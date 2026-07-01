package ecma402

import "github.com/agentable/go-intl/internal/unitid"

// SanctionedUnitIdentifiers returns the namespaced single unit identifiers
// permitted by the ECMA-402 sanctioned single unit table.
// IsSanctionedSimpleUnitIdentifier matches the de-namespaced form (the part
// after the first hyphen).
func SanctionedUnitIdentifiers() []string {
	return unitid.SanctionedUnitIdentifiers()
}

// SanctionedSimpleUnitIdentifiers returns the sorted, de-namespaced unit
// identifiers exposed by Intl.supportedValuesOf("unit").
func SanctionedSimpleUnitIdentifiers() []string {
	return unitid.SanctionedSimpleUnitIdentifiers()
}
