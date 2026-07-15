// Package timezone exposes generated CLDR time-zone data.
//
// It owns the metazone-period mappings, the localized metazone, zone, and
// exemplar-city display names, and the GMT/hour offset and location formats,
// each behind its own lazy decode gate. Identifier legality, primary mappings,
// and region membership belong to internal/tz.
//
// Only internal CLDR and DateTimeFormat code should use this package; public
// callers use datetimeformat APIs.
package timezone
