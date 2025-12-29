// Package timeutil は時刻表現の共通処理を提供し、永続化I/Oは扱わない。
package timeutil

import "time"

// now は DD-DATA-002 の時刻仕様をテストで固定するための差し替え点。
var now = time.Now

// FormatISO8601 は DD-DATA-002 の日時表記に従い、TZ 付き秒精度で整形する。
func FormatISO8601(value time.Time) string {
	return value.In(time.Local).Truncate(time.Second).Format(time.RFC3339)
}

// NowISO8601 は DD-DATA-002 の日時表記で現在時刻を返す。
func NowISO8601() string {
	return FormatISO8601(now())
}
