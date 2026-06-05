package codegen

import (
	"bytes"
	"maps"
	"slices"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// encodeRelativeTime renders the const-only payload for the relativetime domain.
// It emits a private _data string table plus two independent blobs, each
// prefixed by its record count:
//
//   - _relativeTimeFieldBlob:    per-locale relative-time field data
//     (unit -> style -> {future, past, relative} plural/numeric pattern maps).
//     Locale indices are written as a sorted delta stream; the decoder rebuilds
//     the same map[Locale]RelativeTimeFields the legacy loadRelativeTimeFieldData
//     produced.
//   - _relativeTimeSupportedBlob: the relative-time-supported locale tags in
//     sorted-locale order, a narrow index so SupportedLocales never decodes the
//     field blob.
func encodeRelativeTime(input RuntimeInput, table *StringTable) ([]byte, error) {
	data := input.RelativeTime
	localeIndex := localeIndexMap(input.Locales)

	var fields blobEncoder
	locales := relativeTimeLocales(data)
	fields.appendUvarint(uint64(len(locales)))
	for _, locale := range locales {
		fields.appendDelta(mustLocaleIndex(localeIndex, locale))
		encodeRelativeTimeFields(&fields, data[locale], table)
	}

	var supported blobEncoder
	supportedTags := relativeTimeSupportedLocaleTags(data)
	supported.appendUvarint(uint64(len(supportedTags)))
	for _, tag := range supportedTags {
		supported.appendStringRef(table.Add(tag))
	}

	var b bytes.Buffer
	b.WriteString("package relativetime\n\n")
	if err := table.EmitDataConst(&b, "_data"); err != nil {
		return nil, err
	}
	for _, blob := range []struct {
		name  string
		bytes []byte
	}{
		{"_relativeTimeFieldBlob", fields.bytes()},
		{"_relativeTimeSupportedBlob", supported.bytes()},
	} {
		b.WriteString("\n")
		if err := emitStringConst(&b, blob.name, string(blob.bytes)); err != nil {
			return nil, err
		}
	}
	return FormatFile(b.Bytes())
}

// encodeRelativeTimeFields serializes one locale's RelativeTimeFields: a sorted
// unit map whose values are sorted style maps of {future, past, relative}
// pattern maps. The decoder reads the fields back in this exact order.
func encodeRelativeTimeFields(e *blobEncoder, values cldr.RelativeTimeFields, table *StringTable) {
	units := slices.Sorted(maps.Keys(values))
	e.appendUvarint(uint64(len(units)))
	for _, unit := range units {
		e.appendStringRef(table.Add(unit))
		styles := values[unit]
		styleKeys := slices.Sorted(maps.Keys(styles))
		e.appendUvarint(uint64(len(styleKeys)))
		for _, style := range styleKeys {
			e.appendStringRef(table.Add(style))
			field := styles[style]
			encodeStringMap(e, field.Future, table)
			encodeStringMap(e, field.Past, table)
			encodeStringMap(e, field.Relative, table)
		}
	}
}

// relativeTimeLocales returns the locales with relative-time data in
// sorted-locale order, matching the legacy relativeTimeFieldsByLocale key set.
func relativeTimeLocales(data extract.RelativeTimeFields) []string {
	return sortedLocaleKeys(data)
}

// relativeTimeSupportedLocaleTags returns the relative-time-supported locale
// tags in sorted-locale order, matching the legacy RelativeTimeSupportedLocales
// accessor (which walked the global locale table and kept locales present in the
// relative-time data map).
func relativeTimeSupportedLocaleTags(data extract.RelativeTimeFields) []string {
	return relativeTimeLocales(data)
}
