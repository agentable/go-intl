package ecma402dtf

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
)

func TestPackageDoesNotImportCLDR(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if file == "guardrail_test.go" {
			continue
		}
		t.Run(file, func(t *testing.T) {
			t.Parallel()

			parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatal(err)
			}
			for _, spec := range parsed.Imports {
				if spec.Path.Value == "\"github.com/agentable/go-intl/internal/cldr\"" {
					t.Fatalf("%s imports internal/cldr", file)
				}
			}
		})
	}
}

func TestPackageDoesNotUseInit(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if file == "guardrail_test.go" {
			continue
		}
		t.Run(file, func(t *testing.T) {
			t.Parallel()

			parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, decl := range parsed.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if ok && fn.Name.Name == "init" {
					t.Fatalf("%s declares init", file)
				}
			}
		})
	}
}
