package decimal

import (
	"math/big"
	"slices"
	"strings"
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

func TestQuantizeToIncrementPreservesArbitraryPrecision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		digits    int
		sourceExp int32
		targetExp int32
	}{
		{name: "one digit", digits: 1, sourceExp: 0, targetExp: 1},
		{name: "one hundred digits", digits: 100, sourceExp: 0, targetExp: 1},
		{name: "one hundred one digits", digits: 101, sourceExp: 0, targetExp: 1},
		{name: "two hundred fifty digits negative exponent", digits: 250, sourceExp: -40, targetExp: -39},
		{name: "one thousand digits positive exponent", digits: 1000, sourceExp: 40, targetExp: 41},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			coefficientText := "5"
			if tc.digits > 1 {
				coefficientText = "1" + strings.Repeat("0", tc.digits-2) + "5"
			}
			coefficient, ok := new(big.Int).SetString(coefficientText, 10)
			if !ok {
				t.Fatalf("SetString(%q) failed", coefficientText)
			}
			input := New(false, coefficient, tc.sourceExp)
			wantCoefficient := new(big.Int).Quo(new(big.Int).Set(coefficient), big.NewInt(10))
			want := New(false, wantCoefficient, tc.targetExp)
			got := QuantizeToIncrement(input, 1, tc.targetExp, RoundTrunc)
			if got.Cmp(want) != 0 {
				t.Fatalf("QuantizeToIncrement(%d digits, source exp %d, target exp %d) = %q, want %q", tc.digits, tc.sourceExp, tc.targetExp, got.String(), want.String())
			}
		})
	}
}

func TestQuantizeToIncrementBoundsWorkForExtremeExponents(t *testing.T) {
	t.Parallel()

	huge := New(false, big.NewInt(1), 1_000_000_000)
	if got := QuantizeToIncrement(huge, 5000, -3, RoundHalfEven); got.Cmp(huge) != 0 {
		t.Fatalf("exact huge exponent changed: got exponent %d", got.inner.Exponent)
	}

	tiny := New(false, big.NewInt(1), -1_000_000_000)
	if got := QuantizeToIncrement(tiny, 1, -3, RoundTrunc); !got.IsZero() {
		t.Fatalf("truncated tiny exponent = exponent %d, want zero", got.inner.Exponent)
	}
	wantExpanded := New(false, big.NewInt(1), -3)
	if got := QuantizeToIncrement(tiny, 1, -3, RoundExpand); got.Cmp(wantExpanded) != 0 {
		t.Fatalf("expanded tiny exponent = exponent %d, want -3", got.inner.Exponent)
	}
}

func TestQuantizeToIncrementRoundingModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode       RoundingMode
		wantTie    string
		wantNonTie string
	}{
		{mode: RoundCeil, wantTie: "15", wantNonTie: "15"},
		{mode: RoundFloor, wantTie: "10", wantNonTie: "10"},
		{mode: RoundExpand, wantTie: "15", wantNonTie: "15"},
		{mode: RoundTrunc, wantTie: "10", wantNonTie: "10"},
		{mode: RoundHalfCeil, wantTie: "15", wantNonTie: "10"},
		{mode: RoundHalfFloor, wantTie: "10", wantNonTie: "10"},
		{mode: RoundHalfExpand, wantTie: "15", wantNonTie: "10"},
		{mode: RoundHalfTrunc, wantTie: "10", wantNonTie: "10"},
		{mode: RoundHalfEven, wantTie: "10", wantNonTie: "10"},
	}
	for _, tc := range tests {
		t.Run(tc.mode.String(), func(t *testing.T) {
			t.Parallel()
			mode := UnsignedRoundingMode(tc.mode, false)
			if got := QuantizeToIncrement(mustParseDecimal(t, "12.5"), 5, 0, mode); got.String() != tc.wantTie {
				t.Fatalf("tie = %q, want %q", got.String(), tc.wantTie)
			}
			if got := QuantizeToIncrement(mustParseDecimal(t, "12.4"), 5, 0, mode); got.String() != tc.wantNonTie {
				t.Fatalf("non-tie = %q, want %q", got.String(), tc.wantNonTie)
			}
		})
	}
}

func TestQuantizeToEachSanctionedIncrement(t *testing.T) {
	t.Parallel()

	for _, increment := range RoundingIncrements() {
		t.Run(big.NewInt(int64(increment)).String(), func(t *testing.T) {
			t.Parallel()
			input := New(false, big.NewInt(int64(increment*25)), -1)
			want := New(false, big.NewInt(int64(increment*2)), 0)
			got := QuantizeToIncrement(input, increment, 0, RoundHalfEven)
			if got.Cmp(want) != 0 {
				t.Fatalf("QuantizeToIncrement(%q, %d) = %q, want %q", input.String(), increment, got.String(), want.String())
			}
		})
	}
}

