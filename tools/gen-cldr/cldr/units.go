package cldr

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Units map[string]UnitData

type UnitData struct {
	Patterns map[string]map[string]map[string]string
	Compound map[string]string
}

func loadUnits(root string, locales []string) (map[string]Units, error) {
	out := make(map[string]Units)
	for _, locale := range locales {
		if locale == "und" {
			continue
		}
		path := filepath.Join(root, "cldr-units-full", "main", locale, "units.json")
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var doc struct {
			Main map[string]struct {
				Units map[string]json.RawMessage `json:"units"`
			} `json:"main"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		body, ok := doc.Main[locale]
		if !ok {
			return nil, fmt.Errorf("units body missing for %s", locale)
		}
		units := make(Units)
		compound := make(map[string]string)
		for _, width := range []string{"long", "short", "narrow"} {
			var widthUnits map[string]map[string]string
			if err := json.Unmarshal(body.Units[width], &widthUnits); err != nil {
				return nil, fmt.Errorf("parse %s %s units: %w", path, width, err)
			}
			for _, key := range slices.Sorted(maps.Keys(widthUnits)) {
				fields := widthUnits[key]
				if key == "per" {
					if pattern := fields["compoundUnitPattern"]; pattern != "" {
						compound[width] = pattern
					}
					continue
				}
				unit := strings.TrimPrefix(key, "length-")
				data := units[unit]
				if data.Patterns == nil {
					data.Patterns = make(map[string]map[string]map[string]string)
				}
				if data.Patterns[width] == nil {
					data.Patterns[width] = make(map[string]map[string]string)
				}
				patterns := make(map[string]string)
				for _, field := range slices.Sorted(maps.Keys(fields)) {
					value := fields[field]
					if plural, ok := strings.CutPrefix(field, "unitPattern-count-"); ok {
						patterns[plural] = value
					}
				}
				if len(patterns) > 0 {
					data.Patterns[width][unit] = patterns
					units[unit] = data
				}
			}
		}
		for _, unit := range slices.Sorted(maps.Keys(units)) {
			data := units[unit]
			data.Compound = compound
			units[unit] = data
		}
		out[locale] = units
	}
	return out, nil
}
