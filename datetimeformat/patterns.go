package datetimeformat

import (
	"slices"
	"strings"
	"time"

	"github.com/agentable/go-intl/internal/cldr"
	ecma402dtf "github.com/agentable/go-intl/internal/ecma402/datetimeformat"
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

func (f *DateTimeFormat) selectPattern() selectedPattern {
	resolved := f.resolved
	if resolved.DateStyle != "" && resolved.TimeStyle != "" {
		datePattern := dateStylePattern(f.gregorian, resolved.DateStyle)
		timePattern := timeStylePattern(f.gregorian, resolved.TimeStyle)
		dateFormat, _ := f.styleDateIntervalFormat(datePattern)
		timeFormat, _ := f.styleTimeIntervalFormat(timePattern)
		return selectedPattern{
			kind:                patternDateTime,
			date:                datePattern,
			time:                timePattern,
			dateTime:            dateTimeStylePattern(f.gregorian, resolved.DateStyle),
			dateSkeleton:        dateFormat.Skeleton,
			timeSkeleton:        timeFormat.Skeleton,
			dateIntervalOptions: stylePatternOptions(datePattern),
			timeIntervalOptions: timeStylePatternOptions(timePattern),
		}
	}
	if resolved.DateStyle != "" {
		datePattern := dateStylePattern(f.gregorian, resolved.DateStyle)
		dateFormat, _ := f.styleDateIntervalFormat(datePattern)
		return selectedPattern{
			kind:                patternDate,
			date:                datePattern,
			dateSkeleton:        dateFormat.Skeleton,
			dateIntervalOptions: stylePatternOptions(datePattern),
		}
	}
	if resolved.TimeStyle != "" {
		timePattern := timeStylePattern(f.gregorian, resolved.TimeStyle)
		timeFormat, _ := f.styleTimeIntervalFormat(timePattern)
		return selectedPattern{
			kind:                patternTime,
			time:                timePattern,
			timeSkeleton:        timeFormat.Skeleton,
			timeIntervalOptions: timeStylePatternOptions(timePattern),
		}
	}
	if pattern, ok := f.componentPattern(); ok {
		return pattern
	}
	return selectedPattern{}
}

func (p selectedPattern) parts(f *DateTimeFormat, t time.Time) ([]Part, bool) {
	switch p.kind {
	case patternDate:
		return f.formatDatePattern(p.date, t), true
	case patternTime:
		return f.formatTimePattern(p.time, t), true
	case patternDateTime:
		dateParts := f.formatDatePattern(p.date, t)
		timeParts := f.formatTimePattern(p.time, t)
		return interpolateDateTimeParts(p.dateTime, dateParts, timeParts), true
	case patternNone:
	}
	return nil, false
}

func dateStylePattern(g cldr.Gregorian, style Style) string {
	switch style {
	case FullDateTimeStyle:
		return g.DateFormats[0]
	case LongDateTimeStyle:
		return g.DateFormats[1]
	case MediumDateTimeStyle:
		return g.DateFormats[2]
	case ShortDateTimeStyle:
		return g.DateFormats[3]
	}
	return ""
}

func timeStylePattern(g cldr.Gregorian, style Style) string {
	switch style {
	case FullDateTimeStyle:
		return g.TimeFormats[0]
	case LongDateTimeStyle:
		return g.TimeFormats[1]
	case MediumDateTimeStyle:
		return g.TimeFormats[2]
	case ShortDateTimeStyle:
		return g.TimeFormats[3]
	}
	return ""
}

func dateTimeStylePattern(g cldr.Gregorian, style Style) string {
	if pattern := dateTimeAtStylePattern(g, style); pattern != "" {
		return pattern
	}
	return dateTimeStandardStylePattern(g, style)
}

func dateTimeAtStylePattern(g cldr.Gregorian, style Style) string {
	switch style {
	case FullDateTimeStyle:
		return g.DateTimeAtFormats[0]
	case LongDateTimeStyle:
		return g.DateTimeAtFormats[1]
	case MediumDateTimeStyle:
		return g.DateTimeAtFormats[2]
	case ShortDateTimeStyle:
		return g.DateTimeAtFormats[3]
	}
	return ""
}

func dateTimeStandardStylePattern(g cldr.Gregorian, style Style) string {
	switch style {
	case FullDateTimeStyle:
		return g.DateTimeFormats[0]
	case LongDateTimeStyle:
		return g.DateTimeFormats[1]
	case MediumDateTimeStyle:
		return g.DateTimeFormats[2]
	case ShortDateTimeStyle:
		return g.DateTimeFormats[3]
	}
	return ""
}

func (f *DateTimeFormat) componentPattern() (selectedPattern, bool) {
	opts := f.matcherOptions()
	hasDate := hasDateMatcherOptions(opts)
	hasTime := hasTimeMatcherOptions(opts)
	switch {
	case hasDate && hasTime:
		dateFormat, dateOK := f.matchComponentPattern(dateOnlyMatcherOptions(opts), f.datePatternCandidates())
		timeFormat, timeOK := f.matchComponentPattern(timeOnlyMatcherOptions(opts), f.timePatternCandidates(opts))
		if !dateOK || !timeOK {
			return selectedPattern{}, false
		}
		return selectedPattern{
			kind:                patternDateTime,
			date:                dateFormat.Pattern,
			time:                timeFormat.Pattern,
			dateTime:            f.componentDateTimePattern(opts, dateFormat),
			dateSkeleton:        dateFormat.Skeleton,
			timeSkeleton:        timeFormat.Skeleton,
			dateIntervalOptions: dateOnlyMatcherOptions(opts),
			timeIntervalOptions: timeOnlyMatcherOptions(opts),
		}, true
	case hasDate:
		format, ok := f.matchComponentPattern(opts, f.datePatternCandidates())
		if !ok {
			return selectedPattern{}, false
		}
		return selectedPattern{kind: patternDate, date: format.Pattern, dateSkeleton: format.Skeleton, dateIntervalOptions: opts}, true
	case hasTime:
		format, ok := f.matchComponentPattern(opts, f.timePatternCandidates(opts))
		if !ok {
			return selectedPattern{}, false
		}
		return selectedPattern{kind: patternTime, time: format.Pattern, timeSkeleton: format.Skeleton, timeIntervalOptions: opts}, true
	default:
		return selectedPattern{}, false
	}
}

func (f *DateTimeFormat) componentDateTimePattern(opts ecma402dtf.Options, dateFormat ecma402dtf.Formats) string {
	style := MediumDateTimeStyle
	switch {
	case dateFormat.Month == ecma402dtf.FieldLong && dateFormat.Weekday == ecma402dtf.FieldLong:
		style = FullDateTimeStyle
	case dateFormat.Month == ecma402dtf.FieldLong:
		style = LongDateTimeStyle
	case opts.Year == ecma402dtf.Numeric2Digit && (opts.Month == ecma402dtf.FieldNumeric || opts.Month == ecma402dtf.Field2Digit):
		style = ShortDateTimeStyle
	}
	return dateTimeStylePattern(f.gregorian, style)
}

func (f *DateTimeFormat) styleDateIntervalFormat(pattern string) (ecma402dtf.Formats, bool) {
	return f.matchComponentPattern(stylePatternOptions(pattern), f.datePatternCandidates())
}

func (f *DateTimeFormat) styleTimeIntervalFormat(pattern string) (ecma402dtf.Formats, bool) {
	opts := timeStylePatternOptions(pattern)
	return f.matchComponentPattern(opts, f.timePatternCandidates(opts))
}

func timeStylePatternOptions(pattern string) ecma402dtf.Options {
	opts := stylePatternOptions(pattern)
	opts.DayPeriod = ""
	switch opts.HourCycle {
	case ecma402dtf.HourCycleH11, ecma402dtf.HourCycleH12:
		v := true
		opts.Hour12 = &v
	case ecma402dtf.HourCycleH23, ecma402dtf.HourCycleH24:
		v := false
		opts.Hour12 = &v
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

func (f *DateTimeFormat) matcherOptions() ecma402dtf.Options {
	resolved := f.resolved
	hour12 := resolved.Hour12
	if hour12 == nil {
		switch resolved.HourCycle {
		case H11HourCycle, H12HourCycle:
			v := true
			hour12 = &v
		case H23HourCycle, H24HourCycle:
			v := false
			hour12 = &v
		default:
			v := !f.uses24HourTime()
			hour12 = &v
		}
	}
	return ecma402dtf.Options{
		Weekday:                ecma402dtf.FieldStyle(resolved.Weekday),
		Era:                    ecma402dtf.FieldStyle(resolved.Era),
		Year:                   ecma402dtf.NumericStyle(resolved.Year),
		Month:                  ecma402dtf.FieldStyle(resolved.Month),
		Day:                    ecma402dtf.NumericStyle(resolved.Day),
		Hour:                   ecma402dtf.NumericStyle(resolved.Hour),
		HourCycle:              ecma402dtf.HourCycle(resolved.HourCycle),
		Minute:                 ecma402dtf.NumericStyle(resolved.Minute),
		Second:                 ecma402dtf.NumericStyle(resolved.Second),
		DayPeriod:              ecma402dtf.FieldStyle(resolved.DayPeriod),
		FractionalSecondDigits: resolved.FractionalSecondDigits,
		TimeZoneName:           ecma402dtf.TimeZoneName(resolved.TimeZoneName),
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

func (f *DateTimeFormat) datePatternCandidates() []ecma402dtf.Formats {
	return filterFormats(f.availableFormatCandidates(), hasDateFormatOnly)
}

func (f *DateTimeFormat) timePatternCandidates(opts ecma402dtf.Options) []ecma402dtf.Formats {
	candidates := filterFormats(f.availableFormatCandidates(), hasTimeFormatOnly)
	if opts.FractionalSecondDigits != 0 {
		candidates = appendFractionalSecondCandidates(candidates, opts.FractionalSecondDigits)
	}
	return candidates
}

func (f *DateTimeFormat) availableFormatCandidates() []ecma402dtf.Formats {
	keys := make([]string, 0, len(f.gregorian.AvailableFormats))
	for skeleton := range f.gregorian.AvailableFormats {
		keys = append(keys, skeleton)
	}
	slices.Sort(keys)

	formats := make([]ecma402dtf.Formats, 0, len(keys))
	for _, skeleton := range keys {
		pattern := f.gregorian.AvailableFormats[skeleton]
		if pattern == "" {
			continue
		}
		formats = append(formats, ecma402dtf.Parse(skeleton, pattern, nil, ""))
	}
	return formats
}

func filterFormats(formats []ecma402dtf.Formats, keep func(ecma402dtf.Formats) bool) []ecma402dtf.Formats {
	out := make([]ecma402dtf.Formats, 0, len(formats))
	for _, format := range formats {
		if keep(format) {
			out = append(out, format)
		}
	}
	return out
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

func (f *DateTimeFormat) matchComponentPattern(opts ecma402dtf.Options, candidates []ecma402dtf.Formats) (ecma402dtf.Formats, bool) {
	if len(candidates) == 0 {
		return ecma402dtf.Formats{}, false
	}
	var format ecma402dtf.Formats
	if f.formatMatcher == BasicFormatMatcher {
		format = ecma402dtf.MatchBasic(opts, candidates)
	} else {
		format = ecma402dtf.Match(opts, candidates)
	}
	if opts.TimeZoneName != "" && format.Pattern != "" && !patternHasTimeZoneName(format.Pattern) {
		format = f.appendTimeZoneName(format, opts.TimeZoneName)
	}
	return format, format.Pattern != ""
}

func (f *DateTimeFormat) appendTimeZoneName(format ecma402dtf.Formats, style ecma402dtf.TimeZoneName) ecma402dtf.Formats {
	pattern := f.gregorian.AppendItems["Timezone"]
	if pattern == "" {
		pattern = "{0} {1}"
	}
	field := timeZonePatternField(style)
	format.Pattern = strings.ReplaceAll(strings.ReplaceAll(pattern, "{0}", format.Pattern), "{1}", field)
	format.Skeleton += field
	format.TimeZoneName = style
	return format
}

func patternHasTimeZoneName(pattern string) bool {
	for i := 0; i < len(pattern); {
		if pattern[i] == '\'' {
			i = skipQuotedPatternLiteral(pattern, i)
			continue
		}
		if strings.IndexByte("zZOvVxX", pattern[i]) >= 0 {
			return true
		}
		j := i + 1
		for j < len(pattern) && pattern[j] == pattern[i] {
			j++
		}
		i = j
	}
	return false
}

func timeZonePatternField(style ecma402dtf.TimeZoneName) string {
	switch style {
	case ecma402dtf.TimeZoneNameShort:
		return "z"
	case ecma402dtf.TimeZoneNameLong:
		return "zzzz"
	case ecma402dtf.TimeZoneNameShortOffset:
		return "O"
	case ecma402dtf.TimeZoneNameLongOffset:
		return "OOOO"
	case ecma402dtf.TimeZoneNameShortGeneric:
		return "v"
	case ecma402dtf.TimeZoneNameLongGeneric:
		return "vvvv"
	default:
		return "z"
	}
}

func appendFractionalSecondCandidates(formats []ecma402dtf.Formats, digits int) []ecma402dtf.Formats {
	out := make([]ecma402dtf.Formats, 0, len(formats)*2)
	out = append(out, formats...)
	for _, format := range formats {
		if format.Second == "" || format.FractionalSecondDigits != 0 {
			continue
		}
		withFraction := format
		withFraction.Skeleton += strings.Repeat("S", digits)
		withFraction.Pattern = insertFractionalSecondField(format.Pattern, digits)
		withFraction.FractionalSecondDigits = digits
		out = append(out, withFraction)
	}
	return out
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
