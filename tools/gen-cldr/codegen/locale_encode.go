package codegen

import (
	"maps"
	"slices"
	"strings"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
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
//   - _directionBlob:     known script directions, sorted by script, each a
//     (script, rtl) pair. UNKNOWN and missing source values are absent.
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
	locales.appendStringRefSlice(input.Locales.Tags, table)

	var maximize blobEncoder
	maxKeys := slices.Sorted(maps.Keys(input.LikelySubtags.Maximize))
	appendCountedSlice(&maximize, maxKeys, func(key string) {
		triple := input.LikelySubtags.Maximize[key]
		maximize.appendStringRef(table.Add(key))
		appendSubtagTriple(&maximize, table, triple)
	})

	var direction blobEncoder
	directionScripts := slices.Sorted(maps.Keys(input.ScriptDirections))
	appendCountedSlice(&direction, directionScripts, func(script string) {
		direction.appendStringRef(table.Add(script))
		if input.ScriptDirections[script] {
			direction.appendUvarint(1)
		} else {
			direction.appendUvarint(0)
		}
	})

	var minimize blobEncoder
	minTriples := slices.SortedFunc(maps.Keys(input.LikelySubtags.Minimize), func(a, b extract.SubtagTriple) int {
		return strings.Compare(subtagTripleKey(a), subtagTripleKey(b))
	})
	appendCountedSlice(&minimize, minTriples, func(triple extract.SubtagTriple) {
		minimized := input.LikelySubtags.Minimize[triple]
		appendSubtagTriple(&minimize, table, triple)
		minimize.appendStringRef(table.Add(minimized))
	})

	var numbering blobEncoder
	numberingLocales := localeNumberingOverrideLocales(input.Numbers)
	if err := numbering.appendLocaleDeltaRecords(numberingLocales, localeIndex, func(locale string) {
		numbering.appendStringRef(table.Add(input.Numbers[locale].DefaultNumberingSystem))
	}); err != nil {
		return nil, err
	}

	var hourCycle blobEncoder
	appendStringRefKeyMap(&hourCycle, input.Preferences.HourCycle, table, func(values []string) {
		hourCycle.appendStringRefSlice(values, table)
	})

	var week blobEncoder
	appendStringRefKeyMap(&week, input.Preferences.Week, table, func(w cldr.WeekData) {
		week.appendUvarint(weekdayOrdinal(w.FirstDay))
		week.appendUvarint(weekdayOrdinal(w.WeekendStart))
		week.appendUvarint(weekdayOrdinal(w.WeekendEnd))
		week.appendUvarint(uint64(w.MinDays))
	})

	var calendar blobEncoder
	appendStringRefKeyMap(&calendar, input.Preferences.Calendar, table, func(values []string) {
		calendar.appendStringRefSlice(values, table)
	})

	return renderPayloadFile("cldrlocale", table,
		payloadBlob{"_localeBlob", locales.bytes()},
		payloadBlob{"_maximizeBlob", maximize.bytes()},
		payloadBlob{"_minimizeBlob", minimize.bytes()},
		payloadBlob{"_directionBlob", direction.bytes()},
		payloadBlob{"_numberingBlob", numbering.bytes()},
		payloadBlob{"_hourCycleBlob", hourCycle.bytes()},
		payloadBlob{"_weekBlob", week.bytes()},
		payloadBlob{"_calendarBlob", calendar.bytes()},
	)
}

func appendSubtagTriple(e *blobEncoder, table *StringTable, triple extract.SubtagTriple) {
	e.appendStringRef(table.Add(triple.Lang))
	e.appendStringRef(table.Add(triple.Script))
	e.appendStringRef(table.Add(triple.Region))
}

func subtagTripleKey(triple extract.SubtagTriple) string {
	return triple.Lang + "-" + triple.Script + "-" + triple.Region
}

// localeNumberingOverrideLocales returns the locales whose default numbering
// system is not "latn". Locale-index lookup and ordering belong to
// appendLocaleDeltaRecords.
func localeNumberingOverrideLocales(numbers extract.Numbers) []string {
	var locales []string
	for _, locale := range sortedLocaleKeys(numbers) {
		ns := numbers[locale].DefaultNumberingSystem
		if ns == "" || ns == "latn" {
			continue
		}
		locales = append(locales, locale)
	}
	return locales
}

// weekdayOrdinal maps a CLDR weekday abbreviation to the time.Weekday ordinal
// (Sunday = 0 .. Saturday = 6). An unknown value defaults to Sunday, matching
// the locale preference accessor fallback.
func weekdayOrdinal(day string) uint64 {
	idx := slices.Index(weekdayOrder[:], day)
	if idx < 0 {
		return 0
	}
	return uint64(idx)
}

var weekdayOrder = [...]string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}
