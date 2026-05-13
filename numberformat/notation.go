package numberformat

import (
	"strconv"
	"strings"

	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
)

func (f *NumberFormat) formatCompact(d decimal.Decimal) []Part {
	symbols := f.symbols()
	scaled, pattern, exp := f.resolveCompactPattern(d)
	result := f.formatFiniteResult(scaled)
	formatted := result.Formatted
	pattern = f.compactPatternForFormatted(pattern, formatted, exp)
	if f.useGrouping(formatted) {
		formatted = groupDecimal(formatted, f.grouping)
	}
	parts := partitionDecimal(formatted, symbols)
	parts = f.applySignDisplay(parts, d.Negative())
	return interpolateCompactPattern(pattern, parts)
}

func (f *NumberFormat) resolveCompactPattern(d decimal.Decimal) (decimal.Decimal, string, int) {
	magnitude, err := decimal.Log10Floor(decimal.Abs(d))
	if err != nil || magnitude < 3 {
		return d, "", 0
	}
	exp := int(magnitude)
	for exp >= 3 {
		pattern := f.compactPattern(exp, "other")
		if pattern != "" {
			return decimal.Scale10(d, -int32(exp)), pattern, exp // #nosec G115 -- exp came from decimal.Log10Floor int32.
		}
		exp--
	}
	return d, "", 0
}

func (f *NumberFormat) compactPatternForFormatted(pattern, formatted string, exp int) string {
	if pattern == "" {
		return ""
	}
	plural := pluralCategoryWithExponent(f.resolved.Locale.Tag().String(), strings.TrimPrefix(formatted, "-"), exp)
	if pluralPattern := f.compactPattern(exp, plural); pluralPattern != "" {
		return pluralPattern
	}
	return pattern
}

func (f *NumberFormat) roundedForRange(d decimal.Decimal) (decimal.Decimal, bool) {
	if !d.IsFinite() {
		return decimal.Decimal{}, false
	}
	if f.resolved.Style == "percent" {
		d = decimal.MulInt(d, 100)
	}
	switch f.resolved.Notation {
	case CompactNotation:
		d, _, _ = f.resolveCompactPattern(d)
	case ScientificNotation, EngineeringNotation:
		exponent, ok := f.scientificExponent(d)
		if !ok {
			return decimal.Decimal{}, false
		}
		d = decimal.Scale10(d, -int32(exponent)) // #nosec G115 -- exponent came from decimal.Log10Floor int32.
	case StandardNotation:
	}
	return f.formatFiniteResult(d).Rounded, true
}

func (f *NumberFormat) compactPattern(exp int, plural string) string {
	return f.cldrLoc.CompactPattern(f.resolved.NumberingSystem, string(f.resolved.CompactDisplay), exp, plural)
}

func interpolateCompactPattern(pattern string, parts []Part) []Part {
	start, end := numberPatternBounds(pattern)
	if start < 0 {
		return parts
	}
	out := make([]Part, 0, len(parts)+strings.Count(pattern, " ")+2)
	out = appendPatternTextParts(out, pattern[:start], PartCompact)
	out = append(out, parts...)
	out = appendPatternTextParts(out, pattern[end:], PartCompact)
	return out
}

func (f *NumberFormat) formatScientific(d decimal.Decimal) []Part {
	symbols := f.symbols()
	exponent, ok := f.scientificExponent(d)
	if !ok {
		return []Part{{Type: PartNaN, Value: symbols.NaN}}
	}
	d = decimal.Scale10(d, -int32(exponent)) // #nosec G115 -- exponent came from decimal.Log10Floor int32.
	formatted := f.formatFinite(d)
	if f.useGrouping(formatted) {
		formatted = groupDecimal(formatted, f.grouping)
	}
	parts := partitionDecimal(formatted, symbols)
	parts = f.applySignDisplay(parts, d.Negative())
	parts = append(parts, Part{Type: PartExponentSeparator, Value: symbols.Exponential})
	if exponent < 0 {
		parts = append(parts, Part{Type: PartExponentMinusSign, Value: symbols.Minus})
		exponent = -exponent
	}
	parts = append(parts, Part{Type: PartExponentInteger, Value: ecma402.LocalizeDigits(strconv.Itoa(exponent), f.resolved.NumberingSystem)})
	return parts
}

func (f *NumberFormat) scientificExponent(d decimal.Decimal) (int, bool) {
	abs := decimal.Abs(d)
	if abs.IsZero() {
		return 0, true
	}
	magnitude, err := decimal.Log10Floor(abs)
	if err != nil {
		return 0, false
	}
	exponent := int(magnitude)
	if f.resolved.Notation == "engineering" {
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
