package localeprofile

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProfileNormalize(t *testing.T) {
	t.Parallel()

	profile := Profile{
		Locales: []string{"zh-Hans-CN", "en", "und", "en", ""},
	}
	if err := profile.Normalize(); err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if got, want := strings.Join(profile.Locales, ","), "en,zh-Hans-CN"; got != want {
		t.Fatalf("Locales = %q, want %q", got, want)
	}
}

func TestProfileRejectsEmpty(t *testing.T) {
	t.Parallel()

	profile := Profile{Locales: []string{"und", ""}}
	err := profile.Normalize()
	if err == nil || !strings.Contains(err.Error(), "locales is empty") {
		t.Fatalf("Normalize() error = %v, want containing %q", err, "locales is empty")
	}
}

func TestReadRejectsUnknownKeys(t *testing.T) {
	t.Parallel()

	path := writeProfileTestFile(t, `{"locales":["en"],"pluralLocales":["fr"]}`)
	_, err := Read(path)
	if err == nil || !strings.Contains(err.Error(), `unknown field "pluralLocales"`) {
		t.Fatalf("Read() error = %v, want unknown pluralLocales field", err)
	}
	if !errors.Is(err, json.ErrUnknownName) {
		t.Fatalf("Read() error = %v, want a wrapped json.ErrUnknownName", err)
	}
	semantic, ok := errors.AsType[*json.SemanticError](err)
	if !ok {
		t.Fatalf("Read() error = %v, want a wrapped *json.SemanticError", err)
	}
	if semantic.JSONPointer != "/pluralLocales" {
		t.Errorf("Read() error JSON pointer = %q, want /pluralLocales", semantic.JSONPointer)
	}
}

func TestReadRejectsTrailingValues(t *testing.T) {
	t.Parallel()

	path := writeProfileTestFile(t, `{"locales":["en"]} {"locales":["fr"]}`)
	_, err := Read(path)
	if err == nil || !strings.Contains(err.Error(), "multiple JSON values") {
		t.Fatalf("Read() error = %v, want multiple JSON values", err)
	}
	if _, ok := errors.AsType[*jsontext.SyntacticError](err); !ok {
		t.Fatalf("Read() error = %v, want a wrapped *jsontext.SyntacticError", err)
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
