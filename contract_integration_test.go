package gointl

import (
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/durationformat"
	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
	"github.com/agentable/go-intl/relativetimeformat"
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

func TestMessageformatIntegrationContract_PublicConsumerPattern(t *testing.T) {
	t.Parallel()

	locales, err := locale.ParseList("en-US")
	if err != nil {
		t.Fatal(err)
	}
	nf, err := numberformat.New(locales, numberformat.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := nf.Format(numberformat.Float(1234.5)); got == "" {
		t.Fatal("numberformat produced empty output")
	}
	dtf, err := datetimeformat.New(locales, datetimeformat.Options{
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
	pr, err := pluralrules.New(locales, pluralrules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := pr.Select(pluralrules.Int(2))
	if err != nil {
		t.Fatal(err)
	}
	if got != pluralrules.Other {
		t.Fatalf("pluralrules.SelectInt(2) = %s, want %s", got, pluralrules.Other)
	}
	lf, err := listformat.New(locales, listformat.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := lf.Format([]string{"A", "B"}); got == "" {
		t.Fatal("listformat produced empty output")
	}
	rtf, err := relativetimeformat.New(locales, relativetimeformat.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := rtf.FormatInt(-1, relativetimeformat.Second); err != nil || got == "" {
		t.Fatalf("relativetimeformat.FormatInt(-1, second) = %q, %v; want non-empty output", got, err)
	}
	df, err := durationformat.New(locales, durationformat.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := df.Format(durationformat.Duration{Hours: 1}); err != nil || got == "" {
		t.Fatalf("durationformat.Format({Hours: 1}) = %q, %v; want non-empty output", got, err)
	}
}

func TestRootSupportedValuesContractUsesOwnedData(t *testing.T) {
	t.Parallel()

	imports, err := packageImportPaths(".")
	if err != nil {
		t.Fatal(err)
	}
	if imports["github.com/agentable/go-intl/internal/cldr/supported"] {
		t.Fatal("root supported values must not route through generated forwarding wrappers")
	}
	if !imports["github.com/agentable/go-intl/internal/cldr"] {
		t.Fatal("root supported values must read CLDR-backed values from internal/cldr")
	}
	if !imports["github.com/agentable/go-intl/internal/ecma402"] {
		t.Fatal("root supported values must read sanctioned unit identifiers from internal/ecma402")
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

func fileImportPaths(path string) (map[string]bool, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	imports := map[string]bool{}
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, err
		}
		imports[path] = true
	}
	return imports, nil
}

func packageImportPaths(root string) (map[string]bool, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	imports := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".go" || strings.HasSuffix(name, "_test.go") {
			continue
		}
		fileImports, err := fileImportPaths(filepath.Join(root, name))
		if err != nil {
			return nil, err
		}
		for path := range fileImports {
			imports[path] = true
		}
	}
	return imports, nil
}
