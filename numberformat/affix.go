package numberformat

import (
	"fmt"
	"strings"
	"unicode"

	cldrcurrency "github.com/agentable/go-intl/internal/cldr/currency"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	cldrunit "github.com/agentable/go-intl/internal/cldr/unit"
	"github.com/agentable/go-intl/internal/ecma402"
	pluralop "github.com/agentable/go-intl/internal/plural"
)

func applyCurrencyPatternForPlural(parts []Part, plural pluralop.Category, resolved ResolvedOptions, currencyLoc cldrcurrency.Locale, currency currencyPatternSet) []Part {
	if ecma402.ResolvedScalarValue(resolved.CurrencyDisplay) == CurrencyDisplayName {
		name := currencyDisplayForNumberFormat(currencyLoc, resolved, plural.String())
		return currency.name.pattern(plural).append(splitBidiSignPart(parts), name)
	}
	sign, unsigned := splitLeadingSign(parts)
	pattern, consumedSign := currency.pattern(sign.Type == PartMinusSign)
	out := pattern.append(unsigned)
	if sign.Type != "" && !consumedSign {
		return prependPart(sign, out)
	}
	return out
}

func applyUnitPatternForPlural(parts []Part, plural pluralop.Category, unit unitPatternSet) []Part {
	return unit.pattern(plural).append(parts)
}

func currencyDisplayForNumberFormat(loc cldrcurrency.Locale, opts ResolvedOptions, plural string) string {
	code := ecma402.ResolvedScalarValue(opts.Currency)
	switch ecma402.ResolvedScalarValue(opts.CurrencyDisplay) {
	case CurrencyDisplayCode:
		return code
	case CurrencyDisplayName:
		if name := cldrcurrency.DisplayName(loc, code, plural); name != "" {
			return name
		}
	case CurrencyDisplayNarrowSymbol:
		if symbol := cldrcurrency.NarrowSymbol(loc, code); symbol != "" {
			return symbol
		}
	case CurrencyDisplaySymbol:
		if symbol := cldrcurrency.Symbol(loc, code); symbol != "" {
			return symbol
		}
	default:
	}
	return code
}

func pluralCategory(cardinalRule pluralRuleFunc, formatted string) pluralop.Category {
	return pluralCategoryWithExponent(cardinalRule, formatted, 0)
}

func pluralCategoryForNotation(cardinalRule pluralRuleFunc, formatted string, exponent int) pluralop.Category {
	return pluralCategoryWithExponent(cardinalRule, pluralFormattedWithExponent(formatted, exponent), exponent)
}

func pluralFormattedWithExponent(formatted string, exponent int) string {
	if exponent == 0 {
		return formatted
	}
	sign := ""
	if rest, ok := strings.CutPrefix(formatted, "-"); ok {
		sign = "-"
		formatted = rest
	}
	integer, fraction, _ := strings.Cut(formatted, ".")
	digits := integer + fraction
	decimalIndex := len(integer) + exponent
	switch {
	case decimalIndex <= 0:
		return sign + "0." + strings.Repeat("0", -decimalIndex) + digits
	case decimalIndex >= len(digits):
		return sign + digits + strings.Repeat("0", decimalIndex-len(digits))
	default:
		return sign + digits[:decimalIndex] + "." + digits[decimalIndex:]
	}
}

func pluralCategoryWithExponent(cardinalRule pluralRuleFunc, formatted string, exponent int) pluralop.Category {
	ops := pluralop.GetOperands(formatted, exponent)
	return cardinalRule(ops)
}

const numberPluralCategoryCount = int(pluralop.Other) + 1

var numberPluralCategories = [...]pluralop.Category{
	pluralop.Zero,
	pluralop.One,
	pluralop.Two,
	pluralop.Few,
	pluralop.Many,
	pluralop.Other,
}

type currencyPatternSet struct {
	positive    numberAffixPattern
	negative    numberAffixPattern
	hasNegative bool
	name        currencyNamePatternSet
}

