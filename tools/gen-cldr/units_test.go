package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestRunGeneratesUnits(t *testing.T) {
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

	units, err := os.ReadFile(filepath.Join(out, "units.go"))
	if err != nil {
		t.Fatalf("read units.go: %v", err)
	}
	if !containsAll(string(units), "func (l Locale) UnitPattern", "func (l Locale) CompoundUnitPattern") {
		t.Fatalf("units.go missing expected generated content:\n%s", units)
	}
	if containsAll(string(units), "map[string]map[string]map[string]string") {
		t.Fatalf("units.go still emits nested unit pattern maps:\n%s", units)
	}
	supported, err := os.ReadFile(filepath.Join(out, "supported.go"))
	if err != nil {
		t.Fatalf("read supported.go: %v", err)
	}
	if !containsAll(string(supported), "func UnitSupportedLocales") {
		t.Fatalf("supported.go missing unit supported locales:\n%s", supported)
	}
	stringsData := readGeneratedStringTable(t, filepath.Join(out, "strings.go"))
	if !containsAll(stringsData, "{0}/{1}", "{0} meter", "{0} meters", "{0} hour", "{0} hr", "{0}ms") {
		t.Fatalf("strings.go missing expected unit strings:\n%s", stringsData)
	}
}

func TestRunAcceptsHeterogeneousUnitFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	for _, loc := range []string{"en", "en-US", "zh", "zh-Hans", "zh-Hans-CN"} {
		units := `{"main":{"` + loc + `":{"units":{"durationUnit-type-hm":{"durationUnitPattern":"hh:mm"},"long":{"10p-1":{"unitPrefixPattern":"deci{0}"},"per":{"compoundUnitPattern":"{0} per {1}"},"power2":{"compoundUnitPattern1":"square {0}"},"length-meter":{"displayName":"meters","unitPattern-count-one":"{0} meter","unitPattern-count-other":"{0} meters"},"length-furlong":{"unitPattern-count-other":"{0} yards"}},"short":{"per":{"compoundUnitPattern":"{0}/{1}"},"length-meter":{"unitPattern-count-one":"{0} m","unitPattern-count-other":"{0} m"},"length-furlong":{"unitPattern-count-other":"{0} yd"}},"narrow":{"per":{"compoundUnitPattern":"{0}/{1}"},"length-meter":{"unitPattern-count-one":"{0}m","unitPattern-count-other":"{0}m"},"length-furlong":{"unitPattern-count-other":"{0}yd"}}}}}}`
		path := filepath.Join(root, "cldr-units-full", "main", loc, "units.json")
		if err := os.WriteFile(path, []byte(units), 0o666); err != nil {
			t.Fatalf("write units %s: %v", loc, err)
		}
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
	generated, err := os.ReadFile(filepath.Join(out, "units.go"))
	if err != nil {
		t.Fatalf("read units.go: %v", err)
	}
	if containsAll(string(generated), "furlong") {
		t.Fatalf("units.go contains unsupported furlong patterns:\n%s", generated)
	}
}

func writeUnitCLDRFixture(t *testing.T, root string) {
	t.Helper()
	locales := []string{"en", "en-US", "zh", "zh-Hans", "zh-Hans-CN"}
	for _, loc := range locales {
		dir := filepath.Join(root, "cldr-units-full", "main", loc)
		if err := os.MkdirAll(dir, 0o777); err != nil {
			t.Fatalf("mkdir units %s: %v", loc, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "units.json"), []byte(unitsJSON(loc)), 0o666); err != nil {
			t.Fatalf("write units %s: %v", loc, err)
		}
	}
}

func unitsJSON(locale string) string {
	return `{"main":{"` + locale + `":{"units":{"long":{"length-meter":{"unitPattern-count-one":"{0} meter","unitPattern-count-other":"{0} meters"},"duration-hour":{"unitPattern-count-one":"{0} hour","unitPattern-count-other":"{0} hours"},"duration-millisecond":{"unitPattern-count-one":"{0} millisecond","unitPattern-count-other":"{0} milliseconds"},"per":{"compoundUnitPattern":"{0}/{1}"}},"short":{"length-meter":{"unitPattern-count-one":"{0} m","unitPattern-count-other":"{0} m"},"duration-hour":{"unitPattern-count-one":"{0} hr","unitPattern-count-other":"{0} hr"},"duration-millisecond":{"unitPattern-count-one":"{0} ms","unitPattern-count-other":"{0} ms"},"per":{"compoundUnitPattern":"{0}/{1}"}},"narrow":{"length-meter":{"unitPattern-count-one":"{0}m","unitPattern-count-other":"{0}m"},"duration-hour":{"unitPattern-count-one":"{0}h","unitPattern-count-other":"{0}h"},"duration-millisecond":{"unitPattern-count-one":"{0}ms","unitPattern-count-other":"{0}ms"},"per":{"compoundUnitPattern":"{0}/{1}"}}}}}}`
}
