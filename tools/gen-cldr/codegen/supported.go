package codegen

func renderSupported() ([]byte, error) {
	return FormatFile([]byte(`package cldr

import (
	"maps"
	"slices"

	"github.com/agentable/go-intl/internal/ecma402"
	"golang.org/x/text/language"
)

// NumberSupportedLocales returns locales with generated number data.
func NumberSupportedLocales() []string {
	return numberSupportedLocales
}

// DateSupportedLocales returns locales with generated Gregorian date data.
func DateSupportedLocales() []string {
	return dateSupportedLocales
}

// SupportedCalendars returns canonical ECMA-402 calendar identifiers backed by generated date data.
func SupportedCalendars() []string {
	return supportedCalendars
}

// SupportedCollations returns canonical collation identifiers backed by generated locale data.
func SupportedCollations() []string {
	return supportedCollations
}

// SupportedCurrencies returns canonical ISO 4217 currency codes backed by generated number data.
func SupportedCurrencies() []string {
	return supportedCurrencies
}

// SupportedNumberingSystems returns canonical numbering system identifiers backed by generated number data.
func SupportedNumberingSystems() []string {
	return supportedNumberingSystems
}

// SupportedTimeZones returns canonical IANA time-zone names backed by generated metazone data.
func SupportedTimeZones() []string {
	return supportedTimeZones
}

// NumberLocaleData exposes the relevant extension keys used by Intl.NumberFormat ResolveLocale.
type NumberLocaleData struct{}

func (NumberLocaleData) For(locale, key string) []string {
	if key != "nu" {
		return nil
	}
	return numberingSystemLocaleData(locale)
}

// DateLocaleData exposes the relevant extension keys used by Intl.DateTimeFormat ResolveLocale.
type DateLocaleData struct{}

func (DateLocaleData) For(locale, key string) []string {
	switch key {
	case "ca":
		return []string{"gregory", "iso8601"}
	case "hc":
		return hourCycleLocaleData(locale)
	case "nu":
		return numberingSystemLocaleData(locale)
	default:
		return nil
	}
}

var (
	numberSupportedLocales    = supportedLocaleTags(numbersByLocale)
	dateSupportedLocales      = supportedLocaleTags(datesByLocale)
	supportedCalendars        = supportedCalendarValues()
	supportedCollations       = collationValues
	supportedCurrencies       = supportedCurrencyValues()
	supportedNumberingSystems = supportedNumberingSystemValues()
	supportedTimeZones        = supportedTimeZoneValues()
)

func supportedLocaleTags[T any](data map[Locale]T) []string {
	tags := make([]string, 0, len(data))
	for i := 1; i < len(localeRecords); i++ {
		loc := Locale(i)
		if _, ok := data[loc]; ok {
			tags = append(tags, localeRecords[i].tag.string())
		}
	}
	return tags
}

func supportedCalendarValues() []string {
	seen := map[string]bool{"iso8601": true}
	for _, calendars := range datesByLocale {
		for calendar := range calendars {
			if calendar == "gregorian" {
				calendar = "gregory"
			}
			seen[calendar] = true
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

func supportedCurrencyValues() []string {
	seen := map[string]bool{}
	for code := range currencyFractions {
		if code != "DEFAULT" {
			seen[code] = true
		}
	}
	for _, currencies := range currenciesByLocale {
		for code := range currencies {
			seen[code] = true
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

func supportedNumberingSystemValues() []string {
	seen := map[string]bool{}
	for _, numberingSystem := range ecma402.SimpleNumberingSystems {
		seen[numberingSystem] = true
	}
	for _, data := range numbersByLocale {
		if data.defaultNS != "" {
			seen[data.defaultNS] = true
		}
		for numberingSystem := range data.symbols {
			seen[numberingSystem] = true
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

func supportedTimeZoneValues() []string {
	seen := map[string]bool{}
	for zone := range zoneToMetazones {
		if zone == "Etc/Unknown" {
			continue
		}
		seen[CanonicalTimeZoneLink(zone)] = true
	}
	return slices.Sorted(maps.Keys(seen))
}

func numberingSystemLocaleData(locale string) []string {
	defaultNS := "latn"
	if loc, ok := ResolveLocale(language.Make(locale)); ok {
		if ns := loc.DefaultNumberingSystem(); ns != "" {
			defaultNS = ns
		}
	}
	return keywordLocaleData(defaultNS, ecma402.SimpleNumberingSystems)
}

func hourCycleLocaleData(locale string) []string {
	region := localeRegion(locale)
	if region == "" {
		if lang, _, _ := language.Make(locale).Raw(); lang.String() == "en" {
			return []string{"h12", "h23"}
		}
	}
	return keywordLocaleData("", HourCyclePreference(region))
}

func localeRegion(locale string) string {
	_, _, region := language.Make(locale).Raw()
	if region.IsPrivateUse() || region.String() == "ZZ" {
		return ""
	}
	return region.String()
}

func keywordLocaleData(defaultValue string, values []string) []string {
	out := make([]string, 0, len(values)+1)
	if defaultValue != "" {
		out = append(out, defaultValue)
	}
	for _, value := range values {
		if value != "" && !slices.Contains(out, value) {
			out = append(out, value)
		}
	}
	return out
}
`))
}
