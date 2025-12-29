package attachmentstore

import (
	"errors"
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
	if err := os.MkdirAll(attachDir, 0o755); err != nil {
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
	t.Cleanup(func() { _ = rollback() })

	if len(records) != 1 {
		t.Fatalf("unexpected records: %+v", records)
	}
	if records[0].StoredName != "ATTACH123_report_1.txt" {
		t.Fatalf("unexpected stored name: %s", records[0].StoredName)
	}
	if _, err := os.Stat(records[0].FullPath); err != nil {
		t.Fatalf("expected saved file, err=%v", err)
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
		file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return nil, "", err
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
	if _, err := os.Stat(firstPath); !os.IsNotExist(err) {
		t.Fatalf("expected rollback to delete first file, err=%v", err)
	}
}
