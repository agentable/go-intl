package plural

import (
	"math"
	"testing"
)

func TestCategoryString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		category Category
		want     string
	}{
		{category: Zero, want: "zero"},
		{category: One, want: "one"},
		{category: Two, want: "two"},
		{category: Few, want: "few"},
		{category: Many, want: "many"},
		{category: Other, want: "other"},
		{category: Category(99), want: "other"},
	}
	for _, tc := range tests {
		if got := tc.category.String(); got != tc.want {
			t.Fatalf("Category(%d).String() = %q, want %q", tc.category, got, tc.want)
		}
	}
}

func TestCategoryMarshalText(t *testing.T) {
	t.Parallel()

	got, err := Many.MarshalText()
	if err != nil {
		t.Fatalf("Many.MarshalText() error = %v, want nil", err)
	}
	if string(got) != "many" {
		t.Fatalf("Many.MarshalText() = %q, want many", got)
	}
}

func TestOperandValueHighPrecisionComparisonsAndModulo(t *testing.T) {
	t.Parallel()

	fraction := NewOperandValue("0.00000000000000000001")
	if !fraction.Greater(0) || fraction.Equal(0) {
		t.Fatalf("%s should be greater than zero", fraction)
	}
	if got := fraction.Mod(1).String(); got != "0.00000000000000000001" {
		t.Fatalf("%s %% 1 = %q, want original fraction", fraction, got)
	}

	large := NewOperandValue("9223372036854775807.01")
	if !large.Greater(math.MaxInt64) {
		t.Fatalf("%s should be greater than MaxInt64", large)
	}
	if got := NewOperandValue("100000000000000000003.50").Mod(100).String(); got != "3.50" {
		t.Fatalf("large operand %% 100 = %q, want 3.50", got)
	}
}

func TestOperandValueComparisonsAndModulo(t *testing.T) {
	t.Parallel()

	value := NewOperandValue("-001.2300")
	if got := value.String(); got != "1.2300" {
		t.Fatalf("NewOperandValue(-001.2300).String() = %q, want 1.2300", got)
	}
	if !value.Greater(1) || !value.GreaterOrEqual(1) || !value.Less(2) || !value.LessOrEqual(2) {
		t.Fatalf("comparison helpers failed for %s", value)
	}
	if value.Equal(1) || !value.NotEqual(1) || !value.Between(1, 2) || value.OutsideRange(1, 2) {
		t.Fatalf("range helpers failed for %s", value)
	}
	if got := value.Cmp(-1); got != 1 {
		t.Fatalf("Cmp(-1) = %d, want positive", got)
	}
	if got := value.Mod(1).String(); got != "0.2300" {
		t.Fatalf("Mod(1) = %q, want 0.2300", got)
	}
	if got := value.Mod(0).String(); got != "0" {
		t.Fatalf("Mod(0) = %q, want 0", got)
	}
	if got := NewIntegerOperand(math.MinInt64).String(); got != "9223372036854775808" {
		t.Fatalf("NewIntegerOperand(MinInt64) = %q, want absolute magnitude", got)
	}
	if got := NewIntegerOperand(math.MinInt64).Mod(100).String(); got != "8" {
		t.Fatalf("NewIntegerOperand(MinInt64).Mod(100) = %q, want 8", got)
	}
	if got := NewUnsignedIntegerOperand(^uint64(0)).Mod(100).String(); got != "15" {
		t.Fatalf("NewUnsignedIntegerOperand(MaxUint64).Mod(100) = %q, want 15", got)
	}
}

func TestGetOperands(t *testing.T) {
	t.Parallel()

	ops := GetOperands("-1.2300", 2)
	if ops.N.String() != "1.2300" || ops.I.String() != "1" || ops.V != 4 || ops.W != 2 || ops.F.String() != "2300" || ops.T.String() != "23" || ops.C != 2 || ops.E != 2 {
		t.Fatalf("GetOperands(-1.2300, 2) = %+v", ops)
	}
	integer := GetOperands("42", 0)
	if integer.V != 0 || integer.F.String() != "0" || integer.T.String() != "0" {
		t.Fatalf("GetOperands(42, 0) = %+v, want integer operands", integer)
	}
	fastInteger := GetIntegerOperands(math.MinInt64)
	if fastInteger.N.String() != "9223372036854775808" || fastInteger.I.String() != "9223372036854775808" || fastInteger.V != 0 || fastInteger.F.String() != "0" || fastInteger.T.String() != "0" {
		t.Fatalf("GetIntegerOperands(MinInt64) = %+v, want integer operands", fastInteger)
	}
	fastUnsigned := GetUnsignedIntegerOperands(^uint64(0))
	if fastUnsigned.N.String() != "18446744073709551615" || fastUnsigned.I.String() != "18446744073709551615" || fastUnsigned.V != 0 || fastUnsigned.F.String() != "0" || fastUnsigned.T.String() != "0" {
		t.Fatalf("GetUnsignedIntegerOperands(MaxUint64) = %+v, want integer operands", fastUnsigned)
	}
	if got := (OperandValue{}).String(); got != "0" {
		t.Fatalf("zero OperandValue String() = %q, want 0", got)
	}
	if got := (OperandValue{digits: "not-digits"}).bigInt().String(); got != "0" {
		t.Fatalf("invalid OperandValue bigInt() = %q, want 0", got)
	}
}
