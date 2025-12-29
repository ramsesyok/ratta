// schema.go はスキーマ読み込みと参照制御を担い、検証結果の整形は扱わない。
package schema

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// LoadSchemasFromDir は DD-BE-002 に従いディレクトリ内の JSON Schema をコンパイルし、
// 外部参照は拒否する。
// 目的: スキーマファイルを読み込み内部参照のみ許可する。
// 入力: dir はスキーマディレクトリ。
// 出力: スキーマ名とコンパイル済みスキーマのマップ、エラー。
// エラー: 読み込み・コンパイル失敗時に返す。
// 副作用: スキーマファイルを読み取る。
// 並行性: 読み取りのみでスレッドセーフ。
// 不変条件: 外部参照は拒否する。
// 関連DD: DD-BE-002
func LoadSchemasFromDir(dir string) (map[string]*jsonschema.Schema, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve schema dir: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	compiler.LoadURL = func(ref string) (io.ReadCloser, error) {
		parsed, parseErr := url.Parse(ref)
		if parseErr != nil {
			return nil, fmt.Errorf("parse schema ref: %w", parseErr)
		}
		switch parsed.Scheme {
		case "http", "https":
			return nil, fmt.Errorf("external schema refs are not allowed: %s", ref)
		case "file":
			return openSchemaFile(absDir, parsed.Path)
		case "":
			return openSchemaFile(absDir, parsed.Path)
		default:
			return nil, fmt.Errorf("unsupported schema ref: %s", ref)
		}
	}

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, fmt.Errorf("read schema dir: %w", err)
	}

	compiled := make(map[string]*jsonschema.Schema)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(absDir, entry.Name())
		compiledSchema, compileErr := compiler.Compile(path)
		if compileErr != nil {
			return nil, fmt.Errorf("compile schema %s: %w", entry.Name(), compileErr)
		}
		compiled[entry.Name()] = compiledSchema
	}

	return compiled, nil
}

// openSchemaFile は DD-BE-002 のローカル限定ルールを満たすファイルを開く。
// 目的: スキーマ参照が許可された範囲内であることを保証して開く。
// 入力: baseDir は許可された基準ディレクトリ、path は参照パス。
// 出力: ReadCloser とエラー。
// エラー: 参照範囲外やファイルオープン失敗時に返す。
// 副作用: スキーマファイルを開く。
// 並行性: 読み取りのみでスレッドセーフ。
// 不変条件: baseDir 外部は拒否する。
// 関連DD: DD-BE-002
func openSchemaFile(baseDir, path string) (io.ReadCloser, error) {
	cleaned := filepath.Clean(path)
	if !filepath.IsAbs(cleaned) {
		cleaned = filepath.Join(baseDir, cleaned)
	}
	if !strings.HasPrefix(cleaned, baseDir+string(os.PathSeparator)) && cleaned != baseDir {
		return nil, fmt.Errorf("schema ref outside schema dir: %s", path)
	}
	file, err := os.Open(cleaned)
	if err != nil {
		return nil, fmt.Errorf("open schema: %w", err)
	}
	return file, nil
}
