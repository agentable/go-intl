// Hand-written accessor layer for the date domain. It exposes gregorian
// calendar data, day-period rules, locale data, and narrow supported indexes over
// lazily decoded const blobs.

package date

import (
	"maps"
	"slices"
	"strings"
	"time"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/localeid"
	"github.com/agentable/go-intl/internal/numbering"
)

// ResolveLocale resolves a tag to its kernel locale handle, forwarding to the
// shared locale kernel so date handles index identically to every other domain.
func ResolveLocale(tag string) (Locale, bool) {
	return cldrlocale.ResolveLocale(tag)
}

// SupportedLocales returns the date-supported locale tags in sorted-locale
// order. It reads only the narrow supported blob and never triggers the
// gregorian or day-period blob decode.
func SupportedLocales() []string {
	supportedOnce.Do(loadSupported)
	return slices.Clone(supportedTags)
}

// SupportedCalendars returns the canonical ECMA-402 calendar identifiers backed
// by the generated date payload (gregory plus the iso8601 bridge). It reads only
// the narrow calendar blob and never triggers the gregorian blob decode.
func SupportedCalendars() []string {
	calendarsOnce.Do(loadCalendars)
	return slices.Clone(calendarIDs)
}

// DayPeriodFor returns the day-period type that contains the given wall-clock
// time for the locale, or "" when no rule matches. Point rules (From == To) take
// precedence over range rules.
func DayPeriodFor(loc Locale, hour, minute int) string {
	dayPeriodOnce.Do(loadDayPeriods)
	value := time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute
	rules := dayPeriodRulesForLocale(loc)
	for _, rule := range rules {
		if rule.From == rule.To && value == rule.From {
			return rule.Type
		}
	}
	for _, rule := range rules {
		if rule.From == rule.To {
			continue
		}
		if rule.From < rule.To {
			if value >= rule.From && value < rule.To {
				return rule.Type
			}
			continue
		}
		if value >= rule.From || value < rule.To {
			return rule.Type
		}
	}
	return ""
}

func dayPeriodRulesForLocale(loc Locale) []DayPeriodRange {
	rules, ok := dayPeriodRules[loc]
	if !ok {
		return nil
	}
	return rules
}

// GregorianFor returns the resolved gregorian calendar view for the locale,
// assembled from the per-locale calendar data.
func GregorianFor(loc Locale) Gregorian {
	gregorianOnce.Do(loadGregorian)
	data := gregorianCalendarData(loc)
	var gregorian Gregorian
	wideFormat := data.names[calendarNameKey{width: "wide", context: "format"}]
	abbrFormat := data.names[calendarNameKey{width: "abbreviated", context: "format"}]
	narrowFormat := data.names[calendarNameKey{width: "narrow", context: "format"}]
	wideStandalone := data.names[calendarNameKey{width: "wide", context: "stand-alone"}]
	abbrStandalone := data.names[calendarNameKey{width: "abbreviated", context: "stand-alone"}]
	narrowStandalone := data.names[calendarNameKey{width: "narrow", context: "stand-alone"}]
	copy(gregorian.Eras.Wide[:], wideFormat.eras)
	copy(gregorian.Eras.Abbr[:], abbrFormat.eras)
	copy(gregorian.Eras.Narrow[:], narrowFormat.eras)
	copy(gregorian.Months.Wide[:], wideFormat.months)
	copy(gregorian.Months.Abbr[:], abbrFormat.months)
	copy(gregorian.Months.Narrow[:], narrowFormat.months)
	copy(gregorian.Months.StandWide[:], wideStandalone.months)
	copy(gregorian.Months.StandAbbr[:], abbrStandalone.months)
	copy(gregorian.Months.StandNarrow[:], narrowStandalone.months)
	copy(gregorian.Weekdays.Wide[:], wideFormat.weekdays)
	copy(gregorian.Weekdays.Abbr[:], abbrFormat.weekdays)
	copy(gregorian.Weekdays.Narrow[:], narrowFormat.weekdays)
	copy(gregorian.Weekdays.StandWide[:], wideStandalone.weekdays)
	copy(gregorian.Weekdays.StandAbbr[:], abbrStandalone.weekdays)
	copy(gregorian.Weekdays.StandNarrow[:], narrowStandalone.weekdays)
	gregorian.DayPeriods.AM = dayPeriodNamesAt(wideFormat.dayPeriods, abbrFormat.dayPeriods, narrowFormat.dayPeriods, dayPeriodAMSlot)
	gregorian.DayPeriods.PM = dayPeriodNamesAt(wideFormat.dayPeriods, abbrFormat.dayPeriods, narrowFormat.dayPeriods, dayPeriodPMSlot)
	gregorian.DayPeriods.Flex = flexibleDayPeriodNames(wideFormat.dayPeriods, abbrFormat.dayPeriods, narrowFormat.dayPeriods)
	gregorian.DateFormats = styleArray(data.date)
	gregorian.TimeFormats = styleArray(data.time)
	gregorian.DateTimeFormats = styleArray(data.dateTime)
	gregorian.DateTimeAtFormats = styleArray(data.dateTimeAt)
	gregorian.AvailableFormats = maps.Clone(data.available)
	gregorian.IntervalFormats = cloneIntervalFormats(data.intervals)
	gregorian.IntervalFallback = data.intervalFallback
	gregorian.AppendItems = maps.Clone(data.appendItems)
	return gregorian
}

