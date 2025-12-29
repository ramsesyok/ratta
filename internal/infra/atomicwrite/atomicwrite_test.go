// atomicwrite_test.go は原子的書き込みのテストを行い、UI統合は扱わない。
package atomicwrite

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

type failingWriter struct {
	file *os.File
}

func (w *failingWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

// Close は書き込み失敗時のクローズ失敗も検知できるようにする。
// 目的: Close のエラーをラップして返却しテストで識別可能にする。
// 入力: なし。
// 出力: Close 実行結果のエラー。
// エラー: os.File の Close が失敗した場合に返す。
// 副作用: ファイルディスクリプタを閉じる。
// 並行性: 単一ゴルーチンでの利用を前提とする。
// 不変条件: 返却するエラーは Close の失敗を示す。
// 関連DD: DD-PERSIST-002
func (w *failingWriter) Close() error {
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	return nil
}

type closeFailWriter struct {
	file *os.File
}

func (w *closeFailWriter) Write(p []byte) (int, error) {
	return w.file.Write(p)
}

func (w *closeFailWriter) Close() error {
	_ = w.file.Close()
	return errors.New("close failed")
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

	// #nosec G304 -- テスト用の一時ディレクトリ配下を読むため安全。
	contents, readErr := os.ReadFile(targetPath)
	if readErr != nil {
		t.Fatalf("read target: %v", readErr)
	}
	if string(contents) != "new" {
		t.Fatalf("unexpected contents: %s", string(contents))
	}

	tmpPath := filepath.Join(dir, "issue.json.tmp."+itoa(os.Getpid())+".1700000000")
	if _, statErr := os.Stat(tmpPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected temp file cleanup, got err=%v", statErr)
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

	// #nosec G304 -- テスト用の一時ディレクトリ配下を読むため安全。
	contents, readErr := os.ReadFile(targetPath)
	if readErr != nil {
		t.Fatalf("read target: %v", readErr)
	}
	if string(contents) != "old" {
		t.Fatalf("unexpected contents: %s", string(contents))
	}

	tmpPath := filepath.Join(dir, "issue.json.tmp."+itoa(os.Getpid())+".1700000001")
	if _, statErr := os.Stat(tmpPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected temp file cleanup, got err=%v", statErr)
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
		// #nosec G304 -- テスト用ディレクトリ配下の一時ファイルのみを作成する。
		file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			return nil, "", fmt.Errorf("open temp file: %w", err)
		}
		return &failingWriter{file: file}, tmpPath, nil
	}
	t.Cleanup(func() { createTempFile = previousCreate })

	if err := WriteFile(targetPath, []byte("new")); err == nil {
		t.Fatal("expected write error")
	}

	tmpPath := filepath.Join(dir, "issue.json.tmp."+itoa(os.Getpid())+".1700000002")
	if _, statErr := os.Stat(tmpPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected temp file cleanup, got err=%v", statErr)
	}
}

func TestWriteFile_WriteFailureCleanupError(t *testing.T) {
	// 書き込み失敗時に削除が失敗するとエラーに反映されることを確認する。
	dir := t.TempDir()
	targetPath := filepath.Join(dir, "issue.json")

	previousNow := now
	now = func() time.Time { return time.Unix(1700000003, 0) }
	t.Cleanup(func() { now = previousNow })

	previousRemove := removeFile
	removeFile = func(string) error { return errors.New("remove failed") }
	t.Cleanup(func() { removeFile = previousRemove })

	previousCreate := createTempFile
	createTempFile = func(dir, base string) (io.WriteCloser, string, error) {
		tmpPath := filepath.Join(dir, base+".tmp."+itoa(os.Getpid())+".1700000003")
		// #nosec G304 -- テスト用ディレクトリ配下の一時ファイルのみを作成する。
		file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			return nil, "", fmt.Errorf("open temp file: %w", err)
		}
		return &failingWriter{file: file}, tmpPath, nil
	}
	t.Cleanup(func() { createTempFile = previousCreate })

	if err := WriteFile(targetPath, []byte("new")); err == nil {
		t.Fatal("expected write error with cleanup failure")
	}
}

func TestWriteFile_CloseFailureCleanup(t *testing.T) {
	// Close 失敗時にエラーが返ることを確認する。
	dir := t.TempDir()
	targetPath := filepath.Join(dir, "issue.json")

	previousNow := now
	now = func() time.Time { return time.Unix(1700000004, 0) }
	t.Cleanup(func() { now = previousNow })

	previousCreate := createTempFile
	createTempFile = func(dir, base string) (io.WriteCloser, string, error) {
		tmpPath := filepath.Join(dir, base+".tmp."+itoa(os.Getpid())+".1700000004")
		// #nosec G304 -- テスト用ディレクトリ配下の一時ファイルのみを作成する。
		file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			return nil, "", fmt.Errorf("open temp file: %w", err)
		}
		return &closeFailWriter{file: file}, tmpPath, nil
	}
	t.Cleanup(func() { createTempFile = previousCreate })

	if err := WriteFile(targetPath, []byte("new")); err == nil {
		t.Fatal("expected close error")
	}
}

// itoa はテスト用に PID を文字列化する。
// 目的: テスト内の一時ファイル名を再現する。
// 入力: value は整数値。
// 出力: 10進文字列。
// エラー: なし。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: 返却値は strconv.Itoa と同じ。
// 関連DD: DD-PERSIST-002
func itoa(value int) string {
	return strconv.Itoa(value)
}
