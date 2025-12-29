// attachmentstore_test.go は添付保存のテストを行い、UI統合は扱わない。
package attachmentstore

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

// Close は Close の失敗をテストで観測できるようにラップする。
// 目的: クローズエラーを明示的に返す。
// 入力: なし。
// 出力: Close 実行結果のエラー。
// エラー: os.File の Close が失敗した場合に返す。
// 副作用: ファイルディスクリプタを閉じる。
// 並行性: 単一ゴルーチンでの利用を前提とする。
// 不変条件: 返却するエラーは Close の失敗を示す。
// 関連DD: DD-DATA-005
func (w *failingWriter) Close() error {
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	return nil
}

func TestSanitizeFileName_ReplacesInvalidAndTrailing(t *testing.T) {
	// Windows 禁止文字と末尾のドット/スペースが置換されることを確認する。
	input := `re:port<bad>|name. `
	expected := "re_port_bad__name._"

	if got := sanitizeFileName(input); got != expected {
		t.Fatalf("unexpected sanitized name: %s", got)
	}
}

func TestSaveAll_CollisionAddsSuffix(t *testing.T) {
	// 同名の保存先が存在する場合にサフィックスを付けて回避することを確認する。
	dir := t.TempDir()
	issueID := "abcdefghi"
	attachDir := filepath.Join(dir, issueID+attachmentDirExt)
	if err := os.MkdirAll(attachDir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	previousID := newAttachmentID
	newAttachmentID = func() (string, error) { return "ATTACH123", nil }
	t.Cleanup(func() { newAttachmentID = previousID })

	existing := filepath.Join(attachDir, "ATTACH123_report.txt")
	if err := os.WriteFile(existing, []byte("old"), 0o600); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	records, rollback, err := SaveAll(dir, issueID, []Input{{OriginalName: "report.txt", Data: []byte("new")}})
	if err != nil {
		t.Fatalf("SaveAll error: %v", err)
	}
	t.Cleanup(func() {
		if cleanupErr := rollback(); cleanupErr != nil {
			t.Errorf("rollback error: %v", cleanupErr)
		}
	})

	if len(records) != 1 {
		t.Fatalf("unexpected records: %+v", records)
	}
	if records[0].StoredName != "ATTACH123_report_1.txt" {
		t.Fatalf("unexpected stored name: %s", records[0].StoredName)
	}
	if _, statErr := os.Stat(records[0].FullPath); statErr != nil {
		t.Fatalf("expected saved file, err=%v", statErr)
	}
}

func TestSaveAll_RollbackOnFailure(t *testing.T) {
	// 途中で保存に失敗した場合、保存済みの添付が削除されることを確認する。
	dir := t.TempDir()
	issueID := "abcdefghi"

	previousID := newAttachmentID
	counter := 0
	newAttachmentID = func() (string, error) {
		counter++
		if counter == 1 {
			return "ATTACHAAA", nil
		}
		return "ATTACHBBB", nil
	}
	t.Cleanup(func() { newAttachmentID = previousID })

	previousCreate := createTempFile
	callCount := 0
	createTempFile = func(dir, base string) (io.WriteCloser, string, error) {
		callCount++
		tmpPath := filepath.Join(dir, base+".tmp.1.1")
		// #nosec G304 -- テスト用ディレクトリ配下の一時ファイルのみを作成する。
		file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			return nil, "", fmt.Errorf("open temp file: %w", err)
		}
		if callCount == 2 {
			return &failingWriter{file: file}, tmpPath, nil
		}
		return file, tmpPath, nil
	}
	t.Cleanup(func() { createTempFile = previousCreate })

	previousNow := now
	now = func() time.Time { return time.Unix(1700000000, 0) }
	t.Cleanup(func() { now = previousNow })

	_, _, err := SaveAll(dir, issueID, []Input{
		{OriginalName: "a.txt", Data: []byte("ok")},
		{OriginalName: "b.txt", Data: []byte("ng")},
	})
	if err == nil {
		t.Fatal("expected save error")
	}

	firstPath := filepath.Join(dir, issueID+attachmentDirExt, "ATTACHAAA_a.txt")
	if _, statErr := os.Stat(firstPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected rollback to delete first file, err=%v", statErr)
	}
}

func TestSaveAll_EmptyInputs(t *testing.T) {
	// 入力が空の場合に空結果とロールバック関数が返ることを確認する。
	records, rollback, err := SaveAll("dir", "issue", nil)
	if err != nil {
		t.Fatalf("SaveAll error: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("unexpected records: %+v", records)
	}
	if rollback == nil {
		t.Fatal("expected rollback to be set")
	}
}

func TestTrimToLength_Bounds(t *testing.T) {
	// 最大長が0以下の場合は空文字が返ることを確認する。
	if got := trimToLength("abc", 0); got != "" {
		t.Fatalf("unexpected trimmed value: %s", got)
	}
}

func TestSplitExt_NoExtension(t *testing.T) {
	// 拡張子が無い場合にそのまま返ることを確認する。
	name, ext := splitExt("README")
	if name != "README" || ext != "" {
		t.Fatalf("unexpected split: %s %s", name, ext)
	}
}

func TestRemoveAll_ReportsError(t *testing.T) {
	// 削除失敗が集約されることを確認する。
	previousRemove := removeFile
	removeFile = func(string) error { return errors.New("remove failed") }
	t.Cleanup(func() { removeFile = previousRemove })

	err := removeAll([]SavedAttachment{{FullPath: "path"}})
	if err == nil {
		t.Fatal("expected remove error")
	}
}

func TestWriteWithTemp_CloseFailure(t *testing.T) {
	// Close 失敗時にエラーが返ることを確認する。
	dir := t.TempDir()
	previousCreate := createTempFile
	createTempFile = func(dir, base string) (io.WriteCloser, string, error) {
		tmpPath := filepath.Join(dir, base+".tmp.1.2")
		// #nosec G304 -- テスト用ディレクトリ配下の一時ファイルのみを作成する。
		file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			return nil, "", fmt.Errorf("open temp file: %w", err)
		}
		return &failingWriter{file: file}, tmpPath, nil
	}
	t.Cleanup(func() { createTempFile = previousCreate })

	if err := writeWithTemp(dir, "file.txt", []byte("data")); err == nil {
		t.Fatal("expected writeWithTemp error")
	}
}

func TestWriteWithTemp_RenameFailure(t *testing.T) {
	// リネーム失敗時にエラーとなることを確認する。
	dir := t.TempDir()
	previousRename := renameFile
	renameFile = func(_, _ string) error { return errors.New("rename failed") }
	t.Cleanup(func() { renameFile = previousRename })

	if err := writeWithTemp(dir, "file.txt", []byte("data")); err == nil {
		t.Fatal("expected rename error")
	}
}
