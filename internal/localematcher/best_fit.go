package localematcher

import (
	"math"
	"strings"

	"github.com/agentable/go-intl/locale"
)

func BestFitMatcher(requested, supported []string, defaultLocale string) Result {
	requestedExtensions := map[string]string{}
	noExtensionRequested := make([]string, len(requested))
	for i, loc := range requested {
		noExtensionLocale, extension := removeUnicodeExtension(loc)
		noExtensionRequested[i] = noExtensionLocale
		if extension != "" {
			requestedExtensions[noExtensionLocale] = extension
		}
	}
	result := findBestMatch(noExtensionRequested, supported, DefaultMatchingThreshold)
	if result.matchedSupported == "" {
		return Result{Locale: defaultLocale, DataLocale: defaultLocale}
	}
	noExtensionLocale, supportedExtension := removeUnicodeExtension(result.matchedSupported)
	extension := supportedExtension
	if requestedExtension := requestedExtensions[result.matchedDesired]; requestedExtension != "" {
		extension = requestedExtension
	}
	return Result{Locale: noExtensionLocale, DataLocale: noExtensionLocale, Extension: extension, Distance: result.distance}
}

type bestMatchResult struct {
	matchedDesired   string
	matchedSupported string
	distance         int
}

func findBestMatch(requested, supported []string, threshold int) bestMatchResult {
	lowestDistance := math.MaxInt
	result := bestMatchResult{}
	supportedSet := make(map[string]string, len(supported))
	for _, loc := range supported {
		noExtensionLocale, _ := removeUnicodeExtension(loc)
		supportedSet[noExtensionLocale] = loc
	}

	for i, desired := range requested {
		if original, ok := supportedSet[desired]; ok {
			distance := i * 40
			if distance < lowestDistance {
				lowestDistance = distance
				result = bestMatchResult{matchedDesired: desired, matchedSupported: original, distance: distance}
			}
			if i == 0 {
				return result
			}
		}
	}

	for i, desired := range requested {
		maximized := maximize(desired)
		if maximized == desired {
			continue
		}
		for j, candidate := range getFallbackCandidates(maximized) {
			if candidate == desired {
				continue
			}
			original, ok := supportedSet[candidate]
			if !ok {
				continue
			}
			distance := j*10 + i*40
			if maximize(candidate) == maximized {
				distance = i * 40
			}
			if distance < lowestDistance {
				lowestDistance = distance
				result = bestMatchResult{matchedDesired: desired, matchedSupported: original, distance: distance}
			}
			break
		}
	}
	if result.matchedSupported != "" && lowestDistance == 0 {
		return result
	}

	lowestDistance = math.MaxInt
	for i, desired := range requested {
		for _, supportedLocale := range supported {
			noExtensionLocale, _ := removeUnicodeExtension(supportedLocale)
			distance := findMatchingDistance(desired, noExtensionLocale) + i*40
			if distance < lowestDistance {
				lowestDistance = distance
				result = bestMatchResult{matchedDesired: desired, matchedSupported: supportedLocale, distance: distance}
			}
		}
	}
	if lowestDistance >= threshold {
		return bestMatchResult{}
	}
	return result
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

func maximize(tag string) string {
	loc, err := locale.Parse(tag)
	if err != nil {
		return tag
	}
	return loc.Maximize().String()
}
