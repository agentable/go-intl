package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckDataPinsRejectsMissingCLDRVersion(t *testing.T) {
	t.Parallel()

	fixture := newPreflightFixture(t)
	fixture.writeVersion("icu=78\ntzdata=2025b\n")

	err := checkDataPins(fixture.config)
	if err == nil || !strings.Contains(err.Error(), "internal/cldr/VERSION") {
		t.Fatalf("checkDataPins() error = %v, want missing CLDR version path", err)
	}
}

func TestCheckDataPinsRejectsMalformedCLDRVersion(t *testing.T) {
	t.Parallel()

	fixture := newPreflightFixture(t)
	fixture.writeVersion("cldr=48\nicu=78\ntzdata=2025b\n")

	err := checkDataPins(fixture.config)
	if err == nil || !strings.Contains(err.Error(), "cldr=48") {
		t.Fatalf("checkDataPins() error = %v, want malformed CLDR version", err)
	}
}

func TestCheckDataPinsRejectsMissingCLDRPackagePin(t *testing.T) {
	t.Parallel()

	fixture := newPreflightFixture(t)
	fixture.writeFile("tools/gen-cldr/.cldr-json/package.json", `{"dependencies":{"cldr-numbers-full":"48.1.0"}}`)

	err := checkDataPins(fixture.config)
	if err == nil || !strings.Contains(err.Error(), "cldr-core") || !strings.Contains(err.Error(), "package.json") {
		t.Fatalf("checkDataPins() error = %v, want missing cldr-core package pin", err)
	}
}

func TestCheckDataPinsRejectsMissingTZDataLockVersion(t *testing.T) {
	t.Parallel()

	fixture := newPreflightFixture(t)
	fixture.writeFile("tools/gen-cldr/tzdata.json", `{"url":"https://example.test/tzdata2025b.tar.gz","sha256":"11810413345fc7805017e27ea9fa4885fd74cd61b2911711ad038f5d28d71474"}`)

	err := checkDataPins(fixture.config)
	if err == nil || !strings.Contains(err.Error(), "tzdata.json") || !strings.Contains(err.Error(), "version") {
		t.Fatalf("checkDataPins() error = %v, want missing tzdata lock version", err)
	}
}

func TestCheckDataPinsRejectsInvalidTZDataHash(t *testing.T) {
	t.Parallel()

	fixture := newPreflightFixture(t)
	fixture.writeFile("tools/gen-cldr/tzdata.json", `{"version":"2025b","url":"https://example.test/tzdata2025b.tar.gz","sha256":"not-a-hash"}`)

	err := checkDataPins(fixture.config)
	if err == nil || !strings.Contains(err.Error(), "tzdata.json") || !strings.Contains(err.Error(), "sha256") {
		t.Fatalf("checkDataPins() error = %v, want invalid tzdata sha256", err)
	}
}

func TestCheckDataPinsRejectsOlderGoTransitionData(t *testing.T) {
	t.Parallel()

	fixture := newPreflightFixture(t)
	fixture.writeFile("goroot/lib/time/update.bash", "CODE=2025a\nDATA=2025a\n")

	err := checkDataPins(fixture.config)
	if err == nil || !strings.Contains(err.Error(), "update.bash") || !strings.Contains(err.Error(), "older") {
		t.Fatalf("checkDataPins() error = %v, want older Go transition data", err)
	}
}

func TestCheckDataPinsRejectsCorruptPinInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(preflightFixture)
		want   string
	}{
		{
			name: "package version drift",
			mutate: func(f preflightFixture) {
				f.writeFile("tools/gen-cldr/.cldr-json/package.json", `{"dependencies":{"cldr-core":"48.0.0"}}`)
			},
			want: "version mismatch",
		},
		{
			name: "malformed package json",
			mutate: func(f preflightFixture) {
				f.writeFile("tools/gen-cldr/.cldr-json/package.json", `{`)
			},
			want: "package.json",
		},
		{
			name: "missing tzdata lock",
			mutate: func(f preflightFixture) {
				if err := os.Remove(f.config.tzLockFile); err != nil {
					f.t.Fatal(err)
				}
			},
			want: "tzdata.json",
		},
		{
			name: "missing hash",
			mutate: func(f preflightFixture) {
				f.writeFile("tools/gen-cldr/tzdata.json", `{"version":"2025b","url":"https://example.test/tzdata2025b.tar.gz"}`)
			},
			want: "sha256",
		},
		{
			name: "missing go data version",
			mutate: func(f preflightFixture) {
				f.writeFile("goroot/lib/time/update.bash", "CODE=2025c\n")
			},
			want: "DATA version",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fixture := newPreflightFixture(t)
			tc.mutate(fixture)
			err := checkDataPins(fixture.config)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("checkDataPins() error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestCheckDataPinsAcceptsValidPinsAndNewerGoTransitionData(t *testing.T) {
	t.Parallel()

	fixture := newPreflightFixture(t)
	if err := checkDataPins(fixture.config); err != nil {
		t.Fatalf("checkDataPins() error = %v, want nil", err)
	}
}

type preflightFixture struct {
	t      *testing.T
	root   string
	config preflightConfig
}

func newPreflightFixture(t *testing.T) preflightFixture {
	t.Helper()

	root := t.TempDir()
	fixture := preflightFixture{
		t:    t,
		root: root,
		config: preflightConfig{
			versionFile:  filepath.Join(root, "internal", "cldr", "VERSION"),
			packageFile:  filepath.Join(root, "tools", "gen-cldr", ".cldr-json", "package.json"),
			tzLockFile:   filepath.Join(root, "tools", "gen-cldr", "tzdata.json"),
			goUpdateFile: filepath.Join(root, "goroot", "lib", "time", "update.bash"),
		},
	}
	fixture.writeVersion("cldr=48.1.0\nicu=78\ntzdata=2025b\n")
	fixture.writeFile("tools/gen-cldr/.cldr-json/package.json", `{"dependencies":{"cldr-core":"48.1.0","cldr-numbers-full":"48.1.0"}}`)
	fixture.writeFile("tools/gen-cldr/tzdata.json", `{"version":"2025b","url":"https://example.test/tzdata2025b.tar.gz","sha256":"11810413345fc7805017e27ea9fa4885fd74cd61b2911711ad038f5d28d71474"}`)
	fixture.writeFile("goroot/lib/time/update.bash", "CODE=2025c\nDATA=2025c\n")
	return fixture
}

func (f preflightFixture) writeVersion(contents string) {
	f.t.Helper()

	f.writeFile("internal/cldr/VERSION", contents)
}

func (f preflightFixture) writeFile(rel, contents string) {
	f.t.Helper()

	path := filepath.Join(f.root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		f.t.Fatal(err)
	}
}
