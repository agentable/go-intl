package ecma402dtf

import "testing"

func TestMatchBasicSelectsCompleteCandidate(t *testing.T) {
	t.Parallel()

	formats := []Formats{
		Parse("y", "y", nil, ""),
		Parse("MMMM", "MMMM", nil, ""),
		Parse("MMMMy", "MMMM y", nil, ""),
	}
	got := MatchBasic(Options{Year: NumericNumeric, Month: FieldLong}, formats)
	if got.Pattern != "MMMM y" {
		t.Fatalf("Pattern = %q, want %q", got.Pattern, "MMMM y")
	}
}

func TestMatchBasicPrefersAdditionalFieldsOverMissingRequestedFields(t *testing.T) {
	t.Parallel()

	formats := []Formats{
		Parse("y", "y", nil, ""),
		Parse("yMMMMd", "MMMM d, y", nil, ""),
	}
	got := MatchBasic(Options{Year: NumericNumeric, Month: FieldLong}, formats)
	if got.Pattern != "MMMM d, y" {
		t.Fatalf("Pattern = %q, want %q", got.Pattern, "MMMM d, y")
	}
}

func TestMatchBasicPenalizesNumericTextDifferences(t *testing.T) {
	t.Parallel()

	formats := []Formats{
		Parse("yMMMM", "MMMM y", nil, ""),
		Parse("yM", "M/y", nil, ""),
	}
	got := MatchBasic(Options{Year: NumericNumeric, Month: FieldNumeric}, formats)
	if got.Pattern != "M/y" {
		t.Fatalf("Pattern = %q, want %q", got.Pattern, "M/y")
	}
}

func TestMatchBasicSelectsFirstCandidateOnTie(t *testing.T) {
	t.Parallel()

	formats := []Formats{
		Parse("y", "first y", nil, ""),
		Parse("y", "second y", nil, ""),
	}
	got := MatchBasic(Options{Year: NumericNumeric}, formats)
	if got.Pattern != "first y" {
		t.Fatalf("Pattern = %q, want %q", got.Pattern, "first y")
	}
}

func TestAdjustFieldTypesAdjustsYearWidth(t *testing.T) {
	t.Parallel()

	got := AdjustFieldTypes(Formats{Year: NumericNumeric, Pattern: "y"}, Options{Year: Numeric2Digit})
	if got.Pattern != "yy" {
		t.Fatalf("Pattern = %q, want %q", got.Pattern, "yy")
	}
	if got.Year != Numeric2Digit {
		t.Fatalf("Year = %q, want %q", got.Year, Numeric2Digit)
	}
}

func TestMatchAppliesFieldAdjustment(t *testing.T) {
	t.Parallel()

	formats := []Formats{Parse("y", "y", nil, "")}
	opts := Options{Year: Numeric2Digit}
	basic := MatchBasic(opts, formats)
	if basic.Pattern != "y" {
		t.Fatalf("MatchBasic Pattern = %q, want %q", basic.Pattern, "y")
	}
	got := Match(opts, formats)
	if got.Pattern != "yy" {
		t.Fatalf("Match Pattern = %q, want %q", got.Pattern, "yy")
	}
}

func TestAdjustFieldTypesAdjustsCommonFieldWidths(t *testing.T) {
	t.Parallel()

	got := AdjustFieldTypes(
		Parse("yMdHms", "M/d/y, H:m:s", nil, ""),
		Options{
			Year:   Numeric2Digit,
			Month:  FieldLong,
			Day:    Numeric2Digit,
			Hour:   Numeric2Digit,
			Minute: Numeric2Digit,
			Second: Numeric2Digit,
		},
	)
	if got.Pattern != "M/dd/yy, HH:mm:ss" {
		t.Fatalf("Pattern = %q, want %q", got.Pattern, "M/dd/yy, HH:mm:ss")
	}
}

