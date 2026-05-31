package localeid

import (
	"errors"
	"maps"
	"slices"
	"strings"
)

var ErrInvalidUnicodeExtension = errors.New("invalid Unicode locale extension")

type UnicodeKeyword struct {
	Key   string
	Value string
}

type UnicodeExtension struct {
	attributes []string
	keywords   []UnicodeKeyword
	byKey      map[string]string
}

func SplitUnicodeExtension(tag string) (string, UnicodeExtension, error) {
	base, extension, err := RemoveUnicodeExtension(tag)
	if err != nil || extension == "" {
		return base, UnicodeExtension{}, err
	}
	ext, err := ParseUnicodeExtension(extension)
	if err != nil {
		return "", UnicodeExtension{}, err
	}
	return base, ext, nil
}

func RemoveUnicodeExtension(tag string) (base, extension string, err error) {
	parts := strings.Split(tag, "-")
	for i := range parts {
		if parts[i] != "u" {
			continue
		}
		end := i + 1
		for end < len(parts) && len(parts[end]) != 1 {
			end++
		}
		extension = "-" + strings.Join(parts[i:end], "-")
		baseParts := slices.Clone(parts[:i])
		baseParts = append(baseParts, parts[end:]...)
		if len(baseParts) == 0 {
			return "", "", ErrInvalidUnicodeExtension
		}
		return strings.Join(baseParts, "-"), extension, nil
	}
	return tag, "", nil
}

func ParseUnicodeExtension(extension string) (UnicodeExtension, error) {
	body, ok := strings.CutPrefix(extension, "-u-")
	if !ok {
		return UnicodeExtension{}, ErrInvalidUnicodeExtension
	}
	return parseUnicodeExtensionParts(strings.Split(body, "-"))
}

func NewUnicodeExtension(attributes []string, keywords []UnicodeKeyword) UnicodeExtension {
	ext := UnicodeExtension{
		attributes: uniqueSorted(attributes),
		keywords:   make([]UnicodeKeyword, 0, len(keywords)),
		byKey:      make(map[string]string, len(keywords)),
	}
	for _, keyword := range keywords {
		key := strings.ToLower(keyword.Key)
		if key == "" {
			continue
		}
		if _, ok := ext.byKey[key]; ok {
			continue
		}
		value := strings.ToLower(keyword.Value)
		ext.byKey[key] = value
		ext.keywords = append(ext.keywords, UnicodeKeyword{Key: key, Value: value})
	}
	slices.SortFunc(ext.keywords, func(a, b UnicodeKeyword) int {
		return strings.Compare(a.Key, b.Key)
	})
	return ext
}

func (e UnicodeExtension) Attributes() []string {
	return slices.Clone(e.attributes)
}

func (e UnicodeExtension) Keywords() []UnicodeKeyword {
	return slices.Clone(e.keywords)
}

func (e UnicodeExtension) TypeForKey(key string) (string, bool) {
	if e.byKey == nil {
		for _, keyword := range e.keywords {
			if keyword.Key == key {
				return keyword.Value, true
			}
		}
		return "", false
	}
	value, ok := e.byKey[key]
	return value, ok
}

func (e UnicodeExtension) ValueForKey(key string) string {
	value, ok := e.TypeForKey(key)
	if !ok {
		return ""
	}
	if value == "" {
		return "true"
	}
	return value
}

func (e UnicodeExtension) Parts() []string {
	parts := slices.Clone(e.attributes)
	for _, keyword := range e.keywords {
		parts = append(parts, keyword.Key)
		if keyword.Value != "" && keyword.Value != "true" {
			parts = append(parts, strings.Split(keyword.Value, "-")...)
		}
	}
	return parts
}

func InsertUnicodeExtension(locale string, attributes []string, keywords []UnicodeKeyword) string {
	ext := NewUnicodeExtension(attributes, keywords)
	parts := ext.Parts()
	if len(parts) == 0 {
		return locale
	}
	extension := "-u-" + strings.Join(parts, "-")
	privateIndex := strings.Index(locale, "-x-")
	if privateIndex < 0 {
		return locale + extension
	}
	return locale[:privateIndex] + extension + locale[privateIndex:]
}

func parseUnicodeExtensionParts(parts []string) (UnicodeExtension, error) {
	if len(parts) == 0 {
		return UnicodeExtension{}, ErrInvalidUnicodeExtension
	}
	var attributes []string
	seenAttributes := map[string]bool{}
	i := 0
	for i < len(parts) && len(parts[i]) >= 3 {
		if !isUnicodeExtensionSubtag(parts[i]) {
			return UnicodeExtension{}, ErrInvalidUnicodeExtension
		}
		if !seenAttributes[parts[i]] {
			seenAttributes[parts[i]] = true
			attributes = append(attributes, parts[i])
		}
		i++
	}
	slices.Sort(attributes)

	var keywords []UnicodeKeyword
	seenKeywords := map[string]bool{}
	for i < len(parts) {
		key := parts[i]
		if len(key) != 2 || !asciiAlnum(key) {
			return UnicodeExtension{}, ErrInvalidUnicodeExtension
		}
		i++
		start := i
		for i < len(parts) && len(parts[i]) >= 3 {
			if !isUnicodeExtensionSubtag(parts[i]) {
				return UnicodeExtension{}, ErrInvalidUnicodeExtension
			}
			i++
		}
		if seenKeywords[key] {
			continue
		}
		seenKeywords[key] = true
		keywords = append(keywords, UnicodeKeyword{Key: key, Value: strings.Join(parts[start:i], "-")})
	}
	return NewUnicodeExtension(attributes, keywords), nil
}

func uniqueSorted(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		value = strings.ToLower(value)
		if value != "" {
			seen[value] = true
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

func isUnicodeExtensionSubtag(s string) bool {
	return len(s) >= 3 && len(s) <= 8 && asciiAlnum(s)
}

func asciiAlnum(s string) bool {
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			continue
		}
		return false
	}
	return s != ""
}
