package codegen

type wrapperSource struct {
	name   string
	source string
}

func renderWrapperFiles() ([]generatedFile, error) {
	sources := []wrapperSource{
		{name: "locale/accessors.go", source: `package cldrlocale

import "github.com/agentable/go-intl/internal/localeid"

// Locale is an opaque handle into the generated CLDR locale data.
type Locale uint16

// Undefined is the "und" sentinel locale.
const Undefined Locale = 0

func Maximize(tag string) string {
	return localeid.Maximize(tag, MaximizeSubtags)
}

func SupportedCollations() []string {
	return collationValues
}
`},
		{name: "number/accessors.go", source: `package number

import "github.com/agentable/go-intl/internal/cldr"

type CurrencyData = cldr.CurrencyData
type Locale = cldr.Locale
type NumberLocaleData = cldr.NumberLocaleData
type NumberSymbols = cldr.NumberSymbols

func SupportedLocales() []string {
	return cldr.NumberSupportedLocales()
}

func ResolveLocale(tag string) (Locale, bool) {
	return cldr.ResolveLocale(tag)
}

func CurrencyDigits(code string) CurrencyData {
	return cldr.CurrencyDigits(code)
}
`},
		{name: "date/accessors.go", source: `package date

import "github.com/agentable/go-intl/internal/cldr"

type DateLocaleData = cldr.DateLocaleData
type Gregorian = cldr.Gregorian
type Locale = cldr.Locale

const Undefined = cldr.Undefined

func SupportedLocales() []string {
	return cldr.DateSupportedLocales()
}

func ResolveLocale(tag string) (Locale, bool) {
	return cldr.ResolveLocale(tag)
}

func GregorianFor(loc Locale) Gregorian {
	return cldr.GregorianFor(loc)
}

func DayPeriodFor(loc Locale, hour, minute int) string {
	return cldr.DayPeriodFor(loc, hour, minute)
}
`},
		{name: "relativetime/accessors.go", source: `package relativetime

import "github.com/agentable/go-intl/internal/cldr"

type Locale = cldr.Locale
type RelativeTimeField = cldr.RelativeTimeField
type RelativeTimeFields = cldr.RelativeTimeFields

func SupportedLocales() []string {
	return cldr.RelativeTimeSupportedLocales()
}

func ResolveLocale(tag string) (Locale, bool) {
	return cldr.ResolveLocale(tag)
}

func FieldsFor(loc Locale) RelativeTimeFields {
	return loc.RelativeTimeFields()
}
`},
		{name: "timezone/names.go", source: `package timezone

import "github.com/agentable/go-intl/internal/cldr"

type Locale = cldr.Locale
type TimeZoneName = cldr.TimeZoneName

const (
	TimeZoneNameShort        = cldr.TimeZoneNameShort
	TimeZoneNameLong         = cldr.TimeZoneNameLong
	TimeZoneNameShortOffset  = cldr.TimeZoneNameShortOffset
	TimeZoneNameLongOffset   = cldr.TimeZoneNameLongOffset
	TimeZoneNameShortGeneric = cldr.TimeZoneNameShortGeneric
	TimeZoneNameLongGeneric  = cldr.TimeZoneNameLongGeneric
)

func CanonicalTimeZoneLink(name string) string {
	return cldr.CanonicalTimeZoneLink(name)
}

func TimeZoneDisplayName(loc Locale, zone string, form TimeZoneName, isDST bool, instant int64, offsetMs int64) string {
	return cldr.TimeZoneDisplayName(loc, zone, form, isDST, instant, offsetMs)
}

func GMTOffsetName(loc Locale, offsetMs int64, form TimeZoneName) string {
	return cldr.GMTOffsetName(loc, offsetMs, form)
}
`},
	}
	files := make([]generatedFile, 0, len(sources))
	for _, wrapper := range sources {
		src, err := FormatFile([]byte(wrapper.source))
		if err != nil {
			return nil, err
		}
		files = append(files, generatedFile{name: wrapper.name, src: src})
	}
	return files, nil
}
