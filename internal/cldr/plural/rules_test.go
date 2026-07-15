package plural

import (
	"slices"
	"strconv"
	"strings"
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

func TestCardinalRuleUnknownLocaleReturnsNotFound(t *testing.T) {
	t.Parallel()

	rule, ok := CardinalRule("und")
	if ok || rule != nil {
		t.Fatalf("CardinalRule(und) found = %t, rule nil = %t; want false, true", ok, rule == nil)
	}
}

func TestRuleRejectsMissingFamily(t *testing.T) {
	t.Parallel()

	tests := []struct {
		typ string
	}{
		{typ: "cardinal"},
		{typ: "ordinal"},
	}
	for _, tc := range tests {
		t.Run(tc.typ, func(t *testing.T) {
			t.Parallel()

			rule, err := Rule("und", tc.typ)
			if err == nil || !strings.Contains(err.Error(), `plural: missing `+tc.typ+` rule for data locale "und"`) {
				t.Fatalf("Rule(und, %s) error = %v, want missing-family context", tc.typ, err)
			}
			if rule != nil {
				t.Fatalf("Rule(und, %s) returned non-nil rule", tc.typ)
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

func TestOrdinalRuleUnknownLocaleReturnsNotFound(t *testing.T) {
	t.Parallel()

	rule, ok := OrdinalRule("und")
	if ok || rule != nil {
		t.Fatalf("OrdinalRule(und) found = %t, rule nil = %t; want false, true", ok, rule == nil)
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

func TestRangeRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		loc   string
		start pluralop.Category
		end   pluralop.Category
		want  pluralop.Category
		ok    bool
	}{
		{name: "am one to one", loc: "am", start: pluralop.One, end: pluralop.One, want: pluralop.One, ok: true},
		{name: "ar zero to one", loc: "ar", start: pluralop.Zero, end: pluralop.One, want: pluralop.Zero, ok: true},
		{name: "ar one to few", loc: "ar", start: pluralop.One, end: pluralop.Few, want: pluralop.Few, ok: true},
		{name: "unknown locale", loc: "und", start: pluralop.One, end: pluralop.One},
		{name: "unknown category pair", loc: "af", start: pluralop.One, end: pluralop.One},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := CardinalRange(tc.loc, tc.start, tc.end)
			if ok != tc.ok {
				t.Fatalf("CardinalRange(%q, %s, %s) ok = %v, want %v", tc.loc, tc.start, tc.end, ok, tc.ok)
			}
			if got != tc.want {
				t.Fatalf("CardinalRange(%q, %s, %s) = %s, want %s", tc.loc, tc.start, tc.end, got, tc.want)
			}
		})
	}
}

func TestCategoriesUsesGeneratedOrderAndReturnsFreshSlice(t *testing.T) {
	t.Parallel()

	got := Categories("ar", "cardinal")
	want := []pluralop.Category{pluralop.Zero, pluralop.One, pluralop.Two, pluralop.Few, pluralop.Many, pluralop.Other}
	if !slices.Equal(got, want) {
		t.Fatalf("Categories(ar, cardinal) = %v, want %v", got, want)
	}

	got[0] = pluralop.Other
	if next := Categories("ar", "cardinal"); !slices.Equal(next, want) {
		t.Fatalf("Categories(ar, cardinal) after caller mutation = %v, want %v", next, want)
	}

	ordinal := Categories("en", "ordinal")
	wantOrdinal := []pluralop.Category{pluralop.One, pluralop.Two, pluralop.Few, pluralop.Other}
	if !slices.Equal(ordinal, wantOrdinal) {
		t.Fatalf("Categories(en, ordinal) = %v, want %v", ordinal, wantOrdinal)
	}

	fallback := Categories("und", "cardinal")
	if !slices.Equal(fallback, []pluralop.Category{pluralop.Other}) {
		t.Fatalf("Categories(und, cardinal) = %v, want [other]", fallback)
	}
}

func TestSupportedLocalesReturnsFreshSlice(t *testing.T) {
	t.Parallel()

	got := SupportedLocales()
	want := []string{"af", "am", "ar"}
	if !slices.Equal(got[:3], want) {
		t.Fatalf("SupportedLocales() prefix = %v, want %v", got[:3], want)
	}
	got[0] = "mutated"
	next := SupportedLocales()
	if !slices.Equal(next[:3], want) {
		t.Fatalf("SupportedLocales() after caller mutation prefix = %v, want %v", next[:3], want)
	}
}

func operands(formatted string) pluralop.OperandsRecord {
	return pluralop.GetOperands(formatted, 0)
}
