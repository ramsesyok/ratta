package configrepo

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_MissingUsesDefaults(t *testing.T) {
	// config.json が存在しない場合に既定値が返ることを確認する。
	dir := t.TempDir()
	repo := NewRepository(filepath.Join(dir, "ratta.exe"))

	cfg, ok, err := repo.Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if ok {
		t.Fatal("expected has_config to be false")
	}
	if cfg.UI.PageSize != defaultPageSize {
		t.Fatalf("unexpected page size: %d", cfg.UI.PageSize)
	}
	if cfg.Log.Level != "info" {
		t.Fatalf("unexpected log level: %s", cfg.Log.Level)
	}
}

func TestLoad_CorruptReturnsWarning(t *testing.T) {
	// 破損した config.json は警告エラーとして返し、既定値を維持することを確認する。
	dir := t.TempDir()
	repo := NewRepository(filepath.Join(dir, "ratta.exe"))

	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, ok, err := repo.Load()
	if err == nil {
		t.Fatal("expected parse error")
	}
	if ok {
		t.Fatal("expected has_config to be false on parse error")
	}
	if cfg.FormatVersion != formatVersion {
		t.Fatalf("unexpected format version: %d", cfg.FormatVersion)
	}
}

func TestSaveLastProjectRoot_UpdatesPath(t *testing.T) {
	// last_project_root_path を更新して保存できることを確認する。
	dir := t.TempDir()
	repo := NewRepository(filepath.Join(dir, "ratta.exe"))

	if err := repo.Save(DefaultConfig()); err != nil {
		t.Fatalf("Save error: %v", err)
	}
	if err := repo.SaveLastProjectRoot("C:/proj"); err != nil {
		t.Fatalf("SaveLastProjectRoot error: %v", err)
	}

	cfg, ok, err := repo.Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if !ok {
		t.Fatal("expected has_config to be true")
	}
	if cfg.LastProjectRootPath != "C:/proj" {
		t.Fatalf("unexpected last_project_root_path: %s", cfg.LastProjectRootPath)
	}
}

func TestSaveLastProjectRoot_LoadError(t *testing.T) {
	// 既存設定が破損している場合に保存が失敗することを確認する。
	dir := t.TempDir()
	repo := NewRepository(filepath.Join(dir, "ratta.exe"))

	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := repo.SaveLastProjectRoot("C:/proj"); err == nil {
		t.Fatal("expected save last project root error")
	}
}

func TestSave_AtomicWriteFailure(t *testing.T) {
	// atomic write に失敗した場合にエラーが返ることを確認する。
	dir := t.TempDir()
	repo := NewRepository(filepath.Join(dir, "ratta.exe"))

	previous := writeFile
	writeFile = func(string, []byte) error {
		return errors.New("write failed")
	}
	t.Cleanup(func() { writeFile = previous })

	if err := repo.Save(DefaultConfig()); err == nil {
		t.Fatal("expected save error")
	}
}
