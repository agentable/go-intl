package gointl

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestPublicSurfaceMatchesLedger(t *testing.T) {
	t.Parallel()

	for dir, want := range publicSurfaceLedger() {
		t.Run(dir, func(t *testing.T) {
			t.Parallel()

			want = slices.Clone(want)
			slices.Sort(want)
			got := exportedTopLevelNames(t, packageGoFiles(t, dir))
			if !slices.Equal(got, want) {
				t.Fatalf("%s exported surface = %v, want %v", dir, got, want)
			}
		})
	}
}

func TestPublicSurfaceLedgerEntriesHaveOwnerOrBridgeRationale(t *testing.T) {
	t.Parallel()

	for dir, names := range publicSurfaceLedger() {
		for _, name := range names {
			owner, rationale := publicSurfaceOwnerOrBridge(dir, name)
			if owner == "" && rationale == "" {
				t.Fatalf("%s.%s has no ECMA-402 owner or Go bridge rationale", dir, name)
			}
			if owner == "" && !strings.Contains(rationale, "Go bridge") {
				t.Fatalf("%s.%s rationale %q must identify a Go bridge", dir, name, rationale)
			}
		}
	}
}

func TestPublicMethodSurfaceMatchesLedger(t *testing.T) {
	t.Parallel()

	for dir, want := range publicMethodLedger() {
		t.Run(dir, func(t *testing.T) {
			t.Parallel()

			want = slices.Clone(want)
			slices.Sort(want)
			got := exportedPublicMethodNames(t, packageGoFiles(t, dir))
			if !slices.Equal(got, want) {
				t.Fatalf("%s exported method surface = %v, want %v", dir, got, want)
			}
		})
	}
}

func TestPublicMethodLedgerEntriesHaveOwnerOrBridgeRationale(t *testing.T) {
	t.Parallel()

	for dir, names := range publicMethodLedger() {
		for _, name := range names {
			owner, rationale := publicMethodOwnerOrBridge(dir, name)
			if owner == "" && rationale == "" {
				t.Fatalf("%s.%s has no ECMA-402 owner or Go bridge rationale", dir, name)
			}
			if owner == "" && !strings.Contains(rationale, "Go bridge") {
				t.Fatalf("%s.%s rationale %q must identify a Go bridge", dir, name, rationale)
			}
		}
	}
}

func TestREADMEUsageCoversActiveConstructorSurfaces(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	usage := readmeSection(string(data), "## Usage")
	if usage == "" {
		t.Fatal("README.md missing ## Usage section")
	}
	for _, dir := range activeConstructorPackageDirs() {
		if !strings.Contains(usage, dir+".New(") {
			t.Fatalf("README Usage does not show natural %s.New usage", dir)
		}
	}
}

func TestUsageDocsRejectRetiredConvenienceAPIs(t *testing.T) {
	t.Parallel()

	files := []string{"README.md"}
	for _, dir := range activeConstructorPackageDirs() {
		files = append(files, filepath.Join(dir, "doc.go"))
	}
	files = append(files, filepath.Join("locale", "doc.go"))

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			text := string(data)
			for _, retired := range retiredUsageDocNames() {
				if strings.Contains(text, retired) {
					t.Fatalf("%s mentions retired convenience API %q; usage docs must show the final API shape", file, retired)
				}
			}
		})
	}
}

func TestPublicSourceTextRejectsRetiredConvenienceAPIs(t *testing.T) {
	t.Parallel()

	for _, file := range publicPackageSourceFiles(t) {
		t.Run(file, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			for _, line := range strings.Split(string(data), "\n") {
				if lineMentionsAllowedRetiredText(line) {
					continue
				}
				for _, retired := range retiredSourceTextNames() {
					if strings.Contains(line, retired) {
						t.Fatalf("%s mentions retired public API text %q in %q", file, retired, strings.TrimSpace(line))
					}
				}
			}
		})
	}
}

func TestPublicSurfaceRejectsForbiddenRootAPIs(t *testing.T) {
	t.Parallel()

	forbidden := map[string]bool{
		"Intl":                true,
		"New":                 true,
		"Option":              true,
		"WithTimeZone":        true,
		"FormatNumberInt":     true,
		"FormatNumberInt64":   true,
		"FormatNumberUint":    true,
		"FormatNumberUint64":  true,
		"FormatNumberFloat64": true,
		"FormatNumberDecimal": true,
		"FormatDate":          true,
		"FormatTime":          true,
		"FormatRange":         true,
		"SelectPluralInt":     true,
		"SelectPluralInt64":   true,
		"SelectPluralUint":    true,
		"SelectPluralUint64":  true,
		"SelectPluralFloat64": true,
		"SelectPluralDecimal": true,
		"FormatList":          true,
		"FormatRelativeTime":  true,
		"FormatDuration":      true,
		"WithCache":           true,
		"WithoutCache":        true,
		"ResetGlobalCache":    true,
		"Cache":               true,
		"Version":             true,
	}

	for _, file := range rootGoFiles(t) {
		checkFileExports(t, file, func(name string) {
			if forbidden[name] {
				t.Fatalf("%s exports forbidden root API %q", file, name)
			}
		})
	}
}

