package listformat

import (
	"strconv"
	"strings"
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

func BenchmarkListFormat_Format_Scale(b *testing.B) {
	for _, size := range []int{10, 100, 1000} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			format := benchmarkListFormat(b, Options{})
			list := strings.Split(strings.Repeat("Alpha,", size-1)+"Alpha", ",")
			want := strings.Repeat("Alpha, ", size-1) + "and Alpha"
			var got string
			b.ReportAllocs()
			for b.Loop() {
				got = format.Format(list)
			}
			if got != want {
				b.Fatalf("Format returned incorrect output for %d elements", size)
			}
		})
	}
}

func BenchmarkListFormat_FormatToParts_Scale(b *testing.B) {
	for _, size := range []int{10, 100, 1000} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			format := benchmarkListFormat(b, Options{})
			list := strings.Split(strings.Repeat("Alpha,", size-1)+"Alpha", ",")
			var got []Part
			b.ReportAllocs()
			for b.Loop() {
				got = format.FormatToParts(list)
			}
			if len(got) != 2*size-1 || got[0] != (Part{Type: PartElement, Value: "Alpha"}) || got[len(got)-1] != got[0] {
				b.Fatalf("FormatToParts returned incorrect output for %d elements", size)
			}
		})
	}
}
