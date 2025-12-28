package atomicwrite

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

var now = time.Now
var renameFile = os.Rename
var removeFile = os.Remove

type tempFileCreator func(dir, base string) (io.WriteCloser, string, error)

// createTempFile は DD-PERSIST-002 の命名規則で一時ファイルを作成する。
var createTempFile tempFileCreator = func(dir, base string) (io.WriteCloser, string, error) {
	timestamp := now().Unix()
	tmpName := fmt.Sprintf("%s.tmp.%d.%d", base, os.Getpid(), timestamp)
	tmpPath := filepath.Join(dir, tmpName)
	file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, "", err
	}
	return file, tmpPath, nil
}

// WriteFile は DD-PERSIST-002 に従い、同一ディレクトリに一時ファイルを書き出してから rename する。
// DD-PERSIST-003 に従い fsync は行わない。
func WriteFile(targetPath string, data []byte) error {
	dir := filepath.Dir(targetPath)
	base := filepath.Base(targetPath)

	writer, tmpPath, err := createTempFile(dir, base)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		_ = removeFile(tmpPath)
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := writer.Close(); err != nil {
		_ = removeFile(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := renameFile(tmpPath, targetPath); err != nil {
		_ = removeFile(tmpPath)
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}
