package plural

import pluralop "github.com/agentable/go-intl/internal/plural"

// ResolveCardinalRange returns the explicit CLDR range category, or other when
// the sparse range data has no row for the category pair.
func ResolveCardinalRange(loc string, start, end pluralop.Category) pluralop.Category {
	if category, ok := CardinalRange(loc, start, end); ok {
		return category
	}
	return pluralop.Other
}
