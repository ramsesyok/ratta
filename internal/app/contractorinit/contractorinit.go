// Package contractorinit は contractor.json 生成のユースケースを提供し、UIや通信は扱わない。
// 暗号化の詳細実装は infra 層に委ねる。
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
// 目的: Contractor 認証情報ファイルを生成し所定の配置に保存する。
// 入力: exePath は実行ファイルのパス、force は上書き許可、prompter は入力手段。
// 出力: 成功時は nil、失敗時はエラー。
// エラー: 入力不備、既存ファイル衝突、暗号化や保存失敗時に返す。
// 副作用: auth ディレクトリ作成と contractor.json 書き込みを行う。
// 並行性: 同一パスへの同時実行は想定しない。
// 不変条件: 保存する JSON は暗号化済みパスワードを含む。
// 関連DD: DD-CLI-002, DD-CLI-003, DD-CLI-004
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

	if exists, existsErr := fileExists(targetPath); existsErr != nil {
		return existsErr
	} else if exists && !force {
		return errors.New("contractor.json already exists")
	}

	if mkdirErr := mkdirAll(authDir, 0o750); mkdirErr != nil {
		return fmt.Errorf("create auth dir: %w", mkdirErr)
	}

	auth, err := generateAuth(password)
	if err != nil {
		return fmt.Errorf("generate contractor auth: %w", err)
	}
	data, err := marshalAuth(auth)
	if err != nil {
		return fmt.Errorf("marshal contractor auth: %w", err)
	}
	if writeErr := writeFile(targetPath, data); writeErr != nil {
		return fmt.Errorf("write contractor auth: %w", writeErr)
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
