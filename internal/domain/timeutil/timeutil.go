package timeutil

import "time"

var now = time.Now

// FormatISO8601 は DD-DATA-002 の日時表記に従い、TZ 付き秒精度で整形する。
func FormatISO8601(value time.Time) string {
	return value.In(time.Local).Truncate(time.Second).Format(time.RFC3339)
}

// NowISO8601 は DD-DATA-002 の日時表記で現在時刻を返す。
func NowISO8601() string {
	return FormatISO8601(now())
}
