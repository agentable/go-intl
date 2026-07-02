package cldr

import (
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
)

type TimeZoneAlias struct {
	Alias     string
	Canonical string
}

func loadTimeZoneAliases(root string) ([]TimeZoneAlias, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "aliases.json")
	raw, err := readRequiredFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Supplemental struct {
			Metadata struct {
				Alias struct {
					ZoneAlias map[string]json.RawMessage `json:"zoneAlias"`
				} `json:"alias"`
			} `json:"metadata"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse aliases.json: %w", err)
	}
	if len(doc.Supplemental.Metadata.Alias.ZoneAlias) == 0 {
		return nil, fmt.Errorf("expected supplemental metadata alias zoneAlias map")
	}
	aliases := []TimeZoneAlias{}
	if err := appendTimeZoneAliases(&aliases, nil, doc.Supplemental.Metadata.Alias.ZoneAlias); err != nil {
		return nil, err
	}
	slices.SortFunc(aliases, func(a, b TimeZoneAlias) int {
		if byAlias := strings.Compare(a.Alias, b.Alias); byAlias != 0 {
			return byAlias
		}
		return strings.Compare(a.Canonical, b.Canonical)
	})
	return aliases, nil
}

func appendTimeZoneAliases(out *[]TimeZoneAlias, prefix []string, nodes map[string]json.RawMessage) error {
	for _, key := range slices.Sorted(maps.Keys(nodes)) {
		current := append(slices.Clone(prefix), key)
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(nodes[key], &fields); err != nil {
			return fmt.Errorf("parse zoneAlias %s: %w", strings.Join(current, "/"), err)
		}
		if replacementRaw, ok := fields["_replacement"]; ok {
			var replacement string
			if err := json.Unmarshal(replacementRaw, &replacement); err != nil {
				return fmt.Errorf("parse zoneAlias %s replacement: %w", strings.Join(current, "/"), err)
			}
			if replacement == "" {
				return fmt.Errorf("zoneAlias %s replacement is empty", strings.Join(current, "/"))
			}
			*out = append(*out, TimeZoneAlias{Alias: strings.Join(current, "/"), Canonical: replacement})
			continue
		}

		children := make(map[string]json.RawMessage, len(fields))
		for name, raw := range fields {
			if strings.HasPrefix(name, "_") {
				continue
			}
			children[name] = raw
		}
		if len(children) == 0 {
			return fmt.Errorf("zoneAlias %s missing replacement or children", strings.Join(current, "/"))
		}
		if err := appendTimeZoneAliases(out, current, children); err != nil {
			return err
		}
	}
	return nil
}
