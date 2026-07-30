package number

import (
	"maps"
	"slices"
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/numbering"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/internal/testprocess"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no number-data decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_NUMBER_NARROW_INDEX_SUBPROCESS"

// TestSupportedLocalesDoesNotDecodeNumberBlob asserts the narrow-index rule:
// SupportedLocales reads only the supported blob and must never trigger the
// main number-data decode.
//
// The assertion runs in a fresh process so other number-data tests cannot
// populate the package-level Once state first.
func TestSupportedLocalesDoesNotDecodeNumberBlob(t *testing.T) {
	t.Parallel()

	if !testprocess.RunInFreshProcess(t, narrowIndexSubprocessEnv) {
		return
	}
	testcontract.AssertNarrowStringIndexDoesNotLoad(t, "SupportedLocales", SupportedLocales,
		testcontract.LoadProbe{Name: "number blob", Loaded: func() bool { return byLocale != nil }},
	)
}

func TestSupportedNumberingSystemsDoesNotDecodeNumberBlob(t *testing.T) {
	t.Parallel()

	if !testprocess.RunInFreshProcess(t, narrowIndexSubprocessEnv) {
		return
	}
	testcontract.AssertNarrowStringIndexDoesNotLoad(t, "SupportedNumberingSystems", SupportedNumberingSystems,
		testcontract.LoadProbe{Name: "number blob", Loaded: func() bool { return byLocale != nil }},
	)
}

// TestSupportedNumberingSystemsComesFromGeneratedIndex asserts the numbering
// system supported values are the sorted union of the ECMA-402 simple numbering
// systems and the numbering systems carried by the generated number data, and
// that the result is sorted and unique.
func TestSupportedNumberingSystemsComesFromGeneratedIndex(t *testing.T) {
	t.Parallel()

	got := SupportedNumberingSystems()
	if len(got) == 0 {
		t.Fatal("SupportedNumberingSystems returned no values")
	}
	testcontract.AssertStringSliceSortedUnique(t, "SupportedNumberingSystems", got)

	seen := map[string]bool{}
	for _, ns := range numbering.SimpleNumberingSystems() {
		seen[ns] = true
	}
	for _, ns := range decodeNumberingSystemExtras() {
		seen[ns] = true
	}
	if want := slices.Sorted(maps.Keys(seen)); !slices.Equal(got, want) {
		t.Fatalf("SupportedNumberingSystems = %v, want %v", got, want)
	}
}

func TestGeneratedNumberingSystemExtrasHaveRuntimePayload(t *testing.T) {
	t.Parallel()

	numberingSystemExtras := decodeNumberingSystemExtras()
	if len(numberingSystemExtras) == 0 {
		t.Fatal("generated numbering-system extras are empty; supported numbering systems must reflect CLDR number symbols")
	}
	data := numberDataByLocale()
	for _, numberingSystem := range numberingSystemExtras {
		if numberingSystem == "" {
			t.Fatal("generated numbering-system extras contain empty identifier")
		}
		if !numberingSystemHasRuntimePayload(data, numberingSystem) {
			t.Fatalf("generated numbering-system extra %q has no runtime symbols and decimal pattern", numberingSystem)
		}
	}
}

func TestRuntimeNumberingSystemPayloadsAreAdvertised(t *testing.T) {
	t.Parallel()

	supported := map[string]bool{}
	for _, numberingSystem := range SupportedNumberingSystems() {
		supported[numberingSystem] = true
	}
	for _, data := range numberDataByLocale() {
		for numberingSystem := range data.symbols {
			if !supported[numberingSystem] {
				t.Fatalf("runtime number symbols for %q are not advertised by SupportedNumberingSystems", numberingSystem)
			}
		}
		for numberingSystem := range data.decimal {
			if !supported[numberingSystem] {
				t.Fatalf("runtime decimal pattern for %q is not advertised by SupportedNumberingSystems", numberingSystem)
			}
		}
	}
}

func numberingSystemHasRuntimePayload(data map[Locale]numberData, numberingSystem string) bool {
	for _, localeData := range data {
		symbols := localeData.symbols[numberingSystem]
		if symbols.Decimal == "" || symbols.Group == "" || symbols.Plus == "" || symbols.Minus == "" {
			continue
		}
		if localeData.decimal[numberingSystem] == "" {
			continue
		}
		return true
	}
	return false
}

// TestSupportedNumberingSystemsReturnsCopy asserts the merged result is a fresh
// slice each call so callers cannot corrupt cached state.
func TestSupportedNumberingSystemsReturnsCopy(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceReturnsCopy(t, "SupportedNumberingSystems", SupportedNumberingSystems)
}

func TestSupportedLocalesReturnsCopy(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceReturnsCopy(t, "SupportedLocales", SupportedLocales)
}

// TestSmokeKnownNumberData is a checkout-independent smoke test: it asserts the
// number accessors resolved through the kernel "en" handle return the strings
// recorded from the committed data.go. These values are intentionally
// hard-coded so a silent encoder/decoder regression fails here even when the
// FormatJS fixtures are unavailable.
func TestSmokeKnownNumberData(t *testing.T) {
	t.Parallel()

	loc, ok := ResolveLocale("en")
	if !ok {
		t.Fatal(`ResolveLocale("en") = false, want true`)
	}

	if got, want := loc.DefaultNumberingSystem(), "latn"; got != want {
		t.Fatalf("DefaultNumberingSystem = %q, want %q", got, want)
	}
	symbols := loc.NumberSymbols("latn")
	if symbols.Decimal != "." || symbols.Group != "," || symbols.Percent != "%" {
		t.Fatalf("NumberSymbols = %+v", symbols)
	}
	if got, want := symbols.TimeSeparator, ":"; got != want {
		t.Fatalf("NumberSymbols TimeSeparator = %q, want %q", got, want)
	}
	if got := loc.NumberSymbols("missing-symbol-row"); got != symbols {
		t.Fatalf("NumberSymbols(missing) = %+v, want default %+v", got, symbols)
	}
	if got, want := loc.DecimalPattern("latn"), "#,##0.###"; got != want {
		t.Fatalf("DecimalPattern = %q, want %q", got, want)
	}
	if got, want := loc.DecimalPattern("missing-numbering-system"), "#,##0.###"; got != want {
		t.Fatalf("DecimalPattern(missing numbering system) = %q, want fallback %q", got, want)
	}
	if got, want := loc.PercentPattern("latn"), "#,##0%"; got != want {
		t.Fatalf("PercentPattern = %q, want %q", got, want)
	}
	if got, want := loc.PercentPattern("missing-numbering-system"), "#,##0%"; got != want {
		t.Fatalf("PercentPattern(missing numbering system) = %q, want fallback %q", got, want)
	}
	if got, want := loc.ScientificPattern("latn"), "#E0"; got != want {
		t.Fatalf("ScientificPattern = %q, want %q", got, want)
	}
	if got, want := loc.ScientificPattern("missing-numbering-system"), "#E0"; got != want {
		t.Fatalf("ScientificPattern(missing numbering system) = %q, want fallback %q", got, want)
	}
	if got, want := loc.CurrencyPattern("latn", "standard"), "¤#,##0.00"; got != want {
		t.Fatalf("CurrencyPattern = %q, want %q", got, want)
	}
	if got, want := loc.CurrencyPattern("latn", "accounting"), "¤#,##0.00;(¤#,##0.00)"; got != want {
		t.Fatalf("CurrencyPattern(accounting) = %q, want %q", got, want)
	}
	if got, want := loc.CurrencyPattern("latn", "missing-sign"), "¤#,##0.00"; got != want {
		t.Fatalf("CurrencyPattern(missing sign) = %q, want fallback %q", got, want)
	}
	if got, want := loc.CurrencyPattern("missing-numbering-system", "accounting"), "¤#,##0.00;(¤#,##0.00)"; got != want {
		t.Fatalf("CurrencyPattern(missing numbering system) = %q, want fallback %q", got, want)
	}
	if got, want := loc.CompactPattern("latn", "short", 3, "one"), "0K"; got != want {
		t.Fatalf("CompactPattern(one) = %q, want %q", got, want)
	}
	if got, want := loc.CompactPattern("latn", "short", 3, "few"), "0K"; got != want {
		t.Fatalf("CompactPattern(few) = %q, want fallback %q", got, want)
	}
	if got, want := loc.CompactPattern("latn", "short", 6, "other"), "0M"; got != want {
		t.Fatalf("CompactPattern(6) = %q, want %q", got, want)
	}
	if got, want := loc.CompactPattern("latn", "long", 3, "other"), "0 thousand"; got != want {
		t.Fatalf("CompactPattern(long, 3) = %q, want %q", got, want)
	}
}

func TestSmokeGermanNumberSymbols(t *testing.T) {
	t.Parallel()

	loc, ok := ResolveLocale("de")
	if !ok {
		t.Fatal(`ResolveLocale("de") = false, want true`)
	}
	symbols := loc.NumberSymbols("latn")
	if symbols.Decimal != "," || symbols.Group != "." {
		t.Fatalf(`NumberSymbols("de", "latn") = decimal %q group %q, want "," and "."`, symbols.Decimal, symbols.Group)
	}
	if !slices.Contains(SupportedLocales(), "de") {
		t.Fatal(`SupportedLocales() does not include "de"`)
	}
}

func TestSmokeCurrencyNamePatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, locale, numberingSystem, plural, want string
	}{
		{name: "English other", locale: "en", plural: "other", want: "{0} {1}"},
		{name: "Swahili one", locale: "sw", plural: "one", want: "{0} {1}"},
		{name: "Swahili other", locale: "sw", plural: "other", want: "{1} {0}"},
		{name: "missing category uses other", locale: "sw", plural: "few", want: "{1} {0}"},
		{name: "missing numbering system uses default category", locale: "sw", numberingSystem: "missing-numbering-system", plural: "one", want: "{0} {1}"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loc, ok := ResolveLocale(tc.locale)
			if !ok {
				t.Fatalf("ResolveLocale(%q) = false, want true", tc.locale)
			}
			if got := loc.CurrencyNamePattern(tc.numberingSystem, tc.plural); got != tc.want {
				t.Errorf("CurrencyNamePattern(%q, %q, %q) = %q, want %q", tc.locale, tc.numberingSystem, tc.plural, got, tc.want)
			}
		})
	}
}

