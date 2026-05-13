package plural

import (
	"strconv"
	"testing"

	ecma402pr "github.com/agentable/go-intl/internal/ecma402/pluralrules"
)

func TestCardinalRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		locale string
		op     ecma402pr.OperandsRecord
		want   ecma402pr.Category
	}{
		{locale: "en", op: operands("1"), want: ecma402pr.One},
		{locale: "en", op: operands("2"), want: ecma402pr.Other},
		{locale: "pl", op: operands("1"), want: ecma402pr.One},
		{locale: "pl", op: operands("2"), want: ecma402pr.Few},
		{locale: "pl", op: operands("5"), want: ecma402pr.Many},
		{locale: "ar", op: operands("0"), want: ecma402pr.Zero},
		{locale: "fr", op: operands("1000000"), want: ecma402pr.Many},
		{locale: "fr", op: operands("1000001"), want: ecma402pr.Other},
		{locale: "zh", op: operands("1"), want: ecma402pr.Other},
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
		want ecma402pr.Category
	}{
		{i: 1, want: ecma402pr.One},
		{i: 2, want: ecma402pr.Two},
		{i: 3, want: ecma402pr.Few},
		{i: 4, want: ecma402pr.Other},
		{i: 11, want: ecma402pr.Other},
		{i: 21, want: ecma402pr.One},
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
		want ecma402pr.Category
	}{
		{in: "10000000000001", want: ecma402pr.One},
		{in: "10000000000002", want: ecma402pr.Two},
		{in: "10000000000003", want: ecma402pr.Few},
		{in: "10000000000011", want: ecma402pr.Other},
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

func operands(formatted string) ecma402pr.OperandsRecord {
	return ecma402pr.GetOperands(formatted, 0)
}
