// Package intltest provides shared test helpers for go-intl packages.
package intltest

import (
	"encoding/json"
	"os"
	"reflect"
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

func DiffParts[T any](t testing.TB, got, want []T) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parts = %#v, want %#v", got, want)
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
