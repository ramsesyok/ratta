// Package jsonfmt は JSON の整形とキー順序制御を担い、保存先のI/Oは扱わない。
// フォーマット仕様は詳細設計に従う。
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
// 目的: キー順序を固定せずに標準整形を適用する。
// 入力: value は任意のJSON化対象。
// 出力: 整形済みJSONバイト列とエラー。
// エラー: JSON変換に失敗した場合に返す。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: 改行はLF、インデントは2スペース。
// 関連DD: DD-DATA-001
func MarshalCanonical(value any) ([]byte, error) {
	return marshalWithOrder(value, nil)
}

// MarshalIssue は DD-DATA-003/004/005 のキー順に従って issue JSON を整形する。
// 目的: 課題JSONのキー順を固定し差分を安定化する。
// 入力: value は課題の構造体またはマップ。
// 出力: 整形済みJSONバイト列とエラー。
// エラー: JSON変換に失敗した場合に返す。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: 仕様定義のキー順序を維持する。
// 関連DD: DD-DATA-003, DD-DATA-004, DD-DATA-005
func MarshalIssue(value any) ([]byte, error) {
	return marshalWithOrder(value, issueKeyOrder)
}

// MarshalConfig は DD-DATA-001 のキー順に従って config JSON を整形する。
// 目的: config.json のキー順を固定し差分を安定化する。
// 入力: value は設定構造体またはマップ。
// 出力: 整形済みJSONバイト列とエラー。
// エラー: JSON変換に失敗した場合に返す。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: 仕様定義のキー順序を維持する。
// 関連DD: DD-DATA-001
func MarshalConfig(value any) ([]byte, error) {
	return marshalWithOrder(value, configKeyOrder)
}

// MarshalContractor は DD-DATA-001 のキー順に従って contractor JSON を整形する。
// 目的: contractor.json のキー順を固定し差分を安定化する。
// 入力: value は認証構造体またはマップ。
// 出力: 整形済みJSONバイト列とエラー。
// エラー: JSON変換に失敗した場合に返す。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: 仕様定義のキー順序を維持する。
// 関連DD: DD-DATA-001
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
// 目的: JSONを一度汎用構造に変換し、順序付きで再出力する。
// 入力: value はJSON化対象、order はキー順序定義。
// 出力: 整形済みJSONバイト列とエラー。
// エラー: JSON変換や整形処理に失敗した場合に返す。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: 出力の末尾に改行を付与する。
// 関連DD: DD-DATA-001
func marshalWithOrder(value any, order *keyOrder) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}

	var data any
	if unmarshalErr := json.Unmarshal(raw, &data); unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshal json: %w", unmarshalErr)
	}

	var buf bytes.Buffer
	if writeErr := writeValue(&buf, data, order, 0); writeErr != nil {
		return nil, writeErr
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

// writeValue は DD-DATA-001 の JSON ルールに従い値を出力する。
// 目的: 値の型に応じて正しい表現で書き出す。
// 入力: buf は出力先、value は対象値、order はキー順序定義、level はインデント階層。
// 出力: 成功時は nil、失敗時はエラー。
// エラー: JSON変換に失敗した場合に返す。
// 副作用: buf に書き込む。
// 並行性: buf は呼び出し側で排他する。
// 不変条件: 文字列は JSON エスケープ済みで出力する。
// 関連DD: DD-DATA-001
func writeValue(buf *bytes.Buffer, value any, order *keyOrder, level int) error {
	switch typed := value.(type) {
	case map[string]any:
		return writeObject(buf, typed, order, level)
	case []any:
		return writeArray(buf, typed, order, level)
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return fmt.Errorf("marshal value: %w", err)
		}
		buf.Write(encoded)
		return nil
	}
}

// writeObject は DD-DATA-001 のキー順でオブジェクトを出力する。
// 目的: キー順序定義に従いオブジェクトを整形出力する。
// 入力: buf は出力先、value はマップ、order はキー順序定義、level はインデント階層。
// 出力: 成功時は nil、失敗時はエラー。
// エラー: 値の出力に失敗した場合に返す。
// 副作用: buf に書き込む。
// 並行性: buf は呼び出し側で排他する。
// 不変条件: 既知キーは order の順序で出力する。
// 関連DD: DD-DATA-001
func writeObject(buf *bytes.Buffer, value map[string]any, order *keyOrder, level int) error {
	if len(value) == 0 {
		buf.WriteString("{}")
		return nil
	}

	buf.WriteString("{\n")
	keys := orderedKeys(value, order)
	for i, key := range keys {
		buf.WriteString(strings.Repeat(indent, level+1))
		fmt.Fprintf(buf, "%q", key)
		buf.WriteString(": ")
		childOrder := orderChild(order, key)
		if writeErr := writeValue(buf, value[key], childOrder, level+1); writeErr != nil {
			return writeErr
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
// 目的: 配列要素を正しいインデントで出力する。
// 入力: buf は出力先、value は配列、order は子要素順序、level はインデント階層。
// 出力: 成功時は nil、失敗時はエラー。
// エラー: 要素出力に失敗した場合に返す。
// 副作用: buf に書き込む。
// 並行性: buf は呼び出し側で排他する。
// 不変条件: 要素間はカンマ区切りで出力する。
// 関連DD: DD-DATA-001
func writeArray(buf *bytes.Buffer, value []any, order *keyOrder, level int) error {
	if len(value) == 0 {
		buf.WriteString("[]")
		return nil
	}
	buf.WriteString("[\n")
	for i, item := range value {
		buf.WriteString(strings.Repeat(indent, level+1))
		if writeErr := writeValue(buf, item, order, level+1); writeErr != nil {
			return writeErr
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
// 目的: 定義済みキー順序と未定義キーの辞書順を統合する。
// 入力: value は対象マップ、order はキー順序定義。
// 出力: 反映済みのキー配列。
// エラー: なし。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: 未定義キーは昇順で追加される。
// 関連DD: DD-DATA-001
func orderedKeys(value map[string]any, order *keyOrder) []string {
	seen := make(map[string]struct{}, len(value))
	keys := make([]string, 0, len(value))
	if order != nil {
		for _, key := range order.Order {
			if _, ok := value[key]; ok {
				keys = append(keys, key)
				seen[key] = struct{}{}
			}
		}
	}
	remaining := make([]string, 0, len(value))
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
