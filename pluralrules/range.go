package pluralrules

import (
	"github.com/agentable/go-intl/internal/cldr/plural"
	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
	pluralop "github.com/agentable/go-intl/internal/plural"
)

// SelectRange returns the plural category for a numeric range.
func (f *PluralRules) SelectRange(start, end Value) (Category, error) {
	startNumeric := start.numeric
	endNumeric := end.numeric
	if f.integerOperands {
		switch {
		case startNumeric.Kind == ecma402.NumericValueInt64 && endNumeric.Kind == ecma402.NumericValueInt64:
			startCategory := selectInteger(startNumeric.Int64, f.rule)
			if ecma402.Int64Magnitude(startNumeric.Int64) == ecma402.Int64Magnitude(endNumeric.Int64) {
				return startCategory, nil
			}
			return f.selectRangeCategories(startCategory, selectInteger(endNumeric.Int64, f.rule)), nil
		case startNumeric.Kind == ecma402.NumericValueUint64 && endNumeric.Kind == ecma402.NumericValueUint64:
			startCategory := selectUnsignedInteger(startNumeric.Uint64, f.rule)
			if startNumeric.Uint64 == endNumeric.Uint64 {
				return startCategory, nil
			}
			return f.selectRangeCategories(startCategory, selectUnsignedInteger(endNumeric.Uint64, f.rule)), nil
		}
	}
	if startNumeric.Decimal.IsNaN() {
		return Other, invalidRangeValue("start", startNumeric.Decimal.String(), f.resolved.Locale.String(), decimal.ErrInvalidDecimal)
	}
	if endNumeric.Decimal.IsNaN() {
		return Other, invalidRangeValue("end", endNumeric.Decimal.String(), f.resolved.Locale.String(), decimal.ErrInvalidDecimal)
	}
	startFormatted, startCategory := f.resolveDecimal(startNumeric.Decimal)
	endFormatted, endCategory := f.resolveDecimal(endNumeric.Decimal)
	if startFormatted == endFormatted {
		return startCategory, nil
	}
	return f.selectRangeCategories(startCategory, endCategory), nil
}

func (f *PluralRules) selectRangeCategories(startCategory, endCategory Category) Category {
	return Category(plural.ResolveRange(
		f.dataLocale,
		pluralop.Category(startCategory),
		pluralop.Category(endCategory),
	))
}

func invalidRangeValue(name, value, loc string, err error) error {
	return ecma402.InvalidValueErrorExpected(pluralRulesOwner, name, value, loc, "a numeric value other than NaN", err)
}
