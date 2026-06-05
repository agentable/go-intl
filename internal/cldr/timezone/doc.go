// Package timezone exposes generated CLDR time-zone data.
//
// It owns the metazone-period mappings, the localized metazone, zone, and
// exemplar-city display names, the GMT/hour offset formats, and the narrow
// supported-time-zone index, each behind its own lazy decode gate. Canonical
// links and region zones remain locale-region data owned by the cldr/locale
// kernel; the accessors here forward to it.
//
// Only internal CLDR and DateTimeFormat code should use this package; public
// callers use datetimeformat APIs.
package timezone
