// projectroot_test.go はプロジェクトルート操作のテストを行い、UI統合は扱わない。
package projectroot

import (
	"os"
	"path/filepath"
	"testing"
)

type stubConfigRepo struct {
	savedPath string
	err       error
}

func (s *stubConfigRepo) SaveLastProjectRoot(path string) error {
	if s.err != nil {
		return s.err
	}
	s.savedPath = path
	return nil
}

func TestValidateProjectRoot_InvalidPath(t *testing.T) {
	// 存在しないパスは無効になることを確認する。
	service := NewService(nil)
	result, err := service.ValidateProjectRoot(filepath.Join(os.TempDir(), "no-such-dir"))
	if err != nil {
		t.Fatalf("ValidateProjectRoot error: %v", err)
	}
	if result.IsValid {
		t.Fatal("expected invalid result")
	}
}

func TestValidateProjectRoot_NotDirectory(t *testing.T) {
	// ファイルパスは無効になることを確認する。
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	service := NewService(nil)
	result, err := service.ValidateProjectRoot(path)
	if err != nil {
		t.Fatalf("ValidateProjectRoot error: %v", err)
	}
	if result.IsValid {
		t.Fatal("expected invalid result")
	}
}

func TestValidateProjectRoot_Valid(t *testing.T) {
	// ディレクトリが存在すれば有効となることを確認する。
	dir := t.TempDir()
	service := NewService(nil)
	result, err := service.ValidateProjectRoot(dir)
	if err != nil {
		t.Fatalf("ValidateProjectRoot error: %v", err)
	}
	if !result.IsValid {
		t.Fatal("expected valid result")
	}
	if result.NormalizedPath == "" {
		t.Fatal("expected normalized path")
	}
}

func TestValidateProjectRoot_EmptyPath(t *testing.T) {
	// 空パスは無効として扱われることを確認する。
	service := NewService(nil)
	result, err := service.ValidateProjectRoot("")
	if err != nil {
		t.Fatalf("ValidateProjectRoot error: %v", err)
	}
	if result.IsValid {
		t.Fatal("expected invalid result for empty path")
	}
}

func TestCreateProjectRoot_NewDirectory(t *testing.T) {
	// 存在しないパスは作成されることを確認する。
	dir := t.TempDir()
	path := filepath.Join(dir, "new")
	service := NewService(nil)
	result, err := service.CreateProjectRoot(path)
	if err != nil {
		t.Fatalf("CreateProjectRoot error: %v", err)
	}
	if !result.IsValid {
		t.Fatal("expected valid result")
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("expected directory to exist, err=%v", statErr)
	}
}

func TestCreateProjectRoot_EmptyPath(t *testing.T) {
	// 空パスの作成が拒否されることを確認する。
	service := NewService(nil)
	result, err := service.CreateProjectRoot("")
	if err != nil {
		t.Fatalf("CreateProjectRoot error: %v", err)
	}
	if result.IsValid {
		t.Fatal("expected invalid result for empty path")
	}
}

func TestCreateProjectRoot_ExistingPath(t *testing.T) {
	// 既存パスは作成できず無効になることを確認する。
	dir := t.TempDir()
	service := NewService(nil)
	result, err := service.CreateProjectRoot(dir)
	if err != nil {
		t.Fatalf("CreateProjectRoot error: %v", err)
	}
	if result.IsValid {
		t.Fatal("expected invalid result")
	}
}

func TestSaveLastProjectRoot_Delegates(t *testing.T) {
	// config リポジトリへ保存要求が委譲されることを確認する。
	stub := &stubConfigRepo{}
	service := NewService(nil)

	if err := service.SaveLastProjectRoot("path"); err == nil {
		t.Fatal("expected missing config repo error")
	}

	service.configRepo = stub
	if err := service.SaveLastProjectRoot("path"); err != nil {
		t.Fatalf("SaveLastProjectRoot error: %v", err)
	}
	if stub.savedPath != "path" {
		t.Fatalf("unexpected saved path: %s", stub.savedPath)
	}
}
