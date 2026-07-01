package ecma402

import "github.com/agentable/go-intl/internal/localematcher"

const (
	localeMatcherOptionName = "localeMatcher"
	localeMatcherLookup     = "lookup"
	localeMatcherBestFit    = "best fit"
)

// LocaleMatcherOptionInput returns the localeMatcher rule after a typed Go
// options bag has preserved whether the caller supplied the option.
func LocaleMatcherOptionInput(value string, present bool) StringOption {
	return OptionalStringOptionInput(localeMatcherOptionName, value, present, localeMatcherLookup, localeMatcherBestFit)
}

// LocaleMatcherAlgorithm maps an ECMA-402 localeMatcher option value to the
// internal matcher algorithm.
func LocaleMatcherAlgorithm(value string) (localematcher.Algorithm, bool) {
	switch value {
	case "", localeMatcherBestFit:
		return localematcher.AlgorithmBestFit, true
	case localeMatcherLookup:
		return localematcher.AlgorithmLookup, true
	default:
		return 0, false
	}
}
