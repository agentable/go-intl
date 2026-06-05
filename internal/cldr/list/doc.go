// Package list exposes generated CLDR list-formatting data.
//
// The list patterns live as a const-only blob payload in data.go, decoded
// lazily by decode.go and queried through the accessors in accessors.go. Locale
// handling (resolution, maximize) is borrowed from the cldr/locale kernel so
// list indexes locales identically to every other domain.
//
// Only internal CLDR and ListFormat code should use this package; public
// callers use listformat APIs.
package list
