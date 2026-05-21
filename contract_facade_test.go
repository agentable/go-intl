package gointl

import (
	"os"
	"strings"
	"testing"
)

func TestRootFacadeGuidanceDocumentsDirectImportsAndAggregateCost(t *testing.T) {
	t.Parallel()

	readme := mustReadText(t, "README.md")
	requireContains(t, readme, "Use constructor packages directly in production services that need one `Intl` constructor")
	requireContains(t, readme, "root package is measured as aggregate facade cost")
	requireContains(t, readme, "Treat `go list -deps .` as root aggregate evidence only")
	requireContains(t, readme, "go list -deps ./numberformat | wc -l")
	requireContains(t, readme, "go list -deps ./datetimeformat | wc -l")
	requireContains(t, readme, "label root facade harnesses separately from formatter-only harnesses")

	doc := mustReadText(t, "doc.go")
	requireContains(t, doc, "Package gointl provides the root ECMA-402 Intl namespace for Go.")
	requireContains(t, doc, "Import formatter subpackages directly when an application needs one")
	requireContains(t, doc, "aggregate facade cost")
	requireContains(t, doc, "formatter subpackages separately for single-constructor services")
}

func mustReadText(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func requireContains(t *testing.T, got, want string) {
	t.Helper()

	if !strings.Contains(got, want) {
		t.Fatalf("content missing %q", want)
	}
}