func FuzzQuantizeToIncrement(f *testing.F) {
	f.Add("12345", int8(-2), uint8(0), int8(-1), uint8(6))
	f.Add("1"+strings.Repeat("0", 99)+"5", int8(0), uint8(3), int8(0), uint8(3))
	f.Add("9"+strings.Repeat("9", 999), int8(20), uint8(14), int8(21), uint8(8))

	f.Fuzz(func(t *testing.T, coefficientText string, sourceExp int8, incrementIndex uint8, targetExp int8, modeValue uint8) {
		if coefficientText == "" || len(coefficientText) > 1000 || strings.Trim(coefficientText, "0123456789") != "" {
			return
		}
		coefficient, ok := new(big.Int).SetString(coefficientText, 10)
		if !ok {
			return
		}
		increments := RoundingIncrements()
		increment := increments[int(incrementIndex)%len(increments)]
		mode := UnsignedRoundingMode(roundingModeNames[int(modeValue)%len(roundingModeNames)], false)
		input := New(false, coefficient, int32(sourceExp))
		got := QuantizeToIncrement(input, increment, int32(targetExp), mode)

		want, lower, upper := referenceQuantize(coefficient, int32(sourceExp), increment, int32(targetExp), mode)
		gotRat, ok := new(big.Rat).SetString(got.String())
		if !ok {
			t.Fatalf("SetString(%q) failed", got.String())
		}
		if gotRat.Cmp(want) != 0 {
			t.Fatalf("QuantizeToIncrement(%q, %d, %d, %s) = %q, want %s", input.String(), increment, targetExp, mode, got.String(), want.RatString())
		}
		if gotRat.Cmp(lower) < 0 || gotRat.Cmp(upper) > 0 {
			t.Fatalf("result %s outside adjacent values [%s, %s]", gotRat.RatString(), lower.RatString(), upper.RatString())
		}
		step := decimalRat(big.NewInt(int64(increment)), int32(targetExp))
		multiple := new(big.Rat).Quo(gotRat, step)
		if !multiple.IsInt() {
			t.Fatalf("result %s is not a multiple of step %s", gotRat.RatString(), step.RatString())
		}
	})
}

func referenceQuantize(coefficient *big.Int, sourceExp int32, increment int, targetExp int32, mode RoundingMode) (want, lower, upper *big.Rat) {
	x := decimalRat(coefficient, sourceExp)
	step := decimalRat(big.NewInt(int64(increment)), targetExp)
	quotient := new(big.Rat).Quo(x, step)
	lowerCount := new(big.Int).Quo(quotient.Num(), quotient.Denom())
	lower = new(big.Rat).Mul(new(big.Rat).SetInt(lowerCount), step)
	if quotient.IsInt() {
		return lower, lower, lower
	}
	upperCount := new(big.Int).Add(lowerCount, big.NewInt(1))
	upper = new(big.Rat).Mul(new(big.Rat).SetInt(upperCount), step)
	switch mode {
	case RoundFloor, RoundTrunc:
		return lower, lower, upper
	case RoundCeil, RoundExpand:
		return upper, lower, upper
	case RoundHalfCeil, RoundHalfFloor, RoundHalfExpand, RoundHalfTrunc, RoundHalfEven:
	}
	lowerDistance := new(big.Rat).Sub(x, lower)
	upperDistance := new(big.Rat).Sub(upper, x)
	switch lowerDistance.Cmp(upperDistance) {
	case -1:
		return lower, lower, upper
	case 1:
		return upper, lower, upper
	}
	switch mode {
	case RoundHalfFloor, RoundHalfTrunc:
		return lower, lower, upper
	case RoundHalfCeil, RoundHalfExpand:
		return upper, lower, upper
	case RoundHalfEven:
		if lowerCount.Bit(0) == 0 {
			return lower, lower, upper
		}
		return upper, lower, upper
	case RoundCeil, RoundFloor, RoundExpand, RoundTrunc:
		return lower, lower, upper
	}
	return lower, lower, upper
}

func decimalRat(coefficient *big.Int, exp int32) *big.Rat {
	value := new(big.Rat).SetInt(coefficient)
	scaleExp := int64(exp)
	if scaleExp < 0 {
		scaleExp = -scaleExp
	}
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(scaleExp), nil)
	if exp >= 0 {
		return value.Mul(value, new(big.Rat).SetInt(scale))
	}
	return value.Quo(value, new(big.Rat).SetInt(scale))
}
