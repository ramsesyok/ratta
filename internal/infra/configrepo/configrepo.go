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

const formatVersion = 1
const defaultPageSize = 20

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
// 読み込み失敗時は既定値と警告エラーを返す。
func (r *Repository) Load() (Config, bool, error) {
	data, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return DefaultConfig(), false, nil
	}
	if err != nil {
		return DefaultConfig(), false, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), false, fmt.Errorf("parse config: %w", err)
	}

	return cfg, true, nil
}

// Save は DD-PERSIST-002 に従い config.json を atomic write で保存する。
func (r *Repository) Save(cfg Config) error {
	data, err := jsonfmt.MarshalConfig(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := writeFile(r.path, data); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// SaveLastProjectRoot は DD-BE-003 に従い last_project_root_path を更新して保存する。
func (r *Repository) SaveLastProjectRoot(path string) error {
	cfg, _, err := r.Load()
	if err != nil {
		return err
	}
	cfg.LastProjectRootPath = path
	return r.Save(cfg)
}
