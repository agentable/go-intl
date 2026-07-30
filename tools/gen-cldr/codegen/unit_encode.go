package codegen

import (
	"fmt"
	"maps"
	"slices"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

const (
	unitPatternLocaleShift = 16
	unitPatternUnitShift   = 8
	unitPatternWidthShift  = 4

	unitPatternUnitIDMax    = 1<<8 - 1
	unitPatternLocaleTagMax = 1 << 16
	unitPatternLocaleIDMax  = unitPatternLocaleTagMax - 1
)

// encodeUnits renders the const-only payload for the unit domain. It emits a
// private _data string table plus independent blobs, each prefixed by its
// record count:
//
//   - _unitPatternBlob:   sorted packed unit-pattern delta keys paired with
//     pattern StringRefs.
//   - _compoundUnitBlob:  sorted packed compound-unit delta keys paired with
//     pattern StringRefs.
//   - _perUnitPatternBlob: sorted packed specialized per-unit delta keys paired
//     with pattern StringRefs.
//   - _unitNameBlob:      the sorted unit-name table (id = index+1), so the
//     decoder can rebuild the name->id map the pattern key packing needs.
//   - _unitSupportedBlob: the unit-supported locale tags in sorted-locale order,
//     a narrow index so SupportedLocales never decodes the pattern blobs.
//
// The delta primitive sorts rows by their packed keys, so the runtime binary
// search follows the payload order without row builders owning wire-order
// knowledge.
func encodeUnits(input RuntimeInput, table *StringTable) ([]byte, error) {
	data := input.Units
	localeIndex := localeIndexMap(input.Locales)
	locales := sortedLocaleKeys(data)
	unitNames := sortedUnitPatternNames(data)
	unitIDs := unitPatternIDs(unitNames)

	// Field-width guards. The unit pattern key packs loc, unit id, width, and
	// plural into a uint32. The runtime Locale is also uint16, so the generated
	// locale registry cannot exceed the locale slot even before key packing.
	if n := len(unitNames); n > unitPatternUnitIDMax {
		return nil, fmt.Errorf("unit name ID table has %d entries; the 8-bit unit field in the pattern key holds at most %d", n, unitPatternUnitIDMax)
	}
	if n := len(input.Locales.Tags); n > unitPatternLocaleTagMax {
		return nil, fmt.Errorf("locale table has %d tags; the 16-bit locale field in the unit pattern key holds at most %d", n, unitPatternLocaleTagMax)
	}

	var patterns blobEncoder
	patternRows := unitPatternRows(data)
	if err := appendUint32DeltaSlice(&patterns, patternRows, func(row unitPatternRow) (uint32, error) {
		return unitPatternKeyValue(localeIndex, unitIDs, row)
	}, func(row unitPatternRow) {
		patterns.appendStringRef(table.Add(row.pattern))
	}); err != nil {
		return nil, err
	}

	var compound blobEncoder
	compoundRows := compoundUnitPatternRows(data)
	if err := appendUint32DeltaSlice(&compound, compoundRows, func(row compoundUnitPatternRow) (uint32, error) {
		return compoundUnitKeyValue(localeIndex, row)
	}, func(row compoundUnitPatternRow) {
		compound.appendStringRef(table.Add(row.pattern))
	}); err != nil {
		return nil, err
	}

	var perUnit blobEncoder
	perUnitRows := perUnitPatternRows(data)
	if err := appendUint32DeltaSlice(&perUnit, perUnitRows, func(row perUnitPatternRow) (uint32, error) {
		return perUnitPatternKeyValue(localeIndex, unitIDs, row)
	}, func(row perUnitPatternRow) {
		perUnit.appendStringRef(table.Add(row.pattern))
	}); err != nil {
		return nil, err
	}

	var names blobEncoder
	names.appendStringRefSlice(unitNames, table)

	var supported blobEncoder
	supported.appendStringRefSlice(locales, table)

	return renderPayloadFile("unit", table,
		payloadBlob{"_unitPatternBlob", patterns.bytes()},
		payloadBlob{"_compoundUnitBlob", compound.bytes()},
		payloadBlob{"_perUnitPatternBlob", perUnit.bytes()},
		payloadBlob{"_unitNameBlob", names.bytes()},
		payloadBlob{"_unitSupportedBlob", supported.bytes()},
	)
}

// localeIndexMap mirrors the runtime localeIndex: every available locale tag
// maps to its position in the full sorted tag list, with "und" pinned at 0. The
// unit key packing must use these global indices so a key computed at runtime
// from a cldr.Locale matches the encoded key.
func localeIndexMap(locales extract.Locales) map[string]uint64 {
	idx := make(map[string]uint64, len(locales.Tags))
	for i, tag := range locales.Tags {
		idx[tag] = uint64(i)
	}
	return idx
}

// unitPatternKeyValue packs a unit pattern key identically to the runtime
// makeUnitPatternKey.
func unitPatternKeyValue(localeIndex map[string]uint64, unitIDs map[string]int, row unitPatternRow) (uint32, error) {
	loc, err := localeIndexValue(localeIndex, row.locale)
	if err != nil {
		return 0, err
	}
	unitID, ok := unitIDs[row.unit]
	if !ok {
		return 0, fmt.Errorf("unit %q missing from unit ID table", row.unit)
	}
	width, ok := unitWidthOrdinal(row.width)
	if !ok {
		return 0, fmt.Errorf("unknown unit width %q", row.width)
	}
	plural, ok := unitPluralOrdinal(row.plural)
	if !ok {
		return 0, fmt.Errorf("unknown unit plural %q", row.plural)
	}
	return uint32(loc<<unitPatternLocaleShift | uint64(unitID)<<unitPatternUnitShift | width<<unitPatternWidthShift | plural), nil
}

// compoundUnitKeyValue packs a compound key identically to the runtime
// makeCompoundUnitPatternKey.
func compoundUnitKeyValue(localeIndex map[string]uint64, row compoundUnitPatternRow) (uint32, error) {
	loc, err := localeIndexValue(localeIndex, row.locale)
	if err != nil {
		return 0, err
	}
	width, ok := unitWidthOrdinal(row.width)
	if !ok {
		return 0, fmt.Errorf("unknown unit width %q", row.width)
	}
	return uint32(loc<<unitPatternWidthShift | width), nil
}

// perUnitPatternKeyValue packs a specialized per-unit key identically to the
// runtime makePerUnitPatternKey.
func perUnitPatternKeyValue(localeIndex map[string]uint64, unitIDs map[string]int, row perUnitPatternRow) (uint32, error) {
	loc, err := localeIndexValue(localeIndex, row.locale)
	if err != nil {
		return 0, err
	}
	unitID, ok := unitIDs[row.unit]
	if !ok {
		return 0, fmt.Errorf("unit %q missing from unit ID table", row.unit)
	}
	width, ok := unitWidthOrdinal(row.width)
	if !ok {
		return 0, fmt.Errorf("unknown unit width %q", row.width)
	}
	return uint32(loc<<unitPatternLocaleShift | uint64(unitID)<<unitPatternUnitShift | width<<unitPatternWidthShift), nil
}

func unitWidthOrdinal(width string) (uint64, bool) {
	return ordinalIn(unitWidthOrder[:], width)
}

func unitPluralOrdinal(plural string) (uint64, bool) {
	return ordinalIn(unitPluralOrder[:], plural)
}

func ordinalIn(values []string, value string) (uint64, bool) {
	idx := slices.Index(values, value)
	if idx < 0 {
		return 0, false
	}
	return uint64(idx + 1), true
}

type unitPatternRow struct {
	locale  string
	unit    string
	width   string
	plural  string
	pattern string
}

type compoundUnitPatternRow struct {
	locale  string
	width   string
	pattern string
}

type perUnitPatternRow struct {
	locale  string
	unit    string
	width   string
	pattern string
}

func sortedUnitPatternNames(data extract.Units) []string {
	seen := map[string]bool{}
	for _, units := range data {
		for unit := range units {
			seen[unit] = true
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

func unitPatternIDs(names []string) map[string]int {
	ids := map[string]int{}
	for i, unit := range names {
		ids[unit] = i + 1
	}
	return ids
}

func unitPatternRows(data extract.Units) []unitPatternRow {
	var rows []unitPatternRow
	for _, locale := range sortedLocaleKeys(data) {
		for _, unit := range slices.Sorted(maps.Keys(data[locale])) {
			unitData := data[locale][unit]
			for _, width := range unitWidthOrder[:] {
				widthPatterns := unitData.Patterns[width]
				if len(widthPatterns) == 0 {
					continue
				}
				patterns := widthPatterns[unit]
				for _, plural := range unitPluralOrder[:] {
					if pattern := patterns[plural]; pattern != "" {
						rows = append(rows, unitPatternRow{locale: locale, unit: unit, width: width, plural: plural, pattern: pattern})
					}
				}
			}
		}
	}
	return rows
}

func compoundUnitPatternRows(data extract.Units) []compoundUnitPatternRow {
	var rows []compoundUnitPatternRow
	for _, locale := range sortedLocaleKeys(data) {
		compounds := compoundPatternsForLocale(data[locale])
		for _, width := range unitWidthOrder[:] {
			if pattern := compounds[width]; pattern != "" {
				rows = append(rows, compoundUnitPatternRow{locale: locale, width: width, pattern: pattern})
			}
		}
	}
	return rows
}

func perUnitPatternRows(data extract.Units) []perUnitPatternRow {
	var rows []perUnitPatternRow
	for _, locale := range sortedLocaleKeys(data) {
		for _, unit := range slices.Sorted(maps.Keys(data[locale])) {
			for _, width := range unitWidthOrder[:] {
				if pattern := data[locale][unit].PerUnit[width]; pattern != "" {
					rows = append(rows, perUnitPatternRow{locale: locale, unit: unit, width: width, pattern: pattern})
				}
			}
		}
	}
	return rows
}

func compoundPatternsForLocale(units cldr.Units) map[string]string {
	out := map[string]string{}
	for _, unit := range slices.Sorted(maps.Keys(units)) {
		for _, width := range unitWidthOrder[:] {
			if out[width] == "" {
				out[width] = units[unit].Compound[width]
			}
		}
	}
	return out
}

var unitWidthOrder = [...]string{"long", "narrow", "short"}

var unitPluralOrder = [...]string{"few", "many", "one", "other", "two", "zero"}
