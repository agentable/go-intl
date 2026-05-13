package cldr

import (
	_ "embed"
	"strings"
)

// Locale is an opaque handle into the CLDR data tables generated under
// internal/cldr/. Index 0 is the "und" (Undefined) sentinel; every other
// index is assigned at codegen time in sorted-tag order.
type Locale uint16

// Undefined is the "und" sentinel locale. ResolveLocale returns
// (Undefined, false) when the requested tag does not appear in the embedded
// data set; callers must apply SPEC 11 best-fit matching against
// AvailableLocales in that case.
const Undefined Locale = 0

// VersionInfo carries the CLDR / ICU / tzdata versions pinned in
// internal/cldr/VERSION at codegen time.
type VersionInfo struct {
	CLDR   string
	ICU    string
	TZData string
}

//go:embed VERSION
var versionFile string

// Version returns the pinned data versions read from internal/cldr/VERSION.
// The result is stable across calls (parsed once at package init).
func Version() VersionInfo { return parsedVersion }

var parsedVersion = parseVersion(versionFile)

func parseVersion(raw string) VersionInfo {
	var v VersionInfo
	for line := range strings.SplitSeq(raw, "\n") {
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "cldr":
			v.CLDR = strings.TrimSpace(val)
		case "icu":
			v.ICU = strings.TrimSpace(val)
		case "tzdata":
			v.TZData = strings.TrimSpace(val)
		}
	}
	return v
}
