package codegen

import (
	"testing"

	"github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// TestNumberRoundTrip is the production-path round-trip gate for the number
// domain. It re-derives the extract.Numbers map from the real pinned CLDR
// checkout and asserts that every per-locale default numbering system, symbol
// set, decimal/percent/scientific pattern, currency style, and compact pattern
// the encoder wrote is queried back byte-for-byte through the production number
// accessors over the committed data.go. It exercises encoder, blob, decoder,
// and accessor as one chain — not internal structures.
//
// The gate is meaningful only when the pinned cldr-json checkout is present
// (after task data / data:fetch); without it the test skips, since there is no
// data to round-trip against.
func TestNumberRoundTrip(t *testing.T) {
	t.Parallel()

	input := loadRoundTripSource(t)
	data := extract.ExtractNumbers(input.source.Numbers, input.profile)

	for localeTag, numbers := range data {
		loc := resolveNumberLocale(t, localeTag)

		if got := loc.DefaultNumberingSystem(); got != numbers.DefaultNumberingSystem {
			t.Errorf("DefaultNumberingSystem(%q) = %q, want %q", localeTag, got, numbers.DefaultNumberingSystem)
		}

		for ns, want := range numbers.Symbols {
			if got := loc.NumberSymbols(ns); got != number.NumberSymbols(want) {
				t.Errorf("NumberSymbols(%q, %q) = %+v, want %+v", localeTag, ns, got, want)
			}
		}
		for ns, want := range numbers.DecimalPatterns {
			if got := loc.DecimalPattern(ns); got != want {
				t.Errorf("DecimalPattern(%q, %q) = %q, want %q", localeTag, ns, got, want)
			}
		}
		for ns, want := range numbers.PercentPatterns {
			if got := loc.PercentPattern(ns); got != want {
				t.Errorf("PercentPattern(%q, %q) = %q, want %q", localeTag, ns, got, want)
			}
		}
		for ns, want := range numbers.ScientificPatterns {
			if got := loc.ScientificPattern(ns); got != want {
				t.Errorf("ScientificPattern(%q, %q) = %q, want %q", localeTag, ns, got, want)
			}
		}
		for ns, signs := range numbers.CurrencyPatterns {
			for sign, want := range signs {
				if got := loc.CurrencyPattern(ns, sign); got != want {
					t.Errorf("CurrencyPattern(%q, %q, %q) = %q, want %q", localeTag, ns, sign, got, want)
				}
			}
		}
		for ns, displays := range numbers.CompactPatterns {
			for display, exps := range displays {
				for exp, plurals := range exps {
					for plural, want := range plurals {
						if got := loc.CompactPattern(ns, display, exp, plural); got != want {
							t.Errorf("CompactPattern(%q, %q, %q, %d, %q) = %q, want %q", localeTag, ns, display, exp, plural, got, want)
						}
					}
				}
			}
		}
	}

	// Supported locales narrow index: the encoder wrote exactly the locales with
	// number data, in sorted-locale order.
	wantTags := sortedLocaleKeys(data)
	gotTags := number.SupportedLocales()
	assertStringSliceEqual(t, "SupportedLocales", gotTags, wantTags)
}

// resolveNumberLocale resolves a tag to the number-domain handle the accessors
// take. The number data is keyed by the kernel locale index, so a tag that
// fails to resolve would silently mis-key every lookup.
func resolveNumberLocale(t *testing.T, tag string) number.Locale {
	t.Helper()
	loc, ok := number.ResolveLocale(tag)
	if !ok {
		t.Fatalf("number locale %q not resolvable", tag)
	}
	return loc
}
