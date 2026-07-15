package codegen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

func RenderUnicodeTypeAliases(path string, aliases []cldr.UnicodeTypeAlias) error {
	src, err := renderUnicodeTypeAliases(aliases)
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

func renderUnicodeTypeAliases(aliases []cldr.UnicodeTypeAlias) ([]byte, error) {
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
	return FormatFile(body.Bytes())
}
