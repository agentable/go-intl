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
		if got := rules.Select(Int(int64(tc.in))); got != tc.want {
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
		if got := rules.Select(Int(int64(tc.in))); got != tc.want {
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
		got := ordinal.Select(decimalValue(tc.in))
		if got != tc.want {
			t.Fatalf("ordinal SelectDecimal(%s) = %s, want %s", tc.in, got, tc.want)
		}
	}

	cardinal, err := New(locale.List{intltest.Locale(t, "fr")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	got := cardinal.Select(decimalValue("1000000000000"))
	if got != Many {
		t.Fatalf("fr SelectDecimal(1000000000000) = %s, want %s", got, Many)
	}
	got = cardinal.Select(decimalValue("1000000000001"))
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
			got := rules.Select(decimalValue(tc.input))
			if got != tc.want {
				t.Errorf("Select(ParseValue(%q)) = %s, want %s", tc.input, got, tc.want)
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
	input, err := ParseValue(inputText)
	if err != nil {
		t.Fatal(err)
	}
	if got := rules.Select(input); got != One {
		t.Fatalf("Select(ParseValue(%q)) = %s, want %s", inputText, got, One)
	}
}

func TestPluralRulesNotationAffectsOperands(t *testing.T) {
	t.Parallel()

	standard, err := New(locale.List{intltest.Locale(t, "ca")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	got := standard.Select(decimalValue("12345678"))
	if got != Other {
		t.Fatalf("standard SelectDecimal(12345678) = %s, want %s", got, Other)
	}
	if got := standard.Select(Int(12345678)); got != Other {
		t.Fatalf("standard SelectInt64(12345678) = %s, want %s", got, Other)
	}

	for _, notation := range []Notation{ScientificNotation, EngineeringNotation} {
		for _, tc := range []struct {
			name, loc, text string
			value           float64
			min, max        int
			want            Category
		}{
			{"full magnitude", "en", "1000", 1000, 0, 3, Other},
			{"negative full magnitude", "en", "-1000", -1000, 0, 3, Other},
			{"french exponent", "fr", "12345678", 12345678, 0, 3, Many},
			{"catalan exponent", "ca", "12345678", 12345678, 0, 3, Many},
			{"carry", "ca", "999999.6", 999999.6, 0, 0, Many},
			{"visible zeros", "en", "1", 1, 2, 3, Other},
			{"source rounding", "ru", "1001", 1001, 0, 0, One},
			{"negative exponent", "en", "0.001", 0.001, 2, 3, Other},
			{"source rounded zero", "ar", "0.001", 0.001, 0, 0, Zero},
		} {
			t.Run(string(notation)+"/"+tc.name, func(t *testing.T) {
				t.Parallel()
				rules, err := New(intltest.LocaleList(t, tc.loc), Options{
					Notation:              stringPtr(notation),
					MinimumFractionDigits: intPtr(tc.min),
					MaximumFractionDigits: intPtr(tc.max),
				})
				if err != nil {
					t.Fatal(err)
				}
				values := []Value{Float(tc.value), decimalValue(tc.text)}
				if math.Trunc(tc.value) == tc.value {
					values = append(values, Int(int64(tc.value)))
				}
				for i, value := range values {
					if got := rules.Select(value); got != tc.want {
						t.Errorf("Select(%s, bridge %d) = %s, want %s", tc.text, i, got, tc.want)
					}
				}
			})
		}
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
		{name: "scientific", notation: ScientificNotation, wantFormatted: "12345678", wantExponent: 7},
		{name: "engineering", notation: EngineeringNotation, wantFormatted: "12345678", wantExponent: 6},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotOps pluralop.OperandsRecord
			formatted, category := resolveDecimal(d, tc.notation, ecma402nf.ResolvedDigitOptions{DigitOptions: digitOptions, RoundingType: ecma402nf.RoundingTypeFractionDigits}, compactExponentSet{}, func(ops pluralop.OperandsRecord) pluralop.Category {
				gotOps = ops
				return pluralop.Many
			})
			if formatted != tc.wantFormatted {
				t.Fatalf("resolveDecimal formatted = %q, want %q", formatted, tc.wantFormatted)
			}
			if category != Many {
				t.Fatalf("resolveDecimal category = %s, want %s", category, Many)
			}
			if gotOps.N.String() != "12345678" || gotOps.I.String() != "12345678" || gotOps.V != 0 {
				t.Fatalf("resolveDecimal operands = N:%s I:%s V:%d; want source 12345678 with no fractional digits", gotOps.N, gotOps.I, gotOps.V)
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

	if got := rules.Select(Uint(uint64(1))); got != One {
		t.Fatalf("SelectUint(1) = %s, want %s", got, One)
	}
	if got := rules.Select(Uint(2)); got != Other {
		t.Fatalf("SelectUint64(2) = %s, want %s", got, Other)
	}
	ordinal, err := New(locale.List{intltest.Locale(t, "en")}, Options{Type: stringPtr(Ordinal)})
	if err != nil {
		t.Fatal(err)
	}
	if got := ordinal.Select(Uint(10000000000001)); got != One {
		t.Fatalf("ordinal SelectUint64(10000000000001) = %s, want %s", got, One)
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

			got := rules.Select(Float(tc.value))
			if got != Other {
				t.Fatalf("Select(Float(%v)) = %s, want %s", tc.value, got, Other)
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

			input, err := ParseValue(value)
			if err != nil {
				t.Fatalf("ParseValue(%q) error = %v", value, err)
			}
			got := rules.Select(input)
			if got != Other {
				t.Fatalf("Select(ParseValue(%q)) = %s, want %s", value, got, Other)
			}
		})
	}
}

func TestPluralRulesDecimalRejectsMalformed(t *testing.T) {
	t.Parallel()

	if _, err := ParseValue("not-a-number"); !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(err, decimal.ErrInvalidDecimal) {
		t.Fatalf("ParseValue(invalid) error = %v, want intlerr.ErrInvalidValue and ErrInvalidDecimal", err)
	} else {
		testcontract.AssertIntlError(t, err, intlerr.InvalidValue, "pluralrules", "decimal", "not-a-number", "")
		testcontract.AssertErrorExpected(t, err, "an Intl numeric string (decimal, binary, octal, hexadecimal, or Infinity), or NaN")
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
			if got := rules.Select(Int(int64(in))); got != category {
				t.Errorf("SelectInt(%d) = %s, want %s", in, got, category)
			}
		})
	}
	wg.Wait()
}
