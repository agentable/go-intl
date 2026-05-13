package ecma402types

// MathematicalValue is the ECMA-402 §6.4 normalized representation of a
// numeric input — NaN, +/-Infinity, or a finite value. Concrete implementations
// live in internal/decimal (see SPEC 21); this interface defines the contract
// so internal/ecma402 stays decoupled from the math backend.
type MathematicalValue interface {
	IsNaN() bool
	IsInfinity() bool
	IsNegative() bool
	Sign() int
}
