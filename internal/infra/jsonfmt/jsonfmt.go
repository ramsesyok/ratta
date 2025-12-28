package jsonfmt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const indent = "  "

// MarshalCanonical は DD-DATA-001 のデータ設計に合わせ、
// プロジェクト標準のインデントと LF 改行で JSON を出力する。
func MarshalCanonical(value any) ([]byte, error) {
	return marshalWithOrder(value, nil)
}

// MarshalIssue は DD-DATA-003/004/005 のキー順に従って issue JSON を整形する。
func MarshalIssue(value any) ([]byte, error) {
	return marshalWithOrder(value, issueKeyOrder)
}

// MarshalConfig は DD-DATA-001 のキー順に従って config JSON を整形する。
func MarshalConfig(value any) ([]byte, error) {
	return marshalWithOrder(value, configKeyOrder)
}

// MarshalContractor は DD-DATA-001 のキー順に従って contractor JSON を整形する。
func MarshalContractor(value any) ([]byte, error) {
	return marshalWithOrder(value, contractorKeyOrder)
}

type keyOrder struct {
	Order    []string
	Children map[string]*keyOrder
}

// issueKeyOrder は DD-DATA-003/004/005 のキー順を定義する。
var issueKeyOrder = &keyOrder{
	Order: []string{
		"version",
		"issue_id",
		"category",
		"title",
		"description",
		"status",
		"priority",
		"origin_company",
		"assignee",
		"created_at",
		"updated_at",
		"due_date",
		"comments",
	},
	Children: map[string]*keyOrder{
		"comments": {
			Order: []string{
				"comment_id",
				"body",
				"author_name",
				"author_company",
				"created_at",
				"attachments",
			},
			Children: map[string]*keyOrder{
				"attachments": {
					Order: []string{
						"attachment_id",
						"file_name",
						"stored_name",
						"relative_path",
						"mime_type",
						"size_bytes",
					},
				},
			},
		},
	},
}

// configKeyOrder は DD-DATA-001 のキー順を定義する。
var configKeyOrder = &keyOrder{
	Order: []string{
		"format_version",
		"last_project_root_path",
		"log",
		"ui",
	},
	Children: map[string]*keyOrder{
		"log": {Order: []string{"level"}},
		"ui":  {Order: []string{"page_size"}},
	},
}

// contractorKeyOrder は DD-DATA-001 のキー順を定義する。
var contractorKeyOrder = &keyOrder{
	Order: []string{
		"format_version",
		"kdf",
		"kdf_iterations",
		"salt_b64",
		"nonce_b64",
		"ciphertext_b64",
		"mode",
	},
}

// marshalWithOrder は DD-DATA-001 の canonical 出力ルールに従って整形する。
func marshalWithOrder(value any, order *keyOrder) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var data any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := writeValue(&buf, data, order, 0); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

// writeValue は DD-DATA-001 の JSON ルールに従い値を出力する。
func writeValue(buf *bytes.Buffer, value any, order *keyOrder, level int) error {
	switch typed := value.(type) {
	case map[string]any:
		return writeObject(buf, typed, order, level)
	case []any:
		return writeArray(buf, typed, order, level)
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return err
		}
		buf.Write(encoded)
		return nil
	}
}

// writeObject は DD-DATA-001 のキー順でオブジェクトを出力する。
func writeObject(buf *bytes.Buffer, value map[string]any, order *keyOrder, level int) error {
	if len(value) == 0 {
		buf.WriteString("{}")
		return nil
	}

	buf.WriteString("{\n")
	keys := orderedKeys(value, order)
	for i, key := range keys {
		buf.WriteString(strings.Repeat(indent, level+1))
		buf.WriteString(fmt.Sprintf("%q", key))
		buf.WriteString(": ")
		childOrder := orderChild(order, key)
		if err := writeValue(buf, value[key], childOrder, level+1); err != nil {
			return err
		}
		if i < len(keys)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}
	buf.WriteString(strings.Repeat(indent, level))
	buf.WriteString("}")
	return nil
}

// writeArray は DD-DATA-001 の配列表記で出力する。
func writeArray(buf *bytes.Buffer, value []any, order *keyOrder, level int) error {
	if len(value) == 0 {
		buf.WriteString("[]")
		return nil
	}
	buf.WriteString("[\n")
	for i, item := range value {
		buf.WriteString(strings.Repeat(indent, level+1))
		if err := writeValue(buf, item, order, level+1); err != nil {
			return err
		}
		if i < len(value)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}
	buf.WriteString(strings.Repeat(indent, level))
	buf.WriteString("]")
	return nil
}

// orderedKeys は DD-DATA-001 のキー順と未知キーのソートを適用する。
func orderedKeys(value map[string]any, order *keyOrder) []string {
	seen := make(map[string]struct{}, len(value))
	var keys []string
	if order != nil {
		for _, key := range order.Order {
			if _, ok := value[key]; ok {
				keys = append(keys, key)
				seen[key] = struct{}{}
			}
		}
	}
	var remaining []string
	for key := range value {
		if _, ok := seen[key]; ok {
			continue
		}
		remaining = append(remaining, key)
	}
	sort.Strings(remaining)
	keys = append(keys, remaining...)
	return keys
}

// orderChild は DD-DATA-001 のネスト順序定義を取得する。
func orderChild(order *keyOrder, key string) *keyOrder {
	if order == nil {
		return nil
	}
	return order.Children[key]
}
