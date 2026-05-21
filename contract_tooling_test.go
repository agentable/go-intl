package gointl

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestConformanceHelpersStayOutOfRuntime(t *testing.T) {
	t.Parallel()

	const conformanceImport = "github.com/agentable/go-intl/tools/conformance"
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch path {
			case ".git", ".references", ".tmp", "bin", "tools/gen-cldr/.cldr-json":
				return filepath.SkipDir
			}
			if path != "." && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") || strings.HasPrefix(path, "tools/") {
			return nil
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imported := range file.Imports {
			if strings.Trim(imported.Path.Value, `"`) == conformanceImport {
				t.Fatalf("%s imports %s; conformance helpers must stay in tests and tools", path, conformanceImport)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
