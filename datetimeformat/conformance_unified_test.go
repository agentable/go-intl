package datetimeformat

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", func(t *testing.T, fixture conformance.Fixture) {
		loc := locale.MustParse(fixture.Locale)
		format, err := New(locale.List{loc}, conformanceDateTimeOptions(t, fixture))
		if fixture.ErrorCode != "" {
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		if fixture.ExpectedRange != nil {
			start, end := conformanceDateTimeRangeInput(t, fixture)
			if got := format.FormatRange(start, end); got != *fixture.ExpectedRange {
				t.Fatalf("FormatRange(%v, %v) = %q, want %q", start, end, got, *fixture.ExpectedRange)
			}
			if len(fixture.ExpectedRangeParts) > 0 {
				parts := format.FormatRangeToParts(start, end)
				if len(parts) != len(fixture.ExpectedRangeParts) {
					t.Fatalf("FormatRangeToParts(%v, %v) length = %d, want %d", start, end, len(parts), len(fixture.ExpectedRangeParts))
				}
				for i, part := range parts {
					want := fixture.ExpectedRangeParts[i]
					if string(part.Type) != want.Type || part.Value != want.Value || string(part.Source) != want.Source {
						t.Fatalf("FormatRangeToParts(%v, %v)[%d] = {%q, %q, %q}, want {%q, %q, %q}", start, end, i, part.Type, part.Value, part.Source, want.Type, want.Value, want.Source)
					}
				}
			}
			return
		}
		input := conformanceDateTimeInput(t, fixture)
		if fixture.Expected == nil {
			t.Fatal("fixture expected is required")
		}
		if got := format.Format(input); got != *fixture.Expected {
			t.Fatalf("Format(%v) = %q, want %q", input, got, *fixture.Expected)
		}
		if len(fixture.ExpectedParts) > 0 {
			parts := format.FormatToParts(input)
			if len(parts) != len(fixture.ExpectedParts) {
				t.Fatalf("FormatToParts(%v) length = %d, want %d", input, len(parts), len(fixture.ExpectedParts))
			}
			for i, part := range parts {
				want := fixture.ExpectedParts[i]
				if string(part.Type) != want.Type || part.Value != want.Value {
					t.Fatalf("FormatToParts(%v)[%d] = {%q, %q}, want {%q, %q}", input, i, part.Type, part.Value, want.Type, want.Value)
				}
			}
		}
	})
}

func conformanceDateTimeOptions(t *testing.T, fixture conformance.Fixture) Options {
	t.Helper()

	var options struct {
		Calendar               string `json:"calendar"`
		NumberingSystem        string `json:"numberingSystem"`
		LocaleMatcher          string `json:"localeMatcher"`
		FormatMatcher          string `json:"formatMatcher"`
		TimeZone               string `json:"timeZone"`
		TimeZoneName           string `json:"timeZoneName"`
		Weekday                string `json:"weekday"`
		Era                    string `json:"era"`
		Year                   string `json:"year"`
		Month                  string `json:"month"`
		Day                    string `json:"day"`
		DayPeriod              string `json:"dayPeriod"`
		Hour                   string `json:"hour"`
		Minute                 string `json:"minute"`
		Second                 string `json:"second"`
		HourCycle              string `json:"hourCycle"`
		Hour12                 *bool  `json:"hour12"`
		DateStyle              string `json:"dateStyle"`
		TimeStyle              string `json:"timeStyle"`
		FractionalSecondDigits *int   `json:"fractionalSecondDigits"`
	}
	if err := json.Unmarshal(fixture.Options, &options); err != nil {
		t.Fatal(err)
	}
	var opts Options
	if options.Calendar != "" {
		opts.Calendar = options.Calendar
	}
	if options.NumberingSystem != "" {
		opts.NumberingSystem = options.NumberingSystem
	}
	if options.LocaleMatcher != "" {
		opts.LocaleMatcher = LocaleMatcher(options.LocaleMatcher)
	}
	if options.FormatMatcher != "" {
		opts.FormatMatcher = FormatMatcher(options.FormatMatcher)
	}
	if options.TimeZone != "" {
		opts.TimeZone = options.TimeZone
	}
	if options.TimeZoneName != "" {
		opts.TimeZoneName = TimeZoneName(options.TimeZoneName)
	}
	if options.Weekday != "" {
		opts.Weekday = FieldStyle(options.Weekday)
	}
	if options.Era != "" {
		opts.Era = FieldStyle(options.Era)
	}
	if options.Year != "" {
		opts.Year = NumericStyle(options.Year)
	}
	if options.Month != "" {
		opts.Month = MonthStyle(options.Month)
	}
	if options.Day != "" {
		opts.Day = NumericStyle(options.Day)
	}
	if options.DayPeriod != "" {
		opts.DayPeriod = FieldStyle(options.DayPeriod)
	}
	if options.Hour != "" {
		opts.Hour = NumericStyle(options.Hour)
	}
	if options.Minute != "" {
		opts.Minute = NumericStyle(options.Minute)
	}
	if options.Second != "" {
		opts.Second = NumericStyle(options.Second)
	}
	if options.HourCycle != "" {
		opts.HourCycle = HourCycle(options.HourCycle)
	}
	if options.Hour12 != nil {
		opts.Hour12 = options.Hour12
	}
	if options.DateStyle != "" {
		opts.DateStyle = Style(options.DateStyle)
	}
	if options.TimeStyle != "" {
		opts.TimeStyle = Style(options.TimeStyle)
	}
	if options.FractionalSecondDigits != nil {
		opts.FractionalSecondDigits = options.FractionalSecondDigits
	}
	return opts
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
