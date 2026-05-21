package codegen

import (
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

func renderListPatterns(data extract.ListPatterns, table *StringTable) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldr\n\n")
	b.WriteString("import \"sync\"\n\n")
	b.WriteString("type ListPattern struct{ Pair, Start, Middle, End string }\n\n")
	b.WriteString("type listPatternRefs struct{ pair, start, middle, end string }\n\n")
	b.WriteString(`var (
	listPatternsByLocaleOnce sync.Once
	listPatternsByLocale     map[Locale]map[string]map[string]listPatternRefs
	listSupportedLocalesOnce sync.Once
	listSupportedLocales     []string
)

func listPatternDataByLocale() map[Locale]map[string]map[string]listPatternRefs {
	listPatternsByLocaleOnce.Do(loadListPatternData)
	return listPatternsByLocale
}

func loadListPatternData() {
	listPatternsByLocale = map[Locale]map[string]map[string]listPatternRefs{
`)
	for _, locale := range sortedLocaleKeys(data) {
		b.WriteString("\tlocaleIndex[")
		b.WriteString(strconv.Quote(locale))
		b.WriteString("]: ")
		b.WriteString(listPatternTypeMapLiteral(data[locale], table))
		b.WriteString(",\n")
	}
	b.WriteString("\t}\n")
	b.WriteString("}\n\n")
	b.WriteString(`func ListSupportedLocales() []string {
	listSupportedLocalesOnce.Do(func() {
		listSupportedLocales = supportedLocaleTags(listPatternDataByLocale())
	})
	return listSupportedLocales
}

func (l Locale) ListPattern(typ, style string) ListPattern {
	if typ == "" {
		typ = "conjunction"
	}
	if style == "" {
		style = "long"
	}
	data := listPatternDataByLocale()[l][typ][style]
	return ListPattern{
		Pair:   data.pair,
		Start:  data.start,
		Middle: data.middle,
		End:    data.end,
	}
}
`)
	return FormatFile([]byte(b.String()))
}

func renderListPatternsLeaf(data extract.ListPatterns, table *StringTable) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package list\n\n")
	b.WriteString("import (\n\t\"slices\"\n\t\"strings\"\n\t\"sync\"\n)\n\n")
	b.WriteString("type ListPattern struct{ Pair, Start, Middle, End string }\n\n")
	b.WriteString("type Locale string\n\n")
	b.WriteString("type listPatternRefs struct{ pair, start, middle, end string }\n\n")
	b.WriteString(`var (
	listPatternsByLocaleOnce sync.Once
	listPatternsByLocale     map[string]map[string]map[string]listPatternRefs
	listSupportedLocalesOnce sync.Once
	listSupportedLocales     []string
)

func listPatternDataByLocale() map[string]map[string]map[string]listPatternRefs {
	listPatternsByLocaleOnce.Do(loadListPatternData)
	return listPatternsByLocale
}

func loadListPatternData() {
	listPatternsByLocale = map[string]map[string]map[string]listPatternRefs{
`)
	for _, locale := range sortedLocaleKeys(data) {
		b.WriteString("\t")
		b.WriteString(strconv.Quote(locale))
		b.WriteString(": ")
		b.WriteString(listPatternTypeMapLiteral(data[locale], table))
		b.WriteString(",\n")
	}
	b.WriteString("\t}\n")
	b.WriteString("}\n\n")
	b.WriteString(`func ListSupportedLocales() []string {
	listSupportedLocalesOnce.Do(func() {
		listSupportedLocales = supportedLocales()
	})
	return listSupportedLocales
}

func SupportedLocales() []string {
	return ListSupportedLocales()
}

func ResolveLocale(tag string) (Locale, bool) {
	data := listPatternDataByLocale()
	if _, ok := data[tag]; ok {
		return Locale(tag), true
	}
	if base, _, ok := strings.Cut(tag, "-"); ok {
		if _, ok := data[base]; ok {
			return Locale(base), true
		}
	}
	return "", false
}

func Pattern(locale Locale, typ, style string) ListPattern {
	if typ == "" {
		typ = "conjunction"
	}
	if style == "" {
		style = "long"
	}
	data := listPatternDataByLocale()[string(locale)][typ][style]
	return ListPattern{
		Pair:   data.pair,
		Start:  data.start,
		Middle: data.middle,
		End:    data.end,
	}
}

func supportedLocales() []string {
	data := listPatternDataByLocale()
	tags := make([]string, 0, len(data))
	for tag := range data {
		tags = append(tags, tag)
	}
	slices.Sort(tags)
	return tags
}
`)
	return FormatFile([]byte(b.String()))
}

func listPatternTypeMapLiteral(values cldr.ListPatterns, table *StringTable) string {
	if len(values) == 0 {
		return "nil"
	}
	var b strings.Builder
	b.WriteString("map[string]map[string]listPatternRefs{")
	for _, typ := range slices.Sorted(maps.Keys(values)) {
		b.WriteString(strconv.Quote(typ))
		b.WriteString(": ")
		b.WriteString(listPatternStyleMapLiteral(values[typ], table))
		b.WriteString(", ")
	}
	b.WriteString("}")
	return b.String()
}

func listPatternStyleMapLiteral(values map[string]cldr.ListPattern, table *StringTable) string {
	return stringKeyedMapLiteral("map[string]listPatternRefs", values, func(value cldr.ListPattern) string {
		return listPatternLiteral(value, table)
	})
}

func listPatternLiteral(pattern cldr.ListPattern, table *StringTable) string {
	return "listPatternRefs{" +
		"pair: " + refStringLiteral(pattern.Pair, table) + ", " +
		"start: " + refStringLiteral(pattern.Start, table) + ", " +
		"middle: " + refStringLiteral(pattern.Middle, table) + ", " +
		"end: " + refStringLiteral(pattern.End, table) + "}"
}
