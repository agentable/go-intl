package gointl

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"golang.org/x/text/language"

	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
)

func TestMessageformatIntegrationContract_NoReverseDependency(t *testing.T) {
	t.Parallel()

	out, err := exec.Command("go", "list", "-deps", "./...").CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps ./... failed: %v\n%s", err, out)
	}
	if violations := messageformatDependencyViolations(string(out)); len(violations) > 0 {
		t.Fatalf("go-intl must not depend on messageformat-go: %v", violations)
	}
}

func TestMessageformatIntegrationContract_NoRichTextHooks(t *testing.T) {
	t.Parallel()

	violations, err := forbiddenRichTextHookViolations(".")
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) > 0 {
		t.Fatalf("go-intl must not expose messageformat rich-text hooks: %v", violations)
	}
}

func TestMessageformatIntegrationContract_OperandsRecordIsInternal(t *testing.T) {
	t.Parallel()

	violations, err := publicOperandsRecordViolations(".")
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) > 0 {
		t.Fatalf("OperandsRecord must remain internal: %v", violations)
	}
}

func TestMessageformatIntegrationContract_PublicConsumerPattern(t *testing.T) {
	t.Parallel()

	loc, err := locale.New(language.MustParse("en-US"))
	if err != nil {
		t.Fatal(err)
	}
	nf, err := numberformat.New(loc)
	if err != nil {
		t.Fatal(err)
	}
	if got := nf.FormatFloat64(1234.5); got == "" {
		t.Fatal("numberformat produced empty output")
	}
	dtf, err := datetimeformat.New(loc, datetimeformat.Options{
		Year:  datetimeformat.NumericFieldStyle,
		Month: datetimeformat.ShortMonthStyle,
		Day:   datetimeformat.NumericFieldStyle,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := dtf.Format(time.Date(2026, time.May, 8, 12, 0, 0, 0, time.UTC)); got == "" {
		t.Fatal("datetimeformat produced empty output")
	}
	pr, err := pluralrules.New(loc)
	if err != nil {
		t.Fatal(err)
	}
	if got := pr.SelectInt(2); got != pluralrules.Other {
		t.Fatalf("pluralrules.SelectInt(2) = %s, want %s", got, pluralrules.Other)
	}
}

func TestMessageformatIntegrationContract_ReportTemplate(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("reports/messageformat-go.md")
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, field := range []string{"dependency", "go-intl version", "problem", "trigger", "expected", "actual", "workaround", "upstream issue"} {
		if !strings.Contains(content, field) {
			t.Fatalf("reports/messageformat-go.md missing %q", field)
		}
	}
	for _, required := range []string{"Dependency Issue Reporting", "SPEC 61", "must not be implemented as go-intl reimplementations"} {
		if !strings.Contains(content, required) {
			t.Fatalf("reports/messageformat-go.md missing %q", required)
		}
	}
}

func messageformatDependencyViolations(deps string) []string {
	var violations []string
	for dep := range strings.SplitSeq(deps, "\n") {
		if strings.Contains(dep, "messageformat-go") {
			violations = append(violations, dep)
		}
	}
	return violations
}

func forbiddenRichTextHookViolations(root string) ([]string, error) {
	forbidden := []string{
		"default" + "Rich" + "Text" + "Elements",
		"Rich" + "Text" + "Elements",
		"XML" + "Element" + "Fn",
	}
	return sourceSymbolViolations(root, forbidden, nil)
}

func publicOperandsRecordViolations(root string) ([]string, error) {
	var violations []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if shouldSkipIntegrationScanDir(path) || isInternalPath(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			if exportedDeclContainsOperandsRecord(decl) {
				violations = append(violations, path)
			}
		}
		return nil
	})
	return violations, err
}

func exportedDeclContainsOperandsRecord(decl ast.Decl) bool {
	switch decl := decl.(type) {
	case *ast.FuncDecl:
		return decl.Name.IsExported() && nodeContainsOperandsRecord(decl.Type)
	case *ast.GenDecl:
		if decl.Tok != token.TYPE && decl.Tok != token.VAR && decl.Tok != token.CONST {
			return false
		}
		for _, spec := range decl.Specs {
			switch spec := spec.(type) {
			case *ast.TypeSpec:
				if spec.Name.IsExported() && exportedTypeContainsOperandsRecord(spec.Type) {
					return true
				}
			case *ast.ValueSpec:
				if exportedValueSpecContainsOperandsRecord(spec) {
					return true
				}
			}
		}
	}
	return false
}

func exportedTypeContainsOperandsRecord(expr ast.Expr) bool {
	switch expr := expr.(type) {
	case *ast.StructType:
		return slices.ContainsFunc(expr.Fields.List, exportedFieldContainsOperandsRecord)
	case *ast.InterfaceType:
		return slices.ContainsFunc(expr.Methods.List, exportedFieldContainsOperandsRecord)
	default:
		return nodeContainsOperandsRecord(expr)
	}
}

func exportedFieldContainsOperandsRecord(field *ast.Field) bool {
	if len(field.Names) == 0 {
		return nodeContainsOperandsRecord(field.Type)
	}
	for _, name := range field.Names {
		if name.IsExported() {
			return nodeContainsOperandsRecord(field.Type)
		}
	}
	return false
}

func exportedValueSpecContainsOperandsRecord(spec *ast.ValueSpec) bool {
	exported := false
	for _, name := range spec.Names {
		if name.IsExported() {
			exported = true
			break
		}
	}
	return exported && nodeContainsOperandsRecord(spec.Type)
}

func nodeContainsOperandsRecord(node ast.Node) bool {
	if node == nil {
		return false
	}
	found := false
	ast.Inspect(node, func(node ast.Node) bool {
		if ident, ok := node.(*ast.Ident); ok && ident.Name == "Operands"+"Record" {
			found = true
			return false
		}
		return true
	})
	return found
}

func sourceSymbolViolations(root string, symbols []string, skipFile func(string) bool) ([]string, error) {
	var violations []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if shouldSkipIntegrationScanDir(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || (skipFile != nil && skipFile(path)) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		for _, symbol := range symbols {
			if strings.Contains(content, symbol) {
				violations = append(violations, fmt.Sprintf("%s: %s", path, symbol))
			}
		}
		return nil
	})
	return violations, err
}

func shouldSkipIntegrationScanDir(path string) bool {
	name := filepath.Base(path)
	return name == ".git" || name == ".references" || name == ".plans" || name == ".research" || name == ".tmp" || name == "SPECS" || name == "reports"
}

func isInternalPath(path string) bool {
	for part := range strings.SplitSeq(filepath.Clean(path), string(filepath.Separator)) {
		if part == "internal" {
			return true
		}
	}
	return false
}
