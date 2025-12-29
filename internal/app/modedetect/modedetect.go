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
func (s *Service) VerifyContractorPassword(password string) (mode.Mode, error) {
	data, err := readFile(s.authPath)
	if err != nil {
		return mode.ModeVendor, fmt.Errorf("read contractor auth: %w", err)
	}
	if s.validator != nil {
		result, err := s.validator.ValidateContractor(data)
		if err != nil {
			return mode.ModeVendor, fmt.Errorf("validate contractor auth: %w", err)
		}
		if len(result.Issues) > 0 {
			return mode.ModeVendor, fmt.Errorf("contractor auth schema invalid: %s", result.Detail())
		}
	}

	var auth crypto.ContractorAuth
	if err := json.Unmarshal(data, &auth); err != nil {
		return mode.ModeVendor, fmt.Errorf("parse contractor auth: %w", err)
	}
	ok, err := crypto.VerifyPassword(auth, password)
	if err != nil {
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
