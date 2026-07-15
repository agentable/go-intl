package pluralrules

import (
	"errors"
	"math"
	"strings"
	"sync"
	"testing"

	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
	"github.com/agentable/go-intl/internal/intlerr"
	pluralop "github.com/agentable/go-intl/internal/plural"

	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
)

func TestCategoryMatchesInternalOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		public   Category
		internal pluralop.Category
	}{
		{Zero, pluralop.Zero},
		{One, pluralop.One},
		{Two, pluralop.Two},
		{Few, pluralop.Few},
		{Many, pluralop.Many},
		{Other, pluralop.Other},
	}
	for _, tc := range tests {
		t.Run(tc.public.String(), func(t *testing.T) {
			t.Parallel()

			if Category(tc.internal) != tc.public {
				t.Fatalf("Category(%s) = %s, want %s", tc.internal, Category(tc.internal), tc.public)
			}
			if pluralop.Category(tc.public) != tc.internal {
				t.Fatalf("internal Category(%s) = %s, want %s", tc.public, pluralop.Category(tc.public), tc.internal)
			}
		})
	}
}

func TestPluralRulesSelectCardinal(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		in   int
		want Category
	}{
		{0, Other},
		{1, One},
		{2, Other},
		{-1, One},
		{-2, Other},
	}
	for _, tc := range tests {
		if got := mustCategory(rules.Select(Int(int64(tc.in)))); got != tc.want {
			t.Fatalf("SelectInt(%v) = %s, want %s", tc.in, got, tc.want)
		}
	}
}

func TestPluralRulesSelectOrdinal(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{Type: stringPtr(Ordinal)})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		in   int
		want Category
	}{
		{1, One},
		{2, Two},
		{3, Few},
		{4, Other},
		{11, Other},
		{21, One},
	}
	for _, tc := range tests {
		if got := mustCategory(rules.Select(Int(int64(tc.in)))); got != tc.want {
			t.Fatalf("SelectInt(%v) = %s, want %s", tc.in, got, tc.want)
		}
	}
}

func TestPluralRulesSelectUsesExactLargeOperands(t *testing.T) {
	t.Parallel()

	ordinal, err := New(locale.List{intltest.Locale(t, "en")}, Options{Type: stringPtr(Ordinal)})
	if err != nil {
		t.Fatal(err)
	}
	ordinalTests := []struct {
		in   string
		want Category
	}{
		{in: "10000000000001", want: One},
		{in: "10000000000002", want: Two},
		{in: "10000000000003", want: Few},
		{in: "10000000000011", want: Other},
	}
	for _, tc := range ordinalTests {
		got, err := ordinal.Select(decimalValue(tc.in))
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("ordinal SelectDecimal(%s) = %s, want %s", tc.in, got, tc.want)
		}
	}

	cardinal, err := New(locale.List{intltest.Locale(t, "fr")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := cardinal.Select(decimalValue("1000000000000"))
	if err != nil {
		t.Fatal(err)
	}
	if got != Many {
		t.Fatalf("fr SelectDecimal(1000000000000) = %s, want %s", got, Many)
	}
	got, err = cardinal.Select(decimalValue("1000000000001"))
	if err != nil {
		t.Fatal(err)
	}
	if got != Other {
		t.Fatalf("fr SelectDecimal(1000000000001) = %s, want %s", got, Other)
	}
}

func TestPluralRulesDecimalInputUsesMathematicalValue(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		input string
		want  Category
	}{
		{input: "1", want: One},
		{input: "1.0", want: One},
		{input: "1.00", want: One},
		{input: "1.50", want: Other},
		{input: "-0.00", want: Other},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got, err := rules.Select(decimalValue(tc.input))
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("Select(Decimal(%q)) = %s, want %s", tc.input, got, tc.want)
			}
		})
	}
}

func TestPluralRulesRoundingPreservesArbitraryPrecisionDecimal(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "ru")}, Options{
		MaximumFractionDigits: intPtr(0),
		RoundingMode:          stringPtr(TruncRoundingMode),
	})
	if err != nil {
		t.Fatal(err)
	}
	inputText := "1" + strings.Repeat("0", 98) + "1.5"
	input, err := Decimal(inputText)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := rules.Select(input); err != nil || got != One {
		t.Fatalf("Select(Decimal(%q)) = %s, %v; want %s, nil", inputText, got, err, One)
	}
}

