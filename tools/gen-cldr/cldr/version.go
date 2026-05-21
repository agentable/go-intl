// Package cldr ingests the unicode-org/cldr-json npm distribution into
// typed Go structs for the generator extract layer to consume. It is
// generator-only; the runtime package shares the name but lives at
// internal/cldr/ in the main module.
package cldr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Versions captures the CLDR / ICU / tzdata pin from internal/cldr/VERSION.
type Versions struct {
	CLDR   string
	ICU    string
	TZData string
}

// ReadVersionFile parses the three-line pin file. Unknown lines are ignored;
// any of the required keys (cldr/icu/tzdata) being absent is an error so the
// generator never runs with a partially-specified pin.
func ReadVersionFile(path string) (Versions, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Versions{}, fmt.Errorf("read %s: %w", path, err)
	}
	var v Versions
	for line := range strings.SplitSeq(string(raw), "\n") {
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
	if v.CLDR == "" || v.ICU == "" || v.TZData == "" {
		return Versions{}, fmt.Errorf("incomplete pin in %s: %+v", path, v)
	}
	return v, nil
}

// CrossCheck asserts the local cldr-json checkout matches the pinned CLDR
// version. It reads cldr-core/package.json (always present in any cldr-json
// distribution) and compares its `version` field.
func CrossCheck(cldrJSONRoot string, want Versions) error {
	pkgPath := filepath.Join(cldrJSONRoot, "cldr-core", "package.json")
	raw, err := os.ReadFile(pkgPath)
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
	if meta.Version != want.CLDR {
		return fmt.Errorf("cldr-core version %q at %s does not match pin %q",
			meta.Version, pkgPath, want.CLDR)
	}
	return nil
}
