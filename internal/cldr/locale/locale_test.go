package cldrlocale

import (
	"slices"
	"testing"

	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/internal/testprocess"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no heavy blob decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_LOCALE_NARROW_INDEX_SUBPROCESS"

// TestRegistryQueriesDoNotDecodeHeavyBlobs asserts the narrow-index rule for the
// kernel: resolving locales and listing available tags reads only the locale
// registry blob and must never trigger the likely-subtags, numbering, or
// preference decode.
//
// The assertion runs in a fresh process so other locale-kernel tests cannot
// populate the package-level Once state first.
func TestRegistryQueriesDoNotDecodeHeavyBlobs(t *testing.T) {
	t.Parallel()

	if !testprocess.RunInFreshProcess(t, narrowIndexSubprocessEnv) {
		return
	}
	assertRegistryQueriesDoNotDecodeHeavyBlobs(t)
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

func TestAvailableLocalesReturnsCopy(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceReturnsCopy(t, "AvailableLocales", AvailableLocales)
}

func TestPreferenceAccessorsReturnCopies(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceReturnsCopy(t, "HourCyclePreference(US)", func() []string {
		return HourCyclePreference("US")
	})
	testcontract.AssertStringSliceReturnsCopy(t, "CalendarPreference(TH)", func() []string {
		return CalendarPreference("TH")
	})
}

func TestIntersectSupportedLocales(t *testing.T) {
	t.Parallel()

	got := IntersectSupportedLocales(
		[]string{"en", "fr", "zh", "de"},
		[]string{"fr", "en", "zh"},
		[]string{"zh", "en"},
	)
	if want := []string{"en", "zh"}; !slices.Equal(got, want) {
		t.Fatalf("IntersectSupportedLocales() = %v, want %v", got, want)
	}

	primary := []string{"en", "fr"}
	got = IntersectSupportedLocales(primary)
	if !slices.Equal(got, primary) {
		t.Fatalf("IntersectSupportedLocales(primary) = %v, want %v", got, primary)
	}
	got[0] = "mutated"
	if primary[0] != "en" {
		t.Fatalf("IntersectSupportedLocales(primary) reused caller storage; primary[0] = %q", primary[0])
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

func TestTextDirection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		script string
		want   string
		ok     bool
	}{
		{script: "Latn", want: "ltr", ok: true},
		{script: "Arab", want: "rtl", ok: true},
		{script: "Sogd", want: "rtl", ok: true},
		{script: "Zzzz"},
		{script: ""},
	}
	for _, tc := range tests {
		got, ok := TextDirection(tc.script)
		if got != tc.want || ok != tc.ok {
			t.Errorf("TextDirection(%q) = %q, %t; want %q, %t", tc.script, got, ok, tc.want, tc.ok)
		}
	}
}

func TestMaximizeSubtagsUsesLikelySubtagFallbackOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		language string
		script   string
		region   string
		wantLang string
		wantScr  string
		wantReg  string
	}{
		{language: "en", region: "CA", wantLang: "en", wantScr: "Latn", wantReg: "CA"},
		{language: "zh", script: "Hant", wantLang: "zh", wantScr: "Hant", wantReg: "TW"},
		{language: "und", script: "Arab", wantLang: "ar", wantScr: "Arab", wantReg: "EG"},
	}
	for _, tc := range tests {
		lang, script, region, ok := MaximizeSubtags(tc.language, tc.script, tc.region)
		if !ok || lang != tc.wantLang || script != tc.wantScr || region != tc.wantReg {
			t.Fatalf("MaximizeSubtags(%q, %q, %q) = %q, %q, %q, %v; want %q, %q, %q, true",
				tc.language, tc.script, tc.region, lang, script, region, ok, tc.wantLang, tc.wantScr, tc.wantReg)
		}
	}
}
