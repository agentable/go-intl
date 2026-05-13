package localematcher

type Algorithm int

const (
	AlgorithmLookup Algorithm = iota
	AlgorithmBestFit
)

const DefaultMatchingThreshold = 838

type Result struct {
	Locale     string
	DataLocale string
	Extension  string
	Distance   int
}

func Match(requested, supported []string, defaultLocale string, alg Algorithm) Result {
	if alg == AlgorithmLookup {
		return LookupMatcher(requested, supported, defaultLocale)
	}
	return BestFitMatcher(requested, supported, defaultLocale)
}
