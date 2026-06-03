package decimal

import (
	"errors"
	"testing"
)

func TestParseRoundingMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want RoundingMode
	}{
		{in: "ceil", want: RoundCeil},
		{in: "floor", want: RoundFloor},
		{in: "expand", want: RoundExpand},
		{in: "trunc", want: RoundTrunc},
		{in: "halfCeil", want: RoundHalfCeil},
		{in: "halfFloor", want: RoundHalfFloor},
		{in: "halfExpand", want: RoundHalfExpand},
		{in: "halfTrunc", want: RoundHalfTrunc},
		{in: "halfEven", want: RoundHalfEven},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			got, err := ParseRoundingMode(tc.in)
			if err != nil {
				t.Fatalf("ParseRoundingMode(%q) err = %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("ParseRoundingMode(%q) = %v, want %v", tc.in, got, tc.want)
			}
			if got.String() != tc.in {
				t.Fatalf("String() = %q, want %q", got.String(), tc.in)
			}
		})
	}
}

func TestParseRoundingModeInvalid(t *testing.T) {
	t.Parallel()

	_, err := ParseRoundingMode("halfFoo")
	if !errors.Is(err, ErrInvalidDecimal) {
		t.Fatalf("err = %v, want errors.Is(ErrInvalidDecimal)", err)
	}
}

func TestUnsignedRoundingMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode     RoundingMode
		negative bool
		want     RoundingMode
	}{
		{mode: RoundCeil, want: RoundExpand},
		{mode: RoundFloor, want: RoundTrunc},
		{mode: RoundExpand, want: RoundExpand},
		{mode: RoundTrunc, want: RoundTrunc},
		{mode: RoundHalfCeil, want: RoundHalfExpand},
		{mode: RoundHalfFloor, want: RoundHalfTrunc},
		{mode: RoundHalfExpand, want: RoundHalfExpand},
		{mode: RoundHalfTrunc, want: RoundHalfTrunc},
		{mode: RoundHalfEven, want: RoundHalfEven},
		{mode: RoundCeil, negative: true, want: RoundTrunc},
		{mode: RoundFloor, negative: true, want: RoundExpand},
		{mode: RoundHalfCeil, negative: true, want: RoundHalfTrunc},
		{mode: RoundHalfFloor, negative: true, want: RoundHalfExpand},
	}
	for _, tc := range tests {
		t.Run(tc.mode.String(), func(t *testing.T) {
			t.Parallel()
			got := UnsignedRoundingMode(tc.mode, tc.negative)
			if got != tc.want {
				t.Fatalf("UnsignedRoundingMode(%v, %v) = %v, want %v", tc.mode, tc.negative, got, tc.want)
			}
		})
	}
}

func TestApplyUnsignedRoundingMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		x    string
		r1   string
		r2   string
		mode RoundingMode
		want string
	}{
		{name: "x equals lower", x: "1", r1: "1", r2: "2", mode: RoundTrunc, want: "1"},
		{name: "x equals upper", x: "2", r1: "1", r2: "2", mode: RoundTrunc, want: "2"},
		{name: "same boundaries", x: "1", r1: "1", r2: "1", mode: RoundExpand, want: "1"},
		{name: "zero chooses lower", x: "1", r1: "0", r2: "2", mode: RoundTrunc, want: "0"},
		{name: "infinity chooses upper", x: "1", r1: "0", r2: "2", mode: RoundExpand, want: "2"},
		{name: "closer lower", x: "1", r1: "0", r2: "4", mode: RoundHalfExpand, want: "0"},
		{name: "closer upper", x: "3", r1: "0", r2: "4", mode: RoundHalfExpand, want: "4"},
		{name: "tie half trunc", x: "1", r1: "0", r2: "2", mode: RoundHalfTrunc, want: "0"},
		{name: "tie half expand", x: "1", r1: "0", r2: "2", mode: RoundHalfExpand, want: "2"},
		{name: "tie half even lower", x: "1", r1: "0", r2: "2", mode: RoundHalfEven, want: "0"},
		{name: "tie half even upper", x: "2", r1: "1", r2: "3", mode: RoundHalfEven, want: "3"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			x := mustParseDecimal(t, tc.x)
			r1 := mustParseDecimal(t, tc.r1)
			r2 := mustParseDecimal(t, tc.r2)
			got := ApplyUnsignedRoundingMode(x, r1, r2, tc.mode)
			if got.String() != tc.want {
				t.Fatalf("ApplyUnsignedRoundingMode() = %q, want %q", got.String(), tc.want)
			}
		})
	}
}

func mustParseDecimal(t *testing.T, s string) Decimal {
	t.Helper()
	d, err := ParseString(s)
	if err != nil {
		t.Fatalf("ParseString(%q): %v", s, err)
	}
	return d
}