func TestCompactPatternMissingTupleReturnsEmpty(t *testing.T) {
	t.Parallel()

	loc, ok := ResolveLocale("en")
	if !ok {
		t.Fatal(`ResolveLocale("en") = false, want true`)
	}

	for _, tc := range []struct {
		name                             string
		numberingSystem, display, plural string
		exponent                         int
	}{
		{name: "missing numbering system", numberingSystem: "missing", display: "short", exponent: 3, plural: "other"},
		{name: "missing display", numberingSystem: "latn", display: "missing", exponent: 3, plural: "other"},
		{name: "missing exponent", numberingSystem: "latn", display: "short", exponent: 99, plural: "other"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := loc.CompactPattern(tc.numberingSystem, tc.display, tc.exponent, tc.plural); got != "" {
				t.Fatalf("CompactPattern(%q, %q, %d, %q) = %q, want empty", tc.numberingSystem, tc.display, tc.exponent, tc.plural, got)
			}
		})
	}
}

func TestMissingLocaleReturnsZeroNumberData(t *testing.T) {
	t.Parallel()

	loc := Locale(65535)
	for _, tc := range []struct {
		name string
		got  string
		want string
	}{
		{name: "default numbering system", got: loc.DefaultNumberingSystem()},
		{name: "decimal pattern", got: loc.DecimalPattern("latn")},
		{name: "percent pattern", got: loc.PercentPattern("latn")},
		{name: "scientific pattern", got: loc.ScientificPattern("latn")},
		{name: "compact pattern", got: loc.CompactPattern("latn", "short", 3, "other")},
		{name: "currency pattern", got: loc.CurrencyPattern("latn", "standard"), want: defaultCurrencyPattern},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %q, want %q", tc.name, tc.got, tc.want)
		}
	}

	symbols := loc.NumberSymbols("latn")
	if got := symbols.TimeSeparator; got != defaultTimeSeparator {
		t.Errorf("NumberSymbols(missing).TimeSeparator = %q, want %q", got, defaultTimeSeparator)
	}
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "Decimal", value: symbols.Decimal},
		{name: "Group", value: symbols.Group},
		{name: "Percent", value: symbols.Percent},
		{name: "Plus", value: symbols.Plus},
		{name: "Minus", value: symbols.Minus},
		{name: "NaN", value: symbols.NaN},
		{name: "Infinity", value: symbols.Infinity},
		{name: "ApproxSign", value: symbols.ApproxSign},
		{name: "RangeSign", value: symbols.RangeSign},
		{name: "PerMille", value: symbols.PerMille},
		{name: "Exponential", value: symbols.Exponential},
		{name: "SuperscriptingExponent", value: symbols.SuperscriptingExponent},
	} {
		if tc.value != "" {
			t.Errorf("NumberSymbols(missing).%s = %q, want empty", tc.name, tc.value)
		}
	}
}

func TestNumberSymbolsFillDefaultTimeSeparator(t *testing.T) {
	t.Parallel()

	if got, want := withNumberSymbolDefaults(NumberSymbols{}).TimeSeparator, ":"; got != want {
		t.Fatalf("default TimeSeparator = %q, want %q", got, want)
	}
}

// TestSmokeSupportedLocalesWithinProfile asserts every SupportedLocales tag is a
// member of the kernel locale profile subset, scoped to the number domain's
// borrowed kernel.
func TestSmokeSupportedLocalesWithinProfile(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceSubset(t, "SupportedLocales", SupportedLocales(), "kernel locale profile", cldrlocale.AvailableLocales())
}
