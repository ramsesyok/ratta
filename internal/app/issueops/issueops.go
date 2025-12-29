package issueops

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"ratta/internal/domain/id"
	"ratta/internal/domain/issue"
	mod "ratta/internal/domain/mode"
	"ratta/internal/domain/timeutil"
	"ratta/internal/infra/atomicwrite"
	"ratta/internal/infra/attachmentstore"
	"ratta/internal/infra/jsonfmt"
	"ratta/internal/infra/schema"
)

// IssueDetail は DD-LOAD-004/DD-DATA-003 の課題詳細を表す。
type IssueDetail struct {
	IsSchemaInvalid bool
	Issue           issue.Issue
	Path            string
}

// IssueCreateInput は DD-DATA-003 の課題作成入力を表す。
type IssueCreateInput struct {
	Title       string
	Description string
	DueDate     string
	Priority    issue.Priority
	Assignee    string
}

// IssueUpdateInput は DD-DATA-003 の課題更新入力を表す。
type IssueUpdateInput struct {
	Title       string
	Description string
	DueDate     string
	Priority    issue.Priority
	Status      issue.Status
	Assignee    string
}

// CommentCreateInput は DD-DATA-004 のコメント作成入力を表す。
type CommentCreateInput struct {
	Body        string
	AuthorName  string
	Attachments []CommentAttachmentInput
}

// CommentAttachmentInput は DD-DATA-005 の添付入力を表す。
type CommentAttachmentInput struct {
	OriginalName string
	Data         []byte
	MimeType     string
}

// IssueListQuery は DD-BE-003 の IssueListQueryDTO に合わせた条件を表す。
type IssueListQuery struct {
	Page      int
	PageSize  int
	SortBy    string
	SortOrder string
}

// IssueList は DD-BE-003 の IssueListDTO を表す。
type IssueList struct {
	Category string
	Total    int
	Page     int
	PageSize int
	Issues   []IssueSummary
}

// IssueSummary は DD-LOAD-004 の課題一覧項目を表す。
type IssueSummary struct {
	IssueID         string
	Title           string
	Status          string
	Priority        string
	OriginCompany   string
	UpdatedAt       string
	DueDate         string
	Category        string
	IsSchemaInvalid bool
	Path            string
}

// Service は DD-BE-003 の課題永続化と操作を担う。
type Service struct {
	projectRoot string
	validator   *schema.Validator
}

// maxCommentAttachments は DD-DATA-004 の添付上限数を表す。
const maxCommentAttachments = 5

var (
	saveAttachments = attachmentstore.SaveAll
	newCommentID    = id.NewCommentID
	nowISO          = timeutil.NowISO8601
	writeIssueFunc  = func(s *Service, path string, value issue.Issue) error { return s.writeIssue(path, value) }
)

// NewService は DD-BE-003 の課題操作に必要な設定を受け取って生成する。
func NewService(projectRoot string, validator *schema.Validator) *Service {
	return &Service{
		projectRoot: projectRoot,
		validator:   validator,
	}
}

// GetIssue は DD-BE-003 の課題詳細読み込みを行う。
func (s *Service) GetIssue(category, issueID string) (IssueDetail, error) {
	path := filepath.Join(s.projectRoot, category, issueID+".json")
	return s.readIssue(path, category)
}

// CreateIssue は DD-BE-003 の課題作成を行う。
func (s *Service) CreateIssue(category string, currentMode mod.Mode, input IssueCreateInput) (IssueDetail, error) {
	if err := s.ensureCategoryDir(category); err != nil {
		return IssueDetail{}, err
	}

	issueID, err := id.NewIssueID()
	if err != nil {
		return IssueDetail{}, fmt.Errorf("generate issue id: %w", err)
	}

	now := timeutil.NowISO8601()
	newIssue := issue.Issue{
		Version:       1,
		IssueID:       issueID,
		Category:      category,
		Title:         input.Title,
		Description:   input.Description,
		Status:        issue.StatusOpen,
		Priority:      input.Priority,
		OriginCompany: originCompany(currentMode),
		Assignee:      input.Assignee,
		CreatedAt:     now,
		UpdatedAt:     now,
		DueDate:       input.DueDate,
		Comments:      []issue.Comment{},
	}

	if errs := issue.ValidateIssue(newIssue); len(errs) > 0 {
		return IssueDetail{}, errs
	}

	path := filepath.Join(s.projectRoot, category, issueID+".json")
	if err := s.writeIssue(path, newIssue); err != nil {
		return IssueDetail{}, err
	}

	return IssueDetail{Issue: newIssue, Path: path}, nil
}

