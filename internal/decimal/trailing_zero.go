package decimal

import "strings"

type TrailingZeroDisplay int

const (
	TrailingZeroAuto TrailingZeroDisplay = iota
	TrailingZeroStripIfInteger
)

func ApplyTrailingZeroDisplay(formatted string, isInteger bool, display TrailingZeroDisplay) string {
	if display != TrailingZeroStripIfInteger || !isInteger {
		return formatted
	}
	before, _, ok := strings.Cut(formatted, ".")
	if !ok {
		return formatted
	}
	return before
}
