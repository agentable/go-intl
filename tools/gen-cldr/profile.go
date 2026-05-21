package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
)

// LocaleProfile is the maintained record for generated CLDR coverage.
// CLDR-backed surfaces derive payloads from this shared locale list; surfaces
// backed by another runtime engine keep their supported-locale sets separate.
type LocaleProfile struct {
	Locales []string `json:"locales"`
}

func readLocaleProfile(path string) (LocaleProfile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return LocaleProfile{}, fmt.Errorf("read locale profile %s: %w", path, err)
	}
	var profile LocaleProfile
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&profile); err != nil {
		return LocaleProfile{}, fmt.Errorf("parse locale profile %s: %w", path, err)
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return LocaleProfile{}, fmt.Errorf("parse locale profile %s: multiple JSON values", path)
		}
		return LocaleProfile{}, fmt.Errorf("parse locale profile %s: %w", path, err)
	}
	if err := profile.normalize(); err != nil {
		return LocaleProfile{}, fmt.Errorf("invalid locale profile %s: %w", path, err)
	}
	return profile, nil
}

func (p *LocaleProfile) normalize() error {
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
