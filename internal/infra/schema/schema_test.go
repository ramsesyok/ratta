package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSchemasFromDir_AllowsLocalSchemas(t *testing.T) {
	// schemas/ 配下のローカルスキーマが問題なくコンパイルできることを確認する。
	baseDir := filepath.Join("..", "..", "..", "schemas")
	compiled, err := LoadSchemasFromDir(baseDir)
	if err != nil {
		t.Fatalf("LoadSchemasFromDir error: %v", err)
	}

	for _, name := range []string{
		"issue.schema.json",
		"config.schema.json",
		"contractor.schema.json",
	} {
		if compiled[name] == nil {
			t.Fatalf("expected schema %s to be loaded", name)
		}
	}
}

func TestLoadSchemasFromDir_RejectsHTTPRefs(t *testing.T) {
	// 外部 HTTP 参照が含まれる場合に拒否されることを確認する。
	tempDir := t.TempDir()
	schema := `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "root.schema.json",
  "type": "object",
  "properties": {
    "value": {
      "$ref": "http://example.com/other.schema.json"
    }
  }
}`

	if err := os.WriteFile(filepath.Join(tempDir, "root.schema.json"), []byte(schema), 0o600); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	_, err := LoadSchemasFromDir(tempDir)
	if err == nil {
		t.Fatal("expected external ref to be rejected")
	}
}

func TestOpenSchemaFile_RejectsOutside(t *testing.T) {
	// baseDir 外の参照が拒否されることを確認する。
	dir := t.TempDir()
	if _, err := openSchemaFile(dir, filepath.Join("..", "outside.json")); err == nil {
		t.Fatal("expected outside ref error")
	}
}
