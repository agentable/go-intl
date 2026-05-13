package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestRunGeneratesMetazonesAndUnits(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writePhase3CLDRFixture(t, root)
	writePhase4CLDRFixture(t, root)
	writePhase5CLDRFixture(t, root)
	writePhase6CLDRFixture(t, root)
	writePhase7CLDRFixture(t, root)
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

	for _, name := range []string{"metazones.go", "units.go"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatalf("expected generated %s: %v", name, err)
		}
	}
	metazones, err := os.ReadFile(filepath.Join(out, "metazones.go"))
	if err != nil {
		t.Fatalf("read metazones.go: %v", err)
	}
	if !containsAll(string(metazones), "type MetazonePeriod struct", "type TimeZoneFormats struct", "func TimeZoneMetazone", "func ZoneToMetazone", "func (l Locale) MetazoneName", "func (l Locale) ExemplarCity", "func (l Locale) TimeZoneFormats", "Europe/Moscow", "1293840000000", "1385856000000") {
		t.Fatalf("metazones.go missing expected generated content:\n%s", metazones)
	}
	units, err := os.ReadFile(filepath.Join(out, "units.go"))
	if err != nil {
		t.Fatalf("read units.go: %v", err)
	}
	if !containsAll(string(units), "func (l Locale) UnitPattern", "func (l Locale) CompoundUnitPattern") {
		t.Fatalf("units.go missing expected generated content:\n%s", units)
	}
	stringsData := readGeneratedStringTable(t, filepath.Join(out, "strings.go"))
	if !containsAll(stringsData, "0K", "America_Pacific", "Eastern Time", "GMT{0}", "{0}/{1}") {
		t.Fatalf("strings.go missing expected metazone/unit strings:\n%s", stringsData)
	}
}