func TestPublicSurfaceRejectsAppendAndDeprecatedShims(t *testing.T) {
	t.Parallel()

	for _, file := range publicPackageGoFiles(t) {
		fset := token.NewFileSet()
		parsed, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range parsed.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				if !decl.Name.IsExported() {
					continue
				}
				if strings.HasPrefix(decl.Name.Name, "Append") {
					t.Fatalf("%s exports %q; public Append APIs are outside ECMA-402", file, decl.Name.Name)
				}
				if hasDeprecatedComment(decl.Doc) {
					t.Fatalf("%s exports deprecated shim %q", file, decl.Name.Name)
				}
			case *ast.GenDecl:
				if !hasDeprecatedComment(decl.Doc) {
					continue
				}
				for _, spec := range decl.Specs {
					for _, name := range exportedSpecNames(spec) {
						t.Fatalf("%s exports deprecated shim %q", file, name)
					}
				}
			}
		}
	}
}

func TestPublicSurfaceRejectsRetiredConvenienceNames(t *testing.T) {
	t.Parallel()

	for dir, names := range map[string][]string{
		"locale": {
			"MustParse",
			"MustParseList",
		},
		"numberformat": {
			"BigFloat",
			"CurrencyCode",
			"UnitIdentifier",
		},
		"pluralrules": {
			"BigFloat",
		},
		"relativetimeformat": {
			"Days",
			"FormatDecimal",
			"FormatDecimalToParts",
			"FormatFloat64",
			"FormatFloat64ToParts",
			"FormatInt",
			"FormatInt64",
			"FormatInt64ToParts",
			"FormatIntToParts",
			"FormatUint",
			"FormatUint64",
			"FormatUint64ToParts",
			"FormatUintToParts",
			"Hours",
			"Minutes",
			"Months",
			"Quarters",
			"Seconds",
			"Weeks",
			"Years",
		},
	} {
		for _, name := range names {
			if slices.Contains(publicSurfaceLedger()[dir], name) || slices.Contains(exportedMethodNames(t, packageGoFiles(t, dir)), name) {
				t.Fatalf("%s.%s is retired convenience surface; keep the canonical typed value instead", dir, name)
			}
		}
	}
}

func TestPublicSurfaceRejectsInternalTypeAliases(t *testing.T) {
	t.Parallel()

	for _, file := range publicPackageGoFiles(t) {
		if filepath.Dir(file) == "." {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		imports := importNames(parsed)
		for _, decl := range parsed.Decls {
			decl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range decl.Specs {
				spec, ok := spec.(*ast.TypeSpec)
				if !ok || spec.Assign == 0 || !spec.Name.IsExported() {
					continue
				}
				if selector, ok := spec.Type.(*ast.SelectorExpr); ok && selectorUsesInternalImport(selector, imports) {
					t.Fatalf("%s exports internal type alias %q", file, spec.Name.Name)
				}
			}
		}
	}
}

func TestPublicConstructorsUseSharedLocaleResolution(t *testing.T) {
	t.Parallel()

	for _, file := range publicPackageGoFiles(t) {
		if filepath.Dir(file) == "." {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		imports := importNames(parsed)
		ast.Inspect(parsed, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || !selectorUsesImport(selector, imports, "github.com/agentable/go-intl/internal/localematcher") {
				return true
			}
			switch selector.Sel.Name {
			case "ResolveLocale", "Match", "MatchWithMaximizer", "LookupMatcher", "BestFitMatcher", "BestFitMatcherWithMaximizer":
				t.Fatalf("%s calls localematcher.%s directly; public constructors must use internal/ecma402", file, selector.Sel.Name)
			}
			return true
		})
	}
}

func TestResolveLocaleHasOneProductionEntryPoint(t *testing.T) {
	t.Parallel()

	const allowed = "internal/ecma402/constructor_locale.go"
	for _, file := range productionGoFiles(t) {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		imports := importNames(parsed)
		ast.Inspect(parsed, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "ResolveLocale" {
				return true
			}
			if !selectorUsesImport(selector, imports, "github.com/agentable/go-intl/internal/localematcher") {
				return true
			}
			if file != allowed {
				t.Fatalf("%s calls localematcher.ResolveLocale directly; constructor negotiation must enter through %s", file, allowed)
			}
			return true
		})
	}
}

