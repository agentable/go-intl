package datapin

// RequiredCLDRPackages returns the npm packages consumed by CLDR generation.
// Each call owns its slice; preflight, installed-input checks and manifests share this list.
func RequiredCLDRPackages() []string {
	return []string{
		"cldr-bcp47",
		"cldr-core",
		"cldr-dates-full",
		"cldr-localenames-full",
		"cldr-misc-full",
		"cldr-numbers-full",
		"cldr-units-full",
	}
}
