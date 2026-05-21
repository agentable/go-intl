package ecma402

import (
	"regexp"
	"slices"
	"strings"
)

var unicodeTypePattern = regexp.MustCompile(`^[a-z0-9]{3,8}(-[a-z0-9]{3,8})*$`)

// IsWellFormedUnicodeType checks the BCP 47 Unicode locale extension type
// syntax used by ca/co/nu and similar Intl options.
func IsWellFormedUnicodeType(value string) bool {
	return value != "" && unicodeTypePattern.MatchString(strings.ToLower(value))
}

// IsWellFormedCurrencyCode mirrors ECMA-402 sec-iswellformedcurrencycode.
// A currency is well-formed iff its uppercase form is exactly three ASCII
// letters A-Z. No registry lookup is performed.
func IsWellFormedCurrencyCode(currency string) bool {
	if len(currency) != 3 {
		return false
	}
	upper := strings.ToUpper(currency)
	for i := range len(upper) {
		c := upper[i]
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	return true
}

// IsSanctionedSimpleUnitIdentifier mirrors ECMA-402
// sec-issanctionedsimpleunitidentifier — exact membership check against the
// de-namespaced sanctioned unit list.
func IsSanctionedSimpleUnitIdentifier(unit string) bool {
	_, ok := slices.BinarySearch(sanctionedSimpleUnits, unit)
	return ok
}

// IsWellFormedUnitIdentifier mirrors ECMA-402 sec-iswellformedunitidentifier.
// The identifier is either a sanctioned simple unit, or a "<simple>-per-<simple>"
// composite of two sanctioned simple units.
func IsWellFormedUnitIdentifier(unit string) bool {
	if IsSanctionedSimpleUnitIdentifier(unit) {
		return true
	}
	num, denom, ok := strings.Cut(unit, "-per-")
	if !ok {
		return false
	}
	return IsSanctionedSimpleUnitIdentifier(num) &&
		IsSanctionedSimpleUnitIdentifier(denom)
}
