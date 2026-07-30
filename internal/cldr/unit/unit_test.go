package unit

import (
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/internal/testprocess"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no pattern decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_UNIT_NARROW_INDEX_SUBPROCESS"

// TestSupportedLocalesDoesNotDecodePatternBlobs asserts the narrow-index rule:
// SupportedLocales reads only the supported blob and must never trigger the
// pattern, compound, or name-table decode.
//
// The assertion runs in a fresh process so other unit-pattern tests cannot
// populate the package-level Once state first.
func TestSupportedLocalesDoesNotDecodePatternBlobs(t *testing.T) {
	t.Parallel()

	if !testprocess.RunInFreshProcess(t, narrowIndexSubprocessEnv) {
		return
	}
	testcontract.AssertNarrowStringIndexDoesNotLoad(t, "SupportedLocales", SupportedLocales,
		testcontract.LoadProbe{Name: "unit pattern blob", Loaded: func() bool { return unitPatterns != nil }},
		testcontract.LoadProbe{Name: "per-unit pattern blob", Loaded: func() bool { return perUnitPatternRows != nil }},
		testcontract.LoadProbe{Name: "compound unit blob", Loaded: func() bool { return compoundUnitRows != nil }},
		testcontract.LoadProbe{Name: "unit name table", Loaded: func() bool { return unitNameIDs != nil }},
	)
}

func TestSupportedLocalesReturnsCopy(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceReturnsCopy(t, "SupportedLocales", SupportedLocales)
}

// TestSmokeKnownPatterns is a checkout-independent smoke test: it asserts a few
// known (locale, unit, width, plural) tuples resolved through the kernel "en"
// handle return the strings recorded from the committed data.go. These values
// are intentionally hard-coded so a silent encoder/decoder regression fails here
// even when the FormatJS fixtures are unavailable.
func TestSmokeKnownPatterns(t *testing.T) {
	t.Parallel()

	loc, ok := cldrlocale.ResolveLocale("en")
	if !ok {
		t.Fatal(`ResolveLocale("en") = false, want true`)
	}

	for _, tc := range []struct {
		unit, width, plural, want string
	}{
		{"meter", "long", "one", "{0} meter"},
		{"meter", "long", "other", "{0} meters"},
		{"meter", "short", "one", "{0} m"},
		{"meter", "narrow", "one", "{0}m"},
	} {
		if got := UnitPattern(loc, tc.unit, tc.width, tc.plural); got != tc.want {
			t.Errorf("UnitPattern(en, %q, %q, %q) = %q, want %q", tc.unit, tc.width, tc.plural, got, tc.want)
		}
	}

	if got, want := CompoundUnitPattern(loc, "long"), "{0} per {1}"; got != want {
		t.Errorf("CompoundUnitPattern(en, long) = %q, want %q", got, want)
	}
	for _, tc := range []struct {
		unit, width, want string
	}{
		{"meter", "long", "{0} per meter"},
		{"meter", "short", "{0}/m"},
		{"meter", "narrow", "{0}/m"},
		{"megabyte", "long", ""},
	} {
		if got := PerUnitPattern(loc, tc.unit, tc.width); got != tc.want {
			t.Errorf("PerUnitPattern(en, %q, %q) = %q, want %q", tc.unit, tc.width, got, tc.want)
		}
	}
}

func TestUnknownUnitPatternComponentsReturnEmpty(t *testing.T) {
	t.Parallel()

	loc, ok := cldrlocale.ResolveLocale("en")
	if !ok {
		t.Fatal(`ResolveLocale("en") = false, want true`)
	}

	for _, tc := range []struct {
		name   string
		unit   string
		width  string
		plural string
	}{
		{name: "unknown unit", unit: "not-a-unit", width: "long", plural: "other"},
		{name: "unknown width", unit: "meter", width: "wide", plural: "other"},
		{name: "unknown plural", unit: "meter", width: "long", plural: "several"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := UnitPattern(loc, tc.unit, tc.width, tc.plural); got != "" {
				t.Errorf("UnitPattern(en, %q, %q, %q) = %q, want empty pattern", tc.unit, tc.width, tc.plural, got)
			}
		})
	}
}

// TestSmokeSupportedLocalesWithinProfile asserts every SupportedLocales tag is a
// member of the kernel locale profile subset (the available-locale set the unit
// key packing indexes against). This mirrors the deleted root snapshot subset
// assertion, scoped to the unit domain's borrowed kernel.
func TestSmokeSupportedLocalesWithinProfile(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceSubset(t, "SupportedLocales", SupportedLocales(), "kernel locale profile", cldrlocale.AvailableLocales())
}

// TestBinarySearchedBlobsStrictlyAscending pins the binary-search precondition
// over the committed payload: pattern and compound keys must be strictly
// ascending. The round-trip gate also proves this, but it skips when the pinned
// CLDR checkout is absent (as in CI), so this checkout-independent assertion is
// the always-on guard. The generator additionally panics on a regressing delta
// stream at generation time; this test catches a desorted committed blob.
func TestBinarySearchedBlobsStrictlyAscending(t *testing.T) {
	unitPatternOnce.Do(loadUnitPatterns)
	for i := 1; i < len(unitPatterns); i++ {
		if unitPatterns[i].key <= unitPatterns[i-1].key {
			t.Fatalf("unit pattern keys not strictly ascending at %d: %d then %d", i, unitPatterns[i-1].key, unitPatterns[i].key)
		}
	}
	compoundUnitOnce.Do(loadCompoundUnits)
	for i := 1; i < len(compoundUnitRows); i++ {
		if compoundUnitRows[i].key <= compoundUnitRows[i-1].key {
			t.Fatalf("compound unit keys not strictly ascending at %d: %d then %d", i, compoundUnitRows[i-1].key, compoundUnitRows[i].key)
		}
	}
	perUnitPatternOnce.Do(loadPerUnitPatterns)
	for i := 1; i < len(perUnitPatternRows); i++ {
		if perUnitPatternRows[i].key <= perUnitPatternRows[i-1].key {
			t.Fatalf("per-unit pattern keys not strictly ascending at %d: %d then %d", i, perUnitPatternRows[i-1].key, perUnitPatternRows[i].key)
		}
	}
	if len(unitPatterns) == 0 || len(compoundUnitRows) == 0 || len(perUnitPatternRows) == 0 {
		t.Fatal("decoded unit blobs are empty; payload or decoder is broken")
	}
}
