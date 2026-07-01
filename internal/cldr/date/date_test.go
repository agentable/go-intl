package date

import (
	"slices"
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/internal/testprocess"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no main-blob decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_DATE_NARROW_INDEX_SUBPROCESS"

// TestNarrowIndexDoesNotDecodeMainBlobs asserts the narrow-index rule:
// SupportedLocales and SupportedCalendars read only their narrow blobs and must
// never trigger the gregorian or day-period decode.
//
// The assertion runs in a fresh process so other date-data tests cannot populate
// the package-level Once state first.
func TestNarrowIndexDoesNotDecodeMainBlobs(t *testing.T) {
	t.Parallel()

	if !testprocess.RunInFreshProcess(t, narrowIndexSubprocessEnv) {
		return
	}
	probes := []testcontract.LoadProbe{
		{Name: "gregorian blob", Loaded: func() bool { return gregorianData != nil }},
		{Name: "day-period blob", Loaded: func() bool { return dayPeriodRules != nil }},
	}
	testcontract.AssertNarrowStringIndexDoesNotLoad(t, "SupportedLocales", SupportedLocales, probes...)
	testcontract.AssertNarrowStringIndexDoesNotLoad(t, "SupportedCalendars", SupportedCalendars, probes...)
}

func TestSupportedLocalesReturnsCopy(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceReturnsCopy(t, "SupportedLocales", SupportedLocales)
}

func TestSupportedCalendarsReturnsCopy(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceReturnsCopy(t, "SupportedCalendars", SupportedCalendars)
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

func TestGregorianForReturnsMapCopies(t *testing.T) {
	t.Parallel()

	loc := resolveSmoke(t, "en-US")
	g := GregorianFor(loc)
	if g.AvailableFormats["yMMMd"] == "" || g.AppendItems["Timezone"] == "" || g.IntervalFormats["yMMMd"]["d"] == "" {
		t.Fatal("GregorianFor(en-US) missing expected format maps")
	}
	wantAvailable := g.AvailableFormats["yMMMd"]
	wantAppendItem := g.AppendItems["Timezone"]
	wantInterval := g.IntervalFormats["yMMMd"]["d"]

	g.AvailableFormats["yMMMd"] = "mutated available"
	g.AppendItems["Timezone"] = "mutated append item"
	g.IntervalFormats["yMMMd"]["d"] = "mutated interval"
	delete(g.IntervalFormats, "yMMMd")

	again := GregorianFor(loc)
	if got := again.AvailableFormats["yMMMd"]; got != wantAvailable {
		t.Errorf("GregorianFor(en-US).AvailableFormats[yMMMd] = %q, want %q", got, wantAvailable)
	}
	if got := again.AppendItems["Timezone"]; got != wantAppendItem {
		t.Errorf("GregorianFor(en-US).AppendItems[Timezone] = %q, want %q", got, wantAppendItem)
	}
	if got := again.IntervalFormats["yMMMd"]["d"]; got != wantInterval {
		t.Errorf("GregorianFor(en-US).IntervalFormats[yMMMd][d] = %q, want %q", got, wantInterval)
	}
}

func TestGregorianForMissingLocaleReturnsEmptyView(t *testing.T) {
	t.Parallel()

	g := GregorianFor(Locale(65535))
	var wantEras [2]string
	if got := g.Eras.Wide; got != wantEras {
		t.Errorf("GregorianFor(missing).Eras.Wide = %v, want %v", got, wantEras)
	}
	var wantMonths [12]string
	if got := g.Months.Wide; got != wantMonths {
		t.Errorf("GregorianFor(missing).Months.Wide = %v, want %v", got, wantMonths)
	}
	var wantWeekdays [7]string
	if got := g.Weekdays.Wide; got != wantWeekdays {
		t.Errorf("GregorianFor(missing).Weekdays.Wide = %v, want %v", got, wantWeekdays)
	}
	var wantStyles [4]string
	if got := g.DateFormats; got != wantStyles {
		t.Errorf("GregorianFor(missing).DateFormats = %v, want %v", got, wantStyles)
	}
	if got := len(g.DayPeriods.Flex); got != 0 {
		t.Errorf("len(GregorianFor(missing).DayPeriods.Flex) = %d, want 0", got)
	}
	if g.AvailableFormats != nil || g.IntervalFormats != nil || g.AppendItems != nil {
		t.Errorf("GregorianFor(missing) maps = (%v, %v, %v), want nil maps", g.AvailableFormats, g.IntervalFormats, g.AppendItems)
	}
	if g.IntervalFallback != "" {
		t.Errorf("GregorianFor(missing).IntervalFallback = %q, want empty", g.IntervalFallback)
	}
}

func TestStyleArraySlotOrder(t *testing.T) {
	t.Parallel()

	got := styleArray(map[string]string{
		"full":   "full style",
		"long":   "long style",
		"medium": "medium style",
		"short":  "short style",
	})
	want := [4]string{"full style", "long style", "medium style", "short style"}
	if got != want {
		t.Fatalf("styleArray() = %v, want %v", got, want)
	}
}

func TestFixedDayPeriodSlotOrder(t *testing.T) {
	t.Parallel()

	wide := []string{"midnight wide", "AM wide", "noon wide", "PM wide"}
	abbr := []string{"midnight abbr", "AM abbr", "noon abbr", "PM abbr"}
	narrow := []string{"midnight narrow", "A", "noon narrow", "P"}
	for _, tc := range []struct {
		name string
		slot int
		want dayPeriodNames
	}{
		{name: "AM", slot: dayPeriodAMSlot, want: dayPeriodNames{Wide: "AM wide", Abbr: "AM abbr", Narrow: "A"}},
		{name: "PM", slot: dayPeriodPMSlot, want: dayPeriodNames{Wide: "PM wide", Abbr: "PM abbr", Narrow: "P"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := dayPeriodNamesAt(wide, abbr, narrow, tc.slot); got != tc.want {
				t.Fatalf("dayPeriodNamesAt(%s slot) = %+v, want %+v", tc.name, got, tc.want)
			}
		})
	}
}

func TestFlexibleDayPeriodSlotOrder(t *testing.T) {
	t.Parallel()

	wide := []string{
		"midnight wide", "AM wide", "noon wide", "PM wide",
		"morning1 wide", "morning2 wide", "afternoon1 wide", "afternoon2 wide",
		"evening1 wide", "evening2 wide", "night1 wide", "night2 wide",
	}
	abbr := []string{
		"midnight abbr", "AM abbr", "noon abbr", "PM abbr",
		"morning1 abbr", "morning2 abbr", "afternoon1 abbr", "afternoon2 abbr",
		"evening1 abbr", "evening2 abbr", "night1 abbr", "night2 abbr",
	}
	narrow := []string{
		"midnight narrow", "A", "noon narrow", "P",
		"morning1 narrow", "morning2 narrow", "afternoon1 narrow", "afternoon2 narrow",
		"evening1 narrow", "evening2 narrow", "night1 narrow", "night2 narrow",
	}
	got := flexibleDayPeriodNames(wide, abbr, narrow)
	want := map[string]dayPeriodNames{
		"midnight":   {Wide: "midnight wide", Abbr: "midnight abbr", Narrow: "midnight narrow"},
		"noon":       {Wide: "noon wide", Abbr: "noon abbr", Narrow: "noon narrow"},
		"morning1":   {Wide: "morning1 wide", Abbr: "morning1 abbr", Narrow: "morning1 narrow"},
		"morning2":   {Wide: "morning2 wide", Abbr: "morning2 abbr", Narrow: "morning2 narrow"},
		"afternoon1": {Wide: "afternoon1 wide", Abbr: "afternoon1 abbr", Narrow: "afternoon1 narrow"},
		"afternoon2": {Wide: "afternoon2 wide", Abbr: "afternoon2 abbr", Narrow: "afternoon2 narrow"},
		"evening1":   {Wide: "evening1 wide", Abbr: "evening1 abbr", Narrow: "evening1 narrow"},
		"evening2":   {Wide: "evening2 wide", Abbr: "evening2 abbr", Narrow: "evening2 narrow"},
		"night1":     {Wide: "night1 wide", Abbr: "night1 abbr", Narrow: "night1 narrow"},
		"night2":     {Wide: "night2 wide", Abbr: "night2 abbr", Narrow: "night2 narrow"},
	}
	if len(got) != len(want) {
		t.Fatalf("flexibleDayPeriodNames() len = %d, want %d: %v", len(got), len(want), got)
	}
	for key, wantNames := range want {
		if gotNames := got[key]; gotNames != wantNames {
			t.Fatalf("flexibleDayPeriodNames()[%q] = %+v, want %+v", key, gotNames, wantNames)
		}
	}
	for _, skipped := range []string{"AM", "PM"} {
		if _, ok := got[skipped]; ok {
			t.Fatalf("flexibleDayPeriodNames()[%q] is present, want skipped", skipped)
		}
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

func TestDayPeriodForMissingLocaleReturnsEmpty(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		hour, minute int
	}{
		{0, 0},
		{12, 0},
		{23, 59},
	} {
		if got := DayPeriodFor(Locale(65535), tc.hour, tc.minute); got != "" {
			t.Errorf("DayPeriodFor(missing, %d:%02d) = %q, want empty", tc.hour, tc.minute, got)
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

func TestDateLocaleDataEnglishHourCyclePreference(t *testing.T) {
	t.Parallel()

	first := DateLocaleData{}.For("en", "hc")
	if want := []string{"h12", "h23"}; !slices.Equal(first, want) {
		t.Fatalf(`DateLocaleData.For("en", "hc") = %v, want %v`, first, want)
	}
	first[0] = "mutated"

	got := DateLocaleData{}.For("en", "hc")
	if want := []string{"h12", "h23"}; !slices.Equal(got, want) {
		t.Fatalf(`DateLocaleData.For("en", "hc") after caller mutation = %v, want %v`, got, want)
	}
}

func TestLocaleRegionUsesLocaleSubtagGrammar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tag  string
		want string
	}{
		{name: "alpha region", tag: "en-US", want: "US"},
		{name: "numeric region", tag: "es-419", want: "419"},
		{name: "script before region", tag: "sr-Latn-RS", want: "RS"},
		{name: "extension singleton stops scan", tag: "en-u-hc-h12"},
		{name: "world fallback region omitted", tag: "en-ZZ"},
		{name: "invalid region", tag: "en-Latn-12x"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := localeRegion(tc.tag); got != tc.want {
				t.Fatalf("localeRegion(%q) = %q, want %q", tc.tag, got, tc.want)
			}
		})
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
