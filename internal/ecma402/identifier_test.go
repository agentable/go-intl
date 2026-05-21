package ecma402_test

import (
	"slices"
	"testing"

	"github.com/agentable/go-intl/internal/ecma402"
)

func TestIsWellFormedCurrencyCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want bool
	}{
		{"USD", true},
		{"usd", true},
		{"EuR", true},
		{"US", false},
		{"USDA", false},
		{"US1", false},
		{"", false},
		{"Ü€$", false},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := ecma402.IsWellFormedCurrencyCode(tc.in); got != tc.want {
				t.Errorf("IsWellFormedCurrencyCode(%q) = %v, want %v",
					tc.in, got, tc.want)
			}
		})
	}
}

func TestIsWellFormedUnicodeType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want bool
	}{
		{"gregory", true},
		{"islamic-umalqura", true},
		{"arab", true},
		{"a", false},
		{"ab", false},
		{"abcdefghi", false},
		{"islamic_", false},
		{"", false},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := ecma402.IsWellFormedUnicodeType(tc.in); got != tc.want {
				t.Fatalf("IsWellFormedUnicodeType(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestLocalizeDigits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		numberingSystem string
		want            string
	}{
		{name: "arab", numberingSystem: "arab", want: "١٬٢٣٤٫٥"},
		{name: "fullwide", numberingSystem: "fullwide", want: "１٬２３４٫５"},
		{name: "hanidec", numberingSystem: "hanidec", want: "一٬二三四٫五"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := ecma402.LocalizeDigits("1٬234٫5", tc.numberingSystem); got != tc.want {
				t.Fatalf("LocalizeDigits() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestIsSanctionedSimpleUnitIdentifier(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want bool
	}{
		{"meter", true},
		{"kilometer", true},
		{"mile-scandinavian", true},
		{"percent", true},
		{"celsius", true},
		{"Meter", false}, // case-sensitive at this layer
		{"length-meter", false},
		{"foo", false},
		{"", false},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := ecma402.IsSanctionedSimpleUnitIdentifier(tc.in); got != tc.want {
				t.Errorf("IsSanctionedSimpleUnitIdentifier(%q) = %v, want %v",
					tc.in, got, tc.want)
			}
		})
	}
}

func TestIsWellFormedUnitIdentifier(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want bool
	}{
		{"meter", true},
		{"meter-per-second", true},
		{"microsecond", true},
		{"nanosecond-per-second", true},
		{"kilometer-per-hour", true},
		{"foot-per-foot", true},
		{"METER", false},
		{"meter-PER-second", false},
		{"foo", false},
		{"meter-per-foo", false},
		{"foo-per-meter", false},
		{"meter-per-meter-per-second", false},
		{"", false},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := ecma402.IsWellFormedUnitIdentifier(tc.in); got != tc.want {
				t.Errorf("IsWellFormedUnitIdentifier(%q) = %v, want %v",
					tc.in, got, tc.want)
			}
		})
	}
}

func TestSanctionedUnitsCount(t *testing.T) {
	t.Parallel()
	if got := len(ecma402.SanctionedUnits); got != 45 {
		t.Errorf("SanctionedUnits count = %d, want 45", got)
	}
}

func TestSanctionedSimpleUnitIdentifiers(t *testing.T) {
	t.Parallel()

	got := ecma402.SanctionedSimpleUnitIdentifiers()
	if len(got) != len(ecma402.SanctionedUnits) {
		t.Fatalf("SanctionedSimpleUnitIdentifiers() length = %d, want %d", len(got), len(ecma402.SanctionedUnits))
	}
	if !slices.IsSorted(got) {
		t.Fatalf("SanctionedSimpleUnitIdentifiers() = %v, want sorted", got)
	}
	for _, want := range []string{"celsius", "meter", "mile-scandinavian", "percent"} {
		if !slices.Contains(got, want) {
			t.Fatalf("SanctionedSimpleUnitIdentifiers() missing %q in %v", want, got)
		}
	}

	got[0] = "mutated"
	if slices.Contains(ecma402.SanctionedSimpleUnitIdentifiers(), "mutated") {
		t.Fatal("SanctionedSimpleUnitIdentifiers() returned shared mutable storage")
	}
}
