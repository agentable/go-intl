package pluralrules

import (
	"errors"
	"math"
	"sync"
	"testing"

	"github.com/agentable/go-intl/locale"
)

func TestPluralRulesSelectCardinal(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.MustParse("en"))
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
		if got := rules.SelectInt(tc.in); got != tc.want {
			t.Fatalf("SelectInt(%v) = %s, want %s", tc.in, got, tc.want)
		}
	}
}

func TestPluralRulesSelectOrdinal(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.MustParse("en"), Options{Type: Ordinal})
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
		if got := rules.SelectInt(tc.in); got != tc.want {
			t.Fatalf("SelectInt(%v) = %s, want %s", tc.in, got, tc.want)
		}
	}
}

func TestPluralRulesSelectUsesExactLargeOperands(t *testing.T) {
	t.Parallel()

	ordinal, err := New(locale.MustParse("en"), Options{Type: Ordinal})
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
		got, err := ordinal.SelectDecimal(tc.in)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("ordinal SelectDecimal(%s) = %s, want %s", tc.in, got, tc.want)
		}
	}

	cardinal, err := New(locale.MustParse("fr"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := cardinal.SelectDecimal("1000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if got != Many {
		t.Fatalf("fr SelectDecimal(1000000000000) = %s, want %s", got, Many)
	}
	got, err = cardinal.SelectDecimal("1000000000001")
	if err != nil {
		t.Fatal(err)
	}
	if got != Other {
		t.Fatalf("fr SelectDecimal(1000000000001) = %s, want %s", got, Other)
	}
}

func TestPluralRulesTypedSelectErrors(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.MustParse("en"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rules.SelectFloat64(math.NaN()); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("SelectFloat64(NaN) error = %v, want ErrInvalidValue", err)
	}
	if _, err := rules.SelectDecimal("not-a-number"); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("SelectDecimal(invalid) error = %v, want ErrInvalidValue", err)
	}
	if _, err := rules.SelectRangeFloat64(1, math.Inf(1)); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("SelectRangeFloat64(infinite) error = %v, want ErrInvalidValue", err)
	}
	if _, err := rules.SelectRangeDecimal("1", "NaN"); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("SelectRangeDecimal(NaN) error = %v, want ErrInvalidValue", err)
	}
}

func TestPluralRulesInvalidType(t *testing.T) {
	t.Parallel()

	_, err := New(locale.MustParse("en"), Options{Type: Type(99)})
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("New() error = %v, want ErrInvalidOption", err)
	}
}

func TestPluralRulesConcurrentSelect(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.MustParse("en"), Options{Type: Ordinal})
	if err != nil {
		t.Fatal(err)
	}
	want := map[int]Category{1: One, 2: Two, 3: Few, 4: Other, 21: One}

	var wg sync.WaitGroup
	for in, category := range want {
		wg.Go(func() {
			if got := rules.SelectInt(in); got != category {
				t.Errorf("SelectInt(%d) = %s, want %s", in, got, category)
			}
		})
	}
	wg.Wait()
}
