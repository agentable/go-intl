package listformat

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func BenchmarkListFormat_Format_Cached(b *testing.B) {
	format := benchmarkListFormat(b, Options{})
	list := []string{"Alpha", "Beta", "Gamma", "Delta"}

	b.ReportAllocs()
	for b.Loop() {
		if got := format.Format(list); got != "Alpha, Beta, Gamma, and Delta" {
			b.Fatalf("Format(%v) = %q", list, got)
		}
	}
}

func BenchmarkListFormat_FormatToParts_Cached(b *testing.B) {
	format := benchmarkListFormat(b, Options{})
	list := []string{"Alpha", "Beta", "Gamma", "Delta"}

	b.ReportAllocs()
	for b.Loop() {
		parts := format.FormatToParts(list)
		if len(parts) != 7 || parts[0].Value != "Alpha" || parts[len(parts)-1].Value != "Delta" {
			b.Fatalf("FormatToParts(%v) = %#v", list, parts)
		}
	}
}

func benchmarkListFormat(b *testing.B, opts Options) *ListFormat {
	b.Helper()

	format, err := New(locale.List{intltest.Locale(b, "en-US")}, opts)
	if err != nil {
		b.Fatal(err)
	}
	return format
}
