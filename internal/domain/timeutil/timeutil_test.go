package timeutil

import (
	"testing"
	"time"
)

func TestFormatISO8601_SecondPrecisionWithTZ(t *testing.T) {
	// DD-DATA-002 の ISO 8601 仕様に従い、秒精度と TZ が付与されることを確認する。
	location := time.FixedZone("JST", 9*60*60)
	value := time.Date(2024, 1, 2, 3, 4, 5, 123456789, location)

	got := FormatISO8601(value)
	if got != "2024-01-02T03:04:05+09:00" {
		t.Fatalf("unexpected format: %s", got)
	}
}

func TestNowISO8601_UsesCurrentTimeFormat(t *testing.T) {
	// 現在時刻の取得でも秒精度の ISO 8601 形式になることを確認する。
	previous := now
	nowValue := time.Date(2024, 2, 3, 4, 5, 6, 987654321, time.Local)
	now = func() time.Time { return nowValue }
	t.Cleanup(func() { now = previous })

	got := NowISO8601()
	expected := FormatISO8601(nowValue)
	if got != expected {
		t.Fatalf("unexpected format: %s", got)
	}
}
