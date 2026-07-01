package ecma402nf

import "github.com/agentable/go-intl/internal/decimal"

// ScientificExponent returns the decimal exponent used by scientific and
// engineering notation.
func ScientificExponent(d decimal.Decimal, engineering bool) (int, bool) {
	abs := decimal.Abs(d)
	if abs.IsZero() {
		return 0, true
	}
	magnitude, err := decimal.Log10Floor(abs)
	if err != nil {
		return 0, false
	}
	exponent := int(magnitude)
	if engineering {
		exponent -= positiveMod(exponent, 3)
	}
	return exponent, true
}

func positiveMod(n, mod int) int {
	out := n % mod
	if out < 0 {
		out += mod
	}
	return out
}
