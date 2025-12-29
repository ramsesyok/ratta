// Package issuescan はカテゴリ配下の課題走査を担い、編集操作は扱わない。
// スキーマ検証の詳細実装は infra 層に委ねる。
package issuescan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"ratta/internal/infra/schema"
)

// IssueSummary は DD-LOAD-003/004 の課題一覧向け最小情報を表す。
type IssueSummary struct {
	IssueID         string
	Title           string
	Status          string
	Priority        string
	OriginCompany   string
	UpdatedAt       string
	DueDate         string
	Category        string
	IsSchemaInvalid bool
	Path            string
}

// LoadError は DD-LOAD-004 の読み込みエラーを表す。
type LoadError struct {
	Path    string
	Message string
}

// ScanResult は DD-LOAD-003/004 の分類結果を表す。
type ScanResult struct {
	Items      []IssueSummary
	LoadErrors []LoadError
}

// Scanner は DD-LOAD-003 の課題走査を行う。
type Scanner struct {
	validator *schema.Validator
}

// NewScanner は DD-LOAD-003 のスキーマ検証を受け取って生成する。
func NewScanner(validator *schema.Validator) *Scanner {
	return &Scanner{validator: validator}
}

// ScanCategory は DD-LOAD-003/004 のルールでカテゴリ配下を走査する。
// 目的: カテゴリ配下の課題JSONを読み込み一覧項目を収集する。
// 入力: categoryPath はカテゴリパス、categoryName はカテゴリ名。
// 出力: ScanResult とエラー。
// エラー: カテゴリディレクトリの読み取り失敗時に返す。
// 副作用: なし。
// 並行性: 読み取りのみでスレッドセーフ。
// 不変条件: スキーマ不整合の課題は LoadErrors ではなく IsSchemaInvalid で表現する。
// 関連DD: DD-LOAD-003, DD-LOAD-004
func (s *Scanner) ScanCategory(categoryPath, categoryName string) (ScanResult, error) {
	entries, err := os.ReadDir(categoryPath)
	if err != nil {
		return ScanResult{}, fmt.Errorf("read category: %w", err)
	}

	var result ScanResult
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(categoryPath, entry.Name())
		item, readErr := s.readIssue(path, categoryName)
		if readErr != nil {
			result.LoadErrors = append(result.LoadErrors, LoadError{
				Path:    path,
				Message: readErr.Error(),
			})
			continue
		}
		if item != nil {
			result.Items = append(result.Items, *item)
		}
	}

	return result, nil
}

// readIssue は DD-LOAD-004 の課題JSONを読み込み一覧向け情報を抽出する。
// 目的: JSONを解析しスキーマ検証結果を付与して返す。
// 入力: path は課題JSONのパス、categoryName はカテゴリ名。
// 出力: IssueSummary とエラー。
// エラー: 読み取り・JSON解析・検証失敗時に返す。
// 副作用: なし。
// 並行性: 読み取りのみでスレッドセーフ。
// 不変条件: スキーマ不整合時は schemaInvalid を true にする。
// 関連DD: DD-LOAD-004
func (s *Scanner) readIssue(path, categoryName string) (*IssueSummary, error) {
	// #nosec G304 -- カテゴリ配下の列挙結果から生成したパスのみを読む。
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		return nil, fmt.Errorf("read issue: %w", readErr)
	}

	var raw map[string]any
	if unmarshalErr := json.Unmarshal(data, &raw); unmarshalErr != nil {
		return nil, fmt.Errorf("parse json: %w", unmarshalErr)
	}

	if s.validator != nil {
		result, validateErr := s.validator.ValidateIssue(data)
		if validateErr != nil {
			return nil, fmt.Errorf("validate issue: %w", validateErr)
		}
		if len(result.Issues) > 0 {
			return buildSummary(raw, categoryName, path, true), nil
		}
	}

	return buildSummary(raw, categoryName, path, false), nil
}

// buildSummary は DD-LOAD-004 の一覧表示向けフィールドを抽出する。
func buildSummary(raw map[string]any, categoryName, path string, schemaInvalid bool) *IssueSummary {
	return &IssueSummary{
		IssueID:         readString(raw, "issue_id"),
		Title:           readString(raw, "title"),
		Status:          readString(raw, "status"),
		Priority:        readString(raw, "priority"),
		OriginCompany:   readString(raw, "origin_company"),
		UpdatedAt:       readString(raw, "updated_at"),
		DueDate:         readString(raw, "due_date"),
		Category:        categoryName,
		IsSchemaInvalid: schemaInvalid,
		Path:            path,
	}
}

// readString は DD-LOAD-004 の部分表示のために文字列を取り出す。
func readString(raw map[string]any, key string) string {
	value, ok := raw[key]
	if !ok {
		return ""
	}
	typed, ok := value.(string)
	if !ok {
		return ""
	}
	return typed
}
