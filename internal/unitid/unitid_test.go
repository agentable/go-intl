package unitid

import (
	"slices"
	"testing"
)

func TestSanctionedSimpleUnitIdentifiers(t *testing.T) {
	t.Parallel()

	got := SanctionedSimpleUnitIdentifiers()
	namespaced := SanctionedUnitIdentifiers()
	if len(got) != len(namespaced) {
		t.Fatalf("SanctionedSimpleUnitIdentifiers() length = %d, want %d", len(got), len(namespaced))
	}
	if !slices.IsSorted(got) {
		t.Fatalf("SanctionedSimpleUnitIdentifiers() = %v, want sorted", got)
	}
	want := []string{
		"acre",
		"bit",
		"byte",
		"celsius",
		"centimeter",
		"day",
		"degree",
		"fahrenheit",
		"fluid-ounce",
		"foot",
		"gallon",
		"gigabit",
		"gigabyte",
		"gram",
		"hectare",
		"hour",
		"inch",
		"kilobit",
		"kilobyte",
		"kilogram",
		"kilometer",
		"liter",
		"megabit",
		"megabyte",
		"meter",
		"microsecond",
		"mile",
		"mile-scandinavian",
		"milliliter",
		"millimeter",
		"millisecond",
		"minute",
		"month",
		"nanosecond",
		"ounce",
		"percent",
		"petabyte",
		"pound",
		"second",
		"stone",
		"terabit",
		"terabyte",
		"week",
		"yard",
		"year",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("SanctionedSimpleUnitIdentifiers() = %v, want %v", got, want)
	}
	for i := 1; i < len(got); i++ {
		if got[i-1] == got[i] {
			t.Fatalf("SanctionedSimpleUnitIdentifiers() contains duplicate %q", got[i])
		}
	}
	for _, unit := range []string{"fluid-ounce", "kilobit", "megabit", "mile-scandinavian", "terabit"} {
		if !slices.Contains(got, unit) {
			t.Fatalf("SanctionedSimpleUnitIdentifiers() missing %q", unit)
		}
	}
	got[0] = "mutated"
	if SanctionedSimpleUnitIdentifiers()[0] == "mutated" {
		t.Fatal("SanctionedSimpleUnitIdentifiers() returned shared backing storage")
	}
}

func TestSanctionedUnitIdentifiersReturnsCopy(t *testing.T) {
	t.Parallel()

	got := SanctionedUnitIdentifiers()
	if len(got) != 45 {
		t.Fatalf("SanctionedUnitIdentifiers() length = %d, want 45", len(got))
	}
	for _, unit := range []string{"digital-kilobit", "length-mile-scandinavian", "volume-fluid-ounce"} {
		if !slices.Contains(got, unit) {
			t.Fatalf("SanctionedUnitIdentifiers() missing %q", unit)
		}
	}
	got[0] = "mutated"
	if SanctionedUnitIdentifiers()[0] == "mutated" {
		t.Fatal("SanctionedUnitIdentifiers() returned shared backing storage")
	}
}

func TestIsWellFormedUnitIdentifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		unit string
		want bool
	}{
		{unit: "meter", want: true},
		{unit: "meter-per-second", want: true},
		{unit: "fluid-ounce", want: true},
		{unit: "hertz", want: false},
		{unit: "milligram", want: false},
		{unit: "meter-per-hertz", want: false},
		{unit: "meter-per-second-per-hour", want: false},
		{unit: "METER", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.unit, func(t *testing.T) {
			t.Parallel()

			if got := IsWellFormedUnitIdentifier(tt.unit); got != tt.want {
				t.Fatalf("IsWellFormedUnitIdentifier(%q) = %v, want %v", tt.unit, got, tt.want)
			}
		})
	}
}
