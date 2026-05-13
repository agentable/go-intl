// Package cldr is the data layer for every go-intl formatter. It exposes
// O(1), zero-allocation accessors over locale-aware CLDR tables (number
// symbols, calendar names, currency precision, metazone display names, likely
// subtags, language-matching distances, etc.).
//
// All data is compiled at codegen time by tools/gen-cldr/ from the
// unicode-org/cldr-json npm distribution; runtime never parses JSON, never
// performs file I/O, and never reaches the network. The pinned CLDR / ICU /
// tzdata versions live in internal/cldr/VERSION (single source of truth) —
// see SPECS/50-cldr-data.md for the upgrade flow.
//
// Generated files (numbers.go, dates.go, metazones.go, currencies.go,
// units.go, preference.go, likely_subtags.go, locale_matching.go,
// regions.go, locales.go, strings.go) are not hand-edited; the snapshot test
// in this package fails when committed bytes drift from generator output.
package cldr
