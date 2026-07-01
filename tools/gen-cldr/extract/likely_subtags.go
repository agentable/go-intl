package extract

import (
	"maps"
	"slices"
	"strings"
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

func ExtractLikelySubtags(raw map[string]string) LikelySubtags {
	max := make(map[string]SubtagTriple, len(raw))
	min := make(map[SubtagTriple]string, len(raw))
	for _, key := range slices.Sorted(maps.Keys(raw)) {
		triple := parseTriple(raw[key])
		normalized := strings.ReplaceAll(key, "_", "-")
		max[normalized] = triple
		if strings.HasPrefix(normalized, "und") {
			continue
		}
		if _, ok := min[triple]; !ok {
			min[triple] = normalized
		}
	}
	return LikelySubtags{Maximize: max, Minimize: min}
}

func parseTriple(tag string) SubtagTriple {
	parts := strings.FieldsFunc(tag, func(r rune) bool { return r == '-' || r == '_' })
	triple := SubtagTriple{}
	if len(parts) > 0 {
		triple.Lang = parts[0]
	}
	if len(parts) > 1 {
		triple.Script = parts[1]
	}
	if len(parts) > 2 {
		triple.Region = parts[2]
	}
	return triple
}
