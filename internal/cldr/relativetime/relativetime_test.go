package relativetime

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no field decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_RELATIVETIME_NARROW_INDEX_SUBPROCESS"

// TestSupportedLocalesDoesNotDecodeFieldBlob asserts the narrow-index rule:
// SupportedLocales reads only the supported blob and must never trigger the
// field blob decode.
//
// The assertion is order-sensitive: any other test in this binary that touches a
// field lookup (for example the smoke checks below) populates the field map,
// after which a same-binary assertion would see it already decoded. Rather than
// depend on test order, this test re-executes the test binary as a subprocess
// that runs only the inner assertion via -test.run, so the assertion always
// observes a process that has decoded nothing. When adding new field-touching
// tests, keep the assertion in this subprocess form.
func TestSupportedLocalesDoesNotDecodeFieldBlob(t *testing.T) {
	t.Parallel()

	if os.Getenv(narrowIndexSubprocessEnv) == "1" {
		assertSupportedLocalesDoesNotDecodeFieldBlob(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestSupportedLocalesDoesNotDecodeFieldBlob$", "-test.v")
	cmd.Env = append(os.Environ(), narrowIndexSubprocessEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("narrow-index subprocess failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "PASS") {
		t.Fatalf("narrow-index subprocess did not report PASS:\n%s", out)
	}
}

func assertSupportedLocalesDoesNotDecodeFieldBlob(t *testing.T) {
	t.Helper()

	tags := SupportedLocales()
	if len(tags) == 0 {
		t.Fatal("SupportedLocales returned no tags")
	}
	if fieldsByLocale != nil {
		t.Error("SupportedLocales decoded the relative-time field blob; narrow index must not")
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

// TestSmokeKnownFields is a checkout-independent smoke test: it asserts a few
// known (unit, style, section, key) tuples resolved through the kernel "en"
// handle return the strings recorded from the committed data.go. These values
// are intentionally hard-coded so a silent encoder/decoder regression fails here
// even when the FormatJS fixtures are unavailable.
func TestSmokeKnownFields(t *testing.T) {
	t.Parallel()

	loc, ok := cldrlocale.ResolveLocale("en")
	if !ok {
		t.Fatal(`ResolveLocale("en") = false, want true`)
	}

	fields := FieldsFor(loc)
	if len(fields) == 0 {
		t.Fatal("FieldsFor(en) returned no fields")
	}

	day := fields["day"]["long"]
	if got, want := day.Future["other"], "in {0} days"; got != want {
		t.Errorf("day/long future[other] = %q, want %q", got, want)
	}
	if got, want := day.Past["one"], "{0} day ago"; got != want {
		t.Errorf("day/long past[one] = %q, want %q", got, want)
	}
	if got, want := day.Relative["0"], "today"; got != want {
		t.Errorf("day/long relative[0] = %q, want %q", got, want)
	}
}

// TestSmokeSupportedLocalesWithinProfile asserts every SupportedLocales tag is a
// member of the kernel locale profile subset. This mirrors the deleted root
// snapshot subset assertion, scoped to the relativetime domain's borrowed
// kernel.
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
