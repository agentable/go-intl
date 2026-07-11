package ecma402nf

import (
	"strings"

	"github.com/agentable/go-intl/internal/decimal"
)

const (
	MinCompactMagnitude = 3
	MaxCompactMagnitude = 14
)

type CompactExponentLookup func(magnitude int) (exponent int, ok bool)

// ResolveCompactMagnitude scales d into its compact mantissa via the shared
// computeExponent engine (which owns the post-rounding carry recheck) and returns
// the scaled value, its magnitude, the compact exponent, and whether compact
// notation applies. Below MinCompactMagnitude it does not apply.
func ResolveCompactMagnitude(d decimal.Decimal, digitOptions DigitOptions, lookup CompactExponentLookup) (decimal.Decimal, int, int, bool) {
	exponent, magnitude, ok := computeExponent(d, digitOptions, compactExponentForMagnitude(lookup))
	if !ok {
		return d, 0, 0, false
	}
	return decimal.Scale10(d, -int32(exponent)), magnitude, exponent, true // #nosec G115 -- compact exponents are small generated data keys.
}

func compactExponentForMagnitude(lookup CompactExponentLookup) exponentForMagnitude {
	return func(magnitude int) (int, bool) {
		if magnitude < MinCompactMagnitude || lookup == nil {
			return 0, false
		}
		return lookup(magnitude)
	}
}

func CompactExponentForPattern(magnitude int, pattern string) int {
	if pattern == "0" {
		return 0
	}
	zeroCount := compactPatternZeroCount(pattern)
	if zeroCount == 0 {
		return magnitude
	}
	return magnitude + 1 - zeroCount
}

func compactPatternZeroCount(pattern string) int {
	start, end := compactPatternNumberBounds(pattern)
	if start < 0 {
		return 0
	}
	count := 0
	for _, r := range pattern[start:end] {
		if r == '0' {
			count++
			continue
		}
		if count > 0 {
			return count
		}
	}
	return count
}

func compactPatternNumberBounds(pattern string) (int, int) {
	start := strings.IndexAny(pattern, "#0")
	if start < 0 {
		return -1, -1
	}
	end := start
	for end < len(pattern) && strings.ContainsRune("#0,.", rune(pattern[end])) {
		end++
	}
	return start, end
}
