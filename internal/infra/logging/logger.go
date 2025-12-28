package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	maxSizeBytes   = 1 << 20
	maxGenerations = 3
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelError
)

// Logger は BD-FILES-003 に従った構造化ログを提供する。
type Logger struct {
	mu   sync.Mutex
	path string
	lvl  Level
}

// NewLogger は DD-BE-002 に従い実行ファイルと同じディレクトリの logs/ratta.log を使う。
func NewLogger(exePath string, level Level) *Logger {
	return &Logger{
		path: filepath.Join(filepath.Dir(exePath), "logs", "ratta.log"),
		lvl:  level,
	}
}

// SetLevel はログレベルを更新する。
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lvl = level
}

// Debug はデバッグログを記録する。
func (l *Logger) Debug(message string, fields map[string]any) {
	l.write(LevelDebug, message, fields)
}

// Info は情報ログを記録する。
func (l *Logger) Info(message string, fields map[string]any) {
	l.write(LevelInfo, message, fields)
}

// Error はエラーログを記録する。
func (l *Logger) Error(message string, fields map[string]any) {
	l.write(LevelError, message, fields)
}

// write は DD-BE-002/BD-FILES-003 のフォーマットでログ行を出力する。
func (l *Logger) write(level Level, message string, fields map[string]any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.lvl {
		return
	}

	if err := ensureDir(filepath.Dir(l.path)); err != nil {
		return
	}

	if err := rotateIfNeeded(l.path); err != nil {
		return
	}

	record := map[string]any{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     levelString(level),
		"message":   message,
	}
	for key, value := range fields {
		record[key] = value
	}

	line, err := json.Marshal(record)
	if err != nil {
		return
	}
	line = append(line, '\n')

	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer file.Close()

	_, _ = file.Write(line)
}

// levelString は DD-BE-002/BD-FILES-003 のログレベル表記を返す。
func levelString(level Level) string {
	switch level {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	default:
		return "error"
	}
}

// ensureDir は DD-BE-002/BD-FILES-003 のログ出力先ディレクトリを作成する。
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

// rotateIfNeeded は BD-FILES-003 のローテーション仕様に従う。
func rotateIfNeeded(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.Size() < maxSizeBytes {
		return nil
	}

	for i := maxGenerations; i >= 1; i-- {
		if i == maxGenerations {
			_ = os.Remove(fmt.Sprintf("%s.%d", path, i))
		}
	}
	for i := maxGenerations - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", path, i)
		if _, err := os.Stat(oldPath); err == nil {
			newPath := fmt.Sprintf("%s.%d", path, i+1)
			if err := os.Rename(oldPath, newPath); err != nil {
				return err
			}
		}
	}
	if err := os.Rename(path, fmt.Sprintf("%s.1", path)); err != nil {
		return err
	}

	return nil
}
