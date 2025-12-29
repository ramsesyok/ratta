// contractorinit_test.go は contractor.json 作成処理のテストを行い、UI統合は扱わない。
package contractorinit

import (
	"errors"
	"os"
	"path/filepath"
	"ratta/internal/infra/crypto"
	"testing"
)

type stubPrompter struct {
	values []string
	index  int
}

func (s *stubPrompter) PromptHidden(_ string) (string, error) {
	if s.index >= len(s.values) {
		return "", errors.New("no input")
	}
	value := s.values[s.index]
	s.index++
	return value, nil
}

func TestRun_CreatesAuthFile(t *testing.T) {
	// contractor.json が存在しない場合に新規作成されることを確認する。
	dir := t.TempDir()
	exePath := filepath.Join(dir, "ratta.exe")

	previousGenerate := generateAuth
	previousMarshal := marshalAuth
	previousWrite := writeFile
	generateAuth = func(string) (crypto.ContractorAuth, error) {
		return crypto.ContractorAuth{FormatVersion: 1}, nil
	}
	marshalAuth = func(any) ([]byte, error) { return []byte("{\"ok\":true}\n"), nil }
	writeFile = func(path string, data []byte) error {
		return os.WriteFile(path, data, 0o600)
	}
	t.Cleanup(func() {
		generateAuth = previousGenerate
		marshalAuth = previousMarshal
		writeFile = previousWrite
	})

	prompter := &stubPrompter{values: []string{"secret", "secret"}}
	if err := Run(exePath, false, prompter); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "auth", "contractor.json")); err != nil {
		t.Fatalf("expected contractor.json to exist, err=%v", err)
	}
}

func TestRun_RejectsOverwriteWithoutForce(t *testing.T) {
	// --force なしで既存ファイルがある場合にエラーになることを確認する。
	dir := t.TempDir()
	exePath := filepath.Join(dir, "ratta.exe")
	authPath := filepath.Join(dir, "auth", "contractor.json")
	if err := os.MkdirAll(filepath.Dir(authPath), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(authPath, []byte("existing"), 0o600); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	prompter := &stubPrompter{values: []string{"secret", "secret"}}
	if err := Run(exePath, false, prompter); err == nil {
		t.Fatal("expected overwrite to be rejected")
	}
}

func TestRun_AllowsOverwriteWithForce(t *testing.T) {
	// --force 指定時は既存ファイルを上書きできることを確認する。
	dir := t.TempDir()
	exePath := filepath.Join(dir, "ratta.exe")
	authPath := filepath.Join(dir, "auth", "contractor.json")
	if err := os.MkdirAll(filepath.Dir(authPath), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(authPath, []byte("existing"), 0o600); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	previousGenerate := generateAuth
	previousMarshal := marshalAuth
	previousWrite := writeFile
	generateAuth = func(string) (crypto.ContractorAuth, error) {
		return crypto.ContractorAuth{FormatVersion: 1}, nil
	}
	marshalAuth = func(any) ([]byte, error) { return []byte("{\"ok\":true}\n"), nil }
	writeFile = func(path string, data []byte) error {
		return os.WriteFile(path, data, 0o600)
	}
	t.Cleanup(func() {
		generateAuth = previousGenerate
		marshalAuth = previousMarshal
		writeFile = previousWrite
	})

	prompter := &stubPrompter{values: []string{"secret", "secret"}}
	if err := Run(exePath, true, prompter); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	// #nosec G304 -- テスト用ディレクトリ配下の固定パスを読むため安全。
	data, readErr := os.ReadFile(authPath)
	if readErr != nil {
		t.Fatalf("read contractor.json: %v", readErr)
	}
	if string(data) != "{\"ok\":true}\n" {
		t.Fatalf("unexpected content: %s", string(data))
	}
}
