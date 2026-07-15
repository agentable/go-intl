package cldr

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestLoadLanguageMatchingPreservesOrderedContract(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteLanguageMatchingContainment(t, root)
	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "languageMatching.json"), `{
		"supplemental":{"languageMatching":{"written-new":{
			"paradigmLocales":{"_locales":["en","es-419"]},
			"matchVariables":{"$americas":{"_value":"019"}},
			"languageMatch":[
				{"_desired":"nn","_supported":"nb","_distance":20},
				{"_desired":"gsw","_supported":"de","_distance":4,"_oneway":true},
				{"_desired":"es-*-$americas","_supported":"es-*-$americas","_distance":4},
				{"_desired":"*-*-*","_supported":"*-*-*","_distance":4}
			]
		}}}
	}`)

	got, err := loadLanguageMatching(root)
	if err != nil {
		t.Fatalf("loadLanguageMatching() error = %v", err)
	}
	if !slices.Equal(got.ParadigmLocales, []string{"en", "es-419"}) {
		t.Fatalf("ParadigmLocales = %#v", got.ParadigmLocales)
	}
	if len(got.MatchVariables) != 1 {
		t.Fatalf("MatchVariables = %#v", got.MatchVariables)
	}
	variable := got.MatchVariables[0]
	if variable.Name != "americas" || !slices.Equal(variable.SourceRegions, []string{"019"}) ||
		!slices.Equal(variable.ExpandedRegions, []string{"005", "019", "021", "AR", "BR", "CA", "MX", "US"}) {
		t.Fatalf("MatchVariables[0] = %#v", variable)
	}
	wantRules := []LanguageMatchRule{
		{Desired: "nn", Supported: "nb", Distance: 20},
		{Desired: "gsw", Supported: "de", Distance: 4, OneWay: true},
		{Desired: "es-*-$americas", Supported: "es-*-$americas", Distance: 4},
		{Desired: "*-*-*", Supported: "*-*-*", Distance: 4},
	}
	if !slices.Equal(got.Rules, wantRules) {
		t.Fatalf("Rules = %#v, want source order %#v", got.Rules, wantRules)
	}
}

func TestLoadLanguageMatchingRejectsInvalidContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		paradigms string
		variables string
		rules     string
		want      string
	}{
		{name: "duplicate paradigm", paradigms: `"en","en"`, rules: catchAllLanguageMatchRule, want: "duplicate paradigm"},
		{name: "missing desired", paradigms: `"en"`, rules: `{"_supported":"en","_distance":1},` + catchAllLanguageMatchRule, want: "desired"},
		{name: "unknown variable", paradigms: `"en"`, rules: `{"_desired":"en-*-$missing","_supported":"en-*-US","_distance":1},` + catchAllLanguageMatchRule, want: "unknown match variable"},
		{name: "invalid distance", paradigms: `"en"`, rules: `{"_desired":"en","_supported":"fr","_distance":101},` + catchAllLanguageMatchRule, want: "distance"},
		{name: "missing catch all", paradigms: `"en"`, rules: `{"_desired":"en","_supported":"fr","_distance":1}`, want: "final catch-all"},
		{name: "invalid variable", paradigms: `"en"`, variables: `"$bad":{"_value":""}`, rules: catchAllLanguageMatchRule, want: "empty region list"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteLanguageMatchingContainment(t, root)
			raw := `{"supplemental":{"languageMatching":{"written-new":{` +
				`"paradigmLocales":{"_locales":[` + tc.paradigms + `]},` +
				`"matchVariables":{` + tc.variables + `},` +
				`"languageMatch":[` + tc.rules + `]}}}}`
			mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "languageMatching.json"), raw)
			if _, err := loadLanguageMatching(root); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("loadLanguageMatching() error = %v, want containing %q", err, tc.want)
			}
		})
	}
}

func TestPinnedLanguageMatchingContract(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", ".cldr-json", "node_modules")
	got, err := loadLanguageMatching(root)
	if err != nil {
		t.Fatalf("loadLanguageMatching(pinned) error = %v", err)
	}
	if len(got.Rules) != 378 {
		t.Fatalf("pinned rule count = %d, want 378", len(got.Rules))
	}
	if len(got.MatchVariables) != 4 {
		t.Fatalf("pinned variable count = %d, want 4", len(got.MatchVariables))
	}
	if !slices.ContainsFunc(got.Rules, func(rule LanguageMatchRule) bool {
		return rule.Desired == "gsw" && rule.Supported == "de" && rule.Distance == 4 && rule.OneWay
	}) {
		t.Fatal("pinned rules missing gsw -> de one-way witness")
	}
}

const catchAllLanguageMatchRule = `{"_desired":"*-*-*","_supported":"*-*-*","_distance":4}`

func mustWriteLanguageMatchingContainment(t *testing.T, root string) {
	t.Helper()
	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "territoryContainment.json"), `{
		"supplemental":{"territoryContainment":{
			"019":{"_contains":["021","005"]},
			"021":{"_contains":["CA","MX","US"]},
			"005":{"_contains":["AR","BR"]}
		}}
	}`)
}

func BenchmarkLoadLanguageMatching(b *testing.B) {
	root := filepath.Join("..", ".cldr-json", "node_modules")
	for b.Loop() {
		if _, err := loadLanguageMatching(root); err != nil {
			b.Fatal(err)
		}
	}
}
