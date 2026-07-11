// Hand-written accessor layer for the unit domain. It exposes unit patterns,
// compound patterns, and the narrow supported-locale index over lazily decoded
// const blobs.

package unit

import (
	"cmp"
	"slices"
)

// UnitPattern returns the pattern for a (locale, unit, width, plural) tuple, or
// "" when the tuple is absent or any component is unrecognized.
func UnitPattern(loc Locale, unit, width, plural string) string {
	unitID := unitNameID(unit)
	widthID := unitWidthID(width)
	pluralID := unitPluralID(plural)
	if unitID == 0 || widthID == 0 || pluralID == 0 {
		return ""
	}
	key := makeUnitPatternKey(loc, unitID, widthID, pluralID)
	unitPatternOnce.Do(loadUnitPatterns)
	if i, ok := slices.BinarySearchFunc(unitPatterns, key, func(row unitPatternRecord, key uint32) int {
		return cmp.Compare(row.key, key)
	}); ok {
		return unitPatterns[i].pattern
	}
	return ""
}

// CompoundUnitPattern returns the compound pattern for a (locale, width) pair,
// or "" when absent or the width is unrecognized.
func CompoundUnitPattern(loc Locale, width string) string {
	widthID := unitWidthID(width)
	if widthID == 0 {
		return ""
	}
	key := makeCompoundUnitPatternKey(loc, widthID)
	compoundUnitOnce.Do(loadCompoundUnits)
	if i, ok := slices.BinarySearchFunc(compoundUnitRows, key, func(row compoundUnitPatternRecord, key uint32) int {
		return cmp.Compare(row.key, key)
	}); ok {
		return compoundUnitRows[i].pattern
	}
	return ""
}

// SupportedLocales returns the unit-supported locale tags in sorted-locale
// order. It reads only the narrow supported blob and never triggers the pattern
// or compound blob decode.
func SupportedLocales() []string {
	return supported.Get()
}

func makeUnitPatternKey(loc Locale, unitID, width, plural uint32) uint32 {
	return uint32(loc)<<16 | unitID<<8 | width<<4 | plural
}

func makeCompoundUnitPatternKey(loc Locale, width uint32) uint32 {
	return uint32(loc)<<4 | width
}

// unitNameID resolves a unit name to its 1-based id, or 0 when unknown. The id
// table is decoded with the pattern blob, so a lookup gates that decode.
func unitNameID(unit string) uint32 {
	unitPatternOnce.Do(loadUnitPatterns)
	id, ok := unitNameIDs[unit]
	if !ok {
		return 0
	}
	return id
}

func unitWidthID(width string) uint32 {
	switch width {
	case "", "long":
		return uint32(unitWidthLong)
	case "narrow":
		return uint32(unitWidthNarrow)
	case "short":
		return uint32(unitWidthShort)
	default:
		return 0
	}
}

func unitPluralID(plural string) uint32 {
	switch plural {
	case "few":
		return uint32(unitPluralFew)
	case "many":
		return uint32(unitPluralMany)
	case "one":
		return uint32(unitPluralOne)
	case "", "other":
		return uint32(unitPluralOther)
	case "two":
		return uint32(unitPluralTwo)
	case "zero":
		return uint32(unitPluralZero)
	default:
		return 0
	}
}