func TestRunAcceptsSingleMetazoneObject(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writePhase3CLDRFixture(t, root)
	writePhase4CLDRFixture(t, root)
	writePhase5CLDRFixture(t, root)
	writePhase6CLDRFixture(t, root)
	writePhase7CLDRFixture(t, root)
	supp := filepath.Join(root, "cldr-core", "supplemental")
	metaZones := `{"supplemental":{"metaZones":{"metazoneInfo":{"timezone":{"America":{"Los_Angeles":{"usesMetazone":{"_mzone":"America_Pacific"}}}}}}}}`
	if err := os.WriteFile(filepath.Join(supp, "metaZones.json"), []byte(metaZones), 0o666); err != nil {
		t.Fatalf("write metaZones: %v", err)
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

func TestRunAcceptsHeterogeneousUnitFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writePhase3CLDRFixture(t, root)
	writePhase4CLDRFixture(t, root)
	writePhase5CLDRFixture(t, root)
	writePhase6CLDRFixture(t, root)
	writePhase7CLDRFixture(t, root)
	for _, loc := range []string{"en", "en-US", "zh", "zh-Hans", "zh-Hans-CN"} {
		units := `{"main":{"` + loc + `":{"units":{"durationUnit-type-hm":{"durationUnitPattern":"hh:mm"},"long":{"10p-1":{"unitPrefixPattern":"deci{0}"},"per":{"compoundUnitPattern":"{0} per {1}"},"power2":{"compoundUnitPattern1":"square {0}"},"length-meter":{"displayName":"meters","unitPattern-count-one":"{0} meter","unitPattern-count-other":"{0} meters"},"length-yard":{"unitPattern-count-other":"{0} yards"}},"short":{"per":{"compoundUnitPattern":"{0}/{1}"},"length-meter":{"unitPattern-count-one":"{0} m","unitPattern-count-other":"{0} m"},"length-yard":{"unitPattern-count-other":"{0} yd"}},"narrow":{"per":{"compoundUnitPattern":"{0}/{1}"},"length-meter":{"unitPattern-count-one":"{0}m","unitPattern-count-other":"{0}m"},"length-yard":{"unitPattern-count-other":"{0}yd"}}}}}}`
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
	if containsAll(string(generated), "yard") {
		t.Fatalf("units.go contains unsupported yard patterns:\n%s", generated)
	}
}

func writePhase7CLDRFixture(t *testing.T, root string) {
	t.Helper()
	locales := []string{"en", "en-US", "zh", "zh-Hans", "zh-Hans-CN"}
	for _, loc := range locales {
		datesDir := filepath.Join(root, "cldr-dates-full", "main", loc)
		if err := os.MkdirAll(datesDir, 0o777); err != nil {
			t.Fatalf("mkdir timezone names %s: %v", loc, err)
		}
		if err := os.WriteFile(filepath.Join(datesDir, "timeZoneNames.json"), []byte(timeZoneNamesJSON(loc)), 0o666); err != nil {
			t.Fatalf("write timeZoneNames %s: %v", loc, err)
		}
		unitsDir := filepath.Join(root, "cldr-units-full", "main", loc)
		if err := os.MkdirAll(unitsDir, 0o777); err != nil {
			t.Fatalf("mkdir units %s: %v", loc, err)
		}
		if err := os.WriteFile(filepath.Join(unitsDir, "units.json"), []byte(unitsJSON(loc)), 0o666); err != nil {
			t.Fatalf("write units %s: %v", loc, err)
		}
	}
	supp := filepath.Join(root, "cldr-core", "supplemental")
	metaZones := `{"supplemental":{"metaZones":{"metazoneInfo":{"timezone":{"America":{"Los_Angeles":[{"usesMetazone":{"_mzone":"America_Pacific"}}],"New_York":[{"usesMetazone":{"_mzone":"America_Eastern"}}]},"Europe":{"Moscow":[{"usesMetazone":{"_mzone":"Moscow","_to":"2011-01-01 00:00"}},{"usesMetazone":{"_mzone":"Europe_Further_Eastern","_from":"2011-01-01 00:00","_to":"2013-12-01 00:00"}},{"usesMetazone":{"_mzone":"Moscow","_from":"2013-12-01 00:00"}}]}}}}}}`
	if err := os.WriteFile(filepath.Join(supp, "metaZones.json"), []byte(metaZones), 0o666); err != nil {
		t.Fatalf("write metaZones: %v", err)
	}
}

func timeZoneNamesJSON(locale string) string {
	return `{"main":{"` + locale + `":{"dates":{"timeZoneNames":{"hourFormat":"+HH:mm;-HH:mm","gmtFormat":"GMT{0}","gmtZeroFormat":"GMT","metazone":{"America_Pacific":{"long":{"generic":"Pacific Time","standard":"Pacific Standard Time","daylight":"Pacific Daylight Time"},"short":{"generic":"PT","standard":"PST","daylight":"PDT"}},"America_Eastern":{"long":{"generic":"Eastern Time","standard":"Eastern Standard Time","daylight":"Eastern Daylight Time"}}},"zone":{"America":{"Los_Angeles":{"exemplarCity":"Los Angeles"},"New_York":{"exemplarCity":"New York"}},"Etc":{"UTC":{"long":{"standard":"Coordinated Universal Time"},"short":{"standard":"UTC"}}}}}}}}}`
}

func unitsJSON(locale string) string {
	return `{"main":{"` + locale + `":{"units":{"long":{"length-meter":{"unitPattern-count-one":"{0} meter","unitPattern-count-other":"{0} meters"},"per":{"compoundUnitPattern":"{0}/{1}"}},"short":{"length-meter":{"unitPattern-count-one":"{0} m","unitPattern-count-other":"{0} m"},"per":{"compoundUnitPattern":"{0}/{1}"}},"narrow":{"length-meter":{"unitPattern-count-one":"{0}m","unitPattern-count-other":"{0}m"},"per":{"compoundUnitPattern":"{0}/{1}"}}}}}}`
}
