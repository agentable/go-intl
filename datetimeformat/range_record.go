package datetimeformat

import (
	"strings"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	"github.com/agentable/go-intl/internal/ecma402"
	ecma402dtf "github.com/agentable/go-intl/internal/ecma402/datetimeformat"
	"github.com/agentable/go-intl/internal/pattern"
)

type dateRangeField uint8

const (
	dateRangeEra dateRangeField = iota
	dateRangeYear
	dateRangeMonth
	dateRangeDay
	dateRangeFieldCount
)

type rangePatternRecord struct {
	dateFields    [dateRangeFieldCount]intervalProgram
	dateFallbacks [dateRangeFieldCount]patternProgram
	timeFields    [timeRangeFieldCount]intervalProgram
	periodKind    rangePeriodKind
	timeStart     timeRangeField
	hasTime       bool
	fallback      rangeFallbackProgram
}

type rangeRelation struct {
	equal           bool
	program         intervalProgram
	fallbackProgram patternProgram
}

type timeRangeField uint8

const (
	timeRangePeriod timeRangeField = iota
	timeRangeHour
	timeRangeMinute
	timeRangeSecond
	timeRangeFractionalSecond
	timeRangeFieldCount
)

type rangePeriodKind uint8

const (
	rangePeriodNone rangePeriodKind = iota
	rangePeriodStandard
	rangePeriodFlexible
)

func newRangePatternRecord(selected selectedPattern, patterns *patternData, formatMatcher FormatMatcher, gregorian cldrdate.Gregorian) rangePatternRecord {
	record := rangePatternRecord{fallback: compileRangeFallbackProgram(gregorian.IntervalFallback)}
	record.addDatePatterns(selected.dateFormat, gregorian.IntervalFormats)
	record.addTimePatterns(selected.timeFormat, gregorian.IntervalFormats)
	record.addDateFallbackPatterns(selected, patterns, formatMatcher, gregorian)
	return record
}

func (r *rangePatternRecord) addDateFallbackPatterns(selected selectedPattern, patterns *patternData, formatMatcher FormatMatcher, gregorian cldrdate.Gregorian) {
	if selected.kind == patternTime {
		dateOptions := ecma402dtf.Options{
			Year:  ecma402dtf.NumericNumeric,
			Month: ecma402dtf.FieldNumeric,
			Day:   ecma402dtf.NumericNumeric,
		}
		dateFormat, ok := matchComponentPattern(formatMatcher, dateOptions, patterns.dateCandidates, gregorian.AppendItems)
		if !ok {
			return
		}
		dateTimePattern := dateTimeStylePattern(gregorian, ShortDateTimeStyle)
		fallback := combineRangeEndpointPattern(dateTimePattern, dateFormat.Pattern, selected.time)
		for field := dateRangeYear; field < dateRangeFieldCount; field++ {
			r.dateFallbacks[field] = compilePatternProgram(fallback)
		}
		dateOptions.Era = ecma402dtf.FieldShort
		if eraFormat, ok := matchComponentPattern(formatMatcher, dateOptions, patterns.dateCandidates, gregorian.AppendItems); ok {
			r.dateFallbacks[dateRangeEra] = compilePatternProgram(combineRangeEndpointPattern(dateTimePattern, eraFormat.Pattern, selected.time))
		}
		return
	}
	if selected.kind != patternDateTime {
		return
	}
	dateOptions := formatOptions(selected.dateFormat)
	// Larger differences retain the missing smaller fields added for their fallbacks.
	for field := dateRangeDay; ; field-- {
		if addMissingDateRangeField(&dateOptions, field) {
			if dateFormat, ok := matchComponentPattern(formatMatcher, dateOptions, patterns.dateCandidates, gregorian.AppendItems); ok {
				r.dateFallbacks[field] = compilePatternProgram(combineRangeEndpointPattern(selected.dateTime, dateFormat.Pattern, selected.time))
			}
		}
		if field == dateRangeEra {
			break
		}
	}
}

func addMissingDateRangeField(options *ecma402dtf.Options, field dateRangeField) bool {
	switch field {
	case dateRangeEra:
		if options.Era != "" {
			return false
		}
		options.Era = ecma402dtf.FieldShort
	case dateRangeYear:
		if options.Year != "" {
			return false
		}
		options.Year = ecma402dtf.NumericNumeric
	case dateRangeMonth:
		if options.Month != "" {
			return false
		}
		options.Month = ecma402dtf.FieldNumeric
	case dateRangeDay:
		if options.Day != "" {
			return false
		}
		options.Day = ecma402dtf.NumericNumeric
	case dateRangeFieldCount:
		return false
	}
	return true
}

func combineRangeEndpointPattern(dateTimePattern, datePattern, timePattern string) string {
	return pattern.FormatIndexed(dateTimePattern, timePattern, datePattern)
}

