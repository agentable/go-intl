package durationformat

import (
	"encoding/json/v2"
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestDurationFormatConformance(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, testcontract.FixtureRunnerRoot("."), runDurationConformanceFixture)
}

func runDurationConformanceFixture(t *testing.T, fixture conformance.Fixture) {
	t.Helper()

	if fixture.IsSupportedLocalesOf() {
		testcontract.AssertSupportedLocalesOfFixture(t, fixture, intltest.LocaleListJSON, func(locales locale.List) (locale.List, error) {
			return SupportedLocalesOf(locales, conformanceDurationOptions(t, fixture))
		}, func(code string) error {
			return conformanceDurationError(t, code)
		})
		return
	}
	constructorError := fixture.ErrorCode == "invalid_option" || fixture.ErrorCode == "invalid-option"
	if fixture.ExpectedRange != nil || fixture.ExpectedRangeParts != nil || fixture.ExpectedOK != nil || fixture.ExpectedLocales != nil ||
		fixture.ErrorCode != "" && (fixture.Expected != nil || fixture.ExpectedParts != nil) ||
		constructorError && fixture.ExpectedResolved != nil {
		t.Fatalf("fixture %s: unsupported durationformat observation", fixture.ID)
	}
	loc := intltest.Locale(t, fixture.Locale)
	format, err := New(locale.List{loc}, conformanceDurationOptions(t, fixture))
	if constructorError {
		testcontract.AssertErrorCode(t, "New("+fixture.Locale+")", err, fixture.ErrorCode, func(code string) error {
			return conformanceDurationError(t, code)
		})
		return
	}
	if err != nil {
		t.Fatalf("New(%q) error = %v", fixture.Locale, err)
	}
	if fixture.ExpectedResolved != nil {
		testcontract.AssertResolvedOptionsJSON(t, format.ResolvedOptions(), fixture.ExpectedResolved)
	}
	if fixture.ErrorCode == "" && fixture.Expected == nil && fixture.ExpectedParts == nil {
		return
	}
	input := conformanceDurationInput(t, fixture)
	if fixture.ErrorCode != "" {
		operation := "Format()"
		switch fixture.Feature {
		case "", "format":
			_, err = format.Format(input)
		case "formatToParts":
			operation = "FormatToParts()"
			_, err = format.FormatToParts(input)
		default:
			t.Fatalf("unsupported durationformat feature %q", fixture.Feature)
		}
		testcontract.AssertErrorCode(t, operation, err, fixture.ErrorCode, func(code string) error {
			return testcontract.IntlErrorCode(t, "durationformat runtime", code, "invalid_value")
		})
		return
	}
	if fixture.Expected != nil {
		got, err := format.Format(input)
		if err != nil {
			t.Fatalf("Format(%v) error = %v", input, err)
		}
		if got != *fixture.Expected {
			t.Fatalf("Format(%v) = %q, want %q", input, got, *fixture.Expected)
		}
	}
	if fixture.ExpectedParts != nil {
		got, err := format.FormatToParts(input)
		if err != nil {
			t.Fatalf("FormatToParts(%v) error = %v", input, err)
		}
		testcontract.AssertParts(t, "FormatToParts", got, fixture.ExpectedParts, conformanceDurationPart)
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
		Years:        raw["years"],
		Months:       raw["months"],
		Weeks:        raw["weeks"],
		Days:         raw["days"],
		Hours:        raw["hours"],
		Minutes:      raw["minutes"],
		Seconds:      raw["seconds"],
		Milliseconds: raw["milliseconds"],
		Microseconds: raw["microseconds"],
		Nanoseconds:  raw["nanoseconds"],
	}
}

func conformanceDurationPart(part Part) conformance.Part {
	return conformance.Part{Type: string(part.Type), Value: part.Value, Unit: string(part.Unit)}
}
