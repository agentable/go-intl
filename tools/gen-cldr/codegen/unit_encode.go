package codegen

import (
	"bytes"
	"fmt"
	"maps"
	"slices"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// encodeUnits renders the const-only payload for the unit domain. It emits a
// private _data string table plus three independent blobs, each prefixed by its
// record count:
//
//   - _unitPatternBlob:   sorted (loc<<16|unit<<8|width<<4|plural) delta keys
//     paired with pattern StringRefs.
//   - _compoundUnitBlob:  sorted (loc<<4|width) delta keys paired with pattern
//     StringRefs.
//   - _unitNameBlob:      the sorted unit-name table (id = index+1), so the
//     decoder can rebuild the name->id map the pattern key packing needs.
//   - _unitSupportedBlob: the unit-supported locale tags in sorted-locale order,
//     a narrow index so SupportedLocales never decodes the pattern blobs.
//
// The keys are emitted in the exact ascending order the pattern rows already
// arrive in, so the delta stream is monotonic by construction and the runtime
// binary search sees the same order as the legacy literal table.
func encodeUnits(input RuntimeInput, table *StringTable) ([]byte, error) {
	data := input.Units
	localeIndex := localeIndexMap(input.Locales)
	unitIDs := unitPatternIDs(data)

	// Field-width guards. The unit pattern key packs four fields into a uint32:
	// loc<<16 | unit<<8 | width<<4 | plural. unit occupies the 8-bit slot, so a
	// name id above 255 would collide into the width field; loc occupies the
	// upper 16 bits, so an index above 65535 would overflow the uint32 key (and
	// the runtime Locale is a uint16, so it cannot represent such an index). The
	// width and plural ordinals are bounded to 1..3 and 1..6 by their respective
	// ordinal switches, which panic on any unexpected value.
	if n := len(sortedUnitPatternNames(data)); n > 255 {
		panic(fmt.Sprintf("unit name id table has %d entries; the 8-bit unit field in the pattern key holds at most 255", n))
	}
	if n := len(input.Locales.Tags); n > 65535 {
		panic(fmt.Sprintf("locale table has %d tags; the 16-bit locale field in the unit pattern key holds at most 65535", n))
	}

	var patterns blobEncoder
	patternRows := unitPatternRows(data)
	patterns.appendUvarint(uint64(len(patternRows)))
	for _, row := range patternRows {
		key := unitPatternKeyValue(mustLocaleIndex(localeIndex, row.locale), unitIDs[row.unit], row.width, row.plural)
		patterns.appendDelta(key)
		patterns.appendStringRef(table.Add(row.pattern))
	}

	var compound blobEncoder
	compoundRows := compoundUnitPatternRows(data)
	compound.appendUvarint(uint64(len(compoundRows)))
	for _, row := range compoundRows {
		key := compoundUnitKeyValue(mustLocaleIndex(localeIndex, row.locale), row.width)
		compound.appendDelta(key)
		compound.appendStringRef(table.Add(row.pattern))
	}

	var names blobEncoder
	sortedNames := sortedUnitPatternNames(data)
	names.appendUvarint(uint64(len(sortedNames)))
	for _, name := range sortedNames {
		names.appendStringRef(table.Add(name))
	}

	var supported blobEncoder
	supportedTags := unitSupportedLocaleTags(data)
	supported.appendUvarint(uint64(len(supportedTags)))
	for _, tag := range supportedTags {
		supported.appendStringRef(table.Add(tag))
	}

	var b bytes.Buffer
	b.WriteString("package unit\n\n")
	if err := table.EmitDataConst(&b, "_data"); err != nil {
		return nil, err
	}
	b.WriteString("\n")
	if err := emitStringConst(&b, "_unitPatternBlob", string(patterns.bytes())); err != nil {
		return nil, err
	}
	b.WriteString("\n")
	if err := emitStringConst(&b, "_compoundUnitBlob", string(compound.bytes())); err != nil {
		return nil, err
	}
	b.WriteString("\n")
	if err := emitStringConst(&b, "_unitNameBlob", string(names.bytes())); err != nil {
		return nil, err
	}
	b.WriteString("\n")
	if err := emitStringConst(&b, "_unitSupportedBlob", string(supported.bytes())); err != nil {
		return nil, err
	}
	return FormatFile(b.Bytes())
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

// mustLocaleIndex resolves a data locale against the kernel tag set. A locale
// present in domain data but absent from the kernel registry would silently
// collide onto index 0 ("und") and overwrite its rows; fail loudly at
// generation time instead.
func mustLocaleIndex(idx map[string]uint64, locale string) uint64 {
	i, ok := idx[locale]
	if !ok {
		panic("locale " + locale + " missing from kernel locale registry")
	}
	return i
}

// unitPatternKeyValue packs a unit pattern key identically to the runtime
// makeUnitPatternKey: loc<<16 | unit<<8 | width<<4 | plural.
func unitPatternKeyValue(loc uint64, unitID int, width, plural string) uint64 {
	return loc<<16 | uint64(unitID)<<8 | unitWidthOrdinal(width)<<4 | unitPluralOrdinal(plural)
}

// compoundUnitKeyValue packs a compound key identically to the runtime
// makeCompoundUnitPatternKey: loc<<4 | width.
func compoundUnitKeyValue(loc uint64, width string) uint64 {
	return loc<<4 | unitWidthOrdinal(width)
}

func unitWidthOrdinal(width string) uint64 {
	switch width {
	case "long":
		return 1
	case "narrow":
		return 2
	case "short":
		return 3
	default:
		panic("unknown unit width " + width)
	}
}

func unitPluralOrdinal(plural string) uint64 {
	switch plural {
	case "few":
		return 1
	case "many":
		return 2
	case "one":
		return 3
	case "other":
		return 4
	case "two":
		return 5
	case "zero":
		return 6
	default:
		panic("unknown unit plural " + plural)
	}
}

// unitSupportedLocaleTags returns the unit-supported locale tags in
// sorted-locale order, matching the legacy unitSupportedLocaleTags runtime
// accessor (which walked unitLocales in the same order).
func unitSupportedLocaleTags(data extract.Units) []string {
	return sortedLocaleKeys(data)
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

func sortedUnitPatternNames(data extract.Units) []string {
	seen := map[string]bool{}
	for _, units := range data {
		for unit := range units {
			seen[unit] = true
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

func unitPatternIDs(data extract.Units) map[string]int {
	ids := map[string]int{}
	for i, unit := range sortedUnitPatternNames(data) {
		ids[unit] = i + 1
	}
	return ids
}

func unitPatternRows(data extract.Units) []unitPatternRow {
	var rows []unitPatternRow
	for _, locale := range sortedLocaleKeys(data) {
		for _, unit := range slices.Sorted(maps.Keys(data[locale])) {
			unitData := data[locale][unit]
			for _, width := range unitWidthOrder {
				widthPatterns := unitData.Patterns[width]
				if len(widthPatterns) == 0 {
					continue
				}
				patterns := widthPatterns[unit]
				for _, plural := range unitPluralOrder {
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
		for _, width := range unitWidthOrder {
			if pattern := compounds[width]; pattern != "" {
				rows = append(rows, compoundUnitPatternRow{locale: locale, width: width, pattern: pattern})
			}
		}
	}
	return rows
}

func compoundPatternsForLocale(units cldr.Units) map[string]string {
	out := map[string]string{}
	for _, unit := range slices.Sorted(maps.Keys(units)) {
		for _, width := range unitWidthOrder {
			if out[width] == "" {
				out[width] = units[unit].Compound[width]
			}
		}
	}
	return out
}

var unitWidthOrder = []string{"long", "narrow", "short"}

var unitPluralOrder = []string{"few", "many", "one", "other", "two", "zero"}
