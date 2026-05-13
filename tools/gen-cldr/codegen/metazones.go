package codegen

import (
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

func renderMetazones(data extract.Metazones, table *StringTable) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldr\n\n")
	b.WriteString("import (\n\t\"strconv\"\n\t\"strings\"\n)\n\n")
	b.WriteString("type MetazoneNames struct{ LongGeneric, LongStandard, LongDaylight, ShortGeneric, ShortStandard, ShortDaylight string }\n\n")
	b.WriteString("type TimeZoneFormats struct{ GMTFormat, GMTZeroFormat, HourFormat string }\n\n")
	b.WriteString("type MetazonePeriod struct {\n\tMetazone string\n\tStart    int64\n\tEnd      int64\n}\n\n")
	b.WriteString("type TimeZoneName string\n\n")
	b.WriteString("const (\n\tTimeZoneNameShort TimeZoneName = \"short\"\n\tTimeZoneNameLong TimeZoneName = \"long\"\n\tTimeZoneNameShortOffset TimeZoneName = \"shortOffset\"\n\tTimeZoneNameLongOffset TimeZoneName = \"longOffset\"\n\tTimeZoneNameShortGeneric TimeZoneName = \"shortGeneric\"\n\tTimeZoneNameLongGeneric TimeZoneName = \"longGeneric\"\n)\n\n")
	b.WriteString("var zoneToMetazones = ")
	b.WriteString(metazonePeriodMapLiteral(data.ZoneToMetazones, table))
	b.WriteString("\n\n")
	b.WriteString("var metazoneNamesByLocale = map[Locale]map[string]MetazoneNames{\n")
	for _, locale := range sortedLocaleKeys(data.Names) {
		b.WriteString("\tlocaleIndex[")
		b.WriteString(strconv.Quote(locale))
		b.WriteString("]: ")
		b.WriteString(metazoneNamesMapLiteral(data.Names[locale], table))
		b.WriteString(",\n")
	}
	b.WriteString("}\n\n")
	b.WriteString("var exemplarCitiesByLocale = map[Locale]map[string]string{\n")
	for _, locale := range sortedLocaleKeys(data.ExemplarCities) {
		b.WriteString("\tlocaleIndex[")
		b.WriteString(strconv.Quote(locale))
		b.WriteString("]: ")
		b.WriteString(stringMapLiteral(data.ExemplarCities[locale], table))
		b.WriteString(",\n")
	}
	b.WriteString("}\n\n")
	b.WriteString("var timeZoneFormatsByLocale = map[Locale]TimeZoneFormats{\n")
	for _, locale := range sortedLocaleKeys(data.Formats) {
		b.WriteString("\tlocaleIndex[")
		b.WriteString(strconv.Quote(locale))
		b.WriteString("]: ")
		b.WriteString(timeZoneFormatsLiteral(data.Formats[locale], table))
		b.WriteString(",\n")
	}
	b.WriteString("}\n\n")
	b.WriteString(`func ZoneToMetazone(zone string) string {
	return TimeZoneMetazone(zone, 0)
}

func TimeZoneMetazone(zone string, instant int64) string {
	for _, period := range zoneToMetazones[zone] {
		if instant >= period.Start && instant < period.End {
			return period.Metazone
		}
	}
	return ""
}

func TimeZoneDisplayName(loc Locale, zone string, form TimeZoneName, isDST bool, instant int64, offsetMs int64) string {
	kind := "long-generic"
	switch form {
	case TimeZoneNameLong:
		if isDST {
			kind = "long-daylight"
		} else {
			kind = "long-standard"
		}
	case TimeZoneNameShort:
		if isDST {
			kind = "short-daylight"
		} else {
			kind = "short-standard"
		}
	case TimeZoneNameShortGeneric:
		kind = "short-generic"
	}
	if name := loc.MetazoneName(TimeZoneMetazone(zone, instant), kind); name != "" {
		return name
	}
	if city := loc.ExemplarCity(zone); city != "" {
		return "Time in " + city
	}
	return GMTOffsetName(loc, offsetMs, form)
}

func (l Locale) MetazoneName(metazone, kind string) string {
	data := metazoneNamesByLocale[l][metazone]
	switch kind {
	case "long-standard":
		return data.LongStandard
	case "long-daylight":
		return data.LongDaylight
	case "short-generic":
		return data.ShortGeneric
	case "short-standard":
		return data.ShortStandard
	case "short-daylight":
		return data.ShortDaylight
	default:
		return data.LongGeneric
	}
}

func (l Locale) ExemplarCity(zone string) string {
	return exemplarCitiesByLocale[l][zone]
}

func (l Locale) TimeZoneFormats() TimeZoneFormats {
	data := timeZoneFormatsByLocale[l]
	if data.GMTFormat == "" {
		data.GMTFormat = "GMT{0}"
	}
	if data.GMTZeroFormat == "" {
		data.GMTZeroFormat = "GMT"
	}
	if data.HourFormat == "" {
		data.HourFormat = "+HH:mm;-HH:mm"
	}
	return data
}

// GMTOffsetName formats an offset time-zone name using locale GMT patterns.
func GMTOffsetName(loc Locale, offsetMs int64, form TimeZoneName) string {
	formats := loc.TimeZoneFormats()
	long := form == TimeZoneNameLongOffset || form == TimeZoneNameLong
	offset := offsetPattern(formats.HourFormat, offsetMs, long)
	if offset == "" && !long {
		return formats.GMTZeroFormat
	}
	return strings.ReplaceAll(formats.GMTFormat, "{0}", offset)
}

func offsetPattern(hourFormat string, offsetMs int64, long bool) string {
	positive, negative, ok := strings.Cut(hourFormat, ";")
	if !ok {
		positive, negative = "+HH:mm", "-HH:mm"
	}
	pattern := positive
	if offsetMs < 0 {
		pattern = negative
		offsetMs = -offsetMs
	}
	totalMinutes := offsetMs / 60000
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	if !long {
		if hours == 0 && minutes == 0 {
			return ""
		}
		if minutes == 0 {
			pattern = removeMinuteField(pattern)
		}
		return replaceOffsetFields(pattern, int(hours), int(minutes), false)
	}
	return replaceOffsetFields(pattern, int(hours), int(minutes), true)
}

func removeMinuteField(pattern string) string {
	i := strings.IndexByte(pattern, 'm')
	if i < 0 {
		return pattern
	}
	start := i
	if start > 0 && pattern[start-1] == ':' {
		start--
	}
	end := i
	for end < len(pattern) && pattern[end] == 'm' {
		end++
	}
	return pattern[:start] + pattern[end:]
}

func replaceOffsetFields(pattern string, hours int, minutes int, padded bool) string {
	hour := strconv.Itoa(hours)
	minute := strconv.Itoa(minutes)
	if padded {
		hour = twoDigitASCII(hours)
		minute = twoDigitASCII(minutes)
	}
	out := replaceFieldRun(pattern, 'H', hour)
	out = replaceFieldRun(out, 'm', minute)
	return out
}

func replaceFieldRun(pattern string, field byte, replacement string) string {
	i := strings.IndexByte(pattern, field)
	if i < 0 {
		return pattern
	}
	end := i
	for end < len(pattern) && pattern[end] == field {
		end++
	}
	return pattern[:i] + replacement + pattern[end:]
}

func twoDigitASCII(value int) string {
	if value < 10 {
		return "0" + strconv.Itoa(value)
	}
	return strconv.Itoa(value)
}
`)
	return FormatFile([]byte(b.String()))
}

