package locale

import (
	"slices"
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/cldr"

	"golang.org/x/text/language"
)

func TestGetWeekInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in       string
		firstDay time.Weekday
		weekend  []time.Weekday
	}{
		{in: "en-US", firstDay: time.Sunday, weekend: []time.Weekday{time.Saturday, time.Sunday}},
		{in: "de-DE", firstDay: time.Monday, weekend: []time.Weekday{time.Saturday, time.Sunday}},
		{in: "en-US-u-fw-mon", firstDay: time.Monday, weekend: []time.Weekday{time.Saturday, time.Sunday}},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			got := MustParse(tc.in).GetWeekInfo()
			if got.FirstDay != tc.firstDay || !slices.Equal(got.Weekend, tc.weekend) {
				t.Fatalf("GetWeekInfo() = %#v", got)
			}
		})
	}
}

func TestLocaleInfoRegionPreference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		in         string
		calendars  []string
		hourCycles []string
		firstDay   time.Weekday
	}{
		{
			name:      "rg overrides calendar preference",
			in:        "en-US-u-rg-thzzzz",
			calendars: []string{"buddhist", "gregory"},
			firstDay:  time.Sunday,
		},
		{
			name:       "rg overrides hour cycle and week preference",
			in:         "en-US-u-rg-gbzzzz",
			hourCycles: []string{"h23", "h12"},
			firstDay:   time.Monday,
		},
		{
			name:       "sd supplies region when base locale has none",
			in:         "und-u-sd-gbeng",
			hourCycles: []string{"h23", "h12"},
			firstDay:   time.Monday,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loc := MustParse(tc.in)
			if tc.calendars != nil {
				if got := loc.GetCalendars(); !slices.Equal(got, tc.calendars) {
					t.Fatalf("GetCalendars() = %#v, want %#v", got, tc.calendars)
				}
			}
			if tc.hourCycles != nil {
				if got := loc.GetHourCycles(); !slices.Equal(got, tc.hourCycles) {
					t.Fatalf("GetHourCycles() = %#v, want %#v", got, tc.hourCycles)
				}
			}
			if got := loc.GetWeekInfo().FirstDay; got != tc.firstDay {
				t.Fatalf("GetWeekInfo().FirstDay = %v, want %v", got, tc.firstDay)
			}
		})
	}
}

func TestLocaleInfoGetters(t *testing.T) {
	t.Parallel()

	loc := MustParse("en-US")
	if got := loc.GetCalendars(); !slices.Equal(got, []string{"gregory"}) {
		t.Fatalf("GetCalendars() = %#v", got)
	}
	withCalendar, err := New(language.MustParse("en-US"), Options{Calendar: "buddhist"})
	if err != nil {
		t.Fatal(err)
	}
	if got := withCalendar.GetCalendars(); !slices.Equal(got, []string{"buddhist"}) {
		t.Fatalf("GetCalendars() with calendar = %#v", got)
	}
	if got := MustParse("und").GetCollations(); !slices.Equal(got, []string{"emoji", "eor"}) {
		t.Fatalf("GetCollations(und) = %#v, want root collation fallback", got)
	}
	if got := loc.GetCollations(); !slices.Equal(got, cldr.SupportedCollations()) {
		t.Fatalf("GetCollations(en-US) = %#v, want generated supported collations", got)
	}
	withCollation, err := New(language.MustParse("en-US"), Options{Collation: "phonebk"})
	if err != nil {
		t.Fatal(err)
	}
	if got := withCollation.GetCollations(); !slices.Equal(got, []string{"phonebk"}) {
		t.Fatalf("GetCollations() with collation = %#v", got)
	}
	if got := loc.GetHourCycles(); !slices.Equal(got, []string{"h12", "h23"}) {
		t.Fatalf("GetHourCycles() = %#v", got)
	}
	withHourCycle, err := New(language.MustParse("en-US"), Options{HourCycle: "h23"})
	if err != nil {
		t.Fatal(err)
	}
	if got := withHourCycle.GetHourCycles(); !slices.Equal(got, []string{"h23"}) {
		t.Fatalf("GetHourCycles() with hourCycle = %#v", got)
	}
	withNumberingSystem, err := New(language.MustParse("en-US"), Options{NumberingSystem: "arab"})
	if err != nil {
		t.Fatal(err)
	}
	if got := withNumberingSystem.GetNumberingSystems(); !slices.Equal(got, []string{"arab"}) {
		t.Fatalf("GetNumberingSystems() = %#v", got)
	}
	if got := MustParse("fr-FR").GetNumberingSystems(); !slices.Equal(got, []string{"latn"}) {
		t.Fatalf("GetNumberingSystems() fallback = %#v, want latn", got)
	}
	if got := MustParse("en-US").GetTimeZones(); !slices.Contains(got, "America/New_York") || !slices.Contains(got, "America/Los_Angeles") {
		t.Fatalf("GetTimeZones(en-US) = %#v, want canonical US zones", got)
	}
	if got := MustParse("en-US").GetTimeZones(); !slices.IsSorted(got) {
		t.Fatalf("GetTimeZones(en-US) = %#v, want lexicographic order", got)
	}
	if got := MustParse("en-GB").GetTimeZones(); !slices.Equal(got, []string{"Europe/London"}) {
		t.Fatalf("GetTimeZones(en-GB) = %#v, want Europe/London", got)
	}
	if got := MustParse("zh-CN").GetTimeZones(); !slices.Equal(got, []string{"Asia/Shanghai", "Asia/Urumqi"}) {
		t.Fatalf("GetTimeZones(zh-CN) = %#v, want CLDR China zones", got)
	}
	if got := MustParse("en-IN").GetTimeZones(); !slices.Equal(got, []string{"Asia/Calcutta"}) {
		t.Fatalf("GetTimeZones(en-IN) = %#v, want CLDR India zone", got)
	}
	if got := MustParse("ar").GetTimeZones(); got != nil {
		t.Fatalf("GetTimeZones(ar) = %#v, want nil without region", got)
	}
}

func TestTextInfo(t *testing.T) {
	t.Parallel()

	if got := MustParse("ar-SA").GetTextInfo().Direction; got != "rtl" {
		t.Fatalf("ar-SA direction = %q, want rtl", got)
	}
	if got := MustParse("und-Arab").GetTextInfo().Direction; got != "rtl" {
		t.Fatalf("und-Arab direction = %q, want rtl", got)
	}
	if got := MustParse("yi").GetTextInfo().Direction; got != "rtl" {
		t.Fatalf("yi direction = %q, want rtl from likely subtags", got)
	}
	if got := MustParse("dv").GetTextInfo().Direction; got != "rtl" {
		t.Fatalf("dv direction = %q, want rtl from likely subtags", got)
	}
	if got := MustParse("en-US").GetTextInfo().Direction; got != "ltr" {
		t.Fatalf("en-US direction = %q, want ltr", got)
	}
}
