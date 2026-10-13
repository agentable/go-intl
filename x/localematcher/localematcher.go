package localematcher

import (
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/localematcher"
)

// Algorithm selects the locale matching algorithm, mirroring the
// Intl.LocaleMatcher proposal's options.algorithm values.
type Algorithm = localematcher.Algorithm

const (
	// AlgorithmLookup is RFC 4647 lookup (ECMA-402 LookupMatcher): exact and
	// subtag-truncation matching only.
	AlgorithmLookup = localematcher.AlgorithmLookup

	// AlgorithmBestFit is the implementation-defined best-fit algorithm
	// (ECMA-402 BestFitMatcher) driven by the pinned CLDR language-matching
	// profile.
	AlgorithmBestFit = localematcher.AlgorithmBestFit
)

// DefaultMatchingThreshold is the best-fit distance threshold: a match at or
// above this distance counts as a miss and falls back to the default locale.
const DefaultMatchingThreshold = localematcher.DefaultMatchingThreshold

// Result describes one locale match. Locale is the matched available locale;
// DataLocale points at the concrete payload locale behind derived fallbacks;
// Extension carries the matched request's Unicode extension sequence;
// Distance is the CLDR matching distance, zero for an equivalent match.
type Result = localematcher.Result

// Matcher is a compiled available-locale index for repeated matching against
// a fixed available-locale set. Build it once and reuse it; it is safe for
// concurrent use.
type Matcher = localematcher.Matcher

// New compiles available into a Matcher with the same CLDR likely-subtags
// maximizer the Intl constructors use.
func New(available []string) *Matcher {
	return localematcher.NewMatcher(available, cldrlocale.Maximize)
}

// Match selects the best available locale for the requested list, mirroring
// the Stage 1 Intl.LocaleMatcher.match proposal. It falls back to
// defaultLocale when no requested locale matches at an acceptable distance.
func Match(requested, available []string, defaultLocale string, algorithm Algorithm) Result {
	return localematcher.MatchWithMaximizer(requested, available, defaultLocale, algorithm, cldrlocale.Maximize)
}