// UpdateIssue は DD-BE-003 の課題更新を行う。
func (s *Service) UpdateIssue(category, issueID string, currentMode mod.Mode, input IssueUpdateInput) (IssueDetail, error) {
	path := filepath.Join(s.projectRoot, category, issueID+".json")
	current, err := s.readIssue(path, category)
	if err != nil {
		return IssueDetail{}, err
	}
	if current.IsSchemaInvalid {
		return IssueDetail{}, errors.New("schema invalid issue is read-only")
	}
	if current.Issue.Status.IsEndState() {
		return IssueDetail{}, errors.New("closed or rejected issue cannot be updated")
	}
	if !mod.CanTransitionStatus(current.Issue.Status, input.Status, currentMode) {
		return IssueDetail{}, errors.New("status transition not allowed")
	}

	updated := current.Issue
	updated.Title = input.Title
	updated.Description = input.Description
	updated.DueDate = input.DueDate
	updated.Priority = input.Priority
	updated.Status = input.Status
	updated.Assignee = input.Assignee
	updated.UpdatedAt = timeutil.NowISO8601()

	if errs := issue.ValidateIssue(updated); len(errs) > 0 {
		return IssueDetail{}, errs
	}

	if err := s.writeIssue(path, updated); err != nil {
		return IssueDetail{}, err
	}

	return IssueDetail{Issue: updated, Path: path}, nil
}

// AddComment は DD-BE-003/DD-DATA-004 のコメント追加を行う。
func (s *Service) AddComment(category, issueID string, currentMode mod.Mode, input CommentCreateInput) (IssueDetail, error) {
	path := filepath.Join(s.projectRoot, category, issueID+".json")
	current, err := s.readIssue(path, category)
	if err != nil {
		return IssueDetail{}, err
	}
	if current.IsSchemaInvalid {
		return IssueDetail{}, errors.New("schema invalid issue is read-only")
	}
	if current.Issue.Status.IsEndState() {
		return IssueDetail{}, errors.New("closed or rejected issue cannot be updated")
	}

	if len(input.Attachments) > maxCommentAttachments {
		return IssueDetail{}, errors.New("too many attachments")
	}

	commentID, err := newCommentID()
	if err != nil {
		return IssueDetail{}, fmt.Errorf("generate comment id: %w", err)
	}

	issueDir := filepath.Join(s.projectRoot, category)
	var storeInputs []attachmentstore.Input
	for _, attachment := range input.Attachments {
		storeInputs = append(storeInputs, attachmentstore.Input{
			OriginalName: attachment.OriginalName,
			Data:         attachment.Data,
		})
	}
	saved, rollback, err := saveAttachments(issueDir, issueID, storeInputs)
	if err != nil {
		return IssueDetail{}, err
	}

	comment := issue.Comment{
		CommentID:     commentID,
		Body:          input.Body,
		AuthorName:    input.AuthorName,
		AuthorCompany: originCompany(currentMode),
		CreatedAt:     nowISO(),
	}
	for i, savedAttachment := range saved {
		mime := input.Attachments[i].MimeType
		comment.Attachments = append(comment.Attachments, issue.AttachmentRef{
			AttachmentID: savedAttachment.AttachmentID,
			FileName:     savedAttachment.OriginalName,
			StoredName:   savedAttachment.StoredName,
			RelativePath: savedAttachment.RelativePath,
			MimeType:     mime,
			SizeBytes:    int64(len(input.Attachments[i].Data)),
		})
	}

	updated := current.Issue
	updated.Comments = append(updated.Comments, comment)
	updated.UpdatedAt = nowISO()

	if errs := issue.ValidateIssue(updated); len(errs) > 0 {
		if rollback != nil {
			_ = rollback()
		}
		return IssueDetail{}, errs
	}

	if err := writeIssueFunc(s, path, updated); err != nil {
		if rollback != nil {
			_ = rollback()
		}
		return IssueDetail{}, err
	}

	return IssueDetail{Issue: updated, Path: path}, nil
}

// ListIssues は DD-BE-003/DD-LOAD-003 の一覧取得を行う。
func (s *Service) ListIssues(category string, query IssueListQuery) (IssueList, error) {
	categoryPath := filepath.Join(s.projectRoot, category)
	entries, err := os.ReadDir(categoryPath)
	if err != nil {
		return IssueList{}, fmt.Errorf("read category: %w", err)
	}

	var items []IssueSummary
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(categoryPath, entry.Name())
		item, err := s.readIssue(path, category)
		if err != nil {
			continue
		}
		items = append(items, IssueSummary{
			IssueID:         item.Issue.IssueID,
			Title:           item.Issue.Title,
			Status:          string(item.Issue.Status),
			Priority:        string(item.Issue.Priority),
			OriginCompany:   string(item.Issue.OriginCompany),
			UpdatedAt:       item.Issue.UpdatedAt,
			DueDate:         item.Issue.DueDate,
			Category:        category,
			IsSchemaInvalid: item.IsSchemaInvalid,
			Path:            item.Path,
		})
	}

	applySort(items, query.SortBy, query.SortOrder)
	total := len(items)
	pageSize := normalizePageSize(query.PageSize)
	page := normalizePage(query.Page)
	paged := paginate(items, page, pageSize)

	return IssueList{
		Category: category,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Issues:   paged,
	}, nil
}

