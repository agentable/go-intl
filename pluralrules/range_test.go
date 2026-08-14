package pluralrules

import (
	"errors"
	"math"
	"testing"

	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
)

func TestPluralRulesSelectRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
		start  int
		end    int
		want   Category
	}{
		{name: "english one to other", locale: "en", start: 1, end: 2, want: Other},
		{name: "english equal", locale: "en", start: 1, end: 1, want: One},
		{name: "french equal zero", locale: "fr", start: 0, end: 0, want: One},
		{name: "french one to other", locale: "fr", start: 0, end: 2, want: Other},
		{name: "czech missing few to one", locale: "cs", start: 2, end: 1, want: Other},
		{name: "czech explicit one to few", locale: "cs", start: 1, end: 2, want: Few},
		{name: "chinese fallback", locale: "zh", start: 1, end: 5, want: Other},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rules, err := New(locale.List{intltest.Locale(t, tc.locale)}, Options{})
			if err != nil {
				t.Fatal(err)
			}
			if got := mustRangeCategory(rules.SelectRange(Int(int64(tc.start)), Int(int64(tc.end)))); got != tc.want {
				t.Fatalf("SelectRangeInt(%v, %v) = %s, want %s", tc.start, tc.end, got, tc.want)
			}
		})
	}
}

func TestPluralRulesSelectRangeDecimal(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := rules.SelectRange(decimalValue("1"), decimalValue("1.0"))
	if err != nil {
		t.Fatal(err)
	}
	if got != One {
		t.Fatalf("SelectRangeDecimal(1, 1.0) = %s, want %s", got, One)
	}

	got, err = rules.SelectRange(decimalValue("1.0"), decimalValue("1.00"))
	if err != nil {
		t.Fatal(err)
	}
	if got != One {
		t.Fatalf("SelectRangeDecimal(1.0, 1.00) = %s, want %s", got, One)
	}

	got, err = rules.SelectRange(decimalValue("1"), decimalValue("1"))
	if err != nil {
		t.Fatal(err)
	}
	if got != One {
		t.Fatalf("SelectRangeDecimal(1, 1) = %s, want %s", got, One)
	}
}

func TestPluralRulesSelectRangeReversedPreservesInputOrder(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "az")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := mustRangeCategory(rules.SelectRange(Int(int64(2)), Int(int64(1)))); got != One {
		t.Fatalf("SelectRangeInt(2, 1) = %s, want %s", got, One)
	}
	if got := mustRangeCategory(rules.SelectRange(Uint(uint64(2)), Uint(uint64(1)))); got != One {
		t.Fatalf("SelectRangeUint(2, 1) = %s, want %s", got, One)
	}
	gotFloat, err := rules.SelectRange(Float(2), Float(1))
	if err != nil {
		t.Fatal(err)
	}
	if gotFloat != One {
		t.Fatalf("SelectRangeFloat64(2, 1) = %s, want %s", gotFloat, One)
	}
	gotDecimal, err := rules.SelectRange(decimalValue("2"), decimalValue("1"))
	if err != nil {
		t.Fatal(err)
	}
	if gotDecimal != One {
		t.Fatalf("SelectRangeDecimal(2, 1) = %s, want %s", gotDecimal, One)
	}
}

func TestPluralRulesSelectRangeUsesFormattedEqualityAfterRounding(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{MaximumFractionDigits: intPtr(0)})
	if err != nil {
		t.Fatal(err)
	}
	got, err := rules.SelectRange(decimalValue("1.2"), decimalValue("1.3"))
	if err != nil {
		t.Fatal(err)
	}
	if got != One {
		t.Fatalf("SelectRangeDecimal(1.2, 1.3) = %s, want %s", got, One)
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

func TestPluralRulesUnsignedSelectionWrappers(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := mustRangeCategory(rules.SelectRange(Uint(uint64(1)), Uint(uint64(2)))); got != Other {
		t.Fatalf("SelectRangeUint(1, 2) = %s, want %s", got, Other)
	}
	if got := mustRangeCategory(rules.SelectRange(Uint(uint64(1)), Uint(uint64(1)))); got != One {
		t.Fatalf("SelectRangeUint(1, 1) = %s, want %s", got, One)
	}
	if got := mustRangeCategory(rules.SelectRange(Uint(2), Uint(2))); got != Other {
		t.Fatalf("SelectRangeUint64(2, 2) = %s, want %s", got, Other)
	}
}
