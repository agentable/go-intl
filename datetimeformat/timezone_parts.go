package datetimeformat

import (
	"time"

	cldrtimezone "github.com/agentable/go-intl/internal/cldr/timezone"
	"github.com/agentable/go-intl/internal/ecma402"
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

func (f *DateTimeFormat) offsetTimeZonePatternName(t time.Time) string {
	_, info := resolvedTimeZoneInfo(f.resolved.TimeZone, f.location, t)
	form := cldrtimezone.TimeZoneNameShortOffset
	if ecma402.ResolvedScalarValue(f.resolved.TimeZoneName) == LongOffsetTimeZoneName {
		form = cldrtimezone.TimeZoneNameLongOffset
	}
	return cldrtimezone.GMTOffsetName(cldrtimezone.Locale(f.cldrLoc), info.OffsetMs, form)
}

func (f *DateTimeFormat) localizedTimeZonePatternName(form cldrtimezone.TimeZoneName, width int, t time.Time) string {
	loc := cldrtimezone.Locale(f.cldrLoc)
	zone, info := resolvedTimeZoneInfo(f.resolved.TimeZone, f.location, t)
	if zone != "" && zone != "Local" {
		if name := cldrtimezone.TimeZoneDisplayName(loc, zone, form, info.IsDST, t.UnixMilli(), info.OffsetMs); name != "" {
			return name
		}
	}
	if form == cldrtimezone.TimeZoneNameShort && width < 4 && info.Abbrv != "" {
		return info.Abbrv
	}
	return cldrtimezone.GMTOffsetName(loc, info.OffsetMs, form)
}

func resolvedTimeZoneInfo(zone string, location *time.Location, t time.Time) (string, tz.ZoneInfo) {
	if location != nil {
		return zone, tz.LookupAt(location, t)
	}
	if zone == "" {
		zone = "UTC"
	}
	return zone, tz.LookupAt(time.UTC, t)
}
