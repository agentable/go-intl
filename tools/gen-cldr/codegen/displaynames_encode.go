package codegen

import "github.com/agentable/go-intl/tools/gen-cldr/cldr"

// encodeDisplayNames renders the const-only payload for the displaynames domain.
// It emits a private _data string table plus six independent blobs, each split
// along an accessor-reachable boundary so that importing displaynames and
// querying only one kind never decodes the others:
//
//   - _dnLanguageBlob:      per-locale language display names (dialect and
//     standard, each long/short/narrow) plus the locale pattern, which the
//     language accessor uses when composing language-with-region names.
//   - _dnTerritoryBlob:     per-locale region display names (long/short/narrow).
//   - _dnScriptBlob:        per-locale script display names.
//   - _dnCalendarBlob:      per-locale calendar display names.
//   - _dnDateTimeFieldBlob: per-locale date-time-field display names.
//   - _dnSupportedBlob:     the supported locale tags in sorted order, a narrow
//     index so SupportedLocales never decodes any names blob.
//
// Every names blob is keyed by the locale tag itself (a StringRef into _data),
// written in sorted-tag order. The decoder rebuilds each per-kind
// map[string]styledNames behind a per-blob sync.Once. The language accessor's
// region composition reads the territory blob, so a language lookup may also
// gate the territory decode; that is a genuine data dependency, not a shared
// Once.
func encodeDisplayNames(input RuntimeInput, table *StringTable) ([]byte, error) {
	data := input.DisplayNames
	locales := sortedLocaleKeys(data)

	var language blobEncoder
	appendStringRefKeyMap(&language, data, table, func(d cldr.DisplayNames) {
		encodeLanguageDisplayNames(&language, d, table)
	})

	territory := encodeStyledLocales(table, func(d cldr.DisplayNames) cldr.StyledNames { return d.Territories }, data)
	script := encodeStyledLocales(table, func(d cldr.DisplayNames) cldr.StyledNames { return d.Scripts }, data)
	calendar := encodeStyledLocales(table, func(d cldr.DisplayNames) cldr.StyledNames { return d.Calendars }, data)
	field := encodeStyledLocales(table, func(d cldr.DisplayNames) cldr.StyledNames { return d.DateTimeFields }, data)

	var supported blobEncoder
	supported.appendStringRefSlice(locales, table)

	return renderPayloadFile("displaynames", table,
		payloadBlob{"_dnLanguageBlob", language.bytes()},
		payloadBlob{"_dnTerritoryBlob", territory.bytes()},
		payloadBlob{"_dnScriptBlob", script.bytes()},
		payloadBlob{"_dnCalendarBlob", calendar.bytes()},
		payloadBlob{"_dnDateTimeFieldBlob", field.bytes()},
		payloadBlob{"_dnSupportedBlob", supported.bytes()},
	)
}

// encodeStyledLocales serializes one styled-name kind for every locale: the
// locale-tag StringRef followed by the long/short/narrow string maps selected by
// pick. Locales are written in the caller's sorted order, matching the decode
// loop.
func encodeStyledLocales(table *StringTable, pick func(cldr.DisplayNames) cldr.StyledNames, data map[string]cldr.DisplayNames) blobEncoder {
	var e blobEncoder
	appendStringRefKeyMap(&e, data, table, func(d cldr.DisplayNames) {
		encodeStyledNames(&e, pick(d), table)
	})
	return e
}

func encodeLanguageDisplayNames(e *blobEncoder, d cldr.DisplayNames, table *StringTable) {
	encodeStyledNames(e, d.Languages.Dialect, table)
	encodeStyledNames(e, d.Languages.Standard, table)
	e.appendStringRef(table.Add(d.LocalePattern))
}

// encodeStyledNames serializes a StyledNames as three string maps in
// long/short/narrow order, the order the decoder reads them back.
func encodeStyledNames(e *blobEncoder, s cldr.StyledNames, table *StringTable) {
	e.appendStringRefMap(s.Long, table)
	e.appendStringRefMap(s.Short, table)
	e.appendStringRefMap(s.Narrow, table)
}