func TestAdjustFieldTypesAdjustsTextFractionalAndTimeZoneFields(t *testing.T) {
	t.Parallel()

	got := AdjustFieldTypes(
		Parse("GEBSz", "G E B S z", nil, ""),
		Options{
			Era:                    FieldNarrow,
			Weekday:                FieldLong,
			DayPeriod:              FieldNarrow,
			FractionalSecondDigits: 3,
			TimeZoneName:           TimeZoneNameLongGeneric,
		},
	)
	if got.Pattern != "GGGGG EEEE BBBBB SSS vvvv" {
		t.Fatalf("Pattern = %q, want %q", got.Pattern, "GGGGG EEEE BBBBB SSS vvvv")
	}
	if got.FractionalSecondDigits != 3 {
		t.Fatalf("FractionalSecondDigits = %d, want 3", got.FractionalSecondDigits)
	}
	if got.TimeZoneName != TimeZoneNameLongGeneric {
		t.Fatalf("TimeZoneName = %q, want %q", got.TimeZoneName, TimeZoneNameLongGeneric)
	}
}

func TestAdjustFieldTypesDoesNotRewriteNumericMonthToText(t *testing.T) {
	t.Parallel()

	got := AdjustFieldTypes(Parse("M", "M", nil, ""), Options{Month: FieldLong})
	if got.Pattern != "M" {
		t.Fatalf("Pattern = %q, want M", got.Pattern)
	}
	if got.Month != FieldNumeric {
		t.Fatalf("Month = %q, want %q", got.Month, FieldNumeric)
	}
}

func TestAdjustFieldTypesRewritesCompatibleMonthWidths(t *testing.T) {
	t.Parallel()

	got := AdjustFieldTypes(Parse("M", "M", nil, ""), Options{Month: Field2Digit})
	if got.Pattern != "MM" {
		t.Fatalf("2-digit month Pattern = %q, want MM", got.Pattern)
	}
	got = AdjustFieldTypes(Parse("MMMM", "MMMM", nil, ""), Options{Month: FieldNumeric})
	if got.Pattern != "M" {
		t.Fatalf("numeric month Pattern = %q, want M", got.Pattern)
	}
}

func TestAdjustFieldTypesAdjustsTimeZoneNameWidths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		base Formats
		want TimeZoneName
		out  string
	}{
		{name: "short", base: Parse("zzzz", "zzzz", nil, ""), want: TimeZoneNameShort, out: "z"},
		{name: "long", base: Parse("z", "z", nil, ""), want: TimeZoneNameLong, out: "zzzz"},
		{name: "short offset", base: Parse("z", "z", nil, ""), want: TimeZoneNameShortOffset, out: "O"},
		{name: "long offset", base: Parse("z", "z", nil, ""), want: TimeZoneNameLongOffset, out: "OOOO"},
		{name: "short generic", base: Parse("z", "z", nil, ""), want: TimeZoneNameShortGeneric, out: "v"},
		{name: "long generic", base: Parse("z", "z", nil, ""), want: TimeZoneNameLongGeneric, out: "vvvv"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := AdjustFieldTypes(tc.base, Options{TimeZoneName: tc.want})
			if got.Pattern != tc.out {
				t.Fatalf("Pattern = %q, want %q", got.Pattern, tc.out)
			}
		})
	}
}

func TestAdjustFieldTypesPreservesHourCyclePatternCharacters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		skeleton string
		want     string
	}{
		{name: "h11", skeleton: "K", want: "KK"},
		{name: "h12", skeleton: "h", want: "hh"},
		{name: "h23", skeleton: "H", want: "HH"},
		{name: "h24", skeleton: "k", want: "kk"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := AdjustFieldTypes(Parse(tc.skeleton, tc.skeleton, nil, ""), Options{Hour: Numeric2Digit})
			if got.Pattern != tc.want {
				t.Fatalf("Pattern = %q, want %q", got.Pattern, tc.want)
			}
		})
	}
}

