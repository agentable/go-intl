package codegen

import (
	"maps"
	"slices"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// encodeTimezone renders the const-only payload for the timezone domain. It emits
// a private _data string table plus four independent blobs, each prefixed by its
// record count, so the granularity follows accessor reachability: a binary that
// formats no time zones links none of them.
//
//   - _tzMetazonePeriodBlob: zone -> []metazonePeriod (metazone, start, end). The
//     boundaries are int64 instants whose open ends are the int64 min/max
//     sentinels, written as a zigzag-signed stream. Drives TimeZoneMetazone.
//   - _tzNamesBlob: per-locale metazone names, zone-specific names, and exemplar
//     cities. Locale indices are a sorted delta stream; the decoder rebuilds the
//     per-locale maps that drive the localized display-name lookups in
//     TimeZoneDisplayName.
//   - _tzFormatsBlob: per-locale GMT/hour offset and location formats. Drives
//     GMTOffsetName and exemplar-city fallback in TimeZoneDisplayName.
//
// Identifier legality, primary mappings, and region membership are generated
// separately into internal/tz from the pinned IANA and CLDR BCP47 sources.
func encodeTimezone(input RuntimeInput, table *StringTable) ([]byte, error) {
	data := input.Metazones
	localeIndex := localeIndexMap(input.Locales)

	periods := encodeMetazonePeriods(data, table)
	names, err := encodeMetazoneNames(data, localeIndex, table)
	if err != nil {
		return nil, err
	}
	formats, err := encodeTimeZoneFormats(data, localeIndex, table)
	if err != nil {
		return nil, err
	}

	return renderPayloadFile("timezone", table,
		payloadBlob{"_tzMetazonePeriodBlob", periods.bytes()},
		payloadBlob{"_tzNamesBlob", names.bytes()},
		payloadBlob{"_tzFormatsBlob", formats.bytes()},
	)
}

// encodeMetazonePeriods serializes the zone -> []metazonePeriod map. Zones are
// written in sorted order with each zone's period count, then each period as
// (metazone stringRef, zigzag start, zigzag end). The decoder reconstructs the
// lookup map used by TimeZoneMetazone.
func encodeMetazonePeriods(data extract.Metazones, table *StringTable) blobEncoder {
	var e blobEncoder
	appendStringRefKeyMap(&e, data.ZoneToMetazones, table, func(periods []cldr.MetazonePeriod) {
		appendCountedSlice(&e, periods, func(period cldr.MetazonePeriod) {
			e.appendStringRef(table.Add(period.Metazone))
			e.appendZigzag(period.Start)
			e.appendZigzag(period.End)
		})
	})
	return e
}

// encodeMetazoneNames serializes the per-locale display-name data: the metazone
// name map, the zone-specific name map, and the exemplar-city map. Locales carry
// a delta-encoded index and are ordered by that index so the decoder can replay
// the sorted delta stream. Every locale that appears in any of the three source
// maps is emitted, each section prefixed by its own count, so a locale present in
// only one section still round-trips.
func encodeMetazoneNames(data extract.Metazones, localeIndex map[string]uint64, table *StringTable) (blobEncoder, error) {
	var e blobEncoder
	locales := metazoneNameLocales(data)
	if err := e.appendLocaleDeltaRecords(locales, localeIndex, func(locale string) {
		encodeMetazoneNameMap(&e, data.Names[locale], table)
		encodeMetazoneNameMap(&e, data.ZoneNames[locale], table)
		e.appendStringRefMap(data.ExemplarCities[locale], table)
	}); err != nil {
		return blobEncoder{}, err
	}
	return e, nil
}

// encodeMetazoneNameMap writes a sorted key -> metazoneNames (six stringRefs) map.
func encodeMetazoneNameMap(e *blobEncoder, values map[string]cldr.MetazoneNames, table *StringTable) {
	appendStringRefKeyMap(e, values, table, func(names cldr.MetazoneNames) {
		appendMetazoneNames(e, names, table)
	})
}

// appendMetazoneNames owns the six-field wire order shared by metazone and
// zone-specific name maps.
func appendMetazoneNames(e *blobEncoder, names cldr.MetazoneNames, table *StringTable) {
	e.appendStringRef(table.Add(names.LongGeneric))
	e.appendStringRef(table.Add(names.LongStandard))
	e.appendStringRef(table.Add(names.LongDaylight))
	e.appendStringRef(table.Add(names.ShortGeneric))
	e.appendStringRef(table.Add(names.ShortStandard))
	e.appendStringRef(table.Add(names.ShortDaylight))
}

// encodeTimeZoneFormats serializes the per-locale GMT/hour offset and location
// formats. Only locales with a formats record are emitted, ordered by
// delta-encoded index.
func encodeTimeZoneFormats(data extract.Metazones, localeIndex map[string]uint64, table *StringTable) (blobEncoder, error) {
	var e blobEncoder
	locales := slices.Sorted(maps.Keys(data.Formats))
	if err := e.appendLocaleDeltaRecords(locales, localeIndex, func(locale string) {
		appendTimeZoneFormats(&e, data.Formats[locale], table)
	}); err != nil {
		return blobEncoder{}, err
	}
	return e, nil
}

// appendTimeZoneFormats owns the four-field wire order for GMT/hour offset and
// location formatting records.
func appendTimeZoneFormats(e *blobEncoder, formats cldr.TimeZoneFormats, table *StringTable) {
	e.appendStringRef(table.Add(formats.GMTFormat))
	e.appendStringRef(table.Add(formats.GMTZeroFormat))
	e.appendStringRef(table.Add(formats.HourFormat))
	e.appendStringRef(table.Add(formats.RegionFormat))
}

// metazoneNameLocales returns the union of locales present in the metazone-name,
// zone-name, or exemplar-city maps so all three sections share one locale-indexed
// blob.
func metazoneNameLocales(data extract.Metazones) []string {
	seen := map[string]bool{}
	for locale := range data.Names {
		seen[locale] = true
	}
	for locale := range data.ZoneNames {
		seen[locale] = true
	}
	for locale := range data.ExemplarCities {
		seen[locale] = true
	}
	return slices.Sorted(maps.Keys(seen))
}
