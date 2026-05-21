package ecma402pr

import (
	"testing"
)

func TestGetOperands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		formatted string
		exponent  int
		want      operandExpectation
	}{
		{name: "trailing zeros", formatted: "1.00", want: operandExpectation{n: "1.00", i: "1", v: 2, f: "0", t: "0"}},
		{name: "fraction without trailing zero", formatted: "12.340", exponent: 3, want: operandExpectation{n: "12.340", i: "12", v: 3, w: 2, f: "340", t: "34", c: 3, e: 3}},
		{name: "large integer is exact", formatted: "100000000000000000001", want: operandExpectation{n: "100000000000000000001", i: "100000000000000000001", f: "0", t: "0"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := GetOperands(tc.formatted, tc.exponent)
			if got.N.String() != tc.want.n || got.I.String() != tc.want.i || got.V != tc.want.v || got.W != tc.want.w || got.F.String() != tc.want.f || got.T.String() != tc.want.t || got.C != tc.want.c || got.E != tc.want.e {
				t.Fatalf("GetOperands() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

type operandExpectation struct {
	n string
	i string
	v int
	w int
	f string
	t string
	c int
	e int
}

func TestOperandValueExactIntegerComparisons(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		mod  int64
		want int64
	}{
		{name: "large ordinal one", in: "10000000000001", mod: 10, want: 1},
		{name: "large ordinal not eleven", in: "10000000000001", mod: 100, want: 1},
		{name: "large french million", in: "1000000000000", mod: 1000000, want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := NewOperandValue(tc.in).Mod(tc.mod); !got.Equal(tc.want) {
				t.Fatalf("%s %% %d = %s, want %d", tc.in, tc.mod, got, tc.want)
			}
		})
	}
}

func TestOperandValueFractionModulo(t *testing.T) {
	t.Parallel()

	remainder := NewOperandValue("103.50").Mod(100)
	if !remainder.Between(3, 10) {
		t.Fatalf("103.50 %% 100 = %s, want between 3 and 10", remainder)
	}
	if remainder.Equal(3) {
		t.Fatalf("103.50 %% 100 = %s, unexpectedly equal to integer 3", remainder)
	}
}
