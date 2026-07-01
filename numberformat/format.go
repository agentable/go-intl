package numberformat

import (
	"slices"
	"strings"

	cldrcurrency "github.com/agentable/go-intl/internal/cldr/currency"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
	ecma402pr "github.com/agentable/go-intl/internal/ecma402/pluralrules"
	"github.com/agentable/go-intl/internal/numbering"
)

// Format formats a numeric value.
func (f *NumberFormat) Format(v Value) string {
	numeric := v.numeric
	switch numeric.Kind {
	case ecma402.NumericValueInt64:
		if s, ok := formatFastInt64(numeric.Int64, f.integerFastPath); ok {
			return s
		}
	case ecma402.NumericValueUint64:
		if s, ok := formatFastUint64(numeric.Uint64, f.integerFastPath); ok {
			return s
		}
	case ecma402.NumericValueDecimal:
	}
	return formatDecimalText(numeric.Decimal, &f.formatState)
}

// FormatToParts formats a numeric value into ECMA-402 parts.
func (f *NumberFormat) FormatToParts(v Value) []Part {
	return formatDecimalToPartsAppend(nil, v.numeric.Decimal, &f.formatState)
}

type decimalFormatState struct {
	resolved     ResolvedOptions
	symbols      cldrnumber.NumberSymbols
	grouping     digitGrouping
	digitOptions ecma402nf.DigitOptions
	cardinalRule pluralRuleFunc
	currencyLoc  cldrcurrency.Locale
	currency     currencyPatternSet
	unit         unitPatternSet
	compact      compactPatternSet
}

func formatDecimalToPartsAppend(parts []Part, d decimal.Decimal, state *decimalFormatState) []Part {
	resolved := state.resolved
	style := resolved.Style
	notation := resolved.Notation
	signDisplay := resolved.SignDisplay
	symbols := state.symbols
	grouping := state.grouping
	digitOptions := state.digitOptions
	if d.IsNaN() {
		parts = applySpecialSignDisplay(append(parts, Part{Type: PartNaN, Value: symbols.NaN}), false, true, signDisplay, symbols)
		return applyStylePattern(parts, d, state)
	}
	if d.IsInf() {
		if d.Negative() {
			parts = applySpecialSignDisplay(append(parts, Part{Type: PartInfinity, Value: symbols.Infinity}), true, false, signDisplay, symbols)
			return applyStylePattern(parts, d, state)
		}
		parts = applySpecialSignDisplay(append(parts, Part{Type: PartInfinity, Value: symbols.Infinity}), false, false, signDisplay, symbols)
		return applyStylePattern(parts, d, state)
	}
	if style == PercentStyle {
		d = decimal.MulInt(d, 100)
	}
	var rounded decimal.Decimal
	switch notation {
	case ScientificNotation, EngineeringNotation:
		parts, rounded = formatScientificAppend(parts, d, notation, state)
	case CompactNotation:
		parts, rounded = formatCompactAppend(parts, d, state)
	case StandardNotation:
		result := ecma402nf.FormatNumericToString(d, digitOptions)
		formatted := result.Formatted
		if shouldUseGrouping(resolved.UseGrouping, formatted) {
			formatted = groupDecimal(formatted, grouping)
		}
		parts = appendDecimalParts(parts, formatted, symbols)
		parts = applySignDisplay(parts, d.Negative(), signDisplay, symbols)
		rounded = result.Rounded
	}
	return applyStylePattern(parts, rounded, state)
}

func formatDecimalText(d decimal.Decimal, state *decimalFormatState) string {
	resolved := state.resolved
	style := resolved.Style
	notation := resolved.Notation
	signDisplay := resolved.SignDisplay
	symbols := state.symbols
	grouping := state.grouping
	digitOptions := state.digitOptions
	if d.IsNaN() {
		text := applySpecialSignDisplayText(symbols.NaN, false, true, signDisplay, symbols)
		return applyStylePatternText(text, d, state)
	}
	if d.IsInf() {
		text := symbols.Infinity
		if d.Negative() {
			text = applySpecialSignDisplayText(text, true, false, signDisplay, symbols)
			return applyStylePatternText(text, d, state)
		}
		text = applySpecialSignDisplayText(text, false, false, signDisplay, symbols)
		return applyStylePatternText(text, d, state)
	}
	if style == PercentStyle {
		d = decimal.MulInt(d, 100)
	}
	var text string
	var rounded decimal.Decimal
	switch notation {
	case ScientificNotation, EngineeringNotation:
		text, rounded = formatScientificText(d, notation, state)
	case CompactNotation:
		text, rounded = formatCompactText(d, state)
	case StandardNotation:
		result := ecma402nf.FormatNumericToString(d, digitOptions)
		formatted := result.Formatted
		if shouldUseGrouping(resolved.UseGrouping, formatted) {
			formatted = groupDecimal(formatted, grouping)
		}
		text = localizeFormattedNumberText(formatted, symbols, resolved.NumberingSystem)
		text = applySignDisplayText(text, formatted, d.Negative(), signDisplay, symbols)
		rounded = result.Rounded
	}
	return applyStylePatternText(text, rounded, state)
}