func TestAdjustFieldTypesDoesNotRewriteQuotedPatternFields(t *testing.T) {
	t.Parallel()

	got := AdjustFieldTypes(Parse("yM", "'M' M y", nil, ""), Options{Month: FieldLong, Year: Numeric2Digit})
	if got.Pattern != "'M' M yy" {
		t.Fatalf("Pattern = %q, want quoted M preserved and year adjusted", got.Pattern)
	}
	if got.Month != FieldNumeric {
		t.Fatalf("Month = %q, want numeric because numeric month should not become text", got.Month)
	}
	if got.Year != Numeric2Digit {
		t.Fatalf("Year = %q, want %q", got.Year, Numeric2Digit)
	}
}

func TestMatchBasicScoresHour12Compatibility(t *testing.T) {
	t.Parallel()

	hour12 := true
	formats := []Formats{
		Parse("H", "H", nil, ""),
		Parse("h", "h a", nil, ""),
	}
	got := MatchBasic(Options{Hour: NumericNumeric, Hour12: &hour12}, formats)
	if got.Pattern != "h a" {
		t.Fatalf("Pattern = %q, want %q", got.Pattern, "h a")
	}
}

func TestMatchBasicScoresTimeZoneNameFallbacks(t *testing.T) {
	t.Parallel()

	shortFormats := []Formats{
		Parse("zzzz", "zzzz", nil, ""),
		Parse("O", "O", nil, ""),
	}
	got := MatchBasic(Options{TimeZoneName: TimeZoneNameShort}, shortFormats)
	if got.Pattern != "O" {
		t.Fatalf("short TimeZoneName Pattern = %q, want O", got.Pattern)
	}

	longFormats := []Formats{
		Parse("z", "z", nil, ""),
		Parse("OOOO", "OOOO", nil, ""),
	}
	got = MatchBasic(Options{TimeZoneName: TimeZoneNameLong}, longFormats)
	if got.Pattern != "OOOO" {
		t.Fatalf("long TimeZoneName Pattern = %q, want OOOO", got.Pattern)
	}

	genericFormats := []Formats{
		Parse("v", "v", nil, ""),
		Parse("vvvv", "vvvv", nil, ""),
	}
	got = MatchBasic(Options{TimeZoneName: TimeZoneNameLongGeneric}, genericFormats)
	if got.Pattern != "vvvv" {
		t.Fatalf("long generic TimeZoneName Pattern = %q, want vvvv", got.Pattern)
	}

	offsetFormats := []Formats{
		Parse("O", "O", nil, ""),
		Parse("z", "z", nil, ""),
	}
	got = MatchBasic(Options{TimeZoneName: TimeZoneNameLongOffset}, offsetFormats)
	if got.Pattern != "O" {
		t.Fatalf("long offset TimeZoneName Pattern = %q, want O", got.Pattern)
	}
}

func TestMatchBasicScoresFractionalSecondDigits(t *testing.T) {
	t.Parallel()

	formats := []Formats{
		Parse("S", "S", nil, ""),
		Parse("SSS", "SSS", nil, ""),
	}
	got := MatchBasic(Options{FractionalSecondDigits: 3}, formats)
	if got.Pattern != "SSS" {
		t.Fatalf("Pattern = %q, want SSS", got.Pattern)
	}
}

func TestMatchBasicScoresTextWidthDistance(t *testing.T) {
	t.Parallel()

	formats := []Formats{
		Parse("MMMMM", "MMMMM", nil, ""),
		Parse("MMM", "MMM", nil, ""),
	}
	got := MatchBasic(Options{Month: FieldLong}, formats)
	if got.Pattern != "MMM" {
		t.Fatalf("long month Pattern = %q, want MMM", got.Pattern)
	}

	formats = []Formats{
		Parse("MMMMM", "MMMMM", nil, ""),
		Parse("MMMM", "MMMM", nil, ""),
	}
	got = MatchBasic(Options{Month: FieldShort}, formats)
	if got.Pattern != "MMMM" {
		t.Fatalf("short month Pattern = %q, want MMMM", got.Pattern)
	}
}
