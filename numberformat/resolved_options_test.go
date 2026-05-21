package numberformat

import (
	"errors"
	"strings"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/ecma402"
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

	format, err := New(locale.List{locale.MustParse("en")}, Options{})
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
}

func TestNumberFormatResolvedOptionsNumberingSystem(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-u-nu-latn")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.ResolvedOptions().NumberingSystem; got != "latn" {
		t.Fatalf("NumberingSystem = %q, want latn", got)
	}

	format, err = New(locale.List{locale.MustParse("en-u-nu-arab")}, Options{NumberingSystem: "latn"})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.ResolvedOptions().NumberingSystem; got != "latn" {
		t.Fatalf("NumberingSystem = %q, want explicit option latn", got)
	}
}

func TestNumberFormatUsesMatchedNumberDataLocale(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("de-DE")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.ResolvedOptions().NumberingSystem; got != "latn" {
		t.Fatalf("NumberingSystem = %q, want latn", got)
	}
	if got := format.formatValue(1234); got != "1.234" {
		t.Fatalf("Format(1234) = %q, want German grouping %q", got, "1.234")
	}
}

func TestNumberFormatCurrencyRequiresCurrency(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{locale.MustParse("en-US")}, Options{Style: CurrencyStyle})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	}
	optErr, ok := errors.AsType[*ecma402.OptionError](err)
	if !ok {
		t.Fatalf("New() error = %T, want OptionError", err)
	}
	if optErr.Owner != "numberformat" || optErr.Kind != "invalidOption" || optErr.Name != "currency" || optErr.Value != "" || optErr.Locale != "en-US" {
		t.Fatalf("OptionError = %+v, want numberformat invalid currency for en-US", optErr)
	}
}

func TestNumberFormatRejectsInvalidCurrencyCode(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{locale.MustParse("en-US")}, Options{Style: CurrencyStyle, Currency: CurrencyCode("US")})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	}
}

func TestNumberFormatRejectsInvalidCurrencyCodeWhenStyleIsNotCurrency(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{locale.MustParse("en-US")}, Options{Currency: Currency("US")})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	}
}

func TestNumberFormatRejectsNonASCIICurrencyCode(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{locale.MustParse("en-US")}, Options{Style: CurrencyStyle, Currency: Currency("円円円")})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	}
}

func TestNumberFormatRejectsInvalidUnit(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{locale.MustParse("en")}, Options{Style: UnitStyle, Unit: UnitIdentifier("bad-unit")})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	}
}

func TestNumberFormatRejectsInvalidUnitWhenStyleIsNotUnit(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{locale.MustParse("en")}, Options{Unit: UnitIdentifier("bad-unit")})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	}
}

func TestNumberFormatAcceptsSanctionedUnits(t *testing.T) {
	t.Parallel()

	for _, unit := range []string{"meter", "kilometer", "meter-per-second", "microsecond", "nanosecond-per-second"} {
		format, err := New(locale.List{locale.MustParse("en")}, Options{Style: UnitStyle, Unit: UnitIdentifier(unit)})
		if err != nil {
			t.Fatalf("New(Options{Unit: UnitIdentifier(%q)}) error = %v, want nil", unit, err)
		}
		if got := format.ResolvedOptions().Unit; got != strings.ToLower(unit) {
			t.Fatalf("ResolvedOptions().Unit = %q, want %q", got, strings.ToLower(unit))
		}
	}
}

