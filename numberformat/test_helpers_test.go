package numberformat

import (
	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
)

func (f *NumberFormat) formatValue(v any) string {
	return f.Format(anyValue(v))
}

func (f *NumberFormat) formatToPartsValue(v any) []Part {
	return f.FormatToParts(anyValue(v))
}

func (f *NumberFormat) formatRangeValue(start, end any) (string, error) {
	return f.FormatRange(anyValue(start), anyValue(end))
}

func (f *NumberFormat) formatRangeToPartsValue(start, end any) ([]RangePart, error) {
	return f.FormatRangeToParts(anyValue(start), anyValue(end))
}

func anyValue(v any) Value {
	if v == nil {
		return Value{numeric: ecma402.DecimalNumericValue(decimal.NaNValue)}
	}
	d, err := decimal.ToIntlMathematicalValue(v)
	if err != nil {
		return Value{numeric: ecma402.DecimalNumericValue(decimal.NaNValue)}
	}
	return Value{numeric: ecma402.DecimalNumericValue(d)}
}
