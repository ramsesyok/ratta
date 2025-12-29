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
