package decimal

import "testing"

func TestIsValidRoundingIncrement(t *testing.T) {
	t.Parallel()

	valid := map[int]bool{
		1: true, 2: true, 5: true, 10: true, 20: true, 25: true, 50: true,
		100: true, 200: true, 250: true, 500: true,
		1000: true, 2000: true, 2500: true, 5000: true,
	}
	for i := range 5001 {
		got := IsValidRoundingIncrement(i)
		if got != valid[i] {
			t.Fatalf("IsValidRoundingIncrement(%d) = %v, want %v", i, got, valid[i])
		}
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
