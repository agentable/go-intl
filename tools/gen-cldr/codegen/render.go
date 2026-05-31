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

type RuntimeInput struct {
	Manifest       ManifestInput
	Locales        extract.Locales
	LikelySubtags  extract.LikelySubtags
	Numbers        extract.Numbers
	Currencies     extract.CurrencyData
	Collations     []string
	LocaleMatching extract.LocaleMatching
	Dates          extract.Dates
	Preferences    extract.PreferenceData
	Metazones      extract.Metazones
	Units          extract.Units
	ListPatterns   extract.ListPatterns
	RelativeTime   extract.RelativeTimeFields
	DisplayNames   extract.DisplayNames
}

type generatedFile struct {
	name string
	src  []byte
}

func RenderRuntime(outDir string, input RuntimeInput) error {
	table := NewStringTable()
	files, err := renderRuntimeFiles(input, table)
	if err != nil {
		return err
	}
	stringsFile, err := renderStringTableFile(table)
	if err != nil {
		return err
	}
	files = append(files, stringsFile)
	for _, file := range files {
		path := filepath.Join(outDir, file.name)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o777); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
		if err := os.WriteFile(path, file.src, 0o666); err != nil {
			return fmt.Errorf("write %s: %w", file.name, err)
		}
	}
	return nil
}

