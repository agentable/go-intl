package codegen

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentable/go-intl/internal/cldr/currency"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
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

	repoRoot := filepath.Clean("../../..")
	cldrDir := filepath.Join(repoRoot, "tools", "gen-cldr", ".cldr-json", "node_modules")
	if _, err := os.Stat(cldrDir); err != nil {
		t.Skipf("pinned cldr-json checkout absent (%v); run task data:fetch", err)
	}

	versions, err := cldr.ReadVersionFile(filepath.Join(repoRoot, "internal", "cldr", "VERSION"))
	if err != nil {
		t.Fatalf("read VERSION: %v", err)
	}
	profile := readUnitTestProfile(t, filepath.Join(repoRoot, "tools", "locale-profile.json"))

	source, err := cldr.LoadAll(context.Background(), cldrDir, versions, profile)
	if err != nil {
		t.Fatalf("load cldr-json: %v", err)
	}

	data := extract.ExtractCurrencies(source.CurrencyFractions, source.Currencies, profile)

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
		loc := resolveCurrencyLocale(t, localeTag)
		for code, names := range currencies {
			for plural, want := range names.Display {
				if got := currency.DisplayName(loc, code, plural); got != want {
					t.Errorf("DisplayName(%q, %q, %q) = %q, want %q", localeTag, code, plural, got, want)
				}
			}
			if names.Symbol != "" {
				if got := currency.Symbol(loc, code, "symbol"); got != names.Symbol {
					t.Errorf("Symbol(%q, %q, symbol) = %q, want %q", localeTag, code, got, names.Symbol)
				}
			}
			if names.Narrow != "" {
				if got := currency.Symbol(loc, code, "narrow"); got != names.Narrow {
					t.Errorf("Symbol(%q, %q, narrow) = %q, want %q", localeTag, code, got, names.Narrow)
				}
			}
		}
	}

	// Supported codes narrow index.
	wantCodes := supportedCurrencyValues(data)
	gotCodes := currency.SupportedCodes()
	if len(gotCodes) != len(wantCodes) {
		t.Fatalf("SupportedCodes len = %d, want %d", len(gotCodes), len(wantCodes))
	}
	for i := range wantCodes {
		if gotCodes[i] != wantCodes[i] {
			t.Errorf("SupportedCodes[%d] = %q, want %q", i, gotCodes[i], wantCodes[i])
		}
	}
}

// resolveCurrencyLocale resolves a tag to the kernel handle the currency
// accessors take. The names map is keyed by the kernel locale index, so a tag
// that fails to resolve would silently mis-key every name lookup.
func resolveCurrencyLocale(t *testing.T, tag string) cldrlocale.Locale {
	t.Helper()
	loc, ok := cldrlocale.ResolveLocale(tag)
	if !ok {
		t.Fatalf("kernel locale %q not resolvable", tag)
	}
	return loc
}
