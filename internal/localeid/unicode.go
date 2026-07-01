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
		if parts[i] == "x" {
			break
		}
		if parts[i] != "u" {
			continue
		}
		end := i + 1
		for end < len(parts) && len(parts[end]) != 1 {
			end++
		}
		baseLen := len(parts) - (end - i)
		if baseLen == 0 {
			return "", "", ErrInvalidUnicodeExtension
		}
		extension = "-" + strings.Join(parts[i:end], "-")
		baseParts := make([]string, baseLen)
		n := copy(baseParts, parts[:i])
		copy(baseParts[n:], parts[end:])
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
		key := asciiLower(keyword.Key)
		if key == "" {
			continue
		}
		if _, ok := ext.byKey[key]; ok {
			continue
		}
		value := asciiLower(keyword.Value)
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

// LowercaseUnicodeLocaleID lowercases ASCII letters in a Unicode locale
// identifier without applying Unicode case folding.
func LowercaseUnicodeLocaleID(tag string) string {
	return asciiLower(tag)
}

// IsUnicodeType reports whether value has BCP 47 Unicode locale extension type
// syntax: one or more 3-8 byte ASCII alphanumeric subtags.
func IsUnicodeType(value string) bool {
	if value == "" {
		return false
	}
	value = asciiLower(value)
	for {
		subtag, rest, ok := strings.Cut(value, "-")
		if !isUnicodeExtensionSubtag(subtag) {
			return false
		}
		if !ok {
			return true
		}
		value = rest
	}
}

// CanonicalUnicodeType lowercases a Unicode locale extension type and applies
// the aliases accepted by Intl.Locale before validating the canonical value.
func CanonicalUnicodeType(value string) (string, bool) {
	value = canonicalUnicodeTypeAlias(asciiLower(value))
	if !IsUnicodeType(value) {
		return "", false
	}
	return value, true
}

// IsUnicodeLanguageSubtag reports whether subtag has Unicode language subtag
// syntax: 2-3 or 5-8 ASCII letters.
func IsUnicodeLanguageSubtag(subtag string) bool {
	length := len(subtag)
	return (length >= 2 && length <= 3 || length >= 5 && length <= 8) && asciiAlpha(subtag)
}

// CanonicalUnicodeLanguageSubtag validates and lowercases a Unicode language
// subtag.
func CanonicalUnicodeLanguageSubtag(subtag string) (string, bool) {
	if !IsUnicodeLanguageSubtag(subtag) {
		return "", false
	}
	return asciiLower(subtag), true
}

// IsUnicodeScriptSubtag reports whether subtag has Unicode script subtag
// syntax: exactly four ASCII letters.
func IsUnicodeScriptSubtag(subtag string) bool {
	return len(subtag) == 4 && asciiAlpha(subtag)
}

// CanonicalUnicodeScriptSubtag validates and title-cases a Unicode script
// subtag.
func CanonicalUnicodeScriptSubtag(subtag string) (string, bool) {
	if !IsUnicodeScriptSubtag(subtag) {
		return "", false
	}
	return asciiTitle(subtag), true
}

// IsUnicodeRegionSubtag reports whether subtag has Unicode region subtag
// syntax: two ASCII letters or three ASCII digits.
func IsUnicodeRegionSubtag(subtag string) bool {
	switch len(subtag) {
	case 2:
		return asciiAlpha(subtag)
	case 3:
		return asciiDigits(subtag)
	default:
		return false
	}
}

// CanonicalUnicodeRegionSubtag validates and uppercases a Unicode region
// subtag.
func CanonicalUnicodeRegionSubtag(subtag string) (string, bool) {
	if !IsUnicodeRegionSubtag(subtag) {
		return "", false
	}
	return asciiUpper(subtag), true
}

// IsUnicodeVariantSubtag reports whether subtag has Unicode variant subtag
// syntax: 5-8 ASCII alphanumerics, or digit plus three ASCII alphanumerics.
func IsUnicodeVariantSubtag(subtag string) bool {
	length := len(subtag)
	if length >= 5 && length <= 8 {
		return asciiAlnum(subtag)
	}
	return length == 4 && subtag[0] >= '0' && subtag[0] <= '9' && asciiAlnum(subtag)
}

