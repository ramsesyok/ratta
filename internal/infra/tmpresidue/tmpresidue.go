package tmpresidue

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ErrCodeIOWrite      = "E_IO_WRITE"
	ErrCodeTmpRemaining = "E_TMP_REMAINING"
)

const staleThreshold = 24 * time.Hour

var now = time.Now
var removeFile = os.Remove
var walkDir = filepath.WalkDir

// ScanResult は一時ファイル残骸の検出結果を表す。
type ScanResult struct {
	ErrorCode string
	Message   string
	Target    string
	Hint      string
}

// ScanAndHandle は DD-PERSIST-004 に従い *.tmp.* を検出し、削除または警告を記録する。
func ScanAndHandle(root string) ([]ScanResult, error) {
	var results []ScanResult

	err := walkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if shouldSkipDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !isTmpArtifact(entry.Name()) {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		age := now().Sub(info.ModTime())
		if age < staleThreshold {
			if err := removeFile(path); err != nil {
				results = append(results, ScanResult{
					ErrorCode: ErrCodeIOWrite,
					Message:   "一時ファイルの削除に失敗しました。",
					Target:    path,
					Hint:      "対象ファイルの権限や利用状況を確認してください。",
				})
			}
			return nil
		}

		results = append(results, ScanResult{
			ErrorCode: ErrCodeTmpRemaining,
			Message:   "24時間以上残っている一時ファイルがあります。",
			Target:    path,
			Hint:      "不要な場合は手動で削除してください。",
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return results, nil
}

func isTmpArtifact(name string) bool {
	matched, err := filepath.Match("*.tmp.*", name)
	if err == nil {
		return matched
	}
	return strings.Contains(name, ".tmp.")
}

func shouldSkipDir(name string) bool {
	if name == ".tmp_rename" {
		return false
	}
	return strings.HasPrefix(name, ".")
}
