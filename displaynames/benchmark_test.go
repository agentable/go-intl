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

func BenchmarkDisplayNames_OfCompoundLanguage_Cached(b *testing.B) {
	for _, tc := range []struct {
		display displaynames.LanguageDisplay
		code    string
		want    string
	}{
		{displaynames.StandardLanguageDisplay, "sl-Latn-SI-rozaj", "Slovenian (Latin, Slovenia, Resian)"},
		{displaynames.DialectLanguageDisplay, "zh-Hans-CN", "Simplified Chinese (China)"},
	} {
		b.Run(string(tc.display), func(b *testing.B) {
			dn := benchmarkDisplayNames(b, displaynames.Options{
				Type:            stringPtr(displaynames.Language),
				LanguageDisplay: stringPtr(tc.display),
			})
			b.ReportAllocs()
			for b.Loop() {
				if got, ok, err := dn.Of(tc.code); err != nil || !ok || got != tc.want {
					b.Fatalf("Of(%s) = %q, %v, %v; want %q, true, nil", tc.code, got, ok, err, tc.want)
				}
			}
		})
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
