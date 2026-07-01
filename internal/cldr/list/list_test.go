package list

import (
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/internal/testprocess"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no pattern decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_LIST_NARROW_INDEX_SUBPROCESS"

const missingLocale Locale = 65535

// TestSupportedLocalesDoesNotDecodePatternBlob asserts the narrow-index rule:
// SupportedLocales reads only the supported blob and must never trigger the
// pattern blob decode.
//
// The assertion runs in a fresh process so other list-pattern tests cannot
// populate the package-level Once state first.
func TestSupportedLocalesDoesNotDecodePatternBlob(t *testing.T) {
	t.Parallel()

	if !testprocess.RunInFreshProcess(t, narrowIndexSubprocessEnv) {
		return
	}
	testcontract.AssertNarrowStringIndexDoesNotLoad(t, "SupportedLocales", SupportedLocales,
		testcontract.LoadProbe{Name: "list pattern blob", Loaded: func() bool { return patternsByLocale != nil }},
	)
}

func TestSupportedLocalesReturnsCopy(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceReturnsCopy(t, "SupportedLocales", SupportedLocales)
}

// TestSmokeKnownPatterns is a checkout-independent smoke test: it asserts a few
// known (locale, type, style) tuples resolved through the kernel "en" handle
// return the strings recorded from the committed data.go. These values are
// intentionally hard-coded so a silent encoder/decoder regression fails here even
// when the FormatJS fixtures are unavailable.
func TestSmokeKnownPatterns(t *testing.T) {
	t.Parallel()

	loc, ok := ResolveLocale("en")
	if !ok {
		t.Fatal(`ResolveLocale("en") = false, want true`)
	}

	for _, tc := range []struct {
		typ, style string
		want       ListPattern
	}{
		{"conjunction", "long", ListPattern{Pair: "{0} and {1}", Start: "{0}, {1}", Middle: "{0}, {1}", End: "{0}, and {1}"}},
		{"disjunction", "long", ListPattern{Pair: "{0} or {1}", Start: "{0}, {1}", Middle: "{0}, {1}", End: "{0}, or {1}"}},
		{"unit", "long", ListPattern{Pair: "{0}, {1}", Start: "{0}, {1}", Middle: "{0}, {1}", End: "{0}, {1}"}},
	} {
		if got := Pattern(loc, tc.typ, tc.style); got != tc.want {
			t.Errorf("Pattern(en, %q, %q) = %+v, want %+v", tc.typ, tc.style, got, tc.want)
		}
	}

	// Empty type/style default to conjunction/long, matching the legacy accessor.
	if got, want := Pattern(loc, "", ""), (ListPattern{Pair: "{0} and {1}", Start: "{0}, {1}", Middle: "{0}, {1}", End: "{0}, and {1}"}); got != want {
		t.Errorf("Pattern(en, defaults) = %+v, want %+v", got, want)
	}
}

func TestPatternMissingTupleReturnsZero(t *testing.T) {
	t.Parallel()

	loc, ok := ResolveLocale("en")
	if !ok {
		t.Fatal(`ResolveLocale("en") = false, want true`)
	}

	for _, tc := range []struct {
		name       string
		loc        Locale
		typ, style string
	}{
		{name: "missing locale", loc: missingLocale, typ: "conjunction", style: "long"},
		{name: "missing type", loc: loc, typ: "missing", style: "long"},
		{name: "missing style", loc: loc, typ: "conjunction", style: "missing"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := Pattern(tc.loc, tc.typ, tc.style); got != (ListPattern{}) {
				t.Fatalf("Pattern(%d, %q, %q) = %+v, want zero", tc.loc, tc.typ, tc.style, got)
			}
		})
	}
}

// TestSmokeSupportedLocalesWithinProfile asserts every SupportedLocales tag is a
// member of the kernel locale profile subset (the available-locale set the list
// key packing indexes against). This mirrors the deleted root snapshot subset
// assertion, scoped to the list domain's borrowed kernel.
func TestSmokeSupportedLocalesWithinProfile(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceSubset(t, "SupportedLocales", SupportedLocales(), "kernel locale profile", cldrlocale.AvailableLocales())
}
