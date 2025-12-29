// error_test.go はエラー変換のテストを行い、UI統合は扱わない。
package present

import (
	"errors"
	"testing"

	"ratta/internal/domain/issue"
)

func TestMapError_ValidationErrors(t *testing.T) {
	// 検証エラーが E_VALIDATION に変換されることを確認する。
	errs := issue.ValidationErrors{
		{Field: "title", Message: "required"},
	}
	dto := MapError(errs)
	if dto.ErrorCode != ErrorValidation {
		t.Fatalf("unexpected code: %s", dto.ErrorCode)
	}
	if dto.Detail == "" {
		t.Fatal("expected detail to be set")
	}
}

func TestMapError_Permission(t *testing.T) {
	// 権限エラーが E_PERMISSION に変換されることを確認する。
	dto := MapError(errors.New("permission denied"))
	if dto.ErrorCode != ErrorPermission {
		t.Fatalf("unexpected code: %s", dto.ErrorCode)
	}
}

func TestMapError_NotFound(t *testing.T) {
	// not found が E_NOT_FOUND になることを確認する。
	dto := MapError(errors.New("category not found"))
	if dto.ErrorCode != ErrorNotFound {
		t.Fatalf("unexpected code: %s", dto.ErrorCode)
	}
}

func TestMapError_Conflict(t *testing.T) {
	// conflict が E_CONFLICT になることを確認する。
	dto := MapError(errors.New("category not empty"))
	if dto.ErrorCode != ErrorConflict {
		t.Fatalf("unexpected code: %s", dto.ErrorCode)
	}
}

func TestMapError_Internal(t *testing.T) {
	// 未分類エラーが E_INTERNAL になることを確認する。
	dto := MapError(errors.New("unexpected"))
	if dto.ErrorCode != ErrorInternal {
		t.Fatalf("unexpected code: %s", dto.ErrorCode)
	}
}

func TestOkAndFail_ResponseEnvelope(t *testing.T) {
	// 成功時と失敗時のレスポンス形式が正しく設定されることを確認する。
	ok := Ok("data")
	if !ok.Ok {
		t.Fatal("expected Ok to be true")
	}
	if ok.Data != "data" {
		t.Fatalf("unexpected data: %v", ok.Data)
	}
	if ok.Error != nil {
		t.Fatal("expected error to be nil")
	}

	fail := Fail(errors.New("permission denied"))
	if fail.Ok {
		t.Fatal("expected Ok to be false")
	}
	if fail.Error == nil {
		t.Fatal("expected error to be set")
	}
	if fail.Error.ErrorCode != ErrorPermission {
		t.Fatalf("unexpected error code: %s", fail.Error.ErrorCode)
	}
}
