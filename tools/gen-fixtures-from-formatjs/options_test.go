package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFormatJSOptionsPreserveSourceSemantics(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		options string
		want    string
	}{
		{name: "variable", options: "{minimumFractionDigits: digits}"},
		{name: "template value", options: "{currency: `USD`, style: 'currency'}"},
		{name: "regex value", options: "{currency: /USD/, style: 'currency'}"},
		{name: "arithmetic", options: "{minimumFractionDigits: 1 + 1}"},
		{name: "spread", options: "{minimumFractionDigits: 2, ...other}"},
		{name: "computed key", options: "{['minimumFractionDigits']: 2}"},
		{name: "shorthand", options: "{minimumFractionDigits}"},
		{name: "nested value", options: "{minimumFractionDigits: {value: 2}}"},
		{name: "duplicate number", options: "{minimumFractionDigits: 1, minimumFractionDigits: 2}", want: `{"minimumFractionDigits":2}`},
		{name: "duplicate types", options: "{useGrouping: 'auto', useGrouping: false}", want: `{"useGrouping":false}`},
		{name: "comments", options: "{ /* minimumFractionDigits: 9 */ minimumFractionDigits: 2, // trailing comment\n}", want: `{"minimumFractionDigits":2}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			source := t.TempDir()
			out := t.TempDir()
			for _, name := range []string{
				formatJSNumberFormatPackageDir, formatJSPluralRulesPackageDir,
				formatJSDateTimeFormatPackageDir, formatJSLocalePackageDir,
				formatJSCanonicalLocalesPackageDir, formatJSListFormatPackageDir,
				formatJSRelativeTimeFormatPackageDir, formatJSDurationFormatPackageDir,
			} {
				if err := os.MkdirAll(formatJSPackageTestsRoot(source, name), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			mustWriteFile(t, filepath.Join(formatJSPackageTestsRoot(source, formatJSNumberFormatPackageDir), "options.test.ts"),
				"var nf = new NumberFormat('de', {style: 'percent'});\nvar nf = new NumberFormat('en', "+tc.options+");\nexpect(nf.format(1)).toBe('1.00');\n")
			if err := run([]string{"-formatjs", source, "-out", out}); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(formatJSConformanceRoot(out, "numberformat"), "options-test-ts.json")
			if tc.want == "" {
				assertPathAbsent(t, "ununderstood options fixture", path)
				assertFileContainsAll(t, "source-owned skip", filepath.Join(out, repositorySkipListFile),
					formatJSNumberFormatTestSourcePrefix+"options.test.ts", skipCategoryUnsupportedExtractorShape)
				return
			}
			fixtures := mustReadFixtures(t, path)
			if len(fixtures) != 1 {
				t.Fatalf("fixtures = %d, want 1", len(fixtures))
			}
			assertJSONEqual(t, "exported options", fixtures[0].Options, tc.want)
			assertJSONEqual(t, "skip list", mustReadSkips(t, filepath.Join(out, repositorySkipListFile)), "[]")
		})
	}
}

func TestUnunderstoodConstructorDoesNotReuseEarlierOptions(t *testing.T) {
	t.Parallel()
	for _, options := range []string{"{minimumFractionDigits: digits}", "options", "{minimumFractionDigits: 2} as any"} {
		t.Run(options, func(t *testing.T) {
			t.Parallel()
			source := `const nf = new NumberFormat('en', {minimumFractionDigits: 2});
expect(nf.format(1)).toBe('1.00');
const nf = new NumberFormat('en', ` + options + `);
expect(nf.format(2)).toBe('2.00');`
			fixtures := extractNumberFormatFixtures("options.test.ts", scanJSSource(source))
			if len(fixtures) != 1 {
				t.Fatalf("extracted %d fixtures, want only the first understood declaration", len(fixtures))
			}
			assertJSONEqual(t, "understood constructor", fixtures[0].Options, `{"minimumFractionDigits":2}`)
		})
	}
}

func TestRunUnsupportedLocaleConstructorBlocksEarlierDeclaration(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ constructor, packageDir, outputPackage, options, assertion string }{
		{"NumberFormat", formatJSNumberFormatPackageDir, "numberformat", "{}", "expect(f.format(1)).toBe('1');"},
		{"PluralRules", formatJSPluralRulesPackageDir, "pluralrules", "{}", "expect(f.select(1)).toBe('one');"},
		{"DateTimeFormat", formatJSDateTimeFormatPackageDir, "datetimeformat", "{year: 'numeric', month: 'short', day: 'numeric', timeZone: 'UTC'}", "expect(f.format(instant)).toBe('Jan 1, 2020');"},
	} {
		for _, localeExpression := range []string{"localeName", "['fr', 'en']", "`fr`", "/fr/"} {
			t.Run(tc.constructor+"/"+localeExpression, func(t *testing.T) {
				t.Parallel()
				source, out := t.TempDir(), t.TempDir()
				for _, packageDir := range []string{
					formatJSNumberFormatPackageDir, formatJSPluralRulesPackageDir,
					formatJSDateTimeFormatPackageDir, formatJSLocalePackageDir,
					formatJSCanonicalLocalesPackageDir, formatJSListFormatPackageDir,
					formatJSRelativeTimeFormatPackageDir, formatJSDurationFormatPackageDir,
				} {
					if err := os.MkdirAll(formatJSPackageTestsRoot(source, packageDir), 0o755); err != nil {
						t.Fatal(err)
					}
				}
				data := "const instant = new Date('2020-01-01T00:00:00Z');\n" +
					"it('supported', () => { const f = new " + tc.constructor + "('en', " + tc.options + "); " + tc.assertion + " });\n" +
					"it('unsupported', () => { const localeName = 'fr'; const f = new " + tc.constructor + "(" + localeExpression + ", " + tc.options + "); " + tc.assertion + " });\n"
				mustWriteFile(t, filepath.Join(formatJSPackageTestsRoot(source, tc.packageDir), "locale.test.ts"), data)
				if err := run([]string{"-formatjs", source, "-out", out}); err != nil {
					t.Fatal(err)
				}
				fixtures := mustReadFixtures(t, filepath.Join(formatJSConformanceRoot(out, tc.outputPackage), "locale-test-ts.json"))
				if len(fixtures) != 1 || fixtures[0].Locale != "en" {
					t.Fatalf("fixtures = %#v, want only the first understood constructor", fixtures)
				}
				assertFileContainsAll(t, "source-owned extraction audit", filepath.Join(out, repositorySkipListFile), formatJSTestSourcePrefix(tc.packageDir)+"locale.test.ts", skipCategoryPartialExtraction)
			})
		}
	}
}

func TestRunReassignmentInvalidatesConstructor(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ constructor, packageDir, outputPackage, assertion string }{
		{"NumberFormat", formatJSNumberFormatPackageDir, "numberformat", "expect(f.format(1)).toBe('$1.00');"},
		{"NumberFormatParts", formatJSNumberFormatPackageDir, "numberformat", "expect(f.formatToParts(1)).toEqual([{type: 'integer', value: '1'}]);"},
		{"NumberFormatRange", formatJSNumberFormatPackageDir, "numberformat", "expect(f.formatRange(1, 2)).toBe('1–2');"},
		{"NumberFormatRangeParts", formatJSNumberFormatPackageDir, "numberformat", "expect(f.formatRangeToParts(1, 2)).toEqual([{type: 'integer', value: '1', source: 'startRange'}]);"},
		{"PluralRules", formatJSPluralRulesPackageDir, "pluralrules", "expect(f.select(1)).toBe('one');"},
		{"PluralRulesRange", formatJSPluralRulesPackageDir, "pluralrules", "expect(f.selectRange(1, 2)).toBe('other');"},
		{"DateTimeFormat", formatJSDateTimeFormatPackageDir, "datetimeformat", "expect(f.format(instant)).toBe('2020');"},
		{"DateTimeFormatRange", formatJSDateTimeFormatPackageDir, "datetimeformat", "expect(f.formatRange(instant, instant)).toBe('2020');"},
		{"DateTimeFormatRangeParts", formatJSDateTimeFormatPackageDir, "datetimeformat", "expect(f.formatRangeToParts(instant, instant)).toEqual([{type: 'year', value: '2020', source: 'shared'}]);"},
	} {
		t.Run(tc.constructor, func(t *testing.T) {
			t.Parallel()
			constructor := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(tc.constructor, "RangeParts"), "Range"), "Parts")
			for _, before := range []bool{false, true} {
				source, out := t.TempDir(), t.TempDir()
				for _, packageDir := range []string{
					formatJSNumberFormatPackageDir, formatJSPluralRulesPackageDir,
					formatJSDateTimeFormatPackageDir, formatJSLocalePackageDir,
					formatJSCanonicalLocalesPackageDir, formatJSListFormatPackageDir,
					formatJSRelativeTimeFormatPackageDir, formatJSDurationFormatPackageDir,
				} {
					if err := os.MkdirAll(formatJSPackageTestsRoot(source, packageDir), 0o755); err != nil {
						t.Fatal(err)
					}
				}
				data := "const instant = new Date('2020-01-01T00:00:00Z');\nlet f = new Intl." + constructor + "('en', {});\n"
				if before {
					data += tc.assertion + "\n"
				}
				data += "f = new Intl." + constructor + "('fr', {});\n" + tc.assertion
				mustWriteFile(t, filepath.Join(formatJSPackageTestsRoot(source, tc.packageDir), "assignment.test.ts"), data)
				if err := run([]string{"-formatjs", source, "-out", out}); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(formatJSConformanceRoot(out, tc.outputPackage), "assignment-test-ts.json")
				category := skipCategoryUnsupportedExtractorShape
				if before {
					fixtures := mustReadFixtures(t, path)
					if len(fixtures) != 1 || fixtures[0].Locale != "en" {
						t.Fatalf("fixtures = %#v, want only pre-assignment observation", fixtures)
					}
					category = skipCategoryPartialExtraction
				} else {
					assertPathAbsent(t, "reassigned fixture", path)
				}
				assertFileContainsAll(t, "assignment audit", filepath.Join(out, repositorySkipListFile), formatJSTestSourcePrefix(tc.packageDir)+"assignment.test.ts", category)
			}
		})
	}
}

func TestConstructorAssignmentsRespectCodePositions(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		code string
		want int
	}{
		{"direct", "f = replacement;", 1},
		{"compact chained assignment", "other=f = replacement;", 1},
		{"spaced chained assignment", "other = f = replacement;", 1},
		{"comment trivia", "f /* binding */ = replacement;", 1},
		{"template substitution", "const text = `${f = replacement}`;", 1},
		{"comment", "/* f = replacement; */", 2},
		{"string", `const text = "f = replacement;";`, 2},
		{"template text", "const text = `f = replacement;`;", 2},
		{"regex", `const re = /f = replacement/;`, 2},
		{"property", "obj.f = replacement;", 2},
		{"equality", "f === replacement;", 2},
		{"arrow", "const identity = f => f;", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			data := "let f = new Intl.NumberFormat('en', {});\n" + tc.code + `
expect(f.format(1)).toBe('1');
expect(new Intl.NumberFormat('en').format(2)).toBe('2');`
			fixtures := extractNumberFormatFixtures("binding.test.ts", scanJSSource(data))
			if len(fixtures) != tc.want {
				t.Fatalf("fixtures = %d, want %d (independent literal must remain)", len(fixtures), tc.want)
			}
			if tc.want == 1 && (fixtures[0].Locale != "en" || fixtures[0].Expected == nil || *fixtures[0].Expected != "2") {
				t.Fatalf("fixture = %#v, want independent en literal formatting 2", fixtures[0])
			}
		})
	}
}
