package codegen

import (
	"strconv"
	"strings"

	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

func renderLocaleNumbering(numbers extract.Numbers) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldrlocale\n\n")
	b.WriteString("func (l Locale) DefaultNumberingSystem() string {\n")
	b.WriteString("\tif int(l) >= len(localeRecords) {\n")
	b.WriteString("\t\treturn \"latn\"\n")
	b.WriteString("\t}\n")
	b.WriteString("\tswitch localeRecords[l].tag.string() {\n")
	for _, locale := range sortedLocaleKeys(numbers) {
		ns := numbers[locale].DefaultNumberingSystem
		if ns == "" || ns == "latn" {
			continue
		}
		b.WriteString("\tcase ")
		b.WriteString(strconv.Quote(locale))
		b.WriteString(":\n")
		b.WriteString("\t\treturn ")
		b.WriteString(strconv.Quote(ns))
		b.WriteString("\n")
	}
	b.WriteString("\tdefault:\n")
	b.WriteString("\t\treturn \"latn\"\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n")
	return FormatFile([]byte(b.String()))
}
