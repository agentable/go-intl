package codegen

import (
	"bytes"
	"maps"
	"slices"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// encodeList renders the const-only payload for the list domain. It emits a
// private _data string table plus two independent blobs, each prefixed by its
// record count:
//
//   - _listPatternBlob:    per-locale list-pattern data (type -> style ->
//     {pair, start, middle, end}). Locale indices are written as a sorted delta
//     stream; the decoder rebuilds the same
//     map[Locale]map[string]map[string]listPatternRefs the legacy
//     loadListPatternData produced.
//   - _listSupportedBlob:  the list-supported locale tags in sorted-locale
//     order, a narrow index so SupportedLocales never decodes the pattern blob.
func encodeList(input RuntimeInput, table *StringTable) ([]byte, error) {
	data := input.ListPatterns
	localeIndex := localeIndexMap(input.Locales)

	var patterns blobEncoder
	locales := listLocales(data)
	patterns.appendUvarint(uint64(len(locales)))
	for _, locale := range locales {
		patterns.appendDelta(mustLocaleIndex(localeIndex, locale))
		encodeListPatterns(&patterns, data[locale], table)
	}

	var supported blobEncoder
	supportedTags := listSupportedLocaleTags(data)
	supported.appendUvarint(uint64(len(supportedTags)))
	for _, tag := range supportedTags {
		supported.appendStringRef(table.Add(tag))
	}

	var b bytes.Buffer
	b.WriteString("package list\n\n")
	if err := table.EmitDataConst(&b, "_data"); err != nil {
		return nil, err
	}
	for _, blob := range []struct {
		name  string
		bytes []byte
	}{
		{"_listPatternBlob", patterns.bytes()},
		{"_listSupportedBlob", supported.bytes()},
	} {
		b.WriteString("\n")
		if err := emitStringConst(&b, blob.name, string(blob.bytes)); err != nil {
			return nil, err
		}
	}
	return FormatFile(b.Bytes())
}

// encodeListPatterns serializes one locale's list patterns: a sorted type map
// whose values are sorted style maps of {pair, start, middle, end} pattern
// strings. The decoder reads them back in this exact order.
func encodeListPatterns(e *blobEncoder, values cldr.ListPatterns, table *StringTable) {
	types := slices.Sorted(maps.Keys(values))
	e.appendUvarint(uint64(len(types)))
	for _, typ := range types {
		e.appendStringRef(table.Add(typ))
		styles := values[typ]
		styleKeys := slices.Sorted(maps.Keys(styles))
		e.appendUvarint(uint64(len(styleKeys)))
		for _, style := range styleKeys {
			e.appendStringRef(table.Add(style))
			pattern := styles[style]
			e.appendStringRef(table.Add(pattern.Pair))
			e.appendStringRef(table.Add(pattern.Start))
			e.appendStringRef(table.Add(pattern.Middle))
			e.appendStringRef(table.Add(pattern.End))
		}
	}
}

// listLocales returns the locales with list-pattern data in sorted-locale order,
// matching the legacy listPatternsByLocale key set.
func listLocales(data extract.ListPatterns) []string {
	return sortedLocaleKeys(data)
}

// listSupportedLocaleTags returns the list-supported locale tags in
// sorted-locale order, matching the legacy supportedLocales accessor (which
// walked the list pattern data map and sorted its keys).
func listSupportedLocaleTags(data extract.ListPatterns) []string {
	return listLocales(data)
}
