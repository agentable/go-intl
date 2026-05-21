package numberformat

import (
	"strconv"
	"strings"
)

func (f *NumberFormat) formatFastInt64(v int64) (string, bool) {
	if !f.canUseDecimalIntegerFastPath() {
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
	if !f.canUseDecimalIntegerFastPath() {
		return "", false
	}
	var scratch [20]byte
	raw := strconv.AppendUint(scratch[:0], v, 10)
	return f.formatFastInteger(raw, false)
}

func (f *NumberFormat) canUseDecimalIntegerFastPath() bool {
	return f.resolved.Style == DecimalStyle &&
		f.resolved.Notation == StandardNotation &&
		f.resolved.SignDisplay == AutoSignDisplay &&
		f.resolved.NumberingSystem == "latn" &&
		f.digits.minInt == 1 &&
		f.digits.minFrac == 0 &&
		f.digits.maxFrac == 3 &&
		f.digits.minSig == 0 &&
		f.digits.maxSig == 0 &&
		f.resolved.RoundingIncrement == 1 &&
		f.resolved.RoundingMode == HalfExpandRoundingMode &&
		f.resolved.RoundingPriority == AutoRoundingPriority &&
		f.resolved.TrailingZeroDisplay == AutoTrailingZeroDisplay
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
