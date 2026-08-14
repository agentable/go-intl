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
			if startNumeric.Int64 == endNumeric.Int64 {
				return startCategory, nil
			}
			return selectRangeCategories(startCategory, selectInteger(endNumeric.Int64, f.rule), f), nil
		case startNumeric.Kind == ecma402.NumericValueUint64 && endNumeric.Kind == ecma402.NumericValueUint64:
			startCategory := selectUnsignedInteger(startNumeric.Uint64, f.rule)
			if startNumeric.Uint64 == endNumeric.Uint64 {
				return startCategory, nil
			}
			return selectRangeCategories(startCategory, selectUnsignedInteger(endNumeric.Uint64, f.rule), f), nil
		}
	}
	if startNumeric.Decimal.IsNaN() {
		return Other, invalidRangeValue("start", startNumeric.Decimal.String(), f.resolved.Locale.String(), decimal.ErrInvalidDecimal)
	}
	if endNumeric.Decimal.IsNaN() {
		return Other, invalidRangeValue("end", endNumeric.Decimal.String(), f.resolved.Locale.String(), decimal.ErrInvalidDecimal)
	}
	return selectRangeDecimal(startNumeric.Decimal, endNumeric.Decimal, f), nil
}

func selectRangeDecimal(start, end decimal.Decimal, f *PluralRules) Category {
	notation := f.resolved.Notation
	digitOptions := f.digitOptions
	compact := f.compact
	rule := f.rule
	startFormatted, _, startCategory := resolveDecimal(start, notation, digitOptions, compact, rule)
	endFormatted, _, endCategory := resolveDecimal(end, notation, digitOptions, compact, rule)
	return selectRangeResolved(startFormatted, startCategory, endFormatted, endCategory, f)
}

func selectRangeResolved(startFormatted string, startCategory Category, endFormatted string, endCategory Category, f *PluralRules) Category {
	if startFormatted == endFormatted {
		return startCategory
	}
	return selectRangeCategories(startCategory, endCategory, f)
}

func selectRangeCategories(startCategory Category, endCategory Category, f *PluralRules) Category {
	if f.resolved.Type != Cardinal {
		return endCategory
	}
	return Category(plural.ResolveCardinalRange(
		f.dataLocale,
		pluralop.Category(startCategory),
		pluralop.Category(endCategory),
	))
}

func invalidRangeValue(name, value, loc string, err error) error {
	return ecma402.InvalidValueErrorExpected(pluralRulesOwner, name, value, loc, "a numeric value other than NaN", err)
}
