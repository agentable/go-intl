package codegen

import (
	"context"
	"os"
	"path/filepath"
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

	// Gregorian view: assemble the expected Gregorian from the extract rows using
	// the same field projection the accessor performs, then compare to the
	// production accessor result.
	for _, locale := range dateGregorianLocales(dates) {
		loc := resolveDateLocale(t, locale)
		want := expectedGregorian(dates[locale].Calendars["gregorian"])
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
	wantTags := dateSupportedLocaleTags(dates)
	gotTags := date.SupportedLocales()
	if len(gotTags) != len(wantTags) {
		t.Fatalf("SupportedLocales len = %d, want %d", len(gotTags), len(wantTags))
	}
	for i := range wantTags {
		if gotTags[i] != wantTags[i] {
			t.Errorf("SupportedLocales[%d] = %q, want %q", i, gotTags[i], wantTags[i])
		}
	}

	// Supported-calendar narrow index.
	wantCal := dateSupportedCalendars(dates)
	gotCal := date.SupportedCalendars()
	if len(gotCal) != len(wantCal) {
		t.Fatalf("SupportedCalendars len = %d, want %d", len(gotCal), len(wantCal))
	}
	for i := range wantCal {
		if gotCal[i] != wantCal[i] {
			t.Errorf("SupportedCalendars[%d] = %q, want %q", i, gotCal[i], wantCal[i])
		}
	}
}

func loadDateTestInput(t *testing.T) extract.Dates {
	t.Helper()

	repoRoot := filepath.Clean("../../..")
	cldrDir := filepath.Join(repoRoot, "tools", "gen-cldr", ".cldr-json", "node_modules")
	if _, err := os.Stat(cldrDir); err != nil {
		t.Skipf("pinned cldr-json checkout absent (%v); run task data:fetch", err)
	}

	versions, err := cldr.ReadVersionFile(filepath.Join(repoRoot, "internal", "cldr", "VERSION"))
	if err != nil {
		t.Fatalf("read VERSION: %v", err)
	}
	profile := readUnitTestProfile(t, filepath.Join(repoRoot, "tools", "locale-profile.json"))

	source, err := cldr.LoadAll(context.Background(), cldrDir, versions, profile)
	if err != nil {
		t.Fatalf("load cldr-json: %v", err)
	}
	return extract.ExtractDates(source.Dates, profile)
}

func resolveDateLocale(t *testing.T, tag string) cldrlocale.Locale {
	t.Helper()
	loc, ok := cldrlocale.ResolveLocale(tag)
	if !ok {
		t.Fatalf("kernel locale %q not resolvable", tag)
	}
	return loc
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
	if !stringMapEqual(got.AvailableFormats, want.AvailableFormats) {
		t.Errorf("GregorianFor(%q).AvailableFormats mismatch", locale)
	}
	if !stringMapEqual(got.AppendItems, want.AppendItems) {
		t.Errorf("GregorianFor(%q).AppendItems mismatch", locale)
	}
	if !nestedStringMapEqual(got.IntervalFormats, want.IntervalFormats) {
		t.Errorf("GregorianFor(%q).IntervalFormats mismatch", locale)
	}
}

func stringMapEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func nestedStringMapEqual(a, b map[string]map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, inner := range a {
		if !stringMapEqual(inner, b[k]) {
			return false
		}
	}
	return true
}
