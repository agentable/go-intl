package extract

import "github.com/agentable/go-intl/tools/gen-cldr/cldr"

type Dates = map[string]cldr.Dates

type PreferenceData = cldr.PreferenceData

func ExtractDates(raw map[string]cldr.Dates, locales []string) Dates {
	selected := localeSet(locales)
	out := make(Dates, len(selected))
	for locale, dates := range raw {
		if selected[locale] {
			out[locale] = dates
		}
	}
	return out
}

func ExtractPreference(raw cldr.PreferenceData) PreferenceData { return raw }
