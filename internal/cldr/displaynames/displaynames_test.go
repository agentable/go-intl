package displaynames

import (
	"testing"

	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/internal/testprocess"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no names decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_DISPLAYNAMES_NARROW_INDEX_SUBPROCESS"

// TestSupportedLocalesDoesNotDecodeNameBlobs asserts the narrow-index rule:
// SupportedLocales reads only the supported blob and must never trigger any of
// the per-kind names blob decodes.
//
// The assertion runs in a fresh process so other display-name tests cannot
// populate the package-level Once state first.
func TestSupportedLocalesDoesNotDecodeNameBlobs(t *testing.T) {
	t.Parallel()

	if !testprocess.RunInFreshProcess(t, narrowIndexSubprocessEnv) {
		return
	}
	testcontract.AssertNarrowStringIndexDoesNotLoad(t, "SupportedLocales", SupportedLocales,
		testcontract.LoadProbe{Name: "language blob", Loaded: func() bool { return languageByLocale != nil }},
		testcontract.LoadProbe{Name: "territory blob", Loaded: func() bool { return territoryByLocale != nil }},
		testcontract.LoadProbe{Name: "script blob", Loaded: func() bool { return scriptByLocale != nil }},
		testcontract.LoadProbe{Name: "calendar blob", Loaded: func() bool { return calendarByLocale != nil }},
		testcontract.LoadProbe{Name: "date-time-field blob", Loaded: func() bool { return fieldByLocale != nil }},
	)
}

func TestSupportedLocalesSortedAndUnique(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceSortedUnique(t, "SupportedLocales", SupportedLocales())
}

func TestSupportedLocalesReturnsCopy(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceReturnsCopy(t, "SupportedLocales", SupportedLocales)
}

// TestSmokeKnownDisplayNames is a checkout-independent smoke test: a few known
// (locale, kind, code) tuples resolve to the display names recorded in the
// committed data.go, including the language-with-region composition that reads
// the territory blob. These values are intentionally hard-coded so a silent
// encoder/decoder regression fails here even when the FormatJS fixtures are
// unavailable.
func TestSmokeKnownDisplayNames(t *testing.T) {
	t.Parallel()

	cases := []struct {
		loc, kind, style, languageDisplay, code, want string
		fallbackCode                                  bool
	}{
		{loc: "en", kind: "language", style: "long", languageDisplay: "dialect", code: "fr", fallbackCode: true, want: "French"},
		{loc: "fr", kind: "language", style: "long", languageDisplay: "dialect", code: "en", fallbackCode: true, want: "anglais"},
		{loc: "en", kind: "language", style: "long", languageDisplay: "dialect", code: "en-CA", fallbackCode: true, want: "Canadian English"},
		{loc: "en", kind: "region", style: "long", code: "US", fallbackCode: true, want: "United States"},
		{loc: "en", kind: "script", style: "long", code: "Latn", fallbackCode: true, want: "Latin"},
		{loc: "en", kind: "calendar", style: "long", code: "gregory", fallbackCode: true, want: "Gregorian Calendar"},
		{loc: "en", kind: "calendar", style: "long", code: "ethioaa", fallbackCode: true, want: "Ethiopic Amete Alem Calendar"},
		{loc: "en", kind: "dateTimeField", style: "long", code: "year", fallbackCode: true, want: "year"},
		{loc: "en", kind: "dateTimeField", style: "long", code: "dayPeriod", fallbackCode: true, want: "AM/PM"},
	}
	for _, c := range cases {
		got, ok := Of(c.loc, c.kind, c.style, c.languageDisplay, c.code, c.fallbackCode)
		if !ok || got != c.want {
			t.Errorf("Of(%q, %q, %q, %q, %q) = %q (ok=%v), want %q", c.loc, c.kind, c.style, c.languageDisplay, c.code, got, ok, c.want)
		}
	}
}

func TestLanguageRegionFallbackUsesBooleanBoundary(t *testing.T) {
	t.Parallel()

	got, ok := Of("en", "language", "long", "dialect", "en-QQ", true)
	if !ok || got != "English (QQ)" {
		t.Fatalf(`Of("en", "language", "en-QQ", fallbackCode=true) = %q (ok=%v), want "English (QQ)"`, got, ok)
	}
	got, ok = Of("en", "language", "long", "dialect", "en-QQ", false)
	if ok || got != "" {
		t.Fatalf(`Of("en", "language", "en-QQ", fallbackCode=false) = %q (ok=%v), want no value`, got, ok)
	}
}

func TestLanguageBaseAndRegionUsesLocaleSubtagGrammar(t *testing.T) {
	t.Parallel()

	display := styledNames{
		long: map[string]string{
			"en":      "English",
			"en-Latn": "Latin English",
		},
	}
	tests := []struct {
		name       string
		parts      []string
		wantBase   string
		wantRegion string
		wantOK     bool
	}{
		{name: "alpha region", parts: []string{"en", "US"}, wantBase: "English", wantRegion: "US", wantOK: true},
		{name: "numeric region", parts: []string{"en", "419"}, wantBase: "English", wantRegion: "419", wantOK: true},
		{name: "script before region", parts: []string{"en", "Latn", "US"}, wantBase: "Latin English", wantRegion: "US", wantOK: true},
		{name: "invalid region", parts: []string{"en", "12x"}, wantBase: "English", wantOK: true},
		{name: "unknown language", parts: []string{"fr", "US"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			base, region, ok := languageBaseAndRegion(display, "long", tc.parts)
			if base != tc.wantBase || region != tc.wantRegion || ok != tc.wantOK {
				t.Fatalf("languageBaseAndRegion(%v) = %q, %q, %v; want %q, %q, %v", tc.parts, base, region, ok, tc.wantBase, tc.wantRegion, tc.wantOK)
			}
		})
	}
}

// TestCurrencyKindDelegates confirms the currency kind still routes through the
// shared currency name accessors rather than local name blobs or NumberFormat
// symbols.
func TestCurrencyKindDelegates(t *testing.T) {
	t.Parallel()

	got, ok := Of("en", "currency", "long", "", "USD", true)
	if !ok || got != "US Dollar" {
		t.Errorf("Of(en, currency, USD) = %q (ok=%v), want %q", got, ok, "US Dollar")
	}
	got, ok = Of("en", "currency", "narrow", "", "EUR", true)
	if !ok || got != "Euro" {
		t.Errorf("Of(en, currency narrow, EUR) = %q (ok=%v), want %q", got, ok, "Euro")
	}
}

func TestCurrencyKindUsesLocaleFallback(t *testing.T) {
	t.Parallel()

	got, ok := Of("und", "currency", "long", "", "USD", true)
	if !ok || got != "US Dollar" {
		t.Fatalf(`Of("und", "currency", "USD") = %q (ok=%v), want "US Dollar" from fallback locale`, got, ok)
	}
}
