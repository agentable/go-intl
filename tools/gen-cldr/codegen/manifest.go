package codegen

import (
	"strconv"
	"strings"
)

type ManifestInput struct {
	Generator     string
	CLDR          string
	ICU           string
	TZData        string
	LocaleProfile []string
	InputHashes   []ManifestHash
}

type ManifestHash struct {
	Name   string
	SHA256 string
}

func renderManifest(input ManifestInput) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldr\n\n")
	b.WriteString("import \"slices\"\n\n")
	b.WriteString("type DataManifest struct {\n")
	b.WriteString("\tGenerator string\n")
	b.WriteString("\tCLDR string\n")
	b.WriteString("\tICU string\n")
	b.WriteString("\tTZData string\n")
	b.WriteString("\tLocaleProfile []string\n")
	b.WriteString("\tInputHashes []InputHash\n")
	b.WriteString("}\n\n")
	b.WriteString("type InputHash struct {\n")
	b.WriteString("\tName string\n")
	b.WriteString("\tSHA256 string\n")
	b.WriteString("}\n\n")
	b.WriteString("var dataManifest = DataManifest{\n")
	b.WriteString("\tGenerator: ")
	b.WriteString(strconv.Quote(input.Generator))
	b.WriteString(",\n\tCLDR: ")
	b.WriteString(strconv.Quote(input.CLDR))
	b.WriteString(",\n\tICU: ")
	b.WriteString(strconv.Quote(input.ICU))
	b.WriteString(",\n\tTZData: ")
	b.WriteString(strconv.Quote(input.TZData))
	b.WriteString(",\n\tLocaleProfile: []string{\n")
	for _, locale := range input.LocaleProfile {
		b.WriteString("\t\t")
		b.WriteString(strconv.Quote(locale))
		b.WriteString(",\n")
	}
	b.WriteString("\t},\n\tInputHashes: []InputHash{\n")
	for _, hash := range input.InputHashes {
		b.WriteString("\t\t{Name: ")
		b.WriteString(strconv.Quote(hash.Name))
		b.WriteString(", SHA256: ")
		b.WriteString(strconv.Quote(hash.SHA256))
		b.WriteString("},\n")
	}
	b.WriteString("\t},\n}\n\n")
	b.WriteString(`func Manifest() DataManifest {
	manifest := dataManifest
	manifest.LocaleProfile = slices.Clone(dataManifest.LocaleProfile)
	manifest.InputHashes = slices.Clone(dataManifest.InputHashes)
	return manifest
}
`)
	return FormatFile([]byte(b.String()))
}
