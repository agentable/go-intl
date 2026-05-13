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
		want      OperandsRecord
	}{
		{name: "trailing zeros", formatted: "1.00", want: OperandsRecord{N: NewOperandValue("1.00"), I: NewIntegerOperand(1), V: 2, W: 0, F: NewIntegerOperand(0), T: NewIntegerOperand(0)}},
		{name: "fraction without trailing zero", formatted: "12.340", exponent: 3, want: OperandsRecord{N: NewOperandValue("12.340"), I: NewIntegerOperand(12), V: 3, W: 2, F: NewIntegerOperand(340), T: NewIntegerOperand(34), C: 3, E: 3}},
		{name: "large integer is exact", formatted: "100000000000000000001", want: OperandsRecord{N: NewOperandValue("100000000000000000001"), I: NewOperandValue("100000000000000000001"), F: NewIntegerOperand(0), T: NewIntegerOperand(0)}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := GetOperands(tc.formatted, tc.exponent)
			if got != tc.want {
				t.Fatalf("GetOperands() = %#v, want %#v", got, tc.want)
			}
		})
	}
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
			if got := NewOperandValue(tc.in).ModInt(tc.mod); !got.EqualInt(tc.want) {
				t.Fatalf("%s %% %d = %s, want %d", tc.in, tc.mod, got, tc.want)
			}
		})
	}
}

func TestOperandValueFractionModulo(t *testing.T) {
	t.Parallel()

	remainder := NewOperandValue("103.50").ModInt(100)
	if !remainder.BetweenInt(3, 10) {
		t.Fatalf("103.50 %% 100 = %s, want between 3 and 10", remainder)
	}
	if remainder.EqualInt(3) {
		t.Fatalf("103.50 %% 100 = %s, unexpectedly equal to integer 3", remainder)
	}
}
