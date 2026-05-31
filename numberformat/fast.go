package numberformat

import (
	"strconv"
	"strings"
)

func (f *NumberFormat) formatFastInt64(v int64) (string, bool) {
	if !f.decimalIntegerFastPath {
		return "", false
	}
	var scratch [20]byte
	raw := strconv.AppendInt(scratch[:0], v, 10)
	negative := len(raw) > 0 && raw[0] == '-'
	if negative {
		raw = raw[1:]
	}
	return f.formatFastInteger(raw, negative)
}

func (f *NumberFormat) formatFastUint64(v uint64) (string, bool) {
	if !f.decimalIntegerFastPath {
		return "", false
	}
	var scratch [20]byte
	raw := strconv.AppendUint(scratch[:0], v, 10)
	return f.formatFastInteger(raw, false)
}

func canUseDecimalIntegerFastPath(resolved ResolvedOptions, digits digitState) bool {
	return resolved.Style == DecimalStyle &&
		resolved.Notation == StandardNotation &&
		resolved.SignDisplay == AutoSignDisplay &&
		resolved.NumberingSystem == "latn" &&
		digits.minInt == 1 &&
		digits.minFrac == 0 &&
		digits.maxFrac == 3 &&
		digits.minSig == 0 &&
		digits.maxSig == 0 &&
		resolved.RoundingIncrement == 1 &&
		resolved.RoundingMode == HalfExpandRoundingMode &&
		resolved.RoundingPriority == AutoRoundingPriority &&
		resolved.TrailingZeroDisplay == AutoTrailingZeroDisplay
}

func (f *NumberFormat) formatFastInteger(digits []byte, negative bool) (string, bool) {
	symbols := f.symbols()
	grouped := f.useGroupingDigits(len(digits)) && needsGrouping(len(digits), f.grouping)
	size := len(digits)
	if negative {
		size += len(symbols.Minus)
	}
	if grouped {
		size += groupSeparatorCount(len(digits), f.grouping) * len(symbols.Group)
	}

	var b strings.Builder
	b.Grow(size)
	if negative {
		b.WriteString(symbols.Minus)
	}
	if grouped {
		writeGroupedBytes(&b, digits, f.grouping, symbols.Group)
	} else {
		b.Write(digits)
	}
	return b.String(), true
}
