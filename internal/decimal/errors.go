package decimal

import "errors"

var (
	// ErrInvalidDecimal classifies malformed or unsupported decimal input.
	ErrInvalidDecimal = errors.New("decimal: invalid numeric literal")
	// ErrInvalidRoundingIncrement classifies increments outside ECMA-402's valid set.
	ErrInvalidRoundingIncrement = errors.New("decimal: invalid rounding increment")
	// ErrNaNComparison classifies comparisons where either operand is NaN.
	ErrNaNComparison = errors.New("decimal: NaN in comparison")
	// ErrLog10Domain classifies log10 inputs outside the finite, non-zero domain.
	ErrLog10Domain = errors.New("decimal: log10 of zero or non-finite")
)
