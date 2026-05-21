package ecma402

// MathematicalValue is the ECMA-402 normalized representation of a numeric
// input: NaN, infinity, or a finite value.
type MathematicalValue interface {
	IsNaN() bool
	IsInfinity() bool
	IsNegative() bool
	Sign() int
}

// Part is the unit of output from PartitionPattern and formatter partition
// algorithms.
type Part struct {
	Type  string
	Value string
}

// Pattern is the slice form returned by PartitionPattern.
type Pattern = []Part
