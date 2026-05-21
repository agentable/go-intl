package numberformat

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/agentable/go-intl/internal/decimal"
)

// FormatRange formats a numeric range.
func (f *NumberFormat) FormatRange(start, end Value) string {
	return joinRangeParts(f.FormatRangeToParts(start, end))
}

// FormatRangeToParts formats a numeric range into ECMA-402 range parts.
func (f *NumberFormat) FormatRangeToParts(start, end Value) []RangePart {
	return f.formatRangeToPartsDecimal(start.decimal, end.decimal)
}

func (f *NumberFormat) formatRangeValue(start, end any) string {
	return f.FormatRange(anyValue(start), anyValue(end))
}

func (f *NumberFormat) formatRangeToPartsValue(start, end any) []RangePart {
	return f.FormatRangeToParts(anyValue(start), anyValue(end))
}

func (f *NumberFormat) formatRangeToPartsDecimal(start, end decimal.Decimal) []RangePart {
	if !start.IsFinite() || !end.IsFinite() {
		return nil
	}
	startParts := f.formatDecimalToParts(start)
	endParts := f.formatDecimalToParts(end)
	if f.roundedRangeEqual(start, end) || slices.Equal(startParts, endParts) {
		out := make([]RangePart, 0, len(startParts)+1)
		out = append(out, RangePart{Type: PartApproximatelySign, Value: f.symbols().ApproxSign, Source: SourceShared})
		out = append(out, rangeParts(startParts, SourceShared)...)
		return out
	}
	out := rangeParts(startParts, SourceStartRange)
	out = append(out, RangePart{Type: PartLiteral, Value: "–", Source: SourceShared})
	out = append(out, rangeParts(endParts, SourceEndRange)...)
	return collapseRangeParts(out)
}

func (f *NumberFormat) roundedRangeEqual(start, end decimal.Decimal) bool {
	startRounded, ok := f.roundedForRange(start)
	if !ok {
		return false
	}
	endRounded, ok := f.roundedForRange(end)
	return ok && startRounded.Cmp(endRounded) == 0
}

func rangeParts(parts []Part, source RangeSource) []RangePart {
	out := make([]RangePart, len(parts))
	for i, part := range parts {
		out[i] = RangePart{Type: part.Type, Value: part.Value, Source: source}
	}
	return out
}

func joinRangeParts(parts []RangePart) string {
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

func collapseRangeParts(parts []RangePart) []RangePart {
	separator := rangeSeparatorIndex(parts)
	if separator < 0 {
		return parts
	}
	if count := collapsibleSuffixCount(parts[:separator]); rangePartWidth(parts[separator-count:separator]) > 1 {
		return append(parts[:separator-count], parts[separator:]...)
	}
	if count := collapsiblePrefixCount(parts[separator+1:]); rangePartWidth(parts[separator+1:separator+1+count]) > 1 {
		return append(parts[:separator+1], parts[separator+1+count:]...)
	}
	return parts
}

func rangeSeparatorIndex(parts []RangePart) int {
	for i, part := range parts {
		if part.Type == PartLiteral && part.Source == SourceShared && part.Value == "–" {
			return i
		}
	}
	return -1
}

func collapsibleSuffixCount(parts []RangePart) int {
	count := 0
	for i := len(parts) - 1; i >= 0; i-- {
		if !isCollapsibleRangePart(parts[i].Type) {
			break
		}
		count++
	}
	return count
}

func collapsiblePrefixCount(parts []RangePart) int {
	count := 0
	for _, part := range parts {
		if !isCollapsibleRangePart(part.Type) {
			break
		}
		count++
	}
	return count
}

func rangePartWidth(parts []RangePart) int {
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
