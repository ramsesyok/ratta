// Package issue は課題ドメインの検証と型定義を提供し、永続化は扱わない。
// 検証ルールは詳細設計に従う。
package issue

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maxNameLength       = 255
	maxCommentBodyBytes = 100 * 1024
	maxAttachments      = 5
)

// ValidationError は DD-DATA-003/004 の入力不整合を表す。
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors は DD-DATA-003/004 の複数エラーをまとめる。
type ValidationErrors []ValidationError

// Error は DD-DATA-003/004 の検証エラーを連結して返す。
// 目的: 複数エラーを単一の文字列にまとめる。
// 入力: e は検証エラー群。
// 出力: エラーメッセージ文字列。
// エラー: なし。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: 各エラーは ", " で連結する。
// 関連DD: DD-DATA-003, DD-DATA-004
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	parts := make([]string, 0, len(e))
	for _, item := range e {
		parts = append(parts, item.Error())
	}
	return strings.Join(parts, ", ")
}

// ValidateCategoryName は DD-DATA-003 のカテゴリ名ルールを検証する。
func ValidateCategoryName(name string) ValidationErrors {
	var errs ValidationErrors
	if name == "" {
		errs = append(errs, ValidationError{Field: "category", Message: "required"})
		return errs
	}
	if utf8.RuneCountInString(name) > maxNameLength {
		errs = append(errs, ValidationError{Field: "category", Message: "too long"})
	}
	if hasInvalidCategoryChar(name) {
		errs = append(errs, ValidationError{Field: "category", Message: "contains invalid characters"})
	}
	if hasTrailingDotOrSpace(name) {
		errs = append(errs, ValidationError{Field: "category", Message: "trailing dot or space"})
	}
	return errs
}

// ValidateIssue は DD-DATA-003/004 の必須項目・形式を検証する。
func ValidateIssue(issue Issue) ValidationErrors {
	var errs ValidationErrors

	if issue.IssueID == "" {
		errs = append(errs, ValidationError{Field: "issue_id", Message: "required"})
	}
	errs = append(errs, ValidateCategoryName(issue.Category)...)
	if err := validateRequiredLength("title", issue.Title, maxNameLength); err != nil {
		errs = append(errs, *err)
	}
	if err := validateRequiredLength("description", issue.Description, maxNameLength); err != nil {
		errs = append(errs, *err)
	}
	if !issue.Status.IsValid() {
		errs = append(errs, ValidationError{Field: "status", Message: "invalid"})
	}
	if !issue.Priority.IsValid() {
		errs = append(errs, ValidationError{Field: "priority", Message: "invalid"})
	}
	if !issue.OriginCompany.IsValid() {
		errs = append(errs, ValidationError{Field: "origin_company", Message: "invalid"})
	}
	if issue.CreatedAt == "" {
		errs = append(errs, ValidationError{Field: "created_at", Message: "required"})
	}
	if issue.UpdatedAt == "" {
		errs = append(errs, ValidationError{Field: "updated_at", Message: "required"})
	}
	if issue.DueDate == "" {
		errs = append(errs, ValidationError{Field: "due_date", Message: "required"})
	} else if !isValidDate(issue.DueDate) {
		errs = append(errs, ValidationError{Field: "due_date", Message: "invalid format"})
	}
	if issue.Comments == nil {
		errs = append(errs, ValidationError{Field: "comments", Message: "required"})
	} else {
		for i, comment := range issue.Comments {
			errs = append(errs, prefixErrors(fmt.Sprintf("comments[%d].", i), ValidateComment(comment))...)
		}
	}

	return errs
}

// ValidateComment は DD-DATA-004 のコメント必須項目を検証する。
func ValidateComment(comment Comment) ValidationErrors {
	var errs ValidationErrors
	if comment.CommentID == "" {
		errs = append(errs, ValidationError{Field: "comment_id", Message: "required"})
	}
	if comment.Body == "" {
		errs = append(errs, ValidationError{Field: "body", Message: "required"})
	} else if len([]byte(comment.Body)) > maxCommentBodyBytes {
		errs = append(errs, ValidationError{Field: "body", Message: "too large"})
	}
	if err := validateRequiredLength("author_name", comment.AuthorName, maxNameLength); err != nil {
		errs = append(errs, *err)
	}
	if !comment.AuthorCompany.IsValid() {
		errs = append(errs, ValidationError{Field: "author_company", Message: "invalid"})
	}
	if comment.CreatedAt == "" {
		errs = append(errs, ValidationError{Field: "created_at", Message: "required"})
	}
	if len(comment.Attachments) > maxAttachments {
		errs = append(errs, ValidationError{Field: "attachments", Message: "too many"})
	}
	return errs
}

// validateRequiredLength は DD-DATA-003/004 の必須・長さ制約を検証する。
// 目的: 必須項目と最大長の制約を検証する。
// 入力: field は対象フィールド名、value は値、maxLen は最大文字数。
// 出力: エラーがあれば ValidationError、なければ nil。
// エラー: なし。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: 空文字は required エラーとなる。
// 関連DD: DD-DATA-003, DD-DATA-004
func validateRequiredLength(field, value string, maxLen int) *ValidationError {
	if value == "" {
		return &ValidationError{Field: field, Message: "required"}
	}
	if utf8.RuneCountInString(value) > maxLen {
		return &ValidationError{Field: field, Message: "too long"}
	}
	return nil
}

// prefixErrors は DD-DATA-003/004 の配列項目エラーにプレフィックスを付ける。
// 目的: 配列要素のエラーにインデックス接頭辞を付与する。
// 入力: prefix は付与する接頭辞、errs は元エラー。
// 出力: 接頭辞付きの ValidationErrors。
// エラー: なし。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: errs が空の場合は nil を返す。
// 関連DD: DD-DATA-003, DD-DATA-004
func prefixErrors(prefix string, errs ValidationErrors) ValidationErrors {
	if len(errs) == 0 {
		return nil
	}
	prefixed := make(ValidationErrors, 0, len(errs))
	for _, err := range errs {
		prefixed = append(prefixed, ValidationError{
			Field:   prefix + err.Field,
			Message: err.Message,
		})
	}
	return prefixed
}

// isValidDate は DD-DATA-002 の日付フォーマットを検証する。
func isValidDate(value string) bool {
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

// hasInvalidCategoryChar は DD-DATA-003 の禁止文字を検出する。
func hasInvalidCategoryChar(value string) bool {
	for _, r := range value {
		if r < 0x20 {
			return true
		}
		switch r {
		case '\\', '/', ':', '*', '?', '"', '<', '>', '|':
			return true
		default:
		}
	}
	return false
}

// hasTrailingDotOrSpace は DD-DATA-003 の末尾記号ルールを検証する。
func hasTrailingDotOrSpace(value string) bool {
	if value == "" {
		return false
	}
	last := value[len(value)-1]
	return last == '.' || last == ' '
}
