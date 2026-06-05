// Hand-written decode layer for the timezone domain. It expands the const blobs
// in data.go into the same in-memory structures the legacy root loadMetazoneData
// produced, behind per-blob sync.Once gates so that granularity follows accessor
// reachability: metazone periods, localized display names, and GMT offset formats
// each decode on first use of their own accessor and not before.
//
// Locale handle ownership: the names and formats blobs pack a locale index that
// must match the runtime locale assignment, so this package borrows the Locale
// handle type from the cldr/locale kernel package. That kernel assigns locale
// indices identically to the legacy root package, so keys align; importing the
// kernel (rather than the root) keeps timezone's compile cost off the root data
// bomb. The dependency is one-way (timezone -> cldr/locale; the kernel never
// imports timezone, so there is no cycle), and it is the permanent home every
// domain points at.

package timezone

import (
	"sync"

	"github.com/agentable/go-intl/internal/cldr/codec"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// Locale is the borrowed locale handle (see file header).
type Locale = cldrlocale.Locale

// metazoneNames carries the six width/usage variants of one metazone or zone
// label. It mirrors the legacy root metazoneNames field-for-field; the blob
// StringRefs are resolved against _data at decode time, so each field already
// holds the final string the accessor returns.
type metazoneNames struct {
	longGeneric, longStandard, longDaylight    string
	shortGeneric, shortStandard, shortDaylight string
}

// metazonePeriod is one resolved [start, end) window tagging a zone with the
// metazone in force during it. start/end are unix-milli instants whose open ends
// are the int64 min/max sentinels. It mirrors the legacy root metazonePeriod.
type metazonePeriod struct {
	metazone   string
	start, end int64
}

// timeZoneFormatRefs holds the resolved GMT/hour offset format strings for one
// locale, mirroring the legacy root timeZoneFormatRefs (after StringRef
// resolution).
type timeZoneFormatRefs struct {
	gmtFormat, gmtZeroFormat, hourFormat string
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

	supportedOnce sync.Once
	supportedTags []string
)

func loadMetazonePeriods() {
	r := codec.NewReader(_tzMetazonePeriodBlob)
	count := r.Uvarint()
	zoneToMetazones = make(map[string][]metazonePeriod, count)
	for i := uint64(0); i < count; i++ {
		zone := r.StringRef(_data)
		nPeriods := r.Uvarint()
		periods := make([]metazonePeriod, nPeriods)
		for j := range periods {
			periods[j] = metazonePeriod{
				metazone: r.StringRef(_data),
				start:    r.Zigzag(),
				end:      r.Zigzag(),
			}
		}
		zoneToMetazones[zone] = periods
	}
}

func loadNames() {
	r := codec.NewReader(_tzNamesBlob)
	count := r.Uvarint()
	metazoneNamesByLocale = make(map[Locale]map[string]metazoneNames, count)
	timeZoneNamesByLocale = make(map[Locale]map[string]metazoneNames, count)
	exemplarCitiesByLocale = make(map[Locale]map[string]string, count)
	indices := codec.Delta(&r)
	for i := uint64(0); i < count; i++ {
		loc := Locale(indices.Next())
		if names := decodeMetazoneNameMap(&r); names != nil {
			metazoneNamesByLocale[loc] = names
		}
		if zoneNames := decodeMetazoneNameMap(&r); zoneNames != nil {
			timeZoneNamesByLocale[loc] = zoneNames
		}
		if cities := decodeExemplarCityMap(&r); cities != nil {
			exemplarCitiesByLocale[loc] = cities
		}
	}
}

func decodeMetazoneNameMap(r *codec.Reader) map[string]metazoneNames {
	n := r.Uvarint()
	if n == 0 {
		return nil
	}
	out := make(map[string]metazoneNames, n)
	for i := uint64(0); i < n; i++ {
		key := r.StringRef(_data)
		out[key] = metazoneNames{
			longGeneric:   r.StringRef(_data),
			longStandard:  r.StringRef(_data),
			longDaylight:  r.StringRef(_data),
			shortGeneric:  r.StringRef(_data),
			shortStandard: r.StringRef(_data),
			shortDaylight: r.StringRef(_data),
		}
	}
	return out
}

func decodeExemplarCityMap(r *codec.Reader) map[string]string {
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

func loadFormats() {
	r := codec.NewReader(_tzFormatsBlob)
	count := r.Uvarint()
	timeZoneFormatsByLocale = make(map[Locale]timeZoneFormatRefs, count)
	indices := codec.Delta(&r)
	for i := uint64(0); i < count; i++ {
		loc := Locale(indices.Next())
		timeZoneFormatsByLocale[loc] = timeZoneFormatRefs{
			gmtFormat:     r.StringRef(_data),
			gmtZeroFormat: r.StringRef(_data),
			hourFormat:    r.StringRef(_data),
		}
	}
}

func loadSupported() {
	r := codec.NewReader(_tzSupportedBlob)
	count := r.Uvarint()
	supportedTags = make([]string, count)
	for i := range supportedTags {
		supportedTags[i] = r.StringRef(_data)
	}
}
