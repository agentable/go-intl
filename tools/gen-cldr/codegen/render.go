package codegen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

type RuntimeInput struct {
	Manifest      ManifestInput
	Locales       extract.Locales
	LikelySubtags extract.LikelySubtags
	Numbers       extract.Numbers
	Currencies    extract.CurrencyData
	Collations    []string
	Dates         extract.Dates
	Preferences   extract.PreferenceData
	Metazones     extract.Metazones
	Units         extract.Units
	ListPatterns  extract.ListPatterns
	RelativeTime  extract.RelativeTimeFields
	DisplayNames  extract.DisplayNames
}

type generatedFile struct {
	name string
	src  []byte
}

func RenderRuntime(outDir string, input RuntimeInput) error {
	files, err := renderRuntimeFiles(input, nil)
	if err != nil {
		return err
	}
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

// renderRuntimeFiles emits the generated CLDR data. The root internal/cldr
// package is retired: every domain (including the locale kernel) owns its data
// directly, so the generator no longer writes any package cldr file. The root
// string table is gone too — each domain carries its own private _data table.
func renderRuntimeFiles(input RuntimeInput, _ *StringTable) ([]generatedFile, error) {
	files, err := renderLeafCLDRFiles(input)
	if err != nil {
		return nil, err
	}
	domainFiles, err := renderDomainFiles(input)
	if err != nil {
		return nil, err
	}
	files = append(files, domainFiles...)
	return files, nil
}

// renderLeafCLDRFiles renders the locale kernel package (cldrlocale). The kernel
// follows the same const-only-payload form as the leaf domains: a generated
// data.go holds _data plus the kernel blobs, and the hand-written decode.go and
// accessors.go expand and query them. The registry data and likely-subtags bomb
// that formerly lived in literal locales.go / likely_subtags.go / preference.go /
// strings.go are dissolved into that single payload; the kernel _data holds only
// the kernel's own strings.
//
// collations.go and timezones.go stay small const/literal files (no compile-bomb
// shape), so they are rendered into the kernel package. numbering values move
// into the kernel payload, so renderLocaleNumbering is retired. The data
// manifest (CLDR/ICU/tzdata pin and locale profile) is owned by the kernel too,
// so manifest.go is rendered here with the kernel package name.
func renderLeafCLDRFiles(input RuntimeInput) ([]generatedFile, error) {
	var files []generatedFile

	localeData, err := encodeLocaleKernel(input, NewStringTable())
	if err != nil {
		return nil, err
	}
	files = append(files, generatedFile{name: "locale/data.go", src: localeData})

	for _, leaf := range []struct {
		name   string
		render func(RuntimeInput) ([]byte, error)
	}{
		{name: "locale/manifest.go", render: func(in RuntimeInput) ([]byte, error) { return renderManifest(in.Manifest) }},
		{name: "locale/collations.go", render: func(in RuntimeInput) ([]byte, error) { return renderCollations(in.Collations) }},
		{name: "locale/timezones.go", render: func(RuntimeInput) ([]byte, error) { return renderTimezones() }},
	} {
		src, err := leaf.render(input)
		if err != nil {
			return nil, err
		}
		src, err = replaceGeneratedPackage(src, "cldrlocale")
		if err != nil {
			return nil, err
		}
		files = append(files, generatedFile{name: leaf.name, src: src})
	}
	return files, nil
}

func replaceGeneratedPackage(src []byte, packageName string) ([]byte, error) {
	replaced := bytes.Replace(src, []byte("package cldr\n"), []byte("package "+packageName+"\n"), 1)
	return FormatFile(replaced)
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
