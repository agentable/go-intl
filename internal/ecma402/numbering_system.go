package ecma402

import "github.com/agentable/go-intl/internal/numbering"

// SimpleNumberingSystems returns ECMA-402 Table 28,
// AvailableCanonicalNumberingSystems, as an independent sorted slice.
func SimpleNumberingSystems() []string {
	return numbering.SimpleNumberingSystems()
}

// LocalizeDigits replaces ASCII decimal digits with the ECMA-402 simple digit
// set for numberingSystem. Unsupported systems are left unchanged.
func LocalizeDigits(s, numberingSystem string) string {
	return numbering.LocalizeDigits(s, numberingSystem)
}
