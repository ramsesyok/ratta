package jsonfmt

import (
	"bytes"
	"encoding/json"
)

const indent = "  "

// MarshalCanonical はプロジェクト標準のインデントと LF 改行で JSON を出力する。
// キー順は TASK-0205 で対応する。
func MarshalCanonical(value any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", indent)

	if err := encoder.Encode(value); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
