package numberformat

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/agentable/go-intl/internal/cldr/plural"
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
	startNumberParts, startOperand := formatDecimalOperandToPartsAppend(nil, start, formatState)
	endNumberParts, endOperand := formatDecimalOperandToPartsAppend(nil, end, formatState)
	startParts := applyStylePattern(slices.Clone(startNumberParts), startOperand, formatState)
	endParts := applyStylePattern(slices.Clone(endNumberParts), endOperand, formatState)
	if partValuesEqual(startParts, endParts) {
		out := make([]RangePart, len(startParts)+1)
		out[0] = RangePart{Type: PartApproximatelySign, Value: formatState.symbols.ApproxSign, Source: SourceShared}
		fillRangeParts(out[1:], startParts, SourceShared)
		return out, nil
	}
	if rangeUsesPluralCategory(formatState.resolved) {
		category := plural.ResolveCardinalRange(
			formatState.dataLocale,
			stylePluralCategory(startOperand, formatState.cardinalRule),
			stylePluralCategory(endOperand, formatState.cardinalRule),
		)
		startParts = applyStylePatternForPlural(startNumberParts, category, formatState)
		endParts = applyStylePatternForPlural(endNumberParts, category, formatState)
	}
	prefix, startParts, endParts, suffix := collapseRangeEndpointParts(startParts, endParts)
	separator := numberRangeSeparator(startParts, formatState.symbols.RangeSign)
	out := make([]RangePart, len(prefix)+len(startParts)+1+len(endParts)+len(suffix))
	n := len(prefix)
	fillRangeParts(out[:n], prefix, SourceShared)
	fillRangeParts(out[n:n+len(startParts)], startParts, SourceStartRange)
	n += len(startParts)
	out[n] = RangePart{Type: PartLiteral, Value: separator, Source: SourceShared}
	n++
	fillRangeParts(out[n:n+len(endParts)], endParts, SourceEndRange)
	n += len(endParts)
	fillRangeParts(out[n:], suffix, SourceShared)
	return out, nil
}

func rangeUsesPluralCategory(resolved ResolvedOptions) bool {
	return resolved.Style == UnitStyle ||
		(resolved.Style == CurrencyStyle && ecma402.ResolvedScalarValue(resolved.CurrencyDisplay) == CurrencyDisplayName)
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

func collapseRangeEndpointParts(startParts, endParts []Part) (prefix, start, end, suffix []Part) {
	prefixCount := commonCollapsiblePrefixCount(startParts, endParts)
	suffixCount := compatibleCollapsibleSuffixCount(startParts[prefixCount:], endParts[prefixCount:])
	prefixWidth := partWidth(startParts[:prefixCount])
	suffixWidth := partWidth(startParts[len(startParts)-suffixCount:])

	sharedPrefix := startParts[:prefixCount]
	sharedSuffix := endParts[len(endParts)-suffixCount:]
	// Paired patterns wrap both endpoints with one logical affix pair. Signs
	// share only with a paired percent suffix; units and currency names keep an
	// identical sign on each endpoint.
	if prefixCount > 0 && suffixCount > 0 && prefixWidth+suffixWidth > 1 &&
		(!containsSignPart(sharedPrefix) || containsPartType(sharedSuffix, PartPercentSign)) {
		return sharedPrefix, startParts[prefixCount : len(startParts)-suffixCount], endParts[prefixCount : len(endParts)-suffixCount], sharedSuffix
	}
	if suffixWidth > 1 || containsPartType(sharedSuffix, PartUnit) {
		return nil, startParts[:len(startParts)-suffixCount], endParts[:len(endParts)-suffixCount], endParts[len(endParts)-suffixCount:]
	}
	if prefixWidth > 1 {
		return startParts[:prefixCount], startParts[prefixCount:], endParts[prefixCount:], nil
	}
	return nil, startParts, endParts, nil
}

func containsSignPart(parts []Part) bool {
	return containsPartType(parts, PartMinusSign) || containsPartType(parts, PartPlusSign)
}

func containsPartType(parts []Part, typ PartType) bool {
	for _, part := range parts {
		if part.Type == typ {
			return true
		}
	}
	return false
}

func compatibleCollapsibleSuffixCount(startParts, endParts []Part) int {
	count := 0
	for count < len(startParts) && count < len(endParts) {
		start := startParts[len(startParts)-1-count]
		end := endParts[len(endParts)-1-count]
		if !isCollapsibleRangePart(start.Type) || start.Type != end.Type {
			break
		}
		if start.Value != end.Value && start.Type != PartUnit && start.Type != PartCurrency {
			break
		}
		count++
	}
	return count
}

func commonCollapsiblePrefixCount(startParts, endParts []Part) int {
	count := 0
	for count < len(startParts) && count < len(endParts) {
		start := startParts[count]
		end := endParts[count]
		if !isCollapsibleRangePart(start.Type) || start != end {
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
