package locale

import (
	"fmt"

	"golang.org/x/text/language"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/localeid"
)

func (l Locale) Maximize() Locale {
	lang, script, region := localeid.Parts(l.tag)
	if maxLang, maxScript, maxRegion, ok := cldrlocale.MaximizeSubtags(lang, script, region); ok {
		if tag, err := localeid.ReplaceLanguageSubtags(l.tag, maxLang, maxScript, maxRegion); err == nil {
			l.tag = tag
		}
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
		if tag, err := localeid.ReplaceLanguageSubtags(l.tag, minLang, minScript, minRegion); err == nil {
			l.tag = tag
		}
		l.freeze()
		return l
	}
	maxLang, maxScript, maxRegion := localeid.Parts(l.Maximize().tag)
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
		trialLang, trialScript, trialRegion := localeid.Parts(trial.Maximize().tag)
		if trialLang == maxLang && trialScript == maxScript && trialRegion == maxRegion {
			minLang, minScript, minRegion := localeid.Parts(trial.tag)
			if tag, err := localeid.ReplaceLanguageSubtags(l.tag, minLang, minScript, minRegion); err == nil {
				l.tag = tag
			}
			l.freeze()
			return l
		}
	}
	l.freeze()
	return l
}

func mustLanguageTag(s string) language.Tag {
	tag, err := language.Parse(s)
	if err != nil {
		panic(fmt.Errorf("locale: invalid internally constructed language tag: %w", err))
	}
	return tag
}
