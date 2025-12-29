package projectroot

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ValidationResult は DD-BE-003 の ValidationResultDTO に合わせた結果を表す。
type ValidationResult struct {
	IsValid        bool
	NormalizedPath string
	Message        string
	Details        string
}

// Service は DD-BE-003 の Project Root 操作を担う。
type Service struct {
	configRepo ConfigSaver
}

// ConfigSaver は DD-BE-003 の config 保存を抽象化する。
type ConfigSaver interface {
	SaveLastProjectRoot(path string) error
}

// NewService は DD-BE-003 の config 保存先を受け取って作成する。
func NewService(configRepo ConfigSaver) *Service {
	return &Service{configRepo: configRepo}
}

// ValidateProjectRoot は DD-BE-003 の Project Root 検証を行う。
func (s *Service) ValidateProjectRoot(path string) (ValidationResult, error) {
	if path == "" {
		return ValidationResult{
			IsValid: false,
			Message: "Path is required.",
		}, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ValidationResult{
				IsValid: false,
				Message: "Path does not exist.",
				Details: err.Error(),
			}, nil
		}
		return ValidationResult{}, fmt.Errorf("stat project root: %w", err)
	}
	if !info.IsDir() {
		return ValidationResult{
			IsValid: false,
			Message: "Path is not a directory.",
		}, nil
	}

	normalized, err := filepath.Abs(path)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("normalize path: %w", err)
	}

	return ValidationResult{
		IsValid:        true,
		NormalizedPath: normalized,
		Message:        "OK",
	}, nil
}

// CreateProjectRoot は DD-BE-003 の Project Root 作成を行う。
func (s *Service) CreateProjectRoot(path string) (ValidationResult, error) {
	if path == "" {
		return ValidationResult{
			IsValid: false,
			Message: "Path is required.",
		}, nil
	}

	if _, err := os.Stat(path); err == nil {
		return ValidationResult{
			IsValid: false,
			Message: "Path already exists.",
		}, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return ValidationResult{}, fmt.Errorf("stat project root: %w", err)
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		return ValidationResult{}, fmt.Errorf("create project root: %w", err)
	}

	normalized, err := filepath.Abs(path)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("normalize path: %w", err)
	}

	return ValidationResult{
		IsValid:        true,
		NormalizedPath: normalized,
		Message:        "OK",
	}, nil
}

// SaveLastProjectRoot は DD-BE-003 の last_project_root_path 更新を行う。
func (s *Service) SaveLastProjectRoot(path string) error {
	if s.configRepo == nil {
		return errors.New("config repository is required")
	}
	return s.configRepo.SaveLastProjectRoot(path)
}
