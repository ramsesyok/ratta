// tmpresidue_test.go は一時ファイル残骸処理のテストを行い、UI統合は扱わない。
package tmpresidue

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScanAndHandle_DeletesRecentTmp(t *testing.T) {
	// 24時間未満の一時ファイルは削除され、警告が残らないことを確認する。
	dir := t.TempDir()
	tmpPath := filepath.Join(dir, "issue.json.tmp.123.456")
	if err := os.WriteFile(tmpPath, []byte("tmp"), 0o600); err != nil {
		t.Fatalf("write tmp: %v", err)
	}

	fixedNow := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	previousNow := now
	now = func() time.Time { return fixedNow }
	t.Cleanup(func() { now = previousNow })

	if err := os.Chtimes(tmpPath, fixedNow.Add(-1*time.Hour), fixedNow.Add(-1*time.Hour)); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	results, err := ScanAndHandle(dir)
	if err != nil {
		t.Fatalf("ScanAndHandle error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("unexpected results: %+v", results)
	}
	if _, statErr := os.Stat(tmpPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected temp file to be deleted, err=%v", statErr)
	}
}

func TestScanAndHandle_ReportsOldTmp(t *testing.T) {
	// 24時間以上の一時ファイルは削除せず、警告として記録されることを確認する。
	dir := t.TempDir()
	tmpPath := filepath.Join(dir, "issue.json.tmp.123.789")
	if err := os.WriteFile(tmpPath, []byte("tmp"), 0o600); err != nil {
		t.Fatalf("write tmp: %v", err)
	}

	fixedNow := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	previousNow := now
	now = func() time.Time { return fixedNow }
	t.Cleanup(func() { now = previousNow })

	if err := os.Chtimes(tmpPath, fixedNow.Add(-25*time.Hour), fixedNow.Add(-25*time.Hour)); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	results, err := ScanAndHandle(dir)
	if err != nil {
		t.Fatalf("ScanAndHandle error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("unexpected results: %+v", results)
	}
	if results[0].Target != tmpPath {
		t.Fatalf("unexpected target: %s", results[0].Target)
	}
	if results[0].Hint == "" {
		t.Fatal("expected hint to be set")
	}

	if _, statErr := os.Stat(tmpPath); statErr != nil {
		t.Fatalf("expected temp file to remain, err=%v", statErr)
	}
}

func TestScanAndHandle_DeleteFailureRecorded(t *testing.T) {
	// 削除失敗時に E_IO_WRITE が記録されることを確認する。
	dir := t.TempDir()
	tmpPath := filepath.Join(dir, "issue.json.tmp.123.999")
	if err := os.WriteFile(tmpPath, []byte("tmp"), 0o600); err != nil {
		t.Fatalf("write tmp: %v", err)
	}

	fixedNow := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	previousNow := now
	now = func() time.Time { return fixedNow }
	t.Cleanup(func() { now = previousNow })

	if err := os.Chtimes(tmpPath, fixedNow.Add(-1*time.Hour), fixedNow.Add(-1*time.Hour)); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	previousRemove := removeFile
	removeFile = func(string) error { return errors.New("remove failed") }
	t.Cleanup(func() { removeFile = previousRemove })

	results, err := ScanAndHandle(dir)
	if err != nil {
		t.Fatalf("ScanAndHandle error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("unexpected results: %+v", results)
	}
	if results[0].ErrorCode != ErrCodeIOWrite {
		t.Fatalf("unexpected error code: %s", results[0].ErrorCode)
	}
	if results[0].Target != tmpPath {
		t.Fatalf("unexpected target: %s", results[0].Target)
	}
	if _, statErr := os.Stat(tmpPath); statErr != nil {
		t.Fatalf("expected temp file to remain, err=%v", statErr)
	}
}
