package codegen

import (
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

func renderDisplayNames(data extract.DisplayNames) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package displaynames\n\n")
	b.WriteString("import \"sync\"\n\n")
	b.WriteString("type styledNames struct{ long, short, narrow map[string]string }\n\n")
	b.WriteString("type languageDisplay struct{ dialect, standard styledNames }\n\n")
	b.WriteString(`type localeData struct {
	languages      languageDisplay
	territories    styledNames
	scripts        styledNames
	calendars      styledNames
	dateTimeFields styledNames
	localePattern  string
}

var (
	displayNamesDataOnce sync.Once
	dataByLocale         map[string]localeData
	supportedLocales     []string
)

func displayNamesData() map[string]localeData {
	displayNamesDataOnce.Do(loadDisplayNamesData)
	return dataByLocale
}

func displayNamesSupportedLocales() []string {
	displayNamesDataOnce.Do(loadDisplayNamesData)
	return supportedLocales
}

func loadDisplayNamesData() {
	dataByLocale = map[string]localeData{
`)
	keys := slices.Sorted(maps.Keys(data))
	for _, locale := range keys {
		b.WriteString("\t\t")
		b.WriteString(strconv.Quote(locale))
		b.WriteString(": ")
		b.WriteString(localeDataLiteral(data[locale]))
		b.WriteString(",\n")
	}
	b.WriteString("\t}\n")
	b.WriteString("\tsupportedLocales = []string{\n")
	for _, locale := range keys {
		b.WriteString("\t\t")
		b.WriteString(strconv.Quote(locale))
		b.WriteString(",\n")
	}
	b.WriteString("\t}\n")
	b.WriteString("}\n")
	return FormatFile([]byte(b.String()))
}

func localeDataLiteral(d cldr.DisplayNames) string {
	var b strings.Builder
	b.WriteString("{")
	b.WriteString("languages: languageDisplay{dialect: ")
	b.WriteString(styledNamesLiteral(d.Languages.Dialect))
	b.WriteString(", standard: ")
	b.WriteString(styledNamesLiteral(d.Languages.Standard))
	b.WriteString("}, territories: ")
	b.WriteString(styledNamesLiteral(d.Territories))
	b.WriteString(", scripts: ")
	b.WriteString(styledNamesLiteral(d.Scripts))
	b.WriteString(", calendars: ")
	b.WriteString(styledNamesLiteral(d.Calendars))
	b.WriteString(", dateTimeFields: ")
	b.WriteString(styledNamesLiteral(d.DateTimeFields))
	b.WriteString(", localePattern: ")
	b.WriteString(strconv.Quote(d.LocalePattern))
	b.WriteString("}")
	return b.String()
}

func styledNamesLiteral(s cldr.StyledNames) string {
	return "styledNames{long: " + plainStringMapLiteral(s.Long) +
		", short: " + plainStringMapLiteral(s.Short) +
		", narrow: " + plainStringMapLiteral(s.Narrow) + "}"
}

func plainStringMapLiteral(m map[string]string) string {
	return stringKeyedMapLiteral("map[string]string", m, strconv.Quote)
}
