// issueops_test.go は課題操作ユースケースのテストを行い、UI統合は扱わない。
package issueops

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ratta/internal/domain/issue"
	"ratta/internal/infra/attachmentstore"
	"ratta/internal/infra/jsonfmt"
	"ratta/internal/infra/schema"

	mod "ratta/internal/domain/mode"
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
	if writeErr := os.WriteFile(path, data, 0o600); writeErr != nil {
		t.Fatalf("write issue: %v", writeErr)
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
	if writeErr := os.WriteFile(filepath.Join(root, category, issueID+".json"), data, 0o600); writeErr != nil {
		t.Fatalf("write issue: %v", writeErr)
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
	if writeErr := os.WriteFile(filepath.Join(root, category, issueID+".json"), data, 0o600); writeErr != nil {
		t.Fatalf("write issue: %v", writeErr)
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

func TestGetIssue_NotFound(t *testing.T) {
	// 存在しない課題を読み込むとエラーになることを確認する。
	root := t.TempDir()
	validator, err := schema.NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}
	service := NewService(root, validator)

	if _, err := service.GetIssue("cat", "missing"); err == nil {
		t.Fatal("expected get issue error")
	}
}

func TestUpdateIssue_Success(t *testing.T) {
	// 更新が成功し、更新日時とステータスが反映されることを確認する。
	root := t.TempDir()
	category := "cat"
	if err := os.MkdirAll(filepath.Join(root, category), 0o750); err != nil {
		t.Fatalf("mkdir category: %v", err)
	}
	path := filepath.Join(root, category, "issue.json")
	base := issue.Issue{
		Version:       1,
		IssueID:       "abc123DEF",
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
	if writeErr := os.WriteFile(path, data, 0o600); writeErr != nil {
		t.Fatalf("write issue: %v", writeErr)
	}

	validator, err := schema.NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}
	service := NewService(root, validator)

	updated, err := service.UpdateIssue(category, "issue", mod.ModeVendor, IssueUpdateInput{
		Title:       "new",
		Description: "new",
		DueDate:     "2024-01-03",
		Priority:    issue.PriorityLow,
		Status:      issue.StatusWorking,
	})
	if err != nil {
		t.Fatalf("UpdateIssue error: %v", err)
	}
	if updated.Issue.Status != issue.StatusWorking {
		t.Fatalf("unexpected status: %s", updated.Issue.Status)
	}
	if updated.Issue.UpdatedAt == "2024-01-01T00:00:00Z" {
		t.Fatal("expected updated_at to change")
	}
}

func TestCreateIssue_CategoryMissing(t *testing.T) {
	// カテゴリが存在しない場合に作成できないことを確認する。
	root := t.TempDir()
	validator, err := schema.NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}
	service := NewService(root, validator)

	if _, err := service.CreateIssue("missing", mod.ModeVendor, IssueCreateInput{
		Title:       "title",
		Description: "desc",
		DueDate:     "2024-01-01",
		Priority:    issue.PriorityHigh,
	}); err == nil {
		t.Fatal("expected create issue error")
	}
}

func TestCreateIssue_ValidationError(t *testing.T) {
	// 必須項目が欠ける場合に検証エラーとなることを確認する。
	root := t.TempDir()
	category := "cat"
	if err := os.MkdirAll(filepath.Join(root, category), 0o750); err != nil {
		t.Fatalf("mkdir category: %v", err)
	}
	service := NewService(root, nil)

	if _, err := service.CreateIssue(category, mod.ModeVendor, IssueCreateInput{
		Title:       "",
		Description: "desc",
		DueDate:     "2024-01-01",
		Priority:    issue.PriorityHigh,
	}); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestEnsureCategoryDir_NotDirectory(t *testing.T) {
	// カテゴリパスがファイルの場合にエラーとなることを確認する。
	root := t.TempDir()
	category := "cat"
	path := filepath.Join(root, category)
	if writeErr := os.WriteFile(path, []byte("x"), 0o600); writeErr != nil {
		t.Fatalf("write file: %v", writeErr)
	}

	service := NewService(root, nil)
	if err := service.ensureCategoryDir(category); err == nil {
		t.Fatal("expected not directory error")
	}
}

func TestRankingHelpers(t *testing.T) {
	// 優先度とステータスの順位付けが想定どおりであることを確認する。
	if got := priorityRank(string(issue.PriorityHigh)); got != 0 {
		t.Fatalf("unexpected priority rank: %d", got)
	}
	if got := statusRank(string(issue.StatusClosed)); got <= statusRank(string(issue.StatusOpen)) {
		t.Fatal("expected closed to be ranked after open")
	}
	if got := priorityRank("unknown"); got == 0 {
		t.Fatal("expected unknown priority to be lowest")
	}
	if got := statusRank("unknown"); got == 0 {
		t.Fatal("expected unknown status to be lowest")
	}
}

func TestApplySort_Defaults(t *testing.T) {
	// 未指定ソートが issue_id の昇順になることを確認する。
	items := []IssueSummary{
		{IssueID: "B"},
		{IssueID: "A"},
	}
	applySort(items, "", "")
	if items[0].IssueID != "A" {
		t.Fatalf("unexpected order: %+v", items)
	}
}

func TestApplySort_ByPriorityDesc(t *testing.T) {
	// 優先度で降順ソートされることを確認する。
	items := []IssueSummary{
		{IssueID: "1", Priority: string(issue.PriorityLow)},
		{IssueID: "2", Priority: string(issue.PriorityHigh)},
	}
	applySort(items, "priority", "desc")
	if items[0].IssueID != "1" {
		t.Fatalf("unexpected order: %+v", items)
	}
}

func TestApplySort_ByStatusAsc(t *testing.T) {
	// ステータスで昇順ソートされることを確認する。
	items := []IssueSummary{
		{IssueID: "1", Status: string(issue.StatusResolved)},
		{IssueID: "2", Status: string(issue.StatusOpen)},
	}
	applySort(items, "status", "asc")
	if items[0].IssueID != "2" {
		t.Fatalf("unexpected order: %+v", items)
	}
}

func TestPaginationHelpers(t *testing.T) {
	// ページング補助関数が境界値を補正することを確認する。
	if got := normalizePageSize(0); got != 20 {
		t.Fatalf("unexpected page size: %d", got)
	}
	if got := normalizePage(0); got != 1 {
		t.Fatalf("unexpected page: %d", got)
	}

	items := []IssueSummary{{IssueID: "A"}}
	if got := paginate(items, 2, 10); len(got) != 0 {
		t.Fatalf("unexpected paged length: %d", len(got))
	}
}

func TestOriginCompany_Contractor(t *testing.T) {
	// Contractor モードでは contractor が返ることを確認する。
	if got := originCompany(mod.ModeContractor); got != issue.CompanyContractor {
		t.Fatalf("unexpected origin company: %s", got)
	}
}

func TestAddComment_TooManyAttachments(t *testing.T) {
	// 添付数上限を超える場合にエラーになることを確認する。
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
	if writeErr := os.WriteFile(filepath.Join(root, category, issueID+".json"), data, 0o600); writeErr != nil {
		t.Fatalf("write issue: %v", writeErr)
	}

	validator, err := schema.NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}
	service := NewService(root, validator)

	attachments := make([]CommentAttachmentInput, maxCommentAttachments+1)
	if _, err := service.AddComment(category, issueID, mod.ModeVendor, CommentCreateInput{
		Body:        "body",
		AuthorName:  "author",
		Attachments: attachments,
	}); err == nil {
		t.Fatal("expected too many attachments error")
	}
}

func TestReadIssue_SchemaInvalidVersion(t *testing.T) {
	// バージョン不一致がスキーマ不整合として扱われることを確認する。
	root := t.TempDir()
	category := "cat"
	if err := os.MkdirAll(filepath.Join(root, category), 0o750); err != nil {
		t.Fatalf("mkdir category: %v", err)
	}
	path := filepath.Join(root, category, "issue.json")
	if writeErr := os.WriteFile(path, []byte(`{"version":2,"issue_id":"id","category":"cat"}`), 0o600); writeErr != nil {
		t.Fatalf("write issue: %v", writeErr)
	}

	service := NewService(root, nil)
	detail, err := service.readIssue(path, category)
	if err != nil {
		t.Fatalf("readIssue error: %v", err)
	}
	if !detail.IsSchemaInvalid {
		t.Fatal("expected schema invalid to be true")
	}
}

func TestWriteIssue_InvalidPath(t *testing.T) {
	// 保存先ディレクトリが存在しない場合にエラーとなることを確認する。
	service := NewService("missing", nil)
	err := service.writeIssue(filepath.Join("missing", "cat", "issue.json"), issue.Issue{
		Version:       1,
		IssueID:       "id",
		Category:      "cat",
		Title:         "title",
		Description:   "desc",
		Status:        issue.StatusOpen,
		Priority:      issue.PriorityHigh,
		OriginCompany: issue.CompanyVendor,
		CreatedAt:     "2024-01-01T00:00:00Z",
		UpdatedAt:     "2024-01-01T00:00:00Z",
		DueDate:       "2024-01-02",
		Comments:      []issue.Comment{},
	})
	if err == nil {
		t.Fatal("expected write error")
	}
}
