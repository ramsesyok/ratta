// issuescan_test.go は課題走査のテストを行い、UI統合は扱わない。
package issuescan

import (
	"os"
	"path/filepath"
	"ratta/internal/infra/jsonfmt"
	"ratta/internal/infra/schema"
	"testing"
)

func TestScanCategory_ClassifiesIssues(t *testing.T) {
	// 破損 JSON はエラーになり、スキーマ不整合は一覧に含まれることを確認する。
	dir := t.TempDir()

	validIssue := map[string]any{
		"version":        1,
		"issue_id":       "abc123DEF",
		"category":       "cat",
		"title":          "Title",
		"description":    "Desc",
		"status":         "Open",
		"priority":       "High",
		"origin_company": "Vendor",
		"created_at":     "2024-01-01T00:00:00Z",
		"updated_at":     "2024-01-02T00:00:00Z",
		"due_date":       "2024-01-03",
		"comments":       []any{},
	}

	validData, err := jsonfmt.MarshalIssue(validIssue)
	if err != nil {
		t.Fatalf("MarshalIssue error: %v", err)
	}

	if writeErr := os.WriteFile(filepath.Join(dir, "valid.json"), validData, 0o600); writeErr != nil {
		t.Fatalf("write valid: %v", writeErr)
	}
	if writeErr := os.WriteFile(filepath.Join(dir, "invalid.json"), []byte("{"), 0o600); writeErr != nil {
		t.Fatalf("write invalid: %v", writeErr)
	}

	schemaInvalid := map[string]any{
		"version":        1,
		"issue_id":       "abc123DEF",
		"category":       "cat",
		"title":          "Title",
		"description":    "Desc",
		"status":         "Open",
		"priority":       "High",
		"origin_company": "Vendor",
		"created_at":     "2024-01-01T00:00:00Z",
		"updated_at":     "2024-01-02T00:00:00Z",
		"due_date":       "2024-01-03",
	}
	invalidSchemaData, err := jsonfmt.MarshalIssue(schemaInvalid)
	if err != nil {
		t.Fatalf("MarshalIssue error: %v", err)
	}
	if writeErr := os.WriteFile(filepath.Join(dir, "schema_invalid.json"), invalidSchemaData, 0o600); writeErr != nil {
		t.Fatalf("write schema invalid: %v", writeErr)
	}

	validator, err := schema.NewValidatorFromDir(filepath.Join("..", "..", "..", "schemas"))
	if err != nil {
		t.Fatalf("NewValidatorFromDir error: %v", err)
	}
	scanner := NewScanner(validator)
	result, err := scanner.ScanCategory(dir, "cat")
	if err != nil {
		t.Fatalf("ScanCategory error: %v", err)
	}
	if len(result.LoadErrors) != 1 {
		t.Fatalf("expected 1 load error, got %d", len(result.LoadErrors))
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	var schemaInvalidFound bool
	for _, item := range result.Items {
		if item.IsSchemaInvalid {
			schemaInvalidFound = true
		}
	}
	if !schemaInvalidFound {
		t.Fatal("expected schema invalid item")
	}
}
