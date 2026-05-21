package decimal

import (
	"fmt"
	"math"
)

func Log10Floor(x Decimal) (int32, error) {
	if !x.IsFinite() || x.Sign() <= 0 {
		return 0, fmt.Errorf("decimal: log10 domain for %s: %w", x.String(), ErrLog10Domain)
	}
	digits := x.inner.NumDigits() - 1
	if digits > math.MaxInt32 || digits < math.MinInt32 {
		return 0, fmt.Errorf("decimal: log10 overflow for %s: %w", x.String(), ErrInvalidDecimal)
	}
	return int32(digits) + x.inner.Exponent, nil
}
