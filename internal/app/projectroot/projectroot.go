// Package projectroot はプロジェクトルートの検証・作成を担い、UI 表示は扱わない。
// 設定保存の詳細は infra 層に委ねる。
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
// 目的: 入力パスが有効なプロジェクトルートかを判定する。
// 入力: path は検証対象のパス。
// 出力: ValidationResult とエラー。
// エラー: 取得失敗や正規化失敗時に返す。
// 副作用: なし。
// 並行性: 読み取りのみでスレッドセーフ。
// 不変条件: 結果は IsValid に応じた Message を持つ。
// 関連DD: DD-BE-003
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
// 目的: プロジェクトルートディレクトリを作成し正規化パスを返す。
// 入力: path は作成対象のパス。
// 出力: ValidationResult とエラー。
// エラー: 既存や作成失敗、正規化失敗時に返す。
// 副作用: ディレクトリを作成する。
// 並行性: 同一パスへの同時作成は想定しない。
// 不変条件: 作成成功時は IsValid=true。
// 関連DD: DD-BE-003
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

	if err := os.MkdirAll(path, 0o750); err != nil {
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
// 目的: 最終選択されたプロジェクトルートを保存する。
// 入力: path は保存するパス。
// 出力: 成功時は nil、失敗時はエラー。
// エラー: リポジトリ未設定や保存失敗時に返す。
// 副作用: 設定ファイルを書き換える。
// 並行性: 同時保存は想定しない。
// 不変条件: path は空文字も保存可能とする。
// 関連DD: DD-BE-003
func (s *Service) SaveLastProjectRoot(path string) error {
	if s.configRepo == nil {
		return errors.New("config repository is required")
	}
	if err := s.configRepo.SaveLastProjectRoot(path); err != nil {
		return fmt.Errorf("save last project root: %w", err)
	}
	return nil
}
