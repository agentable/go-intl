package datetimeformat

import (
	"strings"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	"github.com/agentable/go-intl/internal/ecma402"
	ecma402dtf "github.com/agentable/go-intl/internal/ecma402/datetimeformat"
	"github.com/agentable/go-intl/internal/pattern"
)

type patternKind uint8

const (
	patternNone patternKind = iota
	patternDate
	patternTime
	patternDateTime
)

type selectedPattern struct {
	kind                patternKind
	date                string
	time                string
	dateTime            string
	dateSkeleton        string
	timeSkeleton        string
	dateIntervalOptions ecma402dtf.Options
	timeIntervalOptions ecma402dtf.Options
}

func selectPattern(patterns *patternData, formatMatcher FormatMatcher, resolved ResolvedOptions, uses24Hour bool, gregorian cldrdate.Gregorian) selectedPattern {
	appendItems := gregorian.AppendItems
	dateStyle := ecma402.ResolvedScalarValue(resolved.DateStyle)
	timeStyle := ecma402.ResolvedScalarValue(resolved.TimeStyle)
	if dateStyle != "" && timeStyle != "" {
		datePattern := dateStylePattern(gregorian, dateStyle)
		timePattern := timeStylePattern(gregorian, timeStyle)
		dateFormat := styleDateIntervalFormat(patterns, dateStyle, formatMatcher, appendItems)
		timeFormat := styleTimeIntervalFormat(patterns, timeStyle, formatMatcher, appendItems)
		dateStyleOptions := patterns.style(dateStyle)
		timeStyleOptions := patterns.style(timeStyle)
		return selectedPattern{
			kind:                patternDateTime,
			date:                datePattern,
			time:                timePattern,
			dateTime:            dateTimeStylePattern(gregorian, dateStyle),
			dateSkeleton:        dateFormat.Skeleton,
			timeSkeleton:        timeFormat.Skeleton,
			dateIntervalOptions: dateStyleOptions.dateOptions,
			timeIntervalOptions: timeStyleOptions.timeOptions,
		}
	}
	if dateStyle != "" {
		datePattern := dateStylePattern(gregorian, dateStyle)
		dateFormat := styleDateIntervalFormat(patterns, dateStyle, formatMatcher, appendItems)
		return selectedPattern{
			kind:                patternDate,
			date:                datePattern,
			dateSkeleton:        dateFormat.Skeleton,
			dateIntervalOptions: patterns.style(dateStyle).dateOptions,
		}
	}
	if timeStyle != "" {
		timePattern := timeStylePattern(gregorian, timeStyle)
		timeFormat := styleTimeIntervalFormat(patterns, timeStyle, formatMatcher, appendItems)
		return selectedPattern{
			kind:                patternTime,
			time:                timePattern,
			timeSkeleton:        timeFormat.Skeleton,
			timeIntervalOptions: patterns.style(timeStyle).timeOptions,
		}
	}
	if pattern, ok := componentPattern(patterns, formatMatcher, resolved, uses24Hour, gregorian, appendItems); ok {
		return pattern
	}
	return selectedPattern{}
}

func (p selectedPattern) parts(f *DateTimeFormat, t localTime) []Part {
	switch p.kind {
	case patternDate:
		return f.formatPattern(p.date, t)
	case patternTime:
		return f.formatPattern(p.time, t)
	case patternDateTime:
		dateParts := f.formatPattern(p.date, t)
		timeParts := f.formatPattern(p.time, t)
		return interpolateDateTimeParts(p.dateTime, dateParts, timeParts)
	case patternNone:
	}
	return nil
}

func dateStylePattern(g cldrdate.Gregorian, style Style) string {
	if idx, ok := knownDateTimeStyleIndex(style); ok {
		return g.DateFormats[idx]
	}
	return ""
}

func timeStylePattern(g cldrdate.Gregorian, style Style) string {
	if idx, ok := knownDateTimeStyleIndex(style); ok {
		return g.TimeFormats[idx]
	}
	return ""
}

func dateTimeStylePattern(g cldrdate.Gregorian, style Style) string {
	if pattern := dateTimeAtStylePattern(g, style); pattern != "" {
		return pattern
	}
	return dateTimeStandardStylePattern(g, style)
}

func dateTimeAtStylePattern(g cldrdate.Gregorian, style Style) string {
	if idx, ok := knownDateTimeStyleIndex(style); ok {
		return g.DateTimeAtFormats[idx]
	}
	return ""
}

func dateTimeStandardStylePattern(g cldrdate.Gregorian, style Style) string {
	if idx, ok := knownDateTimeStyleIndex(style); ok {
		return g.DateTimeFormats[idx]
	}
	return ""
}

