package durationformat

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func BenchmarkDurationFormat_Short_PerCall(b *testing.B) {
	locales := locale.List{intltest.Locale(b, "en-US")}
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
	format := benchmarkDurationFormat(b, Options{Style: stringPtr(DigitalStyle)})
	duration := Duration{Hours: 12, Minutes: 3, Seconds: 4, Milliseconds: 567}

	b.ReportAllocs()
	for b.Loop() {
		if _, err := format.Format(duration); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDurationFormat_TextSeconds_Cached(b *testing.B) {
	format := benchmarkDurationFormat(b, Options{Seconds: new("long"), Milliseconds: new("numeric")})
	duration := Duration{Seconds: -1, Milliseconds: -473}

	b.ReportAllocs()
	for b.Loop() {
		if got, err := format.Format(duration); err != nil || got != "-1.473 seconds" {
			b.Fatalf("Format() = %q, %v; want -1.473 seconds", got, err)
		}
	}
}

func BenchmarkDurationFormat_UnitPrograms_New(b *testing.B) {
	for _, tc := range []struct {
		name string
		opts Options
	}{
		{"digital", Options{Style: new("digital")}},
		{"numeric-seconds", Options{Seconds: new("numeric")}},
		{"two-digit-seconds", Options{Seconds: new("2-digit")}},
		{"fixed-fraction", Options{Style: new("digital"), FractionalDigits: new(3)}},
		{"text-seconds", Options{Seconds: new("long"), Milliseconds: new("numeric")}},
		{"text-milliseconds", Options{Milliseconds: new("long"), Microseconds: new("numeric")}},
		{"text-microseconds", Options{Microseconds: new("long"), Nanoseconds: new("numeric")}},
		{"text-units", Options{Style: new("long")}},
	} {
		b.Run(tc.name, func(b *testing.B) {
			locales := intltest.LocaleList(b, "en-US")
			if _, err := New(locales, tc.opts); err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			for b.Loop() {
				if _, err := New(locales, tc.opts); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func benchmarkDurationFormat(b *testing.B, opts Options) *DurationFormat {
	b.Helper()

	format, err := New(locale.List{intltest.Locale(b, "en-US")}, opts)
	if err != nil {
		b.Fatal(err)
	}
	return format
}

func benchmarkDuration() Duration {
	return Duration{Years: 1, Months: 2, Weeks: 3, Days: 4, Hours: 5, Minutes: 6, Seconds: 7, Milliseconds: 8}
}
