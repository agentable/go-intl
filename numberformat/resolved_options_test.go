package numberformat

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
)

func TestNumberFormatUsesDefaultLocaleProvider(t *testing.T) {
	restore := ecma402.OverrideDefaultLocaleForTest("fr")
	t.Cleanup(restore)

	format, err := New(nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.ResolvedOptions().Locale.String(); got != "fr" {
		t.Fatalf("Locale = %q, want fr", got)
	}

	format, err = New(locale.List{}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.ResolvedOptions().Locale.String(); got != "fr" {
		t.Fatalf("Locale for empty list = %q, want fr", got)
	}
}

func TestNumberFormatResolvedOptionsDefaults(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	got := format.ResolvedOptions()

	if got.Locale.Tag().String() != "en" {
		t.Fatalf("Locale = %s, want en", got.Locale.Tag().String())
	}
	if got.NumberingSystem != "latn" {
		t.Fatalf("NumberingSystem = %q, want latn", got.NumberingSystem)
	}
	if got.Style != "decimal" {
		t.Fatalf("Style = %q, want decimal", got.Style)
	}
	if got.MinimumIntegerDigits != 1 {
		t.Fatalf("MinimumIntegerDigits = %d, want 1", got.MinimumIntegerDigits)
	}
	if got.MinimumFractionDigits == nil || got.MaximumFractionDigits == nil ||
		*got.MinimumFractionDigits != 0 || *got.MaximumFractionDigits != 3 {
		t.Fatalf("fraction digits = %v/%v, want 0/3", got.MinimumFractionDigits, got.MaximumFractionDigits)
	}
	if got.MinimumSignificantDigits != nil || got.MaximumSignificantDigits != nil {
		t.Fatalf("significant digits visible under fractionDigits roundingType: %v/%v",
			got.MinimumSignificantDigits, got.MaximumSignificantDigits)
	}
	if got.UseGrouping != "auto" {
		t.Fatalf("UseGrouping = %q, want auto", got.UseGrouping)
	}
	if got.Notation != "standard" {
		t.Fatalf("Notation = %q, want standard", got.Notation)
	}
	if got.RoundingIncrement != 1 {
		t.Fatalf("RoundingIncrement = %d, want 1", got.RoundingIncrement)
	}
	if got.RoundingMode != "halfExpand" {
		t.Fatalf("RoundingMode = %q, want halfExpand", got.RoundingMode)
	}
	if got.RoundingPriority != "auto" {
		t.Fatalf("RoundingPriority = %q, want auto", got.RoundingPriority)
	}
	if got.TrailingZeroDisplay != "auto" {
		t.Fatalf("TrailingZeroDisplay = %q, want auto", got.TrailingZeroDisplay)
	}
	if got.Currency != nil || got.CurrencyDisplay != nil || got.CurrencySign != nil {
		t.Fatalf("currency slots = %v/%v/%v, want omitted", got.Currency, got.CurrencyDisplay, got.CurrencySign)
	}
	if got.Unit != nil || got.UnitDisplay != nil {
		t.Fatalf("unit slots = %v/%v, want omitted", got.Unit, got.UnitDisplay)
	}
	if got.CompactDisplay != nil {
		t.Fatalf("CompactDisplay = %v, want omitted", got.CompactDisplay)
	}
}

func TestNumberFormatResolvedOptionsPreservesEveryRoundingMode(t *testing.T) {
	t.Parallel()

	modes := []RoundingMode{
		CeilRoundingMode,
		FloorRoundingMode,
		ExpandRoundingMode,
		TruncRoundingMode,
		HalfCeilRoundingMode,
		HalfFloorRoundingMode,
		HalfExpandRoundingMode,
		HalfTruncRoundingMode,
		HalfEvenRoundingMode,
	}
	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, "en")}, Options{RoundingMode: stringPtr(mode)})
			if err != nil {
				t.Fatal(err)
			}
			if got := format.ResolvedOptions().RoundingMode; got != mode {
				t.Fatalf("ResolvedOptions().RoundingMode = %q, want %q", got, mode)
			}
		})
	}
}

func TestNumberFormatResolvedOptionsNumberingSystem(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-u-nu-latn")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.ResolvedOptions().NumberingSystem; got != "latn" {
		t.Fatalf("NumberingSystem = %q, want latn", got)
	}

	format, err = New(locale.List{intltest.Locale(t, "en-u-nu-arab")}, Options{NumberingSystem: stringPtr("latn")})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.ResolvedOptions().NumberingSystem; got != "latn" {
		t.Fatalf("NumberingSystem = %q, want explicit option latn", got)
	}
}

