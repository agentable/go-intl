// Hand-written decode layer for the displaynames domain. It expands the const
// blobs in data.go into the same per-locale styled-name maps the legacy root
// loadDisplayNamesData produced, behind per-kind sync.Once gates.
//
// Each display-name kind (language, territory, script, calendar, date-time
// field) rides its own blob and its own Once, so querying one kind never decodes
// the others. The language record additionally carries the locale pattern,
// because the language accessor composes language-with-region names from it; that
// composition also reads the territory map, so a language lookup may gate the
// territory decode as well. That is a genuine data dependency between two blobs,
// not a shared gate.

package displaynames

import (
	"sync"

	"github.com/agentable/go-intl/internal/cldr/codec"
)

// styledNames mirrors the legacy root styledNames type: long/short/narrow code
// maps for one display-name kind.
type styledNames struct{ long, short, narrow map[string]string }

// languageDisplay mirrors the legacy root languageDisplay type: the dialect and
// standard styled-name pair for language names.
type languageDisplay struct{ dialect, standard styledNames }

// languageRecord couples one locale's language display names with its locale
// pattern, the two pieces the language accessor needs together.
type languageRecord struct {
	display       languageDisplay
	localePattern string
}

var (
	languageOnce      sync.Once
	languageByLocale  map[string]languageRecord
	territoryOnce     sync.Once
	territoryByLocale map[string]styledNames
	scriptOnce        sync.Once
	scriptByLocale    map[string]styledNames
	calendarOnce      sync.Once
	calendarByLocale  map[string]styledNames
	fieldOnce         sync.Once
	fieldByLocale     map[string]styledNames

	supportedOnce sync.Once
	supportedTags []string
)

func languageData() map[string]languageRecord {
	languageOnce.Do(loadLanguage)
	return languageByLocale
}

func territoryData() map[string]styledNames {
	territoryOnce.Do(loadTerritory)
	return territoryByLocale
}

func scriptData() map[string]styledNames {
	scriptOnce.Do(loadScript)
	return scriptByLocale
}

func calendarData() map[string]styledNames {
	calendarOnce.Do(loadCalendar)
	return calendarByLocale
}

func fieldData() map[string]styledNames {
	fieldOnce.Do(loadField)
	return fieldByLocale
}

func displayNamesSupportedLocales() []string {
	supportedOnce.Do(loadSupported)
	return supportedTags
}

func loadLanguage() {
	r := codec.NewReader(_dnLanguageBlob)
	count := r.Uvarint()
	languageByLocale = make(map[string]languageRecord, count)
	for i := uint64(0); i < count; i++ {
		tag := r.StringRef(_data)
		dialect := decodeStyledNames(&r)
		standard := decodeStyledNames(&r)
		pattern := r.StringRef(_data)
		languageByLocale[tag] = languageRecord{
			display:       languageDisplay{dialect: dialect, standard: standard},
			localePattern: pattern,
		}
	}
}

func loadTerritory() { territoryByLocale = decodeStyledBlob(_dnTerritoryBlob) }
func loadScript()    { scriptByLocale = decodeStyledBlob(_dnScriptBlob) }
func loadCalendar()  { calendarByLocale = decodeStyledBlob(_dnCalendarBlob) }
func loadField()     { fieldByLocale = decodeStyledBlob(_dnDateTimeFieldBlob) }

func decodeStyledBlob(blob string) map[string]styledNames {
	r := codec.NewReader(blob)
	count := r.Uvarint()
	out := make(map[string]styledNames, count)
	for i := uint64(0); i < count; i++ {
		tag := r.StringRef(_data)
		out[tag] = decodeStyledNames(&r)
	}
	return out
}

func decodeStyledNames(r *codec.Reader) styledNames {
	return styledNames{
		long:   decodeStringMap(r),
		short:  decodeStringMap(r),
		narrow: decodeStringMap(r),
	}
}

func decodeStringMap(r *codec.Reader) map[string]string {
	n := r.Uvarint()
	if n == 0 {
		return nil
	}
	out := make(map[string]string, n)
	for i := uint64(0); i < n; i++ {
		key := r.StringRef(_data)
		out[key] = r.StringRef(_data)
	}
	return out
}

func loadSupported() {
	r := codec.NewReader(_dnSupportedBlob)
	count := r.Uvarint()
	supportedTags = make([]string, count)
	for i := range supportedTags {
		supportedTags[i] = r.StringRef(_data)
	}
}
