package plural

import (
	"math/big"
	"strconv"
	"strings"
)

type Category uint8

const (
	Zero Category = iota
	One
	Two
	Few
	Many
	Other
)

func (c Category) String() string {
	switch c {
	case Zero:
		return "zero"
	case One:
		return "one"
	case Two:
		return "two"
	case Few:
		return "few"
	case Many:
		return "many"
	case Other:
		return "other"
	}
	return "other"
}

func (c Category) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

type OperandValue struct {
	digits  string
	scale   int
	small   uint64
	isSmall bool
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
	return newSmallOperand(absInt64(n), 0)
}

func NewUnsignedIntegerOperand(n uint64) OperandValue {
	return newSmallOperand(n, 0)
}

func (v OperandValue) Equal(n int64) bool {
	return v.Cmp(n) == 0
}

func (v OperandValue) NotEqual(n int64) bool {
	return v.Cmp(n) != 0
}

func (v OperandValue) Less(n int64) bool {
	return v.Cmp(n) < 0
}

func (v OperandValue) LessOrEqual(n int64) bool {
	return v.Cmp(n) <= 0
}

func (v OperandValue) Greater(n int64) bool {
	return v.Cmp(n) > 0
}

func (v OperandValue) GreaterOrEqual(n int64) bool {
	return v.Cmp(n) >= 0
}

func (v OperandValue) Between(start, end int64) bool {
	return v.GreaterOrEqual(start) && v.LessOrEqual(end)
}

func (v OperandValue) OutsideRange(start, end int64) bool {
	return v.Less(start) || v.Greater(end)
}

func (v OperandValue) Cmp(n int64) int {
	if n < 0 {
		return 1
	}
	if cmp, ok := v.cmpSmall(n); ok {
		return cmp
	}
	left := v.bigInt()
	right := new(big.Int).SetInt64(n)
	if v.scale > 0 {
		right.Mul(right, pow10(v.scale))
	}
	return left.Cmp(right)
}

func (v OperandValue) Mod(mod int64) OperandValue {
	if mod <= 0 || v.isZero() {
		return newSmallOperand(0, 0)
	}
	if remainder, ok := v.modSmall(mod); ok {
		return remainder
	}
	divisor := new(big.Int).SetInt64(mod)
	if v.scale > 0 {
		divisor.Mul(divisor, pow10(v.scale))
	}
	remainder := new(big.Int).Mod(v.bigInt(), divisor)
	return newOperandDigits(remainder.String(), v.scale)
}

func (v OperandValue) String() string {
	digits := v.digits
	if digits == "" && v.isSmall {
		digits = strconv.FormatUint(v.small, 10)
	}
	if digits == "" {
		return "0"
	}
	if v.scale == 0 {
		return digits
	}
	padded := digits
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

func GetIntegerOperands(n int64) OperandsRecord {
	return integerOperands(NewIntegerOperand(n))
}

func GetUnsignedIntegerOperands(n uint64) OperandsRecord {
	return integerOperands(NewUnsignedIntegerOperand(n))
}

func integerOperands(value OperandValue) OperandsRecord {
	zero := newSmallOperand(0, 0)
	return OperandsRecord{
		N: value,
		I: value,
		F: zero,
		T: zero,
	}
}

func newOperandDigits(digits string, scale int) OperandValue {
	digits = strings.TrimLeft(digits, "0")
	if digits == "" {
		digits = "0"
	}
	small, err := strconv.ParseUint(digits, 10, 64)
	return OperandValue{
		digits:  digits,
		scale:   scale,
		small:   small,
		isSmall: err == nil,
	}
}

func newSmallOperand(n uint64, scale int) OperandValue {
	return OperandValue{scale: scale, small: n, isSmall: true}
}

func (v OperandValue) bigInt() *big.Int {
	if v.isSmall {
		return new(big.Int).SetUint64(v.small)
	}
	n, ok := new(big.Int).SetString(v.digits, 10)
	if !ok {
		return new(big.Int)
	}
	return n
}

func (v OperandValue) isZero() bool {
	if v.isSmall {
		return v.small == 0
	}
	return v.digits == "" || v.digits == "0"
}

func pow10(scale int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
}

func (v OperandValue) cmpSmall(n int64) (int, bool) {
	if !v.isSmall {
		return 0, false
	}
	if n < 0 {
		return 1, true
	}
	right := uint64(n) // #nosec G115 -- guarded above.
	if v.scale > 0 {
		factor, ok := pow10Uint64(v.scale)
		if !ok {
			return 0, false
		}
		if right > maxUint64/factor {
			return -1, true
		}
		right *= factor
	}
	return cmpUint64(v.small, right), true
}

func (v OperandValue) modSmall(mod int64) (OperandValue, bool) {
	if !v.isSmall {
		return OperandValue{}, false
	}
	if mod <= 0 {
		return OperandValue{}, false
	}
	divisor := uint64(mod) // #nosec G115 -- guarded above.
	if v.scale > 0 {
		factor, ok := pow10Uint64(v.scale)
		if !ok || divisor > maxUint64/factor {
			return OperandValue{}, false
		}
		divisor *= factor
	}
	return newSmallOperand(v.small%divisor, v.scale), true
}

func absInt64(n int64) uint64 {
	if n >= 0 {
		return uint64(n)
	}
	return uint64(-(n + 1)) + 1 // #nosec G115 -- expression is non-negative and handles MinInt64.
}

func cmpUint64(left, right uint64) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

const maxUint64 = ^uint64(0)

func pow10Uint64(scale int) (uint64, bool) {
	if scale < 0 {
		return 0, false
	}
	out := uint64(1)
	for range scale {
		if out > maxUint64/10 {
			return 0, false
		}
		out *= 10
	}
	return out, true
}
