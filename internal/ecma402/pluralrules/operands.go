package ecma402pr

import (
	"math/big"
	"strings"
)

type OperandValue struct {
	digits string
	scale  int
}

type OperandsRecord struct {
	N OperandValue
	I OperandValue
	V int
	W int
	F OperandValue
	T OperandValue
	C int
	E int
}

func NewOperandValue(formatted string) OperandValue {
	formatted = strings.TrimPrefix(formatted, "-")
	integerPart, fractionPart, hasFraction := strings.Cut(formatted, ".")
	if !hasFraction {
		return newOperandDigits(integerPart, 0)
	}
	return newOperandDigits(integerPart+fractionPart, len(fractionPart))
}

func NewIntegerOperand(n int64) OperandValue {
	digits := new(big.Int).SetInt64(n)
	digits.Abs(digits)
	return newOperandDigits(digits.String(), 0)
}

func (v OperandValue) EqualInt(n int64) bool {
	return v.CmpInt(n) == 0
}

func (v OperandValue) NotEqualInt(n int64) bool {
	return v.CmpInt(n) != 0
}

func (v OperandValue) LessInt(n int64) bool {
	return v.CmpInt(n) < 0
}

func (v OperandValue) LessOrEqualInt(n int64) bool {
	return v.CmpInt(n) <= 0
}

func (v OperandValue) GreaterInt(n int64) bool {
	return v.CmpInt(n) > 0
}

func (v OperandValue) GreaterOrEqualInt(n int64) bool {
	return v.CmpInt(n) >= 0
}

func (v OperandValue) BetweenInt(start, end int64) bool {
	return v.GreaterOrEqualInt(start) && v.LessOrEqualInt(end)
}

func (v OperandValue) OutsideIntRange(start, end int64) bool {
	return v.LessInt(start) || v.GreaterInt(end)
}

func (v OperandValue) CmpInt(n int64) int {
	if n < 0 {
		return 1
	}
	left := v.bigInt()
	right := new(big.Int).SetInt64(n)
	if v.scale > 0 {
		right.Mul(right, pow10(v.scale))
	}
	return left.Cmp(right)
}

func (v OperandValue) ModInt(mod int64) OperandValue {
	if mod <= 0 || v.isZero() {
		return OperandValue{digits: "0"}
	}
	divisor := new(big.Int).SetInt64(mod)
	if v.scale > 0 {
		divisor.Mul(divisor, pow10(v.scale))
	}
	remainder := new(big.Int).Mod(v.bigInt(), divisor)
	return newOperandDigits(remainder.String(), v.scale)
}

func (v OperandValue) String() string {
	if v.digits == "" {
		return "0"
	}
	if v.scale == 0 {
		return v.digits
	}
	padded := v.digits
	if len(padded) <= v.scale {
		padded = strings.Repeat("0", v.scale-len(padded)+1) + padded
	}
	cut := len(padded) - v.scale
	return padded[:cut] + "." + padded[cut:]
}

func GetOperands(formatted string, exponent int) OperandsRecord {
	integerPart, fractionPart, hasFraction := strings.Cut(formatted, ".")
	integerPart = strings.TrimPrefix(integerPart, "-")
	ops := OperandsRecord{
		N: NewOperandValue(formatted),
		I: newOperandDigits(integerPart, 0),
		F: newOperandDigits("0", 0),
		T: newOperandDigits("0", 0),
		C: exponent,
		E: exponent,
	}
	if !hasFraction {
		return ops
	}
	ops.V = len(fractionPart)
	ops.F = newOperandDigits(fractionPart, 0)
	trimmed := strings.TrimRight(fractionPart, "0")
	ops.W = len(trimmed)
	if trimmed != "" {
		ops.T = newOperandDigits(trimmed, 0)
	}
	return ops
}

func newOperandDigits(digits string, scale int) OperandValue {
	digits = strings.TrimLeft(digits, "0")
	if digits == "" {
		digits = "0"
	}
	return OperandValue{digits: digits, scale: scale}
}

func (v OperandValue) bigInt() *big.Int {
	n, ok := new(big.Int).SetString(v.digits, 10)
	if !ok {
		return new(big.Int)
	}
	return n
}

func (v OperandValue) isZero() bool {
	return v.digits == "0"
}

func pow10(scale int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
}
