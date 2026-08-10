package datetimeformat

import (
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func BenchmarkDateTimeFormat_DateStyleShort_PerCall(b *testing.B) {
	loc := intltest.Locale(b, "en-US")
	date := benchmarkDate()

	b.ReportAllocs()
	for b.Loop() {
		format, err := New(locale.List{loc}, benchmarkDateOptions())
		if err != nil {
			b.Fatal(err)
		}
		_ = format.Format(date)
	}
}

func BenchmarkDateTimeFormat_DateStyleShort_Cached(b *testing.B) {
	format := benchmarkDateTimeFormat(b)
	date := benchmarkDate()

	b.ReportAllocs()
	for b.Loop() {
		_ = format.Format(date)
	}
}

func BenchmarkDateTimeFormat_DateTimeRange_Cached(b *testing.B) {
	format := benchmarkDateTimeFormat(b)
	start := benchmarkDate()
	end := start.AddDate(0, 0, 2)

	b.ReportAllocs()
	for b.Loop() {
		_ = format.FormatRange(start, end)
	}
}

func BenchmarkDateTimeFormat_FormatToParts_Cached(b *testing.B) {
	format := benchmarkDateTimeFormat(b)
	date := benchmarkDate()

	b.ReportAllocs()
	for b.Loop() {
		_ = format.FormatToParts(date)
	}
}

func BenchmarkDateTimeFormat_New(b *testing.B) {
	loc := intltest.Locale(b, "en-US")

	b.ReportAllocs()
	for b.Loop() {
		format, err := New(locale.List{loc}, benchmarkDateOptions())
		if err != nil {
			b.Fatal(err)
		}
		_ = format
	}
}

func benchmarkDateTimeFormat(b *testing.B) *DateTimeFormat {
	b.Helper()

	format, err := New(locale.List{intltest.Locale(b, "en-US")}, benchmarkDateOptions())
	if err != nil {
		b.Fatal(err)
	}
	return format
}

func benchmarkDateOptions() Options {
	return Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)}
}

func benchmarkDate() time.Time {
	return time.Date(2026, time.May, 8, 12, 0, 0, 0, time.UTC)
}
