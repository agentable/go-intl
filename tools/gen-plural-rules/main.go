package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	pluralop "github.com/agentable/go-intl/internal/plural"
	"github.com/agentable/go-intl/tools/internal/localeprofile"
)

type Category = pluralop.Category

const (
	categoryZero  Category = pluralop.Zero
	categoryOne   Category = pluralop.One
	categoryTwo   Category = pluralop.Two
	categoryFew   Category = pluralop.Few
	categoryMany  Category = pluralop.Many
	categoryOther Category = pluralop.Other
)

type pluralRuleKind string

const (
	cardinalRuleKind pluralRuleKind = "cardinal"
	ordinalRuleKind  pluralRuleKind = "ordinal"

	ordinalRulesFilename  = "ordinals.json"
	pluralRangesKey       = "plurals"
	pluralRuleCountPrefix = "pluralRule-count-"
)

func (kind pluralRuleKind) cldrTypeKey() string {
	return "plurals-type-" + string(kind)
}

func (kind pluralRuleKind) outputFile() string {
	return string(kind) + "_rules.go"
}

func (kind pluralRuleKind) title() string {
	switch kind {
	case cardinalRuleKind:
		return "Cardinal"
	case ordinalRuleKind:
		return "Ordinal"
	default:
		return ""
	}
}

type RangeKey struct {
	Start Category
	End   Category
}

type Rule struct {
	Locale   string
	Category Category
	Expr     string
}

var categoryOrder = [...]Category{
	categoryZero,
	categoryOne,
	categoryTwo,
	categoryFew,
	categoryMany,
	categoryOther,
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("gen-plural-rules", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	pluralsPath := fs.String("plurals", "", "path to cldr-core/supplemental/plurals.json")
	rangesPath := fs.String("ranges", "", "path to cldr-core/supplemental/pluralRanges.json")
	outDir := fs.String("out", "", "output directory for internal/cldr/plural")
	profilePath := fs.String("profile", "", "path to tools/locale-profile.json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *pluralsPath == "" || *rangesPath == "" || *outDir == "" || *profilePath == "" {
		return fmt.Errorf("usage: gen-plural-rules -plurals <plurals.json> -ranges <pluralRanges.json> -out <internal/cldr/plural> -profile <tools/locale-profile.json>")
	}
	version, err := readCLDRVersion(*pluralsPath)
	if err != nil {
		return err
	}
	cldrVersion = version
	profile, err := localeprofile.Read(*profilePath)
	if err != nil {
		return err
	}
	cardinal, ordinal, err := parsePluralRules(*pluralsPath, profile.Locales)
	if err != nil {
		return err
	}
	ranges, err := parsePluralRanges(*rangesPath, profile.Locales)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*outDir, 0o777); err != nil {
		return fmt.Errorf("mkdir %s: %w", *outDir, err)
	}
	cardinalSource, err := renderRuleFile(cardinalRuleKind, cardinal)
	if err != nil {
		return err
	}
	ordinalSource, err := renderRuleFile(ordinalRuleKind, ordinal)
	if err != nil {
		return err
	}
	files := map[string]string{
		cardinalRuleKind.outputFile(): cardinalSource,
		ordinalRuleKind.outputFile():  ordinalSource,
		"range_rules.go":              renderRangeFile(ranges),
		"categories.go":               renderCategoriesFile(cardinal, ordinal),
		"supported.go":                renderSupportedFile(cardinal),
	}
	for name, source := range files {
		if err := writeGoFile(filepath.Join(*outDir, name), source); err != nil {
			return err
		}
	}
	return nil
}

func parsePluralRules(path string, locales []string) (map[string][]Rule, map[string][]Rule, error) {
	cardinal, err := parsePluralRulesJSON(path, cardinalRuleKind)
	if err != nil {
		return nil, nil, err
	}
	ordinalPath := filepath.Join(filepath.Dir(path), ordinalRulesFilename)
	ordinal, err := parsePluralRulesJSON(ordinalPath, ordinalRuleKind)
	if err != nil {
		return nil, nil, err
	}
	return filterActiveRules(cardinal, locales), filterActiveRules(ordinal, locales), nil
}