func gregorianCalendarData(loc Locale) calendarData {
	data, ok := gregorianData[loc]
	if !ok {
		return calendarData{}
	}
	return data
}

// nameAt indexes a day-period name list, returning "" when the index is past the
// list. A decoded nil or short slice must read as empty rather than panic.
func nameAt(values []string, index int) string {
	if index < len(values) {
		return values[index]
	}
	return ""
}

func dayPeriodNamesAt(wide, abbr, narrow []string, index int) dayPeriodNames {
	return dayPeriodNames{
		Wide:   nameAt(wide, index),
		Abbr:   nameAt(abbr, index),
		Narrow: nameAt(narrow, index),
	}
}

var dateStyleSlotOrder = [...]string{"full", "long", "medium", "short"}

func styleArray(values map[string]string) [4]string {
	var styles [4]string
	for i, key := range dateStyleSlotOrder {
		styles[i] = values[key]
	}
	return styles
}

const (
	dayPeriodAMSlot = 1
	dayPeriodPMSlot = 3
)

type flexibleDayPeriodSlot struct {
	key   string
	index int
}

var flexibleDayPeriodSlots = [...]flexibleDayPeriodSlot{
	{key: "midnight", index: 0},
	{key: "noon", index: 2},
	{key: "morning1", index: 4},
	{key: "morning2", index: 5},
	{key: "afternoon1", index: 6},
	{key: "afternoon2", index: 7},
	{key: "evening1", index: 8},
	{key: "evening2", index: 9},
	{key: "night1", index: 10},
	{key: "night2", index: 11},
}

func flexibleDayPeriodNames(wide, abbr, narrow []string) map[string]dayPeriodNames {
	out := make(map[string]dayPeriodNames, len(flexibleDayPeriodSlots))
	for _, period := range flexibleDayPeriodSlots {
		names := dayPeriodNamesAt(wide, abbr, narrow, period.index)
		if names.Wide != "" {
			out[period.key] = names
		}
	}
	return out
}

func cloneIntervalFormats(formats map[string]map[string]string) map[string]map[string]string {
	if formats == nil {
		return nil
	}
	out := make(map[string]map[string]string, len(formats))
	for skeleton, fields := range formats {
		out[skeleton] = maps.Clone(fields)
	}
	return out
}

// DateLocaleData exposes the extension keys Intl.DateTimeFormat ResolveLocale
// consults. It owns the calendar key directly and derives the hour-cycle and
// numbering-system keys from the locale kernel.
type DateLocaleData struct{}

func (DateLocaleData) For(locale, key string) []string {
	switch key {
	case "ca":
		return SupportedCalendars()
	case "hc":
		return hourCycleLocaleData(locale)
	case "nu":
		return numberingSystemLocaleData(locale)
	default:
		return nil
	}
}

func numberingSystemLocaleData(locale string) []string {
	defaultNumberingSystem := numbering.DefaultNumberingSystem
	if loc, ok := cldrlocale.ResolveLocale(locale); ok {
		if ns := loc.DefaultNumberingSystem(); ns != "" {
			defaultNumberingSystem = ns
		}
	}
	return localeid.RelevantExtensionValues(defaultNumberingSystem, numbering.SimpleNumberingSystems()...)
}

func hourCycleLocaleData(locale string) []string {
	region := localeRegion(locale)
	if region == "" && localeLanguage(locale) == "en" {
		return localeid.RelevantExtensionValues("", "h12", "h23")
	}
	return localeid.RelevantExtensionValues("", cldrlocale.HourCyclePreference(region)...)
}

func localeLanguage(locale string) string {
	language, _, _ := strings.Cut(locale, "-")
	return language
}

func localeRegion(locale string) string {
	first := true
	for part := range strings.SplitSeq(locale, "-") {
		if first {
			first = false
			continue
		}
		if len(part) == 1 {
			return ""
		}
		if localeid.IsUnicodeRegionSubtag(part) {
			if part == "ZZ" {
				return ""
			}
			return part
		}
	}
	return ""
}