func componentPattern(patterns *patternData, formatMatcher FormatMatcher, resolved ResolvedOptions, uses24Hour bool, gregorian cldrdate.Gregorian, appendItems map[string]string) (selectedPattern, bool) {
	opts := matcherOptions(resolved, uses24Hour)
	hasDate := hasDateMatcherOptions(opts)
	hasTime := hasTimeMatcherOptions(opts)
	switch {
	case hasDate && hasTime:
		dateOpts := dateOnlyMatcherOptions(opts)
		timeOpts := timeOnlyMatcherOptions(opts)
		dateFormat, dateOK := matchComponentPattern(formatMatcher, dateOpts, patterns.dateCandidates, appendItems)
		timeFormat, timeOK := matchComponentPattern(formatMatcher, timeOpts, patterns.timePatternCandidates(timeOpts.FractionalSecondDigits), appendItems)
		if !dateOK || !timeOK {
			return selectedPattern{}, false
		}
		return selectedPattern{
			kind:                patternDateTime,
			date:                dateFormat.Pattern,
			time:                timeFormat.Pattern,
			dateTime:            componentDateTimePattern(gregorian, opts, dateFormat),
			dateSkeleton:        dateFormat.Skeleton,
			timeSkeleton:        timeFormat.Skeleton,
			dateIntervalOptions: dateOpts,
			timeIntervalOptions: timeOpts,
		}, true
	case hasDate:
		format, ok := matchComponentPattern(formatMatcher, opts, patterns.dateCandidates, appendItems)
		if !ok {
			return selectedPattern{}, false
		}
		return selectedPattern{kind: patternDate, date: format.Pattern, dateSkeleton: format.Skeleton, dateIntervalOptions: opts}, true
	case hasTime:
		format, ok := matchComponentPattern(formatMatcher, opts, patterns.timePatternCandidates(opts.FractionalSecondDigits), appendItems)
		if !ok {
			return selectedPattern{}, false
		}
		return selectedPattern{kind: patternTime, time: format.Pattern, timeSkeleton: format.Skeleton, timeIntervalOptions: opts}, true
	default:
		return selectedPattern{}, false
	}
}

func componentDateTimePattern(gregorian cldrdate.Gregorian, opts ecma402dtf.Options, dateFormat ecma402dtf.Formats) string {
	style := MediumDateTimeStyle
	switch {
	case dateFormat.Month == ecma402dtf.FieldLong && dateFormat.Weekday == ecma402dtf.FieldLong:
		style = FullDateTimeStyle
	case dateFormat.Month == ecma402dtf.FieldLong:
		style = LongDateTimeStyle
	case opts.Year == ecma402dtf.Numeric2Digit && (opts.Month == ecma402dtf.FieldNumeric || opts.Month == ecma402dtf.Field2Digit):
		style = ShortDateTimeStyle
	}
	return dateTimeStylePattern(gregorian, style)
}

func styleDateIntervalFormat(patterns *patternData, style Style, formatMatcher FormatMatcher, appendItems map[string]string) ecma402dtf.Formats {
	opts := patterns.style(style).dateOptions
	format, _ := matchComponentPattern(formatMatcher, opts, patterns.dateCandidates, appendItems)
	return format
}

func styleTimeIntervalFormat(patterns *patternData, style Style, formatMatcher FormatMatcher, appendItems map[string]string) ecma402dtf.Formats {
	opts := patterns.style(style).timeOptions
	format, _ := matchComponentPattern(formatMatcher, opts, patterns.timePatternCandidates(opts.FractionalSecondDigits), appendItems)
	return format
}

func timeStylePatternOptions(pattern string) ecma402dtf.Options {
	opts := stylePatternOptions(pattern)
	opts.DayPeriod = ""
	if implied := hourCycleImpliesHour12(HourCycle(opts.HourCycle)); implied != nil {
		opts.Hour12 = implied
	}
	return opts
}

func stylePatternOptions(pattern string) ecma402dtf.Options {
	format := ecma402dtf.Parse(pattern, pattern, nil, "")
	return ecma402dtf.Options{
		Weekday:                format.Weekday,
		Era:                    format.Era,
		Year:                   format.Year,
		Month:                  format.Month,
		Day:                    format.Day,
		Hour:                   format.Hour,
		HourCycle:              format.HourCycle,
		Minute:                 format.Minute,
		Second:                 format.Second,
		DayPeriod:              format.DayPeriod,
		FractionalSecondDigits: format.FractionalSecondDigits,
		TimeZoneName:           format.TimeZoneName,
	}
}

