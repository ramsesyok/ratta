package contractorinit

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"ratta/internal/infra/atomicwrite"
	"ratta/internal/infra/crypto"
	"ratta/internal/infra/jsonfmt"
)

var (
	generateAuth = crypto.GenerateContractorAuth
	marshalAuth  = jsonfmt.MarshalContractor
	writeFile    = atomicwrite.WriteFile
	statFile     = os.Stat
	mkdirAll     = os.MkdirAll
)

// Prompter は DD-CLI-003 のパスワード入力を抽象化する。
type Prompter interface {
	PromptHidden(label string) (string, error)
}

// Run は DD-CLI-002/003/004 に従い contractor.json を生成する。
func Run(exePath string, force bool, prompter Prompter) error {
	if prompter == nil {
		return errors.New("prompter is required")
	}

	password, err := prompter.PromptHidden("Password: ")
	if err != nil {
		return fmt.Errorf("prompt password: %w", err)
	}
	confirm, err := prompter.PromptHidden("Confirm: ")
	if err != nil {
		return fmt.Errorf("prompt confirm: %w", err)
	}
	if password == "" {
		return errors.New("password is required")
	}
	if password != confirm {
		return errors.New("password confirmation does not match")
	}

	authDir := filepath.Join(filepath.Dir(exePath), "auth")
	targetPath := filepath.Join(authDir, "contractor.json")

	if exists, err := fileExists(targetPath); err != nil {
		return err
	} else if exists && !force {
		return fmt.Errorf("contractor.json already exists")
	}

	if err := mkdirAll(authDir, 0o755); err != nil {
		return fmt.Errorf("create auth dir: %w", err)
	}

	auth, err := generateAuth(password)
	if err != nil {
		return fmt.Errorf("generate contractor auth: %w", err)
	}
	data, err := marshalAuth(auth)
	if err != nil {
		return fmt.Errorf("marshal contractor auth: %w", err)
	}
	if err := writeFile(targetPath, data); err != nil {
		return fmt.Errorf("write contractor auth: %w", err)
	}
	return nil
}

func fileExists(path string) (bool, error) {
	_, err := statFile(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("stat contractor auth: %w", err)
}