func applyStylePattern(parts []Part, rounded decimal.Decimal, state *decimalFormatState) []Part {
	resolved := state.resolved
	style := resolved.Style
	numberingSystem := resolved.NumberingSystem
	if style == PercentStyle {
		parts = append(parts, Part{Type: PartPercentSign, Value: state.symbols.Percent})
	}
	if style == CurrencyStyle {
		if rounded.IsFinite() {
			return localizeParts(applyCurrencyPattern(parts, pluralNumberString(rounded.String()), state.cardinalRule, resolved, state.currencyLoc, state.currency), numberingSystem)
		}
		return localizeParts(applyCurrencyPatternForPlural(parts, "other", resolved, state.currencyLoc, state.currency), numberingSystem)
	}
	if style == UnitStyle {
		if rounded.IsFinite() {
			plural := pluralCategory(state.cardinalRule, pluralNumberString(rounded.String()))
			return localizeParts(applyUnitPatternForPlural(parts, plural, state.unit), numberingSystem)
		}
		return localizeParts(applyUnitPatternForPlural(parts, ecma402pr.Other, state.unit), numberingSystem)
	}
	return localizeParts(parts, numberingSystem)
}

func applyStylePatternText(text string, rounded decimal.Decimal, state *decimalFormatState) string {
	resolved := state.resolved
	style := resolved.Style
	if style == PercentStyle {
		text += state.symbols.Percent
	}
	if style == CurrencyStyle {
		if rounded.IsFinite() {
			return applyCurrencyPatternText(text, pluralNumberString(rounded.String()), state.cardinalRule, resolved, state.currencyLoc, state.currency, state.symbols)
		}
		return applyCurrencyPatternForPluralText(text, "other", resolved, state.currencyLoc, state.currency, state.symbols)
	}
	if style == UnitStyle {
		if rounded.IsFinite() {
			plural := pluralCategory(state.cardinalRule, pluralNumberString(rounded.String()))
			return applyUnitPatternForPluralText(text, plural, state.unit)
		}
		return applyUnitPatternForPluralText(text, ecma402pr.Other, state.unit)
	}
	return text
}

func applySpecialSignDisplay(parts []Part, negative bool, nan bool, signDisplay SignDisplay, symbols cldrnumber.NumberSymbols) []Part {
	sign, ok := displaySign(signDisplay, negative, false, nan)
	if !ok {
		return parts
	}
	return prependPart(Part{Type: sign, Value: signValue(sign, symbols)}, parts)
}

func applySpecialSignDisplayText(text string, negative bool, nan bool, signDisplay SignDisplay, symbols cldrnumber.NumberSymbols) string {
	sign, ok := displaySign(signDisplay, negative, false, nan)
	if !ok {
		return text
	}
	return signValue(sign, symbols) + text
}

func applySignDisplay(parts []Part, negative bool, signDisplay SignDisplay, symbols cldrnumber.NumberSymbols) []Part {
	parts = withoutLeadingSign(parts)
	zero := numericPartsAreZero(parts)
	sign, ok := displaySign(signDisplay, negative, zero, false)
	if !ok {
		return parts
	}
	return prependPart(Part{Type: sign, Value: signValue(sign, symbols)}, parts)
}

func applySignDisplayText(text, formatted string, negative bool, signDisplay SignDisplay, symbols cldrnumber.NumberSymbols) string {
	text = withoutLeadingSignText(text, symbols)
	zero := formattedNumberTextIsZero(formatted)
	sign, ok := displaySign(signDisplay, negative, zero, false)
	if !ok {
		return text
	}
	return signValue(sign, symbols) + text
}

func prependPart(part Part, parts []Part) []Part {
	out := make([]Part, len(parts)+1)
	out[0] = part
	copy(out[1:], parts)
	return out
}

func displaySign(signDisplay SignDisplay, negative, zero, nan bool) (PartType, bool) {
	if nan {
		if signDisplay == AlwaysSignDisplay {
			return PartPlusSign, true
		}
		return "", false
	}
	switch signDisplay {
	case AlwaysSignDisplay:
		if negative {
			return PartMinusSign, true
		}
		return PartPlusSign, true
	case ExceptZeroSignDisplay:
		if zero {
			return "", false
		}
		if negative {
			return PartMinusSign, true
		}
		return PartPlusSign, true
	case NegativeSignDisplay:
		return PartMinusSign, negative && !zero
	case NeverSignDisplay:
		return "", false
	case AutoSignDisplay:
	}
	return PartMinusSign, negative
}