func renderRuntimeFiles(input RuntimeInput, table *StringTable) ([]generatedFile, error) {
	var files []generatedFile
	for _, render := range []func(RuntimeInput, *StringTable) (generatedFile, error){
		renderManifestFile,
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
		renderListPatternsFile,
		renderRelativeTimeFile,
		renderDisplayNamesFile,
		renderTimezonesFile,
	} {
		file, err := render(input, table)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	wrappers, err := renderWrapperFiles()
	if err != nil {
		return nil, err
	}
	files = append(files, wrappers...)
	leafFiles, err := renderLeafCLDRFiles(input, table)
	if err != nil {
		return nil, err
	}
	files = append(files, leafFiles...)
	return files, nil
}

func renderLeafCLDRFiles(input RuntimeInput, table *StringTable) ([]generatedFile, error) {
	var files []generatedFile
	for _, leaf := range []struct {
		name        string
		packageName string
		render      func(RuntimeInput, *StringTable) (generatedFile, error)
	}{
		{name: "locale/locales.go", packageName: "cldrlocale", render: renderLocalesFile},
		{name: "locale/likely_subtags.go", packageName: "cldrlocale", render: renderLikelySubtagsFile},
		{name: "locale/preference.go", packageName: "cldrlocale", render: renderPreferenceFile},
		{name: "locale/collations.go", packageName: "cldrlocale", render: renderCollationsFile},
		{name: "locale/timezones.go", packageName: "cldrlocale", render: renderTimezonesFile},
		{name: "list/likely_subtags.go", packageName: "list", render: renderLikelySubtagsFile},
	} {
		file, err := leaf.render(input, table)
		if err != nil {
			return nil, err
		}
		src, err := replaceGeneratedPackage(file.src, leaf.packageName)
		if err != nil {
			return nil, err
		}
		files = append(files, generatedFile{name: leaf.name, src: src})
	}

	listPatterns, err := renderListPatternsLeaf(input.ListPatterns, table)
	if err != nil {
		return nil, err
	}
	files = append(files, generatedFile{name: "list/patterns.go", src: listPatterns})

	localeAccessors, err := FormatFile([]byte(`package cldrlocale

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
`))
	if err != nil {
		return nil, err
	}
	files = append(files, generatedFile{name: "locale/accessors.go", src: localeAccessors})

	localeNumbering, err := renderLocaleNumbering(input.Numbers)
	if err != nil {
		return nil, err
	}
	files = append(files, generatedFile{name: "locale/numbering.go", src: localeNumbering})

	listLocale, err := FormatFile([]byte(`package list

import "github.com/agentable/go-intl/internal/localeid"

func Maximize(tag string) string {
	return localeid.Maximize(tag, MaximizeSubtags)
}
`))
	if err != nil {
		return nil, err
	}
	files = append(files, generatedFile{name: "list/locale.go", src: listLocale})

	for _, stringFile := range []struct {
		name        string
		packageName string
	}{
		{name: "locale/strings.go", packageName: "cldrlocale"},
		{name: "list/strings.go", packageName: "list"},
	} {
		var b bytes.Buffer
		if err := table.EmitPackage(&b, stringFile.packageName); err != nil {
			return nil, err
		}
		src, err := FormatFile(b.Bytes())
		if err != nil {
			return nil, err
		}
		files = append(files, generatedFile{name: stringFile.name, src: src})
	}
	return files, nil
}

func replaceGeneratedPackage(src []byte, packageName string) ([]byte, error) {
	replaced := bytes.Replace(src, []byte("package cldr\n"), []byte("package "+packageName+"\n"), 1)
	return FormatFile(replaced)
}

func renderManifestFile(input RuntimeInput, _ *StringTable) (generatedFile, error) {
	src, err := renderManifest(input.Manifest)
	return generatedFile{name: "manifest.go", src: src}, err
}

func renderLocalesFile(input RuntimeInput, table *StringTable) (generatedFile, error) {
	src, err := renderLocales(input.Locales, table)
	return generatedFile{name: "locales.go", src: src}, err
}

func renderLikelySubtagsFile(input RuntimeInput, table *StringTable) (generatedFile, error) {
	src, err := renderLikely(input.LikelySubtags, table)
	return generatedFile{name: "likely_subtags.go", src: src}, err
}

func renderNumbersFile(input RuntimeInput, table *StringTable) (generatedFile, error) {
	src, err := renderNumbers(input.Numbers, table)
	return generatedFile{name: "numbers.go", src: src}, err
}

func renderCurrenciesFile(input RuntimeInput, table *StringTable) (generatedFile, error) {
	src, err := renderCurrencies(input.Currencies, table)
	return generatedFile{name: "currencies.go", src: src}, err
}

func renderCollationsFile(input RuntimeInput, _ *StringTable) (generatedFile, error) {
	src, err := renderCollations(input.Collations)
	return generatedFile{name: "collations.go", src: src}, err
}

func renderSupportedFile(input RuntimeInput, _ *StringTable) (generatedFile, error) {
	src, err := renderSupported(input)
	return generatedFile{name: "supported.go", src: src}, err
}

func renderLocaleMatchingFile(input RuntimeInput, table *StringTable) (generatedFile, error) {
	src, err := renderLocaleMatching(input.LocaleMatching, table)
	return generatedFile{name: "locale_matching.go", src: src}, err
}

func renderRegionsFile(input RuntimeInput, table *StringTable) (generatedFile, error) {
	src, err := renderRegions(input.LocaleMatching, table)
	return generatedFile{name: "regions.go", src: src}, err
}

func renderDatesFile(input RuntimeInput, table *StringTable) (generatedFile, error) {
	src, err := renderDates(input.Dates, table)
	return generatedFile{name: "dates.go", src: src}, err
}

func renderPreferenceFile(input RuntimeInput, table *StringTable) (generatedFile, error) {
	src, err := renderPreference(input.Preferences, table)
	return generatedFile{name: "preference.go", src: src}, err
}

func renderMetazonesFile(input RuntimeInput, table *StringTable) (generatedFile, error) {
	src, err := renderMetazones(input.Metazones, table)
	return generatedFile{name: "metazones.go", src: src}, err
}

func renderUnitsFile(input RuntimeInput, table *StringTable) (generatedFile, error) {
	src, err := renderUnits(input.Units, table)
	return generatedFile{name: "units.go", src: src}, err
}

func renderListPatternsFile(input RuntimeInput, table *StringTable) (generatedFile, error) {
	src, err := renderListPatterns(input.ListPatterns, table)
	return generatedFile{name: "list_patterns.go", src: src}, err
}

func renderRelativeTimeFile(input RuntimeInput, table *StringTable) (generatedFile, error) {
	src, err := renderRelativeTime(input.RelativeTime, table)
	return generatedFile{name: "relative_time.go", src: src}, err
}

func renderDisplayNamesFile(input RuntimeInput, _ *StringTable) (generatedFile, error) {
	src, err := renderDisplayNames(input.DisplayNames)
	return generatedFile{name: "displaynames/data.go", src: src}, err
}

func renderTimezonesFile(RuntimeInput, *StringTable) (generatedFile, error) {
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
	b.WriteString("import \"strings\"\n\n")
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
	b.WriteString(`func ResolveLocale(tag string) (Locale, bool) {
	if loc, ok := localeIndex[tag]; ok {
		return loc, true
	}
	if base, _, ok := strings.Cut(tag, "-"); ok {
		if loc, ok := localeIndex[base]; ok {
			return loc, true
		}
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
	b.WriteString("import \"sync\"\n\n")
	b.WriteString(`type maximizeSubtagRecord struct{ key, lang, script, region sliceRef }

type minimizeSubtagRecord struct{ lang, script, region, minimized sliceRef }

`)
	b.WriteString(`var (
	likelySubtagsOnce sync.Once
	likelySubtags     []maximizeSubtagRecord
	minimizeSubtags   []minimizeSubtagRecord
)

func ensureLikelySubtags() {
	likelySubtagsOnce.Do(loadLikelySubtags)
}

func loadLikelySubtags() {
	likelySubtags = []maximizeSubtagRecord{
`)
	for _, key := range slices.Sorted(maps.Keys(likely.Maximize)) {
		triple := likely.Maximize[key]
		b.WriteString("\t")
		b.WriteString("{")
		b.WriteString(table.Add(key).GoLiteral())
		b.WriteString(", ")
		b.WriteString(table.Add(triple.Lang).GoLiteral())
		b.WriteString(", ")
		b.WriteString(table.Add(triple.Script).GoLiteral())
		b.WriteString(", ")
		b.WriteString(table.Add(triple.Region).GoLiteral())
		b.WriteString("},\n")
	}
	b.WriteString("\t}\n\n")
	b.WriteString("\tminimizeSubtags = []minimizeSubtagRecord{\n")
	keys := slices.SortedFunc(maps.Keys(likely.Minimize), func(a, b extract.SubtagTriple) int {
		return strings.Compare(a.Lang+"-"+a.Script+"-"+a.Region, b.Lang+"-"+b.Script+"-"+b.Region)
	})
	for _, triple := range keys {
		minimized := likely.Minimize[triple]
		b.WriteString("\t{")
		b.WriteString(table.Add(triple.Lang).GoLiteral())
		b.WriteString(", ")
		b.WriteString(table.Add(triple.Script).GoLiteral())
		b.WriteString(", ")
		b.WriteString(table.Add(triple.Region).GoLiteral())
		b.WriteString(", ")
		b.WriteString(table.Add(minimized).GoLiteral())
		b.WriteString("},\n")
	}
	b.WriteString("\t}\n")
	b.WriteString("}\n\n")
	b.WriteString(`func MaximizeSubtags(language, script, region string) (lang, scr, reg string, ok bool) {
	ensureLikelySubtags()

	key := language
	if script != "" {
		key += "-" + script
	}
	if region != "" {
		key += "-" + region
	}
	i := searchLikelySubtag(key)
	if i >= len(likelySubtags) || likelySubtags[i].key.string() != key {
		return "", "", "", false
	}
	triple := likelySubtags[i]
	return triple.lang.string(), triple.script.string(), triple.region.string(), true
}

func MinimizeSubtags(language, script, region string) (lang, scr, reg string, ok bool) {
	ensureLikelySubtags()

	i := searchMinimizeSubtag(language, script, region)
	if i >= len(minimizeSubtags) || compareMinimizeSubtag(minimizeSubtags[i], language, script, region) != 0 {
		return "", "", "", false
	}
	return minimizeSubtags[i].minimized.string(), "", "", true
}

func searchLikelySubtag(key string) int {
	low, high := 0, len(likelySubtags)
	for low < high {
		mid := int(uint(low+high) >> 1)
		if likelySubtags[mid].key.string() < key {
			low = mid + 1
		} else {
			high = mid
		}
	}
	return low
}

func searchMinimizeSubtag(language, script, region string) int {
	low, high := 0, len(minimizeSubtags)
	for low < high {
		mid := int(uint(low+high) >> 1)
		if compareMinimizeSubtag(minimizeSubtags[mid], language, script, region) < 0 {
			low = mid + 1
		} else {
			high = mid
		}
	}
	return low
}

func compareMinimizeSubtag(row minimizeSubtagRecord, language, script, region string) int {
	if lang := row.lang.string(); lang != language {
		if lang < language {
			return -1
		}
		return 1
	}
	if scr := row.script.string(); scr != script {
		if scr < script {
			return -1
		}
		return 1
	}
	if reg := row.region.string(); reg != region {
		if reg < region {
			return -1
		}
		return 1
	}
	return 0
}
`)
	return FormatFile([]byte(b.String()))
}

func renderTimezones() ([]byte, error) {
	return FormatFile([]byte(`package cldr

import (
	"slices"
	"sync"
)

var (
	timeZoneDataOnce      sync.Once
	canonicalTimeZoneLinks map[string]string
	timeZonesByRegion      map[string][]string
)

func loadTimeZoneData() {
	canonicalTimeZoneLinks = map[string]string{
		"US/Eastern": "America/New_York",
		"US/Pacific": "America/Los_Angeles",
		"Asia/Calcutta": "Asia/Kolkata",
	}

	timeZonesByRegion = map[string][]string{
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
}

func CanonicalTimeZoneLink(name string) string {
	timeZoneDataOnce.Do(loadTimeZoneData)
	if canonical := canonicalTimeZoneLinks[name]; canonical != "" {
		return canonical
	}
	return name
}

func TimeZonesForRegion(region string) []string {
	timeZoneDataOnce.Do(loadTimeZoneData)
	zones, ok := timeZonesByRegion[region]
	if !ok {
		return nil
	}
	return slices.Clone(zones)
}
`))
}
