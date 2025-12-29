// Package tmpresidue は一時ファイル残骸の検出と削除を担い、UI表示は扱わない。
// 具体的な通知は上位層に委ねる。
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

var (
	now        = time.Now
	removeFile = os.Remove
	walkDir    = filepath.WalkDir
)

// ScanResult は DD-PERSIST-004 の一時ファイル残骸検出結果を表す。
type ScanResult struct {
	ErrorCode string
	Message   string
	Target    string
	Hint      string
}

// ScanAndHandle は DD-PERSIST-004 に従い *.tmp.* を検出し、削除または警告を記録する。
// 目的: 一時ファイル残骸を削除し、削除できない場合は警告結果を返す。
// 入力: root は走査対象のルートパス。
// 出力: ScanResult の配列とエラー。
// エラー: 走査中のI/Oエラーが発生した場合に返す。
// 副作用: 条件に応じて一時ファイルを削除する。
// 並行性: 同時削除は想定しない。
// 不変条件: 24時間未満は削除、24時間超過は警告として返す。
// 関連DD: DD-PERSIST-004
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

		info, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}

		age := now().Sub(info.ModTime())
		if age < staleThreshold {
			if removeErr := removeFile(path); removeErr != nil {
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

// isTmpArtifact は DD-PERSIST-004 の *.tmp.* 判定を行う。
func isTmpArtifact(name string) bool {
	matched, err := filepath.Match("*.tmp.*", name)
	if err == nil {
		return matched
	}
	return strings.Contains(name, ".tmp.")
}

// shouldSkipDir は DD-PERSIST-004 の検出対象を絞るための除外ルールに従う。
func shouldSkipDir(name string) bool {
	if name == ".tmp_rename" {
		return false
	}
	return strings.HasPrefix(name, ".")
}
