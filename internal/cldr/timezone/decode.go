// Hand-written decode layer for the timezone domain. It expands domain-private
// const blobs from data.go into metazone period, localized display-name, and GMT
// offset-format records behind per-blob sync.Once gates, so granularity follows
// accessor reachability.
//
// Locale handle ownership: the names and formats blobs pack the locale index
// assigned by the cldr/locale kernel. Borrowing that handle keeps generated
// timezone display data and formatter locale resolution on one stable index space
// while the dependency stays one-way (timezone -> cldr/locale).

package timezone

import (
	"sync"

	"github.com/agentable/go-intl/internal/cldr/codec"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// Locale is the borrowed locale handle (see file header).
type Locale = cldrlocale.Locale

// metazoneNames carries the six width/usage variants of one metazone or zone
// label. The blob StringRefs are resolved against _data at decode time, so each
// field already holds the final string the accessor returns.
type metazoneNames struct {
	longGeneric, longStandard, longDaylight    string
	shortGeneric, shortStandard, shortDaylight string
}

// metazonePeriod is one resolved [start, end) window tagging a zone with the
// metazone in force during it. start/end are unix-milli instants whose open ends
// are the int64 min/max sentinels.
type metazonePeriod struct {
	metazone   string
	start, end int64
}

// timeZoneFormatRefs holds the resolved GMT/hour offset format strings for one
// locale.
type timeZoneFormatRefs struct {
	gmtFormat, gmtZeroFormat, hourFormat string
}

type localizedNamesRecord struct {
	metazone map[string]metazoneNames
	timeZone map[string]metazoneNames
	cities   map[string]string
}

var (
	metazonePeriodOnce sync.Once
	zoneToMetazones    map[string][]metazonePeriod

	namesOnce              sync.Once
	metazoneNamesByLocale  map[Locale]map[string]metazoneNames
	timeZoneNamesByLocale  map[Locale]map[string]metazoneNames
	exemplarCitiesByLocale map[Locale]map[string]string

	formatsOnce             sync.Once
	timeZoneFormatsByLocale map[Locale]timeZoneFormatRefs
)

func loadMetazonePeriods() {
	r := codec.NewReader(_tzMetazonePeriodBlob)
	zoneToMetazones = codec.StringRefKeyMap[[]metazonePeriod](&r, _data, decodeMetazonePeriods)
}

func decodeMetazonePeriods(r *codec.Reader) []metazonePeriod {
	return codec.CountedSlice[metazonePeriod](r, decodeMetazonePeriod)
}

func decodeMetazonePeriod(r *codec.Reader) metazonePeriod {
	return metazonePeriod{
		metazone: r.StringRef(_data),
		start:    r.Zigzag(),
		end:      r.Zigzag(),
	}
}

func loadNames() {
	r := codec.NewReader(_tzNamesBlob)
	records := codec.Uint16DeltaMap[Locale, localizedNamesRecord](&r, decodeLocalizedNamesRecord)
	metazoneNamesByLocale = make(map[Locale]map[string]metazoneNames, len(records))
	timeZoneNamesByLocale = make(map[Locale]map[string]metazoneNames, len(records))
	exemplarCitiesByLocale = make(map[Locale]map[string]string, len(records))
	for loc, record := range records {
		if record.metazone != nil {
			metazoneNamesByLocale[loc] = record.metazone
		}
		if record.timeZone != nil {
			timeZoneNamesByLocale[loc] = record.timeZone
		}
		if record.cities != nil {
			exemplarCitiesByLocale[loc] = record.cities
		}
	}
}

func decodeLocalizedNamesRecord(r *codec.Reader) localizedNamesRecord {
	return localizedNamesRecord{
		metazone: decodeMetazoneNameMap(r),
		timeZone: decodeMetazoneNameMap(r),
		cities:   r.StringRefMap(_data),
	}
}

func decodeMetazoneNameMap(r *codec.Reader) map[string]metazoneNames {
	return codec.StringRefKeyMap[metazoneNames](r, _data, decodeMetazoneNames)
}

func decodeMetazoneNames(r *codec.Reader) metazoneNames {
	return metazoneNames{
		longGeneric:   r.StringRef(_data),
		longStandard:  r.StringRef(_data),
		longDaylight:  r.StringRef(_data),
		shortGeneric:  r.StringRef(_data),
		shortStandard: r.StringRef(_data),
		shortDaylight: r.StringRef(_data),
	}
}

func loadFormats() {
	r := codec.NewReader(_tzFormatsBlob)
	timeZoneFormatsByLocale = codec.Uint16DeltaMap[Locale, timeZoneFormatRefs](&r, decodeTimeZoneFormatRefs)
}

func decodeTimeZoneFormatRefs(r *codec.Reader) timeZoneFormatRefs {
	return timeZoneFormatRefs{
		gmtFormat:     r.StringRef(_data),
		gmtZeroFormat: r.StringRef(_data),
		hourFormat:    r.StringRef(_data),
	}
}

var supported = codec.NewLazyStrings(_tzSupportedBlob, _data)