func currencyPatternsForNumberFormat(loc cldrnumber.Locale, currencyLoc cldrcurrency.Locale, opts ResolvedOptions) currencyPatternSet {
	if opts.Style != CurrencyStyle {
		return currencyPatternSet{}
	}
	if ecma402.ResolvedScalarValue(opts.CurrencyDisplay) == CurrencyDisplayName {
		var patterns currencyNamePatternSet
		for _, category := range numberPluralCategories {
			patterns[category] = compileCurrencyNamePattern(loc.CurrencyNamePattern(opts.NumberingSystem, category.String()))
		}
		return currencyPatternSet{name: patterns}
	}
	pattern := loc.CurrencyPattern(opts.NumberingSystem, string(ecma402.ResolvedScalarValue(opts.CurrencySign)))
	positive, negative, hasNegative := strings.Cut(pattern, ";")
	affix := Part{Type: PartCurrency, Value: currencyDisplayForNumberFormat(currencyLoc, opts, "other")}
	set := currencyPatternSet{positive: compileNumberAffixPattern(positive, affix)}
	if hasNegative {
		set.negative = compileNumberAffixPattern(negative, affix)
		set.hasNegative = true
	}
	return set
}

func (p currencyPatternSet) pattern(negative bool) (numberAffixPattern, bool) {
	if negative && p.hasNegative {
		return p.negative, true
	}
	return p.positive, false
}

type currencyNamePattern struct {
	prefix, between, suffix string
	currencyFirst           bool
}

type currencyNamePatternSet [numberPluralCategoryCount]currencyNamePattern

func (p currencyNamePatternSet) pattern(plural pluralop.Category) currencyNamePattern {
	if int(plural) < len(p) {
		return p[plural]
	}
	return p[pluralop.Other]
}

func compileCurrencyNamePattern(pattern string) currencyNamePattern {
	numberIndex := strings.Index(pattern, "{0}")
	currencyIndex := strings.Index(pattern, "{1}")
	if currencyIndex < numberIndex {
		prefix, rest, _ := strings.Cut(pattern, "{1}")
		between, suffix, _ := strings.Cut(rest, "{0}")
		return currencyNamePattern{prefix: prefix, between: between, suffix: suffix, currencyFirst: true}
	}
	prefix, rest, _ := strings.Cut(pattern, "{0}")
	between, suffix, _ := strings.Cut(rest, "{1}")
	return currencyNamePattern{prefix: prefix, between: between, suffix: suffix}
}

func (p currencyNamePattern) append(number []Part, name string) []Part {
	out := make([]Part, 0, len(number)+4)
	out = appendLiteral(out, p.prefix)
	if p.currencyFirst {
		out = append(out, Part{Type: PartCurrency, Value: name})
		out = appendLiteral(out, p.between)
		out = append(out, number...)
	} else {
		out = append(out, number...)
		out = appendLiteral(out, p.between)
		out = append(out, Part{Type: PartCurrency, Value: name})
	}
	return appendLiteral(out, p.suffix)
}

func appendLiteral(parts []Part, value string) []Part {
	if value == "" {
		return parts
	}
	return append(parts, Part{Type: PartLiteral, Value: value})
}

func splitBidiSignPart(parts []Part) []Part {
	if len(parts) == 0 || (parts[0].Type != PartMinusSign && parts[0].Type != PartPlusSign) {
		return parts
	}
	value := parts[0].Value
	coreWithSuffix := strings.TrimLeftFunc(value, isBidiSignMark)
	core := strings.TrimRightFunc(coreWithSuffix, isBidiSignMark)
	prefix := value[:len(value)-len(coreWithSuffix)]
	suffix := coreWithSuffix[len(core):]
	if prefix == "" && suffix == "" {
		return parts
	}
	out := make([]Part, 0, len(parts)+2)
	out = appendLiteral(out, prefix)
	out = append(out, Part{Type: parts[0].Type, Value: core})
	out = appendLiteral(out, suffix)
	return append(out, parts[1:]...)
}

func isBidiSignMark(r rune) bool {
	return r == '\u061c' || r == '\u200e' || r == '\u200f'
}

type numberAffixPattern struct {
	prefix []Part
	suffix []Part
}

