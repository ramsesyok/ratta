// logger_test.go はログ出力とローテーションのテストを行い、UI統合は扱わない。
package logging

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRotateIfNeeded_RotatesAndKeepsGenerations(t *testing.T) {
	// 1MB 超過時にローテーションが行われ、最大世代数が維持されることを確認する。
	dir := t.TempDir()
	path := filepath.Join(dir, "ratta.log")

	if err := os.WriteFile(path, make([]byte, maxSizeBytes+1), 0o600); err != nil {
		t.Fatalf("write base log: %v", err)
	}
	if err := os.WriteFile(path+".1", []byte("gen1"), 0o600); err != nil {
		t.Fatalf("write gen1: %v", err)
	}
	if err := os.WriteFile(path+".2", []byte("gen2"), 0o600); err != nil {
		t.Fatalf("write gen2: %v", err)
	}
	if err := os.WriteFile(path+".3", []byte("gen3"), 0o600); err != nil {
		t.Fatalf("write gen3: %v", err)
	}

	if err := rotateIfNeeded(path); err != nil {
		t.Fatalf("rotateIfNeeded error: %v", err)
	}

	if _, statErr := os.Stat(path + ".4"); !os.IsNotExist(statErr) {
		t.Fatalf("expected no generation beyond max, err=%v", statErr)
	}
	if _, statErr := os.Stat(path + ".3"); statErr != nil {
		t.Fatalf("expected generation 3 to exist, err=%v", statErr)
	}
	if _, statErr := os.Stat(path + ".2"); statErr != nil {
		t.Fatalf("expected generation 2 to exist, err=%v", statErr)
	}
	if _, statErr := os.Stat(path + ".1"); statErr != nil {
		t.Fatalf("expected generation 1 to exist, err=%v", statErr)
	}
}

func TestLogger_WritesStructuredLog(t *testing.T) {
	// JSON 形式でログが追記されることを確認する。
	dir := t.TempDir()
	logger := NewLogger(filepath.Join(dir, "ratta.exe"), LevelInfo)

	logger.Info("hello", map[string]any{
		"detail": "value",
	})

	// #nosec G304 -- テスト用ディレクトリ配下のログのみを読むため安全。
	data, readErr := os.ReadFile(filepath.Join(dir, "logs", "ratta.log"))
	if readErr != nil {
		t.Fatalf("read log: %v", readErr)
	}
	var parsed map[string]any
	if unmarshalErr := json.Unmarshal(data[:len(data)-1], &parsed); unmarshalErr != nil {
		t.Fatalf("unmarshal log: %v", unmarshalErr)
	}
	if parsed["message"] != "hello" {
		t.Fatalf("unexpected message: %v", parsed["message"])
	}
	if parsed["level"] != "info" {
		t.Fatalf("unexpected level: %v", parsed["level"])
	}
	if parsed["detail"] != "value" {
		t.Fatalf("unexpected detail: %v", parsed["detail"])
	}
}

func TestLogger_RespectsLevel(t *testing.T) {
	// ログレベルで出力が制御されることを確認する。
	dir := t.TempDir()
	logger := NewLogger(filepath.Join(dir, "ratta.exe"), LevelError)

	logger.Info("skip", nil)

	if _, statErr := os.Stat(filepath.Join(dir, "logs", "ratta.log")); !os.IsNotExist(statErr) {
		t.Fatalf("expected no log output, err=%v", statErr)
	}
}
