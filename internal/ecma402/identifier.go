package ecma402

import (
	"github.com/agentable/go-intl/internal/localeid"
	"github.com/agentable/go-intl/internal/unitid"
)

// IsWellFormedUnicodeType checks the BCP 47 Unicode locale extension type
// syntax used by ca/co/nu and similar Intl options.
func IsWellFormedUnicodeType(value string) bool {
	return localeid.IsUnicodeType(value)
}

// CanonicalUnicodeType returns the ASCII-lowercase canonical form for a Unicode
// locale extension type when the value is well-formed.
func CanonicalUnicodeType(value string) (string, bool) {
	canonical := asciiLower(value)
	if !IsWellFormedUnicodeType(canonical) {
		return "", false
	}
	return canonical, true
}

// ApplyUnicodeTypeOptionInput copies a present Unicode type option, canonicalizes
// well-formed values, and records that the caller supplied it explicitly.
func ApplyUnicodeTypeOptionInput(dst *string, present *bool, value *string) {
	if value == nil {
		return
	}
	*present = true
	if canonical, ok := CanonicalUnicodeType(*value); ok {
		*dst = canonical
		return
	}
	*dst = *value
}

// InvalidUnicodeTypeOptionError returns an invalid-option error for options
// whose value must use Unicode locale extension type syntax.
func InvalidUnicodeTypeOptionError(owner, name, value, loc string) error {
	return InvalidOptionErrorExpected(owner, name, value, loc, "a Unicode locale extension type", nil)
}

// ValidateUnicodeTypeOption validates an optional option whose non-empty value
// must use Unicode locale extension type syntax.
func ValidateUnicodeTypeOption(owner, name, value, loc string) error {
	if value == "" || IsWellFormedUnicodeType(value) {
		return nil
	}
	return InvalidUnicodeTypeOptionError(owner, name, value, loc)
}

// ValidateUnicodeTypeOptionInput validates a Unicode locale extension type
// option while preserving the ECMA-402 distinction between omitted input and an
// explicitly present empty string.
func ValidateUnicodeTypeOptionInput(owner, name, value, loc string, present bool) error {
	if !present {
		return ValidateUnicodeTypeOption(owner, name, value, loc)
	}
	if IsWellFormedUnicodeType(value) {
		return nil
	}
	return InvalidUnicodeTypeOptionError(owner, name, value, loc)
}

// IsWellFormedCurrencyCode mirrors ECMA-402 sec-iswellformedcurrencycode.
// A currency is well-formed iff its uppercase form is exactly three ASCII
// letters A-Z. No registry lookup is performed.
func IsWellFormedCurrencyCode(currency string) bool {
	return len(currency) == 3 && isASCIIAlpha(currency)
}

// CanonicalCurrencyCode returns the ECMA-402 ASCII-uppercase currency code
// canonical form. Callers still validate with IsWellFormedCurrencyCode.
func CanonicalCurrencyCode(currency string) string {
	return asciiUpper(currency)
}

// ApplyCurrencyCodeOptionInput copies a present currency code option,
// canonicalizes its ASCII case, and records that the caller supplied it
// explicitly.
func ApplyCurrencyCodeOptionInput(dst *string, present *bool, value *string) {
	if value == nil {
		return
	}
	*dst = CanonicalCurrencyCode(*value)
	*present = true
}

// InvalidCurrencyCodeOptionError returns an invalid-option error for options
// whose value must use ECMA-402 currency-code syntax.
func InvalidCurrencyCodeOptionError(owner, name, value, loc string) error {
	return InvalidOptionErrorExpected(owner, name, value, loc, "a three-letter ASCII currency code", nil)
}

// IsSanctionedSimpleUnitIdentifier mirrors ECMA-402
// sec-issanctionedsimpleunitidentifier — exact membership check against the
// de-namespaced sanctioned unit list.
func IsSanctionedSimpleUnitIdentifier(unit string) bool {
	return unitid.IsSanctionedSimpleUnitIdentifier(unit)
}

// IsWellFormedUnitIdentifier mirrors ECMA-402 sec-iswellformedunitidentifier.
// The identifier is either a sanctioned simple unit, or a "<simple>-per-<simple>"
// composite of two sanctioned simple units.
func IsWellFormedUnitIdentifier(unit string) bool {
	return unitid.IsWellFormedUnitIdentifier(unit)
}

// InvalidUnitIdentifierOptionError returns an invalid-option error for options
// whose value must use ECMA-402 unit-identifier syntax.
func InvalidUnitIdentifierOptionError(owner, name, value, loc string) error {
	return InvalidOptionErrorExpected(owner, name, value, loc, "a sanctioned unit identifier or <unit>-per-<unit> compound", nil)
}
