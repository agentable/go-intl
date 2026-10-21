package localeid

import (
	"errors"
	"maps"
	"slices"
	"strings"
)

type languageAliasRecord struct{ from, to string }

var ErrInvalidLanguageTag = errors.New("invalid Unicode language tag")

// Tag holds a canonical Unicode identity independently of registry coverage.
// Its zero value denotes und. Public callers construct locale.Locale instead.
type Tag string

func (t Tag) String() string {
	if t == "" {
		return "und"
	}
	return string(t)
}

// ParseTag validates syntax before canonicalization can erase duplicate subtags.
func ParseTag(s string, maximize SubtagMaximizer) (Tag, error) {
	parts := strings.Split(asciiLower(s), "-")
	n := unicodeLanguageIDLength(parts)
	if n == 0 {
		return "", ErrInvalidLanguageTag
	}
	base, err := canonicalLanguageID(parts[:n], maximize)
	if err != nil {
		return "", err
	}
	extensions := map[string]string{}
	for i := n; i < len(parts); {
		key := parts[i]
		i++
		if len(key) != 1 || !asciiAlnum(key) {
			return "", ErrInvalidLanguageTag
		}
		if _, exists := extensions[key]; exists {
			return "", ErrInvalidLanguageTag
		}
		start := i
		if key == "x" {
			for ; i < len(parts); i++ {
				if len(parts[i]) < 1 || len(parts[i]) > 8 || !asciiAlnum(parts[i]) {
					return "", ErrInvalidLanguageTag
				}
			}
		} else {
			for i < len(parts) && len(parts[i]) != 1 {
				if len(parts[i]) < 2 || len(parts[i]) > 8 || !asciiAlnum(parts[i]) {
					return "", ErrInvalidLanguageTag
				}
				i++
			}
		}
		if start == i {
			return "", ErrInvalidLanguageTag
		}
		value := strings.Join(parts[start:i], "-")
		switch key {
		case "u":
			ext, err := parseUnicodeExtensionParts(parts[start:i])
			if err != nil {
				return "", err
			}
			value = strings.Join(ext.Parts(), "-")
		case "t":
			value, err = canonicalTransformed(parts[start:i], maximize)
			if err != nil {
				return "", err
			}
		}
		extensions[key] = value
	}
	result := base
	for _, key := range slices.Sorted(maps.Keys(extensions)) {
		if key != "x" {
			result += "-" + key + "-" + extensions[key]
		}
	}
	if value, ok := extensions["x"]; ok {
		result += "-x-" + value
	}
	return Tag(result), nil
}

func canonicalTransformed(parts []string, maximize SubtagMaximizer) (string, error) {
	i := 0
	var out []string
	if IsUnicodeLanguageSubtag(parts[0]) {
		i = unicodeLanguageIDLength(parts)
		if i == 0 {
			return "", ErrInvalidLanguageTag
		}
		base, err := canonicalLanguageID(parts[:i], maximize)
		if err != nil {
			return "", err
		}
		out = append(out, asciiLower(base))
	}
	fields := map[string]string{}
	for i < len(parts) {
		key := parts[i]
		if !IsTransformedKey(key) {
			return "", ErrInvalidLanguageTag
		}
		i++
		start := i
		for i < len(parts) && isUnicodeExtensionSubtag(parts[i]) {
			i++
		}
		if i == start {
			return "", ErrInvalidLanguageTag
		}
		if _, ok := fields[key]; !ok {
			fields[key] = canonicalUnicodeTypeAlias("t:"+key, strings.Join(parts[start:i], "-"))
		}
	}
	for _, key := range slices.Sorted(maps.Keys(fields)) {
		out = append(out, key, fields[key])
	}
	if len(out) == 0 {
		return "", ErrInvalidLanguageTag
	}
	return strings.Join(out, "-"), nil
}

func Parts(tag Tag) (lang, script, region string) {
	parts := strings.Split(tag.String(), "-")
	lang = parts[0]
	i := 1
	if i < len(parts) && IsUnicodeScriptSubtag(parts[i]) {
		script = parts[i]
		i++
	}
	if i < len(parts) && IsUnicodeRegionSubtag(parts[i]) {
		region = parts[i]
	}
	return
}
func Join(lang, script, region string) string {
	return strings.Join(languageSubtags(lang, script, region), "-")
}
func languageSubtags(lang, script, region string) []string {
	parts := []string{lang}
	if script != "" {
		parts = append(parts, script)
	}
	if region != "" {
		parts = append(parts, region)
	}
	return parts
}
func BaseName(tag Tag) string {
	parts := strings.Split(tag.String(), "-")
	for i, p := range parts {
		if len(p) == 1 {
			return strings.Join(parts[:i], "-")
		}
	}
	return tag.String()
}
func VariantSubtags(tag Tag) []string {
	parts := strings.Split(BaseName(tag), "-")
	i := 1
	if i < len(parts) && IsUnicodeScriptSubtag(parts[i]) {
		i++
	}
	if i < len(parts) && IsUnicodeRegionSubtag(parts[i]) {
		i++
	}
	if i == len(parts) {
		return nil
	}
	return parts[i:]
}

