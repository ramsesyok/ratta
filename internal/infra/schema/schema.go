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

// LoadSchemasFromDir はディレクトリ内の JSON Schema をコンパイルし、
// 外部参照は拒否する。
func LoadSchemasFromDir(dir string) (map[string]*jsonschema.Schema, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve schema dir: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	compiler.LoadURL = func(ref string) (io.ReadCloser, error) {
		parsed, err := url.Parse(ref)
		if err != nil {
			return nil, fmt.Errorf("parse schema ref: %w", err)
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
		schema, err := compiler.Compile(path)
		if err != nil {
			return nil, fmt.Errorf("compile schema %s: %w", entry.Name(), err)
		}
		compiled[entry.Name()] = schema
	}

	return compiled, nil
}

func openSchemaFile(baseDir, path string) (io.ReadCloser, error) {
	cleaned := filepath.Clean(path)
	if !filepath.IsAbs(cleaned) {
		cleaned = filepath.Join(baseDir, cleaned)
	}
	if !strings.HasPrefix(cleaned, baseDir+string(os.PathSeparator)) && cleaned != baseDir {
		return nil, fmt.Errorf("schema ref outside schema dir: %s", path)
	}
	return os.Open(cleaned)
}
