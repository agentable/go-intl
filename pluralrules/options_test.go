package pluralrules

import (
	"errors"
	"slices"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
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
		{name: "rounding mode truncates toward one", opts: Options{MaximumFractionDigits: intPtr(0), RoundingMode: stringPtr(TruncRoundingMode)}, in: 1.9, want: One},
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
		name         string
		opts         Options
		wantName     string
		wantValue    string
		wantExpected string
	}{
		{name: "minimum integer digits zero", opts: Options{MinimumIntegerDigits: intPtr(0)}, wantName: "minimumIntegerDigits", wantValue: "0"},
		{name: "minimum integer digits high", opts: Options{MinimumIntegerDigits: intPtr(22)}, wantName: "minimumIntegerDigits", wantValue: "22"},
		{name: "minimum fraction digits low", opts: Options{MinimumFractionDigits: intPtr(-1)}, wantName: "minimumFractionDigits", wantValue: "-1"},
		{name: "maximum fraction digits high", opts: Options{MaximumFractionDigits: intPtr(101)}, wantName: "maximumFractionDigits", wantValue: "101"},
		{
			name:         "maximum less than minimum",
			opts:         Options{MinimumFractionDigits: intPtr(2), MaximumFractionDigits: intPtr(1)},
			wantName:     "maximumFractionDigits",
			wantValue:    "1",
			wantExpected: "greater than or equal to minimumFractionDigits",
		},
		{name: "bad notation", opts: Options{Notation: stringPtr(Notation("bad"))}, wantName: "notation", wantValue: "bad"},
		{name: "empty notation", opts: Options{Notation: stringPtr("")}, wantName: "notation", wantValue: ""},
		{name: "bad compact display", opts: Options{CompactDisplay: stringPtr(CompactDisplay("bad"))}, wantName: "compactDisplay", wantValue: "bad"},
		{name: "empty compact display", opts: Options{CompactDisplay: stringPtr("")}, wantName: "compactDisplay", wantValue: ""},
		{name: "bad type", opts: Options{Type: stringPtr(Type("bad"))}, wantName: "type", wantValue: "bad"},
		{name: "empty type", opts: Options{Type: stringPtr("")}, wantName: "type", wantValue: ""},
		{name: "bad rounding mode", opts: Options{RoundingMode: stringPtr(RoundingMode("bankers"))}, wantName: "roundingMode", wantValue: "bankers"},
		{name: "empty rounding mode", opts: Options{RoundingMode: stringPtr("")}, wantName: "roundingMode", wantValue: ""},
		{name: "bad rounding priority", opts: Options{RoundingPriority: stringPtr(RoundingPriority("bad"))}, wantName: "roundingPriority", wantValue: "bad"},
		{name: "empty rounding priority", opts: Options{RoundingPriority: stringPtr("")}, wantName: "roundingPriority", wantValue: ""},
		{name: "bad trailing zero display", opts: Options{TrailingZeroDisplay: stringPtr(TrailingZeroDisplay("bad"))}, wantName: "trailingZeroDisplay", wantValue: "bad"},
		{name: "empty trailing zero display", opts: Options{TrailingZeroDisplay: stringPtr("")}, wantName: "trailingZeroDisplay", wantValue: ""},
		{name: "rounding increment zero", opts: Options{RoundingIncrement: intPtr(0)}, wantName: "roundingIncrement", wantValue: "0"},
		{name: "bad rounding increment", opts: Options{RoundingIncrement: intPtr(3)}, wantName: "roundingIncrement", wantValue: "3"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{intltest.Locale(t, "en")}, tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
			}
			testcontract.AssertOptionError(t, err, "pluralrules", intlerr.InvalidOption, tc.wantName, tc.wantValue, "en")
			if tc.wantExpected != "" {
				testcontract.AssertOptionExpected(t, err, tc.wantExpected)
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
	if got.CompactDisplay != nil {
		t.Fatalf("ResolvedOptions().CompactDisplay = %v, want nil for standard notation", got.CompactDisplay)
	}
	if !slices.Equal(got.PluralCategories, []Category{One, Other}) {
		t.Fatalf("PluralCategories = %#v", got.PluralCategories)
	}
}

func TestPluralRuleTypeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		typ  Type
		want string
	}{
		{name: "zero defaults to cardinal", want: "cardinal"},
		{name: "cardinal", typ: Cardinal, want: "cardinal"},
		{name: "ordinal", typ: Ordinal, want: "ordinal"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.typ.String(); got != tc.want {
				t.Fatalf("%q.String() = %q, want %q", tc.typ, got, tc.want)
			}
		})
	}
}

func TestPluralRulesResolvedOptionsPluralCategoriesAreCallerOwned(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	first := rules.ResolvedOptions()
	if !slices.Equal(first.PluralCategories, []Category{One, Other}) {
		t.Fatalf("PluralCategories = %#v, want [one other]", first.PluralCategories)
	}

	first.PluralCategories[0] = Zero

	if got := rules.ResolvedOptions().PluralCategories; !slices.Equal(got, []Category{One, Other}) {
		t.Fatalf("PluralCategories after caller mutation = %#v, want [one other]", got)
	}
}

