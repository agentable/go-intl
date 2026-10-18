package codegen

import (
	"maps"
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/cldr/date"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// TestDateRoundTrip compares pinned loader rows with production accessors over
// committed generated blobs. Encoder unit tests and regeneration separately
// verify the source-to-blob step. This test skips without the pinned checkout.
func TestDateRoundTrip(t *testing.T) {
	t.Parallel()

	dates := loadDateTestInput(t)
	gregLocales := dateGregorianLocales(dates)
	if len(gregLocales) == 0 {
		t.Fatal("pinned date profile contains no Gregorian locales")
	}

	// Gregorian view: assemble the expected Gregorian from the extract rows using
	// the same field projection the accessor performs, then compare to the
	// production accessor result.
	for _, locale := range gregLocales {
		loc := resolveKernelLocale(t, locale)
		calendar := dates[locale].Calendars[dateCLDRGregorianCalendar]
		got := date.GregorianFor(loc)
		assertGregorianEqual(t, locale, got, expectedGregorian(calendar))
		assertDayPeriodNames(t, locale, got, calendar)
	}

	// Check every registered row, including intentional absence. Supplemental
	// keys outside this registry are not encoded and must not resolve via fallback.
	rulesByLocale := dates[gregLocales[0]].DayPeriodRules
	for _, tag := range cldrlocale.AvailableLocales() {
		loc := resolveKernelLocale(t, tag)
		want := expectedDayPeriodMinutes(t, tag, rulesByLocale[tag])
		for minute, period := range want {
			if got := date.DayPeriodFor(loc, minute/60, minute%60); got != period {
				t.Errorf("DayPeriodFor(%q, %02d:%02d) = %q, want %q", tag, minute/60, minute%60, got, period)
				break
			}
		}
	}

	// Supported-locale narrow index.
	wantTags := gregLocales
	gotTags := date.SupportedLocales()
	assertStringSliceEqual(t, "SupportedLocales", gotTags, wantTags)

	// Supported-calendar narrow index.
	wantCal, err := dateSupportedCalendars(dates)
	if err != nil {
		t.Fatalf("dateSupportedCalendars(): %v", err)
	}
	gotCal := date.SupportedCalendars()
	assertStringSliceEqual(t, "SupportedCalendars", gotCal, wantCal)
}

// Paint source intervals into a minute table, then overlay fixed points. This
// derives the expected partition without repeating the accessor's lookup loop.
func expectedDayPeriodMinutes(t *testing.T, tag string, rules []cldr.DayPeriodRange) [24 * 60]string {
	t.Helper()
	var minutes [24 * 60]string
	points := make(map[int]string)
	for _, rule := range rules {
		from, to := int(rule.From/time.Minute), int(rule.To/time.Minute)
		if rule.From%time.Minute != 0 || rule.To%time.Minute != 0 || from < 0 || from >= len(minutes) || to < 0 || to > len(minutes) {
			t.Fatalf("day-period source %q has unsupported boundary: %+v", tag, rule)
		}
		if from == to {
			points[from] = rule.Type
			continue
		}
		if to < from {
			to += len(minutes)
		}
		for minute := from; minute < to; minute++ {
			index := minute % len(minutes)
			if minutes[index] != "" && minutes[index] != rule.Type {
				t.Fatalf("day-period source %q overlaps at minute %d", tag, index)
			}
			minutes[index] = rule.Type
		}
	}
	for minute, period := range points {
		minutes[minute] = period
	}
	return minutes
}

func assertDayPeriodNames(t *testing.T, locale string, got date.Gregorian, calendar cldr.Calendar) {
	t.Helper()
	want := make(map[string][3]string)
	for index, period := range []string{"midnight", "am", "noon", "pm", "morning1", "morning2", "afternoon1", "afternoon2", "evening1", "evening2", "night1", "night2"} {
		var names [3]string
		for widthIndex, width := range []string{"wide", "abbreviated", "narrow"} {
			values := calendar.Names[cldr.CalendarNameKey{Width: width, Context: "format"}].DayPeriods
			if index < len(values) {
				names[widthIndex] = values[index]
			}
		}
		if period == "am" || period == "pm" || names != [3]string{} {
			want[period] = names
		}
	}
	actual := map[string][3]string{
		"am": {got.DayPeriods.AM.Wide, got.DayPeriods.AM.Abbr, got.DayPeriods.AM.Narrow},
		"pm": {got.DayPeriods.PM.Wide, got.DayPeriods.PM.Abbr, got.DayPeriods.PM.Narrow},
	}
	for period, names := range got.DayPeriods.Flex {
		actual[period] = [3]string{names.Wide, names.Abbr, names.Narrow}
	}
	if !maps.Equal(actual, want) {
		t.Errorf("GregorianFor(%q).DayPeriods = %v, want %v", locale, actual, want)
	}
}

func TestDateGregorianCalendarRoutes(t *testing.T) {
	t.Parallel()

	dates := extract.Dates{
		"en": {
			Calendars: map[string]cldr.Calendar{
				dateCLDRGregorianCalendar: {},
			},
		},
		"fr": {
			Calendars: map[string]cldr.Calendar{
				"buddhist": {},
			},
		},
		"zh": {
			Calendars: map[string]cldr.Calendar{
				dateCLDRGregorianCalendar: {},
				"chinese":                 {},
			},
		},
	}

	gotLocales := dateGregorianLocales(dates)
	assertStringSliceEqual(t, "dateGregorianLocales()", gotLocales, []string{"en", "zh"})
	if _, err := dateSupportedCalendars(dates); err == nil {
		t.Fatal("dateSupportedCalendars() accepted calendars without encoder routes")
	}

	gotCalendars, err := dateSupportedCalendars(extract.Dates{
		"en": {Calendars: map[string]cldr.Calendar{dateCLDRGregorianCalendar: {}}},
	})
	if err != nil {
		t.Fatalf("dateSupportedCalendars(gregorian): %v", err)
	}
	assertStringSliceEqual(t, "dateSupportedCalendars(gregorian)", gotCalendars, []string{dateSupportedGregorianCalendar, dateSupportedISO8601Calendar})

	gotCalendars, err = dateSupportedCalendars(nil)
	if err != nil {
		t.Fatalf("dateSupportedCalendars(nil): %v", err)
	}
	if gotCalendars != nil {
		t.Fatalf("dateSupportedCalendars(nil) = %#v, want nil", gotCalendars)
	}
}

func loadDateTestInput(t *testing.T) extract.Dates {
	t.Helper()

	input := loadRoundTripSource(t)
	return extract.ExtractDates(input.source.Dates, input.profile)
}

// expectedGregorian mirrors the production date.GregorianFor projection over the
// extract Calendar, so the round-trip compares the accessor result against an
// independent computation from the source rows.
func expectedGregorian(cal cldr.Calendar) date.Gregorian {
	var g date.Gregorian
	names := func(width, context string) cldr.CalendarNames {
		return cal.Names[cldr.CalendarNameKey{Width: width, Context: context}]
	}
	wf := names("wide", "format")
	af := names("abbreviated", "format")
	nf := names("narrow", "format")
	ws := names("wide", "stand-alone")
	as := names("abbreviated", "stand-alone")
	ns := names("narrow", "stand-alone")
	copy(g.Eras.Wide[:], wf.Eras)
	copy(g.Eras.Abbr[:], af.Eras)
	copy(g.Eras.Narrow[:], nf.Eras)
	copy(g.Months.Wide[:], wf.Months)
	copy(g.Months.Abbr[:], af.Months)
	copy(g.Months.Narrow[:], nf.Months)
	copy(g.Months.StandWide[:], ws.Months)
	copy(g.Months.StandAbbr[:], as.Months)
	copy(g.Months.StandNarrow[:], ns.Months)
	copy(g.Weekdays.Wide[:], wf.Weekdays)
	copy(g.Weekdays.Abbr[:], af.Weekdays)
	copy(g.Weekdays.Narrow[:], nf.Weekdays)
	copy(g.Weekdays.StandWide[:], ws.Weekdays)
	copy(g.Weekdays.StandAbbr[:], as.Weekdays)
	copy(g.Weekdays.StandNarrow[:], ns.Weekdays)
	g.DateFormats = styleArrayFromMap(cal.DateFormats)
	g.TimeFormats = styleArrayFromMap(cal.TimeFormats)
	g.DateTimeFormats = styleArrayFromMap(cal.DateTimeFormats)
	g.DateTimeAtFormats = styleArrayFromMap(cal.DateTimeAtFormats)
	g.AvailableFormats = cal.AvailableFormats
	g.IntervalFormats = cal.IntervalFormats.BySkeleton
	g.IntervalFallback = cal.IntervalFormats.FallbackPattern
	g.AppendItems = cal.AppendItems
	return g
}

func styleArrayFromMap(values map[string]string) [4]string {
	return [4]string{values["full"], values["long"], values["medium"], values["short"]}
}

func assertGregorianEqual(t *testing.T, locale string, got, want date.Gregorian) {
	t.Helper()
	if got.Eras != want.Eras {
		t.Errorf("GregorianFor(%q).Eras = %+v, want %+v", locale, got.Eras, want.Eras)
	}
	if got.Months != want.Months {
		t.Errorf("GregorianFor(%q).Months mismatch", locale)
	}
	if got.Weekdays != want.Weekdays {
		t.Errorf("GregorianFor(%q).Weekdays mismatch", locale)
	}
	if got.DateFormats != want.DateFormats || got.TimeFormats != want.TimeFormats ||
		got.DateTimeFormats != want.DateTimeFormats || got.DateTimeAtFormats != want.DateTimeAtFormats {
		t.Errorf("GregorianFor(%q) style formats mismatch", locale)
	}
	if got.IntervalFallback != want.IntervalFallback {
		t.Errorf("GregorianFor(%q).IntervalFallback = %q, want %q", locale, got.IntervalFallback, want.IntervalFallback)
	}
	if !maps.Equal(got.AvailableFormats, want.AvailableFormats) {
		t.Errorf("GregorianFor(%q).AvailableFormats mismatch", locale)
	}
	if !maps.Equal(got.AppendItems, want.AppendItems) {
		t.Errorf("GregorianFor(%q).AppendItems mismatch", locale)
	}
	if !maps.EqualFunc(got.IntervalFormats, want.IntervalFormats, func(a, b map[string]string) bool {
		return maps.Equal(a, b)
	}) {
		t.Errorf("GregorianFor(%q).IntervalFormats mismatch", locale)
	}
}
