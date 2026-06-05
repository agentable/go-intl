package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunGeneratesUnitDomain asserts the generator emits the unit domain as a
// self-contained const-only payload at internal/cldr/unit/data.go, and no longer
// emits the retired root units.go or the root UnitSupportedLocales accessor.
func TestRunGeneratesUnitDomain(t *testing.T) {
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

	// The retired root literal renderer must not produce units.go anymore.
	if _, err := os.Stat(filepath.Join(out, "units.go")); !os.IsNotExist(err) {
		t.Fatalf("root units.go should no longer be generated, stat err = %v", err)
	}

	// The unit domain payload is a const-only data.go carrying the blobs and the
	// private _data table the decoder reads.
	payload, err := os.ReadFile(filepath.Join(out, "unit", "data.go"))
	if err != nil {
		t.Fatalf("read unit/data.go: %v", err)
	}
	if !containsAll(string(payload), "package unit", "const _data", "_unitPatternBlob", "_compoundUnitBlob", "_unitNameBlob", "_unitSupportedBlob") {
		t.Fatalf("unit/data.go missing expected const payload:\n%s", payload)
	}
	if !containsAll(string(payload), "{0} meter", "{0} meters", "{0} hour", "{0} hr", "{0}ms") {
		t.Fatalf("unit/data.go _data missing expected unit strings:\n%s", payload)
	}

	// The retired root supported.go literal file must no longer be generated; the
	// unit domain owns its own narrow supported index, and currency / numbering /
	// timezone supported values are owned by their domains.
	if _, err := os.Stat(filepath.Join(out, "supported.go")); !os.IsNotExist(err) {
		t.Fatalf("root supported.go should no longer be generated, stat err = %v", err)
	}

	// The number domain owns the numbering-system narrow index now.
	numberPayload, err := os.ReadFile(filepath.Join(out, "number", "data.go"))
	if err != nil {
		t.Fatalf("read number/data.go: %v", err)
	}
	if !containsAll(string(numberPayload), "package number", "_numberingSystemBlob") {
		t.Fatalf("number/data.go missing numbering-system narrow blob:\n%s", numberPayload)
	}
}

// TestRunDropsUnsupportedUnitsFromDomain asserts units outside the sanctioned
// set (here: furlong) never reach the generated unit domain payload.
func TestRunDropsUnsupportedUnitsFromDomain(t *testing.T) {
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
	payload, err := os.ReadFile(filepath.Join(out, "unit", "data.go"))
	if err != nil {
		t.Fatalf("read unit/data.go: %v", err)
	}
	if strings.Contains(string(payload), "yards") || strings.Contains(string(payload), "{0}yd") {
		t.Fatalf("unit/data.go contains unsupported furlong patterns:\n%s", payload)
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
