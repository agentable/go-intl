package locale

import (
	"strings"

	"golang.org/x/text/language"

	"github.com/agentable/go-intl/internal/cldr"
)

func (l Locale) Maximize() Locale {
	lang, script, region := tagParts(l.tag)
	if maxLang, maxScript, maxRegion, ok := cldr.MaximizeSubtags(lang, script, region); ok {
		l.tag = mustLanguageTag(joinTagParts(maxLang, maxScript, maxRegion))
	}
	return l
}

func (l Locale) Minimize() Locale {
	lang, script, region := tagParts(l.tag)
	if minLang, minScript, minRegion, ok := cldr.MinimizeSubtags(lang, script, region); ok {
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
	base, scr, reg := tag.Raw()
	lang = base.String()
	if !scr.IsPrivateUse() {
		if s := scr.String(); s != "Zzzz" {
			script = s
		}
	}
	if !reg.IsPrivateUse() {
		if r := reg.String(); r != "ZZ" {
			region = r
		}
	}
	return lang, script, region
}

func joinTagParts(lang, script, region string) string {
	parts := []string{lang}
	if script != "" {
		parts = append(parts, script)
	}
	if region != "" {
		parts = append(parts, region)
	}
	return strings.Join(parts, "-")
}

func mustLanguageTag(s string) language.Tag {
	return language.MustParse(s)
}
