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
)

type Category string

type RangeKey struct {
	Start Category
	End   Category
}

type Rule struct {
	Locale   string
	Category Category
	Expr     string
}

var categoryOrder = []Category{"zero", "one", "two", "few", "many", "other"}

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
	profile, err := readLocaleProfile(*profilePath)
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
	files := map[string]string{
		"cardinal_rules.go": renderRuleFile("Cardinal", cardinal, false),
		"ordinal_rules.go":  renderRuleFile("Ordinal", ordinal, true),
		"range_rules.go":    renderRangeFile(ranges),
		"categories.go":     renderCategoriesFile(cardinal, ordinal),
		"supported.go":      renderSupportedFile(sortedRuleLocales(cardinal)),
	}
	for name, source := range files {
		if err := writeGoFile(filepath.Join(*outDir, name), source); err != nil {
			return err
		}
	}
	return nil
}

func parsePluralRules(path string, locales []string) (map[string][]Rule, map[string][]Rule, error) {
	cardinal, err := parsePluralRulesJSON(path, "plurals-type-cardinal")
	if err != nil {
		return nil, nil, err
	}
	ordinalPath := filepath.Join(filepath.Dir(path), "ordinals.json")
	ordinal, err := parsePluralRulesJSON(ordinalPath, "plurals-type-ordinal")
	if err != nil {
		return nil, nil, err
	}
	return filterActiveRules(cardinal, locales), filterActiveRules(ordinal, locales), nil
}

func parsePluralRulesJSON(path, typ string) (map[string][]Rule, error) {
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
	var byLocale map[string]map[string]string
	if err := json.Unmarshal(doc.Supplemental[typ], &byLocale); err != nil {
		return nil, fmt.Errorf("parse %s %s: %w", path, typ, err)
	}
	out := make(map[string][]Rule, len(byLocale))
	for _, loc := range slices.Sorted(maps.Keys(byLocale)) {
		rules := byLocale[loc]
		for _, key := range slices.Sorted(maps.Keys(rules)) {
			category, ok := strings.CutPrefix(key, "pluralRule-count-")
			if !ok {
				continue
			}
			expr := strings.TrimSpace(strings.Split(rules[key], "@")[0])
			out[loc] = append(out[loc], Rule{Locale: loc, Category: Category(category), Expr: expr})
		}
	}
	return out, nil
}

