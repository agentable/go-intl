// Hand-written decode layer for the displaynames domain. It expands
// domain-private const blobs from data.go into per-locale styled-name maps,
// behind per-kind sync.Once gates.
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

// styledNames holds long/short/narrow code maps for one display-name kind.
type styledNames struct{ long, short, narrow map[string]string }

// languageDisplay holds the dialect and standard styled-name pair for language
// names.
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

func loadLanguage() {
	r := codec.NewReader(_dnLanguageBlob)
	languageByLocale = codec.StringRefKeyMap[languageRecord](&r, _data, decodeLanguageRecord)
}

func decodeLanguageRecord(r *codec.Reader) languageRecord {
	dialect := decodeStyledNames(r)
	standard := decodeStyledNames(r)
	pattern := r.StringRef(_data)
	return languageRecord{
		display:       languageDisplay{dialect: dialect, standard: standard},
		localePattern: pattern,
	}
}

func loadTerritory() { territoryByLocale = decodeStyledBlob(_dnTerritoryBlob) }
func loadScript()    { scriptByLocale = decodeStyledBlob(_dnScriptBlob) }
func loadCalendar()  { calendarByLocale = decodeStyledBlob(_dnCalendarBlob) }
func loadField()     { fieldByLocale = decodeStyledBlob(_dnDateTimeFieldBlob) }

func decodeStyledBlob(blob string) map[string]styledNames {
	r := codec.NewReader(blob)
	return codec.StringRefKeyMap[styledNames](&r, _data, decodeStyledNames)
}

func decodeStyledNames(r *codec.Reader) styledNames {
	return styledNames{
		long:   r.StringRefMap(_data),
		short:  r.StringRefMap(_data),
		narrow: r.StringRefMap(_data),
	}
}

func loadSupported() {
	supportedTags = codec.StringRefSlice(_dnSupportedBlob, _data)
}
