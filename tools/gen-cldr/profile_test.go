package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocaleProfileNormalize(t *testing.T) {
	t.Parallel()

	profile := LocaleProfile{
		Locales: []string{"zh-Hans-CN", "en", "und", "en", ""},
	}
	if err := profile.normalize(); err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if got, want := strings.Join(profile.Locales, ","), "en,zh-Hans-CN"; got != want {
		t.Fatalf("Locales = %q, want %q", got, want)
	}
}

func TestLocaleProfileRejectsEmpty(t *testing.T) {
	t.Parallel()

	profile := LocaleProfile{Locales: []string{"und", ""}}
	err := profile.normalize()
	if err == nil || !strings.Contains(err.Error(), "locales is empty") {
		t.Fatalf("normalize() error = %v, want containing %q", err, "locales is empty")
	}
}

func TestReadLocaleProfileRejectsUnknownKeys(t *testing.T) {
	t.Parallel()

	path := writeProfileTestFile(t, `{"locales":["en"],"richLocales":["fr"]}`)
	_, err := readLocaleProfile(path)
	if err == nil || !strings.Contains(err.Error(), `unknown field "richLocales"`) {
		t.Fatalf("readLocaleProfile() error = %v, want unknown richLocales field", err)
	}
}

func TestReadLocaleProfileRejectsTrailingValues(t *testing.T) {
	t.Parallel()

	path := writeProfileTestFile(t, `{"locales":["en"]} {"locales":["fr"]}`)
	_, err := readLocaleProfile(path)
	if err == nil || !strings.Contains(err.Error(), "multiple JSON values") {
		t.Fatalf("readLocaleProfile() error = %v, want multiple JSON values", err)
	}
}

func writeProfileTestFile(t *testing.T, data string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "locale-profile.json")
	if err := os.WriteFile(path, []byte(data), 0o666); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	return path
}
