// categoryops_test.go はカテゴリ操作ユースケースのテストを行い、UI の統合動作は扱わない。
package categoryops

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ratta/internal/domain/issue"
	"ratta/internal/infra/jsonfmt"

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

func TestCreateCategory_PermissionDenied(t *testing.T) {
	// Vendor モードではカテゴリ作成できないことを確認する。
	root := t.TempDir()
	service := NewService(root)

	if _, err := service.CreateCategory("cat", mod.ModeVendor); err == nil {
		t.Fatal("expected permission error")
	}
}

func TestCreateCategory_InvalidName(t *testing.T) {
	// 禁止文字を含むカテゴリ名は拒否されることを確認する。
	root := t.TempDir()
	service := NewService(root)

	if _, err := service.CreateCategory("bad:name", mod.ModeContractor); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCreateCategory_Success(t *testing.T) {
	// Contractor モードでカテゴリが作成できることを確認する。
	root := t.TempDir()
	service := NewService(root)

	category, err := service.CreateCategory("cat", mod.ModeContractor)
	if err != nil {
		t.Fatalf("CreateCategory error: %v", err)
	}
	if category.Name != "cat" {
		t.Fatalf("unexpected category name: %s", category.Name)
	}
	if _, statErr := os.Stat(category.Path); statErr != nil {
		t.Fatalf("expected category dir to exist, err=%v", statErr)
	}
}

func TestDeleteCategory_PermissionDenied(t *testing.T) {
	// Vendor モードではカテゴリ削除できないことを確認する。
	root := t.TempDir()
	service := NewService(root)

	if err := service.DeleteCategory("cat", mod.ModeVendor); err == nil {
		t.Fatal("expected permission error")
	}
}

func TestDeleteCategory_ReadOnly(t *testing.T) {
	// .tmp_rename 配下のカテゴリは削除できないことを確認する。
	root := t.TempDir()
	category := "cat"
	if err := os.MkdirAll(filepath.Join(root, ".tmp_rename", category), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	service := NewService(root)
	if err := service.DeleteCategory(category, mod.ModeContractor); err == nil {
		t.Fatal("expected read-only error")
	}
}

func TestRenameCategory_TmpResidue(t *testing.T) {
	// .tmp_rename 残骸がある場合はリネームできないことを確認する。
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "old"), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".tmp_rename", "residue"), 0o750); err != nil {
		t.Fatalf("mkdir residue: %v", err)
	}

	service := NewService(root)
	if _, err := service.RenameCategory("old", "new", mod.ModeContractor); err == nil {
		t.Fatal("expected tmp residue error")
	}
}

func TestRenameCategory_NotFound(t *testing.T) {
	// 対象カテゴリが存在しない場合にエラーとなることを確認する。
	root := t.TempDir()
	service := NewService(root)

	if _, err := service.RenameCategory("missing", "new", mod.ModeContractor); err == nil {
		t.Fatal("expected not found error")
	}
}

func TestRenameCategory_RollbackOnParseError(t *testing.T) {
	// 課題JSONの解析失敗時にリネームがロールバックされることを確認する。
	root := t.TempDir()
	oldName := "old"
	newName := "new"
	if err := os.MkdirAll(filepath.Join(root, oldName), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if writeErr := os.WriteFile(filepath.Join(root, oldName, "bad.json"), []byte("{"), 0o600); writeErr != nil {
		t.Fatalf("write issue: %v", writeErr)
	}

	service := NewService(root)
	if _, err := service.RenameCategory(oldName, newName, mod.ModeContractor); err == nil {
		t.Fatal("expected rename error")
	}
	if _, statErr := os.Stat(filepath.Join(root, oldName)); statErr != nil {
		t.Fatalf("expected old category to remain, err=%v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(root, newName)); !os.IsNotExist(statErr) {
		t.Fatalf("expected new category to be absent, err=%v", statErr)
	}
}

func TestDeleteCategory_NotEmpty(t *testing.T) {
	// JSONファイルが存在する場合に削除できないことを確認する。
	root := t.TempDir()
	category := "cat"
	if err := os.MkdirAll(filepath.Join(root, category), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if writeErr := os.WriteFile(filepath.Join(root, category, "issue.json"), []byte("{}"), 0o600); writeErr != nil {
		t.Fatalf("write issue: %v", writeErr)
	}

	service := NewService(root)
	if err := service.DeleteCategory(category, mod.ModeContractor); err == nil {
		t.Fatal("expected not empty error")
	}
}

func TestRenameCategory_PermissionDenied(t *testing.T) {
	// Vendor モードではリネームできないことを確認する。
	root := t.TempDir()
	service := NewService(root)
	if _, err := service.RenameCategory("old", "new", mod.ModeVendor); err == nil {
		t.Fatal("expected permission error")
	}
}

func TestRenameCategory_NameConflict(t *testing.T) {
	// 既存カテゴリと大小文字違いの衝突が拒否されることを確認する。
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "Cat"), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "old"), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	service := NewService(root)
	if _, err := service.RenameCategory("old", "cat", mod.ModeContractor); err == nil {
		t.Fatal("expected name conflict error")
	}
}
