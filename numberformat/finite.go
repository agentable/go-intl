package numberformat

import (
	"strings"

	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/decimal"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
)

func (f *NumberFormat) formatFiniteResult(d decimal.Decimal) ecma402nf.FormattedNumeric {
	return ecma402nf.FormatNumericToString(d, ecma402nf.DigitOptions{
		MinimumIntegerDigits:     f.digits.minInt,
		MinimumFractionDigits:    f.digits.minFrac,
		MaximumFractionDigits:    f.digits.maxFrac,
		MinimumSignificantDigits: f.digits.minSig,
		MaximumSignificantDigits: f.digits.maxSig,
		RoundingIncrement:        f.resolved.RoundingIncrement,
		RoundingMode:             string(f.resolved.RoundingMode),
		RoundingPriority:         string(f.resolved.RoundingPriority),
		TrailingZeroDisplay:      string(f.resolved.TrailingZeroDisplay),
	})
}

type digitGrouping struct {
	primary   int
	secondary int
}

func groupingForNumberFormat(loc cldrnumber.Locale, opts ResolvedOptions) digitGrouping {
	pattern := loc.DecimalPattern(opts.NumberingSystem)
	switch opts.Style {
	case CurrencyStyle:
		if opts.CurrencyDisplay != CurrencyDisplayName {
			pattern = loc.CurrencyPattern(opts.NumberingSystem, string(opts.CurrencySign))
		}
	case PercentStyle:
		pattern = loc.PercentPattern(opts.NumberingSystem)
	case DecimalStyle, UnitStyle:
	default:
	}
	return groupingFromPattern(pattern)
}

func groupingFromPattern(pattern string) digitGrouping {
	grouping := digitGrouping{primary: 3, secondary: 3}
	positive, _, _ := strings.Cut(pattern, ";")
	start, end := numberPatternBounds(positive)
	if start < 0 {
		return grouping
	}
	numberPattern := positive[start:end]
	integerPattern, _, _ := strings.Cut(numberPattern, ".")
	patternGroups := strings.Split(integerPattern, ",")
	if len(patternGroups) > 1 {
		grouping.primary = len(patternGroups[len(patternGroups)-1])
	}
	if len(patternGroups) > 2 {
		grouping.secondary = len(patternGroups[len(patternGroups)-2])
	}
	if grouping.primary <= 0 {
		grouping.primary = 3
	}
	if grouping.secondary <= 0 {
		grouping.secondary = grouping.primary
	}
	return grouping
}

func groupDecimal(s string, grouping digitGrouping) string {
	original := s
	sign := ""
	if rest, ok := strings.CutPrefix(s, "-"); ok {
		sign = "-"
		s = rest
	}
	integer, fraction, hasFraction := strings.Cut(s, ".")
	if !needsGrouping(len(integer), grouping) {
		return original
	}
	grouped := groupInteger(integer, grouping)
	if !hasFraction {
		return sign + grouped
	}
	return joinSignedDecimalParts(sign, grouped, fraction)
}

func joinDecimalParts(integer, fraction string) string {
	var b strings.Builder
	b.Grow(len(integer) + 1 + len(fraction))
	b.WriteString(integer)
	b.WriteByte('.')
	b.WriteString(fraction)
	return b.String()
}

func joinSignedDecimalParts(sign, integer, fraction string) string {
	var b strings.Builder
	b.Grow(len(sign) + len(integer) + 1 + len(fraction))
	b.WriteString(sign)
	b.WriteString(integer)
	b.WriteByte('.')
	b.WriteString(fraction)
	return b.String()
}

func groupInteger(integer string, grouping digitGrouping) string {
	if !needsGrouping(len(integer), grouping) {
		return integer
	}

	var b strings.Builder
	b.Grow(len(integer) + groupSeparatorCount(len(integer), grouping))
	writeGroupedString(&b, integer, grouping, ",")
	return b.String()
}

func (f *NumberFormat) useGrouping(formatted string) bool {
	switch f.resolved.UseGrouping {
	case UseGroupingFalse:
		return false
	case UseGroupingMin2:
		return integerDigitCount(formatted) >= 5
	case UseGroupingAuto, UseGroupingAlways:
	}
	return true
}

func (f *NumberFormat) useGroupingDigits(digits int) bool {
	switch f.resolved.UseGrouping {
	case UseGroupingFalse:
		return false
	case UseGroupingMin2:
		return digits >= 5
	case UseGroupingAuto, UseGroupingAlways:
	}
	return true
}

func integerDigitCount(formatted string) int {
	formatted = strings.TrimPrefix(formatted, "-")
	integer, _, _ := strings.Cut(formatted, ".")
	return len(integer)
}

func needsGrouping(digits int, grouping digitGrouping) bool {
	return digits > grouping.primary
}

func groupSeparatorCount(digits int, grouping digitGrouping) int {
	if !needsGrouping(digits, grouping) {
		return 0
	}
	remaining := digits - grouping.primary
	return (remaining + grouping.secondary - 1) / grouping.secondary
}

func writeGroupedString(b *strings.Builder, digits string, grouping digitGrouping, separator string) {
	firstGroup, lastGroup := groupingBounds(len(digits), grouping)
	b.WriteString(digits[:firstGroup])
	for start := firstGroup; start < lastGroup; start += grouping.secondary {
		b.WriteString(separator)
		b.WriteString(digits[start : start+grouping.secondary])
	}
	b.WriteString(separator)
	b.WriteString(digits[lastGroup:])
}

func writeGroupedBytes(b *strings.Builder, digits []byte, grouping digitGrouping, separator string) {
	firstGroup, lastGroup := groupingBounds(len(digits), grouping)
	b.Write(digits[:firstGroup])
	for start := firstGroup; start < lastGroup; start += grouping.secondary {
		b.WriteString(separator)
		b.Write(digits[start : start+grouping.secondary])
	}
	b.WriteString(separator)
	b.Write(digits[lastGroup:])
}

func groupingBounds(digits int, grouping digitGrouping) (int, int) {
	lastGroup := digits - grouping.primary
	firstGroup := lastGroup % grouping.secondary
	if firstGroup == 0 {
		firstGroup = grouping.secondary
	}
	return firstGroup, lastGroup
}
