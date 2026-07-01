package decimal

import (
	"slices"
	"testing"
)

func TestIsValidRoundingIncrement(t *testing.T) {
	t.Parallel()

	valid := make(map[int]bool)
	for _, increment := range RoundingIncrements() {
		valid[increment] = true
	}
	for i := range 5001 {
		got := IsValidRoundingIncrement(i)
		if got != valid[i] {
			t.Fatalf("IsValidRoundingIncrement(%d) = %v, want %v", i, got, valid[i])
		}
	}
}

func TestRoundingIncrementsReturnsCopy(t *testing.T) {
	t.Parallel()

	got := RoundingIncrements()
	want := []int{1, 2, 5, 10, 20, 25, 50, 100, 200, 250, 500, 1000, 2000, 2500, 5000}
	if !slices.Equal(got, want) {
		t.Fatalf("RoundingIncrements() = %v, want %v", got, want)
	}
	got[0] = 3
	if next := RoundingIncrements(); !slices.Equal(next, want) {
		t.Fatalf("RoundingIncrements() after caller mutation = %v, want %v", next, want)
	}
}

func TestQuantizeToIncrement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		x         string
		increment int
		exp       int32
		mode      RoundingMode
		want      string
	}{
		{name: "rounds up", x: "1.24", increment: 5, exp: -2, mode: RoundHalfExpand, want: "1.25"},
		{name: "rounds down", x: "1.22", increment: 5, exp: -2, mode: RoundHalfExpand, want: "1.20"},
		{name: "half even lower", x: "1.225", increment: 5, exp: -2, mode: RoundHalfEven, want: "1.20"},
		{name: "half even upper", x: "1.275", increment: 5, exp: -2, mode: RoundHalfEven, want: "1.30"},
		{name: "infinity unchanged", x: "Infinity", increment: 5, exp: -2, mode: RoundHalfExpand, want: "Infinity"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := QuantizeToIncrement(mustParseDecimal(t, tc.x), tc.increment, tc.exp, tc.mode)
			if got.String() != tc.want {
				t.Fatalf("QuantizeToIncrement() = %q, want %q", got.String(), tc.want)
			}
		})
	}
}
