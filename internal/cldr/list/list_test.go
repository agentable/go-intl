package list

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no pattern decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_LIST_NARROW_INDEX_SUBPROCESS"

// TestSupportedLocalesDoesNotDecodePatternBlob asserts the narrow-index rule:
// SupportedLocales reads only the supported blob and must never trigger the
// pattern blob decode.
//
// The assertion is order-sensitive: any other test in this binary that touches a
// pattern lookup (for example the TestSmoke* tuple checks below) populates the
// package-level pattern map, after which a same-binary assertion would see it
// already decoded. Rather than depend on test order, this test re-executes the
// test binary as a subprocess that runs only the inner assertion via -test.run,
// so the assertion always observes a process that has decoded nothing. When
// adding new pattern-touching tests, keep the assertion in this subprocess form;
// do not move the Once check into a plain same-binary test, and do not rely on
// running before other tests.
func TestSupportedLocalesDoesNotDecodePatternBlob(t *testing.T) {
	t.Parallel()

	if os.Getenv(narrowIndexSubprocessEnv) == "1" {
		assertSupportedLocalesDoesNotDecodePatternBlob(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestSupportedLocalesDoesNotDecodePatternBlob$", "-test.v")
	cmd.Env = append(os.Environ(), narrowIndexSubprocessEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("narrow-index subprocess failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "PASS") {
		t.Fatalf("narrow-index subprocess did not report PASS:\n%s", out)
	}
}

func assertSupportedLocalesDoesNotDecodePatternBlob(t *testing.T) {
	t.Helper()

	tags := SupportedLocales()
	if len(tags) == 0 {
		t.Fatal("SupportedLocales returned no tags")
	}
	if patternsByLocale != nil {
		t.Error("SupportedLocales decoded the list pattern blob; narrow index must not")
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

// TestSmokeSupportedLocalesWithinProfile asserts every SupportedLocales tag is a
// member of the kernel locale profile subset (the available-locale set the list
// key packing indexes against). This mirrors the deleted root snapshot subset
// assertion, scoped to the list domain's borrowed kernel.
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
