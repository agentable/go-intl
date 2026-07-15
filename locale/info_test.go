package locale

import (
	"slices"
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/testcontract"
)

func TestGetWeekInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		in             string
		firstDayOption string
		firstDay       time.Weekday
		weekend        []time.Weekday
	}{
		{name: "US default", in: "en-US", firstDay: time.Sunday, weekend: []time.Weekday{time.Saturday, time.Sunday}},
		{name: "German default", in: "de-DE", firstDay: time.Monday, weekend: []time.Weekday{time.Saturday, time.Sunday}},
		{name: "tag keyword", in: "en-US-u-fw-mon", firstDay: time.Monday, weekend: []time.Weekday{time.Saturday, time.Sunday}},
		{name: "numeric option 2", in: "en-US", firstDayOption: "2", firstDay: time.Tuesday, weekend: []time.Weekday{time.Saturday, time.Sunday}},
		{name: "numeric option 3", in: "en-US", firstDayOption: "3", firstDay: time.Wednesday, weekend: []time.Weekday{time.Saturday, time.Sunday}},
		{name: "numeric option 4", in: "en-US", firstDayOption: "4", firstDay: time.Thursday, weekend: []time.Weekday{time.Saturday, time.Sunday}},
		{name: "numeric option 5", in: "en-US", firstDayOption: "5", firstDay: time.Friday, weekend: []time.Weekday{time.Saturday, time.Sunday}},
		{name: "numeric option 6", in: "en-US", firstDayOption: "6", firstDay: time.Saturday, weekend: []time.Weekday{time.Saturday, time.Sunday}},
		{name: "numeric option 7", in: "en-US", firstDayOption: "7", firstDay: time.Sunday, weekend: []time.Weekday{time.Saturday, time.Sunday}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			opts := Options{}
			if tc.firstDayOption != "" {
				opts.FirstDayOfWeek = stringPtr(tc.firstDayOption)
			}
			loc, err := New(tc.in, opts)
			if err != nil {
				t.Fatal(err)
			}
			got := loc.GetWeekInfo()
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
			calendars: []string{"gregory"},
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
		{
			name:       "invalid sd falls back to likely region",
			in:         "und-u-sd-1azzzz",
			hourCycles: []string{"h12", "h23"},
			firstDay:   time.Sunday,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loc := parseLocaleForTest(tc.in)
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

	loc := parseLocaleForTest("en-US")
	if got := loc.GetCalendars(); !slices.Equal(got, []string{"gregory"}) {
		t.Fatalf("GetCalendars() = %#v", got)
	}
	withCalendar, err := New("en-US", Options{Calendar: stringPtr("buddhist")})
	if err != nil {
		t.Fatal(err)
	}
	if got := withCalendar.GetCalendars(); !slices.Equal(got, []string{"buddhist"}) {
		t.Fatalf("GetCalendars() with calendar = %#v", got)
	}
	if got := parseLocaleForTest("und").GetCollations(); len(got) != 0 {
		t.Fatalf("GetCollations(und) = %#v, want no unrelated backend collations", got)
	}
	if got := loc.GetCollations(); len(got) != 0 {
		t.Fatalf("GetCollations(en-US) = %#v, want no German collation tailoring", got)
	}
	if got := parseLocaleForTest("tlh").GetCollations(); len(got) != 0 {
		t.Fatalf("GetCollations(tlh) = %#v, want no unmatched backend collations", got)
	}
	if got := parseLocaleForTest("de-DE").GetCollations(); !slices.Contains(got, "phonebk") {
		t.Fatalf("GetCollations(de-DE) = %#v, want locale-scoped phonebk", got)
	} else {
		testcontract.AssertStringSliceSortedUnique(t, "GetCollations(de-DE)", got)
	}
	withCollation, err := New("en-US", Options{Collation: stringPtr("phonebk")})
	if err != nil {
		t.Fatal(err)
	}
	if got := withCollation.GetCollations(); !slices.Equal(got, []string{"phonebk"}) {
		t.Fatalf("GetCollations() with collation = %#v", got)
	}
	if got := loc.GetHourCycles(); !slices.Equal(got, []string{"h12", "h23"}) {
		t.Fatalf("GetHourCycles() = %#v", got)
	}
	withHourCycle, err := New("en-US", Options{HourCycle: stringPtr("h23")})
	if err != nil {
		t.Fatal(err)
	}
	if got := withHourCycle.GetHourCycles(); !slices.Equal(got, []string{"h23"}) {
		t.Fatalf("GetHourCycles() with hourCycle = %#v", got)
	}
	withNumberingSystem, err := New("en-US", Options{NumberingSystem: stringPtr("arab")})
	if err != nil {
		t.Fatal(err)
	}
	if got := withNumberingSystem.GetNumberingSystems(); !slices.Equal(got, []string{"arab"}) {
		t.Fatalf("GetNumberingSystems() = %#v", got)
	}
	if got := parseLocaleForTest("fr-FR").GetNumberingSystems(); !slices.Equal(got, []string{"latn"}) {
		t.Fatalf("GetNumberingSystems() fallback = %#v, want latn", got)
	}
	if got := parseLocaleForTest("und").GetNumberingSystems(); !slices.Equal(got, []string{"latn"}) {
		t.Fatalf("GetNumberingSystems(und) fallback = %#v, want latn", got)
	}
	if got := parseLocaleForTest("en-US").GetTimeZones(); !slices.Contains(got, "America/New_York") || !slices.Contains(got, "America/Los_Angeles") {
		t.Fatalf("GetTimeZones(en-US) = %#v, want canonical US zones", got)
	}
	testcontract.AssertStringSliceSortedUnique(t, "GetTimeZones(en-US)", parseLocaleForTest("en-US").GetTimeZones())
	if got := parseLocaleForTest("en-GB").GetTimeZones(); !slices.Equal(got, []string{"Europe/London"}) {
		t.Fatalf("GetTimeZones(en-GB) = %#v, want Europe/London", got)
	}
	if got := parseLocaleForTest("zh-CN").GetTimeZones(); !slices.Equal(got, []string{"Asia/Shanghai", "Asia/Urumqi"}) {
		t.Fatalf("GetTimeZones(zh-CN) = %#v, want CLDR China zones", got)
	}
	if got := parseLocaleForTest("en-IN").GetTimeZones(); !slices.Equal(got, []string{"Asia/Kolkata"}) {
		t.Fatalf("GetTimeZones(en-IN) = %#v, want IANA primary India zone", got)
	}
	if got := parseLocaleForTest("en-CA").GetTimeZones(); len(got) < 20 || !slices.Contains(got, "America/Toronto") || !slices.Contains(got, "America/Vancouver") {
		t.Fatalf("GetTimeZones(en-CA) = %#v, want full IANA Canadian projection", got)
	}
	if got := parseLocaleForTest("ar").GetTimeZones(); got != nil {
		t.Fatalf("GetTimeZones(ar) = %#v, want nil without region", got)
	}
}

func TestTextInfo(t *testing.T) {
	t.Parallel()

	if got := parseLocaleForTest("ar-SA").GetTextInfo().Direction; got != "rtl" {
		t.Fatalf("ar-SA direction = %q, want rtl", got)
	}
	if got := parseLocaleForTest("und-Arab").GetTextInfo().Direction; got != "rtl" {
		t.Fatalf("und-Arab direction = %q, want rtl", got)
	}
	if got := parseLocaleForTest("yi").GetTextInfo().Direction; got != "rtl" {
		t.Fatalf("yi direction = %q, want rtl from likely subtags", got)
	}
	if got := parseLocaleForTest("dv").GetTextInfo().Direction; got != "rtl" {
		t.Fatalf("dv direction = %q, want rtl from likely subtags", got)
	}
	if got := parseLocaleForTest("en-US").GetTextInfo().Direction; got != "ltr" {
		t.Fatalf("en-US direction = %q, want ltr", got)
	}
}
