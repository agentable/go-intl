package numberformat

import (
	"strings"

	"github.com/agentable/go-intl/internal/cldr"
	"github.com/agentable/go-intl/internal/decimal"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
)

func (f *NumberFormat) formatFinite(d decimal.Decimal) string {
	return f.formatFiniteResult(d).Formatted
}

func (f *NumberFormat) formatFiniteResult(d decimal.Decimal) ecma402nf.FormattedNumeric {
	return ecma402nf.FormatNumericToString(d, ecma402nf.DigitOptions{
		MinimumIntegerDigits:     f.resolved.MinimumIntegerDigits,
		MinimumFractionDigits:    f.resolved.MinimumFractionDigits,
		MaximumFractionDigits:    f.resolved.MaximumFractionDigits,
		MinimumSignificantDigits: f.resolved.MinimumSignificantDigits,
		MaximumSignificantDigits: f.resolved.MaximumSignificantDigits,
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

func groupingForNumberFormat(loc cldr.Locale, opts ResolvedOptions) digitGrouping {
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
	sign := ""
	if rest, ok := strings.CutPrefix(s, "-"); ok {
		sign = "-"
		s = rest
	}
	integer, fraction, hasFraction := strings.Cut(s, ".")
	grouped := groupInteger(integer, grouping)
	if !hasFraction {
		return sign + grouped
	}
	return sign + grouped + "." + fraction
}

func groupInteger(integer string, grouping digitGrouping) string {
	i := len(integer) - grouping.primary
	if i <= 0 {
		return integer
	}
	groups := []string{integer[i:]}
	for i -= grouping.secondary; i > 0; i -= grouping.secondary {
		groups = append(groups, integer[i:i+grouping.secondary])
	}
	groups = append(groups, integer[:i+grouping.secondary])

	size := len(integer) + len(groups) - 1
	var b strings.Builder
	b.Grow(size)
	for i := len(groups) - 1; i >= 0; i-- {
		if i != len(groups)-1 {
			b.WriteByte(',')
		}
		b.WriteString(groups[i])
	}
	return b.String()
}

func (f *NumberFormat) useGrouping(formatted string) bool {
	switch f.resolved.UseGrouping {
	case UseGroupingFalse:
		return false
	case UseGroupingMin2:
		return integerDigitCount(formatted) >= 5
	case UseGroupingAuto, UseGroupingAlways:
		return true
	default:
		return true
	}
}

func integerDigitCount(formatted string) int {
	formatted = strings.TrimPrefix(formatted, "-")
	integer, _, _ := strings.Cut(formatted, ".")
	return len(integer)
}
