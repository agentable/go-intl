package pluralrules

import (
	"errors"
	"slices"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestPluralRulesMinimumFractionDigitsAffectsSelection(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{MinimumFractionDigits: intPtr(2)})
	if err != nil {
		t.Fatal(err)
	}
	if got := mustCategory(rules.Select(Int(int64(1)))); got != Other {
		t.Fatalf("SelectInt(1) = %s, want other", got)
	}
}

func TestPluralRulesMinimumSignificantDigitsDefaultsMaximum(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{MinimumSignificantDigits: intPtr(2)})
	if err != nil {
		t.Fatal(err)
	}
	if got := mustCategory(rules.Select(Int(int64(1)))); got != Other {
		t.Fatalf("SelectInt(1) = %s, want %s when minimumSignificantDigits formats 1 as 1.0", got, Other)
	}

	resolved := rules.ResolvedOptions()
	if resolved.MinimumSignificantDigits == nil || resolved.MaximumSignificantDigits == nil ||
		*resolved.MinimumSignificantDigits != 2 || *resolved.MaximumSignificantDigits != 21 {
		t.Fatalf("ResolvedOptions() significant digits = %v/%v, want 2/21", resolved.MinimumSignificantDigits, resolved.MaximumSignificantDigits)
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
		{name: "maximum fraction rounds to one", opts: Options{MaximumFractionDigits: intPtr(0)}, in: 1.4, want: One},
		{name: "maximum fraction rounds away from one", opts: Options{MaximumFractionDigits: intPtr(0)}, in: 1.5, want: Other},
		{name: "minimum integer digits keeps numeric value", opts: Options{MinimumIntegerDigits: intPtr(2)}, in: 1, want: One},
		{name: "significant digits share numberformat rounding", opts: Options{MinimumSignificantDigits: intPtr(1), MaximumSignificantDigits: intPtr(1)}, in: 1.5, want: Other},
		{name: "rounding mode truncates toward one", opts: Options{MaximumFractionDigits: intPtr(0), RoundingMode: TruncRoundingMode}, in: 1.9, want: One},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rules, err := New(locale.List{intltest.Locale(t, "en")}, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			got, err := rules.Select(Float(tc.in))
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
		{name: "minimum integer digits zero", opts: Options{MinimumIntegerDigits: intPtr(0)}},
		{name: "minimum integer digits high", opts: Options{MinimumIntegerDigits: intPtr(22)}},
		{name: "minimum fraction digits low", opts: Options{MinimumFractionDigits: intPtr(-1)}},
		{name: "maximum fraction digits high", opts: Options{MaximumFractionDigits: intPtr(101)}},
		{name: "maximum less than minimum", opts: Options{MinimumFractionDigits: intPtr(2), MaximumFractionDigits: intPtr(1)}},
		{name: "bad notation", opts: Options{Notation: Notation("bad")}},
		{name: "bad rounding mode", opts: Options{RoundingMode: RoundingMode("bankers")}},
		{name: "rounding increment zero", opts: Options{RoundingIncrement: intPtr(0)}},
		{name: "bad rounding increment", opts: Options{RoundingIncrement: intPtr(3)}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{intltest.Locale(t, "en")}, tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
			}
		})
	}
}

func TestPluralRulesResolvedOptions(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	got := rules.ResolvedOptions()
	if got.MinimumFractionDigits == nil || got.MaximumFractionDigits == nil ||
		got.Locale.String() != "en" || got.Type != Cardinal ||
		*got.MinimumFractionDigits != 0 || *got.MaximumFractionDigits != 3 || got.MinimumIntegerDigits != 1 {
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

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Notation:                 CompactNotation,
		CompactDisplay:           LongCompactDisplay,
		MinimumSignificantDigits: intPtr(2), MaximumSignificantDigits: intPtr(4),
		RoundingMode:        HalfEvenRoundingMode,
		RoundingPriority:    MorePrecisionRoundingPriority,
		TrailingZeroDisplay: StripIfIntegerTrailingZeroDisplay,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := rules.ResolvedOptions()
	if got.Notation != CompactNotation || got.CompactDisplay != LongCompactDisplay ||
		got.MinimumSignificantDigits == nil || got.MaximumSignificantDigits == nil ||
		*got.MinimumSignificantDigits != 2 || *got.MaximumSignificantDigits != 4 {
		t.Fatalf("ResolvedOptions() = %#v", got)
	}
	if got.MinimumFractionDigits == nil || got.MaximumFractionDigits == nil ||
		*got.MinimumFractionDigits != 0 || *got.MaximumFractionDigits != 3 {
		t.Fatalf("ResolvedOptions() fraction digits = %v/%v, want default fraction digit surface", got.MinimumFractionDigits, got.MaximumFractionDigits)
	}
	if got.RoundingMode != HalfEvenRoundingMode || got.RoundingPriority != MorePrecisionRoundingPriority || got.TrailingZeroDisplay != StripIfIntegerTrailingZeroDisplay {
		t.Fatalf("ResolvedOptions() rounding options = %#v", got)
	}
}

func TestPluralRulesResolvedOptionsArabicCategories(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "ar")}, Options{})
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

	requested := locale.List{intltest.Locale(t, "de-DE"), intltest.Locale(t, "fr-FR"), intltest.Locale(t, "en-US")}
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: LookupLocaleMatcher})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	want := []string{"de-DE", "fr-FR", "en-US"}
	if len(got) != len(want) {
		t.Fatalf("SupportedLocalesOf() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i].String() != want[i] {
			t.Fatalf("SupportedLocalesOf()[%d] = %q, want %q", i, got[i].String(), want[i])
		}
	}
}