func (s *Service) readIssue(path, category string) (IssueDetail, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return IssueDetail{}, fmt.Errorf("read issue: %w", err)
	}

	var parsed issue.Issue
	if err := json.Unmarshal(data, &parsed); err != nil {
		return IssueDetail{}, fmt.Errorf("parse issue: %w", err)
	}
	parsed.Category = category

	schemaInvalid := false
	if s.validator != nil {
		result, err := s.validator.ValidateIssue(data)
		if err != nil {
			return IssueDetail{}, fmt.Errorf("validate issue: %w", err)
		}
		schemaInvalid = len(result.Issues) > 0 || parsed.Version != 1
	} else if parsed.Version != 1 {
		schemaInvalid = true
	}

	return IssueDetail{
		IsSchemaInvalid: schemaInvalid,
		Issue:           parsed,
		Path:            path,
	}, nil
}

// writeIssue は DD-PERSIST-002 に従い課題 JSON を保存する。
func (s *Service) writeIssue(path string, value issue.Issue) error {
	data, err := jsonfmt.MarshalIssue(value)
	if err != nil {
		return fmt.Errorf("marshal issue: %w", err)
	}
	if err := atomicwrite.WriteFile(path, data); err != nil {
		return fmt.Errorf("write issue: %w", err)
	}
	return nil
}

// ensureCategoryDir は DD-LOAD-002 のカテゴリディレクトリ存在を確認する。
func (s *Service) ensureCategoryDir(category string) error {
	path := filepath.Join(s.projectRoot, category)
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat category: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("category is not a directory")
	}
	return nil
}

// originCompany は DD-DATA-003 の origin_company を決定する。
func originCompany(current mod.Mode) issue.Company {
	if current == mod.ModeContractor {
		return issue.CompanyContractor
	}
	return issue.CompanyVendor
}

// normalizePageSize は DD-BE-003 のページサイズ既定値を適用する。
func normalizePageSize(size int) int {
	if size <= 0 {
		return 20
	}
	return size
}

// normalizePage は DD-BE-003 の 1-based ページ仕様に合わせる。
func normalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

// paginate は DD-BE-003 のページングを適用する。
func paginate(items []IssueSummary, page, pageSize int) []IssueSummary {
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []IssueSummary{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

// applySort は DD-BE-003 の sort_by/sort_order に従って並べ替える。
func applySort(items []IssueSummary, sortBy, sortOrder string) {
	order := sortOrder
	if order == "" {
		order = "asc"
	}
	less := func(i, j int) bool { return items[i].IssueID < items[j].IssueID }

	switch sortBy {
	case "updated_at":
		less = func(i, j int) bool { return items[i].UpdatedAt < items[j].UpdatedAt }
	case "due_date":
		less = func(i, j int) bool { return items[i].DueDate < items[j].DueDate }
	case "priority":
		less = func(i, j int) bool { return priorityRank(items[i].Priority) < priorityRank(items[j].Priority) }
	case "status":
		less = func(i, j int) bool { return statusRank(items[i].Status) < statusRank(items[j].Status) }
	case "title":
		less = func(i, j int) bool { return items[i].Title < items[j].Title }
	}

	sort.SliceStable(items, func(i, j int) bool {
		if less(i, j) {
			return order != "desc"
		}
		if less(j, i) {
			return order == "desc"
		}
		return items[i].IssueID < items[j].IssueID
	})
}

// priorityRank は DD-DATA-003 の優先度順を数値化する。
func priorityRank(value string) int {
	switch issue.Priority(value) {
	case issue.PriorityHigh:
		return 0
	case issue.PriorityMedium:
		return 1
	case issue.PriorityLow:
		return 2
	default:
		return 3
	}
}

// statusRank は DD-DATA-003/F-004 のステータス順を数値化する。
func statusRank(value string) int {
	switch issue.Status(value) {
	case issue.StatusOpen:
		return 0
	case issue.StatusWorking:
		return 1
	case issue.StatusInquiry:
		return 2
	case issue.StatusHold:
		return 3
	case issue.StatusFeedback:
		return 4
	case issue.StatusResolved:
		return 5
	case issue.StatusClosed:
		return 6
	case issue.StatusRejected:
		return 7
	default:
		return 8
	}
}