func (r *rangePatternRecord) addDatePatterns(format ecma402dtf.Formats, intervals map[string]map[string]string) {
	patterns := intervals[format.Skeleton]
	for field, key := range [...]string{"G", "y", "M", "d"} {
		r.dateFields[field] = adjustedRangePattern(patterns[key], format)
	}
}

func (r *rangePatternRecord) addTimePatterns(format ecma402dtf.Formats, intervals map[string]map[string]string) {
	patterns := intervals[format.Skeleton]
	if len(patterns) == 0 {
		patterns = intervals[strings.TrimRight(format.Skeleton, "S")]
	}
	periodPattern := patterns["B"]
	if periodPattern != "" {
		r.periodKind = rangePeriodFlexible
	} else {
		periodPattern = patterns["a"]
		if periodPattern != "" {
			r.periodKind = rangePeriodStandard
		}
	}
	r.timeFields[timeRangePeriod] = adjustedRangePattern(periodPattern, format)
	hourPattern := patterns["h"]
	if hourPattern == "" {
		hourPattern = patterns["H"]
	}
	r.timeFields[timeRangeHour] = adjustedRangePattern(hourPattern, format)
	r.timeFields[timeRangeMinute] = adjustedRangePattern(patterns["m"], format)
	r.timeFields[timeRangeSecond] = adjustedRangePattern(patterns["s"], format)
	r.timeFields[timeRangeFractionalSecond] = adjustedRangePattern(patterns["S"], format)
	for field, program := range r.timeFields {
		if program != nil {
			r.timeStart = timeRangeField(field)
			r.hasTime = true
			break
		}
	}
}

func adjustedRangePattern(pattern string, format ecma402dtf.Formats) intervalProgram {
	if pattern == "" {
		return nil
	}
	parsed := ecma402dtf.Parse(pattern, pattern, nil, "")
	return compileIntervalPattern(ecma402dtf.AdjustFieldTypes(parsed, formatOptions(format)).Pattern)
}

func (r rangePatternRecord) dateRelation(start, end localTime) rangeRelation {
	var selected intervalProgram
	for field, program := range r.dateFields {
		if selected != nil && program == nil {
			break
		}
		if program != nil {
			selected = program
		}
		if !dateRangeFieldEqual(dateRangeField(field), start, end) {
			return rangeRelation{program: selected}
		}
	}
	return rangeRelation{equal: true}
}

func dateRangeFieldEqual(field dateRangeField, start, end localTime) bool {
	switch field {
	case dateRangeEra:
		return start.Era == end.Era
	case dateRangeYear:
		return start.Year == end.Year
	case dateRangeMonth:
		return start.Month == end.Month
	case dateRangeDay:
		return start.Day == end.Day
	case dateRangeFieldCount:
	}
	return true
}

func (r rangePatternRecord) timeRelation(format *DateTimeFormat, start, end localTime) rangeRelation {
	for field := dateRangeEra; field < dateRangeFieldCount; field++ {
		if !dateRangeFieldEqual(field, start, end) {
			return rangeRelation{fallbackProgram: r.dateFallbacks[field]}
		}
	}
	if !r.hasTime {
		for field := timeRangeHour; field < timeRangeFieldCount; field++ {
			if !timeRangeFieldEqual(field, rangePeriodNone, format, start, end) {
				return rangeRelation{}
			}
		}
		return rangeRelation{equal: true}
	}
	var selected intervalProgram
	for field := r.timeStart; field < timeRangeFieldCount; field++ {
		program := r.timeFields[field]
		if selected != nil && program == nil {
			break
		}
		if program != nil {
			selected = program
		}
		if !timeRangeFieldEqual(field, r.periodKind, format, start, end) {
			return rangeRelation{program: selected}
		}
	}
	return rangeRelation{equal: true}
}

func timeRangeFieldEqual(field timeRangeField, periodKind rangePeriodKind, format *DateTimeFormat, start, end localTime) bool {
	switch field {
	case timeRangePeriod:
		switch periodKind {
		case rangePeriodStandard:
			return (start.Hour < 12) == (end.Hour < 12)
		case rangePeriodFlexible:
			return flexibleDayPeriodValue(format.cldrLoc, start) == flexibleDayPeriodValue(format.cldrLoc, end)
		case rangePeriodNone:
			return true
		}
	case timeRangeHour:
		return start.Hour == end.Hour
	case timeRangeMinute:
		return start.Minute == end.Minute
	case timeRangeSecond:
		return start.Second == end.Second
	case timeRangeFractionalSecond:
		digits := ecma402.ResolvedScalarValue(format.resolved.FractionalSecondDigits)
		return digits == 0 || fractionalSecondValue(start.Nanosecond, digits) == fractionalSecondValue(end.Nanosecond, digits)
	case timeRangeFieldCount:
	}
	return true
}
