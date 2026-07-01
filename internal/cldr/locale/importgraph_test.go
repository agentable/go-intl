package cldrlocale_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// Import-graph gate: dependencies inside internal/cldr may only point down, and
// the retired root internal/cldr package may never be imported again.
//
// Like the data-shape gate, this asserts an absolute, threshold-free shape. It
// reads only the production (non-test) imports of each package, parsing every
// non-test Go file directly (so build-tag-gated files cannot hide an edge);
// test files are free to reach sideways for round-trip checks without
// weakening the production boundary.
//
// Four directional rules, from the bottom up:
//
//   - codec imports nothing from this module: only the standard library.
//   - locale (the kernel) imports codec plus the stdlib-only locale identity
//     leaf internal/localeid.
//   - every cldr/<domain> leaf imports only codec, locale, the stdlib-only
//     shared ECMA-402 utility leaves, and any leaf->leaf edge named in the
//     exception table below.
//   - no package anywhere imports the dead root internal/cldr package; the path
//     is matched exactly so the boundary cannot silently reappear.
//
// formatter packages legitimately share domains such as currency, so this gate
// makes no claim about which formatter imports which domain. It governs only the
// internal/cldr dependency direction.

const modulePrefix = "github.com/agentable/go-intl/"

// deadRootPackage is the retired root internal/cldr package. No package may
// import it; the match is exact so a domain subpackage (.../internal/cldr/unit)
// is never mistaken for the root.
const deadRootPackage = modulePrefix + "internal/cldr"

// sharedUtilityLeaves are the stdlib-only, go-intl-internal helper packages a
// cldr leaf or the kernel may depend on. Each is a leaf in the import graph
// (it imports no other go-intl package), so depending on one cannot create a
// cycle or pull in another domain's data. The reason column records why the
// edge is allowed to exist; a leaf importing anything outside this set, codec,
// or locale fails the gate.
type sharedUtilityLeaf struct {
	path   string
	reason string
}

var sharedUtilityLeaves = [...]sharedUtilityLeaf{
	{
		path:   "internal/localeid",
		reason: "locale identity and likely-subtag expansion",
	},
	{
		path:   "internal/numbering",
		reason: "numbering-system metadata",
	},
	{
		path:   "internal/pattern",
		reason: "brace-delimited pattern tokenizer",
	},
	{
		path:   "internal/plural",
		reason: "plural rule engine",
	},
}

// leafToLeafEdge is one explicit, reasoned exception to "leaves do not import
// leaves". It names a single directed edge between two cldr domain packages.
type leafToLeafEdge struct {
	// from is the importing domain package (short name under internal/cldr/).
	from string
	// to is the imported domain package (short name under internal/cldr/).
	to string
	// reason states why both formatters share this CLDR owner.
	reason string
}

// leafToLeafEdges is the shrink-only exception table for cross-domain leaf
// edges. Today it holds exactly one: displaynames reads currency names from the
// currency domain, the single CLDR owner shared by NumberFormat and
// DisplayNames. The gate fails if a listed edge is absent from the actual import
// graph, so a row cannot outlive the edge it sanctions.
var leafToLeafEdges = [...]leafToLeafEdge{
	{
		from:   "displaynames",
		to:     "currency",
		reason: "DisplayNames reads currency names from the shared currency owner",
	},
}

// cldrDomains are the leaf domain packages under internal/cldr, excluding the
// codec primitives and the locale kernel, which have their own stricter rules.
// TestCLDRImportGraphDirection asserts this list matches the on-disk set, so a
// new domain cannot silently escape the gate.
var cldrDomains = [...]string{
	"currency",
	"date",
	"displaynames",
	"list",
	"number",
	"plural",
	"relativetime",
	"timezone",
	"unit",
}

func TestCLDRImportGraphDirection(t *testing.T) {
	t.Parallel()

	// Completeness: cldrDomains must equal the on-disk set of Go package
	// directories under internal/cldr (minus codec and the kernel). A domain
	// added on disk but missing here would never have its imports scanned.
	if disk := onDiskDomains(t); !slices.Equal(disk, slices.Sorted(slices.Values(cldrDomains[:]))) {
		t.Errorf("cldrDomains list %v does not match on-disk domain set %v; update the list", cldrDomains, disk)
	}

	// Build the allowed-edge set per leaf: codec, locale, every shared utility
	// leaf, and any explicit leaf->leaf exception that names this domain as the
	// importer. seenEdges records which exceptions actually fire so the table
	// can only shrink.
	seenEdges := make(map[leafToLeafEdge]bool, len(leafToLeafEdges))

	t.Run("codec", func(t *testing.T) {
		t.Parallel()
		imps := moduleImports(t, "codec")
		if len(imps) != 0 {
			t.Errorf("codec must import only the standard library; found go-intl imports %v", imps)
		}
	})

	t.Run("locale", func(t *testing.T) {
		t.Parallel()
		allowed := map[string]bool{
			modulePrefix + "internal/cldr/codec": true,
			modulePrefix + "internal/localeid":   true,
		}
		for _, imp := range moduleImports(t, "locale") {
			if !allowed[imp] {
				t.Errorf("locale kernel imports %s; allowed: codec and internal/localeid only", imp)
			}
		}
	})

	for _, domain := range cldrDomains[:] {
		t.Run(domain, func(t *testing.T) {
			t.Parallel()
			for _, imp := range moduleImports(t, domain) {
				if isAllowedLeafImport(domain, imp) {
					continue
				}
				t.Errorf("domain %s imports %s; not codec, locale, a shared utility leaf, or a sanctioned leaf->leaf edge", domain, imp)
			}
		})
	}

	// Record which leaf->leaf edges are present so missing rows are caught.
	for _, domain := range cldrDomains[:] {
		for _, imp := range moduleImports(t, domain) {
			to, ok := domainOf(imp)
			if !ok {
				continue
			}
			edge := leafToLeafEdge{from: domain, to: to}
			for _, e := range leafToLeafEdges[:] {
				if e.from == edge.from && e.to == edge.to {
					seenEdges[leafToLeafEdge{from: e.from, to: e.to, reason: e.reason}] = true
				}
			}
		}
	}

	// Shrink-only enforcement: every declared edge must exist in the graph.
	for _, e := range leafToLeafEdges[:] {
		if !seenEdges[e] {
			t.Errorf("leaf->leaf exception %s -> %s is no longer in the import graph; remove its row (%s)", e.from, e.to, e.reason)
		}
	}
}

