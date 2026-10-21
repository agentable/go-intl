package locale

import (
	"encoding/json/v2"
	"slices"
	"time"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/localeid"
	"github.com/agentable/go-intl/internal/tz"
)

// WeekInfo describes locale week data. Mirrors Intl.Locale.prototype.getWeekInfo().
type WeekInfo struct {
	// FirstDay is the locale's first day of week. Mirrors getWeekInfo() field "firstDay".
	FirstDay time.Weekday `json:"firstDay"`
	// Weekend is the locale's weekend day range. Mirrors getWeekInfo() field "weekend".
	Weekend []time.Weekday `json:"weekend"`
}

// TextInfo describes locale text direction. Mirrors Intl.Locale.prototype.getTextInfo().
type TextInfo struct {
	// Direction is the locale text direction, or nil when it is unknown.
	// Mirrors getTextInfo() field "direction".
	Direction *string `json:"direction,omitempty"`
}

// MarshalJSON emits ECMA-402 weekday numbers, Monday=1 through Sunday=7.
func (w WeekInfo) MarshalJSON() ([]byte, error) {
	type weekInfo struct {
		FirstDay int   `json:"firstDay"`
		Weekend  []int `json:"weekend"`
	}
	weekend := make([]int, len(w.Weekend))
	for i, day := range w.Weekend {
		weekend[i] = weekdayNumber(day)
	}
	return json.Marshal(weekInfo{
		FirstDay: weekdayNumber(w.FirstDay),
		Weekend:  weekend,
	})
}

func (l Locale) GetCalendars() []string {
	if value, present := l.ext.keywords["ca"]; present {
		return []string{value}
	}
	preference := l.regionPreference()
	calendars := cldrlocale.CalendarPreference(preference.lookupRegion(cldrlocale.HasCalendarPreference))
	supported := cldrdate.SupportedCalendars()
	out := make([]string, 0, len(calendars))
	for _, calendar := range calendars {
		if slices.Contains(supported, calendar) && !slices.Contains(out, calendar) {
			out = append(out, calendar)
		}
	}
	if len(out) == 0 {
		return []string{"gregory"}
	}
	return out
}

func (l Locale) GetCollations() []string {
	if value, present := l.ext.keywords["co"]; present {
		return []string{value}
	}
	return []string{"emoji", "eor"}
}

func (l Locale) GetHourCycles() []string {
	if value, present := l.ext.keywords["hc"]; present {
		return []string{value}
	}
	preference := l.regionPreference()
	return cldrlocale.HourCyclePreference(l.Language(), preference.regionOverride, preference.region)
}

func (l Locale) GetNumberingSystems() []string {
	if value, present := l.ext.keywords["nu"]; present {
		return []string{value}
	}
	if resolved, ok := cldrlocale.ResolveLocale(l.tag.String()); ok {
		if numberingSystem := resolved.DefaultNumberingSystem(); numberingSystem != "" {
			return []string{numberingSystem}
		}
	}
	fallback, _ := cldrlocale.ResolveLocale("en")
	return []string{fallback.DefaultNumberingSystem()}
}

func (l Locale) GetTimeZones() []string {
	region := l.Region()
	if region == "" {
		return nil
	}
	zones := tz.TimeZonesForRegion(region)
	if zones == nil {
		return []string{}
	}
	return zones
}

func (l Locale) GetWeekInfo() WeekInfo {
	preference := l.regionPreference()
	region := preference.lookupRegion(cldrlocale.HasWeekPreference)
	start, end := cldrlocale.Weekend(region)
	weekend := weekdaysBetween(start, end)
	first := cldrlocale.FirstDayOfWeek(region)
	if day, ok := weekdayFromString(l.ext.keywords["fw"]); ok {
		first = day
	}
	return WeekInfo{FirstDay: first, Weekend: weekend}
}

func (l Locale) GetTextInfo() TextInfo {
	_, script, _ := localeid.Parts(l.tag)
	if script == "" {
		_, script, _ = localeid.Parts(l.Maximize().tag)
	}
	if direction, ok := cldrlocale.TextDirection(script); ok {
		return TextInfo{Direction: &direction}
	}
	return TextInfo{}
}

func (l Locale) maximizedRegion() string {
	_, _, region := localeid.Parts(l.Maximize().tag)
	return region
}

type regionPreference struct {
	region         string
	regionOverride string
}

func (l Locale) regionPreference() regionPreference {
	region := l.Region()
	if region == "" {
		region = l.canonicalUnicodeSubdivision("sd")
	}
	if region == "" {
		region = l.maximizedRegion()
	}
	if region == "" {
		region = "001"
	}
	return regionPreference{region: region, regionOverride: l.canonicalUnicodeSubdivision("rg")}
}

func (p regionPreference) lookupRegion(hasData func(string) bool) string {
	if p.regionOverride != "" && hasData(p.regionOverride) {
		return p.regionOverride
	}
	return p.region
}

func (l Locale) canonicalUnicodeSubdivision(key string) string {
	region, ok := localeid.UnicodeSubdivisionRegion(l.ext.keywords[key])
	if !ok {
		return ""
	}
	return localeid.CanonicalRegion("und", "", region, cldrlocale.MaximizeSubtags)
}

func weekdaysBetween(start, end time.Weekday) []time.Weekday {
	weekend := []time.Weekday{start}
	for day := start; day != end; {
		day = (day + 1) % 7
		weekend = append(weekend, day)
	}
	return weekend
}

func weekdayFromString(s string) (time.Weekday, bool) {
	switch s {
	case "sun", "7":
		return time.Sunday, true
	case "mon", "1":
		return time.Monday, true
	case "tue", "2":
		return time.Tuesday, true
	case "wed", "3":
		return time.Wednesday, true
	case "thu", "4":
		return time.Thursday, true
	case "fri", "5":
		return time.Friday, true
	case "sat", "6":
		return time.Saturday, true
	}
	return 0, false
}

func weekdayNumber(day time.Weekday) int {
	if day == time.Sunday {
		return 7
	}
	return int(day)
}
