package numberformat

import (
	"strings"

	cldrcurrency "github.com/agentable/go-intl/internal/cldr/currency"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	cldrunit "github.com/agentable/go-intl/internal/cldr/unit"
	"github.com/agentable/go-intl/internal/ecma402"
	pluralop "github.com/agentable/go-intl/internal/plural"
)

func applyCurrencyPattern(parts []Part, pluralFormatted string, cardinalRule pluralRuleFunc, resolved ResolvedOptions, currencyLoc cldrcurrency.Locale, currency currencyPatternSet) []Part {
	if ecma402.ResolvedScalarValue(resolved.CurrencyDisplay) != CurrencyDisplayName {
		return applyCurrencyPatternForPlural(parts, "other", resolved, currencyLoc, currency)
	}
	return applyCurrencyPatternForPlural(parts, pluralCategory(cardinalRule, strings.TrimPrefix(pluralFormatted, "-")).String(), resolved, currencyLoc, currency)
}

func applyCurrencyPatternForPlural(parts []Part, plural string, resolved ResolvedOptions, currencyLoc cldrcurrency.Locale, currency currencyPatternSet) []Part {
	if ecma402.ResolvedScalarValue(resolved.CurrencyDisplay) == CurrencyDisplayName {
		sign, unsigned := splitLeadingSign(parts)
		name := currencyDisplayForNumberFormat(currencyLoc, resolved, plural)
		out := make([]Part, len(unsigned)+2)
		copy(out, unsigned)
		out[len(unsigned)] = Part{Type: PartLiteral, Value: " "}
		out[len(unsigned)+1] = Part{Type: PartCurrency, Value: name}
		if sign.Type == PartMinusSign && ecma402.ResolvedScalarValue(resolved.CurrencySign) == AccountingCurrencySign {
			wrapped := make([]Part, len(out)+2)
			wrapped[0] = Part{Type: PartLiteral, Value: "("}
			copy(wrapped[1:], out)
			wrapped[len(wrapped)-1] = Part{Type: PartLiteral, Value: ")"}
			return wrapped
		}
		if sign.Type != "" {
			return prependPart(sign, out)
		}
		return out
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

func pluralNumberString(formatted string) string {
	integer, fraction, ok := strings.Cut(formatted, ".")
	if !ok {
		return formatted
	}
	fraction = strings.TrimRight(fraction, "0")
	if fraction == "" {
		return integer
	}
	return joinDecimalParts(integer, fraction)
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
}

func currencyPatternsForNumberFormat(loc cldrnumber.Locale, currencyLoc cldrcurrency.Locale, opts ResolvedOptions) currencyPatternSet {
	if opts.Style != CurrencyStyle || ecma402.ResolvedScalarValue(opts.CurrencyDisplay) == CurrencyDisplayName {
		return currencyPatternSet{}
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

func unitPatternsForNumberFormat(loc cldrlocale.Locale, opts ResolvedOptions) unitPatternSet {
	if opts.Style != UnitStyle {
		return unitPatternSet{}
	}
	unit := ecma402.ResolvedScalarValue(opts.Unit)
	width := string(ecma402.ResolvedScalarValue(opts.UnitDisplay))
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
		out[category] = compileSimpleUnitPattern(pattern)
	}
	return out
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
	prefix []Part
	suffix []Part
}

func compileSimpleUnitPattern(pattern string) simpleUnitPattern {
	before, after, ok := strings.Cut(pattern, "{0}")
	if !ok {
		return simpleUnitPattern{
			suffix: []Part{
				{Type: PartLiteral, Value: " "},
				{Type: PartUnit, Value: strings.TrimSpace(pattern)},
			},
		}
	}
	return simpleUnitPattern{
		prefix: appendPatternTextParts(nil, before, PartUnit),
		suffix: appendPatternTextParts(nil, after, PartUnit),
	}
}

func (p simpleUnitPattern) append(parts []Part) []Part {
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
		trimmed := strings.TrimLeft(text, " ")
		if len(trimmed) != len(text) {
			parts = append(parts, Part{Type: PartLiteral, Value: text[:len(text)-len(trimmed)]})
			text = trimmed
			continue
		}
		idx := strings.IndexByte(text, ' ')
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
