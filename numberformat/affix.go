package numberformat

import (
	"strings"

	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/cldr/plural"
	ecma402pr "github.com/agentable/go-intl/internal/ecma402/pluralrules"
)

func (f *NumberFormat) applyCurrencyPattern(parts []Part, pluralFormatted string) []Part {
	return f.applyCurrencyPatternForPlural(parts, pluralCategory(f.resolved.Locale.Tag().String(), strings.TrimPrefix(pluralFormatted, "-")))
}

func (f *NumberFormat) applyCurrencyPatternForPlural(parts []Part, plural string) []Part {
	if f.resolved.CurrencyDisplay == CurrencyDisplayName {
		sign, unsigned := splitLeadingSign(parts)
		name := f.currencyDisplay(plural)
		out := interpolateUnitPattern("{0} {1}", unsigned, Part{Type: PartCurrency, Value: name})
		if sign.Type == PartMinusSign && f.resolved.CurrencySign == AccountingCurrencySign {
			wrapped := make([]Part, 0, len(out)+2)
			wrapped = append(wrapped, Part{Type: PartLiteral, Value: "("})
			wrapped = append(wrapped, out...)
			wrapped = append(wrapped, Part{Type: PartLiteral, Value: ")"})
			return wrapped
		}
		if sign.Type != "" {
			return prependPart(sign, out)
		}
		return out
	}
	sign, unsigned := splitLeadingSign(parts)
	pattern, consumedSign := f.currency.pattern(sign)
	out := pattern.append(unsigned)
	if sign.Type != "" && !consumedSign {
		return prependPart(sign, out)
	}
	return out
}

func (f *NumberFormat) applyUnitPattern(parts []Part, pluralFormatted string) []Part {
	return f.applyUnitPatternForPlural(parts, pluralCategory(f.resolved.Locale.Tag().String(), pluralFormatted))
}

func (f *NumberFormat) applyUnitPatternForPlural(parts []Part, plural string) []Part {
	return f.unit.pattern(plural).append(parts)
}

func (f *NumberFormat) currencyDisplay(plural string) string {
	return currencyDisplayForNumberFormat(f.cldrLoc, f.resolved, plural)
}

func currencyDisplayForNumberFormat(loc cldrnumber.Locale, opts ResolvedOptions, plural string) string {
	code := opts.Currency
	switch opts.CurrencyDisplay {
	case CurrencyDisplayCode:
		return code
	case CurrencyDisplayName:
		if name := loc.CurrencyDisplayName(code, plural); name != "" {
			return name
		}
	case CurrencyDisplayNarrowSymbol:
		if symbol := loc.CurrencySymbol(code, "narrow"); symbol != "" {
			return symbol
		}
	case CurrencyDisplaySymbol:
		if symbol := loc.CurrencySymbol(code, "symbol"); symbol != "" {
			return symbol
		}
	default:
	}
	return code
}

func pluralCategory(localeTag, formatted string) string {
	return pluralCategoryWithExponent(localeTag, formatted, 0)
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

func pluralCategoryWithExponent(localeTag, formatted string, exponent int) string {
	rule, ok := plural.CardinalRule(localeTag)
	if !ok {
		rule, _ = plural.CardinalRule("en")
	}
	ops := ecma402pr.GetOperands(formatted, exponent)
	return rule(ops).String()
}

type currencyPatternSet struct {
	positive    numberAffixPattern
	negative    numberAffixPattern
	hasNegative bool
}

func currencyPatternsForNumberFormat(loc cldrnumber.Locale, opts ResolvedOptions) currencyPatternSet {
	if opts.Style != CurrencyStyle || opts.CurrencyDisplay == CurrencyDisplayName {
		return currencyPatternSet{}
	}
	pattern := loc.CurrencyPattern(opts.NumberingSystem, string(opts.CurrencySign))
	if pattern == "" {
		pattern = "¤#,##0.00"
	}
	positive, negative, hasNegative := strings.Cut(pattern, ";")
	affix := Part{Type: PartCurrency, Value: currencyDisplayForNumberFormat(loc, opts, "other")}
	set := currencyPatternSet{positive: compileNumberAffixPattern(positive, affix)}
	if hasNegative {
		set.negative = compileNumberAffixPattern(negative, affix)
		set.hasNegative = true
	}
	return set
}

func (p currencyPatternSet) pattern(sign Part) (numberAffixPattern, bool) {
	if sign.Type == PartMinusSign && p.hasNegative {
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
	out := make([]Part, 0, len(p.prefix)+len(parts)+len(p.suffix))
	out = append(out, p.prefix...)
	out = append(out, parts...)
	return append(out, p.suffix...)
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

type unitPatternSet struct {
	zero, one, two, few, many, other simpleUnitPattern
}

func unitPatternsForNumberFormat(loc cldrnumber.Locale, opts ResolvedOptions) unitPatternSet {
	if opts.Style != UnitStyle {
		return unitPatternSet{}
	}
	width := string(opts.UnitDisplay)
	other := loc.UnitPattern(opts.Unit, width, "other")
	if other == "" {
		other = defaultUnitPattern(opts.Unit)
	}
	compile := func(plural string) simpleUnitPattern {
		pattern := loc.UnitPattern(opts.Unit, width, plural)
		if pattern == "" {
			pattern = other
		}
		return compileSimpleUnitPattern(pattern)
	}
	return unitPatternSet{
		zero:  compile("zero"),
		one:   compile("one"),
		two:   compile("two"),
		few:   compile("few"),
		many:  compile("many"),
		other: compile("other"),
	}
}

func defaultUnitPattern(unit string) string {
	var b strings.Builder
	b.Grow(len("{0} ") + len(unit))
	b.WriteString("{0} ")
	b.WriteString(unit)
	return b.String()
}

func (p unitPatternSet) pattern(plural string) simpleUnitPattern {
	switch plural {
	case "zero":
		return p.zero
	case "one":
		return p.one
	case "two":
		return p.two
	case "few":
		return p.few
	case "many":
		return p.many
	default:
		return p.other
	}
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
		prefix: compilePatternTextParts(before, PartUnit),
		suffix: compilePatternTextParts(after, PartUnit),
	}
}

func (p simpleUnitPattern) append(parts []Part) []Part {
	out := make([]Part, 0, len(p.prefix)+len(parts)+len(p.suffix))
	out = append(out, p.prefix...)
	out = append(out, parts...)
	return append(out, p.suffix...)
}

func compilePatternTextParts(text string, typ PartType) []Part {
	return appendPatternTextParts(nil, text, typ)
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

func interpolateUnitPattern(pattern string, parts []Part, unit Part) []Part {
	out := make([]Part, 0, len(parts)+strings.Count(pattern, "{")+1)
	for pattern != "" {
		if rest, ok := strings.CutPrefix(pattern, "{0}"); ok {
			out = append(out, parts...)
			pattern = rest
			continue
		}
		if rest, ok := strings.CutPrefix(pattern, "{1}"); ok {
			out = append(out, unit)
			pattern = rest
			continue
		}
		idx := strings.Index(pattern, "{")
		if idx < 0 {
			out = append(out, Part{Type: PartLiteral, Value: pattern})
			break
		}
		if idx > 0 {
			out = append(out, Part{Type: PartLiteral, Value: pattern[:idx]})
			pattern = pattern[idx:]
			continue
		}
		out = append(out, Part{Type: PartLiteral, Value: pattern[:1]})
		pattern = pattern[1:]
	}
	return out
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
