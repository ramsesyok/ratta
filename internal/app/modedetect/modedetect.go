// Package modedetect は起動時モード判定とパスワード検証を担い、UI 表示は扱わない。
// 認証情報の暗号処理は infra 層に委ねる。
package modedetect

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"ratta/internal/domain/mode"
	"ratta/internal/infra/crypto"
	"ratta/internal/infra/schema"
)

var (
	readFile = os.ReadFile
	statFile = os.Stat
)

// Service は DD-BE-003 のモード判定と検証を担う。
type Service struct {
	authPath  string
	validator *schema.Validator
}

// NewService は DD-BE-003 に従い auth/contractor.json を対象にする。
func NewService(exePath string, validator *schema.Validator) *Service {
	return &Service{
		authPath:  filepath.Join(filepath.Dir(exePath), "auth", "contractor.json"),
		validator: validator,
	}
}

// DetectMode は DD-BE-003 の起動時モード判定を行う。
func (s *Service) DetectMode() (mode.Mode, bool, error) {
	exists, err := s.fileExists()
	if err != nil {
		return mode.ModeVendor, false, err
	}
	if !exists {
		return mode.ModeVendor, false, nil
	}
	return mode.ModeVendor, true, nil
}

// VerifyContractorPassword は DD-BE-003/DD-CLI-005 に従いパスワードを検証する。
// 目的: contractor.json の内容に基づきパスワード一致を判定する。
// 入力: password は入力された平文パスワード。
// 出力: 成功時は ModeContractor、失敗時は ModeVendor とエラー。
// エラー: 読み取り・検証・復号失敗、パスワード不一致時に返す。
// 副作用: contractor.json を読み取る。
// 並行性: 読み取りのみでスレッドセーフ。
// 不変条件: 認証情報が不正な場合は Contractor モードにしない。
// 関連DD: DD-BE-003, DD-CLI-005
func (s *Service) VerifyContractorPassword(password string) (mode.Mode, error) {
	data, err := readFile(s.authPath)
	if err != nil {
		return mode.ModeVendor, fmt.Errorf("read contractor auth: %w", err)
	}
	if s.validator != nil {
		result, validateErr := s.validator.ValidateContractor(data)
		if validateErr != nil {
			return mode.ModeVendor, fmt.Errorf("validate contractor auth: %w", validateErr)
		}
		if len(result.Issues) > 0 {
			return mode.ModeVendor, fmt.Errorf("contractor auth schema invalid: %s", result.Detail())
		}
	}

	var auth crypto.ContractorAuth
	if unmarshalErr := json.Unmarshal(data, &auth); unmarshalErr != nil {
		return mode.ModeVendor, fmt.Errorf("parse contractor auth: %w", unmarshalErr)
	}
	ok, err := crypto.VerifyPassword(auth, password)
	if err != nil {
		if errors.Is(err, crypto.ErrPasswordMismatch) {
			return mode.ModeVendor, errors.New("password verification failed")
		}
		return mode.ModeVendor, fmt.Errorf("verify contractor password: %w", err)
	}
	if !ok {
		return mode.ModeVendor, errors.New("password verification failed")
	}
	return mode.ModeContractor, nil
}

func (s *Service) fileExists() (bool, error) {
	_, err := statFile(s.authPath)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("stat contractor auth: %w", err)
}
