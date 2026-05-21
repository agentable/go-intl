package cldr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func loadCollations(root string) ([]string, error) {
	raw, err := os.ReadFile(filepath.Join(root, "cldr-bcp47", "bcp47", "collation.json"))
	if err != nil {
		return nil, fmt.Errorf("read collation.json: %w", err)
	}
	var doc struct {
		Keyword struct {
			U struct {
				Co map[string]json.RawMessage `json:"co"`
			} `json:"u"`
		} `json:"keyword"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse collation.json: %w", err)
	}
	var collations []string
	for collation, raw := range doc.Keyword.U.Co {
		if strings.HasPrefix(collation, "_") || !supportedCollation(collation) {
			continue
		}
		var data struct {
			Deprecated bool `json:"_deprecated"`
		}
		if err := json.Unmarshal(raw, &data); err != nil || data.Deprecated {
			continue
		}
		collations = append(collations, collation)
	}
	slices.Sort(collations)
	return collations, nil
}

func supportedCollation(collation string) bool {
	switch collation {
	case "ducet", "search", "standard":
		return false
	default:
		return true
	}
}
