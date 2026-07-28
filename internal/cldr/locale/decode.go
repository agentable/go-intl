// Hand-written decode layer for the locale kernel. It expands const blobs from
// data.go into the locale registry, likely-subtag, preference, and numbering
// records consumed by accessors.go, behind per-blob sync.Once gates.
//
// The kernel owns the Locale handle type every formatter domain borrows. A
// domain's packed keys depend on the locale index this package assigns, so the
// registry decode here is authoritative: Locale(i) is the i-th tag in the sorted
// _localeBlob, with "und" pinned at 0.

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

type scriptDirectionRecord struct {
	script string
	rtl    bool
}

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

	scriptDirectionOnce sync.Once
	scriptDirections    []scriptDirectionRecord

	numberingOnce     sync.Once
	numberingByLocale map[Locale]string

	preferenceOnce         sync.Once
	hourCyclePreference    map[string][]string
	weekPreferenceByRegion map[string]weekPreference
	calendarPreference     map[string][]string
)

func loadLocaleRegistry() {
	availableLocaleTags = codec.StringRefSlice(_localeBlob, _data)
	localeIndex = make(map[string]Locale, len(availableLocaleTags))
	for i, tag := range availableLocaleTags {
		localeIndex[tag] = Locale(i)
	}
}

func loadLikelySubtags() {
	max := codec.NewReader(_maximizeBlob)
	likelySubtags = codec.CountedSlice[maximizeSubtagRecord](&max, decodeMaximizeSubtagRecord)

	min := codec.NewReader(_minimizeBlob)
	minimizeSubtags = codec.CountedSlice[minimizeSubtagRecord](&min, decodeMinimizeSubtagRecord)
}

func decodeMaximizeSubtagRecord(r *codec.Reader) maximizeSubtagRecord {
	return maximizeSubtagRecord{
		key:    r.StringRef(_data),
		lang:   r.StringRef(_data),
		script: r.StringRef(_data),
		region: r.StringRef(_data),
	}
}

func decodeMinimizeSubtagRecord(r *codec.Reader) minimizeSubtagRecord {
	return minimizeSubtagRecord{
		lang:      r.StringRef(_data),
		script:    r.StringRef(_data),
		region:    r.StringRef(_data),
		minimized: r.StringRef(_data),
	}
}

func loadScriptDirections() {
	r := codec.NewReader(_directionBlob)
	scriptDirections = codec.CountedSlice[scriptDirectionRecord](&r, decodeScriptDirectionRecord)
}

func decodeScriptDirectionRecord(r *codec.Reader) scriptDirectionRecord {
	return scriptDirectionRecord{script: r.StringRef(_data), rtl: r.Uvarint() != 0}
}

func loadNumbering() {
	r := codec.NewReader(_numberingBlob)
	numberingByLocale = codec.Uint16DeltaMap[Locale, string](&r, decodeDefaultNumberingSystem)
}

func decodeDefaultNumberingSystem(r *codec.Reader) string { return r.StringRef(_data) }

func decodePreferenceList(r *codec.Reader) []string { return r.StringRefSlice(_data) }

func loadPreferenceData() {
	hour := codec.NewReader(_hourCycleBlob)
	hourCyclePreference = codec.StringRefKeyMap[[]string](&hour, _data, decodePreferenceList)

	week := codec.NewReader(_weekBlob)
	weekPreferenceByRegion = codec.StringRefKeyMap[weekPreference](&week, _data, decodeWeekPreference)

	calendar := codec.NewReader(_calendarBlob)
	calendarPreference = codec.StringRefKeyMap[[]string](&calendar, _data, decodePreferenceList)
}

func decodeWeekPreference(r *codec.Reader) weekPreference {
	return weekPreference{
		first:        time.Weekday(r.Uvarint()),
		weekendStart: time.Weekday(r.Uvarint()),
		weekendEnd:   time.Weekday(r.Uvarint()),
		minDays:      int(r.Uvarint()),
	}
}
