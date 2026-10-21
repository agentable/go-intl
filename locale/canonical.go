package locale

import (
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/localeid"
)

func (l Locale) Maximize() Locale {
	lang, script, region := localeid.Parts(l.tag)
	if maxLang, maxScript, maxRegion, ok := cldrlocale.MaximizeSubtags(lang, script, region); ok {
		l.tag = localeid.ReplaceLanguageSubtags(l.tag, maxLang, maxScript, maxRegion)
	}
	l.freeze()
	return l
}

// Minimize prefers a region over a script when both preserve the maximal LSR.
func (l Locale) Minimize() Locale {
	maximal := l.Maximize()
	lang, script, region := localeid.Parts(maximal.tag)
	for _, candidate := range [][3]string{
		{lang, "", ""},
		{lang, "", region},
		{lang, script, ""},
	} {
		trialLang, trialScript, trialRegion, ok := cldrlocale.MaximizeSubtags(candidate[0], candidate[1], candidate[2])
		if !ok || trialLang != lang || trialScript != script || trialRegion != region {
			continue
		}
		l.tag = localeid.ReplaceLanguageSubtags(l.tag, candidate[0], candidate[1], candidate[2])
		l.freeze()
		return l
	}
	return maximal
}
