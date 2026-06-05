// Hand-written accessor layer for the locale kernel. The query semantics mirror
// the legacy literal locales.go, likely_subtags.go, preference.go, and
// numbering.go exactly, so locale negotiation and formatter output are
// byte-for-byte unchanged.

package cldrlocale

import (
	"strings"
	"time"

	"github.com/agentable/go-intl/internal/localeid"
)

// ResolveLocale resolves a tag to its kernel Locale handle, falling back to the
// base language subtag, or (Undefined, false) when neither is known.
func ResolveLocale(tag string) (Locale, bool) {
	localeRegistryOnce.Do(loadLocaleRegistry)
	if loc, ok := localeIndex[tag]; ok {
		return loc, true
	}
	if base, _, ok := strings.Cut(tag, "-"); ok {
		if loc, ok := localeIndex[base]; ok {
			return loc, true
		}
	}
	return Undefined, false
}

// AvailableLocales returns the kernel's available locale tags in sorted order
// (Locale index = position, "und" pinned at 0).
func AvailableLocales() []string {
	localeRegistryOnce.Do(loadLocaleRegistry)
	return availableLocaleTags
}

// Maximize adds likely subtags to tag using the kernel maximize data.
func Maximize(tag string) string {
	return localeid.Maximize(tag, MaximizeSubtags)
}

// MaximizeSubtags returns the likely (language, script, region) for the input
// subtags, or ok=false when the key is absent from the maximize table.
func MaximizeSubtags(language, script, region string) (lang, scr, reg string, ok bool) {
	likelySubtagsOnce.Do(loadLikelySubtags)

	key := language
	if script != "" {
		key += "-" + script
	}
	if region != "" {
		key += "-" + region
	}
	i := searchLikelySubtag(key)
	if i >= len(likelySubtags) || likelySubtags[i].key != key {
		return "", "", "", false
	}
	triple := likelySubtags[i]
	return triple.lang, triple.script, triple.region, true
}

// MinimizeSubtags returns the minimized tag for the input subtags, or ok=false
// when the (language, script, region) triple is absent from the minimize table.
func MinimizeSubtags(language, script, region string) (lang, scr, reg string, ok bool) {
	likelySubtagsOnce.Do(loadLikelySubtags)

	i := searchMinimizeSubtag(language, script, region)
	if i >= len(minimizeSubtags) || compareMinimizeSubtag(minimizeSubtags[i], language, script, region) != 0 {
		return "", "", "", false
	}
	return minimizeSubtags[i].minimized, "", "", true
}

// DefaultNumberingSystem returns the default numbering system for the locale,
// defaulting to "latn" for any locale without a non-latn override.
func (l Locale) DefaultNumberingSystem() string {
	numberingOnce.Do(loadNumbering)
	if ns, ok := numberingByLocale[l]; ok {
		return ns
	}
	return "latn"
}

// SupportedCollations returns the supported collation identifiers.
func SupportedCollations() []string {
	return collationValues
}

// HourCyclePreference returns the region's hour-cycle preference list, falling
// back to the world ("001") default.
func HourCyclePreference(region string) []string {
	preferenceOnce.Do(loadPreferenceData)
	if data, ok := hourCyclePreference[region]; ok {
		return data
	}
	return hourCyclePreference["001"]
}

// HasHourCyclePreference reports whether the region has an explicit hour-cycle
// preference.
func HasHourCyclePreference(region string) bool {
	preferenceOnce.Do(loadPreferenceData)
	_, ok := hourCyclePreference[region]
	return ok
}

// FirstDayOfWeek returns the region's first day of the week.
func FirstDayOfWeek(region string) time.Weekday {
	return weekPreferenceRecord(region).first
}

// Weekend returns the region's weekend start and end days.
func Weekend(region string) (start, end time.Weekday) {
	week := weekPreferenceRecord(region)
	return week.weekendStart, week.weekendEnd
}

// MinimalDaysInFirstWeek returns the region's minimal-days-in-first-week value.
func MinimalDaysInFirstWeek(region string) int {
	return weekPreferenceRecord(region).minDays
}

// CalendarPreference returns the region's calendar preference list, falling back
// to the world ("001") default.
func CalendarPreference(region string) []string {
	preferenceOnce.Do(loadPreferenceData)
	if data, ok := calendarPreference[region]; ok {
		return data
	}
	return calendarPreference["001"]
}

// HasCalendarPreference reports whether the region has an explicit calendar
// preference.
func HasCalendarPreference(region string) bool {
	preferenceOnce.Do(loadPreferenceData)
	_, ok := calendarPreference[region]
	return ok
}

// HasWeekPreference reports whether the region has explicit week data.
func HasWeekPreference(region string) bool {
	preferenceOnce.Do(loadPreferenceData)
	_, ok := weekPreferenceByRegion[region]
	return ok
}

func weekPreferenceRecord(region string) weekPreference {
	preferenceOnce.Do(loadPreferenceData)
	if data, ok := weekPreferenceByRegion[region]; ok {
		return data
	}
	return weekPreferenceByRegion["001"]
}

func searchLikelySubtag(key string) int {
	low, high := 0, len(likelySubtags)
	for low < high {
		mid := int(uint(low+high) >> 1)
		if likelySubtags[mid].key < key {
			low = mid + 1
		} else {
			high = mid
		}
	}
	return low
}

func searchMinimizeSubtag(language, script, region string) int {
	low, high := 0, len(minimizeSubtags)
	for low < high {
		mid := int(uint(low+high) >> 1)
		if compareMinimizeSubtag(minimizeSubtags[mid], language, script, region) < 0 {
			low = mid + 1
		} else {
			high = mid
		}
	}
	return low
}

func compareMinimizeSubtag(row minimizeSubtagRecord, language, script, region string) int {
	if row.lang != language {
		if row.lang < language {
			return -1
		}
		return 1
	}
	if row.script != script {
		if row.script < script {
			return -1
		}
		return 1
	}
	if row.region != region {
		if row.region < region {
			return -1
		}
		return 1
	}
	return 0
}
