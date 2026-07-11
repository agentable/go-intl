package numberformat

import (
	"strconv"
	"strings"

	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
	"github.com/agentable/go-intl/internal/numbering"
	pluralop "github.com/agentable/go-intl/internal/plural"
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
	scaled, magnitude, _, ok := ecma402nf.ResolveCompactMagnitude(d, digitOptions, compact.exponentForMagnitude)
	if !ok {
		return d, compactPatternEntry{}
	}
	entry, _ := compact.patternForMagnitude(magnitude)
	return scaled, entry
}

func compactPatternForFormatted(entry compactPatternEntry, formatted string, cardinalRule pluralRuleFunc) compactAffixPattern {
	if !entry.patterns[pluralop.Other].set {
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
	exponent, ok := ecma402nf.ScientificExponent(d, digitOptions, notation == EngineeringNotation)
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
	exponent, ok := ecma402nf.ScientificExponent(d, digitOptions, notation == EngineeringNotation)
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
	entries := make([]compactPatternEntry, 0, ecma402nf.MaxCompactMagnitude-ecma402nf.MinCompactMagnitude+1)
	for exponent := ecma402nf.MaxCompactMagnitude; exponent >= ecma402nf.MinCompactMagnitude; exponent-- {
		other := loc.CompactPattern(opts.NumberingSystem, display, exponent, "other")
		if other == "" {
			continue
		}
		entry := compactPatternEntry{
			magnitude: exponent,
			exponent:  ecma402nf.CompactExponentForPattern(exponent, other),
		}
		for _, category := range numberPluralCategories {
			entry.patterns[category] = compileCompactAffixPattern(loc.CompactPattern(opts.NumberingSystem, display, exponent, category.String()))
		}
		entry.patterns[pluralop.Other] = compileCompactAffixPattern(other)
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

func (p compactPatternSet) exponentForMagnitude(magnitude int) (int, bool) {
	entry, ok := p.patternForMagnitude(magnitude)
	if !ok {
		return 0, false
	}
	return entry.exponent, true
}

func (p compactPatternEntry) pattern(plural pluralop.Category) compactAffixPattern {
	if int(plural) < len(p.patterns) {
		if pattern := p.patterns[plural]; pattern.set {
			return pattern
		}
	}
	return p.patterns[pluralop.Other]
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