// CanonicalUnicodeVariantSubtag validates and lowercases a Unicode variant
// subtag.
func CanonicalUnicodeVariantSubtag(subtag string) (string, bool) {
	if !IsUnicodeVariantSubtag(subtag) {
		return "", false
	}
	return asciiLower(subtag), true
}

// RelevantExtensionValues builds the ordered candidate values for one Unicode
// extension key: default first, then non-empty supported values without
// duplicates.
func RelevantExtensionValues(defaultValue string, values ...string) []string {
	out := make([]string, 0, len(values)+1)
	if defaultValue != "" {
		out = append(out, defaultValue)
	}
	for _, value := range values {
		if value != "" && !slices.Contains(out, value) {
			out = append(out, value)
		}
	}
	return out
}

func parseUnicodeExtensionParts(parts []string) (UnicodeExtension, error) {
	if len(parts) == 0 {
		return UnicodeExtension{}, ErrInvalidUnicodeExtension
	}
	var attributes []string
	i := 0
	for i < len(parts) && len(parts[i]) >= 3 {
		if !isUnicodeExtensionSubtag(parts[i]) {
			return UnicodeExtension{}, ErrInvalidUnicodeExtension
		}
		attributes = append(attributes, parts[i])
		i++
	}

	var keywords []UnicodeKeyword
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
		value = asciiLower(value)
		if value != "" {
			seen[value] = true
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

func canonicalUnicodeTypeAlias(value string) string {
	switch value {
	case "gregorian":
		return "gregory"
	case "islamic-civil":
		return "islamicc"
	}
	return value
}

func isUnicodeExtensionSubtag(s string) bool {
	return len(s) >= 3 && len(s) <= 8 && asciiAlnum(s)
}

func asciiLower(s string) string {
	for i := range len(s) {
		if s[i] >= 'A' && s[i] <= 'Z' {
			return asciiLowerFrom(s, i)
		}
	}
	return s
}

func asciiLowerFrom(s string, start int) string {
	out := []byte(s)
	for i := start; i < len(out); i++ {
		if out[i] >= 'A' && out[i] <= 'Z' {
			out[i] += 'a' - 'A'
		}
	}
	return string(out)
}

func asciiUpper(s string) string {
	for i := range len(s) {
		if s[i] >= 'a' && s[i] <= 'z' {
			return asciiUpperFrom(s, i)
		}
	}
	return s
}

func asciiUpperFrom(s string, start int) string {
	out := []byte(s)
	for i := start; i < len(out); i++ {
		if out[i] >= 'a' && out[i] <= 'z' {
			out[i] -= 'a' - 'A'
		}
	}
	return string(out)
}

func asciiTitle(s string) string {
	for i := range len(s) {
		canonical := s[i]
		if i == 0 && s[i] >= 'a' && s[i] <= 'z' {
			canonical -= 'a' - 'A'
		}
		if i > 0 && s[i] >= 'A' && s[i] <= 'Z' {
			canonical += 'a' - 'A'
		}
		if canonical != s[i] {
			return asciiTitleFrom(s, i)
		}
	}
	return s
}

func asciiTitleFrom(s string, start int) string {
	out := []byte(s)
	for i := start; i < len(out); i++ {
		if i == 0 && out[i] >= 'a' && out[i] <= 'z' {
			out[i] -= 'a' - 'A'
		}
		if i > 0 && out[i] >= 'A' && out[i] <= 'Z' {
			out[i] += 'a' - 'A'
		}
	}
	return string(out)
}

func asciiAlnum(s string) bool {
	for i := range len(s) {
		c := s[i] | 0x20
		if c >= 'a' && c <= 'z' || s[i] >= '0' && s[i] <= '9' {
			continue
		}
		return false
	}
	return s != ""
}

func asciiAlpha(s string) bool {
	for i := range len(s) {
		c := s[i] | 0x20
		if c < 'a' || c > 'z' {
			return false
		}
	}
	return s != ""
}

func asciiDigits(s string) bool {
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return s != ""
}
