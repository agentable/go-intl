package durationformat

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/intltest"
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
			if fixture.ErrorCode != "" {
				if !errors.Is(err, conformanceDurationError(t, fixture.ErrorCode)) {
					t.Fatalf("New(%q) error = %v, want %q", fixture.Locale, err, fixture.ErrorCode)
				}
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
				assertDurationParts(t, got, fixture.ExpectedParts)
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

func conformanceDurationError(t *testing.T, code string) error {
	t.Helper()

	switch code {
	case "invalid_option", "invalid-option":
		return intlerr.ErrInvalidOption
	default:
		t.Fatalf("unsupported durationformat errorCode %q", code)
		return nil
	}
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

func assertDurationResolvedOptions(t *testing.T, fixture conformance.Fixture, got ResolvedOptions) {
	t.Helper()

	var want map[string]json.RawMessage
	if err := json.Unmarshal(fixture.ExpectedResolved, &want); err != nil {
		t.Fatal(err)
	}
	assertDurationResolvedString(t, want, "locale", got.Locale.String())
	assertDurationResolvedString(t, want, "numberingSystem", got.NumberingSystem)
	assertDurationResolvedString(t, want, "style", string(got.Style))
	assertDurationResolvedString(t, want, "years", string(got.Years))
	assertDurationResolvedString(t, want, "yearsDisplay", string(got.YearsDisplay))
	assertDurationResolvedString(t, want, "months", string(got.Months))
	assertDurationResolvedString(t, want, "monthsDisplay", string(got.MonthsDisplay))
	assertDurationResolvedString(t, want, "weeks", string(got.Weeks))
	assertDurationResolvedString(t, want, "weeksDisplay", string(got.WeeksDisplay))
	assertDurationResolvedString(t, want, "days", string(got.Days))
	assertDurationResolvedString(t, want, "daysDisplay", string(got.DaysDisplay))
	assertDurationResolvedString(t, want, "hours", string(got.Hours))
	assertDurationResolvedString(t, want, "hoursDisplay", string(got.HoursDisplay))
	assertDurationResolvedString(t, want, "minutes", string(got.Minutes))
	assertDurationResolvedString(t, want, "minutesDisplay", string(got.MinutesDisplay))
	assertDurationResolvedString(t, want, "seconds", string(got.Seconds))
	assertDurationResolvedString(t, want, "secondsDisplay", string(got.SecondsDisplay))
	assertDurationResolvedString(t, want, "milliseconds", string(got.Milliseconds))
	assertDurationResolvedString(t, want, "millisecondsDisplay", string(got.MillisecondsDisplay))
	assertDurationResolvedString(t, want, "microseconds", string(got.Microseconds))
	assertDurationResolvedString(t, want, "microsecondsDisplay", string(got.MicrosecondsDisplay))
	assertDurationResolvedString(t, want, "nanoseconds", string(got.Nanoseconds))
	assertDurationResolvedString(t, want, "nanosecondsDisplay", string(got.NanosecondsDisplay))
	assertDurationResolvedOptionalInt(t, want, "fractionalDigits", got.FractionalDigits)
}

func assertDurationResolvedString(t *testing.T, values map[string]json.RawMessage, name, got string) {
	t.Helper()

	raw, ok := values[name]
	if !ok {
		return
	}
	var want string
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("expectedResolvedOptions.%s = %s: %v", name, raw, err)
	}
	if got != want {
		t.Fatalf("ResolvedOptions().%s = %q, want %q", name, got, want)
	}
}

func assertDurationResolvedOptionalInt(t *testing.T, values map[string]json.RawMessage, name string, got *int) {
	t.Helper()

	raw, ok := values[name]
	if !ok {
		return
	}
	if string(raw) == "null" {
		if got != nil {
			t.Fatalf("ResolvedOptions().%s = %d, want omitted", name, *got)
		}
		return
	}
	var want int
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("expectedResolvedOptions.%s = %s: %v", name, raw, err)
	}
	if got == nil {
		t.Fatalf("ResolvedOptions().%s omitted, want %d", name, want)
	}
	if *got != want {
		t.Fatalf("ResolvedOptions().%s = %d, want %d", name, *got, want)
	}
}
