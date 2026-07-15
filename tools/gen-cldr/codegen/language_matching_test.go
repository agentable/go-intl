package codegen

import (
	"strings"
	"testing"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

func TestRenderLanguageMatchingProfile(t *testing.T) {
	t.Parallel()

	src, err := renderLanguageMatchingProfile(cldr.LanguageMatching{
		ParadigmLocales: []string{"en"},
		MatchVariables: []cldr.LanguageMatchVariable{{
			Name: "americas", SourceRegions: []string{"019"}, ExpandedRegions: []string{"BR", "US"},
		}},
		Rules: []cldr.LanguageMatchRule{{Desired: "gsw", Supported: "de", Distance: 4, OneWay: true}},
	})
	if err != nil {
		t.Fatalf("renderLanguageMatchingProfile() error = %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"package localematcher",
		`paradigmLocales: []string{`,
		`"en",`,
		`name: "americas"`,
		`sourceRegions: []string{"019"}`,
		`expandedRegions: []string{"BR", "US"}`,
		`desired: "gsw"`,
		`oneWay: true`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("renderLanguageMatchingProfile() missing %q:\n%s", want, text)
		}
	}
}
