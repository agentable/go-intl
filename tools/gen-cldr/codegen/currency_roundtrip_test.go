package codegen

import (
	"testing"

	"github.com/agentable/go-intl/internal/cldr/currency"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// TestCurrencyRoundTrip is the production-path round-trip gate for the currency
// domain. It re-derives the extract.CurrencyData maps from the real pinned CLDR
// checkout and asserts that every fraction record, every per-locale display
// name / canonical name / symbol, and every supported code the encoder wrote is
// queried back byte-for-byte through the production currency accessors over the
// committed data.go. It exercises encoder, blob, decoder, and accessor as one
// chain — not internal structures.
//
// The gate is meaningful only when the pinned cldr-json checkout is present
// (after task data / data:fetch); without it the test skips, since there is no
// data to round-trip against.
func TestCurrencyRoundTrip(t *testing.T) {
	t.Parallel()

	input := loadRoundTripSource(t)
	data := extract.ExtractCurrencies(input.source.CurrencyFractions, input.source.Currencies, input.profile)

	// Fraction table: every code, including the DEFAULT entry the accessor falls
	// back to, must read back exactly. Digits maps a missing code onto DEFAULT,
	// so a present code must always equal its own record.
	for code, f := range data.Fractions {
		want := currency.Data{DefaultDigits: f.Digits, CashDigits: f.CashDigits, Rounding: f.Rounding}
		if got := currency.Digits(code); got != want {
			t.Errorf("Digits(%q) = %+v, want %+v", code, got, want)
		}
	}

	// Per-locale names: display map per plural, canonical name, and symbol.
	for localeTag, currencies := range data.Currencies {
		loc := resolveKernelLocale(t, localeTag)
		for code, names := range currencies {
			for plural, want := range names.Display {
				if got := currency.DisplayName(loc, code, plural); got != want {
					t.Errorf("DisplayName(%q, %q, %q) = %q, want %q", localeTag, code, plural, got, want)
				}
			}
			if names.Symbol != "" {
				if got := currency.Symbol(loc, code); got != names.Symbol {
					t.Errorf("Symbol(%q, %q) = %q, want %q", localeTag, code, got, names.Symbol)
				}
			}
			if names.Narrow != "" {
				if got := currency.NarrowSymbol(loc, code); got != names.Narrow {
					t.Errorf("NarrowSymbol(%q, %q) = %q, want %q", localeTag, code, got, names.Narrow)
				}
			}
		}
	}

	// Supported codes narrow index.
	wantCodes := supportedCurrencyValues(data)
	gotCodes := currency.SupportedCodes()
	assertStringSliceEqual(t, "SupportedCodes", gotCodes, wantCodes)
}
