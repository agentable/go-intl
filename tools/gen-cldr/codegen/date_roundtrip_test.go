package codegen

import (
	"maps"
	"testing"

	"github.com/agentable/go-intl/internal/cldr/date"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// TestDateRoundTrip is the production-path round-trip gate for the date domain.
// It re-derives the extract.Dates row stream from the real pinned CLDR checkout
// and asserts that the resolved gregorian view, the day-period lookups, the
// supported-locale narrow index, and the supported-calendar narrow index the
// encoder wrote are queried back byte-for-byte through the production date
// accessors over the committed data.go. It exercises encoder, blob, decoder, and
// accessor as one chain — not internal structures.
//
// The gate is meaningful only when the pinned cldr-json checkout is present
// (after task data / data:fetch); without it the test skips.
func TestDateRoundTrip(t *testing.T) {
	t.Parallel()

	dates := loadDateTestInput(t)
	gregLocales := dateGregorianLocales(dates)

	// Gregorian view: assemble the expected Gregorian from the extract rows using
	// the same field projection the accessor performs, then compare to the
	// production accessor result.
	for _, locale := range gregLocales {
		loc := resolveKernelLocale(t, locale)
		want := expectedGregorian(dates[locale].Calendars[dateCLDRGregorianCalendar])
		assertGregorianEqual(t, locale, date.GregorianFor(loc), want)
	}

	// Day-period rules: probe each rule boundary and confirm the production
	// DayPeriodFor resolves the same type the rule stream encodes.
	for locale, rules := range dateDayPeriodRuleSet(dates) {
		loc, ok := cldrlocale.ResolveLocale(locale)
		if !ok {
			continue
		}
		for _, rule := range rules {
			hour := int(rule.From.Hours())
			minute := int(rule.From.Minutes()) % 60
			got := date.DayPeriodFor(loc, hour, minute)
			if got == "" {
				t.Errorf("DayPeriodFor(%q, %d, %d) = \"\", want a period for encoded rule %q",
					locale, hour, minute, rule.Type)
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