func compileNumberAffixPattern(pattern string, affix Part) numberAffixPattern {
	start, end := numberPatternBounds(pattern)
	if start < 0 {
		return numberAffixPattern{prefix: []Part{affix}}
	}
	return numberAffixPattern{
		prefix: compileCurrencyLiteralParts(pattern[:start], affix),
		suffix: compileCurrencyLiteralParts(pattern[end:], affix),
	}
}

func (p numberAffixPattern) append(parts []Part) []Part {
	return joinPatternParts(p.prefix, parts, p.suffix)
}

func compileCurrencyLiteralParts(s string, affix Part) []Part {
	var parts []Part
	for s != "" {
		if rest, ok := strings.CutPrefix(s, "¤"); ok {
			parts = append(parts, affix)
			s = rest
			continue
		}
		idx := strings.Index(s, "¤")
		if idx < 0 {
			parts = append(parts, Part{Type: PartLiteral, Value: s})
			break
		}
		if idx > 0 {
			parts = append(parts, Part{Type: PartLiteral, Value: s[:idx]})
			s = s[idx:]
		}
	}
	return parts
}

type unitPatternSet [numberPluralCategoryCount]simpleUnitPattern

func unitPatternsForNumberFormat(loc cldrlocale.Locale, opts ResolvedOptions, cardinalRule pluralRuleFunc, localeName string) (unitPatternSet, error) {
	if opts.Style != UnitStyle {
		return unitPatternSet{}, nil
	}
	unit := ecma402.ResolvedScalarValue(opts.Unit)
	width := string(ecma402.ResolvedScalarValue(opts.UnitDisplay))
	numerator, denominator, compound := strings.Cut(unit, "-per-")
	if !compound {
		return simpleUnitPatternsForNumberFormat(loc, unit, width), nil
	}
	if directPatterns, ok := generatedUnitPatterns(loc, unit, width); ok {
		return compileCompoundUnitPatternSet(directPatterns), nil
	}
	numeratorPatterns, ok := generatedUnitPatterns(loc, numerator, width)
	if !ok {
		return unitPatternSet{}, missingUnitPatternError(localeName, numerator, width)
	}

	var out unitPatternSet
	if perUnit := cldrunit.PerUnitPattern(loc, denominator, width); perUnit != "" {
		for _, category := range numberPluralCategories {
			pattern := strings.Replace(perUnit, "{0}", numeratorPatterns[category], 1)
			out[category] = compileCompoundUnitPattern(pattern)
		}
		return out, nil
	}

	denominatorPatterns, ok := generatedUnitPatterns(loc, denominator, width)
	if !ok {
		return unitPatternSet{}, missingUnitPatternError(localeName, denominator, width)
	}
	compoundPattern := cldrunit.CompoundUnitPattern(loc, width)
	if compoundPattern == "" {
		return unitPatternSet{}, fmt.Errorf("numberformat: compound unit pattern missing for locale %q and width %q", localeName, width)
	}
	denominatorPattern := denominatorPatterns[pluralCategory(cardinalRule, "1")]
	denominatorText := strings.TrimSpace(strings.Replace(denominatorPattern, "{0}", "", 1))
	for _, category := range numberPluralCategories {
		pattern := strings.Replace(compoundPattern, "{0}", numeratorPatterns[category], 1)
		pattern = strings.Replace(pattern, "{1}", denominatorText, 1)
		out[category] = compileCompoundUnitPattern(pattern)
	}
	return out, nil
}

func simpleUnitPatternsForNumberFormat(loc cldrlocale.Locale, unit, width string) unitPatternSet {
	other := cldrunit.UnitPattern(loc, unit, width, "other")
	if other == "" {
		other = defaultUnitPattern(unit)
	}
	var out unitPatternSet
	for _, category := range numberPluralCategories {
		pattern := cldrunit.UnitPattern(loc, unit, width, category.String())
		if pattern == "" {
			pattern = other
		}
		out[category] = compileCompoundUnitPattern(pattern)
	}
	return out
}

