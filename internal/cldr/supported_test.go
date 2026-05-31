package cldr

import (
	"maps"
	"slices"
	"testing"

	"github.com/agentable/go-intl/internal/numbering"
)

func TestSupportedLocalesComeFromGeneratedData(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		tags []string
		has  func(Locale) bool
	}{
		{name: "number", tags: NumberSupportedLocales(), has: func(loc Locale) bool {
			_, ok := numbersByLocale[loc]
			return ok
		}},
		{name: "date", tags: DateSupportedLocales(), has: func(loc Locale) bool {
			_, ok := datesByLocale[loc]
			return ok
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if len(tc.tags) == 0 {
				t.Fatalf("%s supported locales is empty", tc.name)
			}
			for _, tag := range tc.tags {
				loc, ok := localeIndex[tag]
				if !ok {
					t.Fatalf("%s supported locale %q is missing from localeIndex", tc.name, tag)
				}
				if loc == Undefined {
					t.Fatalf("%s supported locales includes Undefined", tc.name)
				}
				if !tc.has(loc) {
					t.Fatalf("%s supported locale %q has no generated data", tc.name, tag)
				}
			}
		})
	}
}

func TestSupportedCalendarsComeFromGeneratedDateData(t *testing.T) {
	t.Parallel()

	seen := map[string]bool{}
	for _, calendars := range dateDataByLocale() {
		for calendar := range calendars {
			if calendar == "gregorian" {
				seen["gregory"] = true
				continue
			}
			seen[calendar] = true
		}
	}

	got := SupportedCalendars()
	for _, calendar := range got {
		if calendar == "iso8601" {
			continue
		}
		if !seen[calendar] {
			t.Fatalf("SupportedCalendars() includes %q without generated date data: %v", calendar, got)
		}
	}
	if !slices.Contains(got, "gregory") {
		t.Fatalf("SupportedCalendars() = %v, want generated gregory calendar", got)
	}
	if !slices.Contains(got, "iso8601") {
		t.Fatalf("SupportedCalendars() = %v, want ECMA-402 iso8601 bridge", got)
	}
	if slices.Contains(got, "buddhist") {
		t.Fatalf("SupportedCalendars() = %v, want no unsupported Buddhist calendar", got)
	}
}

func TestDateLocaleDataCalendarKeywordUsesSupportedCalendars(t *testing.T) {
	t.Parallel()

	got := DateLocaleData{}.For("en-US", "ca")
	if !slices.Equal(got, SupportedCalendars()) {
		t.Fatalf("DateLocaleData.For(ca) = %v, want SupportedCalendars()", got)
	}
}

func TestSupportedValuesComeFromGeneratedSemanticIndexes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  []string
		want []string
	}{
		{
			name: "currencies",
			got:  SupportedCurrencies(),
			want: supportedCurrencyValuesFromGeneratedData(),
		},
		{
			name: "numbering systems",
			got:  SupportedNumberingSystems(),
			want: supportedNumberingSystemValuesFromGeneratedData(),
		},
		{
			name: "time zones",
			got:  SupportedTimeZones(),
			want: supportedTimeZoneValuesFromGeneratedData(),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if !slices.IsSorted(tc.got) {
				t.Fatalf("%s = %v, want sorted values", tc.name, tc.got)
			}
			if compact := slices.Compact(slices.Clone(tc.got)); !slices.Equal(compact, tc.got) {
				t.Fatalf("%s = %v, want unique values", tc.name, tc.got)
			}
			if !slices.Equal(tc.got, tc.want) {
				t.Fatalf("%s = %v, want %v", tc.name, tc.got, tc.want)
			}
		})
	}
}

func supportedCurrencyValuesFromGeneratedData() []string {
	seen := map[string]bool{}
	for code := range currencyFractionData() {
		if code != "DEFAULT" {
			seen[code] = true
		}
	}
	for _, currencies := range currencyDataByLocale() {
		for code := range currencies {
			seen[code] = true
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

func supportedNumberingSystemValuesFromGeneratedData() []string {
	seen := map[string]bool{}
	for _, numberingSystem := range numbering.SimpleNumberingSystems {
		seen[numberingSystem] = true
	}
	for _, data := range numberDataByLocale() {
		if data.defaultNS != "" {
			seen[data.defaultNS] = true
		}
		for numberingSystem := range data.symbols {
			seen[numberingSystem] = true
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

func supportedTimeZoneValuesFromGeneratedData() []string {
	seen := map[string]bool{}
	for zone := range zoneMetazoneData() {
		if zone == "Etc/Unknown" {
			continue
		}
		seen[CanonicalTimeZoneLink(zone)] = true
	}
	return slices.Sorted(maps.Keys(seen))
}

func BenchmarkAccessorSupportedValues(b *testing.B) {
	_ = SupportedCalendars()
	_ = SupportedCurrencies()
	_ = SupportedNumberingSystems()
	_ = SupportedTimeZones()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = SupportedCalendars()
		_ = SupportedCurrencies()
		_ = SupportedNumberingSystems()
		_ = SupportedTimeZones()
	}
}
