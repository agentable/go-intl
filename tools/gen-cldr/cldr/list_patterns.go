package cldr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ListPatterns map[string]map[string]ListPattern

type ListPattern struct {
	Pair   string `json:"2"`
	Start  string `json:"start"`
	Middle string `json:"middle"`
	End    string `json:"end"`
}

func loadListPatterns(root string, locales []string) (map[string]ListPatterns, error) {
	loaded := make(map[string]ListPatterns)
	for _, locale := range locales {
		if locale == "und" {
			continue
		}
		path := filepath.Join(root, "cldr-misc-full", "main", locale, "listPatterns.json")
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var doc struct {
			Main map[string]struct {
				ListPatterns map[string]ListPattern `json:"listPatterns"`
			} `json:"main"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		body, ok := doc.Main[locale]
		if !ok {
			return nil, fmt.Errorf("listPatterns body missing for %s", locale)
		}
		patterns := make(ListPatterns)
		for _, key := range listPatternKeys {
			pattern, ok := body.ListPatterns[key.cldr]
			if !ok {
				continue
			}
			if patterns[key.typ] == nil {
				patterns[key.typ] = make(map[string]ListPattern)
			}
			patterns[key.typ][key.style] = pattern
		}
		if len(patterns) > 0 {
			loaded[locale] = patterns
		}
	}
	out := make(map[string]ListPatterns, len(locales))
	for _, locale := range locales {
		if locale == "und" {
			continue
		}
		if patterns, ok := loaded[locale]; ok {
			out[locale] = patterns
			continue
		}
		if patterns, ok := inheritedListPatterns(locale, loaded); ok {
			out[locale] = patterns
		}
	}
	return out, nil
}

func inheritedListPatterns(locale string, loaded map[string]ListPatterns) (ListPatterns, bool) {
	for parent := parentLocale(locale); parent != ""; parent = parentLocale(parent) {
		if patterns, ok := loaded[parent]; ok {
			return patterns, true
		}
	}
	return nil, false
}

func parentLocale(locale string) string {
	pos := strings.LastIndex(locale, "-")
	if pos < 0 {
		return ""
	}
	return locale[:pos]
}

var listPatternKeys = []struct {
	cldr  string
	typ   string
	style string
}{
	{cldr: "listPattern-type-standard", typ: "conjunction", style: "long"},
	{cldr: "listPattern-type-standard-short", typ: "conjunction", style: "short"},
	{cldr: "listPattern-type-standard-narrow", typ: "conjunction", style: "narrow"},
	{cldr: "listPattern-type-or", typ: "disjunction", style: "long"},
	{cldr: "listPattern-type-or-short", typ: "disjunction", style: "short"},
	{cldr: "listPattern-type-or-narrow", typ: "disjunction", style: "narrow"},
	{cldr: "listPattern-type-unit", typ: "unit", style: "long"},
	{cldr: "listPattern-type-unit-short", typ: "unit", style: "short"},
	{cldr: "listPattern-type-unit-narrow", typ: "unit", style: "narrow"},
}
