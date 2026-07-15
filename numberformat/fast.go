package numberformat

import (
	"strconv"
	"strings"

	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/decimal"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
	"github.com/agentable/go-intl/internal/numbering"
)

type integerFastPathState struct {
	enabled     bool
	useGrouping UseGrouping
	symbols     cldrnumber.NumberSymbols
	grouping    digitGrouping
}

func formatFastInt64(v int64, state integerFastPathState) (string, bool) {
	if !state.enabled {
		return "", false
	}
	var scratch [20]byte
	raw := strconv.AppendInt(scratch[:0], v, 10)
	negative := len(raw) > 0 && raw[0] == '-'
	if negative {
		raw = raw[1:]
	}
	return formatFastInteger(raw, negative, state.useGrouping, state.symbols, state.grouping)
}

func formatFastUint64(v uint64, state integerFastPathState) (string, bool) {
	if !state.enabled {
		return "", false
	}
	var scratch [20]byte
	raw := strconv.AppendUint(scratch[:0], v, 10)
	return formatFastInteger(raw, false, state.useGrouping, state.symbols, state.grouping)
}

func canUseDecimalIntegerFastPath(resolved ResolvedOptions, digits ecma402nf.ResolvedDigitOptions) bool {
	return resolved.Style == DecimalStyle &&
		resolved.Notation == StandardNotation &&
		resolved.SignDisplay == AutoSignDisplay &&
		resolved.NumberingSystem == numbering.DefaultNumberingSystem &&
		digits.MinimumIntegerDigits == 1 &&
		digits.MinimumFractionDigits == 0 &&
		digits.MaximumFractionDigits == 3 &&
		digits.RoundingType == ecma402nf.RoundingTypeFractionDigits &&
		digits.RoundingIncrement == 1 &&
		digits.RoundingMode == decimal.RoundHalfExpand &&
		digits.RoundingPriority == string(AutoRoundingPriority) &&
		digits.TrailingZeroDisplay == string(AutoTrailingZeroDisplay)
}

func formatFastInteger(digits []byte, negative bool, useGrouping UseGrouping, symbols cldrnumber.NumberSymbols, grouping digitGrouping) (string, bool) {
	grouped := shouldUseGroupingDigits(useGrouping, len(digits)) && needsGrouping(len(digits), grouping)
	size := len(digits)
	if negative {
		size += len(symbols.Minus)
	}
	if grouped {
		size += groupSeparatorCount(len(digits), grouping) * len(symbols.Group)
	}

	var b strings.Builder
	b.Grow(size)
	if negative {
		b.WriteString(symbols.Minus)
	}
	if grouped {
		writeGroupedBytes(&b, digits, grouping, symbols.Group)
	} else {
		b.Write(digits)
	}
	return b.String(), true
}