// TestNoImportOfRetiredRootCLDR scans every package in the main module and fails
// if any production or test import targets the retired root internal/cldr
// package. The exact-path check guards against the boundary reappearing.
func TestNoImportOfRetiredRootCLDR(t *testing.T) {
	t.Parallel()

	root := filepath.Clean("../../..")
	violators := importersOfDeadRoot(t, root)
	for _, v := range violators {
		t.Errorf("%s imports the retired root package %s; import a domain under internal/cldr/ instead", v, deadRootPackage)
	}
}

// TestRetiredRootCLDRHasNoGoFiles makes the data-boundary shape explicit: the
// retired internal/cldr root may hold metadata such as VERSION, but it must not
// become a Go package again.
func TestRetiredRootCLDRHasNoGoFiles(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir("..")
	if err != nil {
		t.Fatalf("read internal/cldr: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		t.Errorf("internal/cldr/%s reintroduces the retired root Go package; put CLDR code in a leaf domain", entry.Name())
	}
}

// isAllowedLeafImport reports whether a leaf domain may import the given go-intl
// package: codec, locale, a shared utility leaf, or a sanctioned leaf->leaf
// edge naming this domain as importer.
func isAllowedLeafImport(domain, imp string) bool {
	switch imp {
	case modulePrefix + "internal/cldr/codec",
		modulePrefix + "internal/cldr/locale":
		return true
	}
	if rel, ok := strings.CutPrefix(imp, modulePrefix); ok {
		if isSharedUtilityLeaf(rel) {
			return true
		}
	}
	if to, ok := domainOf(imp); ok {
		for _, e := range leafToLeafEdges[:] {
			if e.from == domain && e.to == to {
				return true
			}
		}
	}
	return false
}

func isSharedUtilityLeaf(rel string) bool {
	for _, leaf := range sharedUtilityLeaves[:] {
		if leaf.path == rel {
			return true
		}
	}
	return false
}

// domainOf returns the short domain name for an internal/cldr/<domain> import
// path, or false if imp is not such a path.
func domainOf(imp string) (string, bool) {
	const prefix = modulePrefix + "internal/cldr/"
	rel, ok := strings.CutPrefix(imp, prefix)
	if !ok || rel == "" || strings.Contains(rel, "/") {
		return "", false
	}
	return rel, true
}

// importersOfDeadRoot walks the main module from root and returns the
// repo-relative path of every Go file (production or test) that imports the
// retired root internal/cldr package by its exact path. The walk skips
// dot-directories and the tools/ generator module trees, which are separate
// modules with their own import graphs.
func importersOfDeadRoot(t *testing.T, root string) []string {
	t.Helper()

	var violators []string
	fset := token.NewFileSet()
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
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			t.Fatalf("parse imports %s: %v", path, perr)
		}
		for _, spec := range f.Imports {
			p, uerr := strconv.Unquote(spec.Path.Value)
			if uerr != nil {
				continue
			}
			if p == deadRootPackage {
				rel, rerr := filepath.Rel(root, path)
				if rerr != nil {
					rel = path
				}
				violators = append(violators, filepath.ToSlash(rel))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk module: %v", err)
	}
	slices.Sort(violators)
	return violators
}

// moduleImports returns the go-intl production imports of the named cldr leaf,
// parsing every non-test Go file in the package directory regardless of build
// constraints (go/build would drop tag-gated files and hide their edges).
// Standard library imports are dropped; only same-module edges are
// graph-relevant.
func moduleImports(t *testing.T, domain string) []string {
	t.Helper()

	dir := filepath.Join("..", domain)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read internal/cldr/%s: %v", domain, err)
	}
	fset := token.NewFileSet()
	seen := make(map[string]bool)
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, perr := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ImportsOnly)
		if perr != nil {
			t.Fatalf("parse imports internal/cldr/%s/%s: %v", domain, name, perr)
		}
		for _, spec := range f.Imports {
			p, uerr := strconv.Unquote(spec.Path.Value)
			if uerr != nil {
				continue
			}
			if strings.HasPrefix(p, modulePrefix) {
				seen[p] = true
			}
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

// onDiskDomains lists the Go package directories under internal/cldr, sorted,
// excluding the codec primitives and the locale kernel.
func onDiskDomains(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir("..")
	if err != nil {
		t.Fatalf("read internal/cldr: %v", err)
	}
	var out []string
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || name == "codec" || name == "locale" {
			continue
		}
		files, derr := os.ReadDir(filepath.Join("..", name))
		if derr != nil {
			t.Fatalf("read internal/cldr/%s: %v", name, derr)
		}
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".go") {
				out = append(out, name)
				break
			}
		}
	}
	slices.Sort(out)
	return out
}
