package numberformat

import (
	"errors"
	"strings"
	"testing"

	"github.com/agentable/go-intl/locale"
)

func TestNumberFormatResolvedOptionsDefaults(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"))
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
	if got.MinimumFractionDigits != 0 || got.MaximumFractionDigits != 3 {
		t.Fatalf("fraction digits = %d/%d, want 0/3", got.MinimumFractionDigits, got.MaximumFractionDigits)
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

	format, err := New(locale.MustParse("en-u-nu-latn"))
	if err != nil {
		t.Fatal(err)
	}
	if got := format.ResolvedOptions().NumberingSystem; got != "latn" {
		t.Fatalf("NumberingSystem = %q, want latn", got)
	}

	format, err = New(locale.MustParse("en-u-nu-arab"), Options{NumberingSystem: "latn"})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.ResolvedOptions().NumberingSystem; got != "latn" {
		t.Fatalf("NumberingSystem = %q, want explicit option latn", got)
	}
}

func TestNumberFormatFallsBackToNumberDataLocale(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("de-DE"))
	if err != nil {
		t.Fatal(err)
	}
	if got := format.ResolvedOptions().NumberingSystem; got != "latn" {
		t.Fatalf("NumberingSystem = %q, want safe number data fallback", got)
	}
	if got := format.formatValue(1234); got != "1,234" {
		t.Fatalf("Format(1234) = %q, want fallback output %q", got, "1,234")
	}
}

func TestNumberFormatCurrencyRequiresCurrency(t *testing.T) {
	t.Parallel()

	_, err := New(locale.MustParse("en-US"), Options{Style: CurrencyStyle})
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("New() error = %v, want ErrInvalidOption", err)
	}
	if !strings.Contains(err.Error(), "currency") || !strings.Contains(err.Error(), "en-US") {
		t.Fatalf("New() error = %q, want currency and locale context", err)
	}
}

func TestNumberFormatRejectsInvalidCurrencyCode(t *testing.T) {
	t.Parallel()

	_, err := New(locale.MustParse("en-US"), Options{Style: CurrencyStyle, Currency: CurrencyCode("US")})
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("New() error = %v, want ErrInvalidOption", err)
	}
}

func TestNumberFormatRejectsInvalidUnit(t *testing.T) {
	t.Parallel()

	_, err := New(locale.MustParse("en"), Options{Style: UnitStyle, Unit: UnitIdentifier("bad-unit")})
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("New() error = %v, want ErrInvalidOption", err)
	}
}

func TestNumberFormatAcceptsSanctionedUnits(t *testing.T) {
	t.Parallel()

	for _, unit := range []string{"meter", "kilometer", "meter-per-second", "microsecond", "nanosecond-per-second"} {
		format, err := New(locale.MustParse("en"), Options{Style: UnitStyle, Unit: UnitIdentifier(unit)})
		if err != nil {
			t.Fatalf("New(Options{Unit: UnitIdentifier(%q)}) error = %v, want nil", unit, err)
		}
		if got := format.ResolvedOptions().Unit; got != strings.ToLower(unit) {
			t.Fatalf("ResolvedOptions().Unit = %q, want %q", got, strings.ToLower(unit))
		}
	}
}

func TestNumberFormatRejectsCaseChangedUnit(t *testing.T) {
	t.Parallel()

	_, err := New(locale.MustParse("en"), Options{Style: UnitStyle, Unit: UnitIdentifier("METER")})
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("New() error = %v, want ErrInvalidOption", err)
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

			_, err := New(locale.MustParse("en"), tc.opt)
			if !errors.Is(err, ErrInvalidOption) {
				t.Fatalf("New() error = %v, want ErrInvalidOption", err)
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
		{name: "minimum integer digits low", opt: Options{MinimumIntegerDigits: -1}},
		{name: "minimum fraction digits low", opt: Options{FractionDigits: MinimumFractionDigits(-1)}},
		{name: "maximum fraction digits high", opt: Options{FractionDigits: MaximumFractionDigits(101)}},
		{name: "minimum significant digits low", opt: Options{SignificantDigits: MinimumSignificantDigits(0)}},
		{name: "maximum significant digits high", opt: Options{SignificantDigits: MaximumSignificantDigits(22)}},
		{name: "rounding increment", opt: Options{RoundingIncrement: 3}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.MustParse("en"), tc.opt)
			if !errors.Is(err, ErrInvalidOption) {
				t.Fatalf("New() error = %v, want ErrInvalidOption", err)
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

			if _, err := New(locale.MustParse("en"), tc.opts); err != nil {
				t.Fatalf("New() error = %v, want nil", err)
			}
		})
	}
}

func TestNumberFormatResolvedOptionsSignificantDigits(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{SignificantDigits: MinimumSignificantDigits(3)})
	if err != nil {
		t.Fatal(err)
	}
	got := format.ResolvedOptions()
	if got.MinimumSignificantDigits != 3 || got.MaximumSignificantDigits != 21 {
		t.Fatalf("significant digits = %d/%d, want 3/21", got.MinimumSignificantDigits, got.MaximumSignificantDigits)
	}
	if got.MinimumFractionDigits != 0 || got.MaximumFractionDigits != 0 {
		t.Fatalf("fraction digits = %d/%d, want hidden 0/0", got.MinimumFractionDigits, got.MaximumFractionDigits)
	}

	format, err = New(locale.MustParse("en"), Options{SignificantDigits: MaximumSignificantDigits(3)})
	if err != nil {
		t.Fatal(err)
	}
	got = format.ResolvedOptions()
	if got.MinimumSignificantDigits != 1 || got.MaximumSignificantDigits != 3 {
		t.Fatalf("significant digits = %d/%d, want 1/3", got.MinimumSignificantDigits, got.MaximumSignificantDigits)
	}
}

func TestNumberFormatResolvedOptionsRoundingPriority(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{FractionDigits: MaximumFractionDigits(2), RoundingPriority: MorePrecisionRoundingPriority})
	if err != nil {
		t.Fatal(err)
	}
	got := format.ResolvedOptions()
	if got.RoundingPriority != "morePrecision" {
		t.Fatalf("RoundingPriority = %q, want morePrecision", got.RoundingPriority)
	}
	if got.MinimumSignificantDigits != 1 || got.MaximumSignificantDigits != 21 {
		t.Fatalf("significant defaults = %d/%d, want 1/21", got.MinimumSignificantDigits, got.MaximumSignificantDigits)
	}
	if got.MaximumFractionDigits != 2 {
		t.Fatalf("MaximumFractionDigits = %d, want 2", got.MaximumFractionDigits)
	}
}

func TestNumberFormatResolvedOptionsIsSnapshot(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{Notation: ScientificNotation})
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

func TestNumberFormatRejectsMultipleOptions(t *testing.T) {
	t.Parallel()

	_, err := New(locale.MustParse("en-US"), Options{Style: DecimalStyle}, Options{UseGrouping: UseGroupingFalse})
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("New() error = %v, want ErrInvalidOption", err)
	}
}

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()

	requested := []locale.Locale{
		locale.MustParse("fr-FR"),
		locale.MustParse("en-US-u-nu-latn"),
		locale.MustParse("ban"),
	}
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: LookupLocaleMatcher})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	want := []string{"en-US-u-nu-latn"}
	if len(got) != len(want) {
		t.Fatalf("SupportedLocalesOf() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i].String() != want[i] {
			t.Fatalf("SupportedLocalesOf()[%d] = %q, want %q", i, got[i].String(), want[i])
		}
	}
}