func TestNumberFormatUsesMatchedNumberDataLocale(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "de-DE")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.ResolvedOptions().NumberingSystem; got != "latn" {
		t.Fatalf("NumberingSystem = %q, want latn", got)
	}
	if got := format.Format(Float(1234)); got != "1.234" {
		t.Fatalf("Format(1234) = %q, want German grouping %q", got, "1.234")
	}
}

func TestNumberFormatStyleRequiresCompanionOption(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		options      Options
		wantName     string
		wantExpected string
	}{
		{
			name:         "currency",
			options:      Options{Style: stringPtr(CurrencyStyle)},
			wantName:     "currency",
			wantExpected: `a currency code when style is "currency"`,
		},
		{
			name:         "unit",
			options:      Options{Style: stringPtr(UnitStyle)},
			wantName:     "unit",
			wantExpected: `a sanctioned unit identifier when style is "unit"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{intltest.Locale(t, "en-US")}, tc.options)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
			}
			testcontract.AssertOptionError(t, err, "numberformat", intlerr.InvalidOption, tc.wantName, "", "en-US")
			testcontract.AssertOptionExpected(t, err, tc.wantExpected)
		})
	}
}

func TestNumberFormatRejectsInvalidCurrencyCode(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("US")})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	}
	testcontract.AssertOptionError(t, err, "numberformat", intlerr.InvalidOption, "currency", "US", "en-US")
	testcontract.AssertOptionExpected(t, err, "a three-letter ASCII currency code")
}

func TestNumberFormatRejectsInvalidCurrencyCodeWhenStyleIsNotCurrency(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Currency: stringPtr("US")})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	}
}

func TestNumberFormatRejectsEmptyCurrencyCode(t *testing.T) {
	t.Parallel()

	for _, opts := range []Options{
		{Currency: stringPtr("")},
		{Style: stringPtr(CurrencyStyle), Currency: stringPtr("")},
	} {
		_, err := New(locale.List{intltest.Locale(t, "en-US")}, opts)
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Fatalf("New(%+v) error = %v, want intlerr.ErrInvalidOption", opts, err)
		}
		testcontract.AssertOptionError(t, err, "numberformat", intlerr.InvalidOption, "currency", "", "en-US")
		testcontract.AssertOptionExpected(t, err, "a three-letter ASCII currency code")
	}
}

func TestNumberFormatRejectsNonASCIICurrencyCode(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("円円円")})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	}
}

func TestNumberFormatNormalizesCurrencyInConstructor(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("usd")})
	if err != nil {
		t.Fatalf("New(currency=usd) error = %v", err)
	}
	if got := format.ResolvedOptions().Currency; got == nil || *got != "USD" {
		t.Fatalf("ResolvedOptions().Currency = %v, want USD", got)
	}
}

func TestNumberFormatRejectsInvalidUnit(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(UnitStyle), Unit: stringPtr("bad-unit")})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	}
	testcontract.AssertOptionError(t, err, "numberformat", intlerr.InvalidOption, "unit", "bad-unit", "en")
	testcontract.AssertOptionExpected(t, err, "a sanctioned unit identifier or <unit>-per-<unit> compound")
}

func TestNumberFormatRejectsInvalidUnitWhenStyleIsNotUnit(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{intltest.Locale(t, "en")}, Options{Unit: stringPtr("bad-unit")})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	}
}

func TestNumberFormatRejectsEmptyUnit(t *testing.T) {
	t.Parallel()

	for _, opts := range []Options{
		{Unit: stringPtr("")},
		{Style: stringPtr(UnitStyle), Unit: stringPtr("")},
	} {
		_, err := New(locale.List{intltest.Locale(t, "en")}, opts)
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Fatalf("New(%+v) error = %v, want intlerr.ErrInvalidOption", opts, err)
		}
		testcontract.AssertOptionError(t, err, "numberformat", intlerr.InvalidOption, "unit", "", "en")
		testcontract.AssertOptionExpected(t, err, "a sanctioned unit identifier or <unit>-per-<unit> compound")
	}
}

func TestNumberFormatAcceptsSanctionedUnits(t *testing.T) {
	t.Parallel()

	for _, unit := range []string{"meter", "kilometer", "meter-per-second", "microsecond", "nanosecond-per-second"} {
		format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(UnitStyle), Unit: stringPtr(unit)})
		if err != nil {
			t.Fatalf("New(Options{Unit: stringPtr(%q)}) error = %v, want nil", unit, err)
		}
		if got := format.ResolvedOptions().Unit; got == nil || *got != unit {
			t.Fatalf("ResolvedOptions().Unit = %v, want %q", got, unit)
		}
	}
}

func TestNumberFormatResolvedOptionsExposeOnlyStyleSlots(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Currency: stringPtr("USD"),
		Unit:     stringPtr("meter"),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := format.ResolvedOptions()
	if got.Currency != nil || got.CurrencyDisplay != nil || got.CurrencySign != nil {
		t.Fatalf("decimal currency slots = %v/%v/%v, want omitted", got.Currency, got.CurrencyDisplay, got.CurrencySign)
	}
	if got.Unit != nil || got.UnitDisplay != nil {
		t.Fatalf("decimal unit slots = %v/%v, want omitted", got.Unit, got.UnitDisplay)
	}

	format, err = New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), Unit: stringPtr("meter")})
	if err != nil {
		t.Fatal(err)
	}
	got = format.ResolvedOptions()
	if got.Currency == nil || *got.Currency != "USD" ||
		got.CurrencyDisplay == nil || *got.CurrencyDisplay != CurrencyDisplaySymbol ||
		got.CurrencySign == nil || *got.CurrencySign != StandardCurrencySign {
		t.Fatalf("currency slots = %v/%v/%v, want USD/symbol/standard", got.Currency, got.CurrencyDisplay, got.CurrencySign)
	}
	if got.Unit != nil || got.UnitDisplay != nil {
		t.Fatalf("currency unit slots = %v/%v, want omitted", got.Unit, got.UnitDisplay)
	}

	format, err = New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(UnitStyle), Unit: stringPtr("meter"), Currency: stringPtr("USD")})
	if err != nil {
		t.Fatal(err)
	}
	got = format.ResolvedOptions()
	if got.Currency != nil || got.CurrencyDisplay != nil || got.CurrencySign != nil {
		t.Fatalf("unit currency slots = %v/%v/%v, want omitted", got.Currency, got.CurrencyDisplay, got.CurrencySign)
	}
	if got.Unit == nil || *got.Unit != "meter" ||
		got.UnitDisplay == nil || *got.UnitDisplay != ShortUnitDisplay {
		t.Fatalf("unit slots = %v/%v, want meter/short", got.Unit, got.UnitDisplay)
	}
}

func TestNumberFormatRejectsCaseChangedUnit(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(UnitStyle), Unit: stringPtr("METER")})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	}
}

func TestNumberFormatRejectsInvalidStringOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opt  Options
	}{
		{name: "style", opt: Options{Style: stringPtr(Style("bad"))}},
		{name: "style empty", opt: Options{Style: stringPtr("")}},
		{name: "notation", opt: Options{Notation: stringPtr(Notation("bad"))}},
		{name: "notation empty", opt: Options{Notation: stringPtr("")}},
		{name: "compact display", opt: Options{CompactDisplay: stringPtr(CompactDisplay("bad"))}},
		{name: "compact display empty", opt: Options{CompactDisplay: stringPtr("")}},
		{name: "currency display", opt: Options{CurrencyDisplay: stringPtr(CurrencyDisplay("bad"))}},
		{name: "currency display empty", opt: Options{CurrencyDisplay: stringPtr("")}},
		{name: "currency sign", opt: Options{CurrencySign: stringPtr(CurrencySign("bad"))}},
		{name: "currency sign empty", opt: Options{CurrencySign: stringPtr("")}},
		{name: "unit display", opt: Options{UnitDisplay: stringPtr(UnitDisplay("bad"))}},
		{name: "unit display empty", opt: Options{UnitDisplay: stringPtr("")}},
		{name: "sign display", opt: Options{SignDisplay: stringPtr(SignDisplay("bad"))}},
		{name: "sign display empty", opt: Options{SignDisplay: stringPtr("")}},
		{name: "rounding mode", opt: Options{RoundingMode: stringPtr(RoundingMode("bad"))}},
		{name: "rounding mode empty", opt: Options{RoundingMode: stringPtr("")}},
		{name: "rounding priority", opt: Options{RoundingPriority: stringPtr(RoundingPriority("bad"))}},
		{name: "rounding priority empty", opt: Options{RoundingPriority: stringPtr("")}},
		{name: "trailing zero display", opt: Options{TrailingZeroDisplay: stringPtr(TrailingZeroDisplay("bad"))}},
		{name: "trailing zero display empty", opt: Options{TrailingZeroDisplay: stringPtr("")}},
		{name: "locale matcher", opt: Options{LocaleMatcher: stringPtr("bad")}},
		{name: "locale matcher empty", opt: Options{LocaleMatcher: stringPtr("")}},
		{name: "use grouping", opt: Options{UseGrouping: stringPtr(UseGrouping("bad"))}},
		{name: "use grouping empty", opt: Options{UseGrouping: stringPtr("")}},
		{name: "numbering system", opt: Options{NumberingSystem: stringPtr("ab")}},
		{name: "numbering system empty", opt: Options{NumberingSystem: stringPtr("")}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{intltest.Locale(t, "en")}, tc.opt)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
			}
			if tc.name == "numbering system" {
				testcontract.AssertOptionError(t, err, "numberformat", intlerr.InvalidOption, "numberingSystem", "ab", "en")
				testcontract.AssertOptionExpected(t, err, "a Unicode locale extension type")
			}
		})
	}
}

func TestNumberFormatRejectsInvalidNumericOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		opt          Options
		wantExpected string
	}{
		{name: "minimum integer digits zero", opt: Options{MinimumIntegerDigits: intPtr(0)}},
		{name: "minimum integer digits low", opt: Options{MinimumIntegerDigits: intPtr(-1)}},
		{name: "minimum fraction digits low", opt: Options{MinimumFractionDigits: intPtr(-1)}},
		{name: "maximum fraction digits high", opt: Options{MaximumFractionDigits: intPtr(101)}},
		{name: "minimum significant digits low", opt: Options{MinimumSignificantDigits: intPtr(0)}},
		{name: "maximum significant digits high", opt: Options{MaximumSignificantDigits: intPtr(22)}},
		{name: "rounding increment zero", opt: Options{RoundingIncrement: intPtr(0)}},
		{
			name:         "rounding increment invalid",
			opt:          Options{RoundingIncrement: intPtr(3)},
			wantExpected: "one of 1, 2, 5, 10, 20, 25, 50, 100, 200, 250, 500, 1000, 2000, 2500, 5000",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{intltest.Locale(t, "en")}, tc.opt)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
			}
			if tc.wantExpected != "" {
				testcontract.AssertOptionError(t, err, "numberformat", intlerr.InvalidOption, "roundingIncrement", "3", "en")
				testcontract.AssertOptionExpected(t, err, tc.wantExpected)
			}
		})
	}
}

func TestNumberFormatAcceptsDataBackedOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts Options
	}{
		{name: "currency sign", opts: Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencySign: stringPtr(AccountingCurrencySign)}},
		{name: "compact display", opts: Options{Notation: stringPtr(CompactNotation), CompactDisplay: stringPtr(LongCompactDisplay)}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := New(locale.List{intltest.Locale(t, "en")}, tc.opts); err != nil {
				t.Fatalf("New() error = %v, want nil", err)
			}
		})
	}
}

func TestNumberFormatResolvedOptionsSignificantDigits(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{MinimumSignificantDigits: intPtr(3)})
	if err != nil {
		t.Fatal(err)
	}
	got := format.ResolvedOptions()
	if got.MinimumSignificantDigits == nil || got.MaximumSignificantDigits == nil ||
		*got.MinimumSignificantDigits != 3 || *got.MaximumSignificantDigits != 21 {
		t.Fatalf("significant digits = %v/%v, want 3/21", got.MinimumSignificantDigits, got.MaximumSignificantDigits)
	}
	// ECMA-402 §15.5.2: with roundingType=significantDigits, fraction-digit
	// fields must be absent (nil), not zero.
	if got.MinimumFractionDigits != nil || got.MaximumFractionDigits != nil {
		t.Fatalf("fraction digits visible under significantDigits roundingType: %v/%v",
			got.MinimumFractionDigits, got.MaximumFractionDigits)
	}

	format, err = New(locale.List{intltest.Locale(t, "en")}, Options{MaximumSignificantDigits: intPtr(3)})
	if err != nil {
		t.Fatal(err)
	}
	got = format.ResolvedOptions()
	if got.MinimumSignificantDigits == nil || got.MaximumSignificantDigits == nil ||
		*got.MinimumSignificantDigits != 1 || *got.MaximumSignificantDigits != 3 {
		t.Fatalf("significant digits = %v/%v, want 1/3", got.MinimumSignificantDigits, got.MaximumSignificantDigits)
	}
}

func TestNumberFormatResolvedOptionsRoundingPriority(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{MaximumFractionDigits: intPtr(2), RoundingPriority: stringPtr(MorePrecisionRoundingPriority)})
	if err != nil {
		t.Fatal(err)
	}
	got := format.ResolvedOptions()
	if got.RoundingPriority != "morePrecision" {
		t.Fatalf("RoundingPriority = %q, want morePrecision", got.RoundingPriority)
	}
	if got.MinimumSignificantDigits == nil || got.MaximumSignificantDigits == nil ||
		*got.MinimumSignificantDigits != 1 || *got.MaximumSignificantDigits != 21 {
		t.Fatalf("significant defaults = %v/%v, want 1/21", got.MinimumSignificantDigits, got.MaximumSignificantDigits)
	}
	if got.MaximumFractionDigits == nil || *got.MaximumFractionDigits != 2 {
		t.Fatalf("MaximumFractionDigits = %v, want 2", got.MaximumFractionDigits)
	}
}

func TestNumberFormatResolvedOptionsCompactDefaults(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Notation: stringPtr(CompactNotation)})
	if err != nil {
		t.Fatal(err)
	}
	got := format.ResolvedOptions()
	if got.UseGrouping != UseGroupingMin2 {
		t.Fatalf("UseGrouping = %q, want min2", got.UseGrouping)
	}
	if got.RoundingPriority != MorePrecisionRoundingPriority {
		t.Fatalf("RoundingPriority = %q, want morePrecision", got.RoundingPriority)
	}
	if got.MinimumFractionDigits == nil || got.MaximumFractionDigits == nil ||
		*got.MinimumFractionDigits != 0 || *got.MaximumFractionDigits != 0 {
		t.Fatalf("fraction digits = %v/%v, want 0/0", got.MinimumFractionDigits, got.MaximumFractionDigits)
	}
	if got.MinimumSignificantDigits == nil || got.MaximumSignificantDigits == nil ||
		*got.MinimumSignificantDigits != 1 || *got.MaximumSignificantDigits != 2 {
		t.Fatalf("significant digits = %v/%v, want 1/2", got.MinimumSignificantDigits, got.MaximumSignificantDigits)
	}
	if got.CompactDisplay == nil || *got.CompactDisplay != ShortCompactDisplay {
		t.Fatalf("CompactDisplay = %v, want short", got.CompactDisplay)
	}

	format, err = New(locale.List{intltest.Locale(t, "en")}, Options{Notation: stringPtr(CompactNotation), UseGrouping: stringPtr(UseGroupingAlways)})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.ResolvedOptions().UseGrouping; got != UseGroupingAlways {
		t.Fatalf("explicit UseGrouping = %q, want always", got)
	}
}

func TestNumberFormatResolvedOptionsCurrencyDefaultsOnlyForStandardNotation(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Style:    stringPtr(CurrencyStyle),
		Currency: stringPtr("USD"),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := format.ResolvedOptions()
	if got.MinimumFractionDigits == nil || got.MaximumFractionDigits == nil ||
		*got.MinimumFractionDigits != 2 || *got.MaximumFractionDigits != 2 {
		t.Fatalf("standard currency fraction digits = %v/%v, want 2/2", got.MinimumFractionDigits, got.MaximumFractionDigits)
	}

	format, err = New(locale.List{intltest.Locale(t, "en")}, Options{
		Style:    stringPtr(CurrencyStyle),
		Currency: stringPtr("USD"),
		Notation: stringPtr(ScientificNotation),
	})
	if err != nil {
		t.Fatal(err)
	}
	got = format.ResolvedOptions()
	if got.MinimumFractionDigits == nil || got.MaximumFractionDigits == nil ||
		*got.MinimumFractionDigits != 0 || *got.MaximumFractionDigits != 3 {
		t.Fatalf("scientific currency fraction digits = %v/%v, want 0/3", got.MinimumFractionDigits, got.MaximumFractionDigits)
	}
}

func TestNumberFormatResolvedOptionsIsSnapshot(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Notation: stringPtr(ScientificNotation)})
	if err != nil {
		t.Fatal(err)
	}
	first := format.ResolvedOptions()
	first.Notation = "compact"
	if first.MinimumFractionDigits == nil {
		t.Fatal("MinimumFractionDigits = nil, want snapshot scalar")
	}
	*first.MinimumFractionDigits = 9
	if first.Notation != "compact" {
		t.Fatalf("mutated snapshot Notation = %q, want compact", first.Notation)
	}

	got := format.ResolvedOptions()
	if got.Notation != "scientific" {
		t.Fatalf("Notation after mutating snapshot = %q, want scientific", got.Notation)
	}
	if got.MinimumFractionDigits == nil || *got.MinimumFractionDigits != 0 {
		t.Fatalf("MinimumFractionDigits after mutating snapshot = %v, want 0", got.MinimumFractionDigits)
	}

	format, err = New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD")})
	if err != nil {
		t.Fatal(err)
	}
	first = format.ResolvedOptions()
	if first.Currency == nil || first.CurrencyDisplay == nil {
		t.Fatal("currency resolved options omitted, want snapshot scalars")
	}
	*first.Currency = "EUR"
	*first.CurrencyDisplay = CurrencyDisplayCode
	got = format.ResolvedOptions()
	if got.Currency == nil || *got.Currency != "USD" ||
		got.CurrencyDisplay == nil || *got.CurrencyDisplay != CurrencyDisplaySymbol {
		t.Fatalf("currency resolved options after mutating snapshot = %v/%v, want USD/symbol", got.Currency, got.CurrencyDisplay)
	}
}

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()

	requested := locale.List{intltest.Locale(t, "fr-FR"), intltest.Locale(t, "en-US-u-nu-latn"), intltest.Locale(t, "ban")}
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: stringPtr(LookupLocaleMatcher)})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	testcontract.AssertLocaleListStrings(t, "SupportedLocalesOf()", got, []string{"fr-FR", "en-US-u-nu-latn"})
}

func TestSupportedLocalesOfRecognizesDerivedAvailableLocale(t *testing.T) {
	t.Parallel()

	requested := intltest.LocaleList(t, "zh-HK-u-nu-latn")
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: stringPtr(LookupLocaleMatcher)})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	testcontract.AssertLocaleListStrings(t, "SupportedLocalesOf()", got, []string{"zh-HK-u-nu-latn"})

	format, err := New(requested, Options{LocaleMatcher: stringPtr(LookupLocaleMatcher)})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if got := format.ResolvedOptions().Locale.String(); got != "zh-HK-u-nu-latn" {
		t.Fatalf("ResolvedOptions().Locale = %q, want zh-HK-u-nu-latn", got)
	}
}

func TestBestFitUsesGeneratedRelatedLanguageProfile(t *testing.T) {
	t.Parallel()

	requested := intltest.LocaleList(t, "nn")
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: stringPtr(BestFitLocaleMatcher)})
	if err != nil {
		t.Fatalf("SupportedLocalesOf(nn) error = %v", err)
	}
	testcontract.AssertLocaleListStrings(t, "SupportedLocalesOf(nn)", got, []string{"nn"})

	format, err := New(requested, Options{LocaleMatcher: stringPtr(BestFitLocaleMatcher)})
	if err != nil {
		t.Fatalf("New(nn) error = %v", err)
	}
	if got := format.ResolvedOptions().Locale.String(); got != "nb" {
		t.Fatalf("New(nn).ResolvedOptions().Locale = %q, want CLDR related language nb", got)
	}
}

func TestSupportedLocalesOfErrors(t *testing.T) {
	t.Parallel()

	requested := locale.List{intltest.Locale(t, "en-US")}
	for _, matcher := range []string{"bad", ""} {
		t.Run(matcher, func(t *testing.T) {
			t.Parallel()
			_, err := SupportedLocalesOf(requested, Options{LocaleMatcher: stringPtr(matcher)})
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("SupportedLocalesOf(invalid matcher) error = %v, want intlerr.ErrInvalidOption", err)
			}
			testcontract.AssertOptionError(t, err, "numberformat", intlerr.InvalidOption, "localeMatcher", matcher, "en-US")
			testcontract.AssertOptionExpected(t, err, `one of "lookup", "best fit"`)
		})
	}
}