func metazonePeriodMapLiteral(values map[string][]cldr.MetazonePeriod, table *StringTable) string {
	return stringKeyedMapLiteral("map[string][]MetazonePeriod", values, func(value []cldr.MetazonePeriod) string {
		return metazonePeriodSliceLiteral(value, table)
	})
}

func metazonePeriodSliceLiteral(values []cldr.MetazonePeriod, table *StringTable) string {
	if len(values) == 0 {
		return "nil"
	}
	var b strings.Builder
	b.WriteString("[]MetazonePeriod{")
	for _, period := range values {
		b.WriteString("{Metazone: ")
		b.WriteString(refStringLiteral(period.Metazone, table))
		b.WriteString(", Start: ")
		b.WriteString(strconv.FormatInt(period.Start, 10))
		b.WriteString(", End: ")
		b.WriteString(strconv.FormatInt(period.End, 10))
		b.WriteString("}, ")
	}
	b.WriteString("}")
	return b.String()
}

func metazoneNamesMapLiteral(values map[string]cldr.MetazoneNames, table *StringTable) string {
	return stringKeyedMapLiteral("map[string]MetazoneNames", values, func(value cldr.MetazoneNames) string {
		return metazoneNamesLiteral(value, table)
	})
}

func metazoneNamesLiteral(names cldr.MetazoneNames, table *StringTable) string {
	return "MetazoneNames{" +
		"LongGeneric: " + refStringLiteral(names.LongGeneric, table) + ", " +
		"LongStandard: " + refStringLiteral(names.LongStandard, table) + ", " +
		"LongDaylight: " + refStringLiteral(names.LongDaylight, table) + ", " +
		"ShortGeneric: " + refStringLiteral(names.ShortGeneric, table) + ", " +
		"ShortStandard: " + refStringLiteral(names.ShortStandard, table) + ", " +
		"ShortDaylight: " + refStringLiteral(names.ShortDaylight, table) + "}"
}

