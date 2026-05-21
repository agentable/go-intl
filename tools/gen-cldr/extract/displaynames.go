package extract

import "github.com/agentable/go-intl/tools/gen-cldr/cldr"

type DisplayNames = map[string]cldr.DisplayNames

func ExtractDisplayNames(raw map[string]cldr.DisplayNames, locales []string) DisplayNames {
	selected := localeSet(locales)
	out := make(DisplayNames, len(selected))
	for locale, data := range raw {
		if selected[locale] {
			out[locale] = data
		}
	}
	return out
}
