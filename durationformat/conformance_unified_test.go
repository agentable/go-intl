package durationformat

import (
	"encoding/json"
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
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
			loc := intltest.Locale(t, fixture.Locale)
			format, err := New(locale.List{loc}, conformanceDurationOptions(t, fixture))
			if testcontract.AssertErrorCode(t, "New("+fixture.Locale+")", err, fixture.ErrorCode, func(code string) error {
				return conformanceDurationError(t, code)
			}) {
				return
			}
			if err != nil {
				t.Fatalf("New(%q) error = %v", fixture.Locale, err)
			}
			if len(fixture.ExpectedResolved) != 0 {
				assertDurationResolvedOptions(t, fixture, format.ResolvedOptions())
			}
			input := conformanceDurationInput(t, fixture)
			if fixture.Expected != nil {
				got, err := format.Format(input)
				if err != nil {
					t.Fatalf("Format(%v) error = %v", input, err)
				}
				if got != *fixture.Expected {
					t.Fatalf("Format(%v) = %q, want %q", input, got, *fixture.Expected)
				}
			}
			if len(fixture.ExpectedParts) > 0 {
				got, err := format.FormatToParts(input)
				if err != nil {
					t.Fatalf("FormatToParts(%v) error = %v", input, err)
				}
				testcontract.AssertParts(t, "FormatToParts", got, fixture.ExpectedParts, conformanceDurationPart)
				return
			}
			if fixture.Expected == nil {
				t.Fatalf("fixture %s has no expected output", fixture.ID)
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
			opts.LocaleMatcher = stringPtr(value.(string))
		case "numberingSystem":
			numberingSystem := value.(string)
			opts.NumberingSystem = &numberingSystem
		case "style":
			opts.Style = stringPtr(value.(string))
		case "years":
			opts.Years = stringPtr(value.(string))
		case "yearsDisplay":
			opts.YearsDisplay = stringPtr(value.(string))
		case "months":
			opts.Months = stringPtr(value.(string))
		case "monthsDisplay":
			opts.MonthsDisplay = stringPtr(value.(string))
		case "weeks":
			opts.Weeks = stringPtr(value.(string))
		case "weeksDisplay":
			opts.WeeksDisplay = stringPtr(value.(string))
		case "days":
			opts.Days = stringPtr(value.(string))
		case "daysDisplay":
			opts.DaysDisplay = stringPtr(value.(string))
		case "hours":
			opts.Hours = stringPtr(value.(string))
		case "hoursDisplay":
			opts.HoursDisplay = stringPtr(value.(string))
		case "minutes":
			opts.Minutes = stringPtr(value.(string))
		case "minutesDisplay":
			opts.MinutesDisplay = stringPtr(value.(string))
		case "seconds":
			opts.Seconds = stringPtr(value.(string))
		case "secondsDisplay":
			opts.SecondsDisplay = stringPtr(value.(string))
		case "milliseconds":
			opts.Milliseconds = stringPtr(value.(string))
		case "millisecondsDisplay":
			opts.MillisecondsDisplay = stringPtr(value.(string))
		case "microseconds":
			opts.Microseconds = stringPtr(value.(string))
		case "microsecondsDisplay":
			opts.MicrosecondsDisplay = stringPtr(value.(string))
		case "nanoseconds":
			opts.Nanoseconds = stringPtr(value.(string))
		case "nanosecondsDisplay":
			opts.NanosecondsDisplay = stringPtr(value.(string))
		case "fractionalDigits":
			opts.FractionalDigits = new(int(value.(float64)))
		default:
			t.Fatalf("fixture %s has unsupported option %q", fixture.ID, key)
		}
	}
	return opts
}

func conformanceDurationError(t *testing.T, code string) error {
	t.Helper()

	return testcontract.IntlErrorCode(t, "durationformat", code, "invalid_option", "invalid-option")
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

func conformanceDurationPart(part Part) conformance.Part {
	return conformance.Part{Type: string(part.Type), Value: part.Value, Unit: string(part.Unit)}
}

func assertDurationResolvedOptions(t *testing.T, fixture conformance.Fixture, got ResolvedOptions) {
	t.Helper()

	want := testcontract.ExpectedResolvedOptions(t, fixture)
	testcontract.AssertResolvedString(t, want, "locale", got.Locale.String())
	testcontract.AssertResolvedString(t, want, "numberingSystem", got.NumberingSystem)
	testcontract.AssertResolvedString(t, want, "style", string(got.Style))
	testcontract.AssertResolvedString(t, want, "years", string(got.Years))
	testcontract.AssertResolvedString(t, want, "yearsDisplay", string(got.YearsDisplay))
	testcontract.AssertResolvedString(t, want, "months", string(got.Months))
	testcontract.AssertResolvedString(t, want, "monthsDisplay", string(got.MonthsDisplay))
	testcontract.AssertResolvedString(t, want, "weeks", string(got.Weeks))
	testcontract.AssertResolvedString(t, want, "weeksDisplay", string(got.WeeksDisplay))
	testcontract.AssertResolvedString(t, want, "days", string(got.Days))
	testcontract.AssertResolvedString(t, want, "daysDisplay", string(got.DaysDisplay))
	testcontract.AssertResolvedString(t, want, "hours", string(got.Hours))
	testcontract.AssertResolvedString(t, want, "hoursDisplay", string(got.HoursDisplay))
	testcontract.AssertResolvedString(t, want, "minutes", string(got.Minutes))
	testcontract.AssertResolvedString(t, want, "minutesDisplay", string(got.MinutesDisplay))
	testcontract.AssertResolvedString(t, want, "seconds", string(got.Seconds))
	testcontract.AssertResolvedString(t, want, "secondsDisplay", string(got.SecondsDisplay))
	testcontract.AssertResolvedString(t, want, "milliseconds", string(got.Milliseconds))
	testcontract.AssertResolvedString(t, want, "millisecondsDisplay", string(got.MillisecondsDisplay))
	testcontract.AssertResolvedString(t, want, "microseconds", string(got.Microseconds))
	testcontract.AssertResolvedString(t, want, "microsecondsDisplay", string(got.MicrosecondsDisplay))
	testcontract.AssertResolvedString(t, want, "nanoseconds", string(got.Nanoseconds))
	testcontract.AssertResolvedString(t, want, "nanosecondsDisplay", string(got.NanosecondsDisplay))
	testcontract.AssertResolvedOptionalInt(t, want, "fractionalDigits", got.FractionalDigits)
}