func matcherOptions(resolved ResolvedOptions, uses24Hour bool) ecma402dtf.Options {
	hour12 := resolved.Hour12
	hourCycle := ecma402.ResolvedScalarValue(resolved.HourCycle)
	if hour12 == nil {
		hour12 = hourCycleImpliesHour12(HourCycle(hourCycle))
		if hour12 == nil {
			hour12 = ecma402.ResolvedScalar(!uses24Hour)
		}
	}
	return ecma402dtf.Options{
		Weekday:                ecma402dtf.FieldStyle(ecma402.ResolvedScalarValue(resolved.Weekday)),
		Era:                    ecma402dtf.FieldStyle(ecma402.ResolvedScalarValue(resolved.Era)),
		Year:                   ecma402dtf.NumericStyle(ecma402.ResolvedScalarValue(resolved.Year)),
		Month:                  ecma402dtf.FieldStyle(ecma402.ResolvedScalarValue(resolved.Month)),
		Day:                    ecma402dtf.NumericStyle(ecma402.ResolvedScalarValue(resolved.Day)),
		Hour:                   ecma402dtf.NumericStyle(ecma402.ResolvedScalarValue(resolved.Hour)),
		HourCycle:              ecma402dtf.HourCycle(hourCycle),
		Minute:                 ecma402dtf.NumericStyle(ecma402.ResolvedScalarValue(resolved.Minute)),
		Second:                 ecma402dtf.NumericStyle(ecma402.ResolvedScalarValue(resolved.Second)),
		DayPeriod:              ecma402dtf.FieldStyle(ecma402.ResolvedScalarValue(resolved.DayPeriod)),
		FractionalSecondDigits: ecma402.ResolvedScalarValue(resolved.FractionalSecondDigits),
		TimeZoneName:           ecma402dtf.TimeZoneName(ecma402.ResolvedScalarValue(resolved.TimeZoneName)),
		Hour12:                 hour12,
	}
}

func dateOnlyMatcherOptions(opts ecma402dtf.Options) ecma402dtf.Options {
	opts.Hour = ""
	opts.HourCycle = ""
	opts.Minute = ""
	opts.Second = ""
	opts.DayPeriod = ""
	opts.FractionalSecondDigits = 0
	opts.TimeZoneName = ""
	opts.Hour12 = nil
	return opts
}

func timeOnlyMatcherOptions(opts ecma402dtf.Options) ecma402dtf.Options {
	opts.Weekday = ""
	opts.Era = ""
	opts.Year = ""
	opts.Month = ""
	opts.Day = ""
	return opts
}

func hasDateMatcherOptions(opts ecma402dtf.Options) bool {
	return opts.Weekday != "" || opts.Era != "" || opts.Year != "" || opts.Month != "" || opts.Day != ""
}

func hasTimeMatcherOptions(opts ecma402dtf.Options) bool {
	return opts.Hour != "" || opts.Minute != "" || opts.Second != "" || opts.DayPeriod != "" || opts.FractionalSecondDigits != 0 || opts.TimeZoneName != ""
}

func hasDateFormatOnly(format ecma402dtf.Formats) bool {
	return hasDateFormatFields(format) && !hasTimeFormatFields(format)
}

func hasTimeFormatOnly(format ecma402dtf.Formats) bool {
	return hasTimeFormatFields(format) && !hasDateFormatFields(format)
}

func hasDateFormatFields(format ecma402dtf.Formats) bool {
	return format.Weekday != "" || format.Era != "" || format.Year != "" || format.Month != "" || format.Day != ""
}

func hasTimeFormatFields(format ecma402dtf.Formats) bool {
	return format.Hour != "" || format.Minute != "" || format.Second != "" || format.DayPeriod != "" || format.FractionalSecondDigits != 0 || format.TimeZoneName != ""
}

func matchComponentPattern(formatMatcher FormatMatcher, opts ecma402dtf.Options, candidates []ecma402dtf.Formats, appendItems map[string]string) (ecma402dtf.Formats, bool) {
	if len(candidates) == 0 {
		return ecma402dtf.Formats{}, false
	}
	var format ecma402dtf.Formats
	if formatMatcher == BasicFormatMatcher {
		format = ecma402dtf.MatchBasic(opts, candidates)
	} else {
		format = ecma402dtf.Match(opts, candidates)
	}
	if opts.TimeZoneName != "" && format.Pattern != "" && !format.PatternHasTimeZoneName {
		format = appendTimeZoneName(format, opts.TimeZoneName, appendItems)
	}
	return format, format.Pattern != ""
}

func appendTimeZoneName(format ecma402dtf.Formats, style ecma402dtf.TimeZoneName, appendItems map[string]string) ecma402dtf.Formats {
	text := appendItems["Timezone"]
	if text == "" {
		text = "{0} {1}"
	}
	field := ecma402dtf.TimeZonePatternField(style)
	format.Pattern = pattern.FormatIndexed(text, format.Pattern, field)
	format.Skeleton += field
	format.TimeZoneName = style
	format.PatternHasTimeZoneName = true
	return format
}

func insertFractionalSecondField(pattern string, digits int) string {
	field := "." + strings.Repeat("S", digits)
	for i := 0; i < len(pattern); {
		if pattern[i] == '\'' {
			i = skipQuotedPatternLiteral(pattern, i)
			continue
		}
		j := i + 1
		for j < len(pattern) && pattern[j] == pattern[i] {
			j++
		}
		if pattern[i] == 's' {
			return pattern[:j] + field + pattern[j:]
		}
		i = j
	}
	return pattern + field
}

func skipQuotedPatternLiteral(pattern string, start int) int {
	for i := start + 1; i < len(pattern); i++ {
		if pattern[i] != '\'' {
			continue
		}
		if i+1 < len(pattern) && pattern[i+1] == '\'' {
			i++
			continue
		}
		return i + 1
	}
	return len(pattern)
}
