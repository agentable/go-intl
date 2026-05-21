package durationformat

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestDurationFormatConformance(t *testing.T) {
	t.Parallel()

	fixtures, err := conformance.LoadFixtures(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		t.Run(fixture.ID, func(t *testing.T) {
			t.Parallel()
			loc := locale.MustParse(fixture.Locale)
			format, err := New(locale.List{loc}, conformanceDurationOptions(t, fixture))
			if err != nil {
				if fixture.ErrorCode == "invalid-option" {
					return
				}
				t.Fatalf("New(%q) error = %v", fixture.Locale, err)
			}
			input := conformanceDurationInput(t, fixture)
			if len(fixture.ExpectedParts) > 0 {
				got, err := format.FormatToParts(input)
				if err != nil {
					t.Fatalf("FormatToParts(%v) error = %v", input, err)
				}
				assertDurationParts(t, got, fixture.ExpectedParts)
				return
			}
			got, err := format.Format(input)
			if err != nil {
				t.Fatalf("Format(%v) error = %v", input, err)
			}
			if fixture.Expected == nil {
				t.Fatalf("fixture %s has no expected output", fixture.ID)
			}
			if got != *fixture.Expected {
				t.Fatalf("Format(%v) = %q, want %q", input, got, *fixture.Expected)
			}
		})
	}
}

func conformanceDurationOptions(t *testing.T, fixture conformance.Fixture) Options {
	t.Helper()
	var raw map[string]any
	if err := json.Unmarshal(fixture.Options, &raw); err != nil {
		t.Fatalf("fixture %s options: %v", fixture.ID, err)
	}
	var opts Options
	for key, value := range raw {
		switch key {
		case "localeMatcher":
			opts.LocaleMatcher = LocaleMatcher(value.(string))
		case "numberingSystem":
			opts.NumberingSystem = value.(string)
		case "style":
			opts.Style = Style(value.(string))
		case "years":
			opts.Years = UnitStyle(value.(string))
		case "yearsDisplay":
			opts.YearsDisplay = Display(value.(string))
		case "months":
			opts.Months = UnitStyle(value.(string))
		case "monthsDisplay":
			opts.MonthsDisplay = Display(value.(string))
		case "weeks":
			opts.Weeks = UnitStyle(value.(string))
		case "weeksDisplay":
			opts.WeeksDisplay = Display(value.(string))
		case "days":
			opts.Days = UnitStyle(value.(string))
		case "daysDisplay":
			opts.DaysDisplay = Display(value.(string))
		case "hours":
			opts.Hours = UnitStyle(value.(string))
		case "hoursDisplay":
			opts.HoursDisplay = Display(value.(string))
		case "minutes":
			opts.Minutes = UnitStyle(value.(string))
		case "minutesDisplay":
			opts.MinutesDisplay = Display(value.(string))
		case "seconds":
			opts.Seconds = UnitStyle(value.(string))
		case "secondsDisplay":
			opts.SecondsDisplay = Display(value.(string))
		case "milliseconds":
			opts.Milliseconds = UnitStyle(value.(string))
		case "millisecondsDisplay":
			opts.MillisecondsDisplay = Display(value.(string))
		case "microseconds":
			opts.Microseconds = UnitStyle(value.(string))
		case "microsecondsDisplay":
			opts.MicrosecondsDisplay = Display(value.(string))
		case "nanoseconds":
			opts.Nanoseconds = UnitStyle(value.(string))
		case "nanosecondsDisplay":
			opts.NanosecondsDisplay = Display(value.(string))
		case "fractionalDigits":
			opts.FractionalDigits = new(int(value.(float64)))
		default:
			t.Fatalf("fixture %s has unsupported option %q", fixture.ID, key)
		}
	}
	return opts
}

func conformanceDurationInput(t *testing.T, fixture conformance.Fixture) Duration {
	t.Helper()
	var raw map[string]float64
	if err := json.Unmarshal(fixture.Input, &raw); err != nil {
		t.Fatalf("fixture %s input: %v", fixture.ID, err)
	}
	return Duration{
		Years:        int64(raw["years"]),
		Months:       int64(raw["months"]),
		Weeks:        int64(raw["weeks"]),
		Days:         int64(raw["days"]),
		Hours:        int64(raw["hours"]),
		Minutes:      int64(raw["minutes"]),
		Seconds:      int64(raw["seconds"]),
		Milliseconds: int64(raw["milliseconds"]),
		Microseconds: int64(raw["microseconds"]),
		Nanoseconds:  int64(raw["nanoseconds"]),
	}
}

func assertDurationParts(t *testing.T, got []Part, want []conformance.Part) {
	t.Helper()
	converted := make([]Part, len(want))
	for i, part := range want {
		converted[i] = Part{Type: PartType(part.Type), Value: part.Value, Unit: Unit(part.Unit)}
	}
	if !reflect.DeepEqual(got, converted) {
		t.Fatalf("FormatToParts() = %#v, want %#v", got, converted)
	}
}
