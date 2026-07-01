package ecma402

import (
	"slices"
	"strings"

	"github.com/agentable/go-intl/internal/localeid"

	"golang.org/x/text/language"
)

var (
	displayNamesDateTimeFieldValues = [...]string{
		"era",
		"year",
		"quarter",
		"month",
		"weekOfYear",
		"weekday",
		"day",
		"dayPeriod",
		"hour",
		"minute",
		"second",
		"timeZoneName",
	}
	displayNamesDateTimeFieldExpected = "one of " + quotedValues(displayNamesDateTimeFieldValues[:])
)

type displayNamesCodeRule struct {
	canonicalize func(string) (string, bool)
	expected     string
}

// CanonicalCodeForDisplayNames implements ECMA-402
// CanonicalCodeForDisplayNames for Intl.DisplayNames.prototype.of.
func CanonicalCodeForDisplayNames(typ, code string) (string, error) {
	rule, ok := displayNamesCodeRuleFor(typ)
	if !ok {
		return "", ErrInvalidCode
	}
	canonical, ok := rule.canonicalize(code)
	if !ok {
		return "", ErrInvalidCode
	}
	return canonical, nil
}

func displayNamesCodeRuleFor(typ string) (displayNamesCodeRule, bool) {
	switch typ {
	case "language":
		return displayNamesCodeRule{
			canonicalize: canonicalDisplayNamesLanguage,
			expected:     "a Unicode language identifier",
		}, true
	case "region":
		return displayNamesCodeRule{
			canonicalize: canonicalDisplayNamesRegion,
			expected:     "a two-letter ASCII or three-digit region code",
		}, true
	case "script":
		return displayNamesCodeRule{
			canonicalize: canonicalDisplayNamesScript,
			expected:     "a four-letter ASCII script code",
		}, true
	case "calendar":
		return displayNamesCodeRule{
			canonicalize: canonicalDisplayNamesCalendar,
			expected:     "a Unicode locale extension type",
		}, true
	case "dateTimeField":
		return displayNamesCodeRule{
			canonicalize: canonicalDisplayNamesDateTimeField,
			expected:     displayNamesDateTimeFieldExpected,
		}, true
	case "currency":
		return displayNamesCodeRule{
			canonicalize: canonicalDisplayNamesCurrency,
			expected:     "a three-letter ASCII currency code",
		}, true
	default:
		return displayNamesCodeRule{}, false
	}
}

// InvalidDisplayNamesCodeError records an invalid DisplayNames code with
// type-specific guidance from CanonicalCodeForDisplayNames.
func InvalidDisplayNamesCodeError(owner, typ, code, loc string, err error) error {
	return InvalidCodeErrorExpected(owner, typ, code, loc, DisplayNamesCodeExpected(typ), err)
}

// DisplayNamesCodeExpected returns user-facing guidance for DisplayNames code
// validation. It shares the same type cases as CanonicalCodeForDisplayNames.
func DisplayNamesCodeExpected(typ string) string {
	if rule, ok := displayNamesCodeRuleFor(typ); ok {
		return rule.expected
	}
	return "a well-formed DisplayNames code"
}

func canonicalDisplayNamesLanguage(code string) (string, bool) {
	canonical, ok := canonicalizeUnicodeLanguageID(code)
	if !ok {
		return "", false
	}
	tag, err := language.Parse(canonical)
	if err == nil {
		return tag.String(), true
	}
	return canonical, true
}

func canonicalDisplayNamesRegion(code string) (string, bool) {
	if !localeid.IsUnicodeRegionSubtag(code) {
		return "", false
	}
	return asciiUpper(code), true
}

func canonicalDisplayNamesScript(code string) (string, bool) {
	if !localeid.IsUnicodeScriptSubtag(code) {
		return "", false
	}
	return asciiTitle(code), true
}

func canonicalDisplayNamesCalendar(code string) (string, bool) {
	if !IsWellFormedUnicodeType(code) {
		return "", false
	}
	return asciiLower(code), true
}

func canonicalDisplayNamesDateTimeField(code string) (string, bool) {
	if slices.Contains(displayNamesDateTimeFieldValues[:], code) {
		return code, true
	}
	return "", false
}

func canonicalDisplayNamesCurrency(code string) (string, bool) {
	if !IsWellFormedCurrencyCode(code) {
		return "", false
	}
	return CanonicalCurrencyCode(code), true
}

func canonicalizeUnicodeLanguageID(code string) (string, bool) {
	if code == "" || strings.Contains(code, "_") {
		return "", false
	}
	parts := strings.Split(code, "-")

	if !localeid.IsUnicodeLanguageSubtag(parts[0]) {
		return "", false
	}

	out := []string{asciiLower(parts[0])}
	index := 1
	if index < len(parts) && localeid.IsUnicodeScriptSubtag(parts[index]) {
		out = append(out, asciiTitle(parts[index]))
		index++
	}
	if index < len(parts) && localeid.IsUnicodeRegionSubtag(parts[index]) {
		out = append(out, asciiUpper(parts[index]))
		index++
	}

	var seenVariants []string
	for index < len(parts) && localeid.IsUnicodeVariantSubtag(parts[index]) {
		variant := asciiLower(parts[index])
		if slices.Contains(seenVariants, variant) {
			return "", false
		}
		seenVariants = append(seenVariants, variant)
		out = append(out, variant)
		index++
	}

	if index != len(parts) {
		return "", false
	}

	return strings.Join(out, "-"), true
}
