package numberformat

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", func(t *testing.T, fixture conformance.Fixture) {
		loc := intltest.Locale(t, fixture.Locale)
		format, err := New(locale.List{loc}, conformanceNumberOptions(t, fixture))
		if fixture.ErrorCode != "" {
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
			}
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

type formatFixture struct {
	Name                  string `json:"name"`
	Locale                string `json:"locale"`
	Style                 string `json:"style,omitempty"`
	Currency              string `json:"currency,omitempty"`
	Notation              string `json:"notation,omitempty"`
	MaximumFractionDigits *int   `json:"maximumFractionDigits,omitempty"`
	Input                 any    `json:"input"`
	Want                  string `json:"want"`
}

func TestNumberFormatConformanceFixtures(t *testing.T) {
	t.Parallel()

	var fixtures []formatFixture
	intltest.ReadFixture(t, "testdata/format.json", &fixtures)
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()

			var opts Options
			if fixture.Style != "" {
				opts.Style = Style(fixture.Style)
			}
			if fixture.Currency != "" {
				opts.Currency = Currency(fixture.Currency)
			}
			if fixture.Notation != "" {
				opts.Notation = Notation(fixture.Notation)
			}
			if fixture.MaximumFractionDigits != nil {
				opts.MaximumFractionDigits = fixture.MaximumFractionDigits
			}
			format, err := New(intltest.LocaleList(t, fixture.Locale), opts)
			if err != nil {
				t.Fatal(err)
			}
			if got := format.formatValue(fixture.Input); got != fixture.Want {
				t.Fatalf("Format(%v) = %q, want %q", fixture.Input, got, fixture.Want)
			}
		})
	}
}

func assertNumberFormatResolvedOptions(t *testing.T, fixture conformance.Fixture, got ResolvedOptions) {
	t.Helper()

	var want map[string]json.RawMessage
	if err := json.Unmarshal(fixture.ExpectedResolved, &want); err != nil {
		t.Fatal(err)
	}
	assertResolvedString(t, want, "locale", got.Locale.String())
	assertResolvedString(t, want, "numberingSystem", got.NumberingSystem)
	assertResolvedString(t, want, "style", string(got.Style))
	assertResolvedString(t, want, "currency", got.Currency)
	assertResolvedString(t, want, "currencyDisplay", string(got.CurrencyDisplay))
	assertResolvedString(t, want, "currencySign", string(got.CurrencySign))
	assertResolvedString(t, want, "unit", got.Unit)
	assertResolvedString(t, want, "unitDisplay", string(got.UnitDisplay))
	assertResolvedInt(t, want, "minimumIntegerDigits", got.MinimumIntegerDigits)
	assertResolvedOptionalInt(t, want, "minimumFractionDigits", got.MinimumFractionDigits)
	assertResolvedOptionalInt(t, want, "maximumFractionDigits", got.MaximumFractionDigits)
	assertResolvedOptionalInt(t, want, "minimumSignificantDigits", got.MinimumSignificantDigits)
	assertResolvedOptionalInt(t, want, "maximumSignificantDigits", got.MaximumSignificantDigits)
	assertResolvedString(t, want, "useGrouping", string(got.UseGrouping))
	assertResolvedString(t, want, "notation", string(got.Notation))
	assertResolvedString(t, want, "compactDisplay", string(got.CompactDisplay))
	assertResolvedString(t, want, "signDisplay", string(got.SignDisplay))
	assertResolvedInt(t, want, "roundingIncrement", got.RoundingIncrement)
	assertResolvedString(t, want, "roundingMode", string(got.RoundingMode))
	assertResolvedString(t, want, "roundingPriority", string(got.RoundingPriority))
	assertResolvedString(t, want, "trailingZeroDisplay", string(got.TrailingZeroDisplay))
}

func assertResolvedString(t *testing.T, values map[string]json.RawMessage, name, got string) {
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

func assertResolvedInt(t *testing.T, values map[string]json.RawMessage, name string, got int) {
	t.Helper()

	raw, ok := values[name]
	if !ok {
		return
	}
	var want int
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("expectedResolvedOptions.%s = %s: %v", name, raw, err)
	}
	if got != want {
		t.Fatalf("ResolvedOptions().%s = %d, want %d", name, got, want)
	}
}

func assertResolvedOptionalInt(t *testing.T, values map[string]json.RawMessage, name string, got *int) {
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
		opts.Currency = Currency(options.Currency)
	}
	if options.CurrencyDisplay != "" {
		opts.CurrencyDisplay = CurrencyDisplay(options.CurrencyDisplay)
	}
	if options.CurrencySign != "" {
		opts.CurrencySign = CurrencySign(options.CurrencySign)
	}
	if options.Unit != "" {
		opts.Unit = Unit(options.Unit)
	}
	if options.UnitDisplay != "" {
		opts.UnitDisplay = UnitDisplay(options.UnitDisplay)
	}
	if options.MinimumIntegerDigits != nil {
		opts.MinimumIntegerDigits = options.MinimumIntegerDigits
	}
	if options.MinimumFractionDigits != nil {
		opts.MinimumFractionDigits = options.MinimumFractionDigits
	}
	if options.MaximumFractionDigits != nil {
		opts.MaximumFractionDigits = options.MaximumFractionDigits
	}
	if options.MinimumSignificantDigits != nil {
		opts.MinimumSignificantDigits = options.MinimumSignificantDigits
	}
	if options.MaximumSignificantDigits != nil {
		opts.MaximumSignificantDigits = options.MaximumSignificantDigits
	}
	if options.RoundingIncrement != nil {
		opts.RoundingIncrement = options.RoundingIncrement
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
