package pluralrules

import (
	"errors"
	"slices"
	"testing"

	"github.com/agentable/go-intl/locale"
)

func TestPluralRulesMinimumFractionDigitsAffectsSelection(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.MustParse("en"), Options{FractionDigits: MinimumFractionDigits(2)})
	if err != nil {
		t.Fatal(err)
	}
	if got := rules.SelectInt(1); got != Other {
		t.Fatalf("SelectInt(1) = %s, want other", got)
	}
}

func TestPluralRulesDigitOptionsAffectOperands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts Options
		in   float64
		want Category
	}{
		{name: "maximum fraction rounds to one", opts: Options{FractionDigits: MaximumFractionDigits(0)}, in: 1.4, want: One},
		{name: "maximum fraction rounds away from one", opts: Options{FractionDigits: MaximumFractionDigits(0)}, in: 1.5, want: Other},
		{name: "minimum integer digits keeps numeric value", opts: Options{MinimumIntegerDigits: 2}, in: 1, want: One},
		{name: "significant digits share numberformat rounding", opts: Options{SignificantDigits: SignificantDigits(1, 1)}, in: 1.5, want: Other},
		{name: "rounding mode truncates toward one", opts: Options{FractionDigits: MaximumFractionDigits(0), RoundingMode: TruncRoundingMode}, in: 1.9, want: One},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rules, err := New(locale.MustParse("en"), tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			got, err := rules.SelectFloat64(tc.in)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("SelectFloat64(%v) = %s, want %s", tc.in, got, tc.want)
			}
		})
	}
}

func TestPluralRulesInvalidDigitOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts Options
	}{
		{name: "minimum integer digits high", opts: Options{MinimumIntegerDigits: 22}},
		{name: "minimum fraction digits low", opts: Options{FractionDigits: MinimumFractionDigits(-1)}},
		{name: "maximum fraction digits high", opts: Options{FractionDigits: MaximumFractionDigits(101)}},
		{name: "maximum less than minimum", opts: Options{FractionDigits: FractionDigits(2, 1)}},
		{name: "bad notation", opts: Options{Notation: Notation("bad")}},
		{name: "bad rounding mode", opts: Options{RoundingMode: RoundingMode("bankers")}},
		{name: "bad rounding increment", opts: Options{RoundingIncrement: 3}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.MustParse("en"), tc.opts)
			if !errors.Is(err, ErrInvalidOption) {
				t.Fatalf("New() error = %v, want ErrInvalidOption", err)
			}
		})
	}
}

func TestPluralRulesResolvedOptions(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.MustParse("en"))
	if err != nil {
		t.Fatal(err)
	}
	got := rules.ResolvedOptions()
	if got.Locale.String() != "en" || got.Type != Cardinal || got.MinimumFractionDigits != 0 || got.MaximumFractionDigits != 3 || got.MinimumIntegerDigits != 1 {
		t.Fatalf("ResolvedOptions() = %#v", got)
	}
	if got.Notation != StandardNotation || got.RoundingIncrement != 1 || got.RoundingMode != HalfExpandRoundingMode || got.RoundingPriority != AutoRoundingPriority || got.TrailingZeroDisplay != AutoTrailingZeroDisplay {
		t.Fatalf("ResolvedOptions() rounding surface = %#v", got)
	}
	if !slices.Equal(got.PluralCategories, []Category{One, Other}) {
		t.Fatalf("PluralCategories = %#v", got.PluralCategories)
	}
}

func TestPluralRulesResolvedOptionsSignificantDigits(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.MustParse("en"), Options{
		Notation:            CompactNotation,
		CompactDisplay:      LongCompactDisplay,
		SignificantDigits:   SignificantDigits(2, 4),
		RoundingMode:        HalfEvenRoundingMode,
		RoundingPriority:    MorePrecisionRoundingPriority,
		TrailingZeroDisplay: StripIfIntegerTrailingZeroDisplay,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := rules.ResolvedOptions()
	if got.Notation != CompactNotation || got.CompactDisplay != LongCompactDisplay || got.MinimumSignificantDigits != 2 || got.MaximumSignificantDigits != 4 {
		t.Fatalf("ResolvedOptions() = %#v", got)
	}
	if got.MinimumFractionDigits != 0 || got.MaximumFractionDigits != 3 {
		t.Fatalf("ResolvedOptions() fraction digits = %d/%d, want default fraction digit surface", got.MinimumFractionDigits, got.MaximumFractionDigits)
	}
	if got.RoundingMode != HalfEvenRoundingMode || got.RoundingPriority != MorePrecisionRoundingPriority || got.TrailingZeroDisplay != StripIfIntegerTrailingZeroDisplay {
		t.Fatalf("ResolvedOptions() rounding options = %#v", got)
	}
}

func TestPluralRulesResolvedOptionsArabicCategories(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.MustParse("ar"))
	if err != nil {
		t.Fatal(err)
	}
	want := []Category{Zero, One, Two, Few, Many, Other}
	if got := rules.ResolvedOptions().PluralCategories; !slices.Equal(got, want) {
		t.Fatalf("PluralCategories = %#v, want %#v", got, want)
	}
}

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()

	requested := []locale.Locale{
		locale.MustParse("de-DE"),
		locale.MustParse("fr-FR"),
		locale.MustParse("en-US"),
	}
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: LookupLocaleMatcher})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	want := []string{"fr-FR", "en-US"}
	if len(got) != len(want) {
		t.Fatalf("SupportedLocalesOf() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i].String() != want[i] {
			t.Fatalf("SupportedLocalesOf()[%d] = %q, want %q", i, got[i].String(), want[i])
		}
	}
}