func withoutLeadingSign(parts []Part) []Part {
	if len(parts) == 0 {
		return parts
	}
	if parts[0].Type == PartMinusSign || parts[0].Type == PartPlusSign {
		return parts[1:]
	}
	return parts
}

func withoutLeadingSignText(text string, symbols cldrnumber.NumberSymbols) string {
	if symbols.Minus != "" {
		if rest, ok := strings.CutPrefix(text, symbols.Minus); ok {
			return rest
		}
	}
	if symbols.Plus != "" {
		if rest, ok := strings.CutPrefix(text, symbols.Plus); ok {
			return rest
		}
	}
	return text
}

func numericPartsAreZero(parts []Part) bool {
	sawDigit := false
	for _, part := range parts {
		if part.Type != PartInteger && part.Type != PartFraction {
			continue
		}
		for _, r := range part.Value {
			if r < '0' || r > '9' {
				continue
			}
			sawDigit = true
			if r != '0' {
				return false
			}
		}
	}
	return sawDigit
}

func formattedNumberTextIsZero(formatted string) bool {
	sawDigit := false
	for _, r := range formatted {
		if r < '0' || r > '9' {
			continue
		}
		sawDigit = true
		if r != '0' {
			return false
		}
	}
	return sawDigit
}

func signValue(sign PartType, symbols cldrnumber.NumberSymbols) string {
	if sign == PartPlusSign {
		return symbols.Plus
	}
	return symbols.Minus
}

func localizeNumberString(s string, symbols cldrnumber.NumberSymbols) string {
	if symbols.Minus == "-" && symbols.Group == "," && symbols.Decimal == "." {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := range len(s) {
		switch s[i] {
		case '-':
			b.WriteString(symbols.Minus)
		case ',':
			b.WriteString(symbols.Group)
		case '.':
			b.WriteString(symbols.Decimal)
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func localizeFormattedNumberText(s string, symbols cldrnumber.NumberSymbols, numberingSystem string) string {
	s = localizeNumberString(s, symbols)
	if numberingSystem == "" || numberingSystem == numbering.DefaultNumberingSystem {
		return s
	}
	return ecma402.LocalizeDigits(s, numberingSystem)
}

func localizeParts(parts []Part, numberingSystem string) []Part {
	if numberingSystem == "" || numberingSystem == numbering.DefaultNumberingSystem {
		return parts
	}
	out := slices.Clone(parts)
	for i := range out {
		switch out[i].Type {
		case PartInteger, PartFraction, PartExponentInteger:
			out[i].Value = ecma402.LocalizeDigits(out[i].Value, numberingSystem)
		case PartGroup, PartDecimal, PartCurrency, PartPercentSign, PartMinusSign, PartPlusSign, PartNaN, PartInfinity, PartUnit, PartLiteral, PartExponentSeparator, PartExponentMinusSign, PartCompact, PartApproximatelySign:
		}
	}
	return out
}

func appendDecimalParts(parts []Part, s string, symbols cldrnumber.NumberSymbols) []Part {
	rest, negative := strings.CutPrefix(s, "-")
	if negative {
		s = rest
	}
	integer, fraction, hasFraction := strings.Cut(s, ".")
	groupCount := strings.Count(integer, ",")
	partCount := groupCount*2 + 1
	if negative {
		partCount++
	}
	if hasFraction {
		partCount += 2
	}
	partStart := len(parts)
	partEnd := partStart + partCount
	if cap(parts) < partEnd {
		next := make([]Part, partEnd)
		copy(next, parts)
		parts = next
	} else {
		parts = parts[:partEnd]
	}
	nextPart := partStart
	if negative {
		parts[nextPart] = Part{Type: PartMinusSign, Value: symbols.Minus}
		nextPart++
	}
	for {
		segment, rest, hasGroup := strings.Cut(integer, ",")
		parts[nextPart] = Part{Type: PartInteger, Value: segment}
		nextPart++
		if !hasGroup {
			break
		}
		parts[nextPart] = Part{Type: PartGroup, Value: symbols.Group}
		nextPart++
		integer = rest
	}
	if hasFraction {
		parts[nextPart] = Part{Type: PartDecimal, Value: symbols.Decimal}
		parts[nextPart+1] = Part{Type: PartFraction, Value: fraction}
	}
	return parts
}
