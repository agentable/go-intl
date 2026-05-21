package localematcher

import "strings"

func BestFitMatcher(requested, supported []string, defaultLocale string) Result {
	return BestFitMatcherWithMaximizer(requested, supported, defaultLocale, nil)
}

func BestFitMatcherWithMaximizer(requested, supported []string, defaultLocale string, maximizer Maximizer) Result {
	return NewMatcher(supported, maximizer).bestFit(requested, defaultLocale)
}

type bestMatchResult struct {
	matchedDesired   string
	matchedSupported string
	distance         int
}

func getFallbackCandidates(loc string) []string {
	candidates := make([]string, 0, strings.Count(loc, "-")+1)
	for loc != "" {
		candidates = append(candidates, loc)
		lastDash := strings.LastIndex(loc, "-")
		if lastDash < 0 {
			break
		}
		loc = loc[:lastDash]
	}
	return candidates
}
