package codegen

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

type Phase3Data struct {
	Locales    extract.Locales
	Likely     extract.LikelySubtags
	Numbers    extract.Numbers
	Currencies extract.CurrencyData
	Collations []string
	Matching   extract.LocaleMatching
	Dates      extract.Dates
	Preference extract.PreferenceData
	Metazones  extract.Metazones
	Units      extract.Units
}

type generatedFile struct {
	name string
	src  []byte
}

func RenderPhase3(outDir string, data Phase3Data) error {
	table := NewStringTable()
	files, err := renderPhase3Files(data, table)
	if err != nil {
		return err
	}
	stringsFile, err := renderStringTableFile(table)
	if err != nil {
		return err
	}
	files = append(files, stringsFile)
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(outDir, file.name), file.src, 0o666); err != nil {
			return fmt.Errorf("write %s: %w", file.name, err)
		}
	}
	return nil
}

func renderPhase3Files(data Phase3Data, table *StringTable) ([]generatedFile, error) {
	var files []generatedFile
	for _, render := range []func(Phase3Data, *StringTable) (generatedFile, error){
		renderLocalesFile,
		renderLikelySubtagsFile,
		renderNumbersFile,
		renderCurrenciesFile,
		renderCollationsFile,
		renderSupportedFile,
		renderLocaleMatchingFile,
		renderRegionsFile,
		renderDatesFile,
		renderPreferenceFile,
		renderMetazonesFile,
		renderUnitsFile,
		renderTimezonesFile,
	} {
		file, err := render(data, table)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func renderLocalesFile(data Phase3Data, table *StringTable) (generatedFile, error) {
	src, err := renderLocales(data.Locales, table)
	return generatedFile{name: "locales.go", src: src}, err
}

func renderLikelySubtagsFile(data Phase3Data, table *StringTable) (generatedFile, error) {
	src, err := renderLikely(data.Likely, table)
	return generatedFile{name: "likely_subtags.go", src: src}, err
}

func renderNumbersFile(data Phase3Data, table *StringTable) (generatedFile, error) {
	src, err := renderNumbers(data.Numbers, table)
	return generatedFile{name: "numbers.go", src: src}, err
}

func renderCurrenciesFile(data Phase3Data, table *StringTable) (generatedFile, error) {
	src, err := renderCurrencies(data.Currencies, table)
	return generatedFile{name: "currencies.go", src: src}, err
}

func renderCollationsFile(data Phase3Data, _ *StringTable) (generatedFile, error) {
	src, err := renderCollations(data.Collations)
	return generatedFile{name: "collations.go", src: src}, err
}

func renderSupportedFile(Phase3Data, *StringTable) (generatedFile, error) {
	src, err := renderSupported()
	return generatedFile{name: "supported.go", src: src}, err
}

func renderLocaleMatchingFile(data Phase3Data, table *StringTable) (generatedFile, error) {
	src, err := renderLocaleMatching(data.Matching, table)
	return generatedFile{name: "locale_matching.go", src: src}, err
}

func renderRegionsFile(data Phase3Data, table *StringTable) (generatedFile, error) {
	src, err := renderRegions(data.Matching, table)
	return generatedFile{name: "regions.go", src: src}, err
}

func renderDatesFile(data Phase3Data, table *StringTable) (generatedFile, error) {
	src, err := renderDates(data.Dates, table)
	return generatedFile{name: "dates.go", src: src}, err
}

func renderPreferenceFile(data Phase3Data, table *StringTable) (generatedFile, error) {
	src, err := renderPreference(data.Preference, table)
	return generatedFile{name: "preference.go", src: src}, err
}

func renderMetazonesFile(data Phase3Data, table *StringTable) (generatedFile, error) {
	src, err := renderMetazones(data.Metazones, table)
	return generatedFile{name: "metazones.go", src: src}, err
}

func renderUnitsFile(data Phase3Data, table *StringTable) (generatedFile, error) {
	src, err := renderUnits(data.Units, table)
	return generatedFile{name: "units.go", src: src}, err
}

func renderTimezonesFile(Phase3Data, *StringTable) (generatedFile, error) {
	src, err := renderTimezones()
	return generatedFile{name: "timezones.go", src: src}, err
}

func renderStringTableFile(table *StringTable) (generatedFile, error) {
	var b bytes.Buffer
	if err := table.Emit(&b); err != nil {
		return generatedFile{}, err
	}
	src, err := FormatFile(b.Bytes())
	if err != nil {
		return generatedFile{}, err
	}
	return generatedFile{name: "strings.go", src: src}, nil
}

func renderLocales(locales extract.Locales, table *StringTable) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldr\n\n")
	b.WriteString("import \"golang.org/x/text/language\"\n\n")
	b.WriteString("type localeRecord struct{ tag sliceRef }\n\n")
	b.WriteString("var localeRecords = [...]localeRecord{\n")
	for _, tag := range locales.Tags {
		b.WriteString("\t{tag: ")
		b.WriteString(table.Add(tag).GoLiteral())
		b.WriteString("}, // ")
		b.WriteString(tag)
		b.WriteString("\n")
	}
	b.WriteString("}\n\n")
	b.WriteString("var localeIndex = map[string]Locale{\n")
	for i, tag := range locales.Tags {
		b.WriteString("\t")
		b.WriteString(strconv.Quote(tag))
		b.WriteString(": ")
		b.WriteString(fmt.Sprintf("Locale(%d),\n", i))
	}
	b.WriteString("}\n\n")
	b.WriteString("var availableLocaleTags = []string{\n")
	for _, tag := range locales.Tags {
		b.WriteString("\t")
		b.WriteString(strconv.Quote(tag))
		b.WriteString(",\n")
	}
	b.WriteString("}\n\n")
	b.WriteString(`func ResolveLocale(tag language.Tag) (Locale, bool) {
	if loc, ok := localeIndex[tag.String()]; ok {
		return loc, true
	}
	base, _ := tag.Base()
	if loc, ok := localeIndex[base.String()]; ok {
		return loc, true
	}
	return Undefined, false
}

func AvailableLocales() []string { return availableLocaleTags }
`)
	return FormatFile([]byte(b.String()))
}

func renderLikely(likely extract.LikelySubtags, table *StringTable) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldr\n\n")
	b.WriteString("type subtagTriple struct{ lang, script, region sliceRef }\n\n")
	b.WriteString("var likelySubtags = map[string]subtagTriple{\n")
	for _, key := range slices.Sorted(maps.Keys(likely.Maximize)) {
		triple := likely.Maximize[key]
		b.WriteString("\t")
		b.WriteString(strconv.Quote(key))
		b.WriteString(": {lang: ")
		b.WriteString(table.Add(triple.Lang).GoLiteral())
		b.WriteString(", script: ")
		b.WriteString(table.Add(triple.Script).GoLiteral())
		b.WriteString(", region: ")
		b.WriteString(table.Add(triple.Region).GoLiteral())
		b.WriteString("}, // ")
		b.WriteString(triple.Lang + "_" + triple.Script + "_" + triple.Region)
		b.WriteString("\n")
	}
	b.WriteString("}\n\n")
	b.WriteString("var minimizeSubtags = map[subtagTriple]string{\n")
	keys := slices.SortedFunc(maps.Keys(likely.Minimize), func(a, b extract.SubtagTriple) int {
		return strings.Compare(a.Lang+"-"+a.Script+"-"+a.Region, b.Lang+"-"+b.Script+"-"+b.Region)
	})
	for _, triple := range keys {
		b.WriteString("\t{lang: ")
		b.WriteString(table.Add(triple.Lang).GoLiteral())
		b.WriteString(", script: ")
		b.WriteString(table.Add(triple.Script).GoLiteral())
		b.WriteString(", region: ")
		b.WriteString(table.Add(triple.Region).GoLiteral())
		b.WriteString("}: ")
		b.WriteString(strconv.Quote(likely.Minimize[triple]))
		b.WriteString(",\n")
	}
	b.WriteString("}\n\n")
	b.WriteString(`func MaximizeSubtags(language, script, region string) (lang, scr, reg string, ok bool) {
	key := language
	if script != "" {
		key += "-" + script
	}
	if region != "" {
		key += "-" + region
	}
	triple, ok := likelySubtags[key]
	if !ok {
		return "", "", "", false
	}
	return triple.lang.string(), triple.script.string(), triple.region.string(), true
}

func MinimizeSubtags(language, script, region string) (lang, scr, reg string, ok bool) {
	key := subtagTriple{lang: sliceRef{}, script: sliceRef{}, region: sliceRef{}}
	for triple := range minimizeSubtags {
		if triple.lang.string() == language && triple.script.string() == script && triple.region.string() == region {
			minimized := minimizeSubtags[triple]
			return minimized, "", "", true
		}
	}
	_ = key
	return "", "", "", false
}
`)
	return FormatFile([]byte(b.String()))
}

