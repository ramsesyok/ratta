package atomicwrite

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type failingWriter struct {
	file *os.File
}

func (w *failingWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func (w *failingWriter) Close() error {
	return w.file.Close()
}

func TestWriteFile_Success(t *testing.T) {
	// DD-PERSIST-002 の手順で正常に置き換えられることを確認する。
	dir := t.TempDir()
	targetPath := filepath.Join(dir, "issue.json")
	if err := os.WriteFile(targetPath, []byte("old"), 0o600); err != nil {
		t.Fatalf("write original: %v", err)
	}

	previousNow := now
	now = func() time.Time { return time.Unix(1700000000, 0) }
	t.Cleanup(func() { now = previousNow })

	if err := WriteFile(targetPath, []byte("new")); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	contents, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(contents) != "new" {
		t.Fatalf("unexpected contents: %s", string(contents))
	}

	tmpPath := filepath.Join(dir, "issue.json.tmp."+itoa(os.Getpid())+".1700000000")
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Fatalf("expected temp file cleanup, got err=%v", err)
	}
}

func TestWriteFile_RenameFailureCleansTemp(t *testing.T) {
	// rename 失敗時に元データ保持と一時ファイル削除を行うことを確認する。
	dir := t.TempDir()
	targetPath := filepath.Join(dir, "issue.json")
	if err := os.WriteFile(targetPath, []byte("old"), 0o600); err != nil {
		t.Fatalf("write original: %v", err)
	}

	previousNow := now
	now = func() time.Time { return time.Unix(1700000001, 0) }
	t.Cleanup(func() { now = previousNow })

	previousRename := renameFile
	renameFile = func(_, _ string) error { return errors.New("rename failed") }
	t.Cleanup(func() { renameFile = previousRename })

	if err := WriteFile(targetPath, []byte("new")); err == nil {
		t.Fatal("expected rename error")
	}

	contents, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(contents) != "old" {
		t.Fatalf("unexpected contents: %s", string(contents))
	}

	tmpPath := filepath.Join(dir, "issue.json.tmp."+itoa(os.Getpid())+".1700000001")
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Fatalf("expected temp file cleanup, got err=%v", err)
	}
}

func TestWriteFile_WriteFailureCleansTemp(t *testing.T) {
	// 書き込み失敗時に一時ファイルが削除されることを確認する。
	dir := t.TempDir()
	targetPath := filepath.Join(dir, "issue.json")

	previousNow := now
	now = func() time.Time { return time.Unix(1700000002, 0) }
	t.Cleanup(func() { now = previousNow })

	previousCreate := createTempFile
	createTempFile = func(dir, base string) (io.WriteCloser, string, error) {
		tmpPath := filepath.Join(dir, base+".tmp."+itoa(os.Getpid())+".1700000002")
		file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return nil, "", err
		}
		return &failingWriter{file: file}, tmpPath, nil
	}
	t.Cleanup(func() { createTempFile = previousCreate })

	if err := WriteFile(targetPath, []byte("new")); err == nil {
		t.Fatal("expected write error")
	}

	tmpPath := filepath.Join(dir, "issue.json.tmp."+itoa(os.Getpid())+".1700000002")
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Fatalf("expected temp file cleanup, got err=%v", err)
	}
}

func itoa(value int) string {
	return fmt.Sprintf("%d", value)
}
