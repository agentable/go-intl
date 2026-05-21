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
	if !got.PatternHasTimeZoneName {
		t.Fatal("PatternHasTimeZoneName = false, want true")
	}
}

func TestParsePatternHasTimeZoneNameIgnoresQuotedFields(t *testing.T) {
	t.Parallel()

	got := Parse("y", "'z' y", nil, "")
	if got.PatternHasTimeZoneName {
		t.Fatal("PatternHasTimeZoneName = true for quoted z, want false")
	}
	got = Parse("y", "'zone' y O", nil, "")
	if !got.PatternHasTimeZoneName {
		t.Fatal("PatternHasTimeZoneName = false for O field, want true")
	}
}

func TestParseSkeletonNarrowTextFields(t *testing.T) {
	t.Parallel()

	got := Parse("GGGGGEEEEEbbbbb", "GGGGG EEEEE bbbbb", nil, "")
	if got.Era != FieldNarrow {
		t.Fatalf("Era = %q, want %q", got.Era, FieldNarrow)
	}
	if got.Weekday != FieldNarrow {
		t.Fatalf("Weekday = %q, want %q", got.Weekday, FieldNarrow)
	}
	if got.DayPeriod != FieldNarrow {
		t.Fatalf("DayPeriod = %q, want %q", got.DayPeriod, FieldNarrow)
	}
}

func TestParseSkeletonTimeZoneNameWidths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		skeleton string
		want     TimeZoneName
	}{
		{name: "short specific", skeleton: "z", want: TimeZoneNameShort},
		{name: "long specific", skeleton: "zzzz", want: TimeZoneNameLong},
		{name: "short offset", skeleton: "O", want: TimeZoneNameShortOffset},
		{name: "long offset", skeleton: "OOOO", want: TimeZoneNameLongOffset},
		{name: "iso offset", skeleton: "xxxx", want: TimeZoneNameLongOffset},
		{name: "short generic", skeleton: "v", want: TimeZoneNameShortGeneric},
		{name: "long generic", skeleton: "vvvv", want: TimeZoneNameLongGeneric},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Parse(tc.skeleton, tc.skeleton, nil, "")
			if got.TimeZoneName != tc.want {
				t.Fatalf("TimeZoneName = %q, want %q", got.TimeZoneName, tc.want)
			}
		})
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

func TestParseSkeletonAcceptsExtendedYearTokens(t *testing.T) {
	t.Parallel()

	for _, ch := range []byte{'Y', 'u', 'U', 'r'} {
		t.Run(string(ch), func(t *testing.T) {
			t.Parallel()
			got := Parse(string(ch)+string(ch)+"MMMd", "MMM d, y", nil, "")
			if got.Year != Numeric2Digit {
				t.Fatalf("Year for %s%s = %q, want %q", string(ch), string(ch), got.Year, Numeric2Digit)
			}
			got = Parse(string(ch)+"MMMd", "MMM d, y", nil, "")
			if got.Year != NumericNumeric {
				t.Fatalf("Year for %s = %q, want %q", string(ch), got.Year, NumericNumeric)
			}
		})
	}
}

func TestParseSkeletonAcceptsTimeZoneIDToken(t *testing.T) {
	t.Parallel()

	got := Parse("VVVV", "VVVV", nil, "")
	if got.TimeZoneName != TimeZoneNameLongGeneric {
		t.Fatalf("VVVV TimeZoneName = %q, want %q", got.TimeZoneName, TimeZoneNameLongGeneric)
	}
	got = Parse("V", "V", nil, "")
	if got.TimeZoneName != TimeZoneNameShortGeneric {
		t.Fatalf("V TimeZoneName = %q, want %q", got.TimeZoneName, TimeZoneNameShortGeneric)
	}
}

func TestParseSkeletonAcceptsQuarterTokensWithoutSideEffects(t *testing.T) {
	t.Parallel()

	// Quarter tokens are recognized but produce no Formats field today; the
	// active ECMA-402 surface does not expose `quarter`. The test pins this
	// behavior so changes to add a Quarter field are made deliberately.
	got := Parse("yQQQQ", "yQQQQ", nil, "")
	if got.Year != NumericNumeric {
		t.Fatalf("Year = %q, want %q", got.Year, NumericNumeric)
	}
	if got.Month != "" || got.Day != "" || got.Weekday != "" {
		t.Fatalf("Quarter token should not set other fields: %+v", got)
	}
	got = Parse("yqqq", "yqqq", nil, "")
	if got.Year != NumericNumeric {
		t.Fatalf("Year = %q, want %q", got.Year, NumericNumeric)
	}
}
