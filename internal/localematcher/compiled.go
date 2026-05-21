package localematcher

import (
	"math"
	"sync"
)

const derivedFallbackDistancePenalty = 80

// Matcher holds per-surface locale matching indexes built from immutable
// supported locale data.
type Matcher struct {
	supported             []string
	noExtension           []string
	maximized             []string
	derived               []bool
	exact                 map[string]string
	dataLocaleByAvailable map[string]string
	maximizedByLocale     map[string]string
	maximizer             Maximizer
	maximizedRequested    sync.Map
}

// NewMatcher compiles supported locale data for repeated constructor locale
// negotiation.
func NewMatcher(supported []string, maximizer Maximizer) *Matcher {
	available := availableLocalesFor(supported)
	m := &Matcher{
		supported:             make([]string, 0, len(available)),
		noExtension:           make([]string, 0, len(available)),
		maximized:             make([]string, 0, len(available)),
		derived:               make([]bool, 0, len(available)),
		exact:                 make(map[string]string, len(available)),
		dataLocaleByAvailable: make(map[string]string, len(available)),
		maximizedByLocale:     make(map[string]string, len(available)),
		maximizer:             normalizeMaximizer(maximizer),
	}
	for _, loc := range available {
		noExtensionLocale, _ := removeUnicodeExtension(loc.locale)
		maximizedLocale := m.maximizer(noExtensionLocale)
		m.supported = append(m.supported, loc.locale)
		m.noExtension = append(m.noExtension, noExtensionLocale)
		m.maximized = append(m.maximized, maximizedLocale)
		m.derived = append(m.derived, loc.derived)
		m.exact[noExtensionLocale] = loc.locale
		m.dataLocaleByAvailable[loc.locale] = loc.dataLocale
		m.maximizedByLocale[noExtensionLocale] = maximizedLocale
	}
	return m
}

func (m *Matcher) Match(requested []string, defaultLocale string, alg Algorithm) Result {
	if m == nil {
		return MatchWithMaximizer(requested, nil, defaultLocale, alg, nil)
	}
	if alg == AlgorithmLookup {
		return m.lookup(requested, defaultLocale)
	}
	return m.bestFit(requested, defaultLocale)
}

func (m *Matcher) lookup(requested []string, defaultLocale string) Result {
	for _, loc := range requested {
		noExtensionLocale, extension := removeUnicodeExtension(loc)
		availableLocale := m.bestAvailableLocale(noExtensionLocale)
		if availableLocale != "" {
			return Result{Locale: availableLocale, DataLocale: m.dataLocale(availableLocale), Extension: extension}
		}
	}
	return Result{Locale: defaultLocale, DataLocale: defaultLocale}
}

func (m *Matcher) bestAvailableLocale(locale string) string {
	candidate := locale
	for {
		if available, ok := m.exact[candidate]; ok {
			return available
		}
		pos := truncationPosition(candidate)
		if pos < 0 {
			return ""
		}
		candidate = candidate[:pos]
	}
}

func (m *Matcher) bestFit(requested []string, defaultLocale string) Result {
	var requestedExtensions map[string]string
	var small [4]string
	var noExtensionRequested []string
	if len(requested) <= len(small) {
		noExtensionRequested = small[:len(requested)]
	} else {
		noExtensionRequested = make([]string, len(requested))
	}
	for i, loc := range requested {
		noExtensionLocale, extension := removeUnicodeExtension(loc)
		noExtensionRequested[i] = noExtensionLocale
		if extension != "" {
			if requestedExtensions == nil {
				requestedExtensions = map[string]string{}
			}
			requestedExtensions[noExtensionLocale] = extension
		}
	}
	result := m.findBestMatch(noExtensionRequested, DefaultMatchingThreshold)
	if result.matchedSupported == "" {
		return Result{Locale: defaultLocale, DataLocale: defaultLocale}
	}
	noExtensionLocale, supportedExtension := removeUnicodeExtension(result.matchedSupported)
	extension := supportedExtension
	if requestedExtension := requestedExtensions[result.matchedDesired]; requestedExtension != "" {
		extension = requestedExtension
	}
	return Result{Locale: noExtensionLocale, DataLocale: m.dataLocale(noExtensionLocale), Extension: extension, Distance: result.distance}
}

func (m *Matcher) findBestMatch(requested []string, threshold int) bestMatchResult {
	result := findBestMatchExact(requested, m.exact)
	if result.matchedSupported != "" && result.distance == 0 {
		return result
	}
	lowestDistance := resultDistance(result)

	for i, desired := range requested {
		maximized := m.maximize(desired)
		if maximized == desired {
			continue
		}
		for j, candidate := range getFallbackCandidates(maximized) {
			if candidate == desired {
				continue
			}
			original, ok := m.exact[candidate]
			if !ok {
				continue
			}
			distance := j*10 + i*40
			if m.maximizedByLocale[candidate] == maximized {
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
		maximizedDesired := m.maximize(desired)
		for j, supportedLocale := range m.supported {
			distance := cachedMatchingDistance(desired, m.noExtension[j], maximizedDesired, m.maximized[j]) + i*40
			if m.derived[j] {
				distance += derivedFallbackDistancePenalty
			}
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

func (m *Matcher) maximize(locale string) string {
	if maximized, ok := m.maximizedByLocale[locale]; ok {
		return maximized
	}
	if maximized, ok := m.maximizedRequested.Load(locale); ok {
		return maximized.(string)
	}
	maximized := m.maximizer(locale)
	actual, _ := m.maximizedRequested.LoadOrStore(locale, maximized)
	return actual.(string)
}

func (m *Matcher) dataLocale(availableLocale string) string {
	if dataLocale, ok := m.dataLocaleByAvailable[availableLocale]; ok {
		return dataLocale
	}
	return availableLocale
}

func cachedMatchingDistance(desired, supported, maximizedDesired, maximizedSupported string) int {
	key := [4]string{desired, supported, maximizedDesired, maximizedSupported}
	if v, ok := distanceCache.Load(key); ok {
		return v.(int)
	}
	distance := matchingDistance(desired, supported, maximizedDesired, maximizedSupported)
	distanceCache.Store(key, distance)
	return distance
}

func findBestMatchExact(requested []string, exact map[string]string) bestMatchResult {
	lowestDistance := math.MaxInt
	result := bestMatchResult{}
	for i, desired := range requested {
		if original, ok := exact[desired]; ok {
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
	return result
}

func resultDistance(result bestMatchResult) int {
	if result.matchedSupported == "" {
		return math.MaxInt
	}
	return result.distance
}
