package codegen

import (
	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

// encodeList renders the const-only payload for the list domain. It emits a
// private _data string table plus two independent blobs, each prefixed by its
// record count:
//
//   - _listPatternBlob:    per-locale list-pattern data (type -> style ->
//     {pair, start, middle, end}). Locale indices are written as a sorted delta
//     stream; the decoder rebuilds the runtime
//     map[Locale]map[string]map[string]listPatternRefs.
//   - _listSupportedBlob:  the list-supported locale tags in sorted-locale
//     order, a narrow index so SupportedLocales never decodes the pattern blob.
func encodeList(input RuntimeInput, table *StringTable) ([]byte, error) {
	data := input.ListPatterns
	localeIndex := localeIndexMap(input.Locales)

	var patterns blobEncoder
	locales := sortedLocaleKeys(data)
	if err := patterns.appendLocaleDeltaRecords(locales, localeIndex, func(locale string) {
		encodeListPatterns(&patterns, data[locale], table)
	}); err != nil {
		return nil, err
	}

	var supported blobEncoder
	supported.appendStringRefSlice(locales, table)

	return renderPayloadFile("list", table,
		payloadBlob{"_listPatternBlob", patterns.bytes()},
		payloadBlob{"_listSupportedBlob", supported.bytes()},
	)
}

// encodeListPatterns serializes one locale's list patterns: a sorted type map
// whose values are sorted style maps of {pair, start, middle, end} pattern
// strings. The decoder reads them back in this exact order.
func encodeListPatterns(e *blobEncoder, values cldr.ListPatterns, table *StringTable) {
	appendStringRefKeyMap(e, values, table, func(styles map[string]cldr.ListPattern) {
		appendStringRefKeyMap(e, styles, table, func(pattern cldr.ListPattern) {
			appendListPattern(e, pattern, table)
		})
	})
}

// appendListPattern owns the pair/start/middle/end wire order for one
// (type, style) list-pattern payload row.
func appendListPattern(e *blobEncoder, pattern cldr.ListPattern, table *StringTable) {
	e.appendStringRef(table.Add(pattern.Pair))
	e.appendStringRef(table.Add(pattern.Start))
	e.appendStringRef(table.Add(pattern.Middle))
	e.appendStringRef(table.Add(pattern.End))
}
