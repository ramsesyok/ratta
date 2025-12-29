// contractorinit_test.go は contractor.json 作成処理のテストを行い、UI統合は扱わない。
package contractorinit

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"ratta/internal/infra/crypto"
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

type errorPrompter struct {
	err error
}

func (p errorPrompter) PromptHidden(_ string) (string, error) {
	return "", p.err
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

func TestRun_PrompterRequired(t *testing.T) {
	// prompter が nil の場合にエラーとなることを確認する。
	if err := Run("path", false, nil); err == nil {
		t.Fatal("expected missing prompter error")
	}
}

func TestRun_PromptError(t *testing.T) {
	// パスワード入力でエラーが発生した場合に失敗することを確認する。
	errPrompter := errorPrompter{err: errors.New("prompt failed")}
	if err := Run("path", false, errPrompter); err == nil {
		t.Fatal("expected prompt error")
	}
}

func TestRun_PasswordMismatch(t *testing.T) {
	// パスワード確認が一致しない場合に失敗することを確認する。
	prompter := &stubPrompter{values: []string{"secret", "other"}}
	if err := Run("path", false, prompter); err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestRun_EmptyPassword(t *testing.T) {
	// 空パスワードが拒否されることを確認する。
	prompter := &stubPrompter{values: []string{"", ""}}
	if err := Run("path", false, prompter); err == nil {
		t.Fatal("expected empty password error")
	}
}

func TestFileExists_UnexpectedError(t *testing.T) {
	// Stat が予期しないエラーを返した場合にエラーが返ることを確認する。
	previousStat := statFile
	statFile = func(string) (os.FileInfo, error) {
		return nil, errors.New("stat failed")
	}
	t.Cleanup(func() { statFile = previousStat })

	if _, err := fileExists("path"); err == nil {
		t.Fatal("expected stat error")
	}
}

func TestRun_GenerateAuthError(t *testing.T) {
	// 認証情報生成が失敗した場合にエラーとなることを確認する。
	previousGenerate := generateAuth
	generateAuth = func(string) (crypto.ContractorAuth, error) {
		return crypto.ContractorAuth{}, errors.New("generate failed")
	}
	t.Cleanup(func() { generateAuth = previousGenerate })

	prompter := &stubPrompter{values: []string{"secret", "secret"}}
	if err := Run("path", true, prompter); err == nil {
		t.Fatal("expected generate error")
	}
}

func TestRun_MarshalAuthError(t *testing.T) {
	// JSON整形が失敗した場合にエラーとなることを確認する。
	previousGenerate := generateAuth
	previousMarshal := marshalAuth
	generateAuth = func(string) (crypto.ContractorAuth, error) {
		return crypto.ContractorAuth{FormatVersion: 1}, nil
	}
	marshalAuth = func(any) ([]byte, error) {
		return nil, errors.New("marshal failed")
	}
	t.Cleanup(func() {
		generateAuth = previousGenerate
		marshalAuth = previousMarshal
	})

	prompter := &stubPrompter{values: []string{"secret", "secret"}}
	if err := Run("path", true, prompter); err == nil {
		t.Fatal("expected marshal error")
	}
}

func TestRun_WriteFileError(t *testing.T) {
	// 書き込み失敗時にエラーとなることを確認する。
	previousGenerate := generateAuth
	previousMarshal := marshalAuth
	previousWrite := writeFile
	generateAuth = func(string) (crypto.ContractorAuth, error) {
		return crypto.ContractorAuth{FormatVersion: 1}, nil
	}
	marshalAuth = func(any) ([]byte, error) { return []byte("{}"), nil }
	writeFile = func(string, []byte) error {
		return errors.New("write failed")
	}
	t.Cleanup(func() {
		generateAuth = previousGenerate
		marshalAuth = previousMarshal
		writeFile = previousWrite
	})

	prompter := &stubPrompter{values: []string{"secret", "secret"}}
	if err := Run("path", true, prompter); err == nil {
		t.Fatal("expected write error")
	}
}

func TestRun_FileExistsError(t *testing.T) {
	// 存在確認が失敗した場合にエラーとなることを確認する。
	previousStat := statFile
	statFile = func(string) (os.FileInfo, error) {
		return nil, errors.New("stat failed")
	}
	t.Cleanup(func() { statFile = previousStat })

	prompter := &stubPrompter{values: []string{"secret", "secret"}}
	if err := Run("path", false, prompter); err == nil {
		t.Fatal("expected file exists error")
	}
}
