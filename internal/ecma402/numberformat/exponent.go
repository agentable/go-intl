package ecma402nf

import "github.com/agentable/go-intl/internal/decimal"

// exponentForMagnitude maps a base-10 magnitude to a notation exponent. ok=false
// means the notation does not apply at that magnitude (e.g. compact below its
// smallest generated key), in which case the caller falls back to standard.
type exponentForMagnitude func(magnitude int) (exponent int, ok bool)

// computeExponent is the single engine behind scientific, engineering, and
// compact notation. It scales d into its notation mantissa, rounds it with the
// resolved digit options, and applies the ECMA-402 carry recheck: when rounding
// pushes the mantissa up a magnitude (999 -> 1000), the exponent is re-derived
// from magnitude+1. Mirrors FormatJS ComputeExponent
// (.references/formatjs/packages/ecma402-abstract/NumberFormat/ComputeExponent.ts).
func computeExponent(d decimal.Decimal, digitOptions DigitOptions, toExponent exponentForMagnitude) (exponent, magnitude int, ok bool) {
	abs := decimal.Abs(d)
	if abs.IsZero() {
		exponent, ok = toExponent(0)
		return exponent, 0, ok
	}
	mag, err := decimal.Log10Floor(abs)
	if err != nil {
		return 0, 0, false
	}
	magnitude = int(mag)
	exponent, ok = toExponent(magnitude)
	if !ok {
		return 0, 0, false
	}
	scaled := decimal.Scale10(d, -int32(exponent)) // #nosec G115 -- notation exponents come from Log10Floor int32 / small generated keys.
	rounded := decimal.Abs(FormatNumericToString(scaled, digitOptions).Rounded)
	if rounded.IsZero() {
		return exponent, magnitude, true
	}
	roundedMagnitude, err := decimal.Log10Floor(rounded)
	if err != nil || int(roundedMagnitude) == magnitude-exponent {
		return exponent, magnitude, true
	}
	if nextExponent, nextOK := toExponent(magnitude + 1); nextOK {
		return nextExponent, magnitude + 1, true
	}
	return exponent, magnitude, true
}

// ScientificExponent returns the decimal exponent used by scientific and
// engineering notation, including the post-rounding carry recheck.
func ScientificExponent(d decimal.Decimal, digitOptions DigitOptions, engineering bool) (int, bool) {
	exponent, _, ok := computeExponent(d, digitOptions, scientificExponentForMagnitude(engineering))
	return exponent, ok
}

func scientificExponentForMagnitude(engineering bool) exponentForMagnitude {
	return func(magnitude int) (int, bool) {
		if engineering {
			return magnitude - positiveMod(magnitude, 3), true
		}
		return magnitude, true
	}
}

func positiveMod(n, mod int) int {
	out := n % mod
	if out < 0 {
		out += mod
	}
	return out
}
