package durationformat

import (
	"testing"

	"github.com/agentable/go-intl/locale"
)

func BenchmarkDurationFormat_Short_PerCall(b *testing.B) {
	locales := locale.List{locale.MustParse("en-US")}
	duration := benchmarkDuration()

	b.ReportAllocs()
	for b.Loop() {
		format, err := New(locales, Options{})
		if err != nil {
			b.Fatal(err)
		}
		if _, err := format.Format(duration); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDurationFormat_Short_Cached(b *testing.B) {
	format := benchmarkDurationFormat(b, Options{})
	duration := benchmarkDuration()

	b.ReportAllocs()
	for b.Loop() {
		if _, err := format.Format(duration); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDurationFormat_Digital_Cached(b *testing.B) {
	format := benchmarkDurationFormat(b, Options{Style: DigitalStyle})
	duration := Duration{Hours: 12, Minutes: 3, Seconds: 4, Milliseconds: 567}

	b.ReportAllocs()
	for b.Loop() {
		if _, err := format.Format(duration); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkDurationFormat(b *testing.B, opts Options) *DurationFormat {
	b.Helper()

	format, err := New(locale.List{locale.MustParse("en-US")}, opts)
	if err != nil {
		b.Fatal(err)
	}
	return format
}

func benchmarkDuration() Duration {
	return Duration{Years: 1, Months: 2, Weeks: 3, Days: 4, Hours: 5, Minutes: 6, Seconds: 7, Milliseconds: 8}
}
