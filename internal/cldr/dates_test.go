package cldr

import (
	"testing"
)

func TestGregorianForEnglishData(t *testing.T) {
	t.Parallel()

	loc, ok := ResolveLocale("en-US")
	if !ok {
		t.Fatal("ResolveLocale(en-US) ok=false")
	}
	gregorian := GregorianFor(loc)
	if got, want := gregorian.Eras.Wide[1], "Anno Domini"; got != want {
		t.Fatalf("GregorianFor(en-US).Eras.Wide[1] = %q, want %q", got, want)
	}
	if got, want := gregorian.Months.Wide[0], "January"; got != want {
		t.Fatalf("GregorianFor(en-US).Months.Wide[0] = %q, want %q", got, want)
	}
	if got, want := gregorian.Weekdays.Wide[0], "Sunday"; got != want {
		t.Fatalf("GregorianFor(en-US).Weekdays.Wide[0] = %q, want %q", got, want)
	}
	if got, want := gregorian.DayPeriods.AM.Wide, "AM"; got != want {
		t.Fatalf("GregorianFor(en-US).DayPeriods.AM.Wide = %q, want %q", got, want)
	}
	if got, want := gregorian.DateFormats[0], "EEEE, MMMM d, y"; got != want {
		t.Fatalf("GregorianFor(en-US).DateFormats[0] = %q, want %q", got, want)
	}
	if got, want := gregorian.TimeFormats[3], "h:mm\u202fa"; got != want {
		t.Fatalf("GregorianFor(en-US).TimeFormats[3] = %q, want %q", got, want)
	}
	if got, want := gregorian.DateTimeFormats[2], "{1}, {0}"; got != want {
		t.Fatalf("GregorianFor(en-US).DateTimeFormats[2] = %q, want %q", got, want)
	}
	if got, want := gregorian.DateTimeAtFormats[0], "{1} 'at' {0}"; got != want {
		t.Fatalf("GregorianFor(en-US).DateTimeAtFormats[0] = %q, want %q", got, want)
	}
	if got, want := gregorian.AppendItems["Timezone"], "{0} {1}"; got != want {
		t.Fatalf("GregorianFor(en-US).AppendItems[Timezone] = %q, want %q", got, want)
	}
	if got, want := gregorian.AvailableFormats["yMMMd"], "MMM d, y"; got != want {
		t.Fatalf("GregorianFor(en-US).AvailableFormats[yMMMd] = %q, want %q", got, want)
	}
	if got, want := gregorian.IntervalFallback, "{0}\u2009–\u2009{1}"; got != want {
		t.Fatalf("GregorianFor(en-US).IntervalFallback = %q, want %q", got, want)
	}
	if got, want := gregorian.IntervalFormats["yMMMd"]["d"], "MMM d\u2009–\u2009d, y"; got != want {
		t.Fatalf("GregorianFor(en-US).IntervalFormats[yMMMd][d] = %q, want %q", got, want)
	}
}

func TestGregorianForChineseFlexibleDayPeriods(t *testing.T) {
	t.Parallel()

	loc, ok := ResolveLocale("zh-Hans-CN")
	if !ok {
		t.Fatal("ResolveLocale(zh-Hans-CN) ok=false")
	}
	flex := GregorianFor(loc).DayPeriods.Flex
	for _, key := range []string{"morning1", "afternoon1", "afternoon2", "evening1", "night1", "midnight"} {
		if flex[key].Wide == "" {
			t.Fatalf("GregorianFor(zh-Hans-CN).DayPeriods.Flex[%q].Wide is empty", key)
		}
	}
}

func TestDayPeriodForLocaleRules(t *testing.T) {
	t.Parallel()

	en, ok := ResolveLocale("en-US")
	if !ok {
		t.Fatal("ResolveLocale(en-US) ok=false")
	}
	zh, ok := ResolveLocale("zh-Hans-CN")
	if !ok {
		t.Fatal("ResolveLocale(zh-Hans-CN) ok=false")
	}
	for _, tc := range []struct {
		name       string
		loc        Locale
		hour       int
		minute     int
		wantPeriod string
	}{
		{name: "en morning", loc: en, hour: 5, wantPeriod: "morning1"},
		{name: "en noon", loc: en, hour: 12, wantPeriod: "noon"},
		{name: "en midnight", loc: en, wantPeriod: "midnight"},
		{name: "zh afternoon", loc: zh, hour: 13, wantPeriod: "afternoon2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := DayPeriodFor(tc.loc, tc.hour, tc.minute); got != tc.wantPeriod {
				t.Fatalf("DayPeriodFor(%s) = %q, want %q", tc.name, got, tc.wantPeriod)
			}
		})
	}
}

func TestDateAccessors(t *testing.T) {
	t.Parallel()

	loc, ok := ResolveLocale("en-US")
	if !ok {
		t.Fatal("ResolveLocale(en-US) ok=false")
	}
	names := loc.CalendarNames("gregorian", "wide", "format")
	if got, want := names.Eras[1], "Anno Domini"; got != want {
		t.Fatalf("CalendarNames eras[1] = %q, want %q", got, want)
	}
	if got, want := names.Months[0], "January"; got != want {
		t.Fatalf("CalendarNames months[0] = %q, want %q", got, want)
	}
	if got, want := names.Weekdays[0], "Sunday"; got != want {
		t.Fatalf("CalendarNames weekdays[0] = %q, want %q", got, want)
	}
	if got, want := names.Quarters[0], "1st quarter"; got != want {
		t.Fatalf("CalendarNames quarters[0] = %q, want %q", got, want)
	}
	if got, want := names.DayPeriods[1], "AM"; got != want {
		t.Fatalf("CalendarNames dayPeriods[1] = %q, want %q", got, want)
	}
	if got, want := loc.DateFormat("gregorian", "full"), "EEEE, MMMM d, y"; got != want {
		t.Fatalf("DateFormat(full) = %q, want %q", got, want)
	}
	if got, want := loc.TimeFormat("gregorian", "short"), "h:mm\u202fa"; got != want {
		t.Fatalf("TimeFormat(short) = %q, want %q", got, want)
	}
	if got, want := loc.DateTimeFormat("gregorian", "medium"), "{1}, {0}"; got != want {
		t.Fatalf("DateTimeFormat(medium) = %q, want %q", got, want)
	}
	if got, want := loc.AvailableFormats("gregorian")["yMMMd"], "MMM d, y"; got != want {
		t.Fatalf("AvailableFormats[yMMMd] = %q, want %q", got, want)
	}
	intervals := loc.IntervalFormats("gregorian")
	if got, want := intervals.FallbackPattern, "{0}\u2009–\u2009{1}"; got != want {
		t.Fatalf("IntervalFormats fallback = %q, want %q", got, want)
	}
	if got, want := intervals.BySkeleton["yMMMd"]["d"], "MMM d\u2009–\u2009d, y"; got != want {
		t.Fatalf("IntervalFormats[yMMMd][d] = %q, want %q", got, want)
	}
}
