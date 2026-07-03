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

func ResolveCompactMagnitude(d decimal.Decimal, digitOptions DigitOptions, lookup CompactExponentLookup) (decimal.Decimal, int, int, bool) {
	magnitude, err := decimal.Log10Floor(decimal.Abs(d))
	if err != nil || magnitude < MinCompactMagnitude || lookup == nil {
		return d, 0, 0, false
	}
	exponent, ok := lookup(int(magnitude))
	if !ok {
		return d, 0, 0, false
	}
	scaled := decimal.Scale10(d, -int32(exponent)) // #nosec G115 -- compact exponents are small generated data keys.
	result := FormatNumericToString(scaled, digitOptions)
	rounded := decimal.Abs(result.Rounded)
	if rounded.IsZero() {
		return scaled, int(magnitude), exponent, true
	}
	roundedMagnitude, err := decimal.Log10Floor(rounded)
	if err != nil {
		return scaled, int(magnitude), exponent, true
	}
	if int(roundedMagnitude) == int(magnitude)-exponent {
		return scaled, int(magnitude), exponent, true
	}
	nextMagnitude := int(magnitude) + 1
	nextExponent, ok := lookup(nextMagnitude)
	if !ok {
		return scaled, int(magnitude), exponent, true
	}
	return decimal.Scale10(d, -int32(nextExponent)), nextMagnitude, nextExponent, true // #nosec G115 -- compact exponents are small generated data keys.
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
