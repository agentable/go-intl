package codegen

import (
	"bytes"
	"maps"
	"slices"
	"strings"

	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// encodeLocaleKernel renders the const-only payload for the locale kernel
// package (cldrlocale). Unlike the leaf domains, the kernel keeps the package
// name cldrlocale that every formatter borrows the Locale handle from, so the
// payload uses that package clause rather than the directory-derived name.
//
// It emits a private _data string table plus the kernel blobs, each prefixed by
// its record count:
//
//   - _localeBlob:        the available locale tags in sorted order (Locale index
//     = position, "und" pinned at 0). The decoder rebuilds availableLocaleTags
//     and localeIndex from it.
//   - _maximizeBlob:      likely-subtags maximize rows, sorted by key, each a
//     (key, lang, script, region) StringRef quad for binary search.
//   - _minimizeBlob:      likely-subtags minimize rows, sorted by
//     (lang, script, region), each a (lang, script, region, minimized) quad.
//   - _numberingBlob:     the non-latn default-numbering-system overrides as a
//     sorted (localeDelta, system) stream; the decoder defaults every other
//     locale to "latn".
//   - _hourCycleBlob:     per-region hour-cycle preference lists.
//   - _weekBlob:          per-region week data (first/weekendStart/weekendEnd
//     weekday ordinals plus minimal-days-in-first-week).
//   - _calendarBlob:      per-region calendar preference lists.
//
// The kernel _data holds only the kernel's own strings — locale tags, subtag
// triples, regions, numbering systems, and preference values — dissolving the
// formerly shared global string table.
func encodeLocaleKernel(input RuntimeInput, table *StringTable) ([]byte, error) {
	localeIndex := localeIndexMap(input.Locales)

	var locales blobEncoder
	locales.appendUvarint(uint64(len(input.Locales.Tags)))
	for _, tag := range input.Locales.Tags {
		locales.appendStringRef(table.Add(tag))
	}

	var maximize blobEncoder
	maxKeys := slices.Sorted(maps.Keys(input.LikelySubtags.Maximize))
	maximize.appendUvarint(uint64(len(maxKeys)))
	for _, key := range maxKeys {
		triple := input.LikelySubtags.Maximize[key]
		maximize.appendStringRef(table.Add(key))
		maximize.appendStringRef(table.Add(triple.Lang))
		maximize.appendStringRef(table.Add(triple.Script))
		maximize.appendStringRef(table.Add(triple.Region))
	}

	var minimize blobEncoder
	minTriples := slices.SortedFunc(maps.Keys(input.LikelySubtags.Minimize), func(a, b extract.SubtagTriple) int {
		return strings.Compare(a.Lang+"-"+a.Script+"-"+a.Region, b.Lang+"-"+b.Script+"-"+b.Region)
	})
	minimize.appendUvarint(uint64(len(minTriples)))
	for _, triple := range minTriples {
		minimized := input.LikelySubtags.Minimize[triple]
		minimize.appendStringRef(table.Add(triple.Lang))
		minimize.appendStringRef(table.Add(triple.Script))
		minimize.appendStringRef(table.Add(triple.Region))
		minimize.appendStringRef(table.Add(minimized))
	}

	var numbering blobEncoder
	numberingRows := localeNumberingOverrides(input.Numbers, localeIndex)
	numbering.appendUvarint(uint64(len(numberingRows)))
	for _, row := range numberingRows {
		numbering.appendDelta(row.locale)
		numbering.appendStringRef(table.Add(row.system))
	}

	var hourCycle blobEncoder
	hourRegions := slices.Sorted(maps.Keys(input.Preferences.HourCycle))
	hourCycle.appendUvarint(uint64(len(hourRegions)))
	for _, region := range hourRegions {
		hourCycle.appendStringRef(table.Add(region))
		values := input.Preferences.HourCycle[region]
		hourCycle.appendUvarint(uint64(len(values)))
		for _, value := range values {
			hourCycle.appendStringRef(table.Add(value))
		}
	}

	var week blobEncoder
	weekRegions := slices.Sorted(maps.Keys(input.Preferences.Week))
	week.appendUvarint(uint64(len(weekRegions)))
	for _, region := range weekRegions {
		w := input.Preferences.Week[region]
		week.appendStringRef(table.Add(region))
		week.appendUvarint(weekdayOrdinal(w.FirstDay))
		week.appendUvarint(weekdayOrdinal(w.WeekendStart))
		week.appendUvarint(weekdayOrdinal(w.WeekendEnd))
		week.appendUvarint(uint64(w.MinDays))
	}

	var calendar blobEncoder
	calRegions := slices.Sorted(maps.Keys(input.Preferences.Calendar))
	calendar.appendUvarint(uint64(len(calRegions)))
	for _, region := range calRegions {
		calendar.appendStringRef(table.Add(region))
		values := input.Preferences.Calendar[region]
		calendar.appendUvarint(uint64(len(values)))
		for _, value := range values {
			calendar.appendStringRef(table.Add(value))
		}
	}

	var b bytes.Buffer
	b.WriteString("package cldrlocale\n\n")
	if err := table.EmitDataConst(&b, "_data"); err != nil {
		return nil, err
	}
	for _, blob := range []struct {
		name  string
		bytes []byte
	}{
		{"_localeBlob", locales.bytes()},
		{"_maximizeBlob", maximize.bytes()},
		{"_minimizeBlob", minimize.bytes()},
		{"_numberingBlob", numbering.bytes()},
		{"_hourCycleBlob", hourCycle.bytes()},
		{"_weekBlob", week.bytes()},
		{"_calendarBlob", calendar.bytes()},
	} {
		b.WriteString("\n")
		if err := emitStringConst(&b, blob.name, string(blob.bytes)); err != nil {
			return nil, err
		}
	}
	return FormatFile(b.Bytes())
}

type localeNumberingRow struct {
	locale uint64
	system string
}

// localeNumberingOverrides returns the locales whose default numbering system is
// not "latn", in ascending Locale-index order. It mirrors the legacy
// DefaultNumberingSystem switch, which fell through to "latn" for every locale
// without an explicit non-latn override.
func localeNumberingOverrides(numbers extract.Numbers, localeIndex map[string]uint64) []localeNumberingRow {
	var rows []localeNumberingRow
	for _, locale := range sortedLocaleKeys(numbers) {
		ns := numbers[locale].DefaultNumberingSystem
		if ns == "" || ns == "latn" {
			continue
		}
		idx, ok := localeIndex[locale]
		if !ok {
			continue
		}
		rows = append(rows, localeNumberingRow{locale: idx, system: ns})
	}
	slices.SortFunc(rows, func(a, b localeNumberingRow) int { return cmpUint64(a.locale, b.locale) })
	return rows
}

// weekdayOrdinal maps a CLDR weekday abbreviation to the time.Weekday ordinal
// (Sunday = 0 .. Saturday = 6), matching the legacy weekdayLiteral mapping. An
// unknown value defaults to Sunday, as the legacy switch did.
func weekdayOrdinal(day string) uint64 {
	switch day {
	case "sun":
		return 0
	case "mon":
		return 1
	case "tue":
		return 2
	case "wed":
		return 3
	case "thu":
		return 4
	case "fri":
		return 5
	case "sat":
		return 6
	default:
		return 0
	}
}
