package localematcher

import (
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

func TestBestFitMatcher(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		requested []string
		supported []string
		def       string
		locale    string
		extension string
	}{
		{name: "lookup-compatible truncation", requested: []string{"fr-XX", "en"}, supported: []string{"fr", "en"}, def: "en", locale: "fr"},
		{name: "zh traditional fallback", requested: []string{"zh-TW"}, supported: []string{"zh", "zh-Hant"}, def: "en", locale: "zh-Hant"},
		{name: "exact preferred", requested: []string{"en"}, supported: []string{"en", "und"}, def: "en", locale: "en"},
		{name: "preserves extension", requested: []string{"th-u-ca-gregory"}, supported: []string{"th"}, def: "en", locale: "th", extension: "-u-ca-gregory"},
		{name: "falls back when too distant", requested: []string{"es"}, supported: []string{"fr", "en"}, def: "en", locale: "en"},
		{name: "prefers later close locale", requested: []string{"de-DE", "fr"}, supported: []string{"en", "en-US", "fr-FR"}, def: "en-US", locale: "fr-FR"},
		{name: "prefers exact before language family", requested: []string{"en-GB", "en-US", "en"}, supported: []string{"en-US", "nl-NL", "nl"}, def: "en-US", locale: "en-US"},
		{name: "cnsar regional match", requested: []string{"zh-HK"}, supported: []string{"zh-Hant", "zh-MO"}, def: "en", locale: "zh-MO"},
		{name: "enUS regional match", requested: []string{"en-CA"}, supported: []string{"en-GB", "en-US"}, def: "en-US", locale: "en-US"},
		{name: "americas regional match", requested: []string{"es-KY"}, supported: []string{"es", "en", "es-419"}, def: "en", locale: "es-419"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := BestFitMatcherWithMaximizer(tc.requested, tc.supported, tc.def, cldrlocale.Maximize)
			if got.Locale != tc.locale || got.Extension != tc.extension {
				t.Fatalf("BestFitMatcher() = %#v, want locale %q extension %q", got, tc.locale, tc.extension)
			}
		})
	}
}

func TestBestFitMatcherUsesMaximizedFallbackCandidates(t *testing.T) {
	t.Parallel()

	maximizer := func(tag string) string {
		switch tag {
		case "zh-TW", "zh-Hant":
			return "zh-Hant-TW"
		default:
			return tag
		}
	}
	got := BestFitMatcherWithMaximizer([]string{"zh-TW"}, []string{"zh-Hant", "en"}, "en", maximizer)
	if got.Locale != "zh-Hant" || got.DataLocale != "zh-Hant" {
		t.Fatalf("BestFitMatcherWithMaximizer() = %#v, want zh-Hant fallback candidate", got)
	}
	if got.Distance != 0 {
		t.Fatalf("BestFitMatcherWithMaximizer().Distance = %d, want exact maximized distance", got.Distance)
	}
}
