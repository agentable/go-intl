package cldrlocale

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no heavy blob decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_LOCALE_NARROW_INDEX_SUBPROCESS"

// TestRegistryQueriesDoNotDecodeHeavyBlobs asserts the narrow-index rule for the
// kernel: resolving locales and listing available tags reads only the locale
// registry blob and must never trigger the likely-subtags, numbering, or
// preference decode.
//
// The assertion is order-sensitive: any other test in this binary that touches a
// maximize/minimize, numbering, or preference lookup populates those
// package-level structures, after which a same-binary assertion would see them
// already decoded. Rather than depend on test order, this test re-executes the
// test binary as a subprocess that runs only the inner assertion via -test.run,
// so the assertion always observes a process that has decoded nothing heavy.
func TestRegistryQueriesDoNotDecodeHeavyBlobs(t *testing.T) {
	t.Parallel()

	if os.Getenv(narrowIndexSubprocessEnv) == "1" {
		assertRegistryQueriesDoNotDecodeHeavyBlobs(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestRegistryQueriesDoNotDecodeHeavyBlobs$", "-test.v")
	cmd.Env = append(os.Environ(), narrowIndexSubprocessEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("narrow-index subprocess failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "PASS") {
		t.Fatalf("narrow-index subprocess did not report PASS:\n%s", out)
	}
}

func assertRegistryQueriesDoNotDecodeHeavyBlobs(t *testing.T) {
	t.Helper()

	if _, ok := ResolveLocale("en"); !ok {
		t.Fatal(`ResolveLocale("en") = false, want true`)
	}
	if len(AvailableLocales()) == 0 {
		t.Fatal("AvailableLocales returned no tags")
	}
	if likelySubtags != nil {
		t.Error("registry query decoded the likely-subtags blob; narrow index must not")
	}
	if numberingByLocale != nil {
		t.Error("registry query decoded the numbering blob; narrow index must not")
	}
	if hourCyclePreference != nil {
		t.Error("registry query decoded the preference blob; narrow index must not")
	}
}

// TestSmokeRegistry is a checkout-independent smoke test over known locale-index
// tuples. The index assignment is load-bearing: every domain packs locale keys
// against it, so a silent registry-decode regression must fail here even without
// the FormatJS fixtures.
func TestSmokeRegistry(t *testing.T) {
	t.Parallel()

	if Undefined != 0 {
		t.Fatalf("Undefined = %d, want 0", Undefined)
	}
	und, ok := ResolveLocale("und")
	if !ok || und != Undefined {
		t.Fatalf("ResolveLocale(und) = %d, %v; want 0, true", und, ok)
	}
	en, ok := ResolveLocale("en")
	if !ok {
		t.Fatal(`ResolveLocale("en") = false`)
	}
	// Base-language fallback: an unknown region falls back to the base handle.
	enUS, ok := ResolveLocale("en-US")
	if !ok || enUS != en {
		t.Fatalf("ResolveLocale(en-US) = %d, %v; want %d (en fallback), true", enUS, ok, en)
	}

	tags := AvailableLocales()
	if len(tags) == 0 || tags[0] != "und" {
		t.Fatalf("AvailableLocales[0] = %q, want und", firstTag(tags))
	}
}

func firstTag(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return tags[0]
}

// TestSmokeSubtagsAndPreferences exercises the heavy blobs through known tuples
// so an encoder/decoder regression in maximize/minimize, numbering, or
// preference fails independently of the FormatJS fixtures.
func TestSmokeSubtagsAndPreferences(t *testing.T) {
	t.Parallel()

	lang, script, region, ok := MaximizeSubtags("zh", "", "")
	if !ok || lang != "zh" || script != "Hans" || region != "CN" {
		t.Errorf("MaximizeSubtags(zh) = %q, %q, %q, %v; want zh, Hans, CN, true", lang, script, region, ok)
	}

	min, _, _, ok := MinimizeSubtags("zh", "Hans", "CN")
	if !ok || min != "zh" {
		t.Errorf("MinimizeSubtags(zh-Hans-CN) = %q, %v; want zh, true", min, ok)
	}

	enUS, ok := ResolveLocale("en")
	if !ok {
		t.Fatal(`ResolveLocale("en") = false`)
	}
	if got := enUS.DefaultNumberingSystem(); got != "latn" {
		t.Errorf("en DefaultNumberingSystem = %q, want latn", got)
	}
	fa, ok := ResolveLocale("fa")
	if !ok {
		t.Fatal(`ResolveLocale("fa") = false`)
	}
	if got := fa.DefaultNumberingSystem(); got != "arabext" {
		t.Errorf("fa DefaultNumberingSystem = %q, want arabext", got)
	}

	if got := HourCyclePreference("US"); len(got) == 0 {
		t.Error("HourCyclePreference(US) returned no values")
	}
	if !HasWeekPreference("US") {
		t.Error("HasWeekPreference(US) = false, want true")
	}
	if got := CalendarPreference("TH"); len(got) == 0 || got[0] != "buddhist" {
		t.Errorf("CalendarPreference(TH)[0] = %q, want buddhist", firstTag(CalendarPreference("TH")))
	}
}
