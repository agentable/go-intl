package localeid

import "golang.org/x/text/language"

type SubtagMaximizer func(language, script, region string) (lang, scr, reg string, ok bool)

func Maximize(tag string, maximizer SubtagMaximizer) string {
	parsed, err := language.Parse(tag)
	if err != nil {
		return tag
	}
	if maximizer == nil {
		return parsed.String()
	}
	lang, script, region := Parts(parsed)
	if maxLang, maxScript, maxRegion, ok := maximizer(lang, script, region); ok {
		return Join(maxLang, maxScript, maxRegion)
	}
	return parsed.String()
}
