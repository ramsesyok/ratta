// Package categoryops はカテゴリ操作のユースケースを提供し、ファイルI/Oの詳細は扱わない。
// UI 表示やドメイン検証の内部実装には踏み込まない。
package categoryops

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"ratta/internal/domain/issue"
	"ratta/internal/infra/atomicwrite"
	"ratta/internal/infra/jsonfmt"
	"strings"

	mod "ratta/internal/domain/mode"
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
// 目的: 課題カテゴリ用のディレクトリを作成し識別情報を返す。
// 入力: name はカテゴリ名、currentMode は操作モード。
// 出力: 作成した Category とエラー。
// エラー: 権限不足、カテゴリ名検証失敗、同名衝突、作成失敗時に返す。
// 副作用: プロジェクトルート配下にディレクトリを作成する。
// 並行性: 同一プロジェクトルートへの同時実行は呼び出し側で排他する。
// 不変条件: 作成後のカテゴリ名は入力 name と一致する。
// 関連DD: DD-BE-003
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
	if err := os.MkdirAll(path, 0o750); err != nil {
		return Category{}, fmt.Errorf("create category: %w", err)
	}
	return Category{Name: name, Path: path}, nil
}

// DeleteCategory は DD-BE-003 のカテゴリ削除を行う。
// 目的: 空のカテゴリディレクトリを削除する。
// 入力: name はカテゴリ名、currentMode は操作モード。
// 出力: 成功時は nil、失敗時はエラー。
// エラー: 権限不足、読み取り専用、非空、削除失敗時に返す。
// 副作用: カテゴリディレクトリを削除する。
// 並行性: 同時削除は想定しない。
// 不変条件: 削除対象は .json と .files を含まないことを確認する。
// 関連DD: DD-BE-003
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
	removeErr := os.RemoveAll(path)
	if removeErr != nil {
		return fmt.Errorf("delete category: %w", removeErr)
	}
	return nil
}

// RenameCategory は DD-BE-003 のカテゴリ名変更を行う。
// 目的: カテゴリ名変更に伴いディレクトリと課題JSONを更新する。
// 入力: oldName は旧カテゴリ名、newName は新カテゴリ名、currentMode は操作モード。
// 出力: 更新後の Category とエラー。
// エラー: 権限不足、検証失敗、衝突、リネーム失敗時に返す。
// 副作用: ディレクトリ移動と課題JSONの書き換えを行う。
// 並行性: 同時更新は想定しない。
// 不変条件: 更新後の課題JSONの Category は newName。
// 関連DD: DD-BE-003
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
	if err := os.MkdirAll(tmpRoot, 0o750); err != nil {
		return Category{}, fmt.Errorf("create tmp_rename: %w", err)
	}
	if err := os.Rename(oldPath, tmpPath); err != nil {
		return Category{}, fmt.Errorf("rename category: %w", err)
	}

	if err := s.updateIssueCategory(tmpPath, newName); err != nil {
		if renameErr := os.Rename(tmpPath, oldPath); renameErr != nil {
			return Category{}, fmt.Errorf("rollback rename failed: %w; rollback error: %s", err, renameErr.Error())
		}
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
// 目的: カテゴリ配下の課題JSONに新カテゴリ名を反映する。
// 入力: categoryPath は変更対象のカテゴリパス、newName は新カテゴリ名。
// 出力: 成功時は nil、失敗時はエラー。
// エラー: 読み取り・パース・書き込み失敗時に返す。
// 副作用: 課題JSONを書き換える。
// 並行性: 同時書き込みは想定しない。
// 不変条件: 対象JSONの Category フィールドは newName に統一する。
// 関連DD: DD-BE-003
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
		// #nosec G304 -- カテゴリ配下の列挙結果のみを利用するため安全。
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read issue: %w", readErr)
		}
		var parsed issue.Issue
		if unmarshalErr := json.Unmarshal(data, &parsed); unmarshalErr != nil {
			return fmt.Errorf("parse issue: %w", unmarshalErr)
		}
		parsed.Category = newName
		updated, marshalErr := jsonfmt.MarshalIssue(parsed)
		if marshalErr != nil {
			return fmt.Errorf("marshal issue: %w", marshalErr)
		}
		if writeErr := atomicwrite.WriteFile(path, updated); writeErr != nil {
			return fmt.Errorf("write issue: %w", writeErr)
		}
	}
	return nil
}
