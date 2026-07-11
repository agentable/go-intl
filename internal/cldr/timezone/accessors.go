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

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
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
	rootPositiveHourFormat = "+HH:mm"
	rootNegativeHourFormat = "-HH:mm"
	rootHourFormat         = rootPositiveHourFormat + ";" + rootNegativeHourFormat
)

// CanonicalTimeZoneLink resolves a deprecated IANA link to its canonical name,
// forwarding to the locale kernel which owns the small static link table.
func CanonicalTimeZoneLink(name string) string {
	return cldrlocale.CanonicalTimeZoneLink(name)
}

// SupportedTimeZones returns the canonicalized supported IANA zone names in
// sorted order. It reads only the narrow supported blob and never triggers the
// metazone-period or names blob decode.
func SupportedTimeZones() []string {
	return supported.Get()
}

// TimeZoneMetazone returns the metazone in force for the zone at the given
// unix-milli instant, or "" when no period covers it.
func TimeZoneMetazone(zone string, instant int64) string {
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

// TimeZoneDisplayName resolves the localized display name for a zone, falling
// back through zone-specific names, metazone names, exemplar city, and finally
// the GMT offset.
func TimeZoneDisplayName(loc Locale, zone string, form TimeZoneName, isDST bool, instant int64, offsetMs int64) string {
	if isOffsetTimeZoneName(form) {
		return GMTOffsetName(loc, offsetMs, form)
	}
	kind := displayNameKind(form, isDST)
	if name := zoneSpecificName(loc, zone, kind); name != "" {
		return name
	}
	if name := metazoneName(loc, TimeZoneMetazone(zone, instant), kind); name != "" {
		return name
	}
	if city := exemplarCity(loc, zone); city != "" {
		return "Time in " + city
	}
	return GMTOffsetName(loc, offsetMs, form)
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

// timeZoneFormats resolves a locale's GMT/hour offset formats, applying CLDR
// root defaults when a field is absent.
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
func GMTOffsetName(loc Locale, offsetMs int64, form TimeZoneName) string {
	formats := timeZoneFormats(loc)
	long := form == TimeZoneNameLongOffset || form == TimeZoneNameLong
	offset := offsetPattern(formats.hourFormat, offsetMs, long)
	if offset == "" && !long {
		return formats.gmtZeroFormat
	}
	return strings.ReplaceAll(formats.gmtFormat, "{0}", offset)
}

func offsetPattern(hourFormat string, offsetMs int64, long bool) string {
	positive, negative, ok := strings.Cut(hourFormat, ";")
	if !ok {
		positive, negative = rootPositiveHourFormat, rootNegativeHourFormat
	}
	pattern := positive
	if offsetMs < 0 {
		pattern = negative
		offsetMs = -offsetMs
	}
	totalMinutes := offsetMs / 60000
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	if !long {
		if hours == 0 && minutes == 0 {
			return ""
		}
		if minutes == 0 {
			pattern = removeMinuteField(pattern)
		}
		return replaceOffsetFields(pattern, int(hours), int(minutes), false)
	}
	return replaceOffsetFields(pattern, int(hours), int(minutes), true)
}

func removeMinuteField(pattern string) string {
	i := strings.IndexByte(pattern, 'm')
	if i < 0 {
		return pattern
	}
	start := i
	if start > 0 && pattern[start-1] == ':' {
		start--
	}
	end := i
	for end < len(pattern) && pattern[end] == 'm' {
		end++
	}
	return pattern[:start] + pattern[end:]
}

func replaceOffsetFields(pattern string, hours int, minutes int, padded bool) string {
	hour := strconv.Itoa(hours)
	minute := strconv.Itoa(minutes)
	if padded {
		hour = twoDigitASCII(hours)
		minute = twoDigitASCII(minutes)
	}
	out := replaceFieldRun(pattern, 'H', hour)
	out = replaceFieldRun(out, 'm', minute)
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
