package currency

import (
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no fraction or names decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_CURRENCY_NARROW_INDEX_SUBPROCESS"

// TestSupportedCodesDoesNotDecodeOtherBlobs asserts the narrow-index rule:
// SupportedCodes reads only the supported blob and must never trigger the
// fraction or names blob decode.
//
// The assertion is order-sensitive: any other test in this binary that touches a
// fraction or name lookup (for example the smoke checks below) populates the
// corresponding map, after which a same-binary assertion would see it already
// decoded. Rather than depend on test order, this test re-executes the test
// binary as a subprocess that runs only the inner assertion via -test.run, so
// the assertion always observes a process that has decoded nothing. When adding
// new fraction/name-touching tests, keep the assertion in this subprocess form.
func TestSupportedCodesDoesNotDecodeOtherBlobs(t *testing.T) {
	t.Parallel()

	if os.Getenv(narrowIndexSubprocessEnv) == "1" {
		assertSupportedCodesDoesNotDecodeOtherBlobs(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestSupportedCodesDoesNotDecodeOtherBlobs$", "-test.v")
	cmd.Env = append(os.Environ(), narrowIndexSubprocessEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("narrow-index subprocess failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "PASS") {
		t.Fatalf("narrow-index subprocess did not report PASS:\n%s", out)
	}
}

func assertSupportedCodesDoesNotDecodeOtherBlobs(t *testing.T) {
	t.Helper()

	codes := SupportedCodes()
	if len(codes) == 0 {
		t.Fatal("SupportedCodes returned no codes")
	}
	if fractions != nil {
		t.Error("SupportedCodes decoded the fraction blob; narrow index must not")
	}
	if byLocale != nil {
		t.Error("SupportedCodes decoded the names blob; narrow index must not")
	}
}

func TestSupportedCodesReturnsCopy(t *testing.T) {
	t.Parallel()

	a := SupportedCodes()
	if len(a) == 0 {
		t.Fatal("SupportedCodes returned no codes")
	}
	a[0] = "mutated"
	b := SupportedCodes()
	if b[0] == "mutated" {
		t.Error("SupportedCodes returned a shared slice; callers can corrupt the cache")
	}
}

func TestSupportedCodesSortedAndUnique(t *testing.T) {
	t.Parallel()

	codes := SupportedCodes()
	if !slices.IsSorted(codes) {
		t.Errorf("SupportedCodes = %v, want sorted", codes)
	}
	if compact := slices.Compact(slices.Clone(codes)); !slices.Equal(compact, codes) {
		t.Errorf("SupportedCodes = %v, want unique", codes)
	}
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
	if got, want := Symbol(loc, "JPY", "narrow"), "¥"; got != want {
		t.Errorf("Symbol(JPY, narrow) = %q, want %q", got, want)
	}
}