func parsePluralRulesJSON(path string, kind pluralRuleKind) (map[string][]Rule, error) {
	raw, err := readSupplementalMap(path, kind.cldrTypeKey())
	if err != nil {
		return nil, err
	}
	var byLocale map[string]map[string]string
	typeKey := kind.cldrTypeKey()
	if err := json.Unmarshal(raw, &byLocale); err != nil {
		return nil, fmt.Errorf("parse %s %s: %w", path, typeKey, err)
	}
	if len(byLocale) == 0 {
		return nil, fmt.Errorf("parse %s %s: plural locale map missing", path, typeKey)
	}
	out := make(map[string][]Rule, len(byLocale))
	for _, loc := range slices.Sorted(maps.Keys(byLocale)) {
		rules, err := parsePluralLocaleRules(path, typeKey, loc, byLocale[loc])
		if err != nil {
			return nil, err
		}
		out[loc] = rules
	}
	return out, nil
}

func readSupplementalMap(path, key string) (json.RawMessage, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var doc struct {
		Supplemental map[string]json.RawMessage `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if doc.Supplemental == nil {
		return nil, fmt.Errorf("parse %s: supplemental data missing", path)
	}
	payload, ok := doc.Supplemental[key]
	if !ok || len(payload) == 0 {
		return nil, fmt.Errorf("parse %s %s: data missing", path, key)
	}
	return payload, nil
}

func parsePluralLocaleRules(path, typeKey, loc string, rules map[string]string) ([]Rule, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("parse %s %s %s: plural rules missing", path, typeKey, loc)
	}
	out := make([]Rule, 0, len(rules))
	for _, key := range slices.Sorted(maps.Keys(rules)) {
		categoryName, ok := strings.CutPrefix(key, pluralRuleCountPrefix)
		if !ok {
			continue
		}
		category, ok := pluralop.ParseCategory(categoryName)
		if !ok {
			return nil, fmt.Errorf("parse %s %s %s: unknown plural category %q", path, typeKey, loc, categoryName)
		}
		expr := strings.TrimSpace(strings.Split(rules[key], "@")[0])
		out = append(out, Rule{Locale: loc, Category: category, Expr: expr})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("parse %s %s %s: plural rules missing", path, typeKey, loc)
	}
	return out, nil
}

func parsePluralRanges(path string, locales []string) (map[string]map[RangeKey]Category, error) {
	raw, err := readSupplementalMap(path, pluralRangesKey)
	if err != nil {
		return nil, err
	}
	var byLocale map[string]map[string]string
	if err := json.Unmarshal(raw, &byLocale); err != nil {
		return nil, fmt.Errorf("parse %s pluralRanges: %w", path, err)
	}
	if len(byLocale) == 0 {
		return nil, fmt.Errorf("parse %s pluralRanges: locale map missing", path)
	}
	out := make(map[string]map[RangeKey]Category)
	for _, loc := range slices.Sorted(maps.Keys(byLocale)) {
		ranges, err := parsePluralLocaleRanges(path, loc, byLocale[loc])
		if err != nil {
			return nil, err
		}
		out[loc] = ranges
	}
	return filterActiveRanges(out, locales), nil
}

func parsePluralLocaleRanges(path, loc string, rules map[string]string) (map[RangeKey]Category, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("parse %s pluralRanges %s: range rules missing", path, loc)
	}
	ranges := make(map[RangeKey]Category)
	for _, key := range slices.Sorted(maps.Keys(rules)) {
		start, end, ok, err := parseRangeRuleKey(key)
		if err != nil {
			return nil, fmt.Errorf("parse %s pluralRanges %s key %q: %w", path, loc, key, err)
		}
		if !ok {
			continue
		}
		resultName := rules[key]
		result, ok := pluralop.ParseCategory(resultName)
		if !ok {
			return nil, fmt.Errorf("parse %s pluralRanges %s key %q: unknown plural category %q", path, loc, key, resultName)
		}
		ranges[RangeKey{Start: start, End: end}] = result
	}
	if len(ranges) == 0 {
		return nil, fmt.Errorf("parse %s pluralRanges %s: range rules missing", path, loc)
	}
	return ranges, nil
}

func parseRangeRuleKey(key string) (Category, Category, bool, error) {
	const prefix = "pluralRange-start-"
	key, ok := strings.CutPrefix(key, prefix)
	if !ok {
		return 0, 0, false, nil
	}
	start, rest, ok := strings.Cut(key, "-end-")
	if !ok {
		return 0, 0, false, nil
	}
	startCategory, ok := pluralop.ParseCategory(start)
	if !ok {
		return 0, 0, false, fmt.Errorf("unknown plural range start category %q", start)
	}
	endCategory, ok := pluralop.ParseCategory(rest)
	if !ok {
		return 0, 0, false, fmt.Errorf("unknown plural range end category %q", rest)
	}
	return startCategory, endCategory, true, nil
}

func filterActiveRules(rules map[string][]Rule, locales []string) map[string][]Rule {
	out := make(map[string][]Rule, len(locales))
	for _, loc := range locales {
		source := pluralRuleSource(loc, rules)
		if rules[source] != nil {
			out[loc] = cloneRules(loc, rules[source])
		}
	}
	return out
}

func filterActiveRanges(ranges map[string]map[RangeKey]Category, locales []string) map[string]map[RangeKey]Category {
	out := make(map[string]map[RangeKey]Category, len(locales))
	for _, loc := range locales {
		source := pluralRuleSource(loc, ranges)
		if ranges[source] != nil {
			out[loc] = maps.Clone(ranges[source])
		}
	}
	return out
}

func pluralRuleSource[T any](locale string, rules map[string]T) string {
	for candidate := locale; candidate != ""; {
		if _, ok := rules[candidate]; ok {
			return candidate
		}
		idx := strings.LastIndexByte(candidate, '-')
		if idx < 0 {
			return ""
		}
		candidate = candidate[:idx]
	}
	return ""
}

func cloneRules(locale string, rules []Rule) []Rule {
	out := make([]Rule, len(rules))
	for i, rule := range rules {
		rule.Locale = locale
		out[i] = rule
	}
	return out
}

func renderRuleFile(kind pluralRuleKind, rules map[string][]Rule) (string, error) {
	name := kind.title()
	var b strings.Builder
	writeHeader(&b)
	b.WriteString("package plural\n\n")
	b.WriteString("import pluralop \"github.com/agentable/go-intl/internal/plural\"\n\n")
	fmt.Fprintf(&b, "func %sRule(loc string) (func(pluralop.OperandsRecord) pluralop.Category, bool) {\n", name)
	b.WriteString("\tswitch loc {\n")
	for _, loc := range sortedRuleLocales(rules) {
		fmt.Fprintf(&b, "\tcase %q:\n\t\treturn %s%s, true\n", loc, kind, funcSuffix(loc))
	}
	b.WriteString("\t}\n")
	if kind == ordinalRuleKind {
		b.WriteString("\treturn cardinalOther, true\n")
	} else {
		b.WriteString("\treturn nil, false\n")
	}
	b.WriteString("}\n\n")
	for _, loc := range sortedRuleLocales(rules) {
		if err := renderRuleFunc(&b, string(kind)+funcSuffix(loc), rules[loc]); err != nil {
			return "", err
		}
	}
	if kind == cardinalRuleKind {
		b.WriteString("func cardinalOther(pluralop.OperandsRecord) pluralop.Category {\n\treturn pluralop.Other\n}\n")
	}
	return b.String(), nil
}

func renderRuleFunc(b *strings.Builder, name string, rules []Rule) error {
	fmt.Fprintf(b, "func %s(o pluralop.OperandsRecord) pluralop.Category {\n", name)
	for _, category := range categoryOrder[:] {
		if category == categoryOther {
			continue
		}
		rule, ok := findRule(rules, category)
		if !ok || rule.Expr == "" {
			continue
		}
		condition, err := compileCondition(rule.Expr)
		if err != nil {
			return fmt.Errorf("compile %s %s rule for %s: %w", name, category.String(), rule.Locale, err)
		}
		fmt.Fprintf(b, "\tif %s {\n\t\treturn %s\n\t}\n", condition, categoryConst(category))
	}
	b.WriteString("\treturn pluralop.Other\n")
	b.WriteString("}\n\n")
	return nil
}

func renderRangeFile(ranges map[string]map[RangeKey]Category) string {
	var b strings.Builder
	writeHeader(&b)
	b.WriteString("package plural\n\n")
	b.WriteString("import (\n")
	b.WriteString("\t\"cmp\"\n")
	b.WriteString("\t\"slices\"\n\n")
	b.WriteString("\tpluralop \"github.com/agentable/go-intl/internal/plural\"\n")
	b.WriteString(")\n\n")
	b.WriteString("type rangeRecord struct {\n\tloc    string\n\tstart  pluralop.Category\n\tend    pluralop.Category\n\tresult pluralop.Category\n}\n\n")
	b.WriteString(`func CardinalRange(loc string, start, end pluralop.Category) (pluralop.Category, bool) {
	target := rangeRecord{loc: loc, start: start, end: end}
	idx, ok := slices.BinarySearchFunc(cardinalRanges[:], target, compareRangeRecordKey)
	if !ok {
		return 0, false
	}
	return cardinalRanges[idx].result, true
}

func compareRangeRecordKey(a, b rangeRecord) int {
	if byLocale := cmp.Compare(a.loc, b.loc); byLocale != 0 {
		return byLocale
	}
	if byStart := cmp.Compare(a.start, b.start); byStart != 0 {
		return byStart
	}
	return cmp.Compare(a.end, b.end)
}

`)
	b.WriteString("var cardinalRanges = [...]rangeRecord{\n")
	for _, loc := range slices.Sorted(maps.Keys(ranges)) {
		for _, key := range sortedRangeKeys(ranges[loc]) {
			fmt.Fprintf(&b, "\t{loc: %q, start: %s, end: %s, result: %s},\n", loc, categoryConst(key.Start), categoryConst(key.End), categoryConst(ranges[loc][key]))
		}
	}
	b.WriteString("}\n")
	return b.String()
}

func renderCategoriesFile(cardinal, ordinal map[string][]Rule) string {
	var b strings.Builder
	writeHeader(&b)
	b.WriteString("package plural\n\n")
	b.WriteString("import pluralop \"github.com/agentable/go-intl/internal/plural\"\n\n")
	b.WriteString("func Categories(loc, typ string) []pluralop.Category {\n")
	fmt.Fprintf(&b, "\tif typ == %q {\n\t\tswitch loc {\n", ordinalRuleKind)
	for _, loc := range sortedRuleLocales(ordinal) {
		fmt.Fprintf(&b, "\t\tcase %q:\n\t\t\treturn %s\n", loc, categorySliceLiteral(categoriesForRules(ordinal[loc])))
	}
	b.WriteString("\t\t}\n\t\treturn []pluralop.Category{pluralop.Other}\n\t}\n")
	b.WriteString("\tswitch loc {\n")
	for _, loc := range sortedRuleLocales(cardinal) {
		fmt.Fprintf(&b, "\tcase %q:\n\t\treturn %s\n", loc, categorySliceLiteral(categoriesForRules(cardinal[loc])))
	}
	b.WriteString("\t}\n\treturn []pluralop.Category{pluralop.Other}\n}\n")
	return b.String()
}

func renderSupportedFile(rules map[string][]Rule) string {
	var b strings.Builder
	writeHeader(&b)
	b.WriteString("package plural\n\n")
	b.WriteString("import \"slices\"\n\n")
	b.WriteString("var supportedLocales = [...]string{\n")
	for _, loc := range sortedRuleLocales(rules) {
		fmt.Fprintf(&b, "\t%q,\n", loc)
	}
	b.WriteString("}\n\n")
	b.WriteString("func SupportedLocales() []string {\n")
	b.WriteString("\treturn slices.Clone(supportedLocales[:])\n")
	b.WriteString("}\n")
	return b.String()
}

func compileCondition(expr string) (string, error) {
	orTerms := splitRuleExpr(expr, " or ")
	if len(orTerms) == 1 {
		return compileAndCondition(orTerms[0])
	}
	compiled := make([]string, len(orTerms))
	for i, term := range orTerms {
		condition, err := compileAndCondition(term)
		if err != nil {
			return "", err
		}
		compiled[i] = "(" + condition + ")"
	}
	return strings.Join(compiled, " || "), nil
}

func compileAndCondition(expr string) (string, error) {
	andTerms := splitRuleExpr(expr, " and ")
	compiled := make([]string, len(andTerms))
	for i, term := range andTerms {
		relation, err := compileRelation(term)
		if err != nil {
			return "", err
		}
		compiled[i] = relation
	}
	return strings.Join(compiled, " && "), nil
}

func compileRelation(term string) (string, error) {
	term = strings.TrimSpace(term)
	left, right, negated, ok := splitRelation(term)
	if !ok {
		return "", fmt.Errorf("unsupported plural relation %q", term)
	}
	value, err := compileOperand(left)
	if err != nil {
		return "", err
	}
	parts := strings.Split(right, ",")
	compiled := make([]string, len(parts))
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if start, end, ok := strings.Cut(part, ".."); ok {
			compiled[i] = compileRange(value, start, end, negated)
			continue
		}
		compiled[i] = compileEquals(value, part, negated)
	}
	if negated {
		return strings.Join(compiled, " && "), nil
	}
	if len(compiled) == 1 {
		return compiled[0], nil
	}
	return "(" + strings.Join(compiled, " || ") + ")", nil
}

func splitRelation(term string) (left string, right string, negated bool, ok bool) {
	if left, right, ok := strings.Cut(term, " != "); ok {
		return strings.TrimSpace(left), right, true, true
	}
	if left, right, ok := strings.Cut(term, " = "); ok {
		return strings.TrimSpace(left), right, false, true
	}
	return "", "", false, false
}

type compiledOperand struct {
	expr  string
	exact bool
}

func compileOperand(expr string) (compiledOperand, error) {
	operand, mod, hasMod := strings.Cut(expr, "%")
	operand = strings.TrimSpace(operand)
	value, ok := operandExpr(operand)
	if !ok {
		return compiledOperand{}, fmt.Errorf("unsupported plural operand %q", operand)
	}
	if !hasMod {
		return value, nil
	}
	mod = strings.TrimSpace(mod)
	if value.exact {
		return compiledOperand{expr: value.expr + ".Mod(" + mod + ")", exact: true}, nil
	}
	return compiledOperand{expr: "(" + value.expr + " % " + mod + ")"}, nil
}

func operandExpr(operand string) (compiledOperand, bool) {
	switch operand {
	case "n":
		return compiledOperand{expr: "o.N", exact: true}, true
	case "i":
		return compiledOperand{expr: "o.I", exact: true}, true
	case "v":
		return compiledOperand{expr: "o.V"}, true
	case "w":
		return compiledOperand{expr: "o.W"}, true
	case "f":
		return compiledOperand{expr: "o.F", exact: true}, true
	case "t":
		return compiledOperand{expr: "o.T", exact: true}, true
	case "c":
		return compiledOperand{expr: "o.C"}, true
	case "e":
		return compiledOperand{expr: "o.E"}, true
	default:
		return compiledOperand{}, false
	}
}

func compileRange(value compiledOperand, start, end string, negated bool) string {
	start = strings.TrimSpace(start)
	end = strings.TrimSpace(end)
	if value.exact {
		if negated {
			return value.expr + ".OutsideRange(" + start + ", " + end + ")"
		}
		return value.expr + ".Between(" + start + ", " + end + ")"
	}
	if negated {
		return "(" + value.expr + " < " + start + " || " + value.expr + " > " + end + ")"
	}
	return "(" + value.expr + " >= " + start + " && " + value.expr + " <= " + end + ")"
}

func compileEquals(value compiledOperand, want string, negated bool) string {
	want = strings.TrimSpace(want)
	if value.exact {
		if negated {
			return value.expr + ".NotEqual(" + want + ")"
		}
		return value.expr + ".Equal(" + want + ")"
	}
	op := "=="
	if negated {
		op = "!="
	}
	return value.expr + " " + op + " " + want
}

func splitRuleExpr(expr, sep string) []string {
	raw := strings.Split(expr, sep)
	out := make([]string, 0, len(raw))
	for _, part := range raw {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func findRule(rules []Rule, category Category) (Rule, bool) {
	for _, rule := range rules {
		if rule.Category == category {
			return rule, true
		}
	}
	return Rule{}, false
}

func categoriesForRules(rules []Rule) []Category {
	seen := make(map[Category]bool, len(rules)+1)
	for _, rule := range rules {
		seen[rule.Category] = true
	}
	seen[categoryOther] = true
	out := make([]Category, 0, len(seen))
	for _, category := range categoryOrder[:] {
		if seen[category] {
			out = append(out, category)
		}
	}
	return out
}

func categorySliceLiteral(categories []Category) string {
	parts := make([]string, len(categories))
	for i, category := range categories {
		parts[i] = categoryConst(category)
	}
	return "[]pluralop.Category{" + strings.Join(parts, ", ") + "}"
}

func categoryConst(category Category) string {
	switch category {
	case categoryZero:
		return "pluralop.Zero"
	case categoryOne:
		return "pluralop.One"
	case categoryTwo:
		return "pluralop.Two"
	case categoryFew:
		return "pluralop.Few"
	case categoryMany:
		return "pluralop.Many"
	default:
		return "pluralop.Other"
	}
}

func sortedRuleLocales(rules map[string][]Rule) []string {
	return slices.Sorted(maps.Keys(rules))
}

func sortedRangeKeys(ranges map[RangeKey]Category) []RangeKey {
	return slices.SortedFunc(maps.Keys(ranges), func(a, b RangeKey) int {
		if a.Start != b.Start {
			return categoryOrderIndex(a.Start) - categoryOrderIndex(b.Start)
		}
		return categoryOrderIndex(a.End) - categoryOrderIndex(b.End)
	})
}

func categoryOrderIndex(category Category) int {
	for i, candidate := range categoryOrder[:] {
		if category == candidate {
			return i
		}
	}
	return len(categoryOrder)
}

func funcSuffix(locale string) string {
	var b strings.Builder
	upperNext := true
	for _, r := range locale {
		if r == '-' || r == '_' {
			upperNext = true
			continue
		}
		if upperNext {
			b.WriteString(strings.ToUpper(string(r)))
			upperNext = false
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// cldrVersion is the CLDR release the current run generates from, set in run()
// from the source cldr-core/package.json so the generated header can never drift
// from the data — the same single source of truth gen-cldr reads.
var cldrVersion string

func writeHeader(b *strings.Builder) {
	b.WriteString("// Code generated by tools/gen-plural-rules; DO NOT EDIT.\n")
	fmt.Fprintf(b, "// CLDR version: %s\n\n", cldrVersion)
}

// readCLDRVersion reads the CLDR release from the package.json that ships with
// the cldr-core data. In the real layout plurals.json lives at
// <cldr-core>/supplemental/plurals.json and the version is in
// <cldr-core>/package.json (one directory up); it also accepts a package.json
// beside plurals.json for flat fixtures.
func readCLDRVersion(pluralsPath string) (string, error) {
	dir := filepath.Dir(pluralsPath)
	for _, pkgPath := range []string{
		filepath.Join(dir, "package.json"),
		filepath.Join(dir, "..", "package.json"),
	} {
		data, err := os.ReadFile(pkgPath)
		if err != nil {
			continue
		}
		var pkg struct {
			Version string `json:"version"`
		}
		if err := json.Unmarshal(data, &pkg); err != nil {
			return "", fmt.Errorf("parse %s: %w", pkgPath, err)
		}
		if pkg.Version != "" {
			return pkg.Version, nil
		}
	}
	return "", fmt.Errorf("cldr version: no package.json with a version field near %s", pluralsPath)
}

func writeGoFile(path, source string) error {
	formatted, err := format.Source([]byte(source))
	if err != nil {
		return fmt.Errorf("format %s: %w\n%s", path, err, source)
	}
	if err := os.WriteFile(path, formatted, 0o666); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
