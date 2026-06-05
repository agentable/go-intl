// Hand-written accessor layer for the unit domain. The query semantics mirror
// the legacy root cldr unit accessors exactly, so formatter output is
// byte-for-byte unchanged.

package unit

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
	i := searchUnitPattern(unitPatterns, key)
	if i < len(unitPatterns) && unitPatterns[i].key == key {
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
	i := searchCompoundUnitPattern(compoundUnitRows, key)
	if i < len(compoundUnitRows) && compoundUnitRows[i].key == key {
		return compoundUnitRows[i].pattern
	}
	return ""
}

// SupportedLocales returns the unit-supported locale tags in sorted-locale
// order. It reads only the narrow supported blob and never triggers the pattern
// or compound blob decode.
func SupportedLocales() []string {
	unitSupportedOnce.Do(loadUnitSupported)
	tags := make([]string, len(unitSupportedTags))
	copy(tags, unitSupportedTags)
	return tags
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
	return unitNameIDs[unit]
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

func searchUnitPattern(rows []unitPatternRecord, key uint32) int {
	low, high := 0, len(rows)
	for low < high {
		mid := int(uint(low+high) >> 1)
		if rows[mid].key < key {
			low = mid + 1
		} else {
			high = mid
		}
	}
	return low
}

func searchCompoundUnitPattern(rows []compoundUnitPatternRecord, key uint32) int {
	low, high := 0, len(rows)
	for low < high {
		mid := int(uint(low+high) >> 1)
		if rows[mid].key < key {
			low = mid + 1
		} else {
			high = mid
		}
	}
	return low
}