// ReplaceLanguageSubtags projects already canonical LSR values, retaining the suffix.
func ReplaceLanguageSubtags(tag Tag, lang, script, region string) Tag {
	base := BaseName(tag)
	return Tag(strings.Join(append(languageSubtags(lang, script, region), VariantSubtags(tag)...), "-") + strings.TrimPrefix(tag.String(), base))
}
func ReplaceLanguageID(tag Tag, lang, script, region string, variants []string, maximize SubtagMaximizer) (Tag, error) {
	base := BaseName(tag)
	suffix := strings.TrimPrefix(tag.String(), base)
	return ParseTag(strings.Join(append(languageSubtags(lang, script, region), variants...), "-")+suffix, maximize)
}

func canonicalLanguageID(parts []string, maximize SubtagMaximizer) (string, error) {
	tag := Tag(strings.Join(parts, "-"))
	lang, script, region := Parts(tag)
	script = asciiTitle(script)
	region = asciiUpper(region)
	variants := VariantSubtags(tag)
	seen := map[string]bool{}
	for {
		before := strings.Join(append(languageSubtags(lang, script, region), variants...), "-")
		if seen[before] {
			return "", ErrInvalidLanguageTag
		}
		seen[before] = true
		// Prefer the most specific matching CLDR language alias, including und rules.
		best := languageAliasRecord{}
		bestScore := -1
		for _, group := range []string{lang, "und"} {
			for _, rule := range languageAliases[group] {
				from := rule.from
				p := strings.Split(from, "-")
				if unicodeLanguageIDLength(p) != len(p) {
					continue
				}
				l, s, r := Parts(Tag(from))
				vs := VariantSubtags(Tag(from))
				if l != "und" && l != lang || s != "" && s != script || r != "" && r != region {
					continue
				}
				matches := true
				for _, v := range vs {
					if !slices.Contains(variants, v) {
						matches = false
						break
					}
				}
				if !matches {
					continue
				}
				// Count specificity first, then language/script/region presence.
				score := len(vs) * 16
				if l != "und" {
					score += 24
				}
				if s != "" {
					score += 20
				}
				if r != "" {
					score += 18
				}
				if score > bestScore || score == bestScore && from < best.from {
					best = rule
					bestScore = score
				}
			}
		}
		if best.from != "" {
			l, s, r := Parts(Tag(best.from))
			rl, rs, rr := Parts(Tag(best.to))
			if l != "und" || lang == "und" {
				lang = rl
			}
			if s != "" || script == "" {
				script = rs
			}
			if r != "" || region == "" {
				region = rr
			}
			remove := VariantSubtags(Tag(best.from))
			variants = slices.DeleteFunc(variants, func(v string) bool { return slices.Contains(remove, v) })
			for _, v := range VariantSubtags(Tag(best.to)) {
				if !slices.Contains(variants, v) {
					variants = append(variants, v)
				}
			}
		} else {
			region = CanonicalRegion(lang, script, region, maximize)
			if alias, ok := scriptAlias[script]; ok {
				script = alias
			}
			for i, v := range variants {
				if alias, ok := variantAlias[v]; ok {
					variants[i] = alias
				}
			}
			slices.Sort(variants)
			variants = slices.Compact(variants)
		}
		after := strings.Join(append(languageSubtags(lang, script, region), variants...), "-")
		if before == after {
			return after, nil
		}
	}
}

// CanonicalRegion retains unregistered, grammatical regions and resolves CLDR aliases.
func CanonicalRegion(lang, script, region string, maximize SubtagMaximizer) string {
	aliases := strings.Fields(territoryAlias[region])
	if len(aliases) == 0 {
		return region
	}
	if len(aliases) > 1 && maximize != nil {
		_, _, likely, ok := maximize(lang, script, "")
		if ok && slices.Contains(aliases, likely) {
			return likely
		}
	}
	return aliases[0]
}
