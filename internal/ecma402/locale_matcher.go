package ecma402

import "github.com/agentable/go-intl/internal/localematcher"

// LocaleMatcherOption returns the shared ECMA-402 localeMatcher string option
// rule used by constructors.
func LocaleMatcherOption(value string) StringOption {
	return RequiredStringOption("localeMatcher", value, "lookup", "best fit")
}

// LocaleMatcherAlgorithm maps an ECMA-402 localeMatcher option value to the
// internal matcher algorithm.
func LocaleMatcherAlgorithm(value string) (localematcher.Algorithm, bool) {
	switch value {
	case "", "best fit":
		return localematcher.AlgorithmBestFit, true
	case "lookup":
		return localematcher.AlgorithmLookup, true
	default:
		return 0, false
	}
}
