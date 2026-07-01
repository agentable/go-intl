package datetimeformat

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", func(t *testing.T, fixture conformance.Fixture) {
		loc := intltest.Locale(t, fixture.Locale)
		format, err := New(locale.List{loc}, conformanceDateTimeOptions(t, fixture))
		if testcontract.AssertErrorCode(t, "New()", err, fixture.ErrorCode, func(code string) error {
			return conformanceDateTimeError(t, code)
		}) {
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		if fixture.ExpectedRange != nil {
			start, end := conformanceDateTimeRangeInput(t, fixture)
			got, err := format.FormatRange(start, end)
			if err != nil {
				t.Fatalf("FormatRange(%v, %v) error = %v", start, end, err)
			}
			testcontract.AssertExpectedRange(t, "FormatRange", got, fixture.ExpectedRange)
			if len(fixture.ExpectedRangeParts) > 0 {
				parts, err := format.FormatRangeToParts(start, end)
				if err != nil {
					t.Fatalf("FormatRangeToParts(%v, %v) error = %v", start, end, err)
				}
				testcontract.AssertRangeParts(t, "FormatRangeToParts", parts, fixture.ExpectedRangeParts, conformanceDateTimeRangePart)
			}
			return
		}
		input := conformanceDateTimeInput(t, fixture)
		want := fixture.RequiredExpected(t)
		if got := format.Format(input); got != want {
			t.Fatalf("Format(%v) = %q, want %q", input, got, want)
		}
		if len(fixture.ExpectedParts) > 0 {
			parts := format.FormatToParts(input)
			testcontract.AssertParts(t, "FormatToParts", parts, fixture.ExpectedParts, conformanceDateTimePart)
		}
	})
}

func conformanceDateTimeError(t *testing.T, code string) error {
	t.Helper()

	return testcontract.IntlErrorCode(t, "datetimeformat", code, "invalid_option", "unsupported_option")
}

func conformanceDateTimePart(part Part) conformance.Part {
	return conformance.Part{Type: string(part.Type), Value: part.Value}
}

func conformanceDateTimeRangePart(part RangePart) conformance.RangePart {
	return conformance.RangePart{Type: string(part.Type), Value: part.Value, Source: string(part.Source)}
}

func conformanceDateTimeOptions(t *testing.T, fixture conformance.Fixture) Options {
	t.Helper()

	var options struct {
		Calendar               *string `json:"calendar"`
		NumberingSystem        *string `json:"numberingSystem"`
		LocaleMatcher          *string `json:"localeMatcher"`
		FormatMatcher          *string `json:"formatMatcher"`
		TimeZone               *string `json:"timeZone"`
		TimeZoneName           *string `json:"timeZoneName"`
		Weekday                *string `json:"weekday"`
		Era                    *string `json:"era"`
		Year                   *string `json:"year"`
		Month                  *string `json:"month"`
		Day                    *string `json:"day"`
		DayPeriod              *string `json:"dayPeriod"`
		Hour                   *string `json:"hour"`
		Minute                 *string `json:"minute"`
		Second                 *string `json:"second"`
		HourCycle              *string `json:"hourCycle"`
		Hour12                 *bool   `json:"hour12"`
		DateStyle              *string `json:"dateStyle"`
		TimeStyle              *string `json:"timeStyle"`
		FractionalSecondDigits *int    `json:"fractionalSecondDigits"`
	}
	if err := json.Unmarshal(fixture.Options, &options); err != nil {
		t.Fatal(err)
	}
	return Options{
		Calendar:               options.Calendar,
		NumberingSystem:        options.NumberingSystem,
		LocaleMatcher:          options.LocaleMatcher,
		FormatMatcher:          options.FormatMatcher,
		TimeZone:               options.TimeZone,
		TimeZoneName:           options.TimeZoneName,
		Weekday:                options.Weekday,
		Era:                    options.Era,
		Year:                   options.Year,
		Month:                  options.Month,
		Day:                    options.Day,
		DayPeriod:              options.DayPeriod,
		Hour:                   options.Hour,
		Minute:                 options.Minute,
		Second:                 options.Second,
		HourCycle:              options.HourCycle,
		Hour12:                 options.Hour12,
		DateStyle:              options.DateStyle,
		TimeStyle:              options.TimeStyle,
		FractionalSecondDigits: options.FractionalSecondDigits,
	}
}

func conformanceDateTimeInput(t *testing.T, fixture conformance.Fixture) time.Time {
	t.Helper()

	var input string
	if err := json.Unmarshal(fixture.Input, &input); err != nil {
		t.Fatal(err)
	}
	return parseConformanceTime(t, input)
}

func conformanceDateTimeRangeInput(t *testing.T, fixture conformance.Fixture) (time.Time, time.Time) {
	t.Helper()

	var input struct {
		Start string `json:"start"`
		End   string `json:"end"`
	}
	if err := json.Unmarshal(fixture.Input, &input); err != nil {
		t.Fatal(err)
	}
	return parseConformanceTime(t, input.Start), parseConformanceTime(t, input.End)
}

func parseConformanceTime(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
