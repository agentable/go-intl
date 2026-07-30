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
	parts, err := partitionNumberRange(start.numeric.Decimal, end.numeric.Decimal, &f.formatState)
	if err != nil {
		return "", err
	}
	return rangePartsText(parts), nil
}

// FormatRangeToParts formats a numeric range into ECMA-402 range parts.
func (f *NumberFormat) FormatRangeToParts(start, end Value) ([]RangePart, error) {
	return partitionNumberRange(start.numeric.Decimal, end.numeric.Decimal, &f.formatState)
}

func partitionNumberRange(start, end decimal.Decimal, formatState *decimalFormatState) ([]RangePart, error) {
	if start.IsNaN() || end.IsNaN() {
		return nil, invalidNumberRange(start, end, formatState.resolved.Locale.String())
	}
	startParts := formatDecimalToPartsAppend(nil, start, formatState)
	endParts := formatDecimalToPartsAppend(nil, end, formatState)
	if partValuesEqual(startParts, endParts) {
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

func rangePartsText(parts []RangePart) string {
	size := 0
	for _, part := range parts {
		size += len(part.Value)
	}
	var out strings.Builder
	out.Grow(size)
	for _, part := range parts {
		out.WriteString(part.Value)
	}
	return out.String()
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
