// Hand-written decode layer for the locale kernel. It expands the const blobs in
// data.go into the same in-memory structures the legacy literal locales.go,
// likely_subtags.go, preference.go, and numbering.go produced, behind per-blob
// sync.Once gates.
//
// The kernel owns the Locale handle type every formatter domain borrows. A
// domain's packed keys depend on the locale index this package assigns, so the
// registry decode here is authoritative: Locale(i) is the i-th tag in the sorted
// _localeBlob, "und" pinned at 0, exactly as the legacy literal table assigned.

package cldrlocale

import (
	"sync"
	"time"

	"github.com/agentable/go-intl/internal/cldr/codec"
)

// Locale is an opaque handle into the generated CLDR locale data.
type Locale uint16

// Undefined is the "und" sentinel locale.
const Undefined Locale = 0

type maximizeSubtagRecord struct{ key, lang, script, region string }

type minimizeSubtagRecord struct{ lang, script, region, minimized string }

type weekPreference struct {
	first, weekendStart, weekendEnd time.Weekday
	minDays                         int
}

var (
	localeRegistryOnce  sync.Once
	availableLocaleTags []string
	localeIndex         map[string]Locale

	likelySubtagsOnce sync.Once
	likelySubtags     []maximizeSubtagRecord
	minimizeSubtags   []minimizeSubtagRecord

	numberingOnce     sync.Once
	numberingByLocale map[Locale]string

	preferenceOnce         sync.Once
	hourCyclePreference    map[string][]string
	weekPreferenceByRegion map[string]weekPreference
	calendarPreference     map[string][]string
)

func loadLocaleRegistry() {
	r := codec.NewReader(_localeBlob)
	count := r.Uvarint()
	availableLocaleTags = make([]string, count)
	localeIndex = make(map[string]Locale, count)
	for i := range availableLocaleTags {
		tag := r.StringRef(_data)
		availableLocaleTags[i] = tag
		localeIndex[tag] = Locale(i)
	}
}

func loadLikelySubtags() {
	max := codec.NewReader(_maximizeBlob)
	maxCount := max.Uvarint()
	likelySubtags = make([]maximizeSubtagRecord, maxCount)
	for i := range likelySubtags {
		likelySubtags[i] = maximizeSubtagRecord{
			key:    max.StringRef(_data),
			lang:   max.StringRef(_data),
			script: max.StringRef(_data),
			region: max.StringRef(_data),
		}
	}

	min := codec.NewReader(_minimizeBlob)
	minCount := min.Uvarint()
	minimizeSubtags = make([]minimizeSubtagRecord, minCount)
	for i := range minimizeSubtags {
		minimizeSubtags[i] = minimizeSubtagRecord{
			lang:      min.StringRef(_data),
			script:    min.StringRef(_data),
			region:    min.StringRef(_data),
			minimized: min.StringRef(_data),
		}
	}
}

func loadNumbering() {
	r := codec.NewReader(_numberingBlob)
	count := r.Uvarint()
	numberingByLocale = make(map[Locale]string, count)
	locales := codec.Delta(&r)
	for range count {
		loc := Locale(locales.Next32())
		numberingByLocale[loc] = r.StringRef(_data)
	}
}

func loadPreferenceData() {
	hour := codec.NewReader(_hourCycleBlob)
	hourCount := hour.Uvarint()
	hourCyclePreference = make(map[string][]string, hourCount)
	for range hourCount {
		region := hour.StringRef(_data)
		values := make([]string, hour.Uvarint())
		for i := range values {
			values[i] = hour.StringRef(_data)
		}
		hourCyclePreference[region] = values
	}

	week := codec.NewReader(_weekBlob)
	weekCount := week.Uvarint()
	weekPreferenceByRegion = make(map[string]weekPreference, weekCount)
	for range weekCount {
		region := week.StringRef(_data)
		first := time.Weekday(week.Uvarint())
		weekendStart := time.Weekday(week.Uvarint())
		weekendEnd := time.Weekday(week.Uvarint())
		minDays := int(week.Uvarint())
		weekPreferenceByRegion[region] = weekPreference{
			first:        first,
			weekendStart: weekendStart,
			weekendEnd:   weekendEnd,
			minDays:      minDays,
		}
	}

	calendar := codec.NewReader(_calendarBlob)
	calCount := calendar.Uvarint()
	calendarPreference = make(map[string][]string, calCount)
	for range calCount {
		region := calendar.StringRef(_data)
		values := make([]string, calendar.Uvarint())
		for i := range values {
			values[i] = calendar.StringRef(_data)
		}
		calendarPreference[region] = values
	}
}
