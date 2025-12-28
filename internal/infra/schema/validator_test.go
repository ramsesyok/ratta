package schema

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateConfig_Valid(t *testing.T) {
	// config の必須項目が揃っていれば検証を通過することを確認する。
	validator, err := NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}

	input := []byte(`{
  "format_version": 1,
  "last_project_root_path": "C:\\\\proj",
  "log": { "level": "info" },
  "ui": { "page_size": 20 }
}`)

	result, err := validator.ValidateConfig(input)
	if err != nil {
		t.Fatalf("ValidateConfig error: %v", err)
	}
	if len(result.Issues) != 0 {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestValidateConfig_Invalid(t *testing.T) {
	// 型不一致などのエラーで詳細が返ることを確認する。
	validator, err := NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}

	input := []byte(`{
  "format_version": 1,
  "last_project_root_path": "C:\\\\proj",
  "log": { "level": "info" },
  "ui": { "page_size": "twenty" }
}`)

	result, err := validator.ValidateConfig(input)
	if err != nil {
		t.Fatalf("ValidateConfig error: %v", err)
	}
	if len(result.Issues) == 0 {
		t.Fatal("expected validation issues")
	}
	if !strings.Contains(result.Detail(), "/ui/page_size") {
		t.Fatalf("expected detail to include path, got: %s", result.Detail())
	}
}
