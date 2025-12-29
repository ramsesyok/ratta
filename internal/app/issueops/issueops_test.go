// issueops_test.go は課題操作ユースケースのテストを行い、UI統合は扱わない。
package issueops

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ratta/internal/domain/issue"
	mod "ratta/internal/domain/mode"
	"ratta/internal/infra/attachmentstore"
	"ratta/internal/infra/jsonfmt"
	"ratta/internal/infra/schema"
)

func TestCreateIssue_SetsDefaults(t *testing.T) {
	// 作成時に origin_company と status が設定され、comments が空であることを確認する。
	root := t.TempDir()
	category := "cat"
	if err := os.MkdirAll(filepath.Join(root, category), 0o750); err != nil {
		t.Fatalf("mkdir category: %v", err)
	}
	validator, err := schema.NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}
	service := NewService(root, validator)

	detail, err := service.CreateIssue(category, mod.ModeVendor, IssueCreateInput{
		Title:       "title",
		Description: "desc",
		DueDate:     "2024-01-01",
		Priority:    issue.PriorityHigh,
	})
	if err != nil {
		t.Fatalf("CreateIssue error: %v", err)
	}
	if detail.Issue.Status != issue.StatusOpen {
		t.Fatalf("unexpected status: %s", detail.Issue.Status)
	}
	if detail.Issue.OriginCompany != issue.CompanyVendor {
		t.Fatalf("unexpected origin company: %s", detail.Issue.OriginCompany)
	}
	if len(detail.Issue.Comments) != 0 {
		t.Fatal("expected empty comments")
	}
}

func TestUpdateIssue_RejectsEndState(t *testing.T) {
	// Closed/Rejected の課題は更新できないことを確認する。
	root := t.TempDir()
	category := "cat"
	if err := os.MkdirAll(filepath.Join(root, category), 0o750); err != nil {
		t.Fatalf("mkdir category: %v", err)
	}
	path := filepath.Join(root, category, "issue.json")
	closed := issue.Issue{
		Version:       1,
		IssueID:       "abc123DEF",
		Category:      category,
		Title:         "title",
		Description:   "desc",
		Status:        issue.StatusClosed,
		Priority:      issue.PriorityHigh,
		OriginCompany: issue.CompanyVendor,
		CreatedAt:     "2024-01-01T00:00:00Z",
		UpdatedAt:     "2024-01-01T00:00:00Z",
		DueDate:       "2024-01-02",
		Comments:      []issue.Comment{},
	}
	data, err := jsonfmt.MarshalIssue(closed)
	if err != nil {
		t.Fatalf("MarshalIssue error: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write issue: %v", err)
	}

	validator, err := schema.NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}
	service := NewService(root, validator)

	if _, updateErr := service.UpdateIssue(category, "issue", mod.ModeContractor, IssueUpdateInput{
		Title:       "new",
		Description: "new",
		DueDate:     "2024-01-03",
		Priority:    issue.PriorityLow,
		Status:      issue.StatusOpen,
	}); updateErr == nil {
		t.Fatal("expected end-state update to fail")
	}
}

func TestUpdateIssue_RejectsSchemaInvalid(t *testing.T) {
	// スキーマ不整合の課題は更新できないことを確認する。
	root := t.TempDir()
	category := "cat"
	if err := os.MkdirAll(filepath.Join(root, category), 0o750); err != nil {
		t.Fatalf("mkdir category: %v", err)
	}
	path := filepath.Join(root, category, "issue.json")
	if err := os.WriteFile(path, []byte(`{"issue_id":"abc123DEF"}`), 0o600); err != nil {
		t.Fatalf("write issue: %v", err)
	}

	validator, err := schema.NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}
	service := NewService(root, validator)

	if _, updateErr := service.UpdateIssue(category, "issue", mod.ModeContractor, IssueUpdateInput{
		Title:       "new",
		Description: "new",
		DueDate:     "2024-01-03",
		Priority:    issue.PriorityLow,
		Status:      issue.StatusOpen,
	}); updateErr == nil {
		t.Fatal("expected schema invalid update to fail")
	}
}

func TestListIssues_SortAndPage(t *testing.T) {
	// ソートとページングの結果が安定していることを確認する。
	root := t.TempDir()
	category := "cat"
	if err := os.MkdirAll(filepath.Join(root, category), 0o750); err != nil {
		t.Fatalf("mkdir category: %v", err)
	}
	writeIssue := func(filename, title, updatedAt string) {
		item := issue.Issue{
			Version:       1,
			IssueID:       strings.TrimSuffix(filename, ".json"),
			Category:      category,
			Title:         title,
			Description:   "desc",
			Status:        issue.StatusOpen,
			Priority:      issue.PriorityHigh,
			OriginCompany: issue.CompanyVendor,
			CreatedAt:     "2024-01-01T00:00:00Z",
			UpdatedAt:     updatedAt,
			DueDate:       "2024-01-02",
			Comments:      []issue.Comment{},
		}
		data, err := jsonfmt.MarshalIssue(item)
		if err != nil {
			t.Fatalf("MarshalIssue error: %v", err)
		}
		if writeErr := os.WriteFile(filepath.Join(root, category, filename), data, 0o600); writeErr != nil {
			t.Fatalf("write issue: %v", writeErr)
		}
	}

	writeIssue("a.json", "Bravo", "2024-01-03T00:00:00Z")
	writeIssue("b.json", "Alpha", "2024-01-02T00:00:00Z")
	writeIssue("c.json", "Alpha", "2024-01-01T00:00:00Z")

	validator, err := schema.NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}
	service := NewService(root, validator)

	list, err := service.ListIssues(category, IssueListQuery{
		Page:      1,
		PageSize:  2,
		SortBy:    "title",
		SortOrder: "asc",
	})
	if err != nil {
		t.Fatalf("ListIssues error: %v", err)
	}
	if list.Total != 3 {
		t.Fatalf("unexpected total: %d", list.Total)
	}
	if len(list.Issues) != 2 {
		t.Fatalf("unexpected page size: %d", len(list.Issues))
	}
	if list.Issues[0].Title != "Alpha" || list.Issues[1].Title != "Alpha" {
		t.Fatalf("unexpected sort order: %+v", list.Issues)
	}
	if list.Issues[0].IssueID > list.Issues[1].IssueID {
		t.Fatal("expected tie-breaker by issue_id")
	}
}

