package locale

import (
	"encoding/json/v2"
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
		{
			name:       "explicit unknown region does not use subdivision",
			in:         "en-ZZ-u-sd-usca",
			hourCycles: []string{"h23"},
			firstDay:   time.Monday,
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
	for _, tag := range []string{"en-US", "und", "tlh", "de-DE"} {
		loc := parseLocaleForTest(tag)
		got := loc.GetCollations()
		if !slices.Equal(got, []string{"emoji", "eor"}) {
			t.Fatalf("GetCollations(%s) = %#v, want emoji/eor", tag, got)
		}
		got[0] = "changed"
		if !slices.Equal(loc.GetCollations(), []string{"emoji", "eor"}) {
			t.Fatal("collation result aliases shared state")
		}
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
	for _, tag := range []string{"en-ZZ", "en-Zzzz-ZZ", "en-XX"} {
		if got := parseLocaleForTest(tag).GetTimeZones(); got == nil || len(got) != 0 {
			t.Errorf("GetTimeZones(%s) = %#v, want a present empty list", tag, got)
		}
	}
}

func TestTextInfo(t *testing.T) {
	t.Parallel()

	for _, tag := range []string{"und-Sogd", "und-Phli", "und-Narb"} {
		assertTextDirection(t, tag, "rtl")
	}
	assertTextDirection(t, "ar-SA", "rtl")
	assertTextDirection(t, "und-Arab", "rtl")
	assertTextDirection(t, "yi", "rtl")
	assertTextDirection(t, "dv", "rtl")
	assertTextDirection(t, "en-US", "ltr")

	first := parseLocaleForTest("en").GetTextInfo()
	second := parseLocaleForTest("en").GetTextInfo()
	if first.Direction == nil || second.Direction == nil {
		t.Fatalf("GetTextInfo(en).Direction = %v, %v; want two non-nil values", first.Direction, second.Direction)
	}
	*first.Direction = "rtl"
	if got := *second.Direction; got != "ltr" {
		t.Fatalf("independent GetTextInfo direction = %q after mutating prior result, want ltr", got)
	}
}

func assertTextDirection(t *testing.T, tag, want string) {
	t.Helper()
	got := parseLocaleForTest(tag).GetTextInfo().Direction
	if got == nil || *got != want {
		t.Errorf("%s direction = %v, want %q", tag, got, want)
	}
}

func TestTextInfoUnknownDirectionIsAbsent(t *testing.T) {
	t.Parallel()
	for _, tag := range []string{"und-Brai", "en-Zzzz", "ar-Zzzz-ZZ"} {
		t.Run(tag, func(t *testing.T) {
			t.Parallel()
			info := parseLocaleForTest(tag).GetTextInfo()
			if info.Direction != nil {
				t.Errorf("GetTextInfo(%s).Direction = %q, want nil", tag, *info.Direction)
			}
			raw, err := json.Marshal(info)
			if err != nil {
				t.Fatal(err)
			}
			if got := string(raw); got != `{}` {
				t.Errorf("GetTextInfo(%s) JSON = %s, want {}", tag, got)
			}
		})
	}
}

func TestHourCycleLanguageRegionPreferences(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		tag  string
		want []string
	}{
		{"fr-CA", []string{"h23", "h12"}}, {"en-CA", []string{"h12", "h23"}},
		{"ku-SY", []string{"h23"}}, {"en-IL", []string{"h23", "h12"}},
		{"fr-CA-u-rg-uszzzz", []string{"h12", "h23"}},
		{"fr-CA-u-rg-zzzzzz", []string{"h23", "h12"}},
		{"fr-u-sd-caqc", []string{"h23", "h12"}},
		{"fr-US-u-sd-caqc", []string{"h12", "h23"}},
		{"fr-CA-u-hc-h11", []string{"h11"}},
		{"en-XX", []string{"h23"}},
	} {
		t.Run(tc.tag, func(t *testing.T) {
			t.Parallel()
			loc := parseLocaleForTest(tc.tag)
			got := loc.GetHourCycles()
			if !slices.Equal(got, tc.want) {
				t.Fatalf("GetHourCycles() = %v, want %v", got, tc.want)
			}
			got[0] = "modified"
			if !slices.Equal(loc.GetHourCycles(), tc.want) {
				t.Fatal("GetHourCycles exposed shared state")
			}
		})
	}
}

func TestSubdivisionRegionPreference(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		tag       string
		canonical string
		firstDay  time.Weekday
	}{
		{"en-u-sd-gb", "en-u-gb-sd", time.Sunday},
		{"en-u-sd-gba", "en-u-sd-gba", time.Monday},
		{"en-u-sd-gbabcd", "en-u-sd-gbabcd", time.Monday},
		{"en-u-sd-gbabcde", "en-u-sd-gbabcde", time.Sunday},
		{"en-u-sd-001", "en-u-sd-001", time.Sunday},
		{"en-u-sd-001a", "en-u-sd-001a", time.Monday},
		{"en-u-sd-001abcd", "en-u-sd-001abcd", time.Monday},
		{"en-u-sd-001abcde", "en-u-sd-001abcde", time.Sunday},
		{"en-US-u-sd-gba", "en-US-u-sd-gba", time.Sunday},
		{"en-US-u-rg-gba", "en-US-u-rg-gba", time.Monday},
		{"en-US-u-rg-gbabcde", "en-US-u-rg-gbabcde", time.Sunday},
		{"en-US-u-rg-zzzzzz", "en-US-u-rg-zzzzzz", time.Sunday},
	} {
		t.Run(tc.tag, func(t *testing.T) {
			t.Parallel()
			loc, err := Parse(tc.tag)
			if err != nil {
				t.Fatal(err)
			}
			if got := loc.GetWeekInfo().FirstDay; got != tc.firstDay {
				t.Errorf("firstDay = %s; want %s", got, tc.firstDay)
			}
			if got := loc.String(); got != tc.canonical {
				t.Errorf("String() = %q; want %q", got, tc.canonical)
			}
		})
	}
}
