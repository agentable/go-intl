package numberformat

import (
	"strings"

	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/ecma402"
)

type digitGrouping struct {
	primary   int
	secondary int
	minimum   int
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
	grouping := groupingFromPattern(pattern)
	switch opts.UseGrouping {
	case UseGroupingFalse:
		grouping.minimum = 0
	case UseGroupingAuto:
		grouping.minimum = loc.MinimumGroupingDigits()
	case UseGroupingMin2:
		grouping.minimum = max(2, loc.MinimumGroupingDigits())
	case UseGroupingAlways:
	}
	return grouping
}

func groupingFromPattern(pattern string) digitGrouping {
	grouping := digitGrouping{primary: 3, secondary: 3, minimum: 1}
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

func (g digitGrouping) separatorCount(digits int) int {
	if g.minimum <= 0 || digits < g.primary+g.minimum {
		return 0
	}
	remaining := digits - g.primary
	return (remaining + g.secondary - 1) / g.secondary
}

func appendDecimalParts(parts []Part, formatted string, grouping digitGrouping, symbols cldrnumber.NumberSymbols) []Part {
	digits, negative := strings.CutPrefix(formatted, "-")
	integer, fraction, hasFraction := strings.Cut(digits, ".")
	groupCount := grouping.separatorCount(len(integer))
	partCount := groupCount*2 + 1
	if negative {
		partCount++
	}
	if hasFraction {
		partCount += 2
	}
	partStart := len(parts)
	partEnd := partStart + partCount
	if cap(parts) < partEnd {
		next := make([]Part, partEnd)
		copy(next, parts)
		parts = next
	} else {
		parts = parts[:partEnd]
	}

	nextPart := partEnd - 1
	if hasFraction {
		parts[nextPart] = Part{Type: PartFraction, Value: fraction}
		parts[nextPart-1] = Part{Type: PartDecimal, Value: symbols.Decimal}
		nextPart -= 2
	}
	// Group widths are counted from the rightmost integer group.
	width := grouping.primary
	for range groupCount {
		start := len(integer) - width
		parts[nextPart] = Part{Type: PartInteger, Value: integer[start:]}
		parts[nextPart-1] = Part{Type: PartGroup, Value: symbols.Group}
		integer = integer[:start]
		nextPart -= 2
		width = grouping.secondary
	}
	parts[nextPart] = Part{Type: PartInteger, Value: integer}
	if negative {
		parts[partStart] = Part{Type: PartMinusSign, Value: symbols.Minus}
	}
	return parts
}