func TestAddComment_Success(t *testing.T) {
	// コメント追加で添付と本文が保存されることを確認する。
	root := t.TempDir()
	category := "cat"
	if err := os.MkdirAll(filepath.Join(root, category), 0o750); err != nil {
		t.Fatalf("mkdir category: %v", err)
	}
	issueID := "abc123DEF"
	base := issue.Issue{
		Version:       1,
		IssueID:       issueID,
		Category:      category,
		Title:         "title",
		Description:   "desc",
		Status:        issue.StatusOpen,
		Priority:      issue.PriorityHigh,
		OriginCompany: issue.CompanyVendor,
		CreatedAt:     "2024-01-01T00:00:00Z",
		UpdatedAt:     "2024-01-01T00:00:00Z",
		DueDate:       "2024-01-02",
		Comments:      []issue.Comment{},
	}
	data, err := jsonfmt.MarshalIssue(base)
	if err != nil {
		t.Fatalf("MarshalIssue error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, category, issueID+".json"), data, 0o600); err != nil {
		t.Fatalf("write issue: %v", err)
	}

	validator, err := schema.NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}
	service := NewService(root, validator)

	detail, err := service.AddComment(category, issueID, mod.ModeVendor, CommentCreateInput{
		Body:       "hello",
		AuthorName: "author",
		Attachments: []CommentAttachmentInput{
			{OriginalName: "file.txt", Data: []byte("data"), MimeType: "text/plain"},
		},
	})
	if err != nil {
		t.Fatalf("AddComment error: %v", err)
	}
	if len(detail.Issue.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(detail.Issue.Comments))
	}
	comment := detail.Issue.Comments[0]
	if comment.Body != "hello" {
		t.Fatalf("unexpected body: %s", comment.Body)
	}
	if len(comment.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(comment.Attachments))
	}
	if _, statErr := os.Stat(filepath.Join(root, category, issueID+".files", comment.Attachments[0].StoredName)); statErr != nil {
		t.Fatalf("expected attachment file, err=%v", statErr)
	}
}

func TestAddComment_RollbackOnWriteFailure(t *testing.T) {
	// JSON 更新失敗時に添付がロールバックされることを確認する。
	root := t.TempDir()
	category := "cat"
	if err := os.MkdirAll(filepath.Join(root, category), 0o750); err != nil {
		t.Fatalf("mkdir category: %v", err)
	}
	issueID := "abc123DEF"
	base := issue.Issue{
		Version:       1,
		IssueID:       issueID,
		Category:      category,
		Title:         "title",
		Description:   "desc",
		Status:        issue.StatusOpen,
		Priority:      issue.PriorityHigh,
		OriginCompany: issue.CompanyVendor,
		CreatedAt:     "2024-01-01T00:00:00Z",
		UpdatedAt:     "2024-01-01T00:00:00Z",
		DueDate:       "2024-01-02",
		Comments:      []issue.Comment{},
	}
	data, err := jsonfmt.MarshalIssue(base)
	if err != nil {
		t.Fatalf("MarshalIssue error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, category, issueID+".json"), data, 0o600); err != nil {
		t.Fatalf("write issue: %v", err)
	}

	validator, err := schema.NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}
	service := NewService(root, validator)

	previousSave := saveAttachments
	previousWrite := writeIssueFunc
	rolledBack := false
	saveAttachments = func(string, string, []attachmentstore.Input) ([]attachmentstore.SavedAttachment, func() error, error) {
		return []attachmentstore.SavedAttachment{
				{
					AttachmentID: "att123",
					OriginalName: "file.txt",
					StoredName:   "att123_file.txt",
					RelativePath: issueID + ".files/att123_file.txt",
					FullPath:     filepath.Join(root, category, issueID+".files", "att123_file.txt"),
				},
			}, func() error {
				rolledBack = true
				return nil
			}, nil
	}
	writeIssueFunc = func(*Service, string, issue.Issue) error {
		return errors.New("write failed")
	}
	t.Cleanup(func() {
		saveAttachments = previousSave
		writeIssueFunc = previousWrite
	})

	if _, addErr := service.AddComment(category, issueID, mod.ModeVendor, CommentCreateInput{
		Body:       "hello",
		AuthorName: "author",
		Attachments: []CommentAttachmentInput{
			{OriginalName: "file.txt", Data: []byte("data")},
		},
	}); addErr == nil {
		t.Fatal("expected add comment failure")
	}
	if !rolledBack {
		t.Fatal("expected rollback to be called")
	}
}
