package numberformat

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", func(t *testing.T, fixture conformance.Fixture) {
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
		if len(fixture.ExpectedResolved) != 0 {
			assertNumberFormatResolvedOptions(t, fixture, format.ResolvedOptions())
		}
		var input any
		if err := json.Unmarshal(fixture.Input, &input); err != nil {
			t.Fatal(err)
		}
		if fixture.ExpectedRange != nil || len(fixture.ExpectedRangeParts) > 0 {
			rangeInput := conformanceNumberRangeInput(t, fixture)
			if fixture.ExpectedRange != nil {
				got, err := format.formatRangeValue(rangeInput.Start, rangeInput.End)
				if err != nil {
					t.Fatalf("FormatRange(%v, %v) error = %v", rangeInput.Start, rangeInput.End, err)
				}
				testcontract.AssertExpectedRange(t, "FormatRange", got, fixture.ExpectedRange)
			}
			if len(fixture.ExpectedRangeParts) > 0 {
				parts, err := format.formatRangeToPartsValue(rangeInput.Start, rangeInput.End)
				if err != nil {
					t.Fatalf("FormatRangeToParts(%v, %v) error = %v", rangeInput.Start, rangeInput.End, err)
				}
				testcontract.AssertRangeParts(t, "FormatRangeToParts", parts, fixture.ExpectedRangeParts, conformanceNumberRangePart)
			}
			return
		}
		want := fixture.RequiredExpected(t)
		if got := format.formatValue(input); got != want {
			t.Fatalf("Format(%v) = %q, want %q", input, got, want)
		}
		if len(fixture.ExpectedParts) > 0 {
			parts := format.formatToPartsValue(input)
			testcontract.AssertParts(t, "FormatToParts", parts, fixture.ExpectedParts, conformanceNumberPart)
		}
	})
}

type formatFixture struct {
	Name                  string  `json:"name"`
	Locale                string  `json:"locale"`
	Style                 *string `json:"style,omitempty"`
	Currency              *string `json:"currency,omitempty"`
	Notation              *string `json:"notation,omitempty"`
	MaximumFractionDigits *int    `json:"maximumFractionDigits,omitempty"`
	Input                 any     `json:"input"`
	Want                  string  `json:"want"`
}

func TestNumberFormatConformanceFixtures(t *testing.T) {
	t.Parallel()

	var fixtures []formatFixture
	intltest.ReadFixture(t, "testdata/format.json", &fixtures)
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()

			format, err := New(intltest.LocaleList(t, fixture.Locale), Options{
				Style:                 fixture.Style,
				Currency:              fixture.Currency,
				Notation:              fixture.Notation,
				MaximumFractionDigits: fixture.MaximumFractionDigits,
			})
			if err != nil {
				t.Fatal(err)
			}
			if got := format.formatValue(fixture.Input); got != fixture.Want {
				t.Fatalf("Format(%v) = %q, want %q", fixture.Input, got, fixture.Want)
			}
		})
	}
}

func TestConformanceNumberOptionsPreserveExplicitEmptyString(t *testing.T) {
	t.Parallel()

	_, err := New(intltest.LocaleList(t, "en"), conformanceNumberOptions(t, conformance.Fixture{
		Options: json.RawMessage(`{"style":""}`),
	}))
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want %v", err, intlerr.ErrInvalidOption)
	}
	testcontract.AssertOptionError(t, err, "numberformat", intlerr.InvalidOption, "style", "", "en")
	testcontract.AssertOptionExpected(t, err, `one of "decimal", "percent", "currency", "unit"`)
}

func assertNumberFormatResolvedOptions(t *testing.T, fixture conformance.Fixture, got ResolvedOptions) {
	t.Helper()

	want := testcontract.ExpectedResolvedOptions(t, fixture)
	testcontract.AssertResolvedString(t, want, "locale", got.Locale.String())
	testcontract.AssertResolvedString(t, want, "numberingSystem", got.NumberingSystem)
	testcontract.AssertResolvedString(t, want, "style", string(got.Style))
	testcontract.AssertResolvedOptionalString(t, want, "currency", got.Currency)
	testcontract.AssertResolvedOptionalString(t, want, "currencyDisplay", got.CurrencyDisplay)
	testcontract.AssertResolvedOptionalString(t, want, "currencySign", got.CurrencySign)
	testcontract.AssertResolvedOptionalString(t, want, "unit", got.Unit)
	testcontract.AssertResolvedOptionalString(t, want, "unitDisplay", got.UnitDisplay)
	testcontract.AssertResolvedInt(t, want, "minimumIntegerDigits", got.MinimumIntegerDigits)
	testcontract.AssertResolvedOptionalInt(t, want, "minimumFractionDigits", got.MinimumFractionDigits)
	testcontract.AssertResolvedOptionalInt(t, want, "maximumFractionDigits", got.MaximumFractionDigits)
	testcontract.AssertResolvedOptionalInt(t, want, "minimumSignificantDigits", got.MinimumSignificantDigits)
	testcontract.AssertResolvedOptionalInt(t, want, "maximumSignificantDigits", got.MaximumSignificantDigits)
	testcontract.AssertResolvedString(t, want, "useGrouping", string(got.UseGrouping))
	testcontract.AssertResolvedString(t, want, "notation", string(got.Notation))
	testcontract.AssertResolvedOptionalString(t, want, "compactDisplay", got.CompactDisplay)
	testcontract.AssertResolvedString(t, want, "signDisplay", string(got.SignDisplay))
	testcontract.AssertResolvedInt(t, want, "roundingIncrement", got.RoundingIncrement)
	testcontract.AssertResolvedString(t, want, "roundingMode", string(got.RoundingMode))
	testcontract.AssertResolvedString(t, want, "roundingPriority", string(got.RoundingPriority))
	testcontract.AssertResolvedString(t, want, "trailingZeroDisplay", string(got.TrailingZeroDisplay))
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
		Style                    *string `json:"style"`
		Currency                 *string `json:"currency"`
		CurrencyDisplay          *string `json:"currencyDisplay"`
		CurrencySign             *string `json:"currencySign"`
		Unit                     *string `json:"unit"`
		UnitDisplay              *string `json:"unitDisplay"`
		MinimumIntegerDigits     *int    `json:"minimumIntegerDigits"`
		MinimumFractionDigits    *int    `json:"minimumFractionDigits"`
		MaximumFractionDigits    *int    `json:"maximumFractionDigits"`
		MinimumSignificantDigits *int    `json:"minimumSignificantDigits"`
		MaximumSignificantDigits *int    `json:"maximumSignificantDigits"`
		RoundingIncrement        *int    `json:"roundingIncrement"`
		RoundingPriority         *string `json:"roundingPriority"`
		RoundingMode             *string `json:"roundingMode"`
		TrailingZeroDisplay      *string `json:"trailingZeroDisplay"`
		Notation                 *string `json:"notation"`
		CompactDisplay           *string `json:"compactDisplay"`
		UseGrouping              any     `json:"useGrouping"`
		SignDisplay              *string `json:"signDisplay"`
		LocaleMatcher            *string `json:"localeMatcher"`
		NumberingSystem          *string `json:"numberingSystem"`
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

func conformanceUseGroupingOption(t *testing.T, value any) *string {
	t.Helper()

	if value == nil {
		return nil
	}
	return stringPtr(conformanceUseGrouping(t, value))
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
