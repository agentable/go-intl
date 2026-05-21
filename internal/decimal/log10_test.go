package decimal

import (
	"errors"
	"strings"
	"testing"
)

func TestLog10Floor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want int32
	}{
		{in: "1", want: 0},
		{in: "9", want: 0},
		{in: "10", want: 1},
		{in: "99", want: 1},
		{in: "100", want: 2},
		{in: "500", want: 2},
		{in: "0.1", want: -1},
		{in: "0.01", want: -2},
		{in: "0.001", want: -3},
		{in: "0.0099", want: -3},
		{in: "1e41", want: 41},
		{in: "1" + strings.Repeat("0", 350), want: 350},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			got, err := Log10Floor(mustParseDecimal(t, tc.in))
			if err != nil {
				t.Fatalf("Log10Floor(%q) err = %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("Log10Floor(%q) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestLog10FloorDomain(t *testing.T) {
	t.Parallel()

	for _, d := range []Decimal{Zero, FromInt64(-1), NaNValue, PosInfinity, NegInfinity} {
		_, err := Log10Floor(d)
		if !errors.Is(err, ErrLog10Domain) {
			t.Fatalf("err = %v, want errors.Is(ErrLog10Domain)", err)
		}
	}
}
