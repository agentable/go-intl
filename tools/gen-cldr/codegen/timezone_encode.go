package codegen

import (
	"bytes"
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
//     same per-locale maps the legacy loadMetazoneData produced. Drives the
//     localized display-name lookups in TimeZoneDisplayName.
//   - _tzFormatsBlob: per-locale GMT/hour offset formats. Drives GMTOffsetName.
//   - _tzSupportedBlob: the canonicalized supported IANA zone names in sorted
//     order, a narrow index so SupportedTimeZones never decodes the period or
//     names blobs.
//
// Canonical links and region zones are not part of this payload: they are tiny
// static locale-region maps owned by the cldr/locale kernel, and the timezone
// accessors forward to the kernel for them.
func encodeTimezone(input RuntimeInput, table *StringTable) ([]byte, error) {
	data := input.Metazones
	localeIndex := localeIndexMap(input.Locales)

	periods := encodeMetazonePeriods(data, table)
	names := encodeMetazoneNames(data, localeIndex, table)
	formats := encodeTimeZoneFormats(data, localeIndex, table)

	var supported blobEncoder
	supportedTags := timezoneSupportedZones(data)
	supported.appendUvarint(uint64(len(supportedTags)))
	for _, tag := range supportedTags {
		supported.appendStringRef(table.Add(tag))
	}

	var b bytes.Buffer
	b.WriteString("package timezone\n\n")
	if err := table.EmitDataConst(&b, "_data"); err != nil {
		return nil, err
	}
	for _, blob := range []struct {
		name  string
		bytes []byte
	}{
		{"_tzMetazonePeriodBlob", periods.bytes()},
		{"_tzNamesBlob", names.bytes()},
		{"_tzFormatsBlob", formats.bytes()},
		{"_tzSupportedBlob", supported.bytes()},
	} {
		b.WriteString("\n")
		if err := emitStringConst(&b, blob.name, string(blob.bytes)); err != nil {
			return nil, err
		}
	}
	return FormatFile(b.Bytes())
}

// encodeMetazonePeriods serializes the zone -> []metazonePeriod map. Zones are
// written in sorted order with each zone's period count, then each period as
// (metazone stringRef, zigzag start, zigzag end). The decoder reconstructs the
// same map the legacy zoneToMetazones held.
func encodeMetazonePeriods(data extract.Metazones, table *StringTable) blobEncoder {
	var e blobEncoder
	zones := slices.Sorted(maps.Keys(data.ZoneToMetazones))
	e.appendUvarint(uint64(len(zones)))
	for _, zone := range zones {
		e.appendStringRef(table.Add(zone))
		periods := data.ZoneToMetazones[zone]
		e.appendUvarint(uint64(len(periods)))
		for _, period := range periods {
			e.appendStringRef(table.Add(period.Metazone))
			e.appendZigzag(period.Start)
			e.appendZigzag(period.End)
		}
	}
	return e
}

// encodeMetazoneNames serializes the per-locale display-name data: the metazone
// name map, the zone-specific name map, and the exemplar-city map. Locales carry
// a delta-encoded index and are ordered by that index so the decoder can replay
// the sorted delta stream. Every locale that appears in any of the three source
// maps is emitted, each section prefixed by its own count, so a locale present in
// only one section still round-trips.
func encodeMetazoneNames(data extract.Metazones, localeIndex map[string]uint64, table *StringTable) blobEncoder {
	var e blobEncoder
	locales := metazoneNameLocales(data)
	slices.SortFunc(locales, func(a, b string) int {
		return cmpUint64(mustLocaleIndex(localeIndex, a), mustLocaleIndex(localeIndex, b))
	})
	e.appendUvarint(uint64(len(locales)))
	for _, locale := range locales {
		e.appendDelta(mustLocaleIndex(localeIndex, locale))
		encodeMetazoneNameMap(&e, data.Names[locale], table)
		encodeMetazoneNameMap(&e, data.ZoneNames[locale], table)
		encodeExemplarCityMap(&e, data.ExemplarCities[locale], table)
	}
	return e
}

// encodeMetazoneNameMap writes a sorted key -> metazoneNames (six stringRefs) map.
func encodeMetazoneNameMap(e *blobEncoder, values map[string]cldr.MetazoneNames, table *StringTable) {
	keys := slices.Sorted(maps.Keys(values))
	e.appendUvarint(uint64(len(keys)))
	for _, key := range keys {
		e.appendStringRef(table.Add(key))
		names := values[key]
		e.appendStringRef(table.Add(names.LongGeneric))
		e.appendStringRef(table.Add(names.LongStandard))
		e.appendStringRef(table.Add(names.LongDaylight))
		e.appendStringRef(table.Add(names.ShortGeneric))
		e.appendStringRef(table.Add(names.ShortStandard))
		e.appendStringRef(table.Add(names.ShortDaylight))
	}
}

// encodeExemplarCityMap writes a sorted zone -> city stringRef map.
func encodeExemplarCityMap(e *blobEncoder, values map[string]string, table *StringTable) {
	keys := slices.Sorted(maps.Keys(values))
	e.appendUvarint(uint64(len(keys)))
	for _, key := range keys {
		e.appendStringRef(table.Add(key))
		e.appendStringRef(table.Add(values[key]))
	}
}

// encodeTimeZoneFormats serializes the per-locale GMT/hour offset formats. Only
// locales with a formats record are emitted, ordered by delta-encoded index.
func encodeTimeZoneFormats(data extract.Metazones, localeIndex map[string]uint64, table *StringTable) blobEncoder {
	var e blobEncoder
	locales := slices.Sorted(maps.Keys(data.Formats))
	slices.SortFunc(locales, func(a, b string) int {
		return cmpUint64(mustLocaleIndex(localeIndex, a), mustLocaleIndex(localeIndex, b))
	})
	e.appendUvarint(uint64(len(locales)))
	for _, locale := range locales {
		e.appendDelta(mustLocaleIndex(localeIndex, locale))
		formats := data.Formats[locale]
		e.appendStringRef(table.Add(formats.GMTFormat))
		e.appendStringRef(table.Add(formats.GMTZeroFormat))
		e.appendStringRef(table.Add(formats.HourFormat))
	}
	return e
}

// metazoneNameLocales returns the union of locales present in the metazone-name,
// zone-name, or exemplar-city maps, mirroring the legacy loadMetazoneData which
// keyed all three by locale.
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

// timezoneSupportedZones returns the canonicalized supported IANA zone names in
// sorted order, matching the legacy root supportedTimeZoneValues: every zone in
// the metazone-period map (minus Etc/Unknown), canonicalized through the kernel
// links, deduplicated, sorted.
func timezoneSupportedZones(data extract.Metazones) []string {
	seen := map[string]bool{}
	for zone := range data.ZoneToMetazones {
		if zone == "Etc/Unknown" {
			continue
		}
		seen[canonicalTimeZoneLink(zone)] = true
	}
	return slices.Sorted(maps.Keys(seen))
}

// canonicalTimeZoneLink mirrors the kernel CanonicalTimeZoneLink at generation
// time. The link table is the same three static entries the kernel
// loadTimeZoneData installs; keeping a generation-time copy lets the supported
// narrow index be precomputed without decoding the kernel at gen time.
func canonicalTimeZoneLink(name string) string {
	switch name {
	case "US/Eastern":
		return "America/New_York"
	case "US/Pacific":
		return "America/Los_Angeles"
	case "Asia/Calcutta":
		return "Asia/Kolkata"
	default:
		return name
	}
}
