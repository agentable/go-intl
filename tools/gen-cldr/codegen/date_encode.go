package codegen

import (
	"maps"
	"slices"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

const (
	dateCLDRGregorianCalendar      = "gregorian"
	dateSupportedGregorianCalendar = "gregory"
	dateSupportedISO8601Calendar   = "iso8601"
)

// encodeDates renders the const-only payload for the date domain. It emits a
// private _data string table plus four independent blobs, each prefixed by its
// record count:
//
//   - _dateGregorianBlob: per-locale gregorian calendar data (names by
//     width/context, style formats, available formats, interval formats, append
//     items). Locale indices are written as a sorted delta stream; the decoder
//     rebuilds the runtime map[Locale]calendarData.
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
	if err := gregorian.appendLocaleDeltaRecords(gregLocales, localeIndex, func(locale string) {
		encodeCalendar(&gregorian, dates[locale].Calendars[dateCLDRGregorianCalendar], table)
	}); err != nil {
		return nil, err
	}

	var dayPeriods blobEncoder
	ruleSet := dateDayPeriodRuleSet(dates)
	// CLDR day-period rules cover languages beyond the kernel locale registry
	// (kok, yue, zu, ...). Such rules are unreachable at runtime — the locale
	// cannot resolve to a kernel index — and, unfiltered, they all collide onto
	// index 0 and overwrite the genuine "und" rules. Encode only registered
	// locales.
	dpLocales := slices.Sorted(maps.Keys(ruleSet))
	dpLocales = slices.DeleteFunc(dpLocales, func(locale string) bool {
		_, ok := localeIndex[locale]
		return !ok
	})
	if err := dayPeriods.appendLocaleDeltaRecords(dpLocales, localeIndex, func(locale string) {
		rules := ruleSet[locale]
		appendCountedSlice(&dayPeriods, rules, func(rule cldr.DayPeriodRange) {
			appendDayPeriodRange(&dayPeriods, rule, table)
		})
	}); err != nil {
		return nil, err
	}

	var supported blobEncoder
	supported.appendStringRefSlice(gregLocales, table)

	var calendars blobEncoder
	calendarIDs := dateSupportedCalendars(dates)
	calendars.appendStringRefSlice(calendarIDs, table)

	return renderPayloadFile("date", table,
		payloadBlob{"_dateGregorianBlob", gregorian.bytes()},
		payloadBlob{"_dateDayPeriodBlob", dayPeriods.bytes()},
		payloadBlob{"_dateSupportedBlob", supported.bytes()},
		payloadBlob{"_dateCalendarBlob", calendars.bytes()},
	)
}

// encodeCalendar serializes one gregorian Calendar. The decoder reads the fields
// back in this exact order.
func encodeCalendar(e *blobEncoder, cal cldr.Calendar, table *StringTable) {
	// Quarters are intentionally omitted: GregorianFor and flexibleDayPeriodNames
	// read only eras, months, weekdays, and day periods, so quarters are dead
	// weight in the consumed surface.
	for _, key := range calendarNameKeyOrder[:] {
		encodeCalendarNames(e, cal.Names[key], table)
	}
	e.appendStringRefMap(cal.DateFormats, table)
	e.appendStringRefMap(cal.TimeFormats, table)
	e.appendStringRefMap(cal.DateTimeFormats, table)
	e.appendStringRefMap(cal.DateTimeAtFormats, table)
	e.appendStringRefMap(cal.AvailableFormats, table)
	e.appendStringRef(table.Add(cal.IntervalFormats.FallbackPattern))
	encodeIntervalSkeletons(e, cal.IntervalFormats.BySkeleton, table)
	e.appendStringRefMap(cal.AppendItems, table)
}

func encodeCalendarNames(e *blobEncoder, names cldr.CalendarNames, table *StringTable) {
	e.appendStringRefSlice(names.Eras, table)
	e.appendStringRefSlice(names.Months, table)
	e.appendStringRefSlice(names.Weekdays, table)
	e.appendStringRefSlice(names.DayPeriods, table)
}

func appendDayPeriodRange(e *blobEncoder, rule cldr.DayPeriodRange, table *StringTable) {
	e.appendUvarint(uint64(rule.From))
	e.appendUvarint(uint64(rule.To))
	e.appendStringRef(table.Add(rule.Type))
}

// calendarNameKeyOrder is the fixed serialization order of the six CalendarNames
// entries. The decoder rebuilds the calendarNameKey map from this same order.
var calendarNameKeyOrder = [...]cldr.CalendarNameKey{
	{Width: "wide", Context: "format"},
	{Width: "abbreviated", Context: "format"},
	{Width: "narrow", Context: "format"},
	{Width: "wide", Context: "stand-alone"},
	{Width: "abbreviated", Context: "stand-alone"},
	{Width: "narrow", Context: "stand-alone"},
}

func encodeIntervalSkeletons(e *blobEncoder, values map[string]map[string]string, table *StringTable) {
	appendStringRefKeyMap(e, values, table, func(intervals map[string]string) {
		e.appendStringRefMap(intervals, table)
	})
}

// dateGregorianLocales returns the locales with gregorian calendar data in
// sorted payload order.
func dateGregorianLocales(dates extract.Dates) []string {
	out := make([]string, 0, len(dates))
	for _, locale := range sortedLocaleKeys(dates) {
		if _, ok := dates[locale].Calendars[dateCLDRGregorianCalendar]; ok {
			out = append(out, locale)
		}
	}
	return out
}

// dateDayPeriodRuleSet flattens the day-period rule set carried by every
// locale's Dates record into one map keyed by ruleset-locale. Each locale's
// DayPeriodRules holds the same full ruleset map (keyed by the supplemental
// dayPeriods.json locales), so iterating any one locale's map yields the union.
func dateDayPeriodRuleSet(dates extract.Dates) map[string][]cldr.DayPeriodRange {
	out := make(map[string][]cldr.DayPeriodRange)
	for _, locale := range sortedLocaleKeys(dates) {
		for key, rules := range dates[locale].DayPeriodRules {
			out[key] = rules
		}
	}
	return out
}

// dateSupportedCalendars returns the ECMA-402 calendar identifiers supported by
// the date payload: CLDR Gregorian maps to ECMA-402 Gregorian, and its presence
// adds the ECMA-402 ISO 8601 bridge.
func dateSupportedCalendars(dates extract.Dates) []string {
	seen := map[string]bool{}
	for _, data := range dates {
		for calendar := range data.Calendars {
			switch calendar {
			case dateCLDRGregorianCalendar:
				seen[dateSupportedGregorianCalendar] = true
			default:
				seen[calendar] = true
			}
		}
	}
	if seen[dateSupportedGregorianCalendar] {
		seen[dateSupportedISO8601Calendar] = true
	}
	return slices.Sorted(maps.Keys(seen))
}
