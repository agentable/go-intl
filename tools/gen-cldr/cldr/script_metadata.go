package cldr

import (
	"encoding/json/v2"
	"fmt"
	"path/filepath"
)

func loadScriptDirections(root string) (map[string]bool, error) {
	path := filepath.Join(root, "cldr-core", "scriptMetadata.json")
	raw, err := readRequiredFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		ScriptMetadata map[string]struct {
			RTL *string `json:"rtl"`
		} `json:"scriptMetadata"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(doc.ScriptMetadata) == 0 {
		return nil, fmt.Errorf("%s: missing scriptMetadata", path)
	}
	out := make(map[string]bool, len(doc.ScriptMetadata))
	for script, metadata := range doc.ScriptMetadata {
		if !validCanonicalScript(script) {
			return nil, fmt.Errorf("%s: invalid script code %q", path, script)
		}
		if metadata.RTL == nil {
			continue
		}
		switch *metadata.RTL {
		case "YES":
			out[script] = true
		case "NO":
			out[script] = false
		case "UNKNOWN":
		default:
			return nil, fmt.Errorf("%s: script %q has invalid rtl value %q", path, script, *metadata.RTL)
		}
	}
	return out, nil
}

func validCanonicalScript(script string) bool {
	if !validScript(script) || script[0] < 'A' || script[0] > 'Z' {
		return false
	}
	for i := 1; i < len(script); i++ {
		if script[i] < 'a' || script[i] > 'z' {
			return false
		}
	}
	return true
}
