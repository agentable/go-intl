package extract

import "github.com/agentable/go-intl/tools/gen-cldr/cldr"

type LocaleMatching struct {
	Language cldr.LanguageMatching
	Regions  map[string][]string
}

func ExtractLocaleMatching(language cldr.LanguageMatching, regions map[string][]string) LocaleMatching {
	return LocaleMatching{Language: language, Regions: regions}
}
