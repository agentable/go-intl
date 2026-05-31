package codegen

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

func renderSupported(input RuntimeInput) ([]byte, error) {
	var b strings.Builder
	b.WriteString(`package cldr

import (
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/agentable/go-intl/internal/numbering"
)

// NumberSupportedLocales returns locales with generated number data.
func NumberSupportedLocales() []string {
	numberSupportedLocalesOnce.Do(func() {
		numberSupportedLocales = supportedLocaleTags(numberDataByLocale())
	})
	return numberSupportedLocales
}

// DateSupportedLocales returns locales with generated date data.
func DateSupportedLocales() []string {
	dateSupportedLocalesOnce.Do(func() {
		dateSupportedLocales = supportedLocaleTags(dateDataByLocale())
	})
	return dateSupportedLocales
}

// UnitSupportedLocales returns locales with generated unit pattern data.
func UnitSupportedLocales() []string {
	unitSupportedLocalesOnce.Do(func() {
		unitSupportedLocales = unitSupportedLocaleTags()
	})
	return unitSupportedLocales
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
	supportedNumberingSystemsOnce.Do(func() {
		supportedNumberingSystems = supportedNumberingSystemValues()
	})
	return supportedNumberingSystems
}

// SupportedTimeZones returns canonical IANA time-zone names backed by generated metazone data.
func SupportedTimeZones() []string {
	supportedTimeZonesOnce.Do(func() {
		supportedTimeZones = supportedTimeZoneValues()
	})
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
		return SupportedCalendars()
	case "hc":
		return hourCycleLocaleData(locale)
	case "nu":
		return numberingSystemLocaleData(locale)
	default:
		return nil
	}
}

var (
	numberSupportedLocalesOnce    sync.Once
	dateSupportedLocalesOnce      sync.Once
	unitSupportedLocalesOnce      sync.Once
	supportedNumberingSystemsOnce sync.Once
	supportedTimeZonesOnce        sync.Once

	numberSupportedLocales    []string
	dateSupportedLocales      []string
	unitSupportedLocales      []string
	supportedCollations       = collationValues
	supportedNumberingSystems []string
	supportedTimeZones        []string
)
`)
	writeStringSliceVar(&b, "supportedCalendars", supportedCalendarValues(input.Dates))
	writeStringSliceVar(&b, "supportedCurrencies", supportedCurrencyValues(input.Currencies))
	writeStringSliceVar(&b, "supportedNumberingSystemExtras", supportedNumberingSystemExtras(input.Numbers))
	writeStringSliceVar(&b, "supportedTimeZoneCandidates", supportedTimeZoneCandidates(input.Metazones))
	b.WriteString(`
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

func supportedNumberingSystemValues() []string {
	out := slices.Clone(numbering.SimpleNumberingSystems)
	for _, numberingSystem := range supportedNumberingSystemExtras {
		if numberingSystem != "" && !slices.Contains(out, numberingSystem) {
			out = append(out, numberingSystem)
		}
	}
	slices.Sort(out)
	return out
}

func supportedTimeZoneValues() []string {
	seen := map[string]bool{}
	for _, zone := range supportedTimeZoneCandidates {
		seen[CanonicalTimeZoneLink(zone)] = true
	}
	return slices.Sorted(maps.Keys(seen))
}

func numberingSystemLocaleData(locale string) []string {
	defaultNS := "latn"
	if loc, ok := ResolveLocale(locale); ok {
		if ns := loc.DefaultNumberingSystem(); ns != "" {
			defaultNS = ns
		}
	}
	return keywordLocaleData(defaultNS, numbering.SimpleNumberingSystems)
}

func hourCycleLocaleData(locale string) []string {
	region := localeRegion(locale)
	if region == "" {
		if localeLanguage(locale) == "en" {
			return []string{"h12", "h23"}
		}
	}
	return keywordLocaleData("", HourCyclePreference(region))
}

func localeLanguage(locale string) string {
	language, _, _ := strings.Cut(locale, "-")
	return language
}

func localeRegion(locale string) string {
	first := true
	for part := range strings.SplitSeq(locale, "-") {
		if first {
			first = false
			continue
		}
		if len(part) == 1 {
			return ""
		}
		if len(part) == 2 && isASCIIAlpha(part) || len(part) == 3 && isASCIIDigit(part) {
			if part == "ZZ" {
				return ""
			}
			return part
		}
	}
	return ""
}

func isASCIIAlpha(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' && (r < 'a' || r > 'z') {
			return false
		}
	}
	return s != ""
}

func isASCIIDigit(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
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
`)
	return FormatFile([]byte(b.String()))
}

func supportedCalendarValues(dates extract.Dates) []string {
	seen := map[string]bool{}
	for _, data := range dates {
		for calendar := range data.Calendars {
			switch calendar {
			case "gregorian":
				seen["gregory"] = true
			default:
				seen[calendar] = true
			}
		}
	}
	if seen["gregory"] {
		seen["iso8601"] = true
	}
	return slices.Sorted(maps.Keys(seen))
}

func supportedCurrencyValues(data extract.CurrencyData) []string {
	seen := map[string]bool{}
	for code := range data.Fractions {
		if code != "DEFAULT" {
			seen[code] = true
		}
	}
	for _, currencies := range data.Currencies {
		for code := range currencies {
			seen[code] = true
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

func supportedNumberingSystemExtras(numbers extract.Numbers) []string {
	seen := map[string]bool{}
	for _, data := range numbers {
		if data.DefaultNumberingSystem != "" {
			seen[data.DefaultNumberingSystem] = true
		}
		for numberingSystem := range data.Symbols {
			seen[numberingSystem] = true
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

func supportedTimeZoneCandidates(data extract.Metazones) []string {
	seen := map[string]bool{}
	for zone := range data.ZoneToMetazones {
		if zone == "Etc/Unknown" {
			continue
		}
		seen[zone] = true
	}
	return slices.Sorted(maps.Keys(seen))
}

func writeStringSliceVar(b *strings.Builder, name string, values []string) {
	fmt.Fprintf(b, "\nvar %s = []string{\n", name)
	for _, value := range values {
		fmt.Fprintf(b, "\t%s,\n", strconv.Quote(value))
	}
	b.WriteString("}\n")
}
