package datetimeformat

import (
	"testing"
	"time"
)

func TestGregoryLocalTimeEraDisplayYear(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		year        int
		era         string
		displayYear int
	}{
		{name: "ad", year: 2026, era: "AD", displayYear: 2026},
		{name: "year-zero", year: 0, era: "BC", displayYear: 1},
		{name: "negative-year", year: -44, era: "BC", displayYear: 45},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			local := gregoryLocalTime(time.Date(tt.year, time.March, 15, 12, 34, 56, 789, time.UTC))
			if local.Era != tt.era || local.displayYear() != tt.displayYear {
				t.Fatalf("gregoryLocalTime(%d).era/displayYear = %q/%d, want %q/%d", tt.year, local.Era, local.displayYear(), tt.era, tt.displayYear)
			}
		})
	}
}

func TestGregoryTimeInLocationFreezesConvertedFields(t *testing.T) {
	t.Parallel()

	loc := time.FixedZone("UTC+02", 2*60*60)
	instant := time.Date(2026, time.January, 1, 23, 30, 5, 7, time.UTC)
	converted, local := gregoryTimeInLocation(instant, loc)
	want := instant.In(loc)
	if !converted.Equal(want) || !local.Time.Equal(want) {
		t.Fatalf("gregoryTimeInLocation() time = %v/%v, want %v", converted, local.Time, want)
	}
	if local.Year != 2026 || local.Month != time.January || local.Day != 2 ||
		local.Hour != 1 || local.Minute != 30 || local.Second != 5 || local.Nanosecond != 7 ||
		local.Weekday != time.Friday || local.Era != "AD" {
		t.Fatalf("gregoryTimeInLocation() fields = %#v, want converted UTC+02 Gregorian fields", local)
	}
}