func TestPluralRulesNotationAffectsOperands(t *testing.T) {
	t.Parallel()

	standard, err := New(locale.List{intltest.Locale(t, "ca")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := standard.Select(decimalValue("12345678"))
	if err != nil {
		t.Fatal(err)
	}
	if got != Other {
		t.Fatalf("standard SelectDecimal(12345678) = %s, want %s", got, Other)
	}
	if got := mustCategory(standard.Select(Int(12345678))); got != Other {
		t.Fatalf("standard SelectInt64(12345678) = %s, want %s", got, Other)
	}

	for _, tc := range []struct {
		name string
		opts Options
	}{
		{name: "scientific", opts: Options{Notation: stringPtr(ScientificNotation)}},
		{name: "engineering", opts: Options{Notation: stringPtr(EngineeringNotation)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rules, err := New(locale.List{intltest.Locale(t, "ca")}, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			got, err := rules.Select(decimalValue("12345678"))
			if err != nil {
				t.Fatal(err)
			}
			if got != Many {
				t.Fatalf("%s SelectDecimal(12345678) = %s, want %s", tc.name, got, Many)
			}
			if got := mustCategory(rules.Select(Int(12345678))); got != Many {
				t.Fatalf("%s SelectInt64(12345678) = %s, want %s", tc.name, got, Many)
			}
		})
	}
}

func TestResolveDecimalUsesExplicitConstructorState(t *testing.T) {
	t.Parallel()

	d, err := decimal.ParseString("12345678")
	if err != nil {
		t.Fatal(err)
	}
	digitOptions := ecma402nf.DigitOptions{
		MinimumIntegerDigits:  1,
		MaximumFractionDigits: 3,
		RoundingIncrement:     1,
		RoundingMode:          "halfExpand",
		RoundingPriority:      "auto",
		TrailingZeroDisplay:   "auto",
	}
	tests := []struct {
		name          string
		notation      Notation
		wantFormatted string
		wantExponent  int
	}{
		{name: "standard", notation: StandardNotation, wantFormatted: "12345678"},
		{name: "scientific", notation: ScientificNotation, wantFormatted: "1.235", wantExponent: 7},
		{name: "engineering", notation: EngineeringNotation, wantFormatted: "12.346", wantExponent: 6},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotOps pluralop.OperandsRecord
			formatted, rounded, category := resolveDecimal(d, tc.notation, digitOptions, compactExponentSet{}, func(ops pluralop.OperandsRecord) pluralop.Category {
				gotOps = ops
				return pluralop.Many
			})
			if formatted != tc.wantFormatted {
				t.Fatalf("resolveDecimal formatted = %q, want %q", formatted, tc.wantFormatted)
			}
			if rounded.String() != tc.wantFormatted {
				t.Fatalf("resolveDecimal rounded = %s, want %s", rounded.String(), tc.wantFormatted)
			}
			if category != Many {
				t.Fatalf("resolveDecimal category = %s, want %s", category, Many)
			}
			if gotOps.C != tc.wantExponent || gotOps.E != tc.wantExponent {
				t.Fatalf("resolveDecimal operands exponent = c:%d e:%d, want %d", gotOps.C, gotOps.E, tc.wantExponent)
			}
		})
	}
}

func TestPluralRulesUnsignedTypedBridges(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}

	if got := mustCategory(rules.Select(Uint(uint64(1)))); got != One {
		t.Fatalf("SelectUint(1) = %s, want %s", got, One)
	}
	if got := mustCategory(rules.Select(Uint(2))); got != Other {
		t.Fatalf("SelectUint64(2) = %s, want %s", got, Other)
	}
	ordinal, err := New(locale.List{intltest.Locale(t, "en")}, Options{Type: stringPtr(Ordinal)})
	if err != nil {
		t.Fatal(err)
	}
	if got := mustCategory(ordinal.Select(Uint(10000000000001))); got != One {
		t.Fatalf("ordinal SelectUint64(10000000000001) = %s, want %s", got, One)
	}
	if got := mustCategory(rules.SelectRange(Uint(uint64(1)), Uint(uint64(2)))); got != Other {
		t.Fatalf("SelectRangeUint(1, 2) = %s, want %s", got, Other)
	}
	if got := mustCategory(rules.SelectRange(Uint(1), Uint(1))); got != One {
		t.Fatalf("SelectRangeUint64(1, 1) = %s, want %s", got, One)
	}
}

