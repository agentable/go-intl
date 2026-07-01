package displaynames_test

import (
	"testing"

	"github.com/agentable/go-intl/displaynames"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func BenchmarkDisplayNames_OfLanguage_Cached(b *testing.B) {
	dn := benchmarkDisplayNames(b, displaynames.Options{Type: stringPtr(displaynames.Language)})

	b.ReportAllocs()
	for b.Loop() {
		got, ok, err := dn.Of("fr")
		if err != nil || !ok || got != "French" {
			b.Fatalf("Of(fr) = %q, %v, %v; want French, true, nil", got, ok, err)
		}
	}
}

func BenchmarkDisplayNames_OfFallback_Cached(b *testing.B) {
	dn := benchmarkDisplayNames(b, displaynames.Options{Type: stringPtr(displaynames.Region)})

	b.ReportAllocs()
	for b.Loop() {
		got, ok, err := dn.Of("qq")
		if err != nil || !ok || got != "QQ" {
			b.Fatalf("Of(qq) = %q, %v, %v; want QQ, true, nil", got, ok, err)
		}
	}
}

func benchmarkDisplayNames(b *testing.B, opts displaynames.Options) *displaynames.DisplayNames {
	b.Helper()

	dn, err := displaynames.New(locale.List{intltest.Locale(b, "en")}, opts)
	if err != nil {
		b.Fatal(err)
	}
	return dn
}
