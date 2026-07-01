// Package localeprofile owns the tools/locale-profile.json contract shared by
// CLDR-backed generators.
package localeprofile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
)

// Profile is the maintained record for generated CLDR coverage.
// CLDR-backed surfaces derive payloads from this shared locale list; surfaces
// backed by another runtime engine keep their supported-locale sets separate.
type Profile struct {
	Locales []string `json:"locales"`
}

// Read loads and normalizes a tools/locale-profile.json file.
func Read(path string) (Profile, error) {
	raw, err := os.ReadFile(path) // #nosec G304 -- profile path is a maintainer-supplied generator input.
	if err != nil {
		return Profile{}, fmt.Errorf("read locale profile %s: %w", path, err)
	}
	var profile Profile
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&profile); err != nil {
		return Profile{}, fmt.Errorf("parse locale profile %s: %w", path, err)
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return Profile{}, fmt.Errorf("parse locale profile %s: multiple JSON values", path)
		}
		return Profile{}, fmt.Errorf("parse locale profile %s: %w", path, err)
	}
	if err := profile.Normalize(); err != nil {
		return Profile{}, fmt.Errorf("invalid locale profile %s: %w", path, err)
	}
	return profile, nil
}

// Normalize sorts, deduplicates, and removes non-runtime sentinel entries.
func (p *Profile) Normalize() error {
	p.Locales = sortedUniqueLocales(p.Locales)
	if len(p.Locales) == 0 {
		return fmt.Errorf("locales is empty")
	}
	return nil
}

func sortedUniqueLocales(locales []string) []string {
	set := map[string]bool{}
	for _, locale := range locales {
		if locale != "" && locale != "und" {
			set[locale] = true
		}
	}
	return slices.Sorted(maps.Keys(set))
}
