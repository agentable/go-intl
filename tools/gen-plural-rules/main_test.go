package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestParsePluralRulesJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writePluralRuleFixtures(t, dir)

	cardinal, ordinal, err := parsePluralRules(filepath.Join(dir, "plurals.json"), testPluralLocales())
	if err != nil {
		t.Fatal(err)
	}
	if got := cardinal["en"][0]; got.Category != categoryOne || got.Expr != "i = 1 and v = 0" {
		t.Fatalf("cardinal en first rule = %+v", got)
	}
	if _, ok := cardinal["en-US"]; !ok {
		t.Fatal("cardinal en-US alias missing")
	}
	if _, ok := cardinal["zh-Hans-CN"]; !ok {
		t.Fatal("cardinal zh-Hans-CN parent rule missing")
	}
	if got := ordinal["en"][0]; got.Category != categoryOne || got.Expr != "n % 10 = 1 and n % 100 != 11" {
		t.Fatalf("ordinal en first rule = %+v", got)
	}
}

func TestParsePluralRulesRejectsMissingActiveLocale(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writePluralRuleFixtures(t, dir)

	_, _, err := parsePluralRules(filepath.Join(dir, "plurals.json"), []string{"en", "missing"})
	if err == nil || !strings.Contains(err.Error(), `cardinal rule for active locale "missing" is missing`) {
		t.Fatalf("parsePluralRules() error = %v, want missing cardinal locale context", err)
	}
}

func TestParsePluralRulesRejectsMissingActiveOrdinalLocale(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writePluralRuleFixtures(t, dir)
	writeTestFile(t, filepath.Join(dir, "ordinals.json"), `{"supplemental":{"plurals-type-ordinal":{
		"en":{"pluralRule-count-other":" @integer 0"}
	}}}`)

	_, _, err := parsePluralRules(filepath.Join(dir, "plurals.json"), []string{"en", "zh"})
	if err == nil || !strings.Contains(err.Error(), `ordinal rule for active locale "zh" is missing`) {
		t.Fatalf("parsePluralRules() error = %v, want missing ordinal locale context", err)
	}
}

func TestParsePluralRangesJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writePluralRuleFixtures(t, dir)

	ranges, err := parsePluralRanges(filepath.Join(dir, "pluralRanges.json"), testPluralLocales())
	if err != nil {
		t.Fatal(err)
	}
	if got := ranges["en"][RangeKey{Start: categoryOne, End: categoryOther}]; got != categoryOther {
		t.Fatalf("range en one-other = %q, want other", got)
	}
	if _, ok := ranges["en-US"]; !ok {
		t.Fatal("range en-US alias missing")
	}
}

func TestParsePluralRulesRejectsUnknownCategory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pluralsPath := filepath.Join(dir, "plurals.json")
	ordinalsPath := filepath.Join(dir, "ordinals.json")
	if err := os.WriteFile(pluralsPath, []byte(`{"supplemental":{"plurals-type-cardinal":{
		"en":{"pluralRule-count-one":"i = 1 @integer 1","pluralRule-count-sometimes":" @integer 2"}
	}}}`), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ordinalsPath, []byte(`{"supplemental":{"plurals-type-ordinal":{
		"en":{"pluralRule-count-other":" @integer 0"}
	}}}`), 0o666); err != nil {
		t.Fatal(err)
	}

	_, _, err := parsePluralRules(pluralsPath, []string{"en"})
	if err == nil || !strings.Contains(err.Error(), `unknown plural category "sometimes"`) {
		t.Fatalf("parsePluralRules() error = %v, want unknown category", err)
	}
}

