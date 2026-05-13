package extract

import "github.com/agentable/go-intl/tools/gen-cldr/cldr"

type Metazones = cldr.Metazones

type Units = map[string]cldr.Units

func ExtractMetazones(raw cldr.Metazones, locales []string) Metazones {
	selected := localeSet(locales)
	names := make(map[string]map[string]cldr.MetazoneNames, len(selected))
	for locale, value := range raw.Names {
		if selected[locale] {
			names[locale] = value
		}
	}
	cities := make(map[string]map[string]string, len(selected))
	for locale, value := range raw.ExemplarCities {
		if selected[locale] {
			cities[locale] = value
		}
	}
	formats := make(map[string]cldr.TimeZoneFormats, len(selected))
	for locale, value := range raw.Formats {
		if selected[locale] {
			formats[locale] = value
		}
	}
	return Metazones{ZoneToMetazones: raw.ZoneToMetazones, Names: names, ExemplarCities: cities, Formats: formats}
}

func ExtractUnits(raw map[string]cldr.Units, locales []string) Units {
	selected := localeSet(locales)
	out := make(Units, len(selected))
	for locale, units := range raw {
		if !selected[locale] {
			continue
		}
		if data, ok := units["meter"]; ok {
			out[locale] = cldr.Units{"meter": data}
		}
	}
	return out
}
