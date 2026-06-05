// Hand-written accessor layer for the timezone domain. The query semantics
// mirror the legacy root cldr metazone accessors exactly, so DateTimeFormat
// output is byte-for-byte unchanged.
//
// The legacy code attached TimeZoneName / MetazoneName / ExemplarCity /
// TimeZoneFormats as methods on the root Locale type. Here Locale is a type
// alias for the kernel handle, so those are package-level functions taking the
// locale as their first argument; the change is internal to this package and
// invisible to callers, who only see CanonicalTimeZoneLink, TimeZoneDisplayName,
// and GMTOffsetName.

package timezone

import (
	"strconv"
	"strings"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// TimeZoneName is one ECMA-402 timeZoneName option value. It mirrors the legacy
// root TimeZoneName type.
type TimeZoneName string

const (
	TimeZoneNameShort        TimeZoneName = "short"
	TimeZoneNameLong         TimeZoneName = "long"
	TimeZoneNameShortOffset  TimeZoneName = "shortOffset"
	TimeZoneNameLongOffset   TimeZoneName = "longOffset"
	TimeZoneNameShortGeneric TimeZoneName = "shortGeneric"
	TimeZoneNameLongGeneric  TimeZoneName = "longGeneric"
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
	supportedOnce.Do(loadSupported)
	tags := make([]string, len(supportedTags))
	copy(tags, supportedTags)
	return tags
}

// ZoneToMetazone returns the metazone in force for the zone at instant zero,
// matching the legacy root accessor.
func ZoneToMetazone(zone string) string {
	return TimeZoneMetazone(zone, 0)
}

// TimeZoneMetazone returns the metazone in force for the zone at the given
// unix-milli instant, or "" when no period covers it.
func TimeZoneMetazone(zone string, instant int64) string {
	metazonePeriodOnce.Do(loadMetazonePeriods)
	for _, period := range zoneToMetazones[zone] {
		if instant >= period.start && instant < period.end {
			return period.metazone
		}
	}
	return ""
}

// TimeZoneDisplayName resolves the localized display name for a zone, falling
// back through zone-specific names, metazone names, exemplar city, and finally
// the GMT offset, exactly as the legacy root implementation did.
func TimeZoneDisplayName(loc Locale, zone string, form TimeZoneName, isDST bool, instant int64, offsetMs int64) string {
	kind := "long-generic"
	switch form {
	case TimeZoneNameLong:
		if isDST {
			kind = "long-daylight"
		} else {
			kind = "long-standard"
		}
	case TimeZoneNameShort:
		if isDST {
			kind = "short-daylight"
		} else {
			kind = "short-standard"
		}
	case TimeZoneNameShortGeneric:
		kind = "short-generic"
	default:
		// TimeZoneNameLongGeneric and the offset forms resolve through the
		// long-generic name and the GMT fallback below.
	}
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

func zoneSpecificName(loc Locale, zone, kind string) string {
	namesOnce.Do(loadNames)
	return timeZoneNameValue(timeZoneNamesByLocale[loc][zone], kind)
}

func metazoneName(loc Locale, metazone, kind string) string {
	namesOnce.Do(loadNames)
	return timeZoneNameValue(metazoneNamesByLocale[loc][metazone], kind)
}

func timeZoneNameValue(data metazoneNames, kind string) string {
	switch kind {
	case "long-standard":
		return data.longStandard
	case "long-daylight":
		return data.longDaylight
	case "short-generic":
		return data.shortGeneric
	case "short-standard":
		return data.shortStandard
	case "short-daylight":
		return data.shortDaylight
	default:
		return data.longGeneric
	}
}

func exemplarCity(loc Locale, zone string) string {
	namesOnce.Do(loadNames)
	return exemplarCitiesByLocale[loc][zone]
}

// timeZoneFormats resolves a locale's GMT/hour offset formats, applying the same
// CLDR root defaults the legacy implementation used when a field is absent.
func timeZoneFormats(loc Locale) timeZoneFormatRefs {
	formatsOnce.Do(loadFormats)
	formats := timeZoneFormatsByLocale[loc]
	if formats.gmtFormat == "" {
		formats.gmtFormat = "GMT{0}"
	}
	if formats.gmtZeroFormat == "" {
		formats.gmtZeroFormat = "GMT"
	}
	if formats.hourFormat == "" {
		formats.hourFormat = "+HH:mm;-HH:mm"
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
		positive, negative = "+HH:mm", "-HH:mm"
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
