package ecma402dtf

import "testing"

func TestParseSkeletonDateFields(t *testing.T) {
	t.Parallel()

	got := Parse("yMMMd", "MMM d, y", nil, "")
	if got.Skeleton != "yMMMd" {
		t.Fatalf("Skeleton = %q, want %q", got.Skeleton, "yMMMd")
	}
	if got.Pattern != "MMM d, y" {
		t.Fatalf("Pattern = %q, want %q", got.Pattern, "MMM d, y")
	}
	if got.Year != NumericNumeric {
		t.Fatalf("Year = %q, want %q", got.Year, NumericNumeric)
	}
	if got.Month != FieldShort {
		t.Fatalf("Month = %q, want %q", got.Month, FieldShort)
	}
	if got.Day != NumericNumeric {
		t.Fatalf("Day = %q, want %q", got.Day, NumericNumeric)
	}
}

func TestParseSkeletonTimeFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		skeleton  string
		pattern   string
		hour      NumericStyle
		hourCycle HourCycle
		minute    NumericStyle
		second    NumericStyle
	}{
		{
			name:      "hms",
			skeleton:  "hms",
			pattern:   "h:mm:ss a",
			hour:      NumericNumeric,
			hourCycle: HourCycleH12,
			minute:    NumericNumeric,
			second:    NumericNumeric,
		},
		{
			name:      "HHmmss",
			skeleton:  "HHmmss",
			pattern:   "HH:mm:ss",
			hour:      Numeric2Digit,
			hourCycle: HourCycleH23,
			minute:    Numeric2Digit,
			second:    Numeric2Digit,
		},
		{
			name:      "Kms",
			skeleton:  "Kms",
			pattern:   "K:mm:ss",
			hour:      NumericNumeric,
			hourCycle: HourCycleH11,
			minute:    NumericNumeric,
			second:    NumericNumeric,
		},
		{
			name:      "kkmmss",
			skeleton:  "kkmmss",
			pattern:   "kk:mm:ss",
			hour:      Numeric2Digit,
			hourCycle: HourCycleH24,
			minute:    Numeric2Digit,
			second:    Numeric2Digit,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := Parse(test.skeleton, test.pattern, nil, "")
			if got.Hour != test.hour {
				t.Fatalf("Hour = %q, want %q", got.Hour, test.hour)
			}
			if got.HourCycle != test.hourCycle {
				t.Fatalf("HourCycle = %q, want %q", got.HourCycle, test.hourCycle)
			}
			if got.Minute != test.minute {
				t.Fatalf("Minute = %q, want %q", got.Minute, test.minute)
			}
			if got.Second != test.second {
				t.Fatalf("Second = %q, want %q", got.Second, test.second)
			}
		})
	}
}

func TestParseSkeletonHourCycleOverride(t *testing.T) {
	t.Parallel()

	got := Parse("h", "h a", nil, HourCycleH23)
	if got.HourCycle != HourCycleH23 {
		t.Fatalf("HourCycle = %q, want %q", got.HourCycle, HourCycleH23)
	}
}

func TestParseSkeletonAdditionalFields(t *testing.T) {
	t.Parallel()

	got := Parse("GGGGEEEEBBBBSSSZZZZ", "GGGG EEEE BBBB SSS ZZZZ", nil, "")
	if got.Era != FieldLong {
		t.Fatalf("Era = %q, want %q", got.Era, FieldLong)
	}
	if got.Weekday != FieldLong {
		t.Fatalf("Weekday = %q, want %q", got.Weekday, FieldLong)
	}
	if got.DayPeriod != FieldLong {
		t.Fatalf("DayPeriod = %q, want %q", got.DayPeriod, FieldLong)
	}
	if got.FractionalSecondDigits != 3 {
		t.Fatalf("FractionalSecondDigits = %d, want 3", got.FractionalSecondDigits)
	}
	if got.TimeZoneName != TimeZoneNameLongOffset {
		t.Fatalf("TimeZoneName = %q, want %q", got.TimeZoneName, TimeZoneNameLongOffset)
	}
}

func TestParseSkeletonGenericTimeZoneName(t *testing.T) {
	t.Parallel()

	got := Parse("vvvv", "vvvv", nil, "")
	if got.TimeZoneName != TimeZoneNameLongGeneric {
		t.Fatalf("TimeZoneName = %q, want %q", got.TimeZoneName, TimeZoneNameLongGeneric)
	}
}

func TestParseSkeletonIgnoresQuotedLiterals(t *testing.T) {
	t.Parallel()

	got := Parse("'Year:' y", "'Year:' y", nil, "")
	if got.Year != NumericNumeric {
		t.Fatalf("Year = %q, want %q", got.Year, NumericNumeric)
	}
	if got.Era != "" {
		t.Fatalf("Era = %q, want empty", got.Era)
	}
	if got.DayPeriod != "" {
		t.Fatalf("DayPeriod = %q, want empty", got.DayPeriod)
	}
}

func TestParseSkeletonHandlesEscapedQuoteLiterals(t *testing.T) {
	t.Parallel()

	got := Parse("'it''s' y", "'it''s' y", nil, "")
	if got.Second != "" {
		t.Fatalf("Second = %q, want empty", got.Second)
	}
	if got.Year != NumericNumeric {
		t.Fatalf("Year = %q, want %q", got.Year, NumericNumeric)
	}
}