func renderTimezones() ([]byte, error) {
	return FormatFile([]byte(`package cldr

import "slices"

var canonicalTimeZoneLinks = map[string]string{
	"US/Eastern": "America/New_York",
	"US/Pacific": "America/Los_Angeles",
	"Asia/Calcutta": "Asia/Kolkata",
}

var timeZonesByRegion = map[string][]string{
	"BR": {
		"America/Araguaina", "America/Bahia", "America/Belem", "America/Boa_Vista",
		"America/Campo_Grande", "America/Cuiaba", "America/Eirunepe", "America/Fortaleza",
		"America/Maceio", "America/Manaus", "America/Noronha", "America/Porto_Velho",
		"America/Recife", "America/Rio_Branco", "America/Santarem", "America/Sao_Paulo",
	},
	"CN": {"Asia/Shanghai", "Asia/Urumqi"},
	"EG": {"Africa/Cairo"},
	"GB": {"Europe/London"},
	"IN": {"Asia/Calcutta"},
	"SA": {"Asia/Riyadh"},
	"US": {
		"America/Adak", "America/Anchorage", "America/Boise", "America/Chicago",
		"America/Denver", "America/Detroit", "America/Indiana/Knox", "America/Indiana/Marengo",
		"America/Indiana/Petersburg", "America/Indiana/Tell_City", "America/Indiana/Vevay",
		"America/Indiana/Vincennes", "America/Indiana/Winamac", "America/Indianapolis",
		"America/Juneau", "America/Kentucky/Monticello", "America/Los_Angeles", "America/Louisville",
		"America/Menominee", "America/Metlakatla", "America/New_York", "America/Nome",
		"America/North_Dakota/Beulah", "America/North_Dakota/Center", "America/North_Dakota/New_Salem",
		"America/Phoenix", "America/Sitka", "America/Yakutat", "Pacific/Honolulu",
	},
}

func CanonicalTimeZoneLink(name string) string {
	if canonical := canonicalTimeZoneLinks[name]; canonical != "" {
		return canonical
	}
	return name
}

func TimeZonesForRegion(region string) []string {
	zones, ok := timeZonesByRegion[region]
	if !ok {
		return nil
	}
	return slices.Clone(zones)
}
`))
}
