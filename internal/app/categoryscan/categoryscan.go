// Package categoryscan はカテゴリ一覧の走査を担い、書き込みやUI表示は扱わない。
// 走査対象の正当性検証は上位層に委ねる。
package categoryscan

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Category は DD-LOAD-002 のカテゴリ情報を表す。
type Category struct {
	Name       string
	IsReadOnly bool
	Path       string
}

// ScanResult は DD-LOAD-002 のカテゴリ一覧結果を表す。
type ScanResult struct {
	Categories []Category
	ErrorCount int
}

// Scan は DD-LOAD-002 のルールでカテゴリを走査する。
// 目的: プロジェクトルート配下のカテゴリを一覧化する。
// 入力: root はプロジェクトルートパス。
// 出力: ScanResult とエラー。
// エラー: 走査対象ディレクトリの読み取りに失敗した場合に返す。
// 副作用: なし。
// 並行性: 読み取りのみでスレッドセーフ。
// 不変条件: 返却するカテゴリ一覧は名前順にソートされる。
// 関連DD: DD-LOAD-002
func Scan(root string) (ScanResult, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return ScanResult{}, fmt.Errorf("read project root: %w", err)
	}

	categories := make([]Category, 0, len(entries))
	readOnlyNames := make(map[string]struct{})

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == ".tmp_rename" {
			tmpPath := filepath.Join(root, name)
			tmpEntries, readErr := os.ReadDir(tmpPath)
			if readErr != nil {
				return ScanResult{}, fmt.Errorf("read .tmp_rename: %w", readErr)
			}
			for _, tmpEntry := range tmpEntries {
				if !tmpEntry.IsDir() {
					continue
				}
				readOnlyNames[tmpEntry.Name()] = struct{}{}
			}
			continue
		}
		if shouldSkipDir(name) {
			continue
		}
		categories = append(categories, Category{
			Name:       name,
			IsReadOnly: false,
			Path:       filepath.Join(root, name),
		})
	}

	for name := range readOnlyNames {
		categories = append(categories, Category{
			Name:       name,
			IsReadOnly: true,
			Path:       filepath.Join(root, ".tmp_rename", name),
		})
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	return ScanResult{Categories: categories}, nil
}

// shouldSkipDir は DD-LOAD-002 の除外ルールを適用する。
func shouldSkipDir(name string) bool {
	if name == ".tmp_rename" {
		return false
	}
	if strings.HasPrefix(name, ".") {
		return true
	}
	return name == ".git"
}
