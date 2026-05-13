package numberformat

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/internal/decimal"
)

func (f *NumberFormat) formatFast(v any) (string, bool) {
	if !f.canUseFastPath() {
		return "", false
	}
	if f.resolved.Notation != "standard" || f.resolved.Style == "percent" || f.resolved.Style == "unit" {
		return "", false
	}
	formatted, ok := f.formatFiniteFast(v)
	if !ok {
		return "", false
	}
	if f.useGrouping(formatted) {
		formatted = groupDecimal(formatted, f.grouping)
	}
	formatted = localizeNumberStringForSystem(formatted, f.symbols(), f.resolved.NumberingSystem)
	if f.resolved.Style == "currency" {
		return f.formatCurrencyFast(formatted, valueDecimal(v)), true
	}
	return formatted, true
}

func (f *NumberFormat) formatFastInt64(v int64) (string, bool) {
	formatted, ok := f.formatFiniteFastInt64(v)
	if !ok || !f.canUseFullFastPath() {
		return "", false
	}
	if f.useGrouping(formatted) {
		formatted = groupDecimal(formatted, f.grouping)
	}
	formatted = localizeNumberStringForSystem(formatted, f.symbols(), f.resolved.NumberingSystem)
	if f.resolved.Style == "currency" {
		return f.formatCurrencyFast(formatted, decimal.FromInt64(v)), true
	}
	return formatted, true
}

func (f *NumberFormat) canUseFastPath() bool {
	return f.canUseFullFastPath() && f.resolved.Style != "unit"
}

func (f *NumberFormat) canUseFullFastPath() bool {
	return f.resolved.SignDisplay == "auto" &&
		f.resolved.Notation == "standard" &&
		f.resolved.Style != "percent" &&
		f.resolved.Style != "unit" &&
		f.resolved.MinimumIntegerDigits == 1 &&
		f.resolved.MaximumSignificantDigits == 0 &&
		f.resolved.RoundingIncrement == 1 &&
		f.resolved.RoundingMode == "halfExpand" &&
		f.resolved.RoundingPriority == "auto" &&
		f.resolved.TrailingZeroDisplay == "auto"
}

func (f *NumberFormat) formatCurrencyFast(formatted string, d decimal.Decimal) string {
	if f.resolved.CurrencyDisplay == "name" {
		plural := pluralCategory(f.resolved.Locale.Tag().String(), strings.TrimPrefix(d.String(), "-"))
		return formatted + " " + f.currencyDisplay(plural)
	}
	sign := Part{}
	symbols := f.symbols()
	if rest, ok := strings.CutPrefix(formatted, symbols.Minus); ok {
		sign = Part{Type: PartMinusSign, Value: symbols.Minus}
		formatted = rest
	}
	pattern, consumedSign := f.currencyPattern(sign)
	currency := f.currencyDisplay("other")
	out := formatNumberPatternString(pattern, formatted, currency)
	if sign.Type != "" && !consumedSign {
		return sign.Value + out
	}
	return out
}

func formatNumberPatternString(pattern, formatted, affix string) string {
	start, end := numberPatternBounds(pattern)
	if start < 0 {
		return affix + formatted
	}
	prefix := strings.ReplaceAll(pattern[:start], "¤", affix)
	suffix := strings.ReplaceAll(pattern[end:], "¤", affix)
	return prefix + formatted + suffix
}

func (f *NumberFormat) formatFiniteFast(v any) (string, bool) {
	if value, ok := int64Value(v); ok {
		return f.formatFiniteFastInt64(value)
	}
	if value, ok := uint64Value(v); ok {
		return f.formatFiniteFastUint64(value)
	}
	return "", false
}

func (f *NumberFormat) formatFiniteFastInt64(value int64) (string, bool) {
	formatted := strconv.FormatInt(value, 10)
	return f.padFastFraction(formatted), true
}

func (f *NumberFormat) formatFiniteFastUint64(value uint64) (string, bool) {
	formatted := strconv.FormatUint(value, 10)
	return f.padFastFraction(formatted), true
}

func (f *NumberFormat) padFastFraction(formatted string) string {
	digits := f.resolved.MaximumFractionDigits
	if digits > 0 && (f.resolved.MinimumFractionDigits > 0 || digits == f.resolved.MinimumFractionDigits) {
		formatted += "." + strings.Repeat("0", digits)
	}
	return formatted
}

func int64Value(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	}
	return 0, false
}

func uint64Value(v any) (uint64, bool) {
	switch n := v.(type) {
	case uint:
		return uint64(n), true
	case uint8:
		return uint64(n), true
	case uint16:
		return uint64(n), true
	case uint32:
		return uint64(n), true
	case uint64:
		return n, true
	case uintptr:
		return uint64(n), true
	case *big.Int:
		if n == nil || !n.IsUint64() {
			return 0, false
		}
		return n.Uint64(), true
	case big.Int:
		if !n.IsUint64() {
			return 0, false
		}
		return n.Uint64(), true
	}
	return 0, false
}
