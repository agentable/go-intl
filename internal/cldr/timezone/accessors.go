// Hand-written accessor layer for the timezone domain. Supported-zone queries
// read only the narrow index; display-name and format queries decode their own
// blobs on demand.
//
// Locale is the shared CLDR kernel handle. Timezone-specific queries stay as
// package functions so DateTimeFormat imports only the timezone domain.

package timezone

import (
	"strconv"
	"strings"

	"github.com/agentable/go-intl/internal/numbering"
)

// TimeZoneName is one ECMA-402 timeZoneName option value.
type TimeZoneName string

const (
	TimeZoneNameShort        TimeZoneName = "short"
	TimeZoneNameLong         TimeZoneName = "long"
	TimeZoneNameShortOffset  TimeZoneName = "shortOffset"
	TimeZoneNameLongOffset   TimeZoneName = "longOffset"
	TimeZoneNameShortGeneric TimeZoneName = "shortGeneric"
	TimeZoneNameLongGeneric  TimeZoneName = "longGeneric"
)

type zoneNameKind string

const (
	zoneNameLongGeneric   zoneNameKind = "long-generic"
	zoneNameLongStandard  zoneNameKind = "long-standard"
	zoneNameLongDaylight  zoneNameKind = "long-daylight"
	zoneNameShortGeneric  zoneNameKind = "short-generic"
	zoneNameShortStandard zoneNameKind = "short-standard"
	zoneNameShortDaylight zoneNameKind = "short-daylight"
)

const (
	rootGMTFormat          = "GMT{0}"
	rootGMTZeroFormat      = "GMT"
	rootRegionFormat       = "{0}"
	rootPositiveHourFormat = "+HH:mm"
	rootNegativeHourFormat = "-HH:mm"
	rootHourFormat         = rootPositiveHourFormat + ";" + rootNegativeHourFormat
)

// TimeZoneMetazone returns the metazone in force for the zone at the given
// unix-milli instant, or "" when no period covers it.
func TimeZoneMetazone(zone string, instant int64) string {
	zone = displayKey(zone)
	metazonePeriodOnce.Do(loadMetazonePeriods)
	for _, period := range metazonePeriodsForZone(zone) {
		if instant >= period.start && instant < period.end {
			return period.metazone
		}
	}
	return ""
}

func metazonePeriodsForZone(zone string) []metazonePeriod {
	periods, ok := zoneToMetazones[zone]
	if !ok {
		return nil
	}
	return periods
}

// TimeZoneDisplayName resolves zone-specific and metazone names. Only generic
// styles use exemplar-city fallback; specific names fall back to localized GMT.
func TimeZoneDisplayName(loc Locale, zone string, form TimeZoneName, isDST, useStandard bool, instant int64, offsetMs int64, numberingSystem string) string {
	zone = displayKey(zone)
	if isOffsetTimeZoneName(form) {
		return GMTOffsetName(loc, offsetMs, form, numberingSystem)
	}
	kind := displayNameKind(form, isDST)
	if name := zoneSpecificName(loc, zone, kind); name != "" {
		return name
	}
	metazone := TimeZoneMetazone(zone, instant)
	if useStandard && metazone != "" && (form == TimeZoneNameShortGeneric || form == TimeZoneNameLongGeneric) {
		standardKind := zoneNameLongStandard
		if form == TimeZoneNameShortGeneric {
			standardKind = zoneNameShortStandard
		}
		standard := zoneSpecificName(loc, zone, standardKind)
		if standard == "" {
			standard = metazoneName(loc, metazone, standardKind)
		}
		if standard != "" && !strings.EqualFold(standard, metazoneName(loc, metazone, kind)) {
			return standard
		}
	}
	if name := metazoneName(loc, metazone, kind); name != "" {
		return name
	}
	if form == TimeZoneNameShortGeneric || form == TimeZoneNameLongGeneric {
		if city := exemplarCity(loc, zone); city != "" {
			return strings.Replace(timeZoneFormats(loc).regionFormat, "{0}", city, 1)
		}
	}
	return GMTOffsetName(loc, offsetMs, form, numberingSystem)
}

func isOffsetTimeZoneName(form TimeZoneName) bool {
	return form == TimeZoneNameShortOffset || form == TimeZoneNameLongOffset
}

func displayNameKind(form TimeZoneName, isDST bool) zoneNameKind {
	switch form {
	case TimeZoneNameLong:
		if isDST {
			return zoneNameLongDaylight
		}
		return zoneNameLongStandard
	case TimeZoneNameShort:
		if isDST {
			return zoneNameShortDaylight
		}
		return zoneNameShortStandard
	case TimeZoneNameShortGeneric:
		return zoneNameShortGeneric
	default:
		return zoneNameLongGeneric
	}
}

func zoneSpecificName(loc Locale, zone string, kind zoneNameKind) string {
	namesOnce.Do(loadNames)
	return timeZoneNameValue(timeZoneNameRecord(timeZoneNamesByLocale, loc, zone), kind)
}

func metazoneName(loc Locale, metazone string, kind zoneNameKind) string {
	namesOnce.Do(loadNames)
	return timeZoneNameValue(timeZoneNameRecord(metazoneNamesByLocale, loc, metazone), kind)
}

