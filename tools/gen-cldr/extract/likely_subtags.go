package extract

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/agentable/go-intl/internal/localeid"
)

type SubtagTriple struct {
	Lang   string
	Script string
	Region string
}

type LikelySubtags struct {
	Maximize map[string]SubtagTriple
	Minimize map[SubtagTriple]string
}

func ExtractLikelySubtags(raw map[string]string) (LikelySubtags, error) {
	max := make(map[string]SubtagTriple, len(raw))
	min := make(map[SubtagTriple]string, len(raw))
	sourceByNormalizedKey := make(map[string]string, len(raw))
	for _, key := range slices.Sorted(maps.Keys(raw)) {
		value := raw[key]
		normalized, err := normalizeLikelySubtagKey(key)
		if err != nil {
			return LikelySubtags{}, fmt.Errorf("likely subtag %q -> %q: %w", key, value, err)
		}
		if previous, ok := sourceByNormalizedKey[normalized]; ok {
			return LikelySubtags{}, fmt.Errorf("likely subtag %q -> %q: normalized key %q conflicts with source key %q", key, value, normalized, previous)
		}
		sourceByNormalizedKey[normalized] = key
		triple, err := parseTriple(value)
		if err != nil {
			return LikelySubtags{}, fmt.Errorf("likely subtag %q -> %q: %w", key, value, err)
		}
		max[normalized] = triple
		if strings.HasPrefix(normalized, "und") {
			continue
		}
		if _, ok := min[triple]; !ok {
			min[triple] = normalized
		}
	}
	return LikelySubtags{Maximize: max, Minimize: min}, nil
}

func normalizeLikelySubtagKey(key string) (string, error) {
	parts := strings.Split(strings.ReplaceAll(key, "_", "-"), "-")
	if len(parts) == 0 || len(parts) > 3 {
		return "", fmt.Errorf("expected one to three language identifier subtags")
	}
	if slices.Contains(parts, "") {
		return "", fmt.Errorf("expected non-empty language identifier subtags")
	}
	lang, ok := localeid.CanonicalUnicodeLanguageSubtag(parts[0])
	if !ok {
		return "", fmt.Errorf("invalid language subtag %q", parts[0])
	}
	if len(parts) == 1 {
		return lang, nil
	}
	if len(parts) == 2 {
		if script, ok := localeid.CanonicalUnicodeScriptSubtag(parts[1]); ok {
			return localeid.Join(lang, script, ""), nil
		}
		if region, ok := localeid.CanonicalUnicodeRegionSubtag(parts[1]); ok {
			return localeid.Join(lang, "", region), nil
		}
		return "", fmt.Errorf("invalid script or region subtag %q", parts[1])
	}
	script, ok := localeid.CanonicalUnicodeScriptSubtag(parts[1])
	if !ok {
		return "", fmt.Errorf("invalid script subtag %q", parts[1])
	}
	region, ok := localeid.CanonicalUnicodeRegionSubtag(parts[2])
	if !ok {
		return "", fmt.Errorf("invalid region subtag %q", parts[2])
	}
	return localeid.Join(lang, script, region), nil
}

func parseTriple(tag string) (SubtagTriple, error) {
	parts := strings.Split(strings.ReplaceAll(tag, "_", "-"), "-")
	if len(parts) != 3 {
		return SubtagTriple{}, fmt.Errorf("expected language-script-region triple")
	}
	if slices.Contains(parts, "") {
		return SubtagTriple{}, fmt.Errorf("expected non-empty language, script, and region subtags")
	}
	lang, ok := localeid.CanonicalUnicodeLanguageSubtag(parts[0])
	if !ok {
		return SubtagTriple{}, fmt.Errorf("invalid language subtag %q", parts[0])
	}
	script, ok := localeid.CanonicalUnicodeScriptSubtag(parts[1])
	if !ok {
		return SubtagTriple{}, fmt.Errorf("invalid script subtag %q", parts[1])
	}
	region, ok := localeid.CanonicalUnicodeRegionSubtag(parts[2])
	if !ok {
		return SubtagTriple{}, fmt.Errorf("invalid region subtag %q", parts[2])
	}
	return SubtagTriple{Lang: lang, Script: script, Region: region}, nil
}
