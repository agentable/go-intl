package displaynames

import (
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no names decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_DISPLAYNAMES_NARROW_INDEX_SUBPROCESS"

// TestSupportedLocalesDoesNotDecodeNameBlobs asserts the narrow-index rule:
// SupportedLocales reads only the supported blob and must never trigger any of
// the per-kind names blob decodes.
//
// The assertion is order-sensitive: any other test in this binary that touches a
// name lookup populates the corresponding map, after which a same-binary
// assertion would see it already decoded. Rather than depend on test order, this
// test re-executes the test binary as a subprocess that runs only the inner
// assertion via -test.run, so the assertion always observes a process that has
// decoded nothing. When adding new name-touching tests, keep the assertion in
// this subprocess form.
func TestSupportedLocalesDoesNotDecodeNameBlobs(t *testing.T) {
	t.Parallel()

	if os.Getenv(narrowIndexSubprocessEnv) == "1" {
		assertSupportedLocalesDoesNotDecodeNameBlobs(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestSupportedLocalesDoesNotDecodeNameBlobs$", "-test.v")
	cmd.Env = append(os.Environ(), narrowIndexSubprocessEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("narrow-index subprocess failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "PASS") {
		t.Fatalf("narrow-index subprocess did not report PASS:\n%s", out)
	}
}

func assertSupportedLocalesDoesNotDecodeNameBlobs(t *testing.T) {
	t.Helper()

	tags := SupportedLocales()
	if len(tags) == 0 {
		t.Fatal("SupportedLocales returned no tags")
	}
	if languageByLocale != nil {
		t.Error("SupportedLocales decoded the language blob; narrow index must not")
	}
	if territoryByLocale != nil {
		t.Error("SupportedLocales decoded the territory blob; narrow index must not")
	}
	if scriptByLocale != nil {
		t.Error("SupportedLocales decoded the script blob; narrow index must not")
	}
	if calendarByLocale != nil {
		t.Error("SupportedLocales decoded the calendar blob; narrow index must not")
	}
	if fieldByLocale != nil {
		t.Error("SupportedLocales decoded the date-time-field blob; narrow index must not")
	}
}

func TestSupportedLocalesSortedAndUnique(t *testing.T) {
	t.Parallel()

	tags := SupportedLocales()
	if !slices.IsSorted(tags) {
		t.Errorf("SupportedLocales = %v, want sorted", tags)
	}
	if compact := slices.Compact(slices.Clone(tags)); !slices.Equal(compact, tags) {
		t.Errorf("SupportedLocales = %v, want unique", tags)
	}
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
		loc, kind, style, languageDisplay, code, fallback, want string
	}{
		{"en", "language", "long", "dialect", "fr", "code", "French"},
		{"fr", "language", "long", "dialect", "en", "code", "anglais"},
		{"en", "language", "long", "dialect", "en-CA", "code", "Canadian English"},
		{"en", "region", "long", "", "US", "code", "United States"},
		{"en", "script", "long", "", "Latn", "code", "Latin"},
		{"en", "calendar", "long", "", "gregory", "code", "Gregorian Calendar"},
		{"en", "dateTimeField", "long", "", "year", "code", "year"},
	}
	for _, c := range cases {
		got, ok := Of(c.loc, c.kind, c.style, c.languageDisplay, c.code, c.fallback)
		if !ok || got != c.want {
			t.Errorf("Of(%q, %q, %q, %q, %q) = %q (ok=%v), want %q", c.loc, c.kind, c.style, c.languageDisplay, c.code, got, ok, c.want)
		}
	}
}

// TestCurrencyKindDelegates confirms the currency kind still routes through the
// shared currency accessors rather than the local name blobs.
func TestCurrencyKindDelegates(t *testing.T) {
	t.Parallel()

	got, ok := Of("en", "currency", "long", "", "USD", "code")
	if !ok || got != "US Dollar" {
		t.Errorf("Of(en, currency, USD) = %q (ok=%v), want %q", got, ok, "US Dollar")
	}
}
