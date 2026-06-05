package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

// TestRunRetiresRootLocaleMatching asserts the generator no longer emits the
// retired root locale_matching.go / regions.go literal files. The CLDR
// language-matching distance data has no production consumer (the runtime
// matcher in internal/localematcher carries its own distance table), so it
// retired with the root internal/cldr package. The generator still parses the
// languageMatching / territoryContainment fixtures without error; it just emits
// nothing for them.
func TestRunRetiresRootLocaleMatching(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o777); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	versionPath := filepath.Join(out, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := Run(context.Background(), Config{CLDRDir: root, OutDir: out, VersionFile: versionPath, ProfileFile: writeLocaleProfileFixture(t, dir)}, log); err != nil {
		t.Fatalf("Run: %v", err)
	}

	for _, name := range []string{"locale_matching.go", "regions.go"} {
		if _, err := os.Stat(filepath.Join(out, name)); !os.IsNotExist(err) {
			t.Fatalf("root %s should no longer be generated, stat err = %v", name, err)
		}
	}
}

func TestRunAcceptsTerritoryContainmentArray(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	supp := filepath.Join(root, "cldr-core", "supplemental")
	territory := `{"supplemental":{"territoryContainment":{"001":{"_contains":["US","GB","CN"]},"019":{"_contains":["US"]}}}}`
	if err := os.WriteFile(filepath.Join(supp, "territoryContainment.json"), []byte(territory), 0o666); err != nil {
		t.Fatalf("write territoryContainment: %v", err)
	}
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o777); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	versionPath := filepath.Join(out, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := Run(context.Background(), Config{CLDRDir: root, OutDir: out, VersionFile: versionPath, ProfileFile: writeLocaleProfileFixture(t, dir)}, log); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestRunAcceptsBooleanLanguageMatchOneway(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	supp := filepath.Join(root, "cldr-core", "supplemental")
	languageMatching := `{"supplemental":{"languageMatching":{"written_new":{"paradigmLocales":{"_locales":"en en-GB zh-Hans"},"matchVariable":[{"_id":"$enUS","_value":"US"}],"languageMatch":[{"_desired":"en","_supported":"en","_distance":"0","_oneway":true}]}}}}`
	if err := os.WriteFile(filepath.Join(supp, "languageMatching.json"), []byte(languageMatching), 0o666); err != nil {
		t.Fatalf("write languageMatching: %v", err)
	}
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o777); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	versionPath := filepath.Join(out, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := Run(context.Background(), Config{CLDRDir: root, OutDir: out, VersionFile: versionPath, ProfileFile: writeLocaleProfileFixture(t, dir)}, log); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestRunAcceptsNumericLanguageMatchDistance(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	supp := filepath.Join(root, "cldr-core", "supplemental")
	languageMatching := `{"supplemental":{"languageMatching":{"written_new":{"paradigmLocales":{"_locales":"en en-GB zh-Hans"},"matchVariable":[{"_id":"$enUS","_value":"US"}],"languageMatch":[{"_desired":"en","_supported":"en","_distance":0}]}}}}`
	if err := os.WriteFile(filepath.Join(supp, "languageMatching.json"), []byte(languageMatching), 0o666); err != nil {
		t.Fatalf("write languageMatching: %v", err)
	}
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o777); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	versionPath := filepath.Join(out, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := Run(context.Background(), Config{CLDRDir: root, OutDir: out, VersionFile: versionPath, ProfileFile: writeLocaleProfileFixture(t, dir)}, log); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestRunAcceptsParadigmLocalesArray(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	supp := filepath.Join(root, "cldr-core", "supplemental")
	languageMatching := `{"supplemental":{"languageMatching":{"written_new":{"paradigmLocales":{"_locales":["en","en-GB","zh-Hans"]},"matchVariable":[{"_id":"$enUS","_value":"US"}],"languageMatch":[{"_desired":"en","_supported":"en","_distance":"0"}]}}}}`
	if err := os.WriteFile(filepath.Join(supp, "languageMatching.json"), []byte(languageMatching), 0o666); err != nil {
		t.Fatalf("write languageMatching: %v", err)
	}
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o777); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	versionPath := filepath.Join(out, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := Run(context.Background(), Config{CLDRDir: root, OutDir: out, VersionFile: versionPath, ProfileFile: writeLocaleProfileFixture(t, dir)}, log); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func writeMatchingCLDRFixture(t *testing.T, root string) {
	t.Helper()
	supp := filepath.Join(root, "cldr-core", "supplemental")
	languageMatching := `{"supplemental":{"languageMatching":{"written_new":{"paradigmLocales":{"_locales":"en en-GB zh-Hans"},"matchVariable":[{"_id":"$enUS","_value":"US"}],"languageMatch":[{"_desired":"en","_supported":"en","_distance":"0"},{"_desired":"en","_supported":"en-GB","_distance":"3"},{"_desired":"zh","_supported":"zh-Hans","_distance":"5"}]}}}}`
	if err := os.WriteFile(filepath.Join(supp, "languageMatching.json"), []byte(languageMatching), 0o666); err != nil {
		t.Fatalf("write languageMatching: %v", err)
	}
	territory := `{"supplemental":{"territoryContainment":{"001":{"_contains":"US GB CN"},"019":{"_contains":"US"}}}}`
	if err := os.WriteFile(filepath.Join(supp, "territoryContainment.json"), []byte(territory), 0o666); err != nil {
		t.Fatalf("write territoryContainment: %v", err)
	}
}
