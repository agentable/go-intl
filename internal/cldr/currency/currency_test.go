package currency

import (
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/internal/testprocess"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no fraction or names decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_CURRENCY_NARROW_INDEX_SUBPROCESS"

// TestSupportedCodesDoesNotDecodeOtherBlobs asserts the narrow-index rule:
// SupportedCodes reads only the supported blob and must never trigger the
// fraction or names blob decode.
//
// The assertion runs in a fresh process so other currency-data tests cannot
// populate the package-level Once state first.
func TestSupportedCodesDoesNotDecodeOtherBlobs(t *testing.T) {
	t.Parallel()

	if !testprocess.RunInFreshProcess(t, narrowIndexSubprocessEnv) {
		return
	}
	testcontract.AssertNarrowStringIndexDoesNotLoad(t, "SupportedCodes", SupportedCodes,
		testcontract.LoadProbe{Name: "fraction blob", Loaded: func() bool { return fractions != nil }},
		testcontract.LoadProbe{Name: "names blob", Loaded: func() bool { return byLocale != nil }},
	)
}

func TestSupportedCodesReturnsCopy(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceReturnsCopy(t, "SupportedCodes", SupportedCodes)
}

func TestSupportedCodesSortedAndUnique(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceSortedUnique(t, "SupportedCodes", SupportedCodes())
}

// TestSmokeKnownFractions is a checkout-independent smoke test: a few known ISO
// codes resolve to the fraction metadata recorded in the committed data.go,
// including the DEFAULT fallback for an unknown code. These values are
// intentionally hard-coded so a silent encoder/decoder regression fails here
// even when the FormatJS fixtures are unavailable.
func TestSmokeKnownFractions(t *testing.T) {
	t.Parallel()

	if got, want := Digits("USD"), (Data{DefaultDigits: 2, CashDigits: 2, Rounding: 0}); got != want {
		t.Errorf("Digits(USD) = %+v, want %+v", got, want)
	}
	if got, want := Digits("JPY"), (Data{DefaultDigits: 0, CashDigits: 0, Rounding: 0}); got != want {
		t.Errorf("Digits(JPY) = %+v, want %+v", got, want)
	}
	unknown := Digits("XXX")
	if unknown.DefaultDigits != 2 || unknown.CashDigits != 2 {
		t.Errorf("Digits(XXX) = %+v, want DEFAULT 2 digits", unknown)
	}
}

// TestSmokeKnownNames is a checkout-independent smoke test for the per-locale
// display name and symbol accessors, resolved through the kernel "en-US" handle.
func TestSmokeKnownNames(t *testing.T) {
	t.Parallel()

	loc, ok := cldrlocale.ResolveLocale("en-US")
	if !ok {
		t.Fatal(`ResolveLocale("en-US") = false, want true`)
	}

	if got, want := DisplayName(loc, "USD", "one"), "US dollar"; got != want {
		t.Errorf("DisplayName(USD, one) = %q, want %q", got, want)
	}
	other := DisplayName(loc, "USD", "other")
	if other == "" {
		t.Fatal("DisplayName(USD, other) = empty, want CLDR plural fallback anchor")
	}
	if got := DisplayName(loc, "USD", ""); got != other {
		t.Errorf("DisplayName(USD, empty plural) = %q, want other form %q", got, other)
	}
	if got := DisplayName(loc, "USD", "unknown"); got != other {
		t.Errorf("DisplayName(USD, unknown plural) = %q, want other form %q", got, other)
	}
	if got, want := CanonicalName(loc, "USD"), "US Dollar"; got != want {
		t.Errorf("CanonicalName(USD) = %q, want %q", got, want)
	}
	if got, want := Symbol(loc, "USD"), "$"; got != want {
		t.Errorf("Symbol(USD) = %q, want %q", got, want)
	}
	if got, want := NarrowSymbol(loc, "JPY"), "¥"; got != want {
		t.Errorf("NarrowSymbol(JPY) = %q, want %q", got, want)
	}
	if got, want := NarrowSymbol(loc, "USD"), Symbol(loc, "USD"); got != want {
		t.Errorf("NarrowSymbol(USD) = %q, want standard symbol %q", got, want)
	}

	assertMissingCurrencyNames(t, cldrlocale.Undefined, "USD")
	assertMissingCurrencyNames(t, loc, "ZZZ")
}

func TestLocaleNamesMissingCodeReturnsZeroRecord(t *testing.T) {
	t.Parallel()

	loc, ok := cldrlocale.ResolveLocale("en-US")
	if !ok {
		t.Fatal(`ResolveLocale("en-US") = false, want true`)
	}

	names := localeNames(loc, "ZZZ")
	if names.display != nil || names.canonical != "" || names.symbol != "" || names.narrow != "" {
		t.Fatalf("localeNames(en-US, ZZZ) = %+v, want zero currencyNames", names)
	}
}

func assertMissingCurrencyNames(t *testing.T, loc Locale, code string) {
	t.Helper()

	if got := DisplayName(loc, code, "other"); got != "" {
		t.Errorf("DisplayName(%v, %s, other) = %q, want empty", loc, code, got)
	}
	if got := CanonicalName(loc, code); got != "" {
		t.Errorf("CanonicalName(%v, %s) = %q, want empty", loc, code, got)
	}
	if got := Symbol(loc, code); got != "" {
		t.Errorf("Symbol(%v, %s) = %q, want empty", loc, code, got)
	}
	if got := NarrowSymbol(loc, code); got != "" {
		t.Errorf("NarrowSymbol(%v, %s) = %q, want empty", loc, code, got)
	}
}
