package numberformat

import (
	"strconv"
	"strings"

	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
	ecma402pr "github.com/agentable/go-intl/internal/ecma402/pluralrules"
	"github.com/agentable/go-intl/internal/numbering"
)

func formatCompactAppend(parts []Part, d decimal.Decimal, state *decimalFormatState) ([]Part, decimal.Decimal) {
	symbols := state.symbols
	resolved := state.resolved
	signDisplay := resolved.SignDisplay
	grouping := state.grouping
	digitOptions := state.digitOptions
	cardinalRule := state.cardinalRule
	compact := state.compact
	scaled, entry := resolveCompactPattern(d, digitOptions, compact)
	result := ecma402nf.FormatNumericToString(scaled, digitOptions)
	formatted := result.Formatted
	pattern := compactPatternForFormatted(entry, formatted, cardinalRule)
	if shouldUseGrouping(resolved.UseGrouping, formatted) {
		formatted = groupDecimal(formatted, grouping)
	}
	parts = appendDecimalParts(parts, formatted, symbols)
	parts = applySignDisplay(parts, d.Negative(), signDisplay, symbols)
	return pattern.append(parts), result.Rounded
}

func formatCompactText(d decimal.Decimal, state *decimalFormatState) (string, decimal.Decimal) {
	symbols := state.symbols
	resolved := state.resolved
	signDisplay := resolved.SignDisplay
	grouping := state.grouping
	digitOptions := state.digitOptions
	cardinalRule := state.cardinalRule
	compact := state.compact
	scaled, entry := resolveCompactPattern(d, digitOptions, compact)
	result := ecma402nf.FormatNumericToString(scaled, digitOptions)
	formatted := result.Formatted
	pattern := compactPatternForFormatted(entry, formatted, cardinalRule)
	if shouldUseGrouping(resolved.UseGrouping, formatted) {
		formatted = groupDecimal(formatted, grouping)
	}
	text := localizeFormattedNumberText(formatted, symbols, resolved.NumberingSystem)
	text = applySignDisplayText(text, formatted, d.Negative(), signDisplay, symbols)
	return pattern.formatText(text), result.Rounded
}

func resolveCompactPattern(d decimal.Decimal, digitOptions ecma402nf.DigitOptions, compact compactPatternSet) (decimal.Decimal, compactPatternEntry) {
	magnitude, err := decimal.Log10Floor(decimal.Abs(d))
	if err != nil || magnitude < minCompactExponent {
		return d, compactPatternEntry{}
	}
	entry, ok := compact.patternForMagnitude(int(magnitude))
	if !ok {
		return d, compactPatternEntry{}
	}
	scaled := decimal.Scale10(d, -int32(entry.exponent)) // #nosec G115 -- compact exponents are small generated data keys.
	result := ecma402nf.FormatNumericToString(scaled, digitOptions)
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
	next, ok := compact.patternForMagnitude(int(magnitude) + 1)
	if !ok {
		return scaled, entry
	}
	return decimal.Scale10(d, -int32(next.exponent)), next // #nosec G115 -- compact exponents are small generated data keys.
}

func compactPatternForFormatted(entry compactPatternEntry, formatted string, cardinalRule pluralRuleFunc) compactAffixPattern {
	if !entry.patterns[ecma402pr.Other].set {
		return compactAffixPattern{}
	}
	category := pluralCategoryWithExponent(cardinalRule, strings.TrimPrefix(formatted, "-"), entry.exponent)
	return entry.pattern(category)
}

