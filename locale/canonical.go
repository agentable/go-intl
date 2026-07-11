package locale

import (
	"golang.org/x/text/language"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/localeid"
)

func (l Locale) Maximize() Locale {
	lang, script, region := localeid.Parts(l.tag)
	if maxLang, maxScript, maxRegion, ok := cldrlocale.MaximizeSubtags(lang, script, region); ok {
		l.tag = mustLanguageTag(localeid.Join(maxLang, maxScript, maxRegion))
	}
	l.freeze()
	return l
}

// Minimize is a two-tier lookup, not two competing algorithms: a precomputed
// CLDR minimize table for known subtag triples, falling back to the general
// ECMA-402 RemoveLikelySubtags trial below for arbitrary user input. Both tiers
// are driven by the same generated CLDR data (the fallback via Maximize), so they
// are consistent by construction and cannot drift.
func (l Locale) Minimize() Locale {
	lang, script, region := localeid.Parts(l.tag)
	if minLang, minScript, minRegion, ok := cldrlocale.MinimizeSubtags(lang, script, region); ok {
		l.tag = mustLanguageTag(localeid.Join(minLang, minScript, minRegion))
		l.freeze()
		return l
	}
	max := l.Maximize().tag.String()
	for _, candidate := range []string{
		lang,
		localeid.Join(lang, "", region),
		localeid.Join(lang, script, ""),
	} {
		if candidate == "" {
			continue
		}
		trial := l
		trial.tag = mustLanguageTag(candidate)
		if trial.Maximize().tag.String() == max {
			l.tag = trial.tag
			l.freeze()
			return l
		}
	}
	l.freeze()
	return l
}

func mustLanguageTag(s string) language.Tag {
	return language.MustParse(s)
}
