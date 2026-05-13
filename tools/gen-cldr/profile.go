package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
)

type LocaleProfile struct {
	Locales          []string `json:"locales"`
	FormatterLocales []string `json:"formatterLocales"`
	NumberLocales    []string `json:"numberLocales"`
	PluralLocales    []string `json:"pluralLocales"`
}

func readLocaleProfile(path string) (LocaleProfile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return LocaleProfile{}, fmt.Errorf("read locale profile %s: %w", path, err)
	}
	var profile LocaleProfile
	if err := json.Unmarshal(raw, &profile); err != nil {
		return LocaleProfile{}, fmt.Errorf("parse locale profile %s: %w", path, err)
	}
	if err := profile.normalize(); err != nil {
		return LocaleProfile{}, fmt.Errorf("invalid locale profile %s: %w", path, err)
	}
	return profile, nil
}

func (p *LocaleProfile) normalize() error {
	p.Locales = sortedUniqueLocales(p.Locales)
	p.FormatterLocales = sortedUniqueLocales(p.FormatterLocales)
	p.NumberLocales = sortedUniqueLocales(p.NumberLocales)
	p.PluralLocales = sortedUniqueLocales(p.PluralLocales)
	if len(p.Locales) == 0 {
		return fmt.Errorf("locales is empty")
	}
	if len(p.FormatterLocales) == 0 {
		p.FormatterLocales = slices.Clone(p.Locales)
	}
	if len(p.NumberLocales) == 0 {
		p.NumberLocales = slices.Clone(p.FormatterLocales)
	}
	if len(p.PluralLocales) == 0 {
		p.PluralLocales = slices.Clone(p.NumberLocales)
	}
	available := map[string]bool{}
	for _, locale := range p.Locales {
		available[locale] = true
	}
	if err := requireProfileSubset("formatter", p.FormatterLocales, available); err != nil {
		return err
	}
	if err := requireProfileSubset("number", p.NumberLocales, available); err != nil {
		return err
	}
	if err := requireProfileSubset("plural", p.PluralLocales, available); err != nil {
		return err
	}
	return nil
}

func requireProfileSubset(name string, locales []string, available map[string]bool) error {
	for _, locale := range locales {
		if !available[locale] {
			return fmt.Errorf("%s locale %q is not present in locales", name, locale)
		}
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
