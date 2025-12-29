// Package mode は操作モードの定義を提供し、権限判定の実装は別ファイルで扱う。
package mode

// Mode は DD-BE-003 のモード種別を表す。
type Mode string

const (
	ModeContractor Mode = "Contractor"
	ModeVendor     Mode = "Vendor"
)
