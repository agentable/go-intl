package numberformat

import (
	"errors"
	"slices"
	"strings"

	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intlerr"
)

// Format formats a numeric value.
func (f *NumberFormat) Format(v Value) string {
	switch v.kind {
	case valueInt64:
		if s, ok := f.formatFastInt64(v.int64); ok {
			return s
		}
	case valueUint64:
		if s, ok := f.formatFastUint64(v.uint64); ok {
			return s
		}
	case valueDecimal:
	}
	return f.formatNumber(v.decimal)
}

// FormatToParts formats a numeric value into ECMA-402 parts.
func (f *NumberFormat) FormatToParts(v Value) []Part {
	return f.formatDecimalToParts(v.decimal)
}

func (f *NumberFormat) formatValue(v any) string {
	return f.Format(anyValue(v))
}

func (f *NumberFormat) formatToPartsValue(v any) []Part {
	return f.FormatToParts(anyValue(v))
}

func (f *NumberFormat) formatNumber(d decimal.Decimal) string {
	var scratch [16]Part
	return joinParts(f.formatDecimalToPartsAppend(scratch[:0], d))
}

func (f *NumberFormat) formatDecimalToParts(d decimal.Decimal) []Part {
	return f.formatDecimalToPartsAppend(nil, d)
}

func (f *NumberFormat) formatDecimalToPartsAppend(parts []Part, d decimal.Decimal) []Part {
	symbols := f.symbols()
	if d.IsNaN() {
		parts = f.applySpecialSignDisplay(append(parts, Part{Type: PartNaN, Value: symbols.NaN}), false, true)
		return f.applyStylePattern(parts, d)
	}
	if d.IsInf() {
		if d.Negative() {
			parts = f.applySpecialSignDisplay(append(parts, Part{Type: PartInfinity, Value: symbols.Infinity}), true, false)
			return f.applyStylePattern(parts, d)
		}
		parts = f.applySpecialSignDisplay(append(parts, Part{Type: PartInfinity, Value: symbols.Infinity}), false, false)
		return f.applyStylePattern(parts, d)
	}
	if f.resolved.Style == "percent" {
		d = decimal.MulInt(d, 100)
	}
	var rounded decimal.Decimal
	switch f.resolved.Notation {
	case ScientificNotation, EngineeringNotation:
		parts, rounded = f.formatScientificAppend(parts, d)
	case CompactNotation:
		parts, rounded = f.formatCompactAppend(parts, d)
	case StandardNotation:
		result := f.formatFiniteResult(d)
		formatted := result.Formatted
		if f.useGrouping(formatted) {
			formatted = groupDecimal(formatted, f.grouping)
		}
		parts = appendDecimalParts(parts, formatted, symbols)
		parts = f.applySignDisplay(parts, d.Negative())
		rounded = result.Rounded
	}
	return f.applyStylePattern(parts, rounded)
}

func (f *NumberFormat) applyStylePattern(parts []Part, rounded decimal.Decimal) []Part {
	if f.resolved.Style == "percent" {
		symbols := f.symbols()
		parts = append(parts, Part{Type: PartPercentSign, Value: symbols.Percent})
	}
	if f.resolved.Style == "currency" {
		if rounded.IsFinite() {
			return f.localizeParts(f.applyCurrencyPattern(parts, pluralNumberString(rounded.String())))
		}
		return f.localizeParts(f.applyCurrencyPatternForPlural(parts, "other"))
	}
	if f.resolved.Style == "unit" {
		if rounded.IsFinite() {
			return f.localizeParts(f.applyUnitPattern(parts, pluralNumberString(rounded.String())))
		}
		return f.localizeParts(f.applyUnitPatternForPlural(parts, "other"))
	}
	return f.localizeParts(parts)
}

func parseDecimalValue(v string) (decimal.Decimal, error) {
	return parseNamedDecimalValue("decimal", v)
}

func parseNamedDecimalValue(name, v string) (decimal.Decimal, error) {
	d, err := ecma402.ParseDecimalInput(v)
	if err != nil {
		return decimal.Decimal{}, invalidValue(name, v, err)
	}
	return d, nil
}

func invalidValue(name, value string, err error) error {
	cause := intlerr.ErrInvalidValue
	if err != nil {
		cause = errors.Join(intlerr.ErrInvalidValue, err)
	}
	return intlerr.New(intlerr.InvalidValue, "numberformat", name, value, "", cause)
}

