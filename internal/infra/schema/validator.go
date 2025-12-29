// Package schema は JSON スキーマ検証を担い、ファイルI/OやUI表示は扱わない。
// スキーマの読み込みは別ファイルで実施する。
package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

const (
	IssueSchemaName      = "issue.schema.json"
	ConfigSchemaName     = "config.schema.json"
	ContractorSchemaName = "contractor.schema.json"
)

// Validator は DD-BE-002 のスキーマ検証方針に従い検証を行う。
type Validator struct {
	schemas map[string]*jsonschema.Schema
}

// ValidationIssue はスキーマ不整合の詳細を表す。
type ValidationIssue struct {
	InstanceLocation string
	Message          string
}

// ValidationResult は DD-BE-002 のスキーマ検証結果を表す。
type ValidationResult struct {
	Issues []ValidationIssue
}

// Detail は DD-BE-002 のエラー報告に合わせ、APIErrorDTO.detail を組み立てる。
func (r ValidationResult) Detail() string {
	if len(r.Issues) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, issue := range r.Issues {
		builder.WriteString(issue.InstanceLocation)
		builder.WriteString(": ")
		builder.WriteString(issue.Message)
		builder.WriteString("\n")
	}
	return strings.TrimRight(builder.String(), "\n")
}

// NewValidatorFromDir は DD-BE-002 に従い schemas/ 配下のスキーマを読み込む。
// 目的: 検証に必要なスキーマ群を読み込み Validator を生成する。
// 入力: dir はスキーマディレクトリ。
// 出力: Validator とエラー。
// エラー: スキーマ読み込み失敗時に返す。
// 副作用: スキーマファイルを読み取る。
// 並行性: 読み取りのみでスレッドセーフ。
// 不変条件: 必須スキーマが揃っていること。
// 関連DD: DD-BE-002
func NewValidatorFromDir(dir string) (*Validator, error) {
	compiled, err := LoadSchemasFromDir(dir)
	if err != nil {
		return nil, fmt.Errorf("load schemas: %w", err)
	}
	return &Validator{schemas: compiled}, nil
}

// ValidateIssue は DD-DATA-003 の issue スキーマを検証する。
func (v *Validator) ValidateIssue(data []byte) (ValidationResult, error) {
	return v.validateBytes(IssueSchemaName, data)
}

// ValidateConfig は DD-DATA-001 の config スキーマを検証する。
func (v *Validator) ValidateConfig(data []byte) (ValidationResult, error) {
	return v.validateBytes(ConfigSchemaName, data)
}

// ValidateContractor は DD-DATA-001 の contractor スキーマを検証する。
func (v *Validator) ValidateContractor(data []byte) (ValidationResult, error) {
	return v.validateBytes(ContractorSchemaName, data)
}

// validateBytes は DD-BE-002 の共通検証処理を行う。
// 目的: 指定スキーマで JSON データを検証する。
// 入力: schemaName はスキーマ名、data は JSON バイト列。
// 出力: ValidationResult とエラー。
// エラー: パース・検証失敗時に返す。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: スキーマ不整合は ValidationResult に格納する。
// 関連DD: DD-BE-002
func (v *Validator) validateBytes(schemaName string, data []byte) (ValidationResult, error) {
	schema, ok := v.schemas[schemaName]
	if !ok {
		return ValidationResult{}, fmt.Errorf("schema not loaded: %s", schemaName)
	}

	var value any
	if unmarshalErr := json.Unmarshal(data, &value); unmarshalErr != nil {
		return ValidationResult{}, fmt.Errorf("parse json: %w", unmarshalErr)
	}

	if err := schema.Validate(value); err != nil {
		issues := collectIssues(err)
		if len(issues) > 0 {
			return ValidationResult{Issues: issues}, nil
		}
		return ValidationResult{}, fmt.Errorf("validate schema: %w", err)
	}

	return ValidationResult{}, nil
}

// collectIssues は DD-BE-002 の詳細表示向けに検証エラーを収集する。
// 目的: スキーマ検証エラーを一覧に変換する。
// 入力: err はスキーマ検証エラー。
// 出力: ValidationIssue の配列。
// エラー: なし。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: ValidationError 以外は空配列を返す。
// 関連DD: DD-BE-002
func collectIssues(err error) []ValidationIssue {
	var validationErr *jsonschema.ValidationError
	if !errors.As(err, &validationErr) {
		return nil
	}
	var issues []ValidationIssue
	flattenIssues(&issues, validationErr)
	return issues
}

// flattenIssues は DD-BE-002 の詳細表示向けに検証エラーを平坦化する。
func flattenIssues(issues *[]ValidationIssue, err *jsonschema.ValidationError) {
	if len(err.Causes) == 0 {
		location := err.InstanceLocation
		if location == "" {
			location = "/"
		}
		*issues = append(*issues, ValidationIssue{
			InstanceLocation: location,
			Message:          err.Message,
		})
		return
	}
	for _, cause := range err.Causes {
		flattenIssues(issues, cause)
	}
}
