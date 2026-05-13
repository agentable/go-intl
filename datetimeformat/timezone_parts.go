package datetimeformat

import (
	"time"

	"github.com/agentable/go-intl/internal/cldr"
	"github.com/agentable/go-intl/internal/tz"
)

func (f *DateTimeFormat) timeZonePatternName(width int, t time.Time) string {
	form := cldr.TimeZoneNameShort
	if width >= 4 {
		form = cldr.TimeZoneNameLong
	}
	return f.localizedTimeZonePatternName(form, width, t)
}

func (f *DateTimeFormat) genericTimeZonePatternName(width int, t time.Time) string {
	form := cldr.TimeZoneNameShortGeneric
	if width >= 4 {
		form = cldr.TimeZoneNameLongGeneric
	}
	return f.localizedTimeZonePatternName(form, width, t)
}

func (f *DateTimeFormat) offsetTimeZonePatternName(_ int, t time.Time) string {
	_, info := f.timeZoneInfo(t)
	form := cldr.TimeZoneNameShortOffset
	if f.resolved.TimeZoneName == LongOffsetTimeZoneName {
		form = cldr.TimeZoneNameLongOffset
	}
	return cldr.GMTOffsetName(f.cldrLoc, info.OffsetMs, form)
}

func (f *DateTimeFormat) localizedTimeZonePatternName(form cldr.TimeZoneName, width int, t time.Time) string {
	zone, info := f.timeZoneInfo(t)
	if zone != "" && zone != "Local" {
		if name := cldr.TimeZoneDisplayName(f.cldrLoc, zone, form, info.IsDST, t.UnixMilli(), info.OffsetMs); name != "" {
			return name
		}
	}
	if form == cldr.TimeZoneNameShort && width < 4 && info.Abbrv != "" {
		return info.Abbrv
	}
	return cldr.GMTOffsetName(f.cldrLoc, info.OffsetMs, form)
}

func (f *DateTimeFormat) timeZoneInfo(t time.Time) (string, tz.ZoneInfo) {
	zone := f.resolved.TimeZone
	if f.location != nil {
		return zone, tz.LookupAt(f.location, t)
	}
	if zone == "" {
		zone = "UTC"
	}
	return zone, tz.LookupAt(time.UTC, t)
}
