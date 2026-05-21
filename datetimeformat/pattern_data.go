package datetimeformat

import (
	"slices"
	"strings"
	"sync"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	ecma402dtf "github.com/agentable/go-intl/internal/ecma402/datetimeformat"
)

type patternData struct {
	dateCandidates           []ecma402dtf.Formats
	timeCandidates           []ecma402dtf.Formats
	fractionalTimeCandidates [3][]ecma402dtf.Formats
	styleOptions             [4]stylePatternData
}

type stylePatternData struct {
	dateOptions ecma402dtf.Options
	timeOptions ecma402dtf.Options
}

var patternDataCache sync.Map
var gregorianDataCache sync.Map

func gregorianDataFor(loc cldrdate.Locale) cldrdate.Gregorian {
	if loc == cldrdate.Undefined {
		return cldrdate.Gregorian{}
	}
	if data, ok := gregorianDataCache.Load(loc); ok {
		return data.(cldrdate.Gregorian)
	}
	data := cldrdate.GregorianFor(loc)
	actual, _ := gregorianDataCache.LoadOrStore(loc, data)
	return actual.(cldrdate.Gregorian)
}

func patternDataFor(loc cldrdate.Locale, gregorian cldrdate.Gregorian) *patternData {
	if loc == cldrdate.Undefined {
		return buildPatternData(gregorian)
	}
	if data, ok := patternDataCache.Load(loc); ok {
		return data.(*patternData)
	}
	data := buildPatternData(gregorian)
	actual, _ := patternDataCache.LoadOrStore(loc, data)
	return actual.(*patternData)
}

func buildPatternData(gregorian cldrdate.Gregorian) *patternData {
	formats := availableFormatCandidates(gregorian)
	data := &patternData{
		dateCandidates: filterFormats(formats, hasDateFormatOnly),
		timeCandidates: filterFormats(formats, hasTimeFormatOnly),
	}
	for i := range data.fractionalTimeCandidates {
		data.fractionalTimeCandidates[i] = appendFractionalSecondCandidates(data.timeCandidates, i+1)
	}
	for _, style := range dateTimeStyles {
		idx := dateTimeStyleIndex(style)
		data.styleOptions[idx] = stylePatternData{
			dateOptions: stylePatternOptions(dateStylePattern(gregorian, style)),
			timeOptions: timeStylePatternOptions(timeStylePattern(gregorian, style)),
		}
	}
	return data
}

func availableFormatCandidates(gregorian cldrdate.Gregorian) []ecma402dtf.Formats {
	keys := make([]string, 0, len(gregorian.AvailableFormats))
	for skeleton := range gregorian.AvailableFormats {
		keys = append(keys, skeleton)
	}
	slices.Sort(keys)

	formats := make([]ecma402dtf.Formats, 0, len(keys))
	for _, skeleton := range keys {
		pattern := gregorian.AvailableFormats[skeleton]
		if pattern == "" {
			continue
		}
		formats = append(formats, ecma402dtf.Parse(skeleton, pattern, nil, ""))
	}
	return formats
}

func (d *patternData) timePatternCandidates(digits int) []ecma402dtf.Formats {
	if digits >= 1 && digits <= len(d.fractionalTimeCandidates) {
		return d.fractionalTimeCandidates[digits-1]
	}
	return d.timeCandidates
}

func (d *patternData) style(style Style) stylePatternData {
	return d.styleOptions[dateTimeStyleIndex(style)]
}

var dateTimeStyles = [...]Style{
	FullDateTimeStyle,
	LongDateTimeStyle,
	MediumDateTimeStyle,
	ShortDateTimeStyle,
}

func dateTimeStyleIndex(style Style) int {
	switch style {
	case FullDateTimeStyle:
		return 0
	case LongDateTimeStyle:
		return 1
	case MediumDateTimeStyle:
		return 2
	case ShortDateTimeStyle:
		return 3
	default:
		return 2
	}
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
