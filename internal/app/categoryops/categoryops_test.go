// categoryops_test.go はカテゴリ操作ユースケースのテストを行い、UI の統合動作は扱わない。
package categoryops

import (
	"encoding/json"
	"os"
	"path/filepath"
	"ratta/internal/domain/issue"
	"ratta/internal/infra/jsonfmt"
	"testing"

	mod "ratta/internal/domain/mode"
)

func TestCreateCategory_DuplicateCaseInsensitive(t *testing.T) {
	// 大小文字違いを含む重複名が拒否されることを確認する。
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "Cat"), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	service := NewService(root)
	if _, createErr := service.CreateCategory("cat", mod.ModeContractor); createErr == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestDeleteCategory_EmptyWithFilesOnly(t *testing.T) {
	// *.json が無く .files のみの場合は削除できることを確認する。
	root := t.TempDir()
	category := "cat"
	if err := os.MkdirAll(filepath.Join(root, category, "issue.files"), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	service := NewService(root)
	if err := service.DeleteCategory(category, mod.ModeContractor); err != nil {
		t.Fatalf("DeleteCategory error: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, category)); !os.IsNotExist(statErr) {
		t.Fatalf("expected category to be deleted, err=%v", statErr)
	}
}

func TestRenameCategory_UpdatesIssueCategory(t *testing.T) {
	// リネーム時に issue.category が更新されることを確認する。
	root := t.TempDir()
	oldName := "old"
	newName := "new"
	if err := os.MkdirAll(filepath.Join(root, oldName), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	item := issue.Issue{
		Version:       1,
		IssueID:       "abc123DEF",
		Category:      oldName,
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
	data, err := jsonfmt.MarshalIssue(item)
	if err != nil {
		t.Fatalf("MarshalIssue error: %v", err)
	}
	if writeErr := os.WriteFile(filepath.Join(root, oldName, "abc123DEF.json"), data, 0o600); writeErr != nil {
		t.Fatalf("write issue: %v", writeErr)
	}

	service := NewService(root)
	if _, renameErr := service.RenameCategory(oldName, newName, mod.ModeContractor); renameErr != nil {
		t.Fatalf("RenameCategory error: %v", renameErr)
	}

	// #nosec G304 -- テスト用一時ディレクトリ配下の固定ファイルを読むため安全。
	updatedData, readErr := os.ReadFile(filepath.Join(root, newName, "abc123DEF.json"))
	if readErr != nil {
		t.Fatalf("read updated issue: %v", readErr)
	}
	var parsed issue.Issue
	if unmarshalErr := json.Unmarshal(updatedData, &parsed); unmarshalErr != nil {
		t.Fatalf("parse updated issue: %v", unmarshalErr)
	}
	if parsed.Category != newName {
		t.Fatalf("expected updated category: %s", parsed.Category)
	}
}
