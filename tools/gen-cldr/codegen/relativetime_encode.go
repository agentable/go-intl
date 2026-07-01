package codegen

import (
	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

// encodeRelativeTime renders the const-only payload for the relativetime domain.
// It emits a private _data string table plus two independent blobs, each
// prefixed by its record count:
//
//   - _relativeTimeFieldBlob:    per-locale relative-time field data
//     (unit -> style -> {future, past, relative} plural/numeric pattern maps).
//     Locale indices are written as a sorted delta stream; the decoder rebuilds
//     the runtime map[Locale]RelativeTimeFields.
//   - _relativeTimeSupportedBlob: the relative-time-supported locale tags in
//     sorted-locale order, a narrow index so SupportedLocales never decodes the
//     field blob.
func encodeRelativeTime(input RuntimeInput, table *StringTable) ([]byte, error) {
	data := input.RelativeTime
	localeIndex := localeIndexMap(input.Locales)

	var fields blobEncoder
	locales := sortedLocaleKeys(data)
	if err := fields.appendLocaleDeltaRecords(locales, localeIndex, func(locale string) {
		encodeRelativeTimeFields(&fields, data[locale], table)
	}); err != nil {
		return nil, err
	}

	var supported blobEncoder
	supported.appendStringRefSlice(locales, table)

	return renderPayloadFile("relativetime", table,
		payloadBlob{"_relativeTimeFieldBlob", fields.bytes()},
		payloadBlob{"_relativeTimeSupportedBlob", supported.bytes()},
	)
}

// encodeRelativeTimeFields serializes one locale's RelativeTimeFields: a sorted
// unit map whose values are sorted style maps of {future, past, relative}
// pattern maps. The decoder reads the fields back in this exact order.
func encodeRelativeTimeFields(e *blobEncoder, values cldr.RelativeTimeFields, table *StringTable) {
	appendStringRefKeyMap(e, values, table, func(styles map[string]cldr.RelativeTimeField) {
		appendStringRefKeyMap(e, styles, table, func(field cldr.RelativeTimeField) {
			appendRelativeTimeField(e, field, table)
		})
	})
}

// appendRelativeTimeField owns the future/past/relative map order for one
// (unit, style) payload row.
func appendRelativeTimeField(e *blobEncoder, field cldr.RelativeTimeField, table *StringTable) {
	e.appendStringRefMap(field.Future, table)
	e.appendStringRefMap(field.Past, table)
	e.appendStringRefMap(field.Relative, table)
}