func TestPluralRulesSelectNonFiniteReturnsOther(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		value float64
	}{
		{name: "NaN", value: math.NaN()},
		{name: "positive_infinity", value: math.Inf(1)},
		{name: "negative_infinity", value: math.Inf(-1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := rules.Select(Float(tc.value))
			if err != nil || got != Other {
				t.Fatalf("Select(Float(%v)) = %s, %v; want %s, nil", tc.value, got, err, Other)
			}
		})
	}
}

func TestPluralRulesDecimalAcceptsNonFinite(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"NaN", "Infinity", "-Infinity"} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			input, err := Decimal(value)
			if err != nil {
				t.Fatalf("Decimal(%q) error = %v", value, err)
			}
			got, err := rules.Select(input)
			if err != nil || got != Other {
				t.Fatalf("Select(Decimal(%q)) = %s, %v; want %s, nil", value, got, err, Other)
			}
		})
	}
}

func TestPluralRulesSelectRangeAcceptsInfinity(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	positiveInfinity, err := Decimal("Infinity")
	if err != nil {
		t.Fatal(err)
	}
	negativeInfinity, err := Decimal("-Infinity")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name       string
		start, end Value
	}{
		{name: "positive_start", start: Float(math.Inf(1)), end: Int(1)},
		{name: "positive_end", start: Int(1), end: Float(math.Inf(1))},
		{name: "negative_start", start: Float(math.Inf(-1)), end: Int(1)},
		{name: "negative_end", start: Int(1), end: Float(math.Inf(-1))},
		{name: "equal_positive", start: Float(math.Inf(1)), end: Float(math.Inf(1))},
		{name: "decimal_tokens", start: positiveInfinity, end: negativeInfinity},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := rules.SelectRange(tc.start, tc.end)
			if err != nil || got != Other {
				t.Fatalf("SelectRange() = %s, %v; want %s, nil", got, err, Other)
			}
		})
	}
}

func TestPluralRulesSelectRangeRejectsNaN(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	decimalNaN, err := Decimal("NaN")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name       string
		start, end Value
		field      string
	}{
		{name: "float_start", start: Float(math.NaN()), end: Int(1), field: "start"},
		{name: "float_end", start: Int(1), end: Float(math.NaN()), field: "end"},
		{name: "decimal_start", start: decimalNaN, end: Int(1), field: "start"},
		{name: "decimal_end", start: Int(1), end: decimalNaN, field: "end"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := rules.SelectRange(tc.start, tc.end)
			if !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(err, decimal.ErrInvalidDecimal) {
				t.Fatalf("SelectRange() error = %v, want intlerr.ErrInvalidValue and ErrInvalidDecimal", err)
			}
			testcontract.AssertIntlError(t, err, intlerr.InvalidValue, "pluralrules", tc.field, "NaN", "en")
			testcontract.AssertErrorExpected(t, err, "a numeric value other than NaN")
		})
	}
}

func TestPluralRulesDecimalRejectsMalformed(t *testing.T) {
	t.Parallel()

	if _, err := Decimal("not-a-number"); !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(err, decimal.ErrInvalidDecimal) {
		t.Fatalf("Decimal(invalid) error = %v, want intlerr.ErrInvalidValue and ErrInvalidDecimal", err)
	} else {
		testcontract.AssertIntlError(t, err, intlerr.InvalidValue, "pluralrules", "decimal", "not-a-number", "")
		testcontract.AssertErrorExpected(t, err, "a well-formed decimal string, NaN, Infinity, or -Infinity")
	}
}

func TestPluralRulesInvalidType(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{intltest.Locale(t, "en")}, Options{Type: stringPtr(Type("bad"))})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	}
}

func TestPluralRulesConcurrentSelect(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{Type: stringPtr(Ordinal)})
	if err != nil {
		t.Fatal(err)
	}
	want := map[int]Category{1: One, 2: Two, 3: Few, 4: Other, 21: One}

	var wg sync.WaitGroup
	for in, category := range want {
		wg.Go(func() {
			if got := mustCategory(rules.Select(Int(int64(in)))); got != category {
				t.Errorf("SelectInt(%d) = %s, want %s", in, got, category)
			}
		})
	}
	wg.Wait()
}
