package numberformat

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
)

// FormatRange formats a numeric range.
func (f *NumberFormat) FormatRange(start, end Value) (string, error) {
	return formatRangeDecimal(start.numeric.Decimal, end.numeric.Decimal, &f.formatState)
}

// FormatRangeToParts formats a numeric range into ECMA-402 range parts.
func (f *NumberFormat) FormatRangeToParts(start, end Value) ([]RangePart, error) {
	return formatRangeToPartsDecimal(start.numeric.Decimal, end.numeric.Decimal, &f.formatState)
}

func formatRangeDecimal(start, end decimal.Decimal, formatState *decimalFormatState) (string, error) {
	startParts, endParts, approximate, err := rangeEndpointParts(start, end, formatState)
	if err != nil {
		return "", err
	}
	if approximate {
		return formatApproximateRangeText(formatState.symbols.ApproxSign, startParts), nil
	}
	separator := numberRangeSeparator(startParts, formatState.symbols.RangeSign)
	startParts, endParts = collapseRangeEndpointParts(startParts, endParts)
	return joinRangeText(startParts, separator, endParts), nil
}

func formatRangeToPartsDecimal(start, end decimal.Decimal, formatState *decimalFormatState) ([]RangePart, error) {
	startParts, endParts, approximate, err := rangeEndpointParts(start, end, formatState)
	if err != nil {
		return nil, err
	}
	if approximate {
		out := make([]RangePart, len(startParts)+1)
		out[0] = RangePart{Type: PartApproximatelySign, Value: formatState.symbols.ApproxSign, Source: SourceShared}
		fillRangeParts(out[1:], startParts, SourceShared)
		return out, nil
	}
	separator := numberRangeSeparator(startParts, formatState.symbols.RangeSign)
	startParts, endParts = collapseRangeEndpointParts(startParts, endParts)
	out := make([]RangePart, len(startParts)+1+len(endParts))
	fillRangeParts(out[:len(startParts)], startParts, SourceStartRange)
	out[len(startParts)] = RangePart{Type: PartLiteral, Value: separator, Source: SourceShared}
	fillRangeParts(out[len(startParts)+1:], endParts, SourceEndRange)
	return out, nil
}

func rangeEndpointParts(start, end decimal.Decimal, formatState *decimalFormatState) ([]Part, []Part, bool, error) {
	if start.IsNaN() || end.IsNaN() {
		return nil, nil, false, invalidNumberRange(start, end, formatState.resolved.Locale.String())
	}
	startParts := formatDecimalToPartsAppend(nil, start, formatState)
	endParts := formatDecimalToPartsAppend(nil, end, formatState)
	if partValuesEqual(startParts, endParts) {
		return startParts, endParts, true, nil
	}
	return startParts, endParts, false, nil
}

func invalidNumberRange(start, end decimal.Decimal, loc string) error {
	value := fmt.Sprintf("start=%s end=%s", start.String(), end.String())
	return ecma402.InvalidValueErrorExpected(numberFormatOwner, "range", value, loc, "numeric range values that are not NaN", decimal.ErrInvalidDecimal)
}

func numberRangeSeparator(startParts []Part, sign string) string {
	if sign == "" {
		sign = "–"
	}
	if len(startParts) > 0 && isSignPart(startParts[0].Type) {
		return " " + sign + " "
	}
	return sign
}

func isSignPart(typ PartType) bool {
	return typ == PartMinusSign || typ == PartPlusSign
}

func partValuesEqual(startParts, endParts []Part) bool {
	return partsText(startParts) == partsText(endParts)
}

func fillRangeParts(out []RangePart, parts []Part, source RangeSource) {
	for i, part := range parts {
		out[i] = RangePart{Type: part.Type, Value: part.Value, Source: source}
	}
}

func partsText(parts []Part) string {
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

func formatApproximateRangeText(sign string, parts []Part) string {
	size := len(sign)
	for _, part := range parts {
		size += len(part.Value)
	}
	var b strings.Builder
	b.Grow(size)
	b.WriteString(sign)
	for _, part := range parts {
		b.WriteString(part.Value)
	}
	return b.String()
}

func joinRangeText(startParts []Part, separator string, endParts []Part) string {
	size := len(separator)
	for _, part := range startParts {
		size += len(part.Value)
	}
	for _, part := range endParts {
		size += len(part.Value)
	}
	var b strings.Builder
	b.Grow(size)
	for _, part := range startParts {
		b.WriteString(part.Value)
	}
	b.WriteString(separator)
	for _, part := range endParts {
		b.WriteString(part.Value)
	}
	return b.String()
}

func collapseRangeEndpointParts(startParts, endParts []Part) ([]Part, []Part) {
	if count := collapsiblePartSuffixCount(startParts); partWidth(startParts[len(startParts)-count:]) > 1 {
		return startParts[:len(startParts)-count], endParts
	}
	if count := collapsiblePartPrefixCount(endParts); partWidth(endParts[:count]) > 1 {
		return startParts, endParts[count:]
	}
	return startParts, endParts
}

func collapsiblePartSuffixCount(parts []Part) int {
	count := 0
	for i := len(parts) - 1; i >= 0; i-- {
		if !isCollapsibleRangePart(parts[i].Type) {
			break
		}
		count++
	}
	return count
}

func collapsiblePartPrefixCount(parts []Part) int {
	count := 0
	for _, part := range parts {
		if !isCollapsibleRangePart(part.Type) {
			break
		}
		count++
	}
	return count
}

func partWidth(parts []Part) int {
	width := 0
	for _, part := range parts {
		width += utf8.RuneCountInString(part.Value)
	}
	return width
}

func isCollapsibleRangePart(typ PartType) bool {
	switch typ {
	case PartUnit, PartMinusSign, PartPlusSign, PartPercentSign, PartExponentSeparator, PartExponentMinusSign, PartCurrency, PartLiteral:
		return true
	case PartInteger, PartGroup, PartDecimal, PartFraction, PartNaN, PartInfinity, PartExponentInteger, PartCompact, PartApproximatelySign:
		return false
	}
	return false
}
