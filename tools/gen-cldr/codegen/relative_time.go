package codegen

import (
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

func renderRelativeTime(data extract.RelativeTimeFields, table *StringTable) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldr\n\n")
	b.WriteString("import \"sync\"\n\n")
	b.WriteString("type RelativeTimeField struct{ Future, Past, Relative map[string]string }\n\n")
	b.WriteString("type RelativeTimeFields map[string]map[string]RelativeTimeField\n\n")
	b.WriteString(`var (
	relativeTimeFieldsByLocaleOnce sync.Once
	relativeTimeFieldsByLocale     map[Locale]RelativeTimeFields
	relativeTimeSupportedOnce      sync.Once
	relativeTimeSupportedLocales   []string
)

func relativeTimeFieldDataByLocale() map[Locale]RelativeTimeFields {
	relativeTimeFieldsByLocaleOnce.Do(loadRelativeTimeFieldData)
	return relativeTimeFieldsByLocale
}

func loadRelativeTimeFieldData() {
	relativeTimeFieldsByLocale = map[Locale]RelativeTimeFields{
`)
	for _, locale := range sortedLocaleKeys(data) {
		b.WriteString("\tlocaleIndex[")
		b.WriteString(strconv.Quote(locale))
		b.WriteString("]: ")
		b.WriteString(relativeTimeUnitMapLiteral(data[locale], table))
		b.WriteString(",\n")
	}
	b.WriteString("\t}\n")
	b.WriteString("}\n\n")
	b.WriteString(`func RelativeTimeSupportedLocales() []string {
	relativeTimeSupportedOnce.Do(func() {
		relativeTimeSupportedLocales = supportedLocaleTags(relativeTimeFieldDataByLocale())
	})
	return relativeTimeSupportedLocales
}

func (l Locale) RelativeTimeFields() RelativeTimeFields {
	return relativeTimeFieldDataByLocale()[l]
}
`)
	return FormatFile([]byte(b.String()))
}

func relativeTimeUnitMapLiteral(values cldr.RelativeTimeFields, table *StringTable) string {
	if len(values) == 0 {
		return "nil"
	}
	var b strings.Builder
	b.WriteString("RelativeTimeFields{")
	for _, unit := range slices.Sorted(maps.Keys(values)) {
		b.WriteString(strconv.Quote(unit))
		b.WriteString(": ")
		b.WriteString(relativeTimeStyleMapLiteral(values[unit], table))
		b.WriteString(", ")
	}
	b.WriteString("}")
	return b.String()
}

func relativeTimeStyleMapLiteral(values map[string]cldr.RelativeTimeField, table *StringTable) string {
	return stringKeyedMapLiteral("map[string]RelativeTimeField", values, func(value cldr.RelativeTimeField) string {
		return relativeTimeFieldLiteral(value, table)
	})
}

func relativeTimeFieldLiteral(field cldr.RelativeTimeField, table *StringTable) string {
	return "RelativeTimeField{" +
		"Future: " + stringMapLiteral(field.Future, table) + ", " +
		"Past: " + stringMapLiteral(field.Past, table) + ", " +
		"Relative: " + stringMapLiteral(field.Relative, table) + "}"
}
