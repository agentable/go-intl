package pluralrules

import (
	"errors"
	"math"
	"sync"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/locale"
)

func TestPluralRulesSelectCardinal(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{locale.MustParse("en")}, Options{})
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

	rules, err := New(locale.List{locale.MustParse("en")}, Options{Type: Ordinal})
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

	ordinal, err := New(locale.List{locale.MustParse("en")}, Options{Type: Ordinal})
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

	cardinal, err := New(locale.List{locale.MustParse("fr")}, Options{})
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

func TestPluralRulesNotationAffectsOperands(t *testing.T) {
	t.Parallel()

	standard, err := New(locale.List{locale.MustParse("ca")}, Options{})
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
		{name: "scientific", opts: Options{Notation: ScientificNotation}},
		{name: "engineering", opts: Options{Notation: EngineeringNotation}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rules, err := New(locale.List{locale.MustParse("ca")}, tc.opts)
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

func TestPluralRulesUnsignedTypedBridges(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{locale.MustParse("en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}

	if got := mustCategory(rules.Select(Uint(uint64(1)))); got != One {
		t.Fatalf("SelectUint(1) = %s, want %s", got, One)
	}
	if got := mustCategory(rules.Select(Uint(2))); got != Other {
		t.Fatalf("SelectUint64(2) = %s, want %s", got, Other)
	}
	ordinal, err := New(locale.List{locale.MustParse("en")}, Options{Type: Ordinal})
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

func TestPluralRulesTypedSelectErrors(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{locale.MustParse("en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rules.Select(Float(math.NaN())); !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(err, decimal.ErrInvalidDecimal) {
		t.Fatalf("SelectFloat64(NaN) error = %v, want intlerr.ErrInvalidValue and ErrInvalidDecimal", err)
	}
	if _, err := Decimal("not-a-number"); !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(err, decimal.ErrInvalidDecimal) {
		t.Fatalf("Decimal(invalid) error = %v, want intlerr.ErrInvalidValue and ErrInvalidDecimal", err)
	}
	if _, err := rules.SelectRange(Float(1), Float(math.Inf(1))); !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(err, decimal.ErrInvalidDecimal) {
		t.Fatalf("SelectRangeFloat64(infinite) error = %v, want intlerr.ErrInvalidValue and ErrInvalidDecimal", err)
	}
	if _, err := Decimal("NaN"); !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(err, decimal.ErrInvalidDecimal) {
		t.Fatalf("Decimal(NaN) error = %v, want intlerr.ErrInvalidValue and ErrInvalidDecimal", err)
	}
}

func TestPluralRulesInvalidType(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{locale.MustParse("en")}, Options{Type: Type(99)})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	}
}

func TestPluralRulesConcurrentSelect(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.List{locale.MustParse("en")}, Options{Type: Ordinal})
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
