package issue

import (
	"strings"
	"testing"
)

func TestValidateCategoryName_Rules(t *testing.T) {
	// カテゴリ名の禁止文字・末尾記号・長さ制約を確認する。
	if errs := ValidateCategoryName("ok"); len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if errs := ValidateCategoryName("bad."); len(errs) == 0 {
		t.Fatal("expected trailing dot error")
	}
	if errs := ValidateCategoryName("bad "); len(errs) == 0 {
		t.Fatal("expected trailing space error")
	}
	if errs := ValidateCategoryName("bad|name"); len(errs) == 0 {
		t.Fatal("expected invalid char error")
	}
	if errs := ValidateCategoryName(strings.Repeat("a", 256)); len(errs) == 0 {
		t.Fatal("expected length error")
	}
}

func TestStatusPriorityCompanyValidation(t *testing.T) {
	// ステータス・優先度・会社種別の妥当性判定を確認する。
	if !StatusOpen.IsValid() || Status("Bad").IsValid() {
		t.Fatal("unexpected status validation")
	}
	if !PriorityHigh.IsValid() || Priority("Bad").IsValid() {
		t.Fatal("unexpected priority validation")
	}
	if !CompanyVendor.IsValid() || Company("Bad").IsValid() {
		t.Fatal("unexpected company validation")
	}
	if !StatusClosed.IsEndState() || StatusOpen.IsEndState() {
		t.Fatal("unexpected end state evaluation")
	}
}

func TestValidateIssue_RequiredFields(t *testing.T) {
	// 必須項目が欠けている場合にエラーになることを確認する。
	errs := ValidateIssue(Issue{})
	if len(errs) == 0 {
		t.Fatal("expected validation errors")
	}
}

func TestValidateIssue_DueDateFormat(t *testing.T) {
	// due_date が YYYY-MM-DD 以外の場合にエラーになることを確認する。
	issue := Issue{
		IssueID:       "abc",
		Category:      "cat",
		Title:         "t",
		Description:   "d",
		Status:        StatusOpen,
		Priority:      PriorityHigh,
		OriginCompany: CompanyVendor,
		CreatedAt:     "2024-01-01T00:00:00Z",
		UpdatedAt:     "2024-01-01T00:00:00Z",
		DueDate:       "2024/01/01",
		Comments:      []Comment{},
	}
	errs := ValidateIssue(issue)
	if len(errs) == 0 {
		t.Fatal("expected due_date error")
	}
}

func TestValidateComment_BodySizeAndAttachments(t *testing.T) {
	// コメント本文のサイズ制限と添付数上限を確認する。
	comment := Comment{
		CommentID:     "id",
		Body:          strings.Repeat("a", maxCommentBodyBytes+1),
		AuthorName:    "name",
		AuthorCompany: CompanyVendor,
		CreatedAt:     "2024-01-01T00:00:00Z",
		Attachments:   make([]AttachmentRef, maxAttachments+1),
	}
	errs := ValidateComment(comment)
	if len(errs) == 0 {
		t.Fatal("expected comment validation errors")
	}
}

func TestValidationError_ErrorMessage(t *testing.T) {
	// 単一エラーが "field: message" 形式になることを確認する。
	err := ValidationError{Field: "title", Message: "required"}
	if err.Error() != "title: required" {
		t.Fatalf("unexpected error message: %s", err.Error())
	}
}

func TestValidationErrors_ErrorMessage(t *testing.T) {
	// 複数エラーがカンマ区切りで連結されることを確認する。
	errs := ValidationErrors{
		{Field: "title", Message: "required"},
		{Field: "description", Message: "required"},
	}
	if errs.Error() != "title: required, description: required" {
		t.Fatalf("unexpected error message: %s", errs.Error())
	}
}

func TestPrefixErrors_AddsPrefix(t *testing.T) {
	// prefixErrors がフィールド名に接頭辞を付与することを確認する。
	errs := ValidationErrors{
		{Field: "body", Message: "required"},
	}
	prefixed := prefixErrors("comments[0].", errs)
	if len(prefixed) != 1 {
		t.Fatalf("unexpected error count: %d", len(prefixed))
	}
	if prefixed[0].Field != "comments[0].body" {
		t.Fatalf("unexpected field: %s", prefixed[0].Field)
	}
}
