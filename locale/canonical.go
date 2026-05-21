package locale

import (
	"golang.org/x/text/language"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/localeid"
)

func (l Locale) Maximize() Locale {
	lang, script, region := tagParts(l.tag)
	if maxLang, maxScript, maxRegion, ok := cldrlocale.MaximizeSubtags(lang, script, region); ok {
		l.tag = mustLanguageTag(joinTagParts(maxLang, maxScript, maxRegion))
	}
	return l
}

func (l Locale) Minimize() Locale {
	lang, script, region := tagParts(l.tag)
	if minLang, minScript, minRegion, ok := cldrlocale.MinimizeSubtags(lang, script, region); ok {
		l.tag = mustLanguageTag(joinTagParts(minLang, minScript, minRegion))
		return l
	}
	max := l.Maximize().tag.String()
	for _, candidate := range []string{
		lang,
		joinTagParts(lang, "", region),
		joinTagParts(lang, script, ""),
	} {
		if candidate == "" {
			continue
		}
		trial := l
		trial.tag = mustLanguageTag(candidate)
		if trial.Maximize().tag.String() == max {
			l.tag = trial.tag
			return l
		}
	}
	return l
}

func tagParts(tag language.Tag) (lang, script, region string) {
	return localeid.Parts(tag)
}

func joinTagParts(lang, script, region string) string {
	return localeid.Join(lang, script, region)
}

func mustLanguageTag(s string) language.Tag {
	return language.MustParse(s)
}
