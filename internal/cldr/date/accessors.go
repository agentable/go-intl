// Hand-written accessor layer for the date domain. The query semantics mirror
// the legacy root cldr date accessors exactly, so DateTimeFormat output is
// byte-for-byte unchanged.

package date

import (
	"slices"
	"strings"
	"time"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
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
	tags := make([]string, len(supportedTags))
	copy(tags, supportedTags)
	return tags
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
// precedence over range rules, matching the legacy root behavior.
func DayPeriodFor(loc Locale, hour, minute int) string {
	dayPeriodOnce.Do(loadDayPeriods)
	value := time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute
	rules := dayPeriodRules[loc]
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

// GregorianFor returns the resolved gregorian calendar view for the locale,
// assembled from the per-locale calendar data exactly as the legacy root
// accessor did.
func GregorianFor(loc Locale) Gregorian {
	gregorianOnce.Do(loadGregorian)
	data := gregorianData[loc]
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
	gregorian.DayPeriods.AM = dayPeriodNames{Wide: nameAt(wideFormat.dayPeriods, 1), Abbr: nameAt(abbrFormat.dayPeriods, 1), Narrow: nameAt(narrowFormat.dayPeriods, 1)}
	gregorian.DayPeriods.PM = dayPeriodNames{Wide: nameAt(wideFormat.dayPeriods, 3), Abbr: nameAt(abbrFormat.dayPeriods, 3), Narrow: nameAt(narrowFormat.dayPeriods, 3)}
	gregorian.DayPeriods.Flex = flexibleDayPeriodNames(data.names)
	gregorian.DateFormats = styleArray(data.date)
	gregorian.TimeFormats = styleArray(data.time)
	gregorian.DateTimeFormats = styleArray(data.dateTime)
	gregorian.DateTimeAtFormats = styleArray(data.dateTimeAt)
	gregorian.AvailableFormats = data.available
	gregorian.IntervalFormats = data.intervals
	gregorian.IntervalFallback = data.intervalFallback
	gregorian.AppendItems = data.appendItems
	return gregorian
}

// nameAt indexes a day-period name list, returning "" when the index is past the
// list (the legacy literal stored fixed-length slices, but a decoded nil/short
// slice must read as empty rather than panic).
func nameAt(values []string, index int) string {
	if index < len(values) {
		return values[index]
	}
	return ""
}

func styleArray(values map[string]string) [4]string {
	return [4]string{values["full"], values["long"], values["medium"], values["short"]}
}

func dayPeriodName(values []string, key string) string {
	index := dayPeriodNameIndex(key)
	if index < len(values) {
		return values[index]
	}
	return ""
}

func dayPeriodNameIndex(key string) int {
	switch key {
	case "midnight":
		return 0
	case "am":
		return 1
	case "noon":
		return 2
	case "pm":
		return 3
	case "morning1":
		return 4
	case "morning2":
		return 5
	case "afternoon1":
		return 6
	case "afternoon2":
		return 7
	case "evening1":
		return 8
	case "evening2":
		return 9
	case "night1":
		return 10
	case "night2":
		return 11
	default:
		return 100
	}
}

func flexibleDayPeriodNames(names map[calendarNameKey]calendarNames) map[string]dayPeriodNames {
	wide := names[calendarNameKey{width: "wide", context: "format"}].dayPeriods
	abbr := names[calendarNameKey{width: "abbreviated", context: "format"}].dayPeriods
	narrow := names[calendarNameKey{width: "narrow", context: "format"}].dayPeriods
	keys := []string{"midnight", "noon", "morning1", "morning2", "afternoon1", "afternoon2", "evening1", "evening2", "night1", "night2"}
	out := make(map[string]dayPeriodNames)
	for _, key := range keys {
		if name := dayPeriodName(wide, key); name != "" {
			out[key] = dayPeriodNames{Wide: name, Abbr: dayPeriodName(abbr, key), Narrow: dayPeriodName(narrow, key)}
		}
	}
	return out
}

// DateLocaleData exposes the extension keys Intl.DateTimeFormat ResolveLocale
// consults. It owns the calendar key directly and derives the hour-cycle and
// numbering-system keys from the locale kernel, matching the legacy root
// DateLocaleData semantics.
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
	defaultNS := "latn"
	if loc, ok := cldrlocale.ResolveLocale(locale); ok {
		if ns := loc.DefaultNumberingSystem(); ns != "" {
			defaultNS = ns
		}
	}
	return keywordLocaleData(defaultNS, numbering.SimpleNumberingSystems)
}

func hourCycleLocaleData(locale string) []string {
	region := localeRegion(locale)
	if region == "" {
		if localeLanguage(locale) == "en" {
			return []string{"h12", "h23"}
		}
	}
	return keywordLocaleData("", cldrlocale.HourCyclePreference(region))
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
		if len(part) == 2 && isASCIIAlpha(part) || len(part) == 3 && isASCIIDigit(part) {
			if part == "ZZ" {
				return ""
			}
			return part
		}
	}
	return ""
}

func isASCIIAlpha(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' && (r < 'a' || r > 'z') {
			return false
		}
	}
	return s != ""
}

func isASCIIDigit(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

func keywordLocaleData(defaultValue string, values []string) []string {
	out := make([]string, 0, len(values)+1)
	if defaultValue != "" {
		out = append(out, defaultValue)
	}
	for _, value := range values {
		if value != "" && !slices.Contains(out, value) {
			out = append(out, value)
		}
	}
	return out
}
