// Package localeprofile owns the tools/locale-profile.json contract shared by
// CLDR-backed generators.
package localeprofile

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
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
	if err := unmarshalProfile(raw, &profile); err != nil {
		return Profile{}, fmt.Errorf("parse locale profile %s: %w", path, err)
	}
	if err := profile.Normalize(); err != nil {
		return Profile{}, fmt.Errorf("invalid locale profile %s: %w", path, err)
	}
	return profile, nil
}

func unmarshalProfile(raw []byte, profile *Profile) error {
	err := json.Unmarshal(raw, profile, json.RejectUnknownMembers(true))
	if err == nil {
		return nil
	}
	if errors.Is(err, json.ErrUnknownName) {
		var semantic *json.SemanticError
		if errors.As(err, &semantic) {
			name := strings.TrimPrefix(string(semantic.JSONPointer), "/")
			return fmt.Errorf("unknown field %q: %w", name, err)
		}
	}
	var syntactic *jsontext.SyntacticError
	if errors.As(err, &syntactic) && strings.Contains(err.Error(), "after top-level value") {
		return fmt.Errorf("multiple JSON values: %w", err)
	}
	return err
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
