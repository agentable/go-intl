package datetimeformat

import (
	"time"

	cldrtimezone "github.com/agentable/go-intl/internal/cldr/timezone"
	"github.com/agentable/go-intl/internal/tz"
)

func (f *DateTimeFormat) timeZonePatternName(width int, t time.Time) string {
	form := cldrtimezone.TimeZoneNameShort
	if width >= 4 {
		form = cldrtimezone.TimeZoneNameLong
	}
	return f.localizedTimeZonePatternName(form, width, t)
}

func (f *DateTimeFormat) genericTimeZonePatternName(width int, t time.Time) string {
	form := cldrtimezone.TimeZoneNameShortGeneric
	if width >= 4 {
		form = cldrtimezone.TimeZoneNameLongGeneric
	}
	return f.localizedTimeZonePatternName(form, width, t)
}

func (f *DateTimeFormat) offsetTimeZonePatternName(_ int, t time.Time) string {
	_, info := f.timeZoneInfo(t)
	form := cldrtimezone.TimeZoneNameShortOffset
	if f.resolved.TimeZoneName == LongOffsetTimeZoneName {
		form = cldrtimezone.TimeZoneNameLongOffset
	}
	return cldrtimezone.GMTOffsetName(f.cldrLoc, info.OffsetMs, form)
}

func (f *DateTimeFormat) localizedTimeZonePatternName(form cldrtimezone.TimeZoneName, width int, t time.Time) string {
	zone, info := f.timeZoneInfo(t)
	if zone != "" && zone != "Local" {
		if name := cldrtimezone.TimeZoneDisplayName(f.cldrLoc, zone, form, info.IsDST, t.UnixMilli(), info.OffsetMs); name != "" {
			return name
		}
	}
	if form == cldrtimezone.TimeZoneNameShort && width < 4 && info.Abbrv != "" {
		return info.Abbrv
	}
	return cldrtimezone.GMTOffsetName(f.cldrLoc, info.OffsetMs, form)
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
