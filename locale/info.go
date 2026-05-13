package locale

import (
	"slices"
	"strings"
	"time"

	"golang.org/x/text/language"

	"github.com/agentable/go-intl/internal/cldr"
)

type WeekInfo struct {
	FirstDay time.Weekday
	Weekend  []time.Weekday
}

type TextInfo struct {
	Direction string
}

func (l Locale) GetCalendars() []string {
	preference := l.regionPreference()
	calendars := normalizeCalendarList(cldr.CalendarPreference(preference.lookupRegion(cldr.HasCalendarPreference)))
	if l.ext.calendar == "" {
		return calendars
	}
	return []string{l.ext.calendar}
}

func (l Locale) GetCollations() []string {
	if l.ext.collation != "" {
		return []string{l.ext.collation}
	}
	if l.Language() == "und" {
		return []string{"emoji", "eor"}
	}
	return slices.Clone(cldr.SupportedCollations())
}

func (l Locale) GetHourCycles() []string {
	if l.ext.hourCycle != "" {
		return []string{l.ext.hourCycle}
	}
	preference := l.regionPreference()
	return slices.Clone(cldr.HourCyclePreference(preference.lookupRegion(cldr.HasHourCyclePreference)))
}

func (l Locale) GetNumberingSystems() []string {
	if l.ext.numberingSystem != "" {
		return []string{l.ext.numberingSystem}
	}
	if resolved, ok := cldr.ResolveLocale(l.tag); ok {
		if numberingSystem := resolved.DefaultNumberingSystem(); numberingSystem != "" {
			return []string{numberingSystem}
		}
	}
	fallback, _ := cldr.ResolveLocale(language.English)
	return []string{fallback.DefaultNumberingSystem()}
}

func (l Locale) GetTimeZones() []string {
	region := l.Region()
	if region == "" {
		return nil
	}
	return cldr.TimeZonesForRegion(region)
}

func (l Locale) GetWeekInfo() WeekInfo {
	preference := l.regionPreference()
	region := preference.lookupRegion(cldr.HasWeekPreference)
	start, end := cldr.Weekend(region)
	weekend := weekdaysBetween(start, end)
	first := cldr.FirstDayOfWeek(region)
	if l.ext.firstDayOfWeek != "" {
		first = weekdayFromString(l.ext.firstDayOfWeek)
	}
	return WeekInfo{FirstDay: first, Weekend: weekend}
}

func (l Locale) GetTextInfo() TextInfo {
	lang, script, _ := tagParts(l.tag)
	if script == "" {
		_, script, _ = tagParts(l.Maximize().tag)
	}
	if script == "" && lang != "" {
		_, script, _, _ = cldr.MaximizeSubtags(lang, "", "")
	}
	if rtlScripts[script] {
		return TextInfo{Direction: "rtl"}
	}
	return TextInfo{Direction: "ltr"}
}

func (l Locale) region() string {
	_, _, region := tagParts(l.Maximize().tag)
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
		region = l.region()
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
	value := l.ext.keywords[key]
	if len(value) < 4 {
		return ""
	}
	var region string
	switch {
	case len(value) >= 3 && asciiDigits(value[:3]):
		region = value[:3]
	case len(value) >= 2 && asciiLetters(value[:2]):
		region = strings.ToUpper(value[:2])
	default:
		return ""
	}
	for i := len(region); i < len(value); i++ {
		c := value[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			continue
		}
		return ""
	}
	if _, _, region := tagParts(mustLanguageTag("und-" + region)); region != "" {
		return region
	}
	return ""
}

func normalizeCalendarList(in []string) []string {
	out := make([]string, len(in))
	for i, cal := range in {
		out[i], _ = normalizeUnicodeType(cal)
	}
	return out
}

func weekdaysBetween(start, end time.Weekday) []time.Weekday {
	weekend := []time.Weekday{start}
	for day := start; day != end; {
		day = (day + 1) % 7
		weekend = append(weekend, day)
	}
	return weekend
}

func weekdayFromString(s string) time.Weekday {
	switch s {
	case "sun", "7":
		return time.Sunday
	case "mon", "1":
		return time.Monday
	case "tue", "2":
		return time.Tuesday
	case "wed", "3":
		return time.Wednesday
	case "thu", "4":
		return time.Thursday
	case "fri", "5":
		return time.Friday
	case "sat", "6":
		return time.Saturday
	}
	return time.Monday
}

func asciiLetters(s string) bool {
	for i := range len(s) {
		c := s[i]
		if c < 'a' || c > 'z' {
			return false
		}
	}
	return true
}

func asciiDigits(s string) bool {
	for i := range len(s) {
		c := s[i]
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

var rtlScripts = map[string]bool{
	"Adlm": true,
	"Arab": true,
	"Aran": true,
	"Hebr": true,
	"Mand": true,
	"Mani": true,
	"Mend": true,
	"Merc": true,
	"Mero": true,
	"Nkoo": true,
	"Rohg": true,
	"Samr": true,
	"Syrc": true,
	"Thaa": true,
	"Yezi": true,
}
