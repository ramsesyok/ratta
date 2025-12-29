package categoryscan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScan_FlatAndReadOnly(t *testing.T) {
	// 直下ディレクトリのみをカテゴリとし、.tmp_rename は読み取り専用扱いになることを確認する。
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "catA"), 0o755); err != nil {
		t.Fatalf("mkdir catA: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".hidden"), 0o755); err != nil {
		t.Fatalf("mkdir .hidden: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "catA", "nested"), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".tmp_rename", "catB"), 0o755); err != nil {
		t.Fatalf("mkdir tmp_rename: %v", err)
	}

	result, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}
	if len(result.Categories) != 2 {
		t.Fatalf("unexpected category count: %d", len(result.Categories))
	}
	if result.Categories[0].Name != "catA" || result.Categories[0].IsReadOnly {
		t.Fatalf("unexpected category: %+v", result.Categories[0])
	}
	if result.Categories[1].Name != "catB" || !result.Categories[1].IsReadOnly {
		t.Fatalf("unexpected read-only category: %+v", result.Categories[1])
	}
}
