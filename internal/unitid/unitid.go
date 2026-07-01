// Package unitid owns ECMA-402 measurement unit identifier facts.
package unitid

import (
	"slices"
	"strings"
)

var sanctionedUnits = [...]string{
	"angle-degree",
	"area-acre",
	"area-hectare",
	"concentr-percent",
	"digital-bit",
	"digital-byte",
	"digital-gigabit",
	"digital-gigabyte",
	"digital-kilobit",
	"digital-kilobyte",
	"digital-megabit",
	"digital-megabyte",
	"digital-petabyte",
	"digital-terabit",
	"digital-terabyte",
	"duration-day",
	"duration-hour",
	"duration-microsecond",
	"duration-millisecond",
	"duration-minute",
	"duration-month",
	"duration-nanosecond",
	"duration-second",
	"duration-week",
	"duration-year",
	"length-centimeter",
	"length-foot",
	"length-inch",
	"length-kilometer",
	"length-meter",
	"length-mile-scandinavian",
	"length-mile",
	"length-millimeter",
	"length-yard",
	"mass-gram",
	"mass-kilogram",
	"mass-ounce",
	"mass-pound",
	"mass-stone",
	"temperature-celsius",
	"temperature-fahrenheit",
	"volume-fluid-ounce",
	"volume-gallon",
	"volume-liter",
	"volume-milliliter",
}

var sanctionedSimpleUnits = [...]string{
	"acre",
	"bit",
	"byte",
	"celsius",
	"centimeter",
	"day",
	"degree",
	"fahrenheit",
	"fluid-ounce",
	"foot",
	"gallon",
	"gigabit",
	"gigabyte",
	"gram",
	"hectare",
	"hour",
	"inch",
	"kilobit",
	"kilobyte",
	"kilogram",
	"kilometer",
	"liter",
	"megabit",
	"megabyte",
	"meter",
	"microsecond",
	"mile",
	"mile-scandinavian",
	"milliliter",
	"millimeter",
	"millisecond",
	"minute",
	"month",
	"nanosecond",
	"ounce",
	"percent",
	"petabyte",
	"pound",
	"second",
	"stone",
	"terabit",
	"terabyte",
	"week",
	"yard",
	"year",
}

// SanctionedUnitIdentifiers returns the namespaced single unit identifiers
// permitted by the ECMA-402 sanctioned single unit table.
func SanctionedUnitIdentifiers() []string {
	return slices.Clone(sanctionedUnits[:])
}

// SanctionedSimpleUnitIdentifiers returns the sorted, de-namespaced unit
// identifiers exposed by Intl.supportedValuesOf("unit").
func SanctionedSimpleUnitIdentifiers() []string {
	return slices.Clone(sanctionedSimpleUnits[:])
}

// IsSanctionedSimpleUnitIdentifier reports whether unit is a sanctioned single
// unit identifier in de-namespaced ECMA-402 form.
func IsSanctionedSimpleUnitIdentifier(unit string) bool {
	_, ok := slices.BinarySearch(sanctionedSimpleUnits[:], unit)
	return ok
}

// IsWellFormedUnitIdentifier reports whether unit is a sanctioned simple unit
// or a "<simple>-per-<simple>" compound of sanctioned simple units.
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
