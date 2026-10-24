package numberformat

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"math"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", runNumberConformanceFixture)
}

func runNumberConformanceFixture(t *testing.T, fixture conformance.Fixture) {
	t.Helper()

	loc := intltest.Locale(t, fixture.Locale)
	format, err := New(locale.List{loc}, conformanceNumberOptions(t, fixture))
	if testcontract.AssertErrorCode(t, "New()", err, fixture.ErrorCode, func(code string) error {
		return conformanceNumberError(t, code)
	}) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if fixture.ExpectedResolved != nil {
		testcontract.AssertResolvedOptionsJSON(t, format.ResolvedOptions(), fixture.ExpectedResolved)
	}
	if fixture.ExpectedRange != nil || fixture.ExpectedRangeParts != nil {
		rangeInput := conformanceNumberRangeInput(t, fixture)
		if fixture.ExpectedRange != nil {
			got, err := format.FormatRange(rangeInput.Start, rangeInput.End)
			if err != nil {
				t.Fatalf("FormatRange() error = %v", err)
			}
			testcontract.AssertExpectedRange(t, "FormatRange", got, fixture.ExpectedRange)
		}
		if fixture.ExpectedRangeParts != nil {
			parts, err := format.FormatRangeToParts(rangeInput.Start, rangeInput.End)
			if err != nil {
				t.Fatalf("FormatRangeToParts() error = %v", err)
			}
			testcontract.AssertRangeParts(t, "FormatRangeToParts", parts, fixture.ExpectedRangeParts, conformanceNumberRangePart)
		}
		return
	}
	if fixture.Expected == nil && fixture.ExpectedParts == nil {
		return
	}
	input := conformanceNumberInput(t, fixture.Input)
	if fixture.Expected != nil {
		if got := format.Format(input); got != *fixture.Expected {
			t.Fatalf("Format() = %q, want %q", got, *fixture.Expected)
		}
	}
	if fixture.ExpectedParts != nil {
		parts := format.FormatToParts(input)
		testcontract.AssertParts(t, "FormatToParts", parts, fixture.ExpectedParts, conformanceNumberPart)
	}
}

func TestConformanceNumberOptionsPreserveExplicitEmptyString(t *testing.T) {
	t.Parallel()

	_, err := New(intltest.LocaleList(t, "en"), conformanceNumberOptions(t, conformance.Fixture{
		Options: jsontext.Value(`{"style":""}`),
	}))
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want %v", err, intlerr.ErrInvalidOption)
	}
	testcontract.AssertOptionError(t, err, "numberformat", intlerr.InvalidOption, "style", "", "en")
	testcontract.AssertOptionExpected(t, err, `one of "decimal", "percent", "currency", "unit"`)
}

func conformanceNumberPart(part Part) conformance.Part {
	return conformance.Part{Type: string(part.Type), Value: part.Value}
}

func conformanceNumberRangePart(part RangePart) conformance.RangePart {
	return conformance.RangePart{Type: string(part.Type), Value: part.Value, Source: string(part.Source)}
}

func conformanceNumberError(t testing.TB, code string) error {
	t.Helper()

	return testcontract.IntlErrorCode(t, "numberformat", code, "invalid_option")
}

type conformanceNumberRange struct {
	Start Value
	End   Value
}

func conformanceNumberRangeInput(t *testing.T, fixture conformance.Fixture) conformanceNumberRange {
	t.Helper()

	var input struct {
		Start jsontext.Value `json:"start"`
		End   jsontext.Value `json:"end"`
	}
	if err := json.Unmarshal(fixture.Input, &input); err != nil {
		t.Fatal(err)
	}
	return conformanceNumberRange{
		Start: conformanceNumberInput(t, input.Start),
		End:   conformanceNumberInput(t, input.End),
	}
}

func conformanceNumberInput(t testing.TB, raw jsontext.Value) Value {
	t.Helper()

	raw = bytes.TrimSpace(raw)
	if bytes.Equal(raw, []byte("null")) {
		return Float(math.NaN())
	}
	if len(raw) > 0 && raw[0] == '"' {
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			t.Fatal(err)
		}
		input, err := ParseValue(value)
		if err != nil {
			t.Fatal(err)
		}
		return input
	}
	var value float64
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("number input %s: %v", raw, err)
	}
	return Float(value)
}

func conformanceNumberOptions(t *testing.T, fixture conformance.Fixture) Options {
	t.Helper()

	var options struct {
		Style                    *string        `json:"style"`
		Currency                 *string        `json:"currency"`
		CurrencyDisplay          *string        `json:"currencyDisplay"`
		CurrencySign             *string        `json:"currencySign"`
		Unit                     *string        `json:"unit"`
		UnitDisplay              *string        `json:"unitDisplay"`
		MinimumIntegerDigits     *int           `json:"minimumIntegerDigits"`
		MinimumFractionDigits    *int           `json:"minimumFractionDigits"`
		MaximumFractionDigits    *int           `json:"maximumFractionDigits"`
		MinimumSignificantDigits *int           `json:"minimumSignificantDigits"`
		MaximumSignificantDigits *int           `json:"maximumSignificantDigits"`
		RoundingIncrement        *int           `json:"roundingIncrement"`
		RoundingPriority         *string        `json:"roundingPriority"`
		RoundingMode             *string        `json:"roundingMode"`
		TrailingZeroDisplay      *string        `json:"trailingZeroDisplay"`
		Notation                 *string        `json:"notation"`
		CompactDisplay           *string        `json:"compactDisplay"`
		UseGrouping              jsontext.Value `json:"useGrouping"`
		SignDisplay              *string        `json:"signDisplay"`
		LocaleMatcher            *string        `json:"localeMatcher"`
		NumberingSystem          *string        `json:"numberingSystem"`
	}
	if err := json.Unmarshal(fixture.Options, &options); err != nil {
		t.Fatal(err)
	}
	return Options{
		Style:                    options.Style,
		Currency:                 options.Currency,
		CurrencyDisplay:          options.CurrencyDisplay,
		CurrencySign:             options.CurrencySign,
		Unit:                     options.Unit,
		UnitDisplay:              options.UnitDisplay,
		MinimumIntegerDigits:     options.MinimumIntegerDigits,
		MinimumFractionDigits:    options.MinimumFractionDigits,
		MaximumFractionDigits:    options.MaximumFractionDigits,
		MinimumSignificantDigits: options.MinimumSignificantDigits,
		MaximumSignificantDigits: options.MaximumSignificantDigits,
		RoundingIncrement:        options.RoundingIncrement,
		RoundingPriority:         options.RoundingPriority,
		RoundingMode:             options.RoundingMode,
		TrailingZeroDisplay:      options.TrailingZeroDisplay,
		Notation:                 options.Notation,
		CompactDisplay:           options.CompactDisplay,
		UseGrouping:              conformanceUseGroupingOption(t, options.UseGrouping),
		SignDisplay:              options.SignDisplay,
		LocaleMatcher:            options.LocaleMatcher,
		NumberingSystem:          options.NumberingSystem,
	}
}

func conformanceUseGroupingOption(t *testing.T, raw jsontext.Value) *string {
	t.Helper()

	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil
	}
	if raw[0] == '"' {
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			t.Fatal(err)
		}
		return stringPtr(value)
	}
	var value bool
	if err := json.Unmarshal(raw, &value); err == nil {
		if value {
			return stringPtr("always")
		}
		return stringPtr("false")
	}
	t.Fatalf("useGrouping option must be a string or boolean, got %s", raw)
	return nil
}
