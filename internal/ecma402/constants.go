package ecma402

import (
	"slices"
	"strings"
)

// SanctionedUnits lists the namespaced single unit identifiers permitted by
// the ECMA-402 sanctioned single unit table. IsSanctionedSimpleUnitIdentifier
// matches the de-namespaced form (the part after the first hyphen).
var SanctionedUnits = []string{
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

var sanctionedSimpleUnits = func() []string {
	out := make([]string, len(SanctionedUnits))
	for i, u := range SanctionedUnits {
		out[i] = u
		if _, after, ok := strings.Cut(u, "-"); ok {
			out[i] = after
		}
	}
	slices.Sort(out)
	return out
}()

// SanctionedSimpleUnitIdentifiers returns the sorted, de-namespaced unit
// identifiers exposed by Intl.supportedValuesOf("unit").
func SanctionedSimpleUnitIdentifiers() []string {
	return slices.Clone(sanctionedSimpleUnits)
}
