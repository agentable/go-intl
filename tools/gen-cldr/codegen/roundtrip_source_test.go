package codegen

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/internal/localeprofile"
)

type roundTripSource struct {
	source  *cldr.Source
	profile []string
}

func loadRoundTripSource(t *testing.T) roundTripSource {
	t.Helper()

	repoRoot := filepath.Clean("../../..")
	cldrDir := filepath.Join(repoRoot, "tools", "gen-cldr", ".cldr-json", "node_modules")
	if _, err := os.Stat(cldrDir); err != nil {
		t.Skipf("pinned cldr-json checkout absent (%v); run task data:fetch", err)
	}

	versions, err := cldr.ReadVersionFile(filepath.Join(repoRoot, "internal", "cldr", "VERSION"))
	if err != nil {
		t.Fatalf("read VERSION: %v", err)
	}
	profile := readRoundTripTestProfile(t, filepath.Join(repoRoot, "tools", "locale-profile.json"))

	source, err := cldr.LoadAll(context.Background(), cldrDir, versions, profile)
	if err != nil {
		t.Fatalf("load cldr-json: %v", err)
	}
	return roundTripSource{source: source, profile: profile}
}

func readRoundTripTestProfile(t *testing.T, path string) []string {
	t.Helper()
	profile, err := localeprofile.Read(path)
	if err != nil {
		t.Fatalf("read locale profile: %v", err)
	}
	return profile.Locales
}

func resolveKernelLocale(t *testing.T, tag string) cldrlocale.Locale {
	t.Helper()
	loc, ok := cldrlocale.ResolveLocale(tag)
	if !ok {
		t.Fatalf("kernel locale %q not resolvable", tag)
	}
	return loc
}
