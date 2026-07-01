package cldrlocale_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Data-shape gate: test the cause of the compile-memory blowup, not the effect.
//
// Two absolute, threshold-free invariants over every generated Go file in the
// repository, identified by the generator header comment:
//
//   - Rule A (payload files, path internal/cldr/<domain>/data.go): only const
//     declarations are allowed. Zero func, zero var, zero non-const expression,
//     zero import. This is the representation invariant the migration
//     converges on.
//   - Rule B (all other generated files): no CallExpr, IndexExpr, or
//     IndexListExpr may appear inside any composite literal. This kills the
//     compile bomb shape (makeUnitPatternKey(localeIndex["af"], …)) at the
//     data-literal level.
//
// The gate asserts the absolute shape; it never measures a count and carries no
// exemptions.

func TestGeneratedDataShape(t *testing.T) {
	t.Parallel()

	root := filepath.Clean("../../..")
	files := generatedGoFiles(t, root)
	if len(files) == 0 {
		t.Fatal("no generated Go files discovered; data-shape gate would be vacuous")
	}

	for _, path := range files {
		rel := filepath.ToSlash(mustRel(t, root, path))
		violation := generatedFileShapeViolation(t, path, rel)
		t.Run(rel, func(t *testing.T) {
			t.Parallel()
			if violation != "" {
				t.Errorf("rule %s violated in generated file %s", violation, rel)
			}
		})
	}
}

// generatedFileShapeViolation parses one generated file and returns the rule
// name it violates, or "" if the file satisfies its applicable invariant. Rule A
// applies to payload files (internal/cldr/<domain>/data.go); Rule B applies to
// every other generated file.
func generatedFileShapeViolation(t *testing.T, path, rel string) string {
	t.Helper()

	const (
		ruleA = "A (const-only payload)"
		ruleB = "B (no call or index expressions in composite literals)"
	)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse %s: %v", rel, err)
	}

	if isPayloadDataFile(rel) {
		if !isConstOnly(f) {
			return ruleA
		}
		return ""
	}

	if compositeLiteralHasCallOrIndex(f) {
		return ruleB
	}
	return ""
}

// isPayloadDataFile reports whether rel is a domain payload file of the form
// internal/cldr/<domain>/data.go (exactly one path segment between the cldr
// root and the data.go leaf).
func isPayloadDataFile(rel string) bool {
	const prefix = "internal/cldr/"
	const leaf = "/data.go"
	if !strings.HasPrefix(rel, prefix) || !strings.HasSuffix(rel, leaf) {
		return false
	}
	domain := rel[len(prefix) : len(rel)-len(leaf)]
	return domain != "" && !strings.Contains(domain, "/")
}

// isConstOnly reports whether the file contains only const GenDecls: no
// imports, no var, no type, no func — the const-only representation shape.
func isConstOnly(f *ast.File) bool {
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			return false // FuncDecl
		}
		if gen.Tok != token.CONST {
			return false // import, var, or type
		}
	}
	return true
}

// compositeLiteralHasCallOrIndex reports whether any composite literal in the
// file contains a CallExpr, IndexExpr, or IndexListExpr (generic
// instantiation) — the compile-bomb literal shape.
func compositeLiteralHasCallOrIndex(f *ast.File) bool {
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		ast.Inspect(cl, func(m ast.Node) bool {
			switch m.(type) {
			case *ast.CallExpr, *ast.IndexExpr, *ast.IndexListExpr:
				found = true
				return false
			}
			return true
		})
		return true
	})
	return found
}

// generatedGoFiles walks the repository from root and returns every generated
// Go source file, identified by the generator header comment. Dot-directories
// (.git, .tmp, .references) and the tools/ generator module trees are skipped:
// the gate governs committed generated data, not transient build artifacts or
// the generators themselves.
func generatedGoFiles(t *testing.T, root string) []string {
	t.Helper()

	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if path != root && (strings.HasPrefix(name, ".") || name == "tools") {
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		if isGeneratedByHeader(t, path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repository: %v", err)
	}
	slices.Sort(files)
	return files
}

// isGeneratedByHeader reports whether the file carries a recognized generator
// header comment near its top. It parses comments only, avoiding a full parse
// for the discovery pass.
func isGeneratedByHeader(t *testing.T, path string) bool {
	t.Helper()

	prefixes := [...]string{
		"// Code generated by tools/gen-cldr",
		"// Code generated by tools/gen-plural-rules",
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.PackageClauseOnly|parser.ParseComments)
	if err != nil {
		t.Fatalf("parse comments %s: %v", path, err)
	}
	for _, group := range f.Comments {
		for _, c := range group.List {
			for _, prefix := range prefixes {
				if strings.HasPrefix(c.Text, prefix) {
					return true
				}
			}
		}
	}
	return false
}

func mustRel(t *testing.T, root, path string) string {
	t.Helper()

	rel, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatalf("relative path for %s: %v", path, err)
	}
	return rel
}
