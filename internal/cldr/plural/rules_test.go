package plural

import (
	"strconv"
	"testing"

	pluralop "github.com/agentable/go-intl/internal/plural"
)

func TestCardinalRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		locale string
		op     pluralop.OperandsRecord
		want   pluralop.Category
	}{
		{locale: "en", op: operands("1"), want: pluralop.One},
		{locale: "en", op: operands("2"), want: pluralop.Other},
		{locale: "pl", op: operands("1"), want: pluralop.One},
		{locale: "pl", op: operands("2"), want: pluralop.Few},
		{locale: "pl", op: operands("5"), want: pluralop.Many},
		{locale: "ar", op: operands("0"), want: pluralop.Zero},
		{locale: "fr", op: operands("1000000"), want: pluralop.Many},
		{locale: "fr", op: operands("1000001"), want: pluralop.Other},
		{locale: "zh", op: operands("1"), want: pluralop.Other},
	}
	for _, tc := range tests {
		t.Run(tc.locale+"/"+tc.want.String(), func(t *testing.T) {
			t.Parallel()
			rule, ok := CardinalRule(tc.locale)
			if !ok {
				t.Fatalf("CardinalRule(%q) missing", tc.locale)
			}
			if got := rule(tc.op); got != tc.want {
				t.Fatalf("CardinalRule(%q) = %s, want %s", tc.locale, got, tc.want)
			}
		})
	}
}

func TestOrdinalRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		i    int
		want pluralop.Category
	}{
		{i: 1, want: pluralop.One},
		{i: 2, want: pluralop.Two},
		{i: 3, want: pluralop.Few},
		{i: 4, want: pluralop.Other},
		{i: 11, want: pluralop.Other},
		{i: 21, want: pluralop.One},
	}
	rule, ok := OrdinalRule("en")
	if !ok {
		t.Fatal("OrdinalRule(en) missing")
	}
	for _, tc := range tests {
		t.Run(tc.want.String(), func(t *testing.T) {
			t.Parallel()
			if got := rule(operands(strconv.Itoa(tc.i))); got != tc.want {
				t.Fatalf("OrdinalRule(en)(%d) = %s, want %s", tc.i, got, tc.want)
			}
		})
	}
}

func TestOrdinalRulesUseExactLargeOperands(t *testing.T) {
	t.Parallel()

	rule, ok := OrdinalRule("en")
	if !ok {
		t.Fatal("OrdinalRule(en) missing")
	}
	tests := []struct {
		in   string
		want pluralop.Category
	}{
		{in: "10000000000001", want: pluralop.One},
		{in: "10000000000002", want: pluralop.Two},
		{in: "10000000000003", want: pluralop.Few},
		{in: "10000000000011", want: pluralop.Other},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := rule(operands(tc.in)); got != tc.want {
				t.Fatalf("OrdinalRule(en)(%s) = %s, want %s", tc.in, got, tc.want)
			}
		})
	}
}

func operands(formatted string) pluralop.OperandsRecord {
	return pluralop.GetOperands(formatted, 0)
}
