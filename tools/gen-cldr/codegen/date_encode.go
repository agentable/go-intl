package codegen

import (
	"bytes"
	"maps"
	"slices"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// encodeDates renders the const-only payload for the date domain. It emits a
// private _data string table plus four independent blobs, each prefixed by its
// record count:
//
//   - _dateGregorianBlob: per-locale gregorian calendar data (names by
//     width/context, style formats, available formats, interval formats, append
//     items). Locale indices are written as a sorted delta stream; the decoder
//     rebuilds the same map[Locale]calendarData the legacy loadDateData produced.
//   - _dateDayPeriodBlob: per-locale day-period rules (From/To durations and the
//     period type), the input to DayPeriodFor.
//   - _dateSupportedBlob: the date-supported locale tags in sorted-locale order,
//     a narrow index so SupportedLocales never decodes the gregorian blob.
//   - _dateCalendarBlob: the supported ECMA-402 calendar identifiers, a narrow
//     index so SupportedCalendars never decodes the gregorian blob.
//
// Only the gregorian calendar is present in the CLDR payload, so the blob stores
// it directly without a calendar key.
func encodeDates(input RuntimeInput, table *StringTable) ([]byte, error) {
	dates := input.Dates
	localeIndex := localeIndexMap(input.Locales)

	var gregorian blobEncoder
	gregLocales := dateGregorianLocales(dates)
	gregorian.appendUvarint(uint64(len(gregLocales)))
	for _, locale := range gregLocales {
		gregorian.appendDelta(mustLocaleIndex(localeIndex, locale))
		encodeCalendar(&gregorian, dates[locale].Calendars["gregorian"], table)
	}

	var dayPeriods blobEncoder
	ruleSet := dateDayPeriodRuleSet(dates)
	// CLDR day-period rules cover languages beyond the kernel locale registry
	// (kok, yue, zu, ...). Such rules are unreachable at runtime — the locale
	// cannot resolve to a kernel index — and, unfiltered, they all collide onto
	// index 0 and overwrite the genuine "und" rules (the legacy literal renderer
	// shipped Zulu rules as "und" this way). Encode only registered locales.
	dpLocales := slices.Sorted(maps.Keys(ruleSet))
	dpLocales = slices.DeleteFunc(dpLocales, func(locale string) bool {
		_, ok := localeIndex[locale]
		return !ok
	})
	slices.SortFunc(dpLocales, func(a, b string) int {
		return cmpUint64(mustLocaleIndex(localeIndex, a), mustLocaleIndex(localeIndex, b))
	})
	dayPeriods.appendUvarint(uint64(len(dpLocales)))
	for _, locale := range dpLocales {
		dayPeriods.appendDelta(mustLocaleIndex(localeIndex, locale))
		rules := ruleSet[locale]
		dayPeriods.appendUvarint(uint64(len(rules)))
		for _, rule := range rules {
			dayPeriods.appendUvarint(uint64(rule.From))
			dayPeriods.appendUvarint(uint64(rule.To))
			dayPeriods.appendStringRef(table.Add(rule.Type))
		}
	}

	var supported blobEncoder
	supportedTags := dateSupportedLocaleTags(dates)
	supported.appendUvarint(uint64(len(supportedTags)))
	for _, tag := range supportedTags {
		supported.appendStringRef(table.Add(tag))
	}

	var calendars blobEncoder
	calendarIDs := dateSupportedCalendars(dates)
	calendars.appendUvarint(uint64(len(calendarIDs)))
	for _, id := range calendarIDs {
		calendars.appendStringRef(table.Add(id))
	}

	var b bytes.Buffer
	b.WriteString("package date\n\n")
	if err := table.EmitDataConst(&b, "_data"); err != nil {
		return nil, err
	}
	for _, blob := range []struct {
		name  string
		bytes []byte
	}{
		{"_dateGregorianBlob", gregorian.bytes()},
		{"_dateDayPeriodBlob", dayPeriods.bytes()},
		{"_dateSupportedBlob", supported.bytes()},
		{"_dateCalendarBlob", calendars.bytes()},
	} {
		b.WriteString("\n")
		if err := emitStringConst(&b, blob.name, string(blob.bytes)); err != nil {
			return nil, err
		}
	}
	return FormatFile(b.Bytes())
}

// encodeCalendar serializes one gregorian Calendar. The decoder reads the fields
// back in this exact order.
func encodeCalendar(e *blobEncoder, cal cldr.Calendar, table *StringTable) {
	// Quarters are intentionally omitted: GregorianFor and flexibleDayPeriodNames
	// read only eras, months, weekdays, and day periods, so quarters are dead
	// weight in the consumed surface.
	for _, key := range calendarNameKeyOrder {
		names := cal.Names[key]
		encodeStringSlice(e, names.Eras, table)
		encodeStringSlice(e, names.Months, table)
		encodeStringSlice(e, names.Weekdays, table)
		encodeStringSlice(e, names.DayPeriods, table)
	}
	encodeStringMap(e, cal.DateFormats, table)
	encodeStringMap(e, cal.TimeFormats, table)
	encodeStringMap(e, cal.DateTimeFormats, table)
	encodeStringMap(e, cal.DateTimeAtFormats, table)
	encodeStringMap(e, cal.AvailableFormats, table)
	e.appendStringRef(table.Add(cal.IntervalFormats.FallbackPattern))
	encodeIntervalSkeletons(e, cal.IntervalFormats.BySkeleton, table)
	encodeStringMap(e, cal.AppendItems, table)
}

// calendarNameKeyOrder is the fixed serialization order of the six CalendarNames
// entries. The decoder rebuilds the calendarNameKey map from this same order.
var calendarNameKeyOrder = []cldr.CalendarNameKey{
	{Width: "wide", Context: "format"},
	{Width: "abbreviated", Context: "format"},
	{Width: "narrow", Context: "format"},
	{Width: "wide", Context: "stand-alone"},
	{Width: "abbreviated", Context: "stand-alone"},
	{Width: "narrow", Context: "stand-alone"},
}

func encodeStringSlice(e *blobEncoder, values []string, table *StringTable) {
	e.appendUvarint(uint64(len(values)))
	for _, value := range values {
		e.appendStringRef(table.Add(value))
	}
}

func encodeStringMap(e *blobEncoder, values map[string]string, table *StringTable) {
	keys := slices.Sorted(maps.Keys(values))
	e.appendUvarint(uint64(len(keys)))
	for _, key := range keys {
		e.appendStringRef(table.Add(key))
		e.appendStringRef(table.Add(values[key]))
	}
}

func encodeIntervalSkeletons(e *blobEncoder, values map[string]map[string]string, table *StringTable) {
	keys := slices.Sorted(maps.Keys(values))
	e.appendUvarint(uint64(len(keys)))
	for _, key := range keys {
		e.appendStringRef(table.Add(key))
		encodeStringMap(e, values[key], table)
	}
}

// dateGregorianLocales returns the locales with gregorian calendar data in
// sorted order, matching the legacy datesByLocale map key set.
func dateGregorianLocales(dates extract.Dates) []string {
	out := make([]string, 0, len(dates))
	for _, locale := range sortedLocaleKeys(dates) {
		if _, ok := dates[locale].Calendars["gregorian"]; ok {
			out = append(out, locale)
		}
	}
	return out
}

// dateDayPeriodRuleSet flattens the day-period rule set carried by every locale's
// Dates record into one map keyed by ruleset-locale, mirroring the legacy
// anyDayPeriodRules. Each locale's DayPeriodRules holds the same full ruleset
// map (keyed by the supplemental dayPeriods.json locales), so iterating any one
// locale's map yields the union.
func dateDayPeriodRuleSet(dates extract.Dates) map[string][]cldr.DayPeriodRange {
	out := make(map[string][]cldr.DayPeriodRange)
	for _, locale := range sortedLocaleKeys(dates) {
		for key, rules := range dates[locale].DayPeriodRules {
			out[key] = rules
		}
	}
	return out
}

func cmpUint64(a, b uint64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// dateSupportedLocaleTags returns the date-supported locale tags in
// sorted-locale order, matching the legacy DateSupportedLocales accessor (which
// walked the global locale table and kept locales with gregorian data).
func dateSupportedLocaleTags(dates extract.Dates) []string {
	return dateGregorianLocales(dates)
}

// dateSupportedCalendars mirrors the legacy supportedCalendarValues over the date
// payload: gregorian maps to the canonical "gregory", and the presence of
// gregory adds the ECMA-402 iso8601 bridge.
func dateSupportedCalendars(dates extract.Dates) []string {
	seen := map[string]bool{}
	for _, data := range dates {
		for calendar := range data.Calendars {
			switch calendar {
			case "gregorian":
				seen["gregory"] = true
			default:
				seen[calendar] = true
			}
		}
	}
	if seen["gregory"] {
		seen["iso8601"] = true
	}
	return slices.Sorted(maps.Keys(seen))
}
