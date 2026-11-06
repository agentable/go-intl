package localematcher

import (
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/localeid"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

// Options corresponds to the official Intl.LocaleMatcher draft's options bag.
type Options struct {
	// Algorithm is "lookup" or "best fit"; nil defaults to "best fit".
	Algorithm *string
}

// Matcher selects from a fixed set of application resource locales. A Matcher
// is safe for concurrent calls. Its zero value has no available locales.
type Matcher struct {
	compiled  *localematcher.Matcher
	algorithm localematcher.Algorithm
	available map[string]locale.Locale
}

// New snapshots available and opts. Parse raw locale identifiers with
// locale.ParseList before calling New.
func New(available locale.List, opts Options) (*Matcher, error) {
	canonical := ecma402.CanonicalLocaleList(available)
	for _, loc := range canonical {
		_, extension, err := localeid.RemoveUnicodeExtension(loc.String())
		if err != nil || extension != "" {
			return nil, ecma402.InvalidValueErrorExpected("localematcher", "availableLocales", loc.String(), loc.String(), "a locale without a Unicode locale extension", err)
		}
	}
	algorithmValue := "best fit"
	ecma402.ApplyOption(&algorithmValue, opts.Algorithm)
	check := ecma402.RequiredStringOption("algorithm", algorithmValue, "lookup", "best fit")
	if err := ecma402.ValidateStringOptions("localematcher", "", check); err != nil {
		return nil, err
	}
	algorithm, _ := ecma402.LocaleMatcherAlgorithm(algorithmValue)
	m := &Matcher{
		compiled:  localematcher.NewExactMatcher(canonical.Strings(), cldrlocale.Maximize),
		algorithm: algorithm,
		available: make(map[string]locale.Locale, len(canonical)),
	}
	for _, loc := range canonical {
		m.available[loc.String()] = loc
	}
	return m, nil
}

// Match returns a canonical available locale, or ok=false when no locale
// matches. The application owns defaulting; requested Unicode extensions are
// ignored and are not copied into the selected resource locale.
func (m *Matcher) Match(requested locale.List) (matched locale.Locale, ok bool) {
	if m.compiled == nil {
		return locale.Locale{}, false
	}
	result := m.compiled.Match(ecma402.RequestedLocaleStrings(requested), "", m.algorithm)
	matched, ok = m.available[result.Locale]
	return matched, ok
}
