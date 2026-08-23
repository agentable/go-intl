// Package intltest provides shared test helpers for go-intl packages.
package intltest

import (
	"encoding/json/v2"
	"os"
	"testing"
	"time"

	"github.com/agentable/go-intl/locale"
)

func Locale(t testing.TB, tag string) locale.Locale {
	t.Helper()

	loc, err := locale.Parse(tag)
	if err != nil {
		t.Fatalf("locale.Parse(%q) error = %v", tag, err)
	}
	return loc
}

func LocaleList(t testing.TB, tags ...string) locale.List {
	t.Helper()

	locales, err := locale.ParseList(tags...)
	if err != nil {
		t.Fatalf("locale.ParseList(%v) error = %v", tags, err)
	}
	return locales
}

func LocaleListJSON(t testing.TB, data []byte) locale.List {
	t.Helper()

	var tags []string
	if err := json.Unmarshal(data, &tags); err != nil {
		t.Fatalf("decode locale list fixture input: %v", err)
	}
	return LocaleList(t, tags...)
}

func ReadFixture(t testing.TB, path string, v any) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("decode fixture %s: %v", path, err)
	}
}

func MustParseTime(t testing.TB, layout, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(layout, value)
	if err != nil {
		t.Fatalf("time.Parse(%q, %q) error = %v", layout, value, err)
	}
	return parsed
}