func generatedUnitPatterns(loc cldrlocale.Locale, unit, width string) ([numberPluralCategoryCount]string, bool) {
	other := cldrunit.UnitPattern(loc, unit, width, "other")
	if other == "" {
		return [numberPluralCategoryCount]string{}, false
	}
	var out [numberPluralCategoryCount]string
	for _, category := range numberPluralCategories {
		pattern := cldrunit.UnitPattern(loc, unit, width, category.String())
		if pattern == "" {
			pattern = other
		}
		out[category] = pattern
	}
	return out, true
}

func compileCompoundUnitPatternSet(patterns [numberPluralCategoryCount]string) unitPatternSet {
	var out unitPatternSet
	for _, category := range numberPluralCategories {
		out[category] = compileCompoundUnitPattern(patterns[category])
	}
	return out
}

func missingUnitPatternError(localeName, unit, width string) error {
	return fmt.Errorf("numberformat: unit pattern missing for locale %q, unit %q, and width %q", localeName, unit, width)
}

func defaultUnitPattern(unit string) string {
	var b strings.Builder
	b.Grow(len("{0} ") + len(unit))
	b.WriteString("{0} ")
	b.WriteString(unit)
	return b.String()
}

func (p unitPatternSet) pattern(plural pluralop.Category) simpleUnitPattern {
	if int(plural) < len(p) {
		return p[plural]
	}
	return p[pluralop.Other]
}

type simpleUnitPattern struct {
	prefix     []Part
	suffix     []Part
	omitNumber bool
}

func compileCompoundUnitPattern(pattern string) simpleUnitPattern {
	before, after, ok := strings.Cut(pattern, "{0}")
	if !ok {
		return simpleUnitPattern{suffix: []Part{{Type: PartUnit, Value: pattern}}, omitNumber: true}
	}
	return simpleUnitPattern{
		prefix: appendCompoundUnitPrefix(nil, before),
		suffix: appendCompoundUnitSuffix(nil, after),
	}
}

func appendCompoundUnitPrefix(parts []Part, text string) []Part {
	unitText := strings.TrimRightFunc(text, unicode.IsSpace)
	if unitText != "" {
		parts = append(parts, Part{Type: PartUnit, Value: unitText})
	}
	if spaces := text[len(unitText):]; spaces != "" {
		parts = append(parts, Part{Type: PartLiteral, Value: spaces})
	}
	return parts
}

func appendCompoundUnitSuffix(parts []Part, text string) []Part {
	unitText := strings.TrimLeftFunc(text, unicode.IsSpace)
	if spaces := text[:len(text)-len(unitText)]; spaces != "" {
		parts = append(parts, Part{Type: PartLiteral, Value: spaces})
	}
	if unitText != "" {
		parts = append(parts, Part{Type: PartUnit, Value: unitText})
	}
	return parts
}

func (p simpleUnitPattern) append(parts []Part) []Part {
	if p.omitNumber {
		parts = nil
	}
	return joinPatternParts(p.prefix, parts, p.suffix)
}

func splitLeadingSign(parts []Part) (Part, []Part) {
	if len(parts) == 0 {
		return Part{}, parts
	}
	if parts[0].Type != PartMinusSign && parts[0].Type != PartPlusSign {
		return Part{}, parts
	}
	return parts[0], parts[1:]
}

func appendPatternTextParts(parts []Part, text string, typ PartType) []Part {
	for text != "" {
		trimmed := strings.TrimLeftFunc(text, unicode.IsSpace)
		if len(trimmed) != len(text) {
			parts = append(parts, Part{Type: PartLiteral, Value: text[:len(text)-len(trimmed)]})
			text = trimmed
			continue
		}
		idx := strings.IndexFunc(text, unicode.IsSpace)
		if idx < 0 {
			return append(parts, Part{Type: typ, Value: text})
		}
		if idx > 0 {
			parts = append(parts, Part{Type: typ, Value: text[:idx]})
		}
		text = text[idx:]
	}
	return parts
}

func joinPatternParts(prefix, parts, suffix []Part) []Part {
	if len(prefix)+len(suffix) == 0 {
		return parts
	}
	out := make([]Part, len(prefix)+len(parts)+len(suffix))
	n := copy(out, prefix)
	n += copy(out[n:], parts)
	copy(out[n:], suffix)
	return out
}

func numberPatternBounds(pattern string) (int, int) {
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