func TestNumberFormatResolvedOptionsExposeOnlyStyleSlots(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en")}, Options{
		Currency: CurrencyCode("USD"),
		Unit:     UnitIdentifier("meter"),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := format.ResolvedOptions()
	if got.Currency != "" || got.CurrencyDisplay != "" || got.CurrencySign != "" {
		t.Fatalf("decimal currency slots = %q/%q/%q, want empty", got.Currency, got.CurrencyDisplay, got.CurrencySign)
	}
	if got.Unit != "" || got.UnitDisplay != "" {
		t.Fatalf("decimal unit slots = %q/%q, want empty", got.Unit, got.UnitDisplay)
	}

	format, err = New(locale.List{locale.MustParse("en")}, Options{Style: CurrencyStyle, Currency: CurrencyCode("USD"), Unit: UnitIdentifier("meter")})
	if err != nil {
		t.Fatal(err)
	}
	got = format.ResolvedOptions()
	if got.Currency != "USD" || got.CurrencyDisplay != CurrencyDisplaySymbol || got.CurrencySign != StandardCurrencySign {
		t.Fatalf("currency slots = %q/%q/%q, want USD/symbol/standard", got.Currency, got.CurrencyDisplay, got.CurrencySign)
	}
	if got.Unit != "" || got.UnitDisplay != "" {
		t.Fatalf("currency unit slots = %q/%q, want empty", got.Unit, got.UnitDisplay)
	}

	format, err = New(locale.List{locale.MustParse("en")}, Options{Style: UnitStyle, Unit: UnitIdentifier("meter"), Currency: CurrencyCode("USD")})
	if err != nil {
		t.Fatal(err)
	}
	got = format.ResolvedOptions()
	if got.Currency != "" || got.CurrencyDisplay != "" || got.CurrencySign != "" {
		t.Fatalf("unit currency slots = %q/%q/%q, want empty", got.Currency, got.CurrencyDisplay, got.CurrencySign)
	}
	if got.Unit != "meter" || got.UnitDisplay != ShortUnitDisplay {
		t.Fatalf("unit slots = %q/%q, want meter/short", got.Unit, got.UnitDisplay)
	}
}

func TestNumberFormatRejectsCaseChangedUnit(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{locale.MustParse("en")}, Options{Style: UnitStyle, Unit: UnitIdentifier("METER")})
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
		{name: "style", opt: Options{Style: Style("bad")}},
		{name: "notation", opt: Options{Notation: Notation("bad")}},
		{name: "compact display", opt: Options{CompactDisplay: CompactDisplay("bad")}},
		{name: "currency display", opt: Options{CurrencyDisplay: CurrencyDisplay("bad")}},
		{name: "currency sign", opt: Options{CurrencySign: CurrencySign("bad")}},
		{name: "unit display", opt: Options{UnitDisplay: UnitDisplay("bad")}},
		{name: "sign display", opt: Options{SignDisplay: SignDisplay("bad")}},
		{name: "rounding mode", opt: Options{RoundingMode: RoundingMode("bad")}},
		{name: "rounding priority", opt: Options{RoundingPriority: RoundingPriority("bad")}},
		{name: "trailing zero display", opt: Options{TrailingZeroDisplay: TrailingZeroDisplay("bad")}},
		{name: "locale matcher", opt: Options{LocaleMatcher: LocaleMatcher("bad")}},
		{name: "use grouping", opt: Options{UseGrouping: UseGrouping("bad")}},
		{name: "numbering system", opt: Options{NumberingSystem: "ab"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{locale.MustParse("en")}, tc.opt)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
			}
		})
	}
}

func TestNumberFormatRejectsInvalidNumericOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opt  Options
	}{
		{name: "minimum integer digits zero", opt: Options{MinimumIntegerDigits: intPtr(0)}},
		{name: "minimum integer digits low", opt: Options{MinimumIntegerDigits: intPtr(-1)}},
		{name: "minimum fraction digits low", opt: Options{MinimumFractionDigits: intPtr(-1)}},
		{name: "maximum fraction digits high", opt: Options{MaximumFractionDigits: intPtr(101)}},
		{name: "minimum significant digits low", opt: Options{MinimumSignificantDigits: intPtr(0)}},
		{name: "maximum significant digits high", opt: Options{MaximumSignificantDigits: intPtr(22)}},
		{name: "rounding increment zero", opt: Options{RoundingIncrement: intPtr(0)}},
		{name: "rounding increment invalid", opt: Options{RoundingIncrement: intPtr(3)}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{locale.MustParse("en")}, tc.opt)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
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
		{name: "currency sign", opts: Options{Style: CurrencyStyle, Currency: CurrencyCode("USD"), CurrencySign: AccountingCurrencySign}},
		{name: "compact display", opts: Options{Notation: CompactNotation, CompactDisplay: LongCompactDisplay}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := New(locale.List{locale.MustParse("en")}, tc.opts); err != nil {
				t.Fatalf("New() error = %v, want nil", err)
			}
		})
	}
}