func formatScientificAppend(parts []Part, d decimal.Decimal, notation Notation, state *decimalFormatState) ([]Part, decimal.Decimal) {
	symbols := state.symbols
	resolved := state.resolved
	signDisplay := resolved.SignDisplay
	grouping := state.grouping
	digitOptions := state.digitOptions
	exponent, ok := ecma402nf.ScientificExponent(d, notation == EngineeringNotation)
	if !ok {
		return append(parts, Part{Type: PartNaN, Value: symbols.NaN}), decimal.NaNValue
	}
	scaled := decimal.Scale10(d, -int32(exponent)) // #nosec G115 -- exponent came from decimal.Log10Floor int32.
	result := ecma402nf.FormatNumericToString(scaled, digitOptions)
	formatted := result.Formatted
	if shouldUseGrouping(resolved.UseGrouping, formatted) {
		formatted = groupDecimal(formatted, grouping)
	}
	parts = appendDecimalParts(parts, formatted, symbols)
	parts = applySignDisplay(parts, d.Negative(), signDisplay, symbols)
	parts = append(parts, Part{Type: PartExponentSeparator, Value: symbols.Exponential})
	if exponent < 0 {
		parts = append(parts, Part{Type: PartExponentMinusSign, Value: symbols.Minus})
		exponent = -exponent
	}
	exponentInteger := strconv.Itoa(exponent)
	parts = append(parts, Part{Type: PartExponentInteger, Value: exponentInteger})
	return parts, result.Rounded
}

func formatScientificText(d decimal.Decimal, notation Notation, state *decimalFormatState) (string, decimal.Decimal) {
	symbols := state.symbols
	resolved := state.resolved
	signDisplay := resolved.SignDisplay
	grouping := state.grouping
	digitOptions := state.digitOptions
	exponent, ok := ecma402nf.ScientificExponent(d, notation == EngineeringNotation)
	if !ok {
		return symbols.NaN, decimal.NaNValue
	}
	scaled := decimal.Scale10(d, -int32(exponent)) // #nosec G115 -- exponent came from decimal.Log10Floor int32.
	result := ecma402nf.FormatNumericToString(scaled, digitOptions)
	formatted := result.Formatted
	if shouldUseGrouping(resolved.UseGrouping, formatted) {
		formatted = groupDecimal(formatted, grouping)
	}
	text := localizeFormattedNumberText(formatted, symbols, resolved.NumberingSystem)
	text = applySignDisplayText(text, formatted, d.Negative(), signDisplay, symbols)
	text += symbols.Exponential
	if exponent < 0 {
		text += symbols.Minus
		exponent = -exponent
	}
	text += localizeExponentText(strconv.Itoa(exponent), resolved.NumberingSystem)
	return text, result.Rounded
}

func localizeExponentText(text, numberingSystem string) string {
	if numberingSystem == "" || numberingSystem == numbering.DefaultNumberingSystem {
		return text
	}
	return ecma402.LocalizeDigits(text, numberingSystem)
}

const (
	minCompactExponent = 3
	maxCompactExponent = 14
)

type compactPatternSet struct {
	entries []compactPatternEntry
}

type compactPatternEntry struct {
	magnitude, exponent int
	patterns            [numberPluralCategoryCount]compactAffixPattern
}

func compactPatternsForNumberFormat(loc cldrnumber.Locale, opts ResolvedOptions) compactPatternSet {
	if opts.Notation != CompactNotation {
		return compactPatternSet{}
	}
	display := string(ecma402.ResolvedScalarValue(opts.CompactDisplay))
	entries := make([]compactPatternEntry, 0, maxCompactExponent-minCompactExponent+1)
	for exponent := maxCompactExponent; exponent >= minCompactExponent; exponent-- {
		other := loc.CompactPattern(opts.NumberingSystem, display, exponent, "other")
		if other == "" {
			continue
		}
		entry := compactPatternEntry{
			magnitude: exponent,
			exponent:  compactExponentForPattern(exponent, other),
		}
		for _, category := range numberPluralCategories {
			entry.patterns[category] = compileCompactAffixPattern(loc.CompactPattern(opts.NumberingSystem, display, exponent, category.String()))
		}
		entry.patterns[ecma402pr.Other] = compileCompactAffixPattern(other)
		entries = append(entries, entry)
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

func (p compactPatternEntry) pattern(plural ecma402pr.Category) compactAffixPattern {
	if int(plural) < len(p.patterns) {
		if pattern := p.patterns[plural]; pattern.set {
			return pattern
		}
	}
	return p.patterns[ecma402pr.Other]
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
	if !p.set {
		return parts
	}
	return joinPatternParts(p.prefix, parts, p.suffix)
}

func (p compactAffixPattern) formatText(text string) string {
	if !p.set || len(p.prefix)+len(p.suffix) == 0 {
		return text
	}
	return joinPatternText(p.prefix, text, p.suffix)
}
