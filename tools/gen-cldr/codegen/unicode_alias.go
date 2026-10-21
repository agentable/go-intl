package codegen

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

func RenderUnicodeTypeAliases(path string, aliases []cldr.UnicodeTypeAlias, identities cldr.IdentityAliases) error {
	src, err := renderUnicodeTypeAliases(aliases, identities)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, src, 0o666); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func renderUnicodeTypeAliases(aliases []cldr.UnicodeTypeAlias, identities cldr.IdentityAliases) ([]byte, error) {
	var body bytes.Buffer
	if _, err := body.WriteString("package localeid\n\nvar unicodeTypeAliases = [...]unicodeTypeAliasRecord{\n"); err != nil {
		return nil, err
	}
	for _, alias := range aliases {
		if _, err := fmt.Fprintf(&body, "\t{key: %q, alias: %q, canonical: %q},\n", alias.Key, alias.Alias, alias.Canonical); err != nil {
			return nil, err
		}
	}
	if _, err := body.WriteString("}\n"); err != nil {
		return nil, err
	}

	groups := map[string][]string{}
	for _, from := range slices.Sorted(maps.Keys(identities["languageAlias"])) {
		lang, _, _ := strings.Cut(from, "-")
		groups[lang] = append(groups[lang], from)
	}
	body.WriteString("\nvar languageAliases = map[string][]languageAliasRecord{\n")
	for _, lang := range slices.Sorted(maps.Keys(groups)) {
		fmt.Fprintf(&body, "%q: {\n", lang)
		for _, from := range groups[lang] {
			fmt.Fprintf(&body, "{from: %q, to: %q},\n", from, identities["languageAlias"][from])
		}
		body.WriteString("},\n")
	}
	body.WriteString("}\n")
	for _, kind := range []string{"scriptAlias", "territoryAlias", "variantAlias", "subdivisionAlias"} {
		fmt.Fprintf(&body, "\nvar %s = map[string]string{\n", kind)
		for _, from := range slices.Sorted(maps.Keys(identities[kind])) {
			fmt.Fprintf(&body, "%q: %q,\n", from, identities[kind][from])
		}
		body.WriteString("}\n")
	}
	return FormatFile(body.Bytes())
}
