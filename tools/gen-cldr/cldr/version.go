// Package cldr ingests the unicode-org/cldr-json npm distribution into
// typed Go structs for the generator extract layer to consume. It is
// generator-only; the runtime package shares the name but lives at
// internal/cldr/ in the main module.
package cldr

import (
	"encoding/json/v2"
	"fmt"
	"path/filepath"

	"github.com/agentable/go-intl/tools/internal/datapin"
)

// Versions captures the CLDR / ICU / tzdata pin from internal/cldr/VERSION.
type Versions = datapin.Versions

// ReadVersionFile reads the shared strict VERSION grammar.
func ReadVersionFile(path string) (Versions, error) { return datapin.ReadVersions(path) }

// CrossCheck verifies every CLDR package consumed by the generator against
// the pinned package identity before any output is written.
func CrossCheck(cldrJSONRoot string, want Versions) error {
	for _, name := range datapin.RequiredCLDRPackages() {
		pkgPath := filepath.Join(cldrJSONRoot, name, "package.json")
		raw, err := readRequiredFile(pkgPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", pkgPath, err)
		}
		var meta struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}
		if err := json.Unmarshal(raw, &meta); err != nil {
			return fmt.Errorf("parse %s: %w", pkgPath, err)
		}
		if meta.Name != name || meta.Version != want.CLDR {
			return fmt.Errorf("CLDR package %s: expected name %q version %q, got name %q version %q", pkgPath, name, want.CLDR, meta.Name, meta.Version)
		}
	}
	return nil
}
