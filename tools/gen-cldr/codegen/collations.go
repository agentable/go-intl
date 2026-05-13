package codegen

import (
	"fmt"
	"strings"
)

func renderCollations(collations []string) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldr\n\n")
	b.WriteString("var collationValues = []string{\n")
	for _, collation := range collations {
		fmt.Fprintf(&b, "\t%q,\n", collation)
	}
	b.WriteString("}\n")
	return FormatFile([]byte(b.String()))
}