func TestNoRetiredConstructorLocaleFallbackPackage(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("internal/cldrmatch/*.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) > 0 {
		t.Fatalf("internal/cldrmatch is retired; formatter constructors must use internal/ecma402.ResolveConstructorLocale and own their CLDR data fallback, found %v", files)
	}
}

func rootGoFiles(t *testing.T) []string {
	t.Helper()

	return packageGoFiles(t, ".")
}

func publicPackageGoFiles(t *testing.T) []string {
	t.Helper()

	var files []string
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipPublicSurfaceDir(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func productionGoFiles(t *testing.T) []string {
	t.Helper()

	var files []string
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipProductionDir(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(files)
	return files
}

func publicPackageSourceFiles(t *testing.T) []string {
	t.Helper()

	var files []string
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipPublicSurfaceDir(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".go" && path != "public_surface_test.go" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func skipPublicSurfaceDir(path string) bool {
	switch path {
	case ".", "collator", "datetimeformat", "displaynames", "durationformat", "listformat", "locale", "numberformat", "pluralrules", "relativetimeformat", "segmenter":
		return false
	default:
		return true
	}
}

func skipProductionDir(path string) bool {
	name := filepath.Base(path)
	if strings.HasPrefix(name, ".") {
		return path != "."
	}
	switch path {
	case "bin", "reports", "tools":
		return true
	default:
		return false
	}
}

func nonTestFiles(files []string) []string {
	out := files[:0]
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		out = append(out, file)
	}
	return out
}

func packageGoFiles(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	return nonTestFiles(entries)
}

func exportedTopLevelNames(t *testing.T, files []string) []string {
	t.Helper()

	seen := map[string]bool{}
	for _, file := range files {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range parsed.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				if decl.Recv == nil && decl.Name.IsExported() {
					seen[decl.Name.Name] = true
				}
			case *ast.GenDecl:
				for _, spec := range decl.Specs {
					for _, name := range exportedSpecNames(spec) {
						seen[name] = true
					}
				}
			}
		}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func exportedMethodNames(t *testing.T, files []string) []string {
	t.Helper()

	seen := map[string]bool{}
	for _, file := range files {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range parsed.Decls {
			decl, ok := decl.(*ast.FuncDecl)
			if ok && decl.Recv != nil && decl.Name.IsExported() {
				seen[decl.Name.Name] = true
			}
		}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func exportedPublicMethodNames(t *testing.T, files []string) []string {
	t.Helper()

	seen := map[string]bool{}
	for _, file := range files {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range parsed.Decls {
			decl, ok := decl.(*ast.FuncDecl)
			if !ok || decl.Recv == nil || !decl.Name.IsExported() || !exportedReceiverType(decl.Recv) {
				continue
			}
			seen[decl.Name.Name] = true
		}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func exportedReceiverType(recv *ast.FieldList) bool {
	if recv == nil || len(recv.List) == 0 {
		return false
	}
	switch expr := recv.List[0].Type.(type) {
	case *ast.Ident:
		return expr.IsExported()
	case *ast.StarExpr:
		ident, ok := expr.X.(*ast.Ident)
		return ok && ident.IsExported()
	default:
		return false
	}
}

func checkFileExports(t *testing.T, file string, check func(string)) {
	t.Helper()

	parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, decl := range parsed.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Name.IsExported() {
				check(decl.Name.Name)
			}
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				for _, name := range exportedSpecNames(spec) {
					check(name)
				}
			}
		}
	}
}

func exportedSpecNames(spec ast.Spec) []string {
	switch spec := spec.(type) {
	case *ast.TypeSpec:
		if spec.Name.IsExported() {
			return []string{spec.Name.Name}
		}
	case *ast.ValueSpec:
		var names []string
		for _, name := range spec.Names {
			if name.IsExported() {
				names = append(names, name.Name)
			}
		}
		return names
	}
	return nil
}

func hasDeprecatedComment(group *ast.CommentGroup) bool {
	return group != nil && strings.Contains(group.Text(), "Deprecated:")
}

func importNames(file *ast.File) map[string]string {
	imports := map[string]string{}
	for _, spec := range file.Imports {
		path := strings.Trim(spec.Path.Value, `"`)
		name := filepath.Base(path)
		if spec.Name != nil {
			name = spec.Name.Name
		}
		imports[name] = path
	}
	return imports
}

func selectorUsesInternalImport(selector *ast.SelectorExpr, imports map[string]string) bool {
	pkg, ok := selector.X.(*ast.Ident)
	return ok && strings.Contains(imports[pkg.Name], "/internal/")
}

func selectorUsesImport(selector *ast.SelectorExpr, imports map[string]string, path string) bool {
	pkg, ok := selector.X.(*ast.Ident)
	return ok && imports[pkg.Name] == path
}

func readmeSection(text, heading string) string {
	start := strings.Index(text, heading)
	if start < 0 {
		return ""
	}
	rest := text[start+len(heading):]
	end := strings.Index(rest, "\n## ")
	if end < 0 {
		return rest
	}
	return rest[:end]
}

func activeConstructorPackageDirs() []string {
	return []string{
		"collator",
		"datetimeformat",
		"displaynames",
		"durationformat",
		"listformat",
		"numberformat",
		"pluralrules",
		"relativetimeformat",
		"segmenter",
	}
}

func retiredUsageDocNames() []string {
	return []string{
		"locale.MustParse",
		"locale.MustParseList",
		"numberformat.CurrencyCode",
		"numberformat.UnitIdentifier",
		"numberformat.BigFloat",
		"pluralrules.BigFloat",
		".FormatInt(",
		".FormatInt64(",
		".FormatIntToParts(",
		".FormatInt64ToParts(",
		".FormatUint(",
		".FormatUint64(",
		".FormatUintToParts(",
		".FormatUint64ToParts(",
		".FormatFloat64(",
		".FormatFloat64ToParts(",
		".FormatDecimal(",
		".FormatDecimalToParts(",
	}
}

func retiredSourceTextNames() []string {
	return []string{
		"locale.MustParse",
		"locale.MustParseList",
		"numberformat.CurrencyCode",
		"numberformat.UnitIdentifier",
		"numberformat.BigFloat",
		"pluralrules.BigFloat",
		"FormatInt(",
		"FormatInt64(",
		"FormatIntToParts(",
		"FormatInt64ToParts(",
		"FormatUint(",
		"FormatUint64(",
		"FormatUintToParts(",
		"FormatUint64ToParts(",
		"FormatFloat64(",
		"FormatFloat64ToParts(",
		"FormatDecimal(",
		"FormatDecimalToParts(",
	}
}

func lineMentionsAllowedRetiredText(line string) bool {
	return strings.Contains(line, "strconv.FormatInt(") ||
		strings.Contains(line, "strconv.FormatUint(")
}

func publicSurfaceOwnerOrBridge(dir, name string) (owner, rationale string) {
	if rationale := publicSurfaceBridgeRationale(dir, name); rationale != "" {
		return "", rationale
	}
	if owner := publicSurfaceOwner(dir); owner != "" {
		return owner, ""
	}
	return "", ""
}

func publicMethodOwnerOrBridge(dir, name string) (owner, rationale string) {
	if rationale := publicMethodBridgeRationale(dir, name); rationale != "" {
		return "", rationale
	}
	if owner := publicMethodOwner(dir, name); owner != "" {
		return owner, ""
	}
	return "", ""
}

func publicSurfaceOwner(dir string) string {
	switch dir {
	case ".":
		return "Intl"
	case "collator":
		return "Intl.Collator"
	case "datetimeformat":
		return "Intl.DateTimeFormat"
	case "displaynames":
		return "Intl.DisplayNames"
	case "durationformat":
		return "Intl.DurationFormat"
	case "listformat":
		return "Intl.ListFormat"
	case "locale":
		return "Intl.Locale"
	case "numberformat":
		return "Intl.NumberFormat"
	case "pluralrules":
		return "Intl.PluralRules"
	case "relativetimeformat":
		return "Intl.RelativeTimeFormat"
	case "segmenter":
		return "Intl.Segmenter"
	default:
		return ""
	}
}

func publicMethodOwner(dir, name string) string {
	switch dir {
	case "collator":
		if name == "Compare" || name == "ResolvedOptions" {
			return "Intl.Collator"
		}
	case "datetimeformat":
		switch name {
		case "Format", "FormatToParts", "FormatRange", "FormatRangeToParts", "ResolvedOptions":
			return "Intl.DateTimeFormat"
		}
	case "displaynames":
		if name == "Of" || name == "ResolvedOptions" {
			return "Intl.DisplayNames"
		}
	case "durationformat":
		if name == "Format" || name == "FormatToParts" || name == "ResolvedOptions" {
			return "Intl.DurationFormat"
		}
	case "listformat":
		if name == "Format" || name == "FormatToParts" || name == "ResolvedOptions" {
			return "Intl.ListFormat"
		}
	case "locale":
		switch name {
		case "BaseName", "Calendar", "CaseFirst", "Collation", "FirstDayOfWeek", "GetCalendars", "GetCollations", "GetHourCycles", "GetNumberingSystems", "GetTextInfo", "GetTimeZones", "GetWeekInfo", "HourCycle", "Language", "Maximize", "Minimize", "NumberingSystem", "Numeric", "Region", "Script", "String", "Variants":
			return "Intl.Locale"
		}
	case "numberformat":
		switch name {
		case "Format", "FormatToParts", "FormatRange", "FormatRangeToParts", "ResolvedOptions":
			return "Intl.NumberFormat"
		}
	case "pluralrules":
		if name == "Select" || name == "SelectRange" || name == "ResolvedOptions" {
			return "Intl.PluralRules"
		}
	case "relativetimeformat":
		if name == "Format" || name == "FormatToParts" || name == "ResolvedOptions" {
			return "Intl.RelativeTimeFormat"
		}
	case "segmenter":
		if name == "Segment" || name == "ResolvedOptions" {
			return "Intl.Segmenter"
		}
	}
	return ""
}

func publicSurfaceBridgeRationale(dir, name string) string {
	if rationale, ok := publicSurfaceBridgeRationales()[publicSurfaceKey(dir, name)]; ok {
		return rationale
	}
	return ""
}

func publicMethodBridgeRationale(dir, name string) string {
	if rationale, ok := publicMethodBridgeRationales()[publicSurfaceKey(dir, name)]; ok {
		return rationale
	}
	return ""
}

func publicSurfaceKey(dir, name string) string {
	if dir == "." {
		return name
	}
	return dir + "." + name
}

func publicMethodBridgeRationales() map[string]string {
	return map[string]string{
		"locale.Equal":         "Go bridge for comparable Intl.Locale values",
		"locale.MarshalText":   "Go bridge for encoding.TextMarshaler on Intl.Locale",
		"locale.Strings":       "Go bridge for exposing locale list strings without leaking slice storage",
		"locale.Tag":           "Go bridge from Intl.Locale to golang.org/x/text/language.Tag",
		"locale.UnmarshalText": "Go bridge for encoding.TextUnmarshaler on Intl.Locale",
		"locale.MarshalJSON":   "Go bridge for preserving ECMA-402 weekInfo JSON shape",

		"pluralrules.MarshalText": "Go bridge for encoding.TextMarshaler on plural categories",
		"pluralrules.String":      "Go bridge for branch-friendly ECMA-402 plural category strings",

		"segmenter.All":            "Go bridge exposing the ECMA-402 Segments iterator as iter.Seq",
		"segmenter.Containing":     "Go bridge for Intl.Segments.prototype.containing over UTF-8 rune indexes",
		"segmenter.ContainingByte": "Go bridge for Intl.Segments.prototype.containing over Go byte indexes",
	}
}

func publicSurfaceBridgeRationales() map[string]string {
	return map[string]string{
		"Bool":   "Go bridge for optional JS boolean option properties",
		"Error":  "Go bridge for ECMA-402 TypeError/RangeError-style caller-fixable failures",
		"Int":    "Go bridge for optional JS integer option properties",
		"String": "Go bridge for optional JS string option properties",

		"locale.FromTag":   "Go bridge from golang.org/x/text/language.Tag into Intl.Locale",
		"locale.List":      "Go bridge for JavaScript locale list inputs",
		"locale.Parse":     "Go bridge for fallible construction of Intl.Locale values",
		"locale.ParseList": "Go bridge for fallible construction of JavaScript locale list inputs",

		"numberformat.BigInt":  "Go bridge for NumberFormat.prototype.format numeric inputs",
		"numberformat.Decimal": "Go bridge for NumberFormat.prototype.format decimal inputs",
		"numberformat.Float":   "Go bridge for NumberFormat.prototype.format numeric inputs",
		"numberformat.Int":     "Go bridge for NumberFormat.prototype.format numeric inputs",
		"numberformat.Uint":    "Go bridge for NumberFormat.prototype.format numeric inputs",
		"numberformat.Value":   "Go bridge for NumberFormat.prototype.format numeric inputs",

		"pluralrules.BigInt":  "Go bridge for PluralRules.prototype.select numeric inputs",
		"pluralrules.Decimal": "Go bridge for PluralRules.prototype.select decimal inputs",
		"pluralrules.Float":   "Go bridge for PluralRules.prototype.select numeric inputs",
		"pluralrules.Int":     "Go bridge for PluralRules.prototype.select numeric inputs",
		"pluralrules.Uint":    "Go bridge for PluralRules.prototype.select numeric inputs",
		"pluralrules.Value":   "Go bridge for PluralRules.prototype.select numeric inputs",

		"relativetimeformat.Decimal": "Go bridge for RelativeTimeFormat.prototype.format decimal inputs",
		"relativetimeformat.Float":   "Go bridge for RelativeTimeFormat.prototype.format numeric inputs",
		"relativetimeformat.Int":     "Go bridge for RelativeTimeFormat.prototype.format numeric inputs",
		"relativetimeformat.Uint":    "Go bridge for RelativeTimeFormat.prototype.format numeric inputs",
		"relativetimeformat.Value":   "Go bridge for RelativeTimeFormat.prototype.format numeric inputs",
	}
}

func publicMethodLedger() map[string][]string {
	return map[string][]string{
		"collator": {
			"Compare",
			"ResolvedOptions",
		},
		"datetimeformat": {
			"Format",
			"FormatRange",
			"FormatRangeToParts",
			"FormatToParts",
			"ResolvedOptions",
		},
		"displaynames": {
			"Of",
			"ResolvedOptions",
		},
		"durationformat": {
			"Format",
			"FormatToParts",
			"ResolvedOptions",
		},
		"listformat": {
			"Format",
			"FormatToParts",
			"ResolvedOptions",
		},
		"locale": {
			"BaseName",
			"Calendar",
			"CaseFirst",
			"Collation",
			"Equal",
			"FirstDayOfWeek",
			"GetCalendars",
			"GetCollations",
			"GetHourCycles",
			"GetNumberingSystems",
			"GetTextInfo",
			"GetTimeZones",
			"GetWeekInfo",
			"HourCycle",
			"Language",
			"MarshalJSON",
			"MarshalText",
			"Maximize",
			"Minimize",
			"NumberingSystem",
			"Numeric",
			"Region",
			"Script",
			"String",
			"Strings",
			"Tag",
			"UnmarshalText",
			"Variants",
		},
		"numberformat": {
			"Format",
			"FormatRange",
			"FormatRangeToParts",
			"FormatToParts",
			"ResolvedOptions",
		},
		"pluralrules": {
			"MarshalText",
			"ResolvedOptions",
			"Select",
			"SelectRange",
			"String",
		},
		"relativetimeformat": {
			"Format",
			"FormatToParts",
			"ResolvedOptions",
		},
		"segmenter": {
			"All",
			"Containing",
			"ContainingByte",
			"ResolvedOptions",
			"Segment",
		},
	}
}

func publicSurfaceLedger() map[string][]string {
	return map[string][]string{
		".": {
			"Bool",
			"Collator",
			"DateTimeFormat",
			"DisplayNames",
			"DurationFormat",
			"ErrInvalidCode",
			"ErrInvalidKey",
			"ErrInvalidOption",
			"ErrInvalidValue",
			"ErrUnsupportedBackend",
			"ErrUnsupportedLocale",
			"ErrUnsupportedOption",
			"Error",
			"ErrorKind",
			"GetCanonicalLocales",
			"Int",
			"InvalidCode",
			"InvalidKey",
			"InvalidOption",
			"InvalidValue",
			"ListFormat",
			"Locale",
			"NumberFormat",
			"PluralRules",
			"RelativeTimeFormat",
			"Segmenter",
			"String",
			"SupportedCalendars",
			"SupportedCollations",
			"SupportedCurrencies",
			"SupportedNumberingSystems",
			"SupportedTimeZones",
			"SupportedUnits",
			"UnsupportedBackend",
			"UnsupportedLocale",
			"UnsupportedOption",
		},
		"collator": {
			"AccentSensitivity",
			"BaseSensitivity",
			"BestFitLocaleMatcher",
			"CaseFirst",
			"CaseSensitivity",
			"Collator",
			"FalseCaseFirst",
			"LocaleMatcher",
			"LowerCaseFirst",
			"LookupLocaleMatcher",
			"New",
			"Options",
			"ResolvedOptions",
			"SearchUsage",
			"Sensitivity",
			"SortUsage",
			"SupportedLocalesOf",
			"UpperCaseFirst",
			"Usage",
			"VariantSensitivity",
		},
		"datetimeformat": {
			"BasicFormatMatcher",
			"BestFitFormatMatcher",
			"BestFitLocaleMatcher",
			"DateTimeFormat",
			"FieldStyle",
			"FormatMatcher",
			"FullDateTimeStyle",
			"H11HourCycle",
			"H12HourCycle",
			"H23HourCycle",
			"H24HourCycle",
			"HourCycle",
			"LocaleMatcher",
			"LongDateTimeStyle",
			"LongFieldStyle",
			"LongGenericTimeZoneName",
			"LongMonthStyle",
			"LongOffsetTimeZoneName",
			"LongTimeZoneName",
			"LookupLocaleMatcher",
			"MediumDateTimeStyle",
			"MonthStyle",
			"NarrowFieldStyle",
			"NarrowMonthStyle",
			"New",
			"NumericFieldStyle",
			"NumericMonthStyle",
			"NumericStyle",
			"Options",
			"Part",
			"PartDay",
			"PartDayPeriod",
			"PartEra",
			"PartFractionalSecondDigits",
			"PartHour",
			"PartLiteral",
			"PartMinute",
			"PartMonth",
			"PartRelatedYear",
			"PartSecond",
			"PartTimeZoneName",
			"PartType",
			"PartUnknown",
			"PartWeekday",
			"PartYear",
			"PartYearName",
			"RangePart",
			"RangeSource",
			"ResolvedOptions",
			"ShortDateTimeStyle",
			"ShortFieldStyle",
			"ShortGenericTimeZoneName",
			"ShortMonthStyle",
			"ShortOffsetTimeZoneName",
			"ShortTimeZoneName",
			"SourceEndRange",
			"SourceShared",
			"SourceStartRange",
			"Style",
			"SupportedLocalesOf",
			"TimeZoneName",
			"TwoDigitFieldStyle",
			"TwoDigitMonthStyle",
		},
		"displaynames": {
			"BestFitLocaleMatcher",
			"Calendar",
			"CodeFallback",
			"Currency",
			"DateTimeField",
			"DialectLanguageDisplay",
			"DisplayNames",
			"Fallback",
			"Language",
			"LanguageDisplay",
			"LocaleMatcher",
			"LongStyle",
			"LookupLocaleMatcher",
			"NarrowStyle",
			"New",
			"NoneFallback",
			"Options",
			"Region",
			"ResolvedOptions",
			"Script",
			"ShortStyle",
			"StandardLanguageDisplay",
			"Style",
			"SupportedLocalesOf",
			"Type",
		},
		"durationformat": {
			"AlwaysDisplay",
			"AutoDisplay",
			"BestFitLocaleMatcher",
			"Day",
			"DigitalStyle",
			"Display",
			"Duration",
			"DurationFormat",
			"Hour",
			"LocaleMatcher",
			"LongStyle",
			"LongUnitStyle",
			"LookupLocaleMatcher",
			"Microsecond",
			"Millisecond",
			"Minute",
			"Month",
			"NarrowStyle",
			"NarrowUnitStyle",
			"Nanosecond",
			"New",
			"NumericUnitStyle",
			"Options",
			"Part",
			"PartDecimal",
			"PartFraction",
			"PartGroup",
			"PartInfinity",
			"PartInteger",
			"PartLiteral",
			"PartMinusSign",
			"PartNaN",
			"PartPlusSign",
			"PartType",
			"PartUnit",
			"ResolvedOptions",
			"Second",
			"ShortStyle",
			"ShortUnitStyle",
			"Style",
			"SupportedLocalesOf",
			"TwoDigitUnitStyle",
			"Unit",
			"UnitStyle",
			"Week",
			"Year",
		},
		"listformat": {
			"BestFitLocaleMatcher",
			"Conjunction",
			"Disjunction",
			"ListFormat",
			"LocaleMatcher",
			"LongStyle",
			"LookupLocaleMatcher",
			"NarrowStyle",
			"New",
			"Options",
			"Part",
			"PartElement",
			"PartLiteral",
			"PartType",
			"ResolvedOptions",
			"ShortStyle",
			"Style",
			"SupportedLocalesOf",
			"Type",
			"Unit",
		},
		"locale": {
			"FromTag",
			"List",
			"Locale",
			"New",
			"Options",
			"Parse",
			"ParseList",
			"TextInfo",
			"WeekInfo",
		},
		"numberformat": {
			"AccountingCurrencySign",
			"AlwaysSignDisplay",
			"AutoRoundingPriority",
			"AutoSignDisplay",
			"AutoTrailingZeroDisplay",
			"BestFitLocaleMatcher",
			"BigInt",
			"CeilRoundingMode",
			"CompactDisplay",
			"CompactNotation",
			"Currency",
			"CurrencyDisplay",
			"CurrencyDisplayCode",
			"CurrencyDisplayName",
			"CurrencyDisplayNarrowSymbol",
			"CurrencyDisplaySymbol",
			"CurrencySign",
			"CurrencyStyle",
			"Decimal",
			"DecimalStyle",
			"EngineeringNotation",
			"ExceptZeroSignDisplay",
			"ExpandRoundingMode",
			"Float",
			"FloorRoundingMode",
			"HalfCeilRoundingMode",
			"HalfEvenRoundingMode",
			"HalfExpandRoundingMode",
			"HalfFloorRoundingMode",
			"HalfTruncRoundingMode",
			"Int",
			"LessPrecisionRoundingPriority",
			"LocaleMatcher",
			"LongCompactDisplay",
			"LongUnitDisplay",
			"LookupLocaleMatcher",
			"MorePrecisionRoundingPriority",
			"NarrowUnitDisplay",
			"NegativeSignDisplay",
			"NeverSignDisplay",
			"New",
			"Notation",
			"NumberFormat",
			"Options",
			"Part",
			"PartApproximatelySign",
			"PartCompact",
			"PartCurrency",
			"PartDecimal",
			"PartExponentInteger",
			"PartExponentMinusSign",
			"PartExponentSeparator",
			"PartFraction",
			"PartGroup",
			"PartInfinity",
			"PartInteger",
			"PartLiteral",
			"PartMinusSign",
			"PartNaN",
			"PartPercentSign",
			"PartPlusSign",
			"PartType",
			"PartUnit",
			"PercentStyle",
			"RangePart",
			"RangeSource",
			"ResolvedOptions",
			"RoundingMode",
			"RoundingPriority",
			"ScientificNotation",
			"ShortCompactDisplay",
			"ShortUnitDisplay",
			"SignDisplay",
			"SourceEndRange",
			"SourceShared",
			"SourceStartRange",
			"StandardCurrencySign",
			"StandardNotation",
			"StripIfIntegerTrailingZeroDisplay",
			"Style",
			"SupportedLocalesOf",
			"TrailingZeroDisplay",
			"TruncRoundingMode",
			"Uint",
			"Unit",
			"UnitDisplay",
			"UnitStyle",
			"UseGrouping",
			"UseGroupingAlways",
			"UseGroupingAuto",
			"UseGroupingFalse",
			"UseGroupingMin2",
			"Value",
		},
		"pluralrules": {
			"AutoRoundingPriority",
			"AutoTrailingZeroDisplay",
			"BestFitLocaleMatcher",
			"BigInt",
			"Cardinal",
			"Category",
			"CeilRoundingMode",
			"CompactDisplay",
			"CompactNotation",
			"Decimal",
			"EngineeringNotation",
			"ExpandRoundingMode",
			"Few",
			"Float",
			"FloorRoundingMode",
			"HalfCeilRoundingMode",
			"HalfEvenRoundingMode",
			"HalfExpandRoundingMode",
			"HalfFloorRoundingMode",
			"HalfTruncRoundingMode",
			"Int",
			"LessPrecisionRoundingPriority",
			"LocaleMatcher",
			"LongCompactDisplay",
			"LookupLocaleMatcher",
			"Many",
			"MorePrecisionRoundingPriority",
			"New",
			"Notation",
			"One",
			"Options",
			"Ordinal",
			"Other",
			"PluralRules",
			"ResolvedOptions",
			"RoundingMode",
			"RoundingPriority",
			"ScientificNotation",
			"ShortCompactDisplay",
			"StandardNotation",
			"StripIfIntegerTrailingZeroDisplay",
			"SupportedLocalesOf",
			"TrailingZeroDisplay",
			"TruncRoundingMode",
			"Two",
			"Type",
			"Uint",
			"Value",
			"Zero",
		},
		"relativetimeformat": {
			"BestFitLocaleMatcher",
			"Day",
			"Decimal",
			"Float",
			"Hour",
			"Int",
			"LocaleMatcher",
			"LongStyle",
			"LookupLocaleMatcher",
			"Minute",
			"Month",
			"NarrowStyle",
			"New",
			"Numeric",
			"NumericAlways",
			"NumericAuto",
			"Options",
			"Part",
			"PartDecimal",
			"PartFraction",
			"PartGroup",
			"PartInfinity",
			"PartInteger",
			"PartLiteral",
			"PartMinusSign",
			"PartNaN",
			"PartPlusSign",
			"PartType",
			"Quarter",
			"RelativeTimeFormat",
			"ResolvedOptions",
			"Second",
			"ShortStyle",
			"Style",
			"SupportedLocalesOf",
			"Uint",
			"Unit",
			"Value",
			"Week",
			"Year",
		},
		"segmenter": {
			"BestFitLocaleMatcher",
			"Granularity",
			"GraphemeGranularity",
			"LocaleMatcher",
			"LookupLocaleMatcher",
			"New",
			"Options",
			"ResolvedOptions",
			"Segment",
			"Segmenter",
			"Segments",
			"SentenceGranularity",
			"SupportedLocalesOf",
			"WordGranularity",
		},
	}
}
