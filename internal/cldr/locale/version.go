// Hand-written version accessor for the locale kernel. The CLDR / ICU / tzdata
// pins live in internal/cldr/VERSION (single source of truth) and are baked into
// the generated manifest at codegen time, so Version derives from the manifest
// rather than embedding VERSION a second time.

package cldrlocale

// VersionInfo carries the CLDR / ICU / tzdata versions pinned in
// internal/cldr/VERSION at codegen time.
type VersionInfo struct {
	CLDR   string
	ICU    string
	TZData string
}

// Version returns the pinned data versions, sourced from the generated manifest.
func Version() VersionInfo {
	return VersionInfo{
		CLDR:   dataManifest.CLDR,
		ICU:    dataManifest.ICU,
		TZData: dataManifest.TZData,
	}
}
