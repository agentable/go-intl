package codegen

import (
	"bytes"
	"maps"
	"slices"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

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
// written in sorted-tag order. The decoder rebuilds the same per-kind
// map[string]styledNames the legacy localeData literal carried, behind a
// per-blob sync.Once. The language accessor's region composition reads the
// territory blob, so a language lookup may also gate the territory decode; that
// is a genuine data dependency, not a shared Once.
func encodeDisplayNames(input RuntimeInput, table *StringTable) ([]byte, error) {
	data := input.DisplayNames
	locales := sortedDisplayNamesLocales(data)

	var language blobEncoder
	language.appendUvarint(uint64(len(locales)))
	for _, locale := range locales {
		d := data[locale]
		language.appendStringRef(table.Add(locale))
		encodeStyledNames(&language, d.Languages.Dialect, table)
		encodeStyledNames(&language, d.Languages.Standard, table)
		language.appendStringRef(table.Add(d.LocalePattern))
	}

	territory := encodeStyledLocales(locales, table, func(d cldr.DisplayNames) cldr.StyledNames { return d.Territories }, data)
	script := encodeStyledLocales(locales, table, func(d cldr.DisplayNames) cldr.StyledNames { return d.Scripts }, data)
	calendar := encodeStyledLocales(locales, table, func(d cldr.DisplayNames) cldr.StyledNames { return d.Calendars }, data)
	field := encodeStyledLocales(locales, table, func(d cldr.DisplayNames) cldr.StyledNames { return d.DateTimeFields }, data)

	var supported blobEncoder
	supported.appendUvarint(uint64(len(locales)))
	for _, locale := range locales {
		supported.appendStringRef(table.Add(locale))
	}

	var b bytes.Buffer
	b.WriteString("package displaynames\n\n")
	if err := table.EmitDataConst(&b, "_data"); err != nil {
		return nil, err
	}
	for _, blob := range []struct {
		name  string
		bytes []byte
	}{
		{"_dnLanguageBlob", language.bytes()},
		{"_dnTerritoryBlob", territory.bytes()},
		{"_dnScriptBlob", script.bytes()},
		{"_dnCalendarBlob", calendar.bytes()},
		{"_dnDateTimeFieldBlob", field.bytes()},
		{"_dnSupportedBlob", supported.bytes()},
	} {
		b.WriteString("\n")
		if err := emitStringConst(&b, blob.name, string(blob.bytes)); err != nil {
			return nil, err
		}
	}
	return FormatFile(b.Bytes())
}

// sortedDisplayNamesLocales returns the display-name locale tags in sorted
// order, the order every blob and the narrow supported index are written in.
func sortedDisplayNamesLocales(data map[string]cldr.DisplayNames) []string {
	return slices.Sorted(maps.Keys(data))
}

// encodeStyledLocales serializes one styled-name kind for every locale: the
// locale-tag StringRef followed by the long/short/narrow string maps selected by
// pick. Locales are written in the caller's sorted order, matching the decode
// loop and the legacy literal ordering.
func encodeStyledLocales(locales []string, table *StringTable, pick func(cldr.DisplayNames) cldr.StyledNames, data map[string]cldr.DisplayNames) blobEncoder {
	var e blobEncoder
	e.appendUvarint(uint64(len(locales)))
	for _, locale := range locales {
		e.appendStringRef(table.Add(locale))
		encodeStyledNames(&e, pick(data[locale]), table)
	}
	return e
}

// encodeStyledNames serializes a StyledNames as three string maps in
// long/short/narrow order, the order the decoder reads them back.
func encodeStyledNames(e *blobEncoder, s cldr.StyledNames, table *StringTable) {
	encodeStringMap(e, s.Long, table)
	encodeStringMap(e, s.Short, table)
	encodeStringMap(e, s.Narrow, table)
}
