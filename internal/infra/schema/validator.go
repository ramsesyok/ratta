package schema

import (
	"encoding/json"
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

// ValidationResult は検証結果を表す。
type ValidationResult struct {
	Issues []ValidationIssue
}

// Detail は ApiErrorDTO.detail に入れる想定の文字列を返す。
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
func NewValidatorFromDir(dir string) (*Validator, error) {
	compiled, err := LoadSchemasFromDir(dir)
	if err != nil {
		return nil, err
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

func (v *Validator) validateBytes(schemaName string, data []byte) (ValidationResult, error) {
	schema, ok := v.schemas[schemaName]
	if !ok {
		return ValidationResult{}, fmt.Errorf("schema not loaded: %s", schemaName)
	}

	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return ValidationResult{}, fmt.Errorf("parse json: %w", err)
	}

	if err := schema.Validate(value); err != nil {
		issues := collectIssues(err)
		if len(issues) > 0 {
			return ValidationResult{Issues: issues}, nil
		}
		return ValidationResult{}, err
	}

	return ValidationResult{}, nil
}

func collectIssues(err error) []ValidationIssue {
	validationErr, ok := err.(*jsonschema.ValidationError)
	if !ok {
		return nil
	}
	var issues []ValidationIssue
	flattenIssues(&issues, validationErr)
	return issues
}

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
