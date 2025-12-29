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
		item, err := s.readIssue(path, categoryName)
		if err != nil {
			result.LoadErrors = append(result.LoadErrors, LoadError{
				Path:    path,
				Message: err.Error(),
			})
			continue
		}
		if item != nil {
			result.Items = append(result.Items, *item)
		}
	}

	return result, nil
}

func (s *Scanner) readIssue(path, categoryName string) (*IssueSummary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read issue: %w", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	if s.validator != nil {
		result, err := s.validator.ValidateIssue(data)
		if err != nil {
			return nil, fmt.Errorf("validate issue: %w", err)
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
