package numberformat

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/agentable/go-intl/internal/cldr"
	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
)

func (f *NumberFormat) FormatInt(v int) string {
	return f.FormatInt64(int64(v))
}

func (f *NumberFormat) FormatInt64(v int64) string {
	if s, ok := f.formatFastInt64(v); ok {
		return s
	}
	return f.formatNumber(decimal.FromInt64(v))
}

func (f *NumberFormat) FormatUint(v uint) string {
	return f.FormatUint64(uint64(v))
}

func (f *NumberFormat) FormatUint64(v uint64) string {
	return f.formatNumber(decimal.New(false, new(big.Int).SetUint64(v), 0))
}

func (f *NumberFormat) FormatFloat64(v float64) string {
	return f.formatNumber(decimal.FromFloat64(v))
}

func (f *NumberFormat) FormatDecimal(v string) (string, error) {
	d, err := parseDecimalValue(v)
	if err != nil {
		return "", err
	}
	return f.formatNumber(d), nil
}

func (f *NumberFormat) FormatIntToParts(v int) []Part {
	return f.FormatInt64ToParts(int64(v))
}

func (f *NumberFormat) FormatInt64ToParts(v int64) []Part {
	return f.formatDecimalToParts(decimal.FromInt64(v))
}

func (f *NumberFormat) FormatUintToParts(v uint) []Part {
	return f.FormatUint64ToParts(uint64(v))
}

func (f *NumberFormat) FormatUint64ToParts(v uint64) []Part {
	return f.formatDecimalToParts(decimal.New(false, new(big.Int).SetUint64(v), 0))
}

func (f *NumberFormat) FormatFloat64ToParts(v float64) []Part {
	return f.formatDecimalToParts(decimal.FromFloat64(v))
}

func (f *NumberFormat) FormatDecimalToParts(v string) ([]Part, error) {
	d, err := parseDecimalValue(v)
	if err != nil {
		return nil, err
	}
	return f.formatDecimalToParts(d), nil
}

func (f *NumberFormat) formatValue(v any) string {
	if s, ok := f.formatFast(v); ok {
		return s
	}
	return f.formatNumber(valueDecimal(v))
}

func (f *NumberFormat) formatToPartsValue(v any) []Part {
	return f.formatDecimalToParts(valueDecimal(v))
}

func (f *NumberFormat) formatNumber(d decimal.Decimal) string {
	return joinParts(f.formatDecimalToParts(d))
}

func (f *NumberFormat) formatDecimalToParts(d decimal.Decimal) []Part {
	symbols := f.symbols()
	if d.IsNaN() {
		return f.applySpecialSignDisplay([]Part{{Type: PartNaN, Value: symbols.NaN}}, false, true)
	}
	if d.IsInf() {
		if d.Negative() {
			return f.applySpecialSignDisplay([]Part{{Type: PartInfinity, Value: symbols.Infinity}}, true, false)
		}
		return f.applySpecialSignDisplay([]Part{{Type: PartInfinity, Value: symbols.Infinity}}, false, false)
	}
	if f.resolved.Notation == "scientific" || f.resolved.Notation == "engineering" {
		return f.localizeParts(f.formatScientific(d))
	}
	if f.resolved.Notation == "compact" {
		return f.localizeParts(f.formatCompact(d))
	}
	if f.resolved.Style == "percent" {
		d = decimal.MulInt(d, 100)
	}
	formatted := f.formatFinite(d)
	if f.useGrouping(formatted) {
		formatted = groupDecimal(formatted, f.grouping)
	}
	parts := partitionDecimal(formatted, symbols)
	parts = f.applySignDisplay(parts, d.Negative())
	if f.resolved.Style == "percent" {
		parts = append(parts, Part{Type: PartPercentSign, Value: symbols.Percent})
	}
	if f.resolved.Style == "currency" {
		return f.localizeParts(f.applyCurrencyPattern(parts, formatted, d))
	}
	if f.resolved.Style == "unit" {
		return f.localizeParts(f.applyUnitPattern(parts, formatted))
	}
	return f.localizeParts(parts)
}

func valueDecimal(v any) decimal.Decimal {
	if v == nil {
		return decimal.NaNValue
	}
	d, err := decimal.ToIntlMathematicalValue(v)
	if err != nil {
		return decimal.NaNValue
	}
	return d
}

func parseDecimalValue(v string) (decimal.Decimal, error) {
	d, err := decimal.ParseString(v)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("numberformat: invalid decimal %q: %w", v, ErrInvalidValue)
	}
	return d, nil
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
		return PartMinusSign, negative
	default:
		return PartMinusSign, negative
	}
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

func signValue(sign PartType, symbols cldr.NumberSymbols) string {
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

func (f *NumberFormat) symbols() cldr.NumberSymbols {
	symbols := f.cldrLoc.NumberSymbols(f.resolved.NumberingSystem)
	if symbols.Decimal != "" {
		return symbols
	}
	return f.cldrLoc.NumberSymbols(f.cldrLoc.DefaultNumberingSystem())
}

func localizeNumberString(s string, symbols cldr.NumberSymbols) string {
	s = strings.ReplaceAll(s, "-", symbols.Minus)
	s = strings.ReplaceAll(s, ",", symbols.Group)
	return strings.ReplaceAll(s, ".", symbols.Decimal)
}

func localizeNumberStringForSystem(s string, symbols cldr.NumberSymbols, numberingSystem string) string {
	return ecma402.LocalizeDigits(localizeNumberString(s, symbols), numberingSystem)
}

func (f *NumberFormat) localizeParts(parts []Part) []Part {
	if f.resolved.NumberingSystem == "" || f.resolved.NumberingSystem == "latn" {
		return parts
	}
	out := slicesCloneParts(parts)
	for i := range out {
		switch out[i].Type {
		case PartInteger, PartFraction, PartExponentInteger:
			out[i].Value = ecma402.LocalizeDigits(out[i].Value, f.resolved.NumberingSystem)
		case PartGroup, PartDecimal, PartCurrency, PartPercentSign, PartMinusSign, PartPlusSign, PartNaN, PartInfinity, PartUnit, PartLiteral, PartExponentSeparator, PartExponentMinusSign, PartCompact, PartApproximatelySign:
		}
	}
	return out
}

func slicesCloneParts(parts []Part) []Part {
	out := make([]Part, len(parts))
	copy(out, parts)
	return out
}

func partitionDecimal(s string, symbols cldr.NumberSymbols) []Part {
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
	parts := make([]Part, 0, partCount)
	if negative {
		parts = append(parts, Part{Type: PartMinusSign, Value: symbols.Minus})
	}
	segments := strings.Split(integer, ",")
	for i, segment := range segments {
		if i > 0 {
			parts = append(parts, Part{Type: PartGroup, Value: symbols.Group})
		}
		parts = append(parts, Part{Type: PartInteger, Value: segment})
	}
	if hasFraction {
		parts = append(parts, Part{Type: PartDecimal, Value: symbols.Decimal})
		parts = append(parts, Part{Type: PartFraction, Value: fraction})
	}
	return parts
}
