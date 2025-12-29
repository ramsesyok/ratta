package categoryops

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ratta/internal/domain/issue"
	mod "ratta/internal/domain/mode"
	"ratta/internal/infra/jsonfmt"
)

func TestCreateCategory_DuplicateCaseInsensitive(t *testing.T) {
	// 大小文字違いを含む重複名が拒否されることを確認する。
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "Cat"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	service := NewService(root)
	if _, err := service.CreateCategory("cat", mod.ModeContractor); err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestDeleteCategory_EmptyWithFilesOnly(t *testing.T) {
	// *.json が無く .files のみの場合は削除できることを確認する。
	root := t.TempDir()
	category := "cat"
	if err := os.MkdirAll(filepath.Join(root, category, "issue.files"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	service := NewService(root)
	if err := service.DeleteCategory(category, mod.ModeContractor); err != nil {
		t.Fatalf("DeleteCategory error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, category)); !os.IsNotExist(err) {
		t.Fatalf("expected category to be deleted, err=%v", err)
	}
}

func TestRenameCategory_UpdatesIssueCategory(t *testing.T) {
	// リネーム時に issue.category が更新されることを確認する。
	root := t.TempDir()
	oldName := "old"
	newName := "new"
	if err := os.MkdirAll(filepath.Join(root, oldName), 0o755); err != nil {
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
	if err := os.WriteFile(filepath.Join(root, oldName, "abc123DEF.json"), data, 0o600); err != nil {
		t.Fatalf("write issue: %v", err)
	}

	service := NewService(root)
	if _, err := service.RenameCategory(oldName, newName, mod.ModeContractor); err != nil {
		t.Fatalf("RenameCategory error: %v", err)
	}

	updatedData, err := os.ReadFile(filepath.Join(root, newName, "abc123DEF.json"))
	if err != nil {
		t.Fatalf("read updated issue: %v", err)
	}
	var parsed issue.Issue
	if err := json.Unmarshal(updatedData, &parsed); err != nil {
		t.Fatalf("parse updated issue: %v", err)
	}
	if parsed.Category != newName {
		t.Fatalf("expected updated category: %s", parsed.Category)
	}
}
