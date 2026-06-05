// Package currency exposes generated CLDR currency data.
//
// It keeps currency fraction-digit metadata and per-locale display names,
// canonical names, and symbols behind typed accessors. It is the one sanctioned
// shared CLDR owner: both NumberFormat and DisplayNames read currency names and
// symbols through it.
//
// Only internal CLDR, NumberFormat, and DisplayNames code should use this
// package; public callers use numberformat and displaynames APIs.
package currency
