package codegen

import (
	"strings"
	"testing"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

func TestRenderUnicodeTypeAliases(t *testing.T) {
	t.Parallel()

	src, err := renderUnicodeTypeAliases([]cldr.UnicodeTypeAlias{
		{Key: "ca", Alias: "islamicc", Canonical: "islamic-civil"},
		{Key: "ms", Alias: "imperial", Canonical: "uksystem"},
	})
	if err != nil {
		t.Fatalf("renderUnicodeTypeAliases() error = %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"package localeid",
		`{key: "ca", alias: "islamicc", canonical: "islamic-civil"}`,
		`{key: "ms", alias: "imperial", canonical: "uksystem"}`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("renderUnicodeTypeAliases() missing %q:\n%s", want, text)
		}
	}
}
