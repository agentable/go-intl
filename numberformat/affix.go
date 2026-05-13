package numberformat

import (
	"strings"

	"github.com/agentable/go-intl/internal/cldr/plural"
	"github.com/agentable/go-intl/internal/decimal"
	ecma402pr "github.com/agentable/go-intl/internal/ecma402/pluralrules"
)

func (f *NumberFormat) applyCurrencyPattern(parts []Part, formatted string, d decimal.Decimal) []Part {
	if f.resolved.CurrencyDisplay == CurrencyDisplayName {
		name := f.currencyDisplay(pluralCategory(f.resolved.Locale.Tag().String(), strings.TrimPrefix(formatted, "-")))
		return interpolateUnitPattern("{0} {1}", parts, Part{Type: PartCurrency, Value: name})
	}
	sign, unsigned := splitLeadingSign(parts)
	pattern, consumedSign := f.currencyPattern(sign)
	if pattern == "" {
		pattern = "¤#,##0.00"
	}
	out := interpolateNumberPattern(pattern, unsigned, Part{Type: PartCurrency, Value: f.currencyDisplay("other")})
	if sign.Type != "" && !consumedSign {
		return prependPart(sign, out)
	}
	return out
}

func (f *NumberFormat) applyUnitPattern(parts []Part, formatted string) []Part {
	plural := pluralCategory(f.resolved.Locale.Tag().String(), formatted)
	pattern := f.cldrLoc.UnitPattern(f.resolved.Unit, string(f.resolved.UnitDisplay), plural)
	if pattern == "" && plural != "other" {
		pattern = f.cldrLoc.UnitPattern(f.resolved.Unit, string(f.resolved.UnitDisplay), "other")
	}
	if pattern == "" {
		pattern = "{0} " + f.resolved.Unit
	}
	return interpolateSimpleUnitPattern(pattern, parts)
}

func (f *NumberFormat) currencyDisplay(plural string) string {
	code := f.resolved.Currency
	switch f.resolved.CurrencyDisplay {
	case CurrencyDisplayCode:
		return code
	case CurrencyDisplayName:
		if name := f.cldrLoc.CurrencyDisplayName(code, plural); name != "" {
			return name
		}
	case CurrencyDisplayNarrowSymbol:
		if symbol := f.cldrLoc.CurrencySymbol(code, "narrow"); symbol != "" {
			return symbol
		}
	case CurrencyDisplaySymbol:
		if symbol := f.cldrLoc.CurrencySymbol(code, "symbol"); symbol != "" {
			return symbol
		}
	default:
	}
	return code
}

func pluralCategory(localeTag, formatted string) string {
	return pluralCategoryWithExponent(localeTag, formatted, 0)
}

func pluralCategoryWithExponent(localeTag, formatted string, exponent int) string {
	rule, ok := plural.CardinalRule(localeTag)
	if !ok {
		rule, _ = plural.CardinalRule("en")
	}
	ops := ecma402pr.GetOperands(formatted, exponent)
	return rule(ops).String()
}

func interpolateNumberPattern(pattern string, parts []Part, affix Part) []Part {
	start, end := numberPatternBounds(pattern)
	if start < 0 {
		return prependPart(affix, parts)
	}
	out := literalParts(pattern[:start])
	for _, part := range out {
		if part.Value == "¤" {
			out = out[:len(out)-1]
			out = append(out, affix)
		}
	}
	out = append(out, parts...)
	for _, part := range literalParts(pattern[end:]) {
		if part.Value == "¤" {
			out = append(out, affix)
			continue
		}
		out = append(out, part)
	}
	return out
}

func (f *NumberFormat) currencyPattern(sign Part) (string, bool) {
	style := string(f.resolved.CurrencySign)
	pattern := f.cldrLoc.CurrencyPattern(f.resolved.NumberingSystem, style)
	positive, negative, hasNegative := strings.Cut(pattern, ";")
	if sign.Type == PartMinusSign && hasNegative {
		return negative, true
	}
	return positive, false
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

func interpolateSimpleUnitPattern(pattern string, parts []Part) []Part {
	before, after, ok := strings.Cut(pattern, "{0}")
	if !ok {
		return append(parts, Part{Type: PartLiteral, Value: " "}, Part{Type: PartUnit, Value: strings.TrimSpace(pattern)})
	}
	out := make([]Part, 0, len(parts)+2)
	out = appendPatternTextParts(out, before, PartUnit)
	out = append(out, parts...)
	out = appendPatternTextParts(out, after, PartUnit)
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

func literalParts(s string) []Part {
	parts := make([]Part, 0, strings.Count(s, "¤")*2+1)
	for s != "" {
		if rest, ok := strings.CutPrefix(s, "¤"); ok {
			parts = append(parts, Part{Type: PartCurrency, Value: "¤"})
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