func parsePluralRanges(path string, locales []string) (map[string]map[RangeKey]Category, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var doc struct {
		Supplemental struct {
			Plurals map[string]map[string]string `json:"plurals"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	out := make(map[string]map[RangeKey]Category)
	for _, loc := range slices.Sorted(maps.Keys(doc.Supplemental.Plurals)) {
		ranges := make(map[RangeKey]Category)
		for _, key := range slices.Sorted(maps.Keys(doc.Supplemental.Plurals[loc])) {
			start, end, ok := parseRangeRuleKey(key)
			if !ok {
				continue
			}
			ranges[RangeKey{Start: start, End: end}] = Category(doc.Supplemental.Plurals[loc][key])
		}
		if len(ranges) > 0 {
			out[loc] = ranges
		}
	}
	return filterActiveRanges(out, locales), nil
}

func parseRangeRuleKey(key string) (Category, Category, bool) {
	const prefix = "pluralRange-start-"
	key, ok := strings.CutPrefix(key, prefix)
	if !ok {
		return "", "", false
	}
	start, rest, ok := strings.Cut(key, "-end-")
	if !ok {
		return "", "", false
	}
	return Category(start), Category(rest), true
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

func renderRuleFile(name string, rules map[string][]Rule, ordinal bool) string {
	var b strings.Builder
	writeHeader(&b)
	b.WriteString("package plural\n\n")
	b.WriteString("import pluralop \"github.com/agentable/go-intl/internal/plural\"\n\n")
	fmt.Fprintf(&b, "func %sRule(loc string) (func(pluralop.OperandsRecord) pluralop.Category, bool) {\n", name)
	b.WriteString("\tswitch loc {\n")
	for _, loc := range sortedRuleLocales(rules) {
		fmt.Fprintf(&b, "\tcase %q:\n\t\treturn %s%s, true\n", loc, strings.ToLower(name), funcSuffix(loc))
	}
	b.WriteString("\t}\n")
	if ordinal {
		b.WriteString("\treturn cardinalOther, true\n")
	} else {
		b.WriteString("\treturn nil, false\n")
	}
	b.WriteString("}\n\n")
	for _, loc := range sortedRuleLocales(rules) {
		renderRuleFunc(&b, strings.ToLower(name)+funcSuffix(loc), rules[loc])
	}
	if !ordinal {
		b.WriteString("func cardinalOther(pluralop.OperandsRecord) pluralop.Category {\n\treturn pluralop.Other\n}\n")
	}
	return b.String()
}

func renderRuleFunc(b *strings.Builder, name string, rules []Rule) {
	fmt.Fprintf(b, "func %s(o pluralop.OperandsRecord) pluralop.Category {\n", name)
	for _, category := range categoryOrder {
		if category == "other" {
			continue
		}
		rule, ok := findRule(rules, category)
		if !ok || rule.Expr == "" {
			continue
		}
		fmt.Fprintf(b, "\tif %s {\n\t\treturn %s\n\t}\n", compileCondition(rule.Expr), categoryConst(category))
	}
	b.WriteString("\treturn pluralop.Other\n")
	b.WriteString("}\n\n")
}

func renderRangeFile(ranges map[string]map[RangeKey]Category) string {
	var b strings.Builder
	writeHeader(&b)
	b.WriteString("package plural\n\n")
	b.WriteString("import pluralop \"github.com/agentable/go-intl/internal/plural\"\n\n")
	b.WriteString("type rangeKey struct {\n\tstart pluralop.Category\n\tend   pluralop.Category\n}\n\n")
	b.WriteString(`func Range(loc, typ string, start, end pluralop.Category) (pluralop.Category, bool) {
	if typ == "ordinal" {
		return 0, false
	}
	ranges, ok := cardinalRanges[loc]
	if !ok {
		return 0, false
	}
	cat, ok := ranges[rangeKey{start: start, end: end}]
	return cat, ok
}

`)
	b.WriteString("var cardinalRanges = map[string]map[rangeKey]pluralop.Category{\n")
	for _, loc := range slices.Sorted(maps.Keys(ranges)) {
		fmt.Fprintf(&b, "\t%q: {\n", loc)
		for _, key := range sortedRangeKeys(ranges[loc]) {
			fmt.Fprintf(&b, "\t\t{start: %s, end: %s}: %s,\n", categoryConst(key.Start), categoryConst(key.End), categoryConst(ranges[loc][key]))
		}
		b.WriteString("\t},\n")
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
	b.WriteString("\tif typ == \"ordinal\" {\n\t\tswitch loc {\n")
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

func renderSupportedFile(locales []string) string {
	sorted := slices.Clone(locales)
	slices.Sort(sorted)

	var b strings.Builder
	writeHeader(&b)
	b.WriteString("package plural\n\n")
	b.WriteString("func SupportedLocales() []string {\n")
	b.WriteString("\treturn []string{\n")
	for _, loc := range sorted {
		fmt.Fprintf(&b, "\t\t%q,\n", loc)
	}
	b.WriteString("\t}\n")
	b.WriteString("}\n")
	return b.String()
}

func compileCondition(expr string) string {
	orTerms := splitRuleExpr(expr, " or ")
	if len(orTerms) == 1 {
		return compileAndCondition(orTerms[0])
	}
	compiled := make([]string, 0, len(orTerms))
	for _, term := range orTerms {
		compiled = append(compiled, "("+compileAndCondition(term)+")")
	}
	return strings.Join(compiled, " || ")
}

func compileAndCondition(expr string) string {
	andTerms := splitRuleExpr(expr, " and ")
	compiled := make([]string, 0, len(andTerms))
	for _, term := range andTerms {
		compiled = append(compiled, compileRelation(term))
	}
	return strings.Join(compiled, " && ")
}

func compileRelation(term string) string {
	term = strings.TrimSpace(term)
	left, right, negated, ok := splitRelation(term)
	if !ok {
		return "false"
	}
	value := compileOperand(left)
	parts := strings.Split(right, ",")
	compiled := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if start, end, ok := strings.Cut(part, ".."); ok {
			compiled = append(compiled, compileRange(value, start, end, negated))
			continue
		}
		compiled = append(compiled, compileEquals(value, part, negated))
	}
	if negated {
		return strings.Join(compiled, " && ")
	}
	if len(compiled) == 1 {
		return compiled[0]
	}
	return "(" + strings.Join(compiled, " || ") + ")"
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

func compileOperand(expr string) compiledOperand {
	operand, mod, hasMod := strings.Cut(expr, "%")
	operand = strings.TrimSpace(operand)
	value := operandExpr(operand)
	if !hasMod {
		return value
	}
	mod = strings.TrimSpace(mod)
	if value.exact {
		return compiledOperand{expr: value.expr + ".Mod(" + mod + ")", exact: true}
	}
	return compiledOperand{expr: "(" + value.expr + " % " + mod + ")"}
}

func operandExpr(operand string) compiledOperand {
	switch operand {
	case "n":
		return compiledOperand{expr: "o.N", exact: true}
	case "i":
		return compiledOperand{expr: "o.I", exact: true}
	case "v":
		return compiledOperand{expr: "o.V"}
	case "w":
		return compiledOperand{expr: "o.W"}
	case "f":
		return compiledOperand{expr: "o.F", exact: true}
	case "t":
		return compiledOperand{expr: "o.T", exact: true}
	case "c":
		return compiledOperand{expr: "o.C"}
	case "e":
		return compiledOperand{expr: "o.E"}
	default:
		return compiledOperand{expr: "0"}
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
	seen["other"] = true
	out := make([]Category, 0, len(seen))
	for _, category := range categoryOrder {
		if seen[category] {
			out = append(out, category)
		}
	}
	return out
}

func categorySliceLiteral(categories []Category) string {
	parts := make([]string, 0, len(categories))
	for _, category := range categories {
		parts = append(parts, categoryConst(category))
	}
	return "[]pluralop.Category{" + strings.Join(parts, ", ") + "}"
}

func categoryConst(category Category) string {
	switch category {
	case "zero":
		return "pluralop.Zero"
	case "one":
		return "pluralop.One"
	case "two":
		return "pluralop.Two"
	case "few":
		return "pluralop.Few"
	case "many":
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
			return strings.Compare(string(a.Start), string(b.Start))
		}
		return strings.Compare(string(a.End), string(b.End))
	})
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

func writeHeader(b *strings.Builder) {
	b.WriteString("// Code generated by tools/gen-plural-rules; DO NOT EDIT.\n")
	b.WriteString("// CLDR version: 48.1.0\n\n")
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
