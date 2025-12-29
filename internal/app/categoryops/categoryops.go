package categoryops

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ratta/internal/domain/issue"
	mod "ratta/internal/domain/mode"
	"ratta/internal/infra/atomicwrite"
	"ratta/internal/infra/jsonfmt"
)

// Category は DD-LOAD-002 のカテゴリ情報を表す。
type Category struct {
	Name       string
	IsReadOnly bool
	Path       string
}

// Service は DD-BE-003 のカテゴリ操作を担う。
type Service struct {
	projectRoot string
}

// NewService は DD-BE-003 のカテゴリ操作に必要な設定を受け取って生成する。
func NewService(projectRoot string) *Service {
	return &Service{projectRoot: projectRoot}
}

// CreateCategory は DD-BE-003 のカテゴリ作成を行う。
func (s *Service) CreateCategory(name string, currentMode mod.Mode) (Category, error) {
	if currentMode != mod.ModeContractor {
		return Category{}, errors.New("permission denied")
	}
	if errs := issue.ValidateCategoryName(name); len(errs) > 0 {
		return Category{}, errs
	}
	if err := s.ensureNoConflict(name); err != nil {
		return Category{}, err
	}
	path := filepath.Join(s.projectRoot, name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return Category{}, fmt.Errorf("create category: %w", err)
	}
	return Category{Name: name, Path: path}, nil
}

// DeleteCategory は DD-BE-003 のカテゴリ削除を行う。
func (s *Service) DeleteCategory(name string, currentMode mod.Mode) error {
	if currentMode != mod.ModeContractor {
		return errors.New("permission denied")
	}
	if s.isReadOnly(name) {
		return errors.New("read-only category")
	}
	path := filepath.Join(s.projectRoot, name)
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("read category: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), ".files") {
			continue
		}
		if entry.IsDir() {
			return errors.New("category not empty")
		}
		if filepath.Ext(entry.Name()) == ".json" {
			return errors.New("category not empty")
		}
	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	return nil
}

// RenameCategory は DD-BE-003 のカテゴリ名変更を行う。
func (s *Service) RenameCategory(oldName, newName string, currentMode mod.Mode) (Category, error) {
	if currentMode != mod.ModeContractor {
		return Category{}, errors.New("permission denied")
	}
	if errs := issue.ValidateCategoryName(newName); len(errs) > 0 {
		return Category{}, errs
	}
	if err := s.ensureNoConflict(newName); err != nil {
		return Category{}, err
	}
	if s.hasTmpRenameResidue() {
		return Category{}, errors.New("tmp_rename residue exists")
	}
	oldPath := filepath.Join(s.projectRoot, oldName)
	if _, err := os.Stat(oldPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Category{}, errors.New("category not found")
		}
		return Category{}, fmt.Errorf("stat category: %w", err)
	}

	tmpRoot := filepath.Join(s.projectRoot, ".tmp_rename")
	tmpPath := filepath.Join(tmpRoot, newName)
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		return Category{}, fmt.Errorf("create tmp_rename: %w", err)
	}
	if err := os.Rename(oldPath, tmpPath); err != nil {
		return Category{}, fmt.Errorf("rename category: %w", err)
	}

	if err := s.updateIssueCategory(tmpPath, newName); err != nil {
		_ = os.Rename(tmpPath, oldPath)
		return Category{}, err
	}

	finalPath := filepath.Join(s.projectRoot, newName)
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return Category{}, fmt.Errorf("rename category final: %w", err)
	}
	return Category{Name: newName, Path: finalPath}, nil
}

// ensureNoConflict は DD-BE-003 の大小文字違いを含む重複を防ぐ。
func (s *Service) ensureNoConflict(name string) error {
	entries, err := os.ReadDir(s.projectRoot)
	if err != nil {
		return fmt.Errorf("read project root: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		other := entry.Name()
		if strings.EqualFold(other, name) {
			return errors.New("category name conflict")
		}
	}
	return nil
}

// hasTmpRenameResidue は DD-BE-003 の .tmp_rename 残骸検出を行う。
func (s *Service) hasTmpRenameResidue() bool {
	tmpPath := filepath.Join(s.projectRoot, ".tmp_rename")
	entries, err := os.ReadDir(tmpPath)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return true
		}
	}
	return false
}

// isReadOnly は DD-LOAD-002 の読み取り専用カテゴリ判定を行う。
func (s *Service) isReadOnly(name string) bool {
	path := filepath.Join(s.projectRoot, ".tmp_rename", name)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// updateIssueCategory は DD-BE-003 のカテゴリ名変更に伴う課題更新を行う。
func (s *Service) updateIssueCategory(categoryPath, newName string) error {
	entries, err := os.ReadDir(categoryPath)
	if err != nil {
		return fmt.Errorf("read category: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(categoryPath, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read issue: %w", err)
		}
		var parsed issue.Issue
		if err := json.Unmarshal(data, &parsed); err != nil {
			return fmt.Errorf("parse issue: %w", err)
		}
		parsed.Category = newName
		updated, err := jsonfmt.MarshalIssue(parsed)
		if err != nil {
			return fmt.Errorf("marshal issue: %w", err)
		}
		if err := atomicwrite.WriteFile(path, updated); err != nil {
			return fmt.Errorf("write issue: %w", err)
		}
	}
	return nil
}
