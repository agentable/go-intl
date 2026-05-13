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
