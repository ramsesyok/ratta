// Package configrepo は config.json の読み書きを担い、UI表示や暗号処理は扱わない。
// 具体的な保存形式は jsonfmt に委ねる。
package configrepo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"ratta/internal/infra/atomicwrite"
	"ratta/internal/infra/jsonfmt"
)

const (
	formatVersion   = 1
	defaultPageSize = 20
)

// Config は DD-DATA-001 の config.json 仕様を表す。
type Config struct {
	FormatVersion       int    `json:"format_version"`
	LastProjectRootPath string `json:"last_project_root_path"`
	Log                 Log    `json:"log"`
	UI                  UI     `json:"ui"`
}

// Log は DD-DATA-001 の log 設定を表す。
type Log struct {
	Level string `json:"level"`
}

// UI は DD-DATA-001 の UI 設定を表す。
type UI struct {
	PageSize int `json:"page_size"`
}

// DefaultConfig は DD-DATA-001 の既定値に従う。
func DefaultConfig() Config {
	return Config{
		FormatVersion:       formatVersion,
		LastProjectRootPath: "",
		Log: Log{
			Level: "info",
		},
		UI: UI{
			PageSize: defaultPageSize,
		},
	}
}

// Repository は DD-BE-002 の config.json 読み書きを担う。
type Repository struct {
	path string
}

var writeFile = atomicwrite.WriteFile

// NewRepository は DD-BE-002 に従い、実行ファイルと同じディレクトリの config.json を扱う。
func NewRepository(exePath string) *Repository {
	return &Repository{
		path: filepath.Join(filepath.Dir(exePath), "config.json"),
	}
}

// Load は DD-BE-002 に従い config.json を読み込み、存在しなければ既定値を返す。
// 目的: 設定を読み取り、存在しない場合は既定値で続行する。
// 入力: なし。
// 出力: Config、存在フラグ、エラー。
// エラー: 読み取り・パース失敗時に返す。
// 副作用: config.json を読み取る。
// 並行性: 読み取りのみでスレッドセーフ。
// 不変条件: 返却する Config は format_version を含む。
// 関連DD: DD-BE-002
func (r *Repository) Load() (Config, bool, error) {
	data, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return DefaultConfig(), false, nil
	}
	if err != nil {
		return DefaultConfig(), false, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if unmarshalErr := json.Unmarshal(data, &cfg); unmarshalErr != nil {
		return DefaultConfig(), false, fmt.Errorf("parse config: %w", unmarshalErr)
	}

	return cfg, true, nil
}

// Save は DD-PERSIST-002 に従い config.json を atomic write で保存する。
func (r *Repository) Save(cfg Config) error {
	data, err := jsonfmt.MarshalConfig(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if writeErr := writeFile(r.path, data); writeErr != nil {
		return fmt.Errorf("write config: %w", writeErr)
	}
	return nil
}

// SaveLastProjectRoot は DD-BE-003 に従い last_project_root_path を更新して保存する。
// 目的: 最終利用したプロジェクトルートを保存する。
// 入力: path は保存するパス。
// 出力: 成功時は nil、失敗時はエラー。
// エラー: 読み込みや保存失敗時に返す。
// 副作用: config.json を更新する。
// 並行性: 同時更新は想定しない。
// 不変条件: last_project_root_path のみ変更し他の設定は保持する。
// 関連DD: DD-BE-003
func (r *Repository) SaveLastProjectRoot(path string) error {
	cfg, _, err := r.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.LastProjectRootPath = path
	if saveErr := r.Save(cfg); saveErr != nil {
		return fmt.Errorf("save config: %w", saveErr)
	}
	return nil
}
