package jsonfmt

import (
	"bytes"
	"encoding/json"
)

const indent = "  "

// MarshalCanonical encodes JSON with the project's base indentation and LF
// newlines. Key ordering is handled later in TASK-0205.
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
