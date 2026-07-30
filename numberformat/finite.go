package numberformat

import (
	"strings"

	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/ecma402"
)

type digitGrouping struct {
	primary   int
	secondary int
}

func groupingForNumberFormat(loc cldrnumber.Locale, opts ResolvedOptions) digitGrouping {
	pattern := loc.DecimalPattern(opts.NumberingSystem)
	switch opts.Style {
	case CurrencyStyle:
		if ecma402.ResolvedScalarValue(opts.CurrencyDisplay) != CurrencyDisplayName {
			pattern = loc.CurrencyPattern(opts.NumberingSystem, string(ecma402.ResolvedScalarValue(opts.CurrencySign)))
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

func shouldUseGrouping(policy UseGrouping, formatted string) bool {
	return shouldUseGroupingDigits(policy, integerDigitCount(formatted))
}

func shouldUseGroupingDigits(policy UseGrouping, digits int) bool {
	switch policy {
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

func groupingBounds(digits int, grouping digitGrouping) (int, int) {
	lastGroup := digits - grouping.primary
	firstGroup := lastGroup % grouping.secondary
	if firstGroup == 0 {
		firstGroup = grouping.secondary
	}
	return firstGroup, lastGroup
}
