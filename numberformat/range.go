package numberformat

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/agentable/go-intl/internal/decimal"
)

func (f *NumberFormat) FormatRangeInt(start, end int) string {
	return f.FormatRangeInt64(int64(start), int64(end))
}

func (f *NumberFormat) FormatRangeInt64(start, end int64) string {
	return joinRangeParts(f.formatRangeToPartsDecimal(decimal.FromInt64(start), decimal.FromInt64(end)))
}

func (f *NumberFormat) FormatRangeUint(start, end uint) string {
	return f.FormatRangeUint64(uint64(start), uint64(end))
}

func (f *NumberFormat) FormatRangeUint64(start, end uint64) string {
	startDecimal := decimal.New(false, new(big.Int).SetUint64(start), 0)
	endDecimal := decimal.New(false, new(big.Int).SetUint64(end), 0)
	return joinRangeParts(f.formatRangeToPartsDecimal(startDecimal, endDecimal))
}

func (f *NumberFormat) FormatRangeFloat64(start, end float64) (string, error) {
	parts, err := f.FormatRangeFloat64ToParts(start, end)
	if err != nil {
		return "", err
	}
	return joinRangeParts(parts), nil
}

func (f *NumberFormat) FormatRangeDecimal(start, end string) (string, error) {
	parts, err := f.FormatRangeDecimalToParts(start, end)
	if err != nil {
		return "", err
	}
	return joinRangeParts(parts), nil
}

func (f *NumberFormat) FormatRangeIntToParts(start, end int) []RangePart {
	return f.FormatRangeInt64ToParts(int64(start), int64(end))
}

func (f *NumberFormat) FormatRangeInt64ToParts(start, end int64) []RangePart {
	return f.formatRangeToPartsDecimal(decimal.FromInt64(start), decimal.FromInt64(end))
}

func (f *NumberFormat) FormatRangeUintToParts(start, end uint) []RangePart {
	return f.FormatRangeUint64ToParts(uint64(start), uint64(end))
}

func (f *NumberFormat) FormatRangeUint64ToParts(start, end uint64) []RangePart {
	startDecimal := decimal.New(false, new(big.Int).SetUint64(start), 0)
	endDecimal := decimal.New(false, new(big.Int).SetUint64(end), 0)
	return f.formatRangeToPartsDecimal(startDecimal, endDecimal)
}

func (f *NumberFormat) FormatRangeFloat64ToParts(start, end float64) ([]RangePart, error) {
	return f.formatRangeToPartsDecimalChecked(decimal.FromFloat64(start), decimal.FromFloat64(end))
}

func (f *NumberFormat) FormatRangeDecimalToParts(start, end string) ([]RangePart, error) {
	startDecimal, err := parseDecimalValue(start)
	if err != nil {
		return nil, err
	}
	endDecimal, err := parseDecimalValue(end)
	if err != nil {
		return nil, fmt.Errorf("numberformat: invalid range end: %w", err)
	}
	return f.formatRangeToPartsDecimalChecked(startDecimal, endDecimal)
}

func (f *NumberFormat) formatRangeValue(start, end any) string {
	return joinRangeParts(f.formatRangeToPartsValue(start, end))
}

func (f *NumberFormat) formatRangeToPartsValue(start, end any) []RangePart {
	return f.formatRangeToPartsDecimal(valueDecimal(start), valueDecimal(end))
}

func (f *NumberFormat) formatRangeToPartsDecimal(start, end decimal.Decimal) []RangePart {
	if !start.IsFinite() || !end.IsFinite() {
		return nil
	}
	startParts := f.formatDecimalToParts(start)
	endParts := f.formatDecimalToParts(end)
	if f.roundedRangeEqual(start, end) || equalParts(startParts, endParts) {
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

func (f *NumberFormat) formatRangeToPartsDecimalChecked(start, end decimal.Decimal) ([]RangePart, error) {
	if !start.IsFinite() {
		return nil, fmt.Errorf("numberformat: invalid range start: %w", ErrInvalidValue)
	}
	if !end.IsFinite() {
		return nil, fmt.Errorf("numberformat: invalid range end: %w", ErrInvalidValue)
	}
	return f.formatRangeToPartsDecimal(start, end), nil
}

func (f *NumberFormat) roundedRangeEqual(start, end decimal.Decimal) bool {
	startRounded, ok := f.roundedForRange(start)
	if !ok {
		return false
	}
	endRounded, ok := f.roundedForRange(end)
	return ok && startRounded.Cmp(endRounded) == 0
}

func equalParts(a, b []Part) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func rangeParts(parts []Part, source RangeSource) []RangePart {
	out := make([]RangePart, len(parts))
	for i, part := range parts {
		out[i] = RangePart{Type: part.Type, Value: part.Value, Source: source}
	}
	return out
}

func joinRangeParts(parts []RangePart) string {
	var b strings.Builder
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
		for range part.Value {
			width++
		}
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