func (f *NumberFormat) applySpecialSignDisplay(parts []Part, negative bool, nan bool) []Part {
	sign, ok := f.displaySign(negative, false, nan)
	if !ok {
		return parts
	}
	return prependPart(Part{Type: sign, Value: signValue(sign, f.symbols())}, parts)
}

func (f *NumberFormat) applySignDisplay(parts []Part, negative bool) []Part {
	parts = withoutLeadingSign(parts)
	zero := numericPartsAreZero(parts)
	sign, ok := f.displaySign(negative, zero, false)
	if !ok {
		return parts
	}
	symbols := f.symbols()
	return prependPart(Part{Type: sign, Value: signValue(sign, symbols)}, parts)
}

func prependPart(part Part, parts []Part) []Part {
	out := make([]Part, 0, len(parts)+1)
	out = append(out, part)
	return append(out, parts...)
}

func (f *NumberFormat) displaySign(negative, zero, nan bool) (PartType, bool) {
	if nan {
		if f.resolved.SignDisplay == AlwaysSignDisplay {
			return PartPlusSign, true
		}
		return "", false
	}
	switch f.resolved.SignDisplay {
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

func signValue(sign PartType, symbols cldrnumber.NumberSymbols) string {
	if sign == PartPlusSign {
		return symbols.Plus
	}
	return symbols.Minus
}

func joinParts(parts []Part) string {
	size := 0
	for _, part := range parts {
		size += len(part.Value)
	}
	var b strings.Builder
	b.Grow(size)
	for _, part := range parts {
		b.WriteString(part.Value)
	}
	return b.String()
}

func (f *NumberFormat) symbols() cldrnumber.NumberSymbols {
	return f.numberSymbols
}

func numberSymbolsForNumberFormat(loc cldrnumber.Locale, numberingSystem string) cldrnumber.NumberSymbols {
	symbols := loc.NumberSymbols(numberingSystem)
	if symbols.Decimal != "" {
		return symbols
	}
	return loc.NumberSymbols(loc.DefaultNumberingSystem())
}

func localizeNumberString(s string, symbols cldrnumber.NumberSymbols) string {
	s = strings.ReplaceAll(s, "-", symbols.Minus)
	s = strings.ReplaceAll(s, ",", symbols.Group)
	return strings.ReplaceAll(s, ".", symbols.Decimal)
}

func (f *NumberFormat) localizeParts(parts []Part) []Part {
	if f.resolved.NumberingSystem == "" || f.resolved.NumberingSystem == "latn" {
		return parts
	}
	out := slices.Clone(parts)
	for i := range out {
		switch out[i].Type {
		case PartInteger, PartFraction, PartExponentInteger:
			out[i].Value = ecma402.LocalizeDigits(out[i].Value, f.resolved.NumberingSystem)
		case PartGroup, PartDecimal, PartCurrency, PartPercentSign, PartMinusSign, PartPlusSign, PartNaN, PartInfinity, PartUnit, PartLiteral, PartExponentSeparator, PartExponentMinusSign, PartCompact, PartApproximatelySign:
		}
	}
	return out
}

func partitionDecimal(s string, symbols cldrnumber.NumberSymbols) []Part {
	return appendDecimalParts(nil, s, symbols)
}

func appendDecimalParts(parts []Part, s string, symbols cldrnumber.NumberSymbols) []Part {
	rest, negative := strings.CutPrefix(s, "-")
	if negative {
		s = rest
	}
	integer, fraction, hasFraction := strings.Cut(s, ".")
	partCount := strings.Count(integer, ",") + 1
	if negative {
		partCount++
	}
	if hasFraction {
		partCount += 2
	}
	if cap(parts)-len(parts) < partCount {
		next := make([]Part, 0, len(parts)+partCount)
		next = append(next, parts...)
		parts = next
	}
	if negative {
		parts = append(parts, Part{Type: PartMinusSign, Value: symbols.Minus})
	}
	for {
		segment, rest, hasGroup := strings.Cut(integer, ",")
		parts = append(parts, Part{Type: PartInteger, Value: segment})
		if !hasGroup {
			break
		}
		parts = append(parts, Part{Type: PartGroup, Value: symbols.Group})
		integer = rest
	}
	if hasFraction {
		parts = append(parts,
			Part{Type: PartDecimal, Value: symbols.Decimal},
			Part{Type: PartFraction, Value: fraction},
		)
	}
	return parts
}
