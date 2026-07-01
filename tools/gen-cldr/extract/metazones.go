package extract

import (
	"github.com/agentable/go-intl/internal/unitid"
	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

type Metazones = cldr.Metazones

type Units = map[string]cldr.Units

type ListPatterns = map[string]cldr.ListPatterns

type RelativeTimeFields = map[string]cldr.RelativeTimeFields

func ExtractMetazones(raw cldr.Metazones, locales []string) Metazones {
	selected := localeSet(locales)
	names := make(map[string]map[string]cldr.MetazoneNames, len(selected))
	for locale, value := range raw.Names {
		if selected[locale] {
			names[locale] = value
		}
	}
	zoneNames := make(map[string]map[string]cldr.MetazoneNames, len(selected))
	for locale, value := range raw.ZoneNames {
		if selected[locale] {
			zoneNames[locale] = value
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
	return Metazones{ZoneToMetazones: raw.ZoneToMetazones, Names: names, ZoneNames: zoneNames, ExemplarCities: cities, Formats: formats}
}

func ExtractUnits(raw map[string]cldr.Units, locales []string) Units {
	selected := localeSet(locales)
	out := make(Units, len(selected))
	for locale, units := range raw {
		if !selected[locale] {
			continue
		}
		extracted := make(cldr.Units)
		for unit, data := range units {
			if unitid.IsSanctionedSimpleUnitIdentifier(unit) {
				extracted[unit] = data
			}
		}
		if len(extracted) > 0 {
			out[locale] = extracted
		}
	}
	return out
}

func ExtractListPatterns(raw map[string]cldr.ListPatterns, locales []string) ListPatterns {
	selected := localeSet(locales)
	out := make(ListPatterns, len(selected))
	for locale, patterns := range raw {
		if selected[locale] {
			out[locale] = patterns
		}
	}
	return out
}

func ExtractRelativeTimeFields(raw map[string]cldr.RelativeTimeFields, locales []string) RelativeTimeFields {
	selected := localeSet(locales)
	out := make(RelativeTimeFields, len(selected))
	for locale, fields := range raw {
		if selected[locale] {
			out[locale] = fields
		}
	}
	return out
}
