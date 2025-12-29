// validator_test.go はスキーマ検証のテストを行い、UI統合は扱わない。
package schema

import (
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func TestValidateIssue_ReturnsIssues(t *testing.T) {
	// 必須項目が欠落している場合に Issues が返ることを確認する。
	validator, err := NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}

	result, err := validator.ValidateIssue([]byte(`{"issue_id":"abc"}`))
	if err != nil {
		t.Fatalf("ValidateIssue error: %v", err)
	}
	if len(result.Issues) == 0 {
		t.Fatal("expected validation issues")
	}
}

func TestValidateContractor_ReturnsIssues(t *testing.T) {
	// contractor.json の必須項目が欠落している場合に Issues が返ることを確認する。
	validator, err := NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}

	result, err := validator.ValidateContractor([]byte(`{"format_version":1}`))
	if err != nil {
		t.Fatalf("ValidateContractor error: %v", err)
	}
	if len(result.Issues) == 0 {
		t.Fatal("expected validation issues")
	}
}

func TestValidateIssue_SchemaMissing(t *testing.T) {
	// スキーマが未ロードの場合にエラーになることを確認する。
	validator := &Validator{schemas: map[string]*jsonschema.Schema{}}
	if _, err := validator.ValidateIssue([]byte(`{}`)); err == nil {
		t.Fatal("expected schema missing error")
	}
}

func TestValidateConfig_ReturnsIssues(t *testing.T) {
	// config.json の必須項目が欠落している場合に Issues が返ることを確認する。
	validator, err := NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}

	result, err := validator.ValidateConfig([]byte(`{"format_version":1}`))
	if err != nil {
		t.Fatalf("ValidateConfig error: %v", err)
	}
	if len(result.Issues) == 0 {
		t.Fatal("expected validation issues")
	}
}

func TestValidationResult_Detail(t *testing.T) {
	// Detail が空と複数エラーの整形を行うことを確認する。
	if detail := (ValidationResult{}).Detail(); detail != "" {
		t.Fatalf("unexpected detail: %s", detail)
	}

	result := ValidationResult{
		Issues: []ValidationIssue{
			{InstanceLocation: "/title", Message: "required"},
			{InstanceLocation: "/status", Message: "invalid"},
		},
	}
	if detail := result.Detail(); detail == "" {
		t.Fatal("expected detail to be set")
	}
}

func TestNewValidatorFromDir_MissingDir(t *testing.T) {
	// 存在しないディレクトリを指定した場合にエラーとなることを確認する。
	if _, err := NewValidatorFromDir(filepath.Join("..", "no-such-dir")); err == nil {
		t.Fatal("expected load error")
	}
}
