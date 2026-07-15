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
	distanceProfile       compiledLanguageMatchingProfile
	distanceCache         sync.Map
	maximizedRequested    sync.Map
}

// NewMatcher compiles supported locale data for repeated constructor locale
// negotiation.
func NewMatcher(supported []string, maximizer Maximizer) *Matcher {
	maximizer = normalizeMaximizer(maximizer)
	available := availableLocalesFor(supported, maximizer)
	m := &Matcher{
		supported:             make([]string, len(available)),
		noExtension:           make([]string, len(available)),
		maximized:             make([]string, len(available)),
		derived:               make([]bool, len(available)),
		exact:                 make(map[string]string, len(available)),
		dataLocaleByAvailable: make(map[string]string, len(available)),
		maximizedByLocale:     make(map[string]string, len(available)),
		maximizer:             maximizer,
	}
	m.distanceProfile = compileLanguageMatchingProfile(defaultLanguageMatchingProfile(), m.maximizer)
	for i, loc := range available {
		noExtensionLocale, _ := removeUnicodeExtension(loc.locale)
		maximizedLocale := m.maximizer(noExtensionLocale)
		m.supported[i] = loc.locale
		m.noExtension[i] = noExtensionLocale
		m.maximized[i] = maximizedLocale
		m.derived[i] = loc.derived
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
	return m.defaultResult(defaultLocale)
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
	var smallRequested [4]string
	var smallExtensions [4]string
	var noExtensionRequested []string
	var requestedExtensions []string
	if len(requested) <= len(smallRequested) {
		noExtensionRequested = smallRequested[:len(requested)]
		requestedExtensions = smallExtensions[:len(requested)]
	} else {
		noExtensionRequested = make([]string, len(requested))
		requestedExtensions = make([]string, len(requested))
	}
	for i, loc := range requested {
		noExtensionLocale, extension := removeUnicodeExtension(loc)
		noExtensionRequested[i] = noExtensionLocale
		requestedExtensions[i] = extension
	}
	result := m.findBestMatch(noExtensionRequested, DefaultMatchingThreshold)
	if result.matchedSupported == "" {
		return m.defaultResult(defaultLocale)
	}
	noExtensionLocale, supportedExtension := removeUnicodeExtension(result.matchedSupported)
	extension := supportedExtension
	if requestedExtension := requestedExtensions[result.matchedDesiredIndex]; requestedExtension != "" {
		extension = requestedExtension
	}
	return Result{Locale: noExtensionLocale, DataLocale: m.dataLocale(noExtensionLocale), Extension: extension, Distance: result.distance}
}

func (m *Matcher) findBestMatch(requested []string, threshold int) bestMatchResult {
	result := bestMatchResult{}
	lowestDistance := math.MaxInt
	for i, desired := range requested {
		if original, ok := m.exact[desired]; ok {
			result = bestMatchResult{matchedDesiredIndex: i, matchedSupported: original, distance: i * 40}
			if result.distance == 0 {
				return result
			}
			lowestDistance = result.distance
			break
		}
	}

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
				result = bestMatchResult{matchedDesiredIndex: i, matchedSupported: original, distance: distance}
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
			distance := m.cachedMatchingDistance(desired, m.noExtension[j], maximizedDesired, m.maximized[j]) + i*40
			if m.derived[j] {
				distance += derivedFallbackDistancePenalty
			}
			if distance < lowestDistance {
				lowestDistance = distance
				result = bestMatchResult{matchedDesiredIndex: i, matchedSupported: supportedLocale, distance: distance}
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

func (m *Matcher) defaultResult(defaultLocale string) Result {
	return Result{Locale: defaultLocale, DataLocale: m.dataLocale(defaultLocale)}
}

func (m *Matcher) cachedMatchingDistance(desired, supported, maximizedDesired, maximizedSupported string) int {
	key := [4]string{desired, supported, maximizedDesired, maximizedSupported}
	if v, ok := m.distanceCache.Load(key); ok {
		return v.(int)
	}
	distance := m.distanceProfile.distance(maximizedDesired, maximizedSupported)
	m.distanceCache.Store(key, distance)
	return distance
}
