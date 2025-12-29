// Package logging はログの書き込みとローテーションを担い、UI表示は扱わない。
// 取得したログの解析は他層に委ねる。
package logging

import (
	"encoding/json"
	"errors"
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
// 目的: 指定レベル以上のログを構造化形式で追記する。
// 入力: level はログレベル、message は本文、fields は追加フィールド。
// 出力: なし。
// エラー: 内部でエラーが発生した場合は出力を中断する。
// 副作用: ログファイルへの追記とローテーションを行う。
// 並行性: Logger の mutex で排他制御する。
// 不変条件: 出力行は1行1JSONで末尾に改行を付ける。
// 関連DD: DD-BE-002, BD-FILES-003
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

	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	if _, writeErr := file.Write(line); writeErr != nil {
		if closeErr := file.Close(); closeErr != nil {
			return
		}
		return
	}
	if closeErr := file.Close(); closeErr != nil {
		return
	}
}

// levelString は DD-BE-002/BD-FILES-003 のログレベル表記を返す。
// 目的: ログ出力用の文字列表現に変換する。
// 入力: level はログレベル。
// 出力: ログ文字列。
// エラー: なし。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: 未知値は "error" にフォールバックする。
// 関連DD: DD-BE-002, BD-FILES-003
func levelString(level Level) string {
	switch level {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelError:
		return "error"
	default:
		return "error"
	}
}

// ensureDir は DD-BE-002/BD-FILES-003 のログ出力先ディレクトリを作成する。
// 目的: ログ出力先ディレクトリの存在を保証する。
// 入力: dir はディレクトリパス。
// 出力: 成功時は nil、失敗時はエラー。
// エラー: ディレクトリ作成に失敗した場合に返す。
// 副作用: ディレクトリを作成する。
// 並行性: 同時作成は想定しない。
// 不変条件: 成功時に dir は存在する。
// 関連DD: DD-BE-002, BD-FILES-003
func ensureDir(dir string) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}
	return nil
}

// rotateIfNeeded は BD-FILES-003 のローテーション仕様に従う。
// 目的: サイズ上限を超えたログの世代管理を行う。
// 入力: path はログファイルのパス。
// 出力: 成功時は nil、失敗時はエラー。
// エラー: 取得・リネーム・削除に失敗した場合に返す。
// 副作用: ログファイルの移動・削除を行う。
// 並行性: 同時ローテーションは想定しない。
// 不変条件: 世代数は maxGenerations 以内に収める。
// 関連DD: BD-FILES-003
func rotateIfNeeded(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat log: %w", err)
	}
	if info.Size() < maxSizeBytes {
		return nil
	}

	for i := maxGenerations; i >= 1; i-- {
		if i == maxGenerations {
			removeErr := os.Remove(fmt.Sprintf("%s.%d", path, i))
			if removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				return fmt.Errorf("remove log: %w", removeErr)
			}
		}
	}
	for i := maxGenerations - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", path, i)
		if _, statErr := os.Stat(oldPath); statErr == nil {
			newPath := fmt.Sprintf("%s.%d", path, i+1)
			if renameErr := os.Rename(oldPath, newPath); renameErr != nil {
				return fmt.Errorf("rename log: %w", renameErr)
			}
		}
	}
	if renameErr := os.Rename(path, path+".1"); renameErr != nil {
		return fmt.Errorf("rename log: %w", renameErr)
	}

	return nil
}