func timeZoneFormatsLiteral(formats cldr.TimeZoneFormats, table *StringTable) string {
	return "TimeZoneFormats{" +
		"GMTFormat: " + refStringLiteral(formats.GMTFormat, table) + ", " +
		"GMTZeroFormat: " + refStringLiteral(formats.GMTZeroFormat, table) + ", " +
		"HourFormat: " + refStringLiteral(formats.HourFormat, table) + "}"
}

func renderUnits(data extract.Units, table *StringTable) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldr\n\n")
	b.WriteString("type unitData struct{ patterns map[string]map[string]map[string]string; compound map[string]string }\n\n")
	b.WriteString("var unitsByLocale = map[Locale]map[string]unitData{\n")
	for _, locale := range sortedLocaleKeys(data) {
		b.WriteString("\tlocaleIndex[")
		b.WriteString(strconv.Quote(locale))
		b.WriteString("]: {\n")
		for _, unit := range slices.Sorted(maps.Keys(data[locale])) {
			unitData := data[locale][unit]
			b.WriteString("\t\t")
			b.WriteString(strconv.Quote(unit))
			b.WriteString(": {patterns: ")
			b.WriteString(nestedStringMap3Literal(unitData.Patterns, table))
			b.WriteString(", compound: ")
			b.WriteString(stringMapLiteral(unitData.Compound, table))
			b.WriteString("},\n")
		}
		b.WriteString("\t},\n")
	}
	b.WriteString("}\n\n")
	b.WriteString(`func (l Locale) UnitPattern(unit, width, plural string) string {
	if width == "" {
		width = "long"
	}
	if plural == "" {
		plural = "other"
	}
	return unitsByLocale[l][unit].patterns[width][unit][plural]
}

func (l Locale) CompoundUnitPattern(width string) string {
	if width == "" {
		width = "long"
	}
	for _, data := range unitsByLocale[l] {
		if pattern := data.compound[width]; pattern != "" {
			return pattern
		}
	}
	return ""
}
`)
	return FormatFile([]byte(b.String()))
}

func nestedStringMap3Literal(values map[string]map[string]map[string]string, table *StringTable) string {
	return stringKeyedMapLiteral("map[string]map[string]map[string]string", values, func(value map[string]map[string]string) string {
		return nestedStringMapLiteral(value, table)
	})
}
