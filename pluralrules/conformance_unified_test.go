package pluralrules

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/agentable/go-intl/internal/conformance"
	"github.com/agentable/go-intl/locale"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", func(t *testing.T, fixture conformance.Fixture) {
		rules, err := New(locale.MustParse(fixture.Locale), conformancePluralOptions(t, fixture))
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
		if fixture.Expected == nil {
			t.Fatal("fixture expected is required")
		}
		got, err := pluralFixtureValue(rules, fixture.Feature, input)
		if err != nil {
			t.Fatal(err)
		}
		if got.String() != *fixture.Expected {
			t.Fatalf("Select(%v) = %q, want %q", input, got.String(), *fixture.Expected)
		}
	})
}

func pluralFixtureValue(rules *PluralRules, feature string, input any) (Category, error) {
	switch feature {
	case "", "select":
		return selectFixtureValue(rules, input)
	case "selectRange":
		return selectRangeFixtureValue(rules, input)
	default:
		return Other, fmt.Errorf("pluralrules fixture feature %q: %w", feature, ErrInvalidValue)
	}
}

func selectFixtureValue(rules *PluralRules, input any) (Category, error) {
	switch value := input.(type) {
	case float64:
		return rules.SelectFloat64(value)
	case string:
		return rules.SelectDecimal(value)
	default:
		return Other, fmt.Errorf("pluralrules fixture input %T: %w", input, ErrInvalidValue)
	}
}

func selectRangeFixtureValue(rules *PluralRules, input any) (Category, error) {
	values, ok := input.(map[string]any)
	if !ok {
		return Other, fmt.Errorf("pluralrules selectRange fixture input %T: %w", input, ErrInvalidValue)
	}
	start, err := decimalFixtureValue(values["start"])
	if err != nil {
		return Other, err
	}
	end, err := decimalFixtureValue(values["end"])
	if err != nil {
		return Other, err
	}
	return rules.SelectRangeDecimal(start, end)
}

func decimalFixtureValue(value any) (string, error) {
	switch value := value.(type) {
	case float64:
		return strconv.FormatFloat(value, 'g', -1, 64), nil
	case string:
		return value, nil
	default:
		return "", fmt.Errorf("pluralrules decimal fixture value %T: %w", value, ErrInvalidValue)
	}
}

func conformancePluralOptions(t *testing.T, fixture conformance.Fixture) Options {
	t.Helper()

	var options struct {
		Type                     string `json:"type"`
		MinimumIntegerDigits     *int   `json:"minimumIntegerDigits"`
		MinimumFractionDigits    *int   `json:"minimumFractionDigits"`
		MaximumFractionDigits    *int   `json:"maximumFractionDigits"`
		MinimumSignificantDigits *int   `json:"minimumSignificantDigits"`
		MaximumSignificantDigits *int   `json:"maximumSignificantDigits"`
		RoundingIncrement        *int   `json:"roundingIncrement"`
		RoundingMode             string `json:"roundingMode"`
		RoundingPriority         string `json:"roundingPriority"`
		TrailingZeroDisplay      string `json:"trailingZeroDisplay"`
		Notation                 string `json:"notation"`
		CompactDisplay           string `json:"compactDisplay"`
	}
	if err := json.Unmarshal(fixture.Options, &options); err != nil {
		t.Fatal(err)
	}
	typ, ok := typeFromString(options.Type)
	if !ok {
		return Options{Type: Type(99)}
	}
	out := Options{Type: typ}
	if options.MinimumIntegerDigits != nil {
		out.MinimumIntegerDigits = *options.MinimumIntegerDigits
	}
	switch {
	case options.MinimumFractionDigits != nil && options.MaximumFractionDigits != nil:
		out.FractionDigits = FractionDigits(*options.MinimumFractionDigits, *options.MaximumFractionDigits)
	case options.MinimumFractionDigits != nil:
		out.FractionDigits = MinimumFractionDigits(*options.MinimumFractionDigits)
	case options.MaximumFractionDigits != nil:
		out.FractionDigits = MaximumFractionDigits(*options.MaximumFractionDigits)
	}
	switch {
	case options.MinimumSignificantDigits != nil && options.MaximumSignificantDigits != nil:
		out.SignificantDigits = SignificantDigits(*options.MinimumSignificantDigits, *options.MaximumSignificantDigits)
	case options.MinimumSignificantDigits != nil:
		out.SignificantDigits = MinimumSignificantDigits(*options.MinimumSignificantDigits)
	case options.MaximumSignificantDigits != nil:
		out.SignificantDigits = MaximumSignificantDigits(*options.MaximumSignificantDigits)
	}
	if options.RoundingIncrement != nil {
		out.RoundingIncrement = *options.RoundingIncrement
	}
	if options.RoundingMode != "" {
		out.RoundingMode = RoundingMode(options.RoundingMode)
	}
	if options.RoundingPriority != "" {
		out.RoundingPriority = RoundingPriority(options.RoundingPriority)
	}
	if options.TrailingZeroDisplay != "" {
		out.TrailingZeroDisplay = TrailingZeroDisplay(options.TrailingZeroDisplay)
	}
	if options.Notation != "" {
		out.Notation = Notation(options.Notation)
	}
	if options.CompactDisplay != "" {
		out.CompactDisplay = CompactDisplay(options.CompactDisplay)
	}
	return out
}