func TestPluralRulesResolvedOptionsOmitsCompactDisplayForStandardNotation(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{CompactDisplay: stringPtr(LongCompactDisplay)})
	if err != nil {
		t.Fatal(err)
	}
	if got := rules.ResolvedOptions().CompactDisplay; got != nil {
		t.Fatalf("ResolvedOptions().CompactDisplay = %v, want nil when notation is standard", got)
	}
}

func TestPluralRulesResolvedOptionsSignificantDigits(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Notation:                 stringPtr(CompactNotation),
		CompactDisplay:           stringPtr(LongCompactDisplay),
		MinimumSignificantDigits: intPtr(2), MaximumSignificantDigits: intPtr(4),
		RoundingMode:        stringPtr(HalfEvenRoundingMode),
		RoundingPriority:    stringPtr(MorePrecisionRoundingPriority),
		TrailingZeroDisplay: stringPtr(StripIfIntegerTrailingZeroDisplay),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := rules.ResolvedOptions()
	if got.Notation != CompactNotation || got.CompactDisplay == nil || *got.CompactDisplay != LongCompactDisplay ||
		got.MinimumSignificantDigits == nil || got.MaximumSignificantDigits == nil ||
		*got.MinimumSignificantDigits != 2 || *got.MaximumSignificantDigits != 4 {
		t.Fatalf("ResolvedOptions() = %#v", got)
	}
	*got.CompactDisplay = ShortCompactDisplay
	*got.MinimumSignificantDigits = 9
	*got.MaximumSignificantDigits = 9
	if second := rules.ResolvedOptions(); second.CompactDisplay == nil || *second.CompactDisplay != LongCompactDisplay {
		t.Fatalf("ResolvedOptions().CompactDisplay after caller mutation = %v, want %q", second.CompactDisplay, LongCompactDisplay)
	} else if second.MinimumSignificantDigits == nil || second.MaximumSignificantDigits == nil ||
		*second.MinimumSignificantDigits != 2 || *second.MaximumSignificantDigits != 4 {
		t.Fatalf("ResolvedOptions() significant digits after caller mutation = %v/%v, want 2/4", second.MinimumSignificantDigits, second.MaximumSignificantDigits)
	}
	if got.MinimumFractionDigits != nil || got.MaximumFractionDigits != nil {
		t.Fatalf("ResolvedOptions() fraction digits = %v/%v, want nil/nil for compact significant-digit surface", got.MinimumFractionDigits, got.MaximumFractionDigits)
	}
	if got.RoundingMode != HalfEvenRoundingMode || got.RoundingPriority != MorePrecisionRoundingPriority || got.TrailingZeroDisplay != StripIfIntegerTrailingZeroDisplay {
		t.Fatalf("ResolvedOptions() rounding options = %#v", got)
	}
}

func TestPluralRulesResolvedOptionsRoundingPriorityUsesSignificantDigitSurface(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		MaximumFractionDigits: intPtr(2),
		RoundingPriority:      stringPtr(MorePrecisionRoundingPriority),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := rules.ResolvedOptions()
	if got.MinimumFractionDigits != nil || got.MaximumFractionDigits != nil {
		t.Fatalf("ResolvedOptions() fraction digits = %v/%v, want nil/nil", got.MinimumFractionDigits, got.MaximumFractionDigits)
	}
	if got.MinimumSignificantDigits == nil || got.MaximumSignificantDigits == nil ||
		*got.MinimumSignificantDigits != 1 || *got.MaximumSignificantDigits != 21 {
		t.Fatalf("ResolvedOptions() significant digits = %v/%v, want 1/21", got.MinimumSignificantDigits, got.MaximumSignificantDigits)
	}
	if got.RoundingPriority != MorePrecisionRoundingPriority {
		t.Fatalf("ResolvedOptions().RoundingPriority = %q, want %q", got.RoundingPriority, MorePrecisionRoundingPriority)
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

func TestPluralRulesKeepsDataLocaleSeparateFromResolvedLocale(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "zh-HK")}, Options{LocaleMatcher: stringPtr(LookupLocaleMatcher)})
	if err != nil {
		t.Fatal(err)
	}
	if got := rules.ResolvedOptions().Locale.String(); got != "zh-HK" {
		t.Fatalf("ResolvedOptions().Locale = %q, want zh-HK", got)
	}
	if rules.dataLocale != "zh-Hant-HK" {
		t.Fatalf("PluralRules dataLocale = %q, want zh-Hant-HK", rules.dataLocale)
	}
}

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()

	requested := locale.List{intltest.Locale(t, "de-DE"), intltest.Locale(t, "fr-FR"), intltest.Locale(t, "en-US")}
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: stringPtr(LookupLocaleMatcher)})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	testcontract.AssertLocaleListStrings(t, "SupportedLocalesOf()", got, []string{"de-DE", "fr-FR", "en-US"})
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
			testcontract.AssertOptionError(t, err, "pluralrules", intlerr.InvalidOption, "localeMatcher", matcher, "en-US")
			testcontract.AssertOptionExpected(t, err, `one of "lookup", "best fit"`)
		})
	}
}