func TestParsePluralRulesRejectsInvalidDataShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing supplemental",
			body: `{}`,
		},
		{
			name: "null supplemental",
			body: `{"supplemental":null}`,
		},
		{
			name: "missing plural type",
			body: `{"supplemental":{}}`,
		},
		{
			name: "null plural type",
			body: `{"supplemental":{"plurals-type-cardinal":null}}`,
		},
		{
			name: "empty plural type",
			body: `{"supplemental":{"plurals-type-cardinal":{}}}`,
		},
		{
			name: "null locale rules",
			body: `{"supplemental":{"plurals-type-cardinal":{"en":null}}}`,
		},
		{
			name: "empty locale rules",
			body: `{"supplemental":{"plurals-type-cardinal":{"en":{}}}}`,
		},
		{
			name: "locale without plural rules",
			body: `{"supplemental":{"plurals-type-cardinal":{"en":{"displayName":"English"}}}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "plurals.json")
			writeTestFile(t, path, tc.body)

			if _, err := parsePluralRulesJSON(path, cardinalRuleKind); err == nil {
				t.Fatal("parsePluralRulesJSON() succeeded, want error")
			}
		})
	}
}

func TestParsePluralRangesRejectsUnknownCategory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rangesPath := filepath.Join(dir, "pluralRanges.json")
	if err := os.WriteFile(rangesPath, []byte(`{"supplemental":{"plurals":{
		"en":{"pluralRange-start-one-end-sometimes":"other"}
	}}}`), 0o666); err != nil {
		t.Fatal(err)
	}

	_, err := parsePluralRanges(rangesPath, []string{"en"})
	if err == nil || !strings.Contains(err.Error(), `unknown plural range end category "sometimes"`) {
		t.Fatalf("parsePluralRanges() error = %v, want unknown range category", err)
	}

	if err := os.WriteFile(rangesPath, []byte(`{"supplemental":{"plurals":{
		"en":{"pluralRange-start-one-end-other":"sometimes"}
	}}}`), 0o666); err != nil {
		t.Fatal(err)
	}
	_, err = parsePluralRanges(rangesPath, []string{"en"})
	if err == nil || !strings.Contains(err.Error(), `unknown plural category "sometimes"`) {
		t.Fatalf("parsePluralRanges() error = %v, want unknown result category", err)
	}
}

func TestParsePluralRangesRejectsInvalidDataShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing supplemental",
			body: `{}`,
		},
		{
			name: "null supplemental",
			body: `{"supplemental":null}`,
		},
		{
			name: "missing plurals",
			body: `{"supplemental":{}}`,
		},
		{
			name: "null plurals",
			body: `{"supplemental":{"plurals":null}}`,
		},
		{
			name: "empty plurals",
			body: `{"supplemental":{"plurals":{}}}`,
		},
		{
			name: "null locale ranges",
			body: `{"supplemental":{"plurals":{"en":null}}}`,
		},
		{
			name: "empty locale ranges",
			body: `{"supplemental":{"plurals":{"en":{}}}}`,
		},
		{
			name: "locale without range rules",
			body: `{"supplemental":{"plurals":{"en":{"displayName":"English"}}}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "pluralRanges.json")
			writeTestFile(t, path, tc.body)

			if _, err := parsePluralRanges(path, []string{"en"}); err == nil {
				t.Fatal("parsePluralRanges() succeeded, want error")
			}
		})
	}
}

func TestRunGeneratesPluralRuleFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writePluralRuleFixtures(t, dir)
	out := filepath.Join(dir, "out")

	err := run([]string{
		"-plurals", filepath.Join(dir, "plurals.json"),
		"-ranges", filepath.Join(dir, "pluralRanges.json"),
		"-out", out,
		"-profile", writeLocaleProfileFixture(t, dir),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"cardinal_rules.go", "ordinal_rules.go", "range_rules.go", "categories.go", "supported.go"} {
		raw, err := os.ReadFile(filepath.Join(out, name))
		if err != nil {
			t.Fatalf("read generated %s: %v", name, err)
		}
		if !strings.Contains(string(raw), "Code generated by tools/gen-plural-rules") {
			t.Fatalf("%s missing generated header:\n%s", name, raw)
		}
		if !strings.Contains(string(raw), "// CLDR version: 48.1.0") {
			t.Fatalf("%s header missing CLDR version read from package.json:\n%s", name, raw)
		}
	}
	cardinal, err := os.ReadFile(filepath.Join(out, "cardinal_rules.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !containsAll(string(cardinal), "func CardinalRule", "cardinalEnUS", "cardinalZhHansCN", "cardinalPl", "o.N.Mod(100).Between(3, 10)") {
		t.Fatalf("cardinal_rules.go missing compiled rules:\n%s", cardinal)
	}
	supported, err := os.ReadFile(filepath.Join(out, "supported.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !containsAll(string(supported), `import "slices"`, "var supportedLocales = [...]string{", "func SupportedLocales() []string", "return slices.Clone(supportedLocales[:])", `"en-US"`, `"zh-Hans-CN"`) {
		t.Fatalf("supported.go missing plural locale list:\n%s", supported)
	}
	rangeRules, err := os.ReadFile(filepath.Join(out, "range_rules.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !containsAll(string(rangeRules), `"cmp"`, `"slices"`, "type rangeRecord struct", "func CardinalRange(loc string, start, end pluralop.Category)", "slices.BinarySearchFunc(cardinalRanges[:]", "var cardinalRanges = [...]rangeRecord", `{loc: "en", start: pluralop.One, end: pluralop.Other, result: pluralop.Other}`) {
		t.Fatalf("range_rules.go missing fixed range table:\n%s", rangeRules)
	}
	if strings.Contains(string(rangeRules), "map[string]map[") {
		t.Fatalf("range_rules.go still uses mutable map table:\n%s", rangeRules)
	}
	if strings.Contains(string(rangeRules), "func Range(loc, typ string") {
		t.Fatalf("range_rules.go still carries plural type transport:\n%s", rangeRules)
	}
}

func TestSortedRangeKeysUsesRuntimeCategoryOrder(t *testing.T) {
	t.Parallel()

	got := sortedRangeKeys(map[RangeKey]Category{
		{Start: categoryOther, End: categoryOther}: categoryOther,
		{Start: categoryZero, End: categoryOther}:  categoryOther,
		{Start: categoryOne, End: categoryOther}:   categoryOther,
	})
	want := []RangeKey{
		{Start: categoryZero, End: categoryOther},
		{Start: categoryOne, End: categoryOther},
		{Start: categoryOther, End: categoryOther},
	}
	if len(got) != len(want) {
		t.Fatalf("sortedRangeKeys() length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sortedRangeKeys()[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestCategoriesForRulesUsesRuntimeCategoryOrderAndOtherFallback(t *testing.T) {
	t.Parallel()

	got := categoriesForRules([]Rule{
		{Category: categoryMany},
		{Category: categoryOne},
		{Category: categoryMany},
	})
	want := []Category{categoryOne, categoryMany, categoryOther}
	if !slices.Equal(got, want) {
		t.Fatalf("categoriesForRules() = %v, want %v", got, want)
	}

	got = categoriesForRules(nil)
	want = []Category{categoryOther}
	if !slices.Equal(got, want) {
		t.Fatalf("categoriesForRules(nil) = %v, want %v", got, want)
	}
}

func TestCompileCondition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{in: "i = 1 and v = 0", want: "o.I.Equal(1) && o.V == 0"},
		{in: "i % 10 = 2..4 and i % 100 != 12..14", want: "o.I.Mod(10).Between(2, 4) && o.I.Mod(100).OutsideRange(12, 14)"},
		{in: "n % 100 = 3..10", want: "o.N.Mod(100).Between(3, 10)"},
	}
	for _, tc := range tests {
		got, err := compileCondition(tc.in)
		if err != nil {
			t.Fatalf("compileCondition(%q) error = %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("compileCondition(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestCompileConditionRejectsUnsupportedGrammar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "relation", in: "i within 2..4", want: `unsupported plural relation "i within 2..4"`},
		{name: "operand", in: "j = 1", want: `unsupported plural operand "j"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := compileCondition(tc.in); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("compileCondition(%q) error = %v, want %q", tc.in, err, tc.want)
			}
		})
	}
}

func TestRenderRuleFileRejectsUnsupportedCondition(t *testing.T) {
	t.Parallel()

	_, err := renderRuleFile(cardinalRuleKind, map[string][]Rule{
		"en": {
			{Locale: "en", Category: categoryOne, Expr: "j = 1"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), `compile cardinalEn one rule for en: unsupported plural operand "j"`) {
		t.Fatalf("renderRuleFile() error = %v, want unsupported operand context", err)
	}
}

func writePluralRuleFixtures(t *testing.T, dir string) {
	t.Helper()
	plurals := `{"supplemental":{"plurals-type-cardinal":{
		"en":{"pluralRule-count-one":"i = 1 and v = 0 @integer 1","pluralRule-count-other":" @integer 0, 2"},
		"pl":{"pluralRule-count-one":"i = 1 and v = 0 @integer 1","pluralRule-count-few":"v = 0 and i % 10 = 2..4 and i % 100 != 12..14 @integer 2","pluralRule-count-other":" @integer 0"},
		"ar":{"pluralRule-count-few":"n % 100 = 3..10 @integer 3","pluralRule-count-other":" @integer 0"},
		"fr":{"pluralRule-count-one":"i = 0,1 @integer 0, 1","pluralRule-count-other":" @integer 2"},
		"zh":{"pluralRule-count-other":" @integer 0"}
	}}}`
	ordinals := `{"supplemental":{"plurals-type-ordinal":{
		"en":{"pluralRule-count-one":"n % 10 = 1 and n % 100 != 11 @integer 1","pluralRule-count-other":" @integer 0"},
		"pl":{"pluralRule-count-other":" @integer 0"},
		"ar":{"pluralRule-count-other":" @integer 0"},
		"fr":{"pluralRule-count-one":"n = 1 @integer 1","pluralRule-count-other":" @integer 0"},
		"zh":{"pluralRule-count-other":" @integer 0"}
	}}}`
	ranges := `{"supplemental":{"plurals":{
		"en":{"pluralRange-start-one-end-other":"other","pluralRange-start-other-end-other":"other"},
		"fr":{"pluralRange-start-one-end-one":"one","pluralRange-start-one-end-other":"other","pluralRange-start-other-end-other":"other"}
	}}}`
	pkg := `{"name":"cldr-core","version":"48.1.0"}`
	for name, data := range map[string]string{"plurals.json": plurals, "ordinals.json": ordinals, "pluralRanges.json": ranges, "package.json": pkg} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o666); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}

func writeTestFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0o666); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func testPluralLocales() []string {
	return []string{"ar", "en", "en-US", "fr", "pl", "zh", "zh-Hans", "zh-Hans-CN"}
}

func writeLocaleProfileFixture(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "locale-profile.json")
	profile := `{"locales":["ar","en","en-US","fr","pl","zh","zh-Hans","zh-Hans-CN"]}`
	if err := os.WriteFile(path, []byte(profile), 0o666); err != nil {
		t.Fatalf("write locale profile: %v", err)
	}
	return path
}

func containsAll(s string, needles ...string) bool {
	for _, needle := range needles {
		if !strings.Contains(s, needle) {
			return false
		}
	}
	return true
}

func TestReadCLDRVersion(t *testing.T) {
	t.Parallel()

	t.Run("beside plurals", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"version":"99.9.9"}`), 0o666); err != nil {
			t.Fatal(err)
		}
		got, err := readCLDRVersion(filepath.Join(dir, "plurals.json"))
		if err != nil {
			t.Fatal(err)
		}
		if got != "99.9.9" {
			t.Fatalf("readCLDRVersion = %q, want 99.9.9", got)
		}
	})

	t.Run("one directory up (real layout)", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"version":"48.1.0"}`), 0o666); err != nil {
			t.Fatal(err)
		}
		supplemental := filepath.Join(root, "supplemental")
		if err := os.MkdirAll(supplemental, 0o777); err != nil {
			t.Fatal(err)
		}
		got, err := readCLDRVersion(filepath.Join(supplemental, "plurals.json"))
		if err != nil {
			t.Fatal(err)
		}
		if got != "48.1.0" {
			t.Fatalf("readCLDRVersion = %q, want 48.1.0", got)
		}
	})

	t.Run("missing errors", func(t *testing.T) {
		t.Parallel()
		if _, err := readCLDRVersion(filepath.Join(t.TempDir(), "plurals.json")); err == nil {
			t.Fatal("readCLDRVersion did not error on missing package.json")
		}
	})
}
