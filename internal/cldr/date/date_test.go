package date

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no main-blob decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_DATE_NARROW_INDEX_SUBPROCESS"

// TestNarrowIndexDoesNotDecodeMainBlobs asserts the narrow-index rule:
// SupportedLocales and SupportedCalendars read only their narrow blobs and must
// never trigger the gregorian or day-period decode.
//
// The assertion is order-sensitive: any other test in this binary that touches a
// gregorian or day-period lookup populates those package-level maps, after which
// a same-binary assertion would see them already decoded. Rather than depend on
// test order, this test re-executes the test binary as a subprocess that runs
// only the inner assertion via -test.run, so the assertion always observes a
// process that has decoded nothing.
func TestNarrowIndexDoesNotDecodeMainBlobs(t *testing.T) {
	t.Parallel()

	if os.Getenv(narrowIndexSubprocessEnv) == "1" {
		assertNarrowIndexDoesNotDecodeMainBlobs(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestNarrowIndexDoesNotDecodeMainBlobs$", "-test.v")
	cmd.Env = append(os.Environ(), narrowIndexSubprocessEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("narrow-index subprocess failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "PASS") {
		t.Fatalf("narrow-index subprocess did not report PASS:\n%s", out)
	}
}

func assertNarrowIndexDoesNotDecodeMainBlobs(t *testing.T) {
	t.Helper()

	tags := SupportedLocales()
	if len(tags) == 0 {
		t.Fatal("SupportedLocales returned no tags")
	}
	cals := SupportedCalendars()
	if len(cals) == 0 {
		t.Fatal("SupportedCalendars returned no identifiers")
	}
	if gregorianData != nil {
		t.Error("narrow index decoded the gregorian blob; it must not")
	}
	if dayPeriodRules != nil {
		t.Error("narrow index decoded the day-period blob; it must not")
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

func TestSupportedCalendarsReturnsCopy(t *testing.T) {
	t.Parallel()

	a := SupportedCalendars()
	if len(a) == 0 {
		t.Fatal("SupportedCalendars returned no identifiers")
	}
	a[0] = "mutated"
	b := SupportedCalendars()
	if b[0] == "mutated" {
		t.Error("SupportedCalendars returned a shared slice; callers can corrupt the cache")
	}
}

// TestSmokeGregorianEnglish is a checkout-independent smoke test mirroring the
// deleted root dates_test English assertions, scoped to the date domain's
// borrowed kernel handle. These values are hard-coded so a silent
// encoder/decoder regression fails here even without the CLDR checkout.
func TestSmokeGregorianEnglish(t *testing.T) {
	t.Parallel()

	loc := resolveSmoke(t, "en-US")
	g := GregorianFor(loc)

	if got, want := g.Eras.Wide[1], "Anno Domini"; got != want {
		t.Errorf("Eras.Wide[1] = %q, want %q", got, want)
	}
	if got, want := g.Months.Wide[0], "January"; got != want {
		t.Errorf("Months.Wide[0] = %q, want %q", got, want)
	}
	if got, want := g.Weekdays.Wide[0], "Sunday"; got != want {
		t.Errorf("Weekdays.Wide[0] = %q, want %q", got, want)
	}
	if got, want := g.DayPeriods.AM.Wide, "AM"; got != want {
		t.Errorf("DayPeriods.AM.Wide = %q, want %q", got, want)
	}
	if got, want := g.DateFormats[0], "EEEE, MMMM d, y"; got != want {
		t.Errorf("DateFormats[0] = %q, want %q", got, want)
	}
	if got, want := g.DateTimeFormats[2], "{1}, {0}"; got != want {
		t.Errorf("DateTimeFormats[2] = %q, want %q", got, want)
	}
	if got, want := g.DateTimeAtFormats[0], "{1} 'at' {0}"; got != want {
		t.Errorf("DateTimeAtFormats[0] = %q, want %q", got, want)
	}
	if got, want := g.AppendItems["Timezone"], "{0} {1}"; got != want {
		t.Errorf("AppendItems[Timezone] = %q, want %q", got, want)
	}
	if got, want := g.AvailableFormats["yMMMd"], "MMM d, y"; got != want {
		t.Errorf("AvailableFormats[yMMMd] = %q, want %q", got, want)
	}
	if got, want := g.IntervalFallback, "{0} – {1}"; got != want {
		t.Errorf("IntervalFallback = %q, want %q", got, want)
	}
	if got, want := g.IntervalFormats["yMMMd"]["d"], "MMM d – d, y"; got != want {
		t.Errorf("IntervalFormats[yMMMd][d] = %q, want %q", got, want)
	}
}

func TestSmokeFlexibleDayPeriods(t *testing.T) {
	t.Parallel()

	loc := resolveSmoke(t, "zh-Hans-CN")
	flex := GregorianFor(loc).DayPeriods.Flex
	for _, key := range []string{"morning1", "afternoon1", "evening1", "night1"} {
		if flex[key].Wide == "" {
			t.Errorf("DayPeriods.Flex[%q].Wide is empty", key)
		}
	}
}

func TestSmokeDayPeriodForRules(t *testing.T) {
	t.Parallel()

	en := resolveSmoke(t, "en-US")
	zh := resolveSmoke(t, "zh-Hans-CN")
	for _, tc := range []struct {
		loc          Locale
		hour, minute int
		want         string
	}{
		{en, 5, 0, "morning1"},
		{en, 12, 0, "noon"},
		{en, 0, 0, "midnight"},
		{zh, 13, 0, "afternoon2"},
	} {
		if got := DayPeriodFor(tc.loc, tc.hour, tc.minute); got != tc.want {
			t.Errorf("DayPeriodFor(%d:%02d) = %q, want %q", tc.hour, tc.minute, got, tc.want)
		}
	}
}

func TestSmokeSupportedCalendars(t *testing.T) {
	t.Parallel()

	cals := SupportedCalendars()
	wantOrder := []string{"gregory", "iso8601"}
	if len(cals) != len(wantOrder) {
		t.Fatalf("SupportedCalendars() = %v, want %v", cals, wantOrder)
	}
	for i, want := range wantOrder {
		if cals[i] != want {
			t.Fatalf("SupportedCalendars()[%d] = %q, want %q", i, cals[i], want)
		}
	}
}

// TestSmokeDateLocaleData asserts DateLocaleData routes the calendar key through
// the date narrow index and the hour-cycle/numbering keys through the kernel,
// mirroring the legacy root DateLocaleData semantics.
func TestSmokeDateLocaleData(t *testing.T) {
	t.Parallel()

	ca := DateLocaleData{}.For("en-US", "ca")
	if len(ca) == 0 || ca[0] != "gregory" {
		t.Errorf("DateLocaleData.For(ca) = %v, want SupportedCalendars order", ca)
	}
	hc := DateLocaleData{}.For("en-US", "hc")
	if len(hc) == 0 {
		t.Error("DateLocaleData.For(hc) returned no hour-cycle values")
	}
	nu := DateLocaleData{}.For("en-US", "nu")
	if len(nu) == 0 || nu[0] != "latn" {
		t.Errorf("DateLocaleData.For(nu) = %v, want latn first", nu)
	}
	if got := (DateLocaleData{}).For("en-US", "co"); got != nil {
		t.Errorf("DateLocaleData.For(unknown key) = %v, want nil", got)
	}
}

func resolveSmoke(t *testing.T, tag string) Locale {
	t.Helper()
	loc, ok := cldrlocale.ResolveLocale(tag)
	if !ok {
		t.Fatalf("ResolveLocale(%q) = false, want true", tag)
	}
	return loc
}
