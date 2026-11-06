package cldr

import (
	"encoding/json/v2"
	"fmt"
	"path/filepath"
)

// IdentityAliases contains CLDR replacements for Unicode language identifiers.
type IdentityAliases map[string]map[string]string

func loadIdentityAliases(root string) (IdentityAliases, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "aliases.json")
	raw, err := readRequiredFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Supplemental struct {
			Metadata struct {
				Alias map[string]map[string]struct {
					Replacement string `json:"_replacement"`
				} `json:"alias"`
			} `json:"metadata"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	result := IdentityAliases{}
	for _, kind := range []string{"languageAlias", "scriptAlias", "territoryAlias", "variantAlias", "subdivisionAlias"} {
		rows := doc.Supplemental.Metadata.Alias[kind]
		if len(rows) == 0 {
			return nil, fmt.Errorf("%s: missing %s", path, kind)
		}
		result[kind] = map[string]string{}
		for from, row := range rows {
			if row.Replacement == "" {
				return nil, fmt.Errorf("%s: empty replacement for %s %s", path, kind, from)
			}
			result[kind][from] = row.Replacement
		}
	}
	return result, nil
}
