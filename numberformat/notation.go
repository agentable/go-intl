package numberformat

import (
	"fmt"
	"strconv"
	"strings"

	cldrcurrency "github.com/agentable/go-intl/internal/cldr/currency"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
	cldrpattern "github.com/agentable/go-intl/internal/pattern"
	pluralop "github.com/agentable/go-intl/internal/plural"
)

func formatCompactAppend(parts []Part, d decimal.Decimal, state *decimalFormatState) ([]Part, stylePluralOperand) {
	symbols := state.symbols
	resolved := state.resolved
	signDisplay := resolved.SignDisplay
	grouping := state.grouping
	digitOptions := state.digitOptions
	cardinalRule := state.cardinalRule
	compact := state.compact
	scaled, entry := resolveCompactPattern(d, digitOptions, compact)
	result := ecma402nf.FormatNumericToString(scaled, digitOptions)
	pattern := compactPatternForFormatted(entry, result.Formatted, cardinalRule)
	parts = appendDecimalParts(parts, result.Formatted, grouping, symbols)
	parts = applySignDisplay(parts, d.Negative(), signDisplay, symbols)
	return pattern.append(parts), stylePluralOperand{formatted: result.Formatted, exponent: entry.exponent, finite: true}
}

func resolveCompactPattern(d decimal.Decimal, digitOptions ecma402nf.ResolvedDigitOptions, compact compactPatternSet) (decimal.Decimal, compactPatternEntry) {
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

func formatScientificAppend(parts []Part, d decimal.Decimal, notation Notation, state *decimalFormatState) ([]Part, stylePluralOperand) {
	symbols := state.symbols
	resolved := state.resolved
	signDisplay := resolved.SignDisplay
	grouping := state.grouping
	digitOptions := state.digitOptions
	exponent, ok := ecma402nf.ScientificExponent(d, digitOptions, notation == EngineeringNotation)
	if !ok {
		return append(parts, Part{Type: PartNaN, Value: symbols.NaN}), stylePluralOperand{}
	}
	scaled := decimal.Scale10(d, -int32(exponent)) // #nosec G115 -- exponent came from decimal.Log10Floor int32.
	result := ecma402nf.FormatNumericToString(scaled, digitOptions)
	parts = appendDecimalParts(parts, result.Formatted, grouping, symbols)
	parts = applySignDisplay(parts, d.Negative(), signDisplay, symbols)
	parts = append(parts, Part{Type: PartExponentSeparator, Value: symbols.Exponential})
	pluralExponent := exponent
	if exponent < 0 {
		parts = appendBidiSymbol(parts, Part{Type: PartExponentMinusSign, Value: symbols.Minus})
		exponent = -exponent
	}
	exponentInteger := strconv.Itoa(exponent)
	parts = append(parts, Part{Type: PartExponentInteger, Value: exponentInteger})
	return parts, stylePluralOperand{formatted: result.Formatted, exponent: pluralExponent, finite: true}
}

type compactPatternSet struct {
	entries []compactPatternEntry
}

type compactPatternEntry struct {
	magnitude, exponent int
	patterns            [numberPluralCategoryCount]compactAffixPattern
}

func compactPatternsForNumberFormat(loc cldrnumber.Locale, currencyLoc cldrcurrency.Locale, opts ResolvedOptions) (compactPatternSet, error) {
	if opts.Notation != CompactNotation {
		return compactPatternSet{}, nil
	}
	display := string(ecma402.ResolvedScalarValue(opts.CompactDisplay))
	var currency Part
	if opts.Style == CurrencyStyle && ecma402.ResolvedScalarValue(opts.CurrencyDisplay) != CurrencyDisplayName {
		currency = Part{Type: PartCurrency, Value: currencyDisplayForNumberFormat(currencyLoc, opts, "other")}
	}
	for _, candidateDisplay := range []string{display, string(ShortCompactDisplay)} {
		for _, numberingSystem := range []string{opts.NumberingSystem, "latn"} {
			patterns, err := compileCompactPatternFamily(loc, opts, numberingSystem, candidateDisplay, currency)
			if err != nil || len(patterns.entries) != 0 {
				return patterns, err
			}
		}
	}
	return compactPatternSet{}, nil
}

func compileCompactPatternFamily(loc cldrnumber.Locale, opts ResolvedOptions, numberingSystem, display string, currency Part) (compactPatternSet, error) {
	patternFor := func(magnitude int, plural string) string {
		if currency.Type == PartCurrency {
			return loc.CurrencyCompactPattern(numberingSystem, display, magnitude, plural, false)
		}
		return loc.CompactPattern(numberingSystem, display, magnitude, plural)
	}
	entries := make([]compactPatternEntry, 0, ecma402nf.MaxCompactMagnitude-ecma402nf.MinCompactMagnitude+1)
	for exponent := ecma402nf.MaxCompactMagnitude; exponent >= ecma402nf.MinCompactMagnitude; exponent-- {
		other := patternFor(exponent, "other")
		if other == "" {
			continue
		}
		entry := compactPatternEntry{
			magnitude: exponent,
		}
		for _, category := range numberPluralCategories {
			raw := patternFor(exponent, category.String())
			parsed, err := cldrpattern.ParseCompact(exponent, raw)
			if err != nil {
				return compactPatternSet{}, fmt.Errorf("numberformat: compact pattern locale %s numberingSystem %s display %s magnitude %d plural %s: %w", opts.Locale, numberingSystem, display, exponent, category, err)
			}
			entry.patterns[category] = compileCompactAffixPattern(parsed, currency)
			if category == pluralop.Other {
				entry.exponent = parsed.Exponent
			}
		}
		entries = append(entries, entry)
	}
	return compactPatternSet{entries: entries}, nil
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
	prefix    []Part
	suffix    []Part
	hasNumber bool
	set       bool
}

func compileCompactAffixPattern(pattern cldrpattern.CompactPattern, currency Part) compactAffixPattern {
	return compactAffixPattern{
		prefix:    compileCompactAffixParts(pattern.Prefix, currency),
		suffix:    compileCompactAffixParts(pattern.Suffix, currency),
		hasNumber: pattern.HasNumber,
		set:       true,
	}
}

func compileCompactAffixParts(text string, currency Part) []Part {
	if currency.Type != PartCurrency {
		return appendPatternTextParts(nil, text, PartCompact)
	}
	var parts []Part
	for {
		before, after, found := strings.Cut(text, "¤")
		parts = appendPatternTextParts(parts, before, PartCompact)
		if !found {
			return parts
		}
		parts = appendBidiSymbol(parts, currency)
		text = after
	}
}

func (p compactAffixPattern) append(parts []Part) []Part {
	if !p.set {
		return parts
	}
	sign, unsigned := splitLeadingSign(parts)
	if !p.hasNumber {
		unsigned = nil
	}
	out := joinPatternParts(p.prefix, unsigned, p.suffix)
	if sign.Type != "" {
		return prependBidiSymbol(sign, out)
	}
	return out
}
