// Package atomicwrite は原子的なファイル書き込みを提供し、上位の整形や検証は扱わない。
// fsync や同期保証の強化は対象外とする。
package atomicwrite

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

var (
	now        = time.Now
	renameFile = os.Rename
	removeFile = os.Remove
)

type tempFileCreator func(dir, base string) (io.WriteCloser, string, error)

// createTempFile は DD-PERSIST-002 の命名規則で一時ファイルを作成する。
var createTempFile tempFileCreator = func(dir, base string) (io.WriteCloser, string, error) {
	timestamp := now().Unix()
	tmpName := fmt.Sprintf("%s.tmp.%d.%d", base, os.Getpid(), timestamp)
	tmpPath := filepath.Join(dir, tmpName)
	// #nosec G304 -- 生成済みの一時ファイル名のみを利用するため安全。
	file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, "", fmt.Errorf("open temp file: %w", err)
	}
	return file, tmpPath, nil
}

// WriteFile は DD-PERSIST-002 に従い、同一ディレクトリに一時ファイルを書き出してから rename する。
// 目的: 一時ファイルを使って原子的に内容を更新する。
// 入力: targetPath は保存先、data は書き込むバイト列。
// 出力: 成功時は nil、失敗時はエラー。
// エラー: 一時ファイル作成、書き込み、リネーム失敗時に返す。
// 副作用: 一時ファイル作成・削除とターゲットファイル更新を行う。
// 並行性: 同一ファイルへの同時書き込みは想定しない。
// 不変条件: 書き込み失敗時はターゲットファイルを変更しない。
// 関連DD: DD-PERSIST-002, DD-PERSIST-003
func WriteFile(targetPath string, data []byte) error {
	dir := filepath.Dir(targetPath)
	base := filepath.Base(targetPath)

	writer, tmpPath, err := createTempFile(dir, base)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	if _, writeErr := writer.Write(data); writeErr != nil {
		closeErr := writer.Close()
		removeErr := removeFile(tmpPath)
		if closeErr != nil {
			return fmt.Errorf("write temp file failed: %w; close error: %s", writeErr, closeErr.Error())
		}
		if removeErr != nil {
			return fmt.Errorf("write temp file failed: %w; cleanup error: %s", writeErr, removeErr.Error())
		}
		return fmt.Errorf("write temp file: %w", writeErr)
	}

	if closeErr := writer.Close(); closeErr != nil {
		removeErr := removeFile(tmpPath)
		if removeErr != nil {
			return fmt.Errorf("close temp file failed: %w; cleanup error: %s", closeErr, removeErr.Error())
		}
		return fmt.Errorf("close temp file: %w", closeErr)
	}

	if renameErr := renameFile(tmpPath, targetPath); renameErr != nil {
		removeErr := removeFile(tmpPath)
		if removeErr != nil {
			return fmt.Errorf("rename temp file failed: %w; cleanup error: %s", renameErr, removeErr.Error())
		}
		return fmt.Errorf("rename temp file: %w", renameErr)
	}

	return nil
}
