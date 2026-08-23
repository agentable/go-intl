package cldr

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// UnicodeTypeAlias is one canonical Unicode locale extension type mapping.
// Key is part of the identity: the same type spelling may mean unrelated
// things under different Unicode extension keys.
type UnicodeTypeAlias struct {
	Key       string
	Alias     string
	Canonical string
}

type unicodeTypeRecord struct {
	key        string
	name       string
	aliases    []string
	preferred  string
	deprecated bool
	source     string
}

func loadUnicodeTypeAliases(root string) ([]UnicodeTypeAlias, error) {
	dir := filepath.Join(root, "cldr-bcp47", "bcp47")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read cldr-bcp47/bcp47: %w", err)
	}

	records := make(map[string]map[string]unicodeTypeRecord)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		raw, err := readRequiredFile(path)
		if err != nil {
			return nil, err
		}
		var doc struct {
			Keyword map[string]jsontext.Value `json:"keyword"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		uRaw, ok := doc.Keyword["u"]
		if !ok || string(uRaw) == "null" {
			continue
		}
		var keys map[string]jsontext.Value
		if err := json.Unmarshal(uRaw, &keys); err != nil {
			return nil, fmt.Errorf("parse %s keyword.u: %w", path, err)
		}
		for key, keyRaw := range keys {
			if strings.HasPrefix(key, "_") {
				continue
			}
			if !isUnicodeKey(key) {
				return nil, fmt.Errorf("%s keyword.u: invalid key %q", path, key)
			}
			var types map[string]jsontext.Value
			if err := json.Unmarshal(keyRaw, &types); err != nil {
				return nil, fmt.Errorf("parse %s keyword.u.%s: %w", path, key, err)
			}
			if records[key] == nil {
				records[key] = make(map[string]unicodeTypeRecord)
			}
			for name, typeRaw := range types {
				if strings.HasPrefix(name, "_") {
					continue
				}
				if name != strings.ToLower(name) || !isUnicodeType(name) {
					// Some keys contain schema placeholders such as REORDER_CODE.
					// They describe open value spaces rather than concrete BCP 47
					// types and therefore cannot own runtime aliases.
					continue
				}
				if previous, exists := records[key][name]; exists {
					return nil, fmt.Errorf("%s keyword.u.%s.%s duplicates %s", path, key, name, previous.source)
				}
				var metadata struct {
					Alias      string `json:"_alias"`
					Preferred  string `json:"_preferred"`
					Deprecated bool   `json:"_deprecated"`
				}
				if err := json.Unmarshal(typeRaw, &metadata); err != nil {
					return nil, fmt.Errorf("parse %s keyword.u.%s.%s: %w", path, key, name, err)
				}
				aliases := strings.Fields(strings.ToLower(metadata.Alias))
				records[key][name] = unicodeTypeRecord{
					key:        key,
					name:       name,
					aliases:    aliases,
					preferred:  strings.ToLower(metadata.Preferred),
					deprecated: metadata.Deprecated,
					source:     path,
				}
			}
		}
	}

	var out []UnicodeTypeAlias
	for _, key := range slices.Sorted(maps.Keys(records)) {
		aliases, err := resolveUnicodeTypeAliases(key, records[key])
		if err != nil {
			return nil, err
		}
		out = append(out, aliases...)
	}
	return out, nil
}

func resolveUnicodeTypeAliases(key string, records map[string]unicodeTypeRecord) ([]UnicodeTypeAlias, error) {
	direct := make(map[string]string)
	for _, name := range slices.Sorted(maps.Keys(records)) {
		record := records[name]
		if record.preferred != "" {
			if !isUnicodeType(record.preferred) {
				return nil, fmt.Errorf("%s keyword.u.%s.%s: invalid preferred target %q", record.source, key, name, record.preferred)
			}
			if _, ok := records[record.preferred]; !ok {
				return nil, fmt.Errorf("%s keyword.u.%s.%s: preferred target %q is missing", record.source, key, name, record.preferred)
			}
			if err := addUnicodeTypeAlias(direct, key, name, record.preferred); err != nil {
				return nil, err
			}
			continue
		}
		if record.deprecated {
			continue
		}
		for _, alias := range record.aliases {
			// CLDR also carries legacy identifiers such as IANA zone names in
			// _alias. They are not Unicode extension types and cannot enter this
			// table.
			if !isUnicodeType(alias) {
				continue
			}
			if err := addUnicodeTypeAlias(direct, key, alias, name); err != nil {
				return nil, err
			}
		}
	}

	resolved := make(map[string]string, len(direct))
	for _, alias := range slices.Sorted(maps.Keys(direct)) {
		seen := make(map[string]bool)
		current := alias
		for {
			if seen[current] {
				return nil, fmt.Errorf("keyword.u.%s: Unicode type alias cycle at %q", key, current)
			}
			seen[current] = true
			next, ok := direct[current]
			if !ok {
				break
			}
			current = next
		}
		if record, canonicalRecord := records[alias]; canonicalRecord && !record.deprecated && alias != current {
			return nil, fmt.Errorf("keyword.u.%s: alias %q conflicts with canonical type record", key, alias)
		}
		resolved[alias] = current
	}

	out := make([]UnicodeTypeAlias, 0, len(resolved))
	for _, alias := range slices.Sorted(maps.Keys(resolved)) {
		canonical := resolved[alias]
		if alias != canonical {
			out = append(out, UnicodeTypeAlias{Key: key, Alias: alias, Canonical: canonical})
		}
	}
	return out, nil
}

func addUnicodeTypeAlias(aliases map[string]string, key, alias, canonical string) error {
	if alias == canonical {
		return nil
	}
	if previous, ok := aliases[alias]; ok && previous != canonical {
		return fmt.Errorf("keyword.u.%s: conflicting alias %q maps to %q and %q", key, alias, previous, canonical)
	}
	aliases[alias] = canonical
	return nil
}

func isUnicodeKey(value string) bool {
	return len(value) == 2 && asciiAlnumString(value)
}

func isUnicodeType(value string) bool {
	if value == "" {
		return false
	}
	for subtag := range strings.SplitSeq(value, "-") {
		if len(subtag) < 3 || len(subtag) > 8 || !asciiAlnumString(subtag) {
			return false
		}
	}
	return true
}

func asciiAlnumString(value string) bool {
	for i := range len(value) {
		c := value[i]
		if c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' {
			continue
		}
		return false
	}
	return true
}
