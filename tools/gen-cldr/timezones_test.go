package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestRunGeneratesTimeZones(t *testing.T) {
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

	metazones, err := os.ReadFile(filepath.Join(out, "metazones.go"))
	if err != nil {
		t.Fatalf("read metazones.go: %v", err)
	}
	if !containsAll(string(metazones), "type metazonePeriod struct", "type TimeZoneFormats struct", "func TimeZoneMetazone", "func ZoneToMetazone", "func (l Locale) TimeZoneName", "func (l Locale) MetazoneName", "func (l Locale) ExemplarCity", "func (l Locale) TimeZoneFormats", "Europe/Moscow", "Europe/London", "1293840000000", "1385856000000") {
		t.Fatalf("metazones.go missing expected generated content:\n%s", metazones)
	}
	stringsData := readGeneratedStringTable(t, filepath.Join(out, "strings.go"))
	if !containsAll(stringsData, "America_Pacific", "Eastern Time", "British Summer Time", "GMT{0}") {
		t.Fatalf("strings.go missing expected time-zone strings:\n%s", stringsData)
	}
}

func TestRunAcceptsSingleMetazoneObject(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
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

func writeTimeZoneCLDRFixture(t *testing.T, root string) {
	t.Helper()
	locales := []string{"en", "en-US", "zh", "zh-Hans", "zh-Hans-CN"}
	for _, loc := range locales {
		dir := filepath.Join(root, "cldr-dates-full", "main", loc)
		if err := os.MkdirAll(dir, 0o777); err != nil {
			t.Fatalf("mkdir timezone names %s: %v", loc, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "timeZoneNames.json"), []byte(timeZoneNamesJSON(loc)), 0o666); err != nil {
			t.Fatalf("write timeZoneNames %s: %v", loc, err)
		}
	}
	supp := filepath.Join(root, "cldr-core", "supplemental")
	metaZones := `{"supplemental":{"metaZones":{"metazoneInfo":{"timezone":{"America":{"Los_Angeles":[{"usesMetazone":{"_mzone":"America_Pacific"}}],"New_York":[{"usesMetazone":{"_mzone":"America_Eastern"}}]},"Europe":{"London":[{"usesMetazone":{"_mzone":"GMT"}}],"Moscow":[{"usesMetazone":{"_mzone":"Moscow","_to":"2011-01-01 00:00"}},{"usesMetazone":{"_mzone":"Europe_Further_Eastern","_from":"2011-01-01 00:00","_to":"2013-12-01 00:00"}},{"usesMetazone":{"_mzone":"Moscow","_from":"2013-12-01 00:00"}}]}}}}}}`
	if err := os.WriteFile(filepath.Join(supp, "metaZones.json"), []byte(metaZones), 0o666); err != nil {
		t.Fatalf("write metaZones: %v", err)
	}
}

func timeZoneNamesJSON(locale string) string {
	return `{"main":{"` + locale + `":{"dates":{"timeZoneNames":{"hourFormat":"+HH:mm;-HH:mm","gmtFormat":"GMT{0}","gmtZeroFormat":"GMT","metazone":{"America_Pacific":{"long":{"generic":"Pacific Time","standard":"Pacific Standard Time","daylight":"Pacific Daylight Time"},"short":{"generic":"PT","standard":"PST","daylight":"PDT"}},"America_Eastern":{"long":{"generic":"Eastern Time","standard":"Eastern Standard Time","daylight":"Eastern Daylight Time"}}},"zone":{"America":{"Los_Angeles":{"exemplarCity":"Los Angeles"},"New_York":{"exemplarCity":"New York"}},"Europe":{"London":{"long":{"daylight":"British Summer Time"},"short":{"daylight":"BST"}}},"Etc":{"UTC":{"long":{"standard":"Coordinated Universal Time"},"short":{"standard":"UTC"}}}}}}}}}`
}