func timeZoneNameRecord(records map[Locale]map[string]metazoneNames, loc Locale, name string) metazoneNames {
	byName := records[loc]
	if byName == nil {
		return metazoneNames{}
	}
	return byName[name]
}

func timeZoneNameValue(data metazoneNames, kind zoneNameKind) string {
	switch kind {
	case zoneNameLongStandard:
		return data.longStandard
	case zoneNameLongDaylight:
		return data.longDaylight
	case zoneNameShortGeneric:
		return data.shortGeneric
	case zoneNameShortStandard:
		return data.shortStandard
	case zoneNameShortDaylight:
		return data.shortDaylight
	default:
		return data.longGeneric
	}
}

func exemplarCity(loc Locale, zone string) string {
	namesOnce.Do(loadNames)
	cities := exemplarCitiesByLocale[loc]
	if cities == nil {
		return ""
	}
	return cities[zone]
}

// timeZoneFormats resolves a locale's GMT/hour offset and location formats,
// applying CLDR root defaults when a field is absent.
func timeZoneFormats(loc Locale) timeZoneFormatRefs {
	formatsOnce.Do(loadFormats)
	formats := timeZoneFormatsForLocale(loc)
	if formats.gmtFormat == "" {
		formats.gmtFormat = rootGMTFormat
	}
	if formats.gmtZeroFormat == "" {
		formats.gmtZeroFormat = rootGMTZeroFormat
	}
	if formats.hourFormat == "" {
		formats.hourFormat = rootHourFormat
	}
	if formats.regionFormat == "" {
		formats.regionFormat = rootRegionFormat
	}
	return formats
}

func timeZoneFormatsForLocale(loc Locale) timeZoneFormatRefs {
	formats, ok := timeZoneFormatsByLocale[loc]
	if !ok {
		return timeZoneFormatRefs{}
	}
	return formats
}

// GMTOffsetName formats an offset time-zone name using locale GMT patterns.
func GMTOffsetName(loc Locale, offsetMs int64, form TimeZoneName, numberingSystem string) string {
	formats := timeZoneFormats(loc)
	long := form == TimeZoneNameLongOffset || form == TimeZoneNameLong || form == TimeZoneNameLongGeneric
	offset := offsetPattern(formats.hourFormat, offsetMs, long, numberingSystem)
	if offset == "" && !long {
		return formats.gmtZeroFormat
	}
	return strings.ReplaceAll(formats.gmtFormat, "{0}", offset)
}

func offsetPattern(hourFormat string, offsetMs int64, long bool, numberingSystem string) string {
	positive, negative, ok := strings.Cut(hourFormat, ";")
	if !ok {
		positive, negative = rootPositiveHourFormat, rootNegativeHourFormat
	}
	pattern := positive
	if offsetMs < 0 {
		pattern = negative
		offsetMs = -offsetMs
	}
	totalSeconds := offsetMs / 1000
	hours := int(totalSeconds / 3600)
	minutes := int(totalSeconds / 60 % 60)
	seconds := int(totalSeconds % 60)
	if seconds != 0 {
		pattern = expandSecondField(pattern)
	} else if !long {
		if hours == 0 && minutes == 0 {
			return ""
		}
		if minutes == 0 {
			pattern = removeMinuteField(pattern)
		}
	}
	out := replaceOffsetFields(pattern, hours, minutes, long, numberingSystem)
	return replaceFieldRun(out, 's', numbering.LocalizeDigits(twoDigitASCII(seconds), numberingSystem))
}

// CLDR supplies HM; ICU derives HMS by repeating the hour/minute separator.
func expandSecondField(pattern string) string {
	h := strings.LastIndexByte(pattern, 'H')
	m := strings.IndexByte(pattern, 'm')
	if h < 0 || m <= h {
		return pattern
	}
	end := m
	for end < len(pattern) && pattern[end] == 'm' {
		end++
	}
	return pattern[:end] + pattern[h+1:m] + "ss" + pattern[end:]
}

func removeMinuteField(pattern string) string {
	i := strings.IndexByte(pattern, 'H')
	if i < 0 {
		return pattern
	}
	for i < len(pattern) && pattern[i] == 'H' {
		i++
	}
	return pattern[:i]
}

func replaceOffsetFields(pattern string, hours int, minutes int, padded bool, numberingSystem string) string {
	hour := strconv.Itoa(hours)
	minute := twoDigitASCII(minutes)
	if padded {
		hour = twoDigitASCII(hours)
	}
	out := replaceFieldRun(pattern, 'H', numbering.LocalizeDigits(hour, numberingSystem))
	out = replaceFieldRun(out, 'm', numbering.LocalizeDigits(minute, numberingSystem))
	return out
}

func replaceFieldRun(pattern string, field byte, replacement string) string {
	i := strings.IndexByte(pattern, field)
	if i < 0 {
		return pattern
	}
	end := i
	for end < len(pattern) && pattern[end] == field {
		end++
	}
	return pattern[:i] + replacement + pattern[end:]
}

func twoDigitASCII(value int) string {
	if value < 10 {
		return "0" + strconv.Itoa(value)
	}
	return strconv.Itoa(value)
}

func displayKey(zone string) string {
	if key := displayKeys()[zone]; key != "" {
		return key
	}
	return zone
}
