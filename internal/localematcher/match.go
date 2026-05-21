package localematcher

type Algorithm int
type Maximizer func(tag string) string

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
	return MatchWithMaximizer(requested, supported, defaultLocale, alg, nil)
}

func MatchWithMaximizer(requested, supported []string, defaultLocale string, alg Algorithm, maximizer Maximizer) Result {
	return NewMatcher(supported, maximizer).Match(requested, defaultLocale, alg)
}

func normalizeMaximizer(maximizer Maximizer) Maximizer {
	if maximizer != nil {
		return maximizer
	}
	return identityMaximizer
}

func identityMaximizer(tag string) string {
	return tag
}