func TestNumberFormatResolvedOptionsSignificantDigits(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en")}, Options{MinimumSignificantDigits: intPtr(3)})
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

	format, err = New(locale.List{locale.MustParse("en")}, Options{MaximumSignificantDigits: intPtr(3)})
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

	format, err := New(locale.List{locale.MustParse("en")}, Options{MaximumFractionDigits: intPtr(2), RoundingPriority: MorePrecisionRoundingPriority})
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

	format, err := New(locale.List{locale.MustParse("en")}, Options{Notation: CompactNotation})
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

	format, err = New(locale.List{locale.MustParse("en")}, Options{Notation: CompactNotation, UseGrouping: UseGroupingAlways})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.ResolvedOptions().UseGrouping; got != UseGroupingAlways {
		t.Fatalf("explicit UseGrouping = %q, want always", got)
	}
}

func TestNumberFormatResolvedOptionsCurrencyDefaultsOnlyForStandardNotation(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en")}, Options{
		Style:    CurrencyStyle,
		Currency: CurrencyCode("USD"),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := format.ResolvedOptions()
	if got.MinimumFractionDigits == nil || got.MaximumFractionDigits == nil ||
		*got.MinimumFractionDigits != 2 || *got.MaximumFractionDigits != 2 {
		t.Fatalf("standard currency fraction digits = %v/%v, want 2/2", got.MinimumFractionDigits, got.MaximumFractionDigits)
	}

	format, err = New(locale.List{locale.MustParse("en")}, Options{
		Style:    CurrencyStyle,
		Currency: CurrencyCode("USD"),
		Notation: ScientificNotation,
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

	format, err := New(locale.List{locale.MustParse("en")}, Options{Notation: ScientificNotation})
	if err != nil {
		t.Fatal(err)
	}
	first := format.ResolvedOptions()
	first.Notation = "compact"
	if first.Notation != "compact" {
		t.Fatalf("mutated snapshot Notation = %q, want compact", first.Notation)
	}

	if got := format.ResolvedOptions().Notation; got != "scientific" {
		t.Fatalf("Notation after mutating snapshot = %q, want scientific", got)
	}
}

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()

	requested := locale.List{
		locale.MustParse("fr-FR"),
		locale.MustParse("en-US-u-nu-latn"),
		locale.MustParse("ban"),
	}
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: LookupLocaleMatcher})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	want := []string{"fr-FR", "en-US-u-nu-latn"}
	if len(got) != len(want) {
		t.Fatalf("SupportedLocalesOf() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i].String() != want[i] {
			t.Fatalf("SupportedLocalesOf()[%d] = %q, want %q", i, got[i].String(), want[i])
		}
	}
}

func TestSupportedLocalesOfRecognizesDerivedAvailableLocale(t *testing.T) {
	t.Parallel()

	requested := locale.MustParseList("zh-HK-u-nu-latn")
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: LookupLocaleMatcher})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	if len(got) != 1 || got[0].String() != "zh-HK-u-nu-latn" {
		t.Fatalf("SupportedLocalesOf() = %v, want [zh-HK-u-nu-latn]", got)
	}

	format, err := New(requested, Options{LocaleMatcher: LookupLocaleMatcher})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if got := format.ResolvedOptions().Locale.String(); got != "zh-HK-u-nu-latn" {
		t.Fatalf("ResolvedOptions().Locale = %q, want zh-HK-u-nu-latn", got)
	}
}

func TestSupportedLocalesOfErrors(t *testing.T) {
	t.Parallel()

	requested := locale.List{locale.MustParse("en-US")}
	if _, err := SupportedLocalesOf(requested, Options{LocaleMatcher: LocaleMatcher("bad")}); !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("SupportedLocalesOf(invalid matcher) error = %v, want intlerr.ErrInvalidOption", err)
	}
}
