package ecma402

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

var (
	displayNamesRegionSubtag = regexp.MustCompile(`^([a-zA-Z]{2}|[0-9]{3})$`)
	displayNamesScriptSubtag = regexp.MustCompile(`^[a-zA-Z]{4}$`)
)

var displayNamesDateTimeFields = map[string]struct{}{
	"era":          {},
	"year":         {},
	"quarter":      {},
	"month":        {},
	"weekOfYear":   {},
	"weekday":      {},
	"day":          {},
	"dayPeriod":    {},
	"hour":         {},
	"minute":       {},
	"second":       {},
	"timeZoneName": {},
}

// CanonicalCodeForDisplayNames implements ECMA-402
// CanonicalCodeForDisplayNames for Intl.DisplayNames.prototype.of.
func CanonicalCodeForDisplayNames(typ, code string) (string, error) {
	switch typ {
	case "language":
		tag, err := canonicalDisplayNamesLanguage(code)
		if err != nil {
			return "", invalidDisplayNamesCode(typ, code)
		}
		return tag, nil
	case "region":
		if !displayNamesRegionSubtag.MatchString(code) {
			return "", invalidDisplayNamesCode(typ, code)
		}
		return strings.ToUpper(code), nil
	case "script":
		if !displayNamesScriptSubtag.MatchString(code) {
			return "", invalidDisplayNamesCode(typ, code)
		}
		lower := strings.ToLower(code)
		return strings.ToUpper(lower[:1]) + lower[1:], nil
	case "calendar":
		if !IsWellFormedUnicodeType(code) {
			return "", invalidDisplayNamesCode(typ, code)
		}
		return strings.ToLower(code), nil
	case "dateTimeField":
		if _, ok := displayNamesDateTimeFields[code]; !ok {
			return "", invalidDisplayNamesCode(typ, code)
		}
		return code, nil
	case "currency":
		if !IsWellFormedCurrencyCode(code) {
			return "", invalidDisplayNamesCode(typ, code)
		}
		return strings.ToUpper(code), nil
	default:
		return "", invalidDisplayNamesCode(typ, code)
	}
}

func canonicalDisplayNamesLanguage(code string) (string, error) {
	canonical, ok := canonicalizeUnicodeLanguageID(code)
	if !ok {
		return "", invalidDisplayNamesCode("language", code)
	}
	tag, err := language.Parse(canonical)
	if err == nil {
		return tag.String(), nil
	}
	return canonical, nil
}

func canonicalizeUnicodeLanguageID(code string) (string, bool) {
	if code == "" || strings.Contains(code, "_") {
		return "", false
	}
	parts := strings.Split(code, "-")
	for _, part := range parts {
		if part == "" || !isASCIIAlnum(part) {
			return "", false
		}
	}

	if !isUnicodeLanguageSubtag(parts[0]) {
		return "", false
	}

	out := []string{strings.ToLower(parts[0])}
	index := 1
	if index < len(parts) && len(parts[index]) == 4 && isASCIIAlpha(parts[index]) {
		out = append(out, titlecaseASCII(parts[index]))
		index++
	}
	if index < len(parts) && isDisplayNamesRegionSubtag(parts[index]) {
		out = append(out, strings.ToUpper(parts[index]))
		index++
	}

	seenVariants := map[string]struct{}{}
	for index < len(parts) && isVariantSubtag(parts[index]) {
		variant := strings.ToLower(parts[index])
		if _, ok := seenVariants[variant]; ok {
			return "", false
		}
		seenVariants[variant] = struct{}{}
		out = append(out, variant)
		index++
	}

	if index != len(parts) {
		return "", false
	}

	return strings.Join(out, "-"), true
}

func isUnicodeLanguageSubtag(subtag string) bool {
	length := len(subtag)
	return (length >= 2 && length <= 3 || length >= 5 && length <= 8) && isASCIIAlpha(subtag)
}

func isDisplayNamesRegionSubtag(subtag string) bool {
	return displayNamesRegionSubtag.MatchString(subtag)
}

func isVariantSubtag(subtag string) bool {
	length := len(subtag)
	if length >= 5 && length <= 8 {
		return isASCIIAlnum(subtag)
	}
	return length == 4 && isASCIIDigit(subtag[0]) && isASCIIAlnum(subtag)
}

func titlecaseASCII(value string) string {
	lower := strings.ToLower(value)
	return strings.ToUpper(lower[:1]) + lower[1:]
}

func isASCIIAlpha(value string) bool {
	for i := range len(value) {
		if !isASCIIAlphaByte(value[i]) {
			return false
		}
	}
	return true
}

func isASCIIAlnum(value string) bool {
	for i := range len(value) {
		if !isASCIIAlphaByte(value[i]) && !isASCIIDigit(value[i]) {
			return false
		}
	}
	return true
}

func isASCIIAlphaByte(value byte) bool {
	return value >= 'A' && value <= 'Z' || value >= 'a' && value <= 'z'
}

func isASCIIDigit(value byte) bool {
	return value >= '0' && value <= '9'
}

func invalidDisplayNamesCode(typ, code string) error {
	return fmt.Errorf("ecma402: invalid DisplayNames %s code %q: %w", typ, code, ErrInvalidOption)
}
