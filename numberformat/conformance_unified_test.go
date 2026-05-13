package numberformat

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/conformance"
	"github.com/agentable/go-intl/locale"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", func(t *testing.T, fixture conformance.Fixture) {
		loc := locale.MustParse(fixture.Locale)
		format, err := New(loc, conformanceNumberOptions(t, fixture))
		if fixture.ErrorCode != "" {
			if !errors.Is(err, ErrInvalidOption) {
				t.Fatalf("New() error = %v, want ErrInvalidOption", err)
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		var input any
		if err := json.Unmarshal(fixture.Input, &input); err != nil {
			t.Fatal(err)
		}
		if fixture.ExpectedRange != nil || len(fixture.ExpectedRangeParts) > 0 {
			rangeInput := conformanceNumberRangeInput(t, fixture)
			if fixture.ExpectedRange != nil {
				if got := format.formatRangeValue(rangeInput.Start, rangeInput.End); got != *fixture.ExpectedRange {
					t.Fatalf("FormatRange(%v, %v) = %q, want %q", rangeInput.Start, rangeInput.End, got, *fixture.ExpectedRange)
				}
			}
			if len(fixture.ExpectedRangeParts) > 0 {
				parts := format.formatRangeToPartsValue(rangeInput.Start, rangeInput.End)
				if len(parts) != len(fixture.ExpectedRangeParts) {
					t.Fatalf("FormatRangeToParts(%v, %v) length = %d, want %d", rangeInput.Start, rangeInput.End, len(parts), len(fixture.ExpectedRangeParts))
				}
				for i, part := range parts {
					want := fixture.ExpectedRangeParts[i]
					if string(part.Type) != want.Type || part.Value != want.Value || string(part.Source) != want.Source {
						t.Fatalf("FormatRangeToParts(%v, %v)[%d] = {%q, %q, %q}, want {%q, %q, %q}", rangeInput.Start, rangeInput.End, i, part.Type, part.Value, part.Source, want.Type, want.Value, want.Source)
					}
				}
			}
			return
		}
		if fixture.Expected == nil {
			t.Fatal("fixture expected is required")
		}
		if got := format.formatValue(input); got != *fixture.Expected {
			t.Fatalf("Format(%v) = %q, want %q", input, got, *fixture.Expected)
		}
		if len(fixture.ExpectedParts) > 0 {
			parts := format.formatToPartsValue(input)
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

type conformanceNumberRange struct {
	Start any `json:"start"`
	End   any `json:"end"`
}

func conformanceNumberRangeInput(t *testing.T, fixture conformance.Fixture) conformanceNumberRange {
	t.Helper()

	var input conformanceNumberRange
	if err := json.Unmarshal(fixture.Input, &input); err != nil {
		t.Fatal(err)
	}
	return input
}

func conformanceNumberOptions(t *testing.T, fixture conformance.Fixture) Options {
	t.Helper()

	var options struct {
		Style                    string `json:"style"`
		Currency                 string `json:"currency"`
		CurrencyDisplay          string `json:"currencyDisplay"`
		CurrencySign             string `json:"currencySign"`
		Unit                     string `json:"unit"`
		UnitDisplay              string `json:"unitDisplay"`
		MinimumIntegerDigits     *int   `json:"minimumIntegerDigits"`
		MinimumFractionDigits    *int   `json:"minimumFractionDigits"`
		MaximumFractionDigits    *int   `json:"maximumFractionDigits"`
		MinimumSignificantDigits *int   `json:"minimumSignificantDigits"`
		MaximumSignificantDigits *int   `json:"maximumSignificantDigits"`
		RoundingIncrement        *int   `json:"roundingIncrement"`
		RoundingPriority         string `json:"roundingPriority"`
		RoundingMode             string `json:"roundingMode"`
		TrailingZeroDisplay      string `json:"trailingZeroDisplay"`
		Notation                 string `json:"notation"`
		CompactDisplay           string `json:"compactDisplay"`
		UseGrouping              any    `json:"useGrouping"`
		SignDisplay              string `json:"signDisplay"`
		LocaleMatcher            string `json:"localeMatcher"`
		NumberingSystem          string `json:"numberingSystem"`
	}
	if err := json.Unmarshal(fixture.Options, &options); err != nil {
		t.Fatal(err)
	}
	var opts Options
	if options.Style != "" {
		opts.Style = Style(options.Style)
	}
	if options.Currency != "" {
		opts.Currency = CurrencyCode(options.Currency)
	}
	if options.CurrencyDisplay != "" {
		opts.CurrencyDisplay = CurrencyDisplay(options.CurrencyDisplay)
	}
	if options.CurrencySign != "" {
		opts.CurrencySign = CurrencySign(options.CurrencySign)
	}
	if options.Unit != "" {
		opts.Unit = UnitIdentifier(options.Unit)
	}
	if options.UnitDisplay != "" {
		opts.UnitDisplay = UnitDisplay(options.UnitDisplay)
	}
	if options.MinimumIntegerDigits != nil {
		opts.MinimumIntegerDigits = *options.MinimumIntegerDigits
	}
	if options.MinimumFractionDigits != nil {
		opts.FractionDigits = MinimumFractionDigits(*options.MinimumFractionDigits)
	}
	if options.MaximumFractionDigits != nil {
		if opts.FractionDigits.hasMin {
			opts.FractionDigits = FractionDigits(opts.FractionDigits.min, *options.MaximumFractionDigits)
		} else {
			opts.FractionDigits = MaximumFractionDigits(*options.MaximumFractionDigits)
		}
	}
	if options.MinimumSignificantDigits != nil {
		opts.SignificantDigits = MinimumSignificantDigits(*options.MinimumSignificantDigits)
	}
	if options.MaximumSignificantDigits != nil {
		if opts.SignificantDigits.hasMin {
			opts.SignificantDigits = SignificantDigits(opts.SignificantDigits.min, *options.MaximumSignificantDigits)
		} else {
			opts.SignificantDigits = MaximumSignificantDigits(*options.MaximumSignificantDigits)
		}
	}
	if options.RoundingIncrement != nil {
		opts.RoundingIncrement = *options.RoundingIncrement
	}
	if options.RoundingPriority != "" {
		opts.RoundingPriority = RoundingPriority(options.RoundingPriority)
	}
	if options.RoundingMode != "" {
		opts.RoundingMode = RoundingMode(options.RoundingMode)
	}
	if options.TrailingZeroDisplay != "" {
		opts.TrailingZeroDisplay = TrailingZeroDisplay(options.TrailingZeroDisplay)
	}
	if options.Notation != "" {
		opts.Notation = Notation(options.Notation)
	}
	if options.CompactDisplay != "" {
		opts.CompactDisplay = CompactDisplay(options.CompactDisplay)
	}
	if options.UseGrouping != nil {
		opts.UseGrouping = UseGrouping(conformanceUseGrouping(t, options.UseGrouping))
	}
	if options.SignDisplay != "" {
		opts.SignDisplay = SignDisplay(options.SignDisplay)
	}
	if options.LocaleMatcher != "" {
		opts.LocaleMatcher = LocaleMatcher(options.LocaleMatcher)
	}
	if options.NumberingSystem != "" {
		opts.NumberingSystem = options.NumberingSystem
	}
	return opts
}

func conformanceUseGrouping(t *testing.T, value any) string {
	t.Helper()

	switch value := value.(type) {
	case string:
		return value
	case bool:
		if value {
			return "always"
		}
		return "false"
	default:
		t.Fatalf("useGrouping option has type %T, want string or bool", value)
		return ""
	}
}
