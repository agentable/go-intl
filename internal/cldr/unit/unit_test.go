package unit

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no pattern decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_UNIT_NARROW_INDEX_SUBPROCESS"

// TestSupportedLocalesDoesNotDecodePatternBlobs asserts the narrow-index rule:
// SupportedLocales reads only the supported blob and must never trigger the
// pattern, compound, or name-table decode.
//
// The assertion is order-sensitive: any other test in this binary that touches a
// pattern, compound, or name lookup (for example the TestSmoke* tuple checks
// below) populates those package-level slices, after which a same-binary
// assertion would see them already decoded. Rather than depend on test order,
// this test re-executes the test binary as a subprocess that runs only the
// inner assertion via -test.run, so the assertion always observes a process that
// has decoded nothing. When adding new pattern-touching tests, keep the
// assertion in this subprocess form; do not move the Once checks into a plain
// same-binary test, and do not rely on running before other tests.
func TestSupportedLocalesDoesNotDecodePatternBlobs(t *testing.T) {
	t.Parallel()

	if os.Getenv(narrowIndexSubprocessEnv) == "1" {
		assertSupportedLocalesDoesNotDecodePatternBlobs(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestSupportedLocalesDoesNotDecodePatternBlobs$", "-test.v")
	cmd.Env = append(os.Environ(), narrowIndexSubprocessEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("narrow-index subprocess failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "PASS") {
		t.Fatalf("narrow-index subprocess did not report PASS:\n%s", out)
	}
}

func assertSupportedLocalesDoesNotDecodePatternBlobs(t *testing.T) {
	t.Helper()

	tags := SupportedLocales()
	if len(tags) == 0 {
		t.Fatal("SupportedLocales returned no tags")
	}
	if unitPatterns != nil {
		t.Error("SupportedLocales decoded the unit pattern blob; narrow index must not")
	}
	if compoundUnitRows != nil {
		t.Error("SupportedLocales decoded the compound unit blob; narrow index must not")
	}
	if unitNameIDs != nil {
		t.Error("SupportedLocales decoded the unit name table; narrow index must not")
	}
}

func TestSupportedLocalesReturnsCopy(t *testing.T) {
	t.Parallel()

	a := SupportedLocales()
	if len(a) == 0 {
		t.Fatal("SupportedLocales returned no tags")
	}
	a[0] = "mutated"
	b := SupportedLocales()
	if b[0] == "mutated" {
		t.Error("SupportedLocales returned a shared slice; callers can corrupt the cache")
	}
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
}

// TestSmokeSupportedLocalesWithinProfile asserts every SupportedLocales tag is a
// member of the kernel locale profile subset (the available-locale set the unit
// key packing indexes against). This mirrors the deleted root snapshot subset
// assertion, scoped to the unit domain's borrowed kernel.
func TestSmokeSupportedLocalesWithinProfile(t *testing.T) {
	t.Parallel()

	profile := map[string]bool{}
	for _, tag := range cldrlocale.AvailableLocales() {
		profile[tag] = true
	}

	supported := SupportedLocales()
	if len(supported) == 0 {
		t.Fatal("SupportedLocales returned no tags")
	}
	for _, tag := range supported {
		if !profile[tag] {
			t.Errorf("SupportedLocales tag %q is not in the kernel locale profile", tag)
		}
	}
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
	if len(unitPatterns) == 0 || len(compoundUnitRows) == 0 {
		t.Fatal("decoded unit blobs are empty; payload or decoder is broken")
	}
}
