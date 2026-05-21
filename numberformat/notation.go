package numberformat

import (
	"strconv"
	"strings"

	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
)

func (f *NumberFormat) formatCompactAppend(parts []Part, d decimal.Decimal) ([]Part, decimal.Decimal) {
	symbols := f.symbols()
	scaled, entry := f.resolveCompactPattern(d)
	result := f.formatFiniteResult(scaled)
	formatted := result.Formatted
	pattern := f.compactPatternForFormatted(entry, formatted)
	if f.useGrouping(formatted) {
		formatted = groupDecimal(formatted, f.grouping)
	}
	parts = appendDecimalParts(parts, formatted, symbols)
	parts = f.applySignDisplay(parts, d.Negative())
	return pattern.append(parts), result.Rounded
}

func (f *NumberFormat) resolveCompactPattern(d decimal.Decimal) (decimal.Decimal, compactPatternEntry) {
	magnitude, err := decimal.Log10Floor(decimal.Abs(d))
	if err != nil || magnitude < minCompactExponent {
		return d, compactPatternEntry{}
	}
	entry, ok := f.compact.patternForMagnitude(int(magnitude))
	if !ok {
		return d, compactPatternEntry{}
	}
	scaled := decimal.Scale10(d, -int32(entry.exponent)) // #nosec G115 -- compact exponents are small generated data keys.
	result := f.formatFiniteResult(scaled)
	rounded := decimal.Abs(result.Rounded)
	if rounded.IsZero() {
		return scaled, entry
	}
	roundedMagnitude, err := decimal.Log10Floor(rounded)
	if err != nil {
		return scaled, entry
	}
	if int(roundedMagnitude) == int(magnitude)-entry.exponent {
		return scaled, entry
	}
	next, ok := f.compact.patternForMagnitude(int(magnitude) + 1)
	if !ok {
		return scaled, entry
	}
	return decimal.Scale10(d, -int32(next.exponent)), next // #nosec G115 -- compact exponents are small generated data keys.
}

func (f *NumberFormat) compactPatternForFormatted(entry compactPatternEntry, formatted string) compactAffixPattern {
	if !entry.other.set {
		return compactAffixPattern{}
	}
	plural := pluralCategoryWithExponent(f.resolved.Locale.Tag().String(), strings.TrimPrefix(formatted, "-"), entry.exponent)
	return entry.pattern(plural)
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
		d, _ = f.resolveCompactPattern(d)
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

func (f *NumberFormat) formatScientificAppend(parts []Part, d decimal.Decimal) ([]Part, decimal.Decimal) {
	symbols := f.symbols()
	exponent, ok := f.scientificExponent(d)
	if !ok {
		return append(parts, Part{Type: PartNaN, Value: symbols.NaN}), decimal.NaNValue
	}
	scaled := decimal.Scale10(d, -int32(exponent)) // #nosec G115 -- exponent came from decimal.Log10Floor int32.
	result := f.formatFiniteResult(scaled)
	formatted := result.Formatted
	if f.useGrouping(formatted) {
		formatted = groupDecimal(formatted, f.grouping)
	}
	parts = appendDecimalParts(parts, formatted, symbols)
	parts = f.applySignDisplay(parts, d.Negative())
	parts = append(parts, Part{Type: PartExponentSeparator, Value: symbols.Exponential})
	if exponent < 0 {
		parts = append(parts, Part{Type: PartExponentMinusSign, Value: symbols.Minus})
		exponent = -exponent
	}
	parts = append(parts, Part{Type: PartExponentInteger, Value: ecma402.LocalizeDigits(strconv.Itoa(exponent), f.resolved.NumberingSystem)})
	return parts, result.Rounded
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

const (
	minCompactExponent = 3
	maxCompactExponent = 14
)

type compactPatternSet struct {
	entries []compactPatternEntry
}

type compactPatternEntry struct {
	magnitude, exponent              int
	zero, one, two, few, many, other compactAffixPattern
}

func compactPatternsForNumberFormat(loc cldrnumber.Locale, opts ResolvedOptions) compactPatternSet {
	if opts.Notation != CompactNotation {
		return compactPatternSet{}
	}
	display := string(opts.CompactDisplay)
	entries := make([]compactPatternEntry, 0, maxCompactExponent-minCompactExponent+1)
	for exp := maxCompactExponent; exp >= minCompactExponent; exp-- {
		other := loc.CompactPattern(opts.NumberingSystem, display, exp, "other")
		if other == "" {
			continue
		}
		entries = append(entries, compactPatternEntry{
			magnitude: exp,
			exponent:  compactExponentForPattern(exp, other),
			zero:      compileCompactAffixPattern(loc.CompactPattern(opts.NumberingSystem, display, exp, "zero")),
			one:       compileCompactAffixPattern(loc.CompactPattern(opts.NumberingSystem, display, exp, "one")),
			two:       compileCompactAffixPattern(loc.CompactPattern(opts.NumberingSystem, display, exp, "two")),
			few:       compileCompactAffixPattern(loc.CompactPattern(opts.NumberingSystem, display, exp, "few")),
			many:      compileCompactAffixPattern(loc.CompactPattern(opts.NumberingSystem, display, exp, "many")),
			other:     compileCompactAffixPattern(other),
		})
	}
	return compactPatternSet{entries: entries}
}

func (p compactPatternSet) patternForMagnitude(magnitude int) (compactPatternEntry, bool) {
	for _, entry := range p.entries {
		if magnitude >= entry.magnitude {
			return entry, true
		}
	}
	return compactPatternEntry{}, false
}

func compactExponentForPattern(magnitude int, pattern string) int {
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
	start, end := numberPatternBounds(pattern)
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

func (p compactPatternEntry) pattern(plural string) compactAffixPattern {
	switch plural {
	case "zero":
		if p.zero.set {
			return p.zero
		}
	case "one":
		if p.one.set {
			return p.one
		}
	case "two":
		if p.two.set {
			return p.two
		}
	case "few":
		if p.few.set {
			return p.few
		}
	case "many":
		if p.many.set {
			return p.many
		}
	}
	return p.other
}

type compactAffixPattern struct {
	prefix []Part
	suffix []Part
	set    bool
}

func compileCompactAffixPattern(pattern string) compactAffixPattern {
	if pattern == "" {
		return compactAffixPattern{}
	}
	start, end := numberPatternBounds(pattern)
	if start < 0 {
		return compactAffixPattern{set: true}
	}
	return compactAffixPattern{
		prefix: appendPatternTextParts(nil, pattern[:start], PartCompact),
		suffix: appendPatternTextParts(nil, pattern[end:], PartCompact),
		set:    true,
	}
}

func (p compactAffixPattern) append(parts []Part) []Part {
	if !p.set || len(p.prefix)+len(p.suffix) == 0 {
		return parts
	}
	if len(p.prefix) == 0 {
		return append(parts, p.suffix...)
	}
	out := make([]Part, 0, len(p.prefix)+len(parts)+len(p.suffix))
	out = append(out, p.prefix...)
	out = append(out, parts...)
	return append(out, p.suffix...)
}
