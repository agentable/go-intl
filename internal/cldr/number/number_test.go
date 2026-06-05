package number

import (
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/numbering"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no number-data decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_NUMBER_NARROW_INDEX_SUBPROCESS"

// TestSupportedLocalesDoesNotDecodeNumberBlob asserts the narrow-index rule:
// SupportedLocales reads only the supported blob and must never trigger the
// main number-data decode.
//
// The assertion is order-sensitive: any other test in this binary that touches a
// number lookup populates byLocale, after which a same-binary assertion would
// see it already decoded. Rather than depend on test order, this test
// re-executes the test binary as a subprocess that runs only the inner
// assertion via -test.run, so the assertion always observes a process that has
// decoded nothing. When adding new number-touching tests, keep the assertion in
// this subprocess form.
func TestSupportedLocalesDoesNotDecodeNumberBlob(t *testing.T) {
	t.Parallel()

	if os.Getenv(narrowIndexSubprocessEnv) == "1" {
		assertSupportedLocalesDoesNotDecodeNumberBlob(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestSupportedLocalesDoesNotDecodeNumberBlob$", "-test.v")
	cmd.Env = append(os.Environ(), narrowIndexSubprocessEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("narrow-index subprocess failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "PASS") {
		t.Fatalf("narrow-index subprocess did not report PASS:\n%s", out)
	}
}

func assertSupportedLocalesDoesNotDecodeNumberBlob(t *testing.T) {
	t.Helper()

	tags := SupportedLocales()
	if len(tags) == 0 {
		t.Fatal("SupportedLocales returned no tags")
	}
	if byLocale != nil {
		t.Error("SupportedLocales decoded the number blob; narrow index must not")
	}
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
	if !slices.IsSorted(got) {
		t.Fatalf("SupportedNumberingSystems = %v, want sorted values", got)
	}
	if compact := slices.Compact(slices.Clone(got)); !slices.Equal(compact, got) {
		t.Fatalf("SupportedNumberingSystems = %v, want unique values", got)
	}

	seen := map[string]bool{}
	for _, ns := range numbering.SimpleNumberingSystems {
		seen[ns] = true
	}
	numberingSystemOnce.Do(loadNumberingSystemExtras)
	for _, ns := range numberingSystemExtras {
		seen[ns] = true
	}
	if want := slices.Sorted(maps.Keys(seen)); !slices.Equal(got, want) {
		t.Fatalf("SupportedNumberingSystems = %v, want %v", got, want)
	}
}

// TestSupportedNumberingSystemsReturnsCopy asserts the merged result is a fresh
// slice each call so callers cannot corrupt cached state.
func TestSupportedNumberingSystemsReturnsCopy(t *testing.T) {
	t.Parallel()

	a := SupportedNumberingSystems()
	if len(a) == 0 {
		t.Fatal("SupportedNumberingSystems returned no values")
	}
	a[0] = "mutated"
	b := SupportedNumberingSystems()
	if b[0] == "mutated" {
		t.Error("SupportedNumberingSystems returned a shared slice; callers can corrupt the cache")
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
	if got, want := loc.DecimalPattern("latn"), "#,##0.###"; got != want {
		t.Fatalf("DecimalPattern = %q, want %q", got, want)
	}
	if got, want := loc.PercentPattern("latn"), "#,##0%"; got != want {
		t.Fatalf("PercentPattern = %q, want %q", got, want)
	}
	if got, want := loc.CurrencyPattern("latn", "standard"), "¤#,##0.00"; got != want {
		t.Fatalf("CurrencyPattern = %q, want %q", got, want)
	}
	if got, want := loc.CurrencyPattern("latn", "accounting"), "¤#,##0.00;(¤#,##0.00)"; got != want {
		t.Fatalf("CurrencyPattern(accounting) = %q, want %q", got, want)
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

// TestSmokeSupportedLocalesWithinProfile asserts every SupportedLocales tag is a
// member of the kernel locale profile subset, scoped to the number domain's
// borrowed kernel.
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
