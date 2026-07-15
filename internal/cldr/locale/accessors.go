// Hand-written accessor layer for the locale kernel. It owns locale handles,
// likely subtags, preference data, and numbering defaults shared by CLDR-backed
// formatter domains.

package cldrlocale

import (
	"cmp"
	"slices"
	"strings"
	"time"

	"github.com/agentable/go-intl/internal/localeid"
)

const worldRegion = "001"

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
	return slices.Clone(availableLocaleTags)
}

// IntersectSupportedLocales returns primary locales supported by every required
// locale list. It preserves primary order and returns a new slice.
func IntersectSupportedLocales(primary []string, required ...[]string) []string {
	if len(required) == 0 {
		return slices.Clone(primary)
	}
	sets := make([]map[string]bool, len(required))
	for i, locales := range required {
		set := make(map[string]bool, len(locales))
		for _, loc := range locales {
			set[loc] = true
		}
		sets[i] = set
	}
	out := make([]string, 0, len(primary))
	for _, loc := range primary {
		if supportedByAll(loc, sets) {
			out = append(out, loc)
		}
	}
	return out
}

// Maximize adds likely subtags to tag using the kernel maximize data.
func Maximize(tag string) string {
	return localeid.Maximize(tag, MaximizeSubtags)
}

// MaximizeSubtags applies the CLDR Add Likely Subtags fallback order and
// preserves subtags already supplied by the caller.
func MaximizeSubtags(language, script, region string) (lang, scr, reg string, ok bool) {
	likelySubtagsOnce.Do(loadLikelySubtags)
	candidates := [...]string{
		localeid.Join(language, script, region),
		localeid.Join(language, "", region),
		localeid.Join(language, script, ""),
		language,
		localeid.Join("und", script, ""),
		localeid.Join("und", "", region),
		"und",
	}
	for i, key := range candidates {
		if key == "" || slices.Contains(candidates[:i], key) {
			continue
		}
		maxLang, maxScript, maxRegion, found := exactLikelySubtags(key)
		if !found {
			continue
		}
		if language != "" && language != "und" {
			maxLang = language
		}
		if script != "" {
			maxScript = script
		}
		if region != "" {
			maxRegion = region
		}
		return maxLang, maxScript, maxRegion, true
	}
	return "", "", "", false
}

func exactLikelySubtags(key string) (lang, scr, reg string, ok bool) {
	i, ok := slices.BinarySearchFunc(likelySubtags, key, func(row maximizeSubtagRecord, target string) int {
		return cmp.Compare(row.key, target)
	})
	if ok {
		triple := likelySubtags[i]
		return triple.lang, triple.script, triple.region, true
	}
	return "", "", "", false
}

// MinimizeSubtags returns the minimized tag for the input subtags, or ok=false
// when the (language, script, region) triple is absent from the minimize table.
func MinimizeSubtags(language, script, region string) (lang, scr, reg string, ok bool) {
	likelySubtagsOnce.Do(loadLikelySubtags)

	key := minimizeSubtagKey{language: language, script: script, region: region}
	i, ok := slices.BinarySearchFunc(minimizeSubtags, key, compareMinimizeSubtag)
	if ok {
		return minimizeSubtags[i].minimized, "", "", true
	}
	return "", "", "", false
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

// HourCyclePreference returns the region's hour-cycle preference list, falling
// back to the world ("001") default.
func HourCyclePreference(region string) []string {
	preferenceOnce.Do(loadPreferenceData)
	if data, ok := hourCyclePreference[region]; ok {
		return slices.Clone(data)
	}
	return slices.Clone(hourCyclePreference[worldRegion])
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
		return slices.Clone(data)
	}
	return slices.Clone(calendarPreference[worldRegion])
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
	return weekPreferenceByRegion[worldRegion]
}

func supportedByAll(loc string, sets []map[string]bool) bool {
	for _, set := range sets {
		if !set[loc] {
			return false
		}
	}
	return true
}

type minimizeSubtagKey struct {
	language string
	script   string
	region   string
}

func compareMinimizeSubtag(row minimizeSubtagRecord, key minimizeSubtagKey) int {
	if diff := cmp.Compare(row.lang, key.language); diff != 0 {
		return diff
	}
	if diff := cmp.Compare(row.script, key.script); diff != 0 {
		return diff
	}
	return cmp.Compare(row.region, key.region)
}
