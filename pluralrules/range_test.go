package pluralrules

import (
	"testing"

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
		{name: "chinese fallback", locale: "zh", start: 1, end: 5, want: Other},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rules, err := New(locale.MustParse(tc.locale))
			if err != nil {
				t.Fatal(err)
			}
			if got := rules.SelectRangeInt(tc.start, tc.end); got != tc.want {
				t.Fatalf("SelectRangeInt(%v, %v) = %s, want %s", tc.start, tc.end, got, tc.want)
			}
		})
	}
}

func TestPluralRulesSelectRangeDecimal(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.MustParse("en"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := rules.SelectRangeDecimal("1.0", "1.00")
	if err != nil {
		t.Fatal(err)
	}
	if got != Other {
		t.Fatalf("SelectRangeDecimal(1.0, 1.00) = %s, want %s", got, Other)
	}

	got, err = rules.SelectRangeDecimal("1", "1")
	if err != nil {
		t.Fatal(err)
	}
	if got != One {
		t.Fatalf("SelectRangeDecimal(1, 1) = %s, want %s", got, One)
	}
}

func TestPluralRulesSelectRangeUsesRoundedEquality(t *testing.T) {
	t.Parallel()

	rules, err := New(locale.MustParse("en"), Options{FractionDigits: MaximumFractionDigits(0)})
	if err != nil {
		t.Fatal(err)
	}
	got, err := rules.SelectRangeDecimal("1.2", "1.3")
	if err != nil {
		t.Fatal(err)
	}
	if got != One {
		t.Fatalf("SelectRangeDecimal(1.2, 1.3) = %s, want %s", got, One)
	}
}
