package pluralrules

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
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

func TestPluralRulesUnsignedSelectionWrappers(t *testing.T) {
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
	if got := mustRangeCategory(rules.SelectRange(Uint(uint64(1)), Uint(uint64(1)))); got != One {
		t.Fatalf("SelectRangeUint(1, 1) = %s, want %s", got, One)
	}
	if got := mustRangeCategory(rules.SelectRange(Uint(2), Uint(2))); got != Other {
		t.Fatalf("SelectRangeUint64(2, 2) = %s, want %s", got, Other)
	}
}
