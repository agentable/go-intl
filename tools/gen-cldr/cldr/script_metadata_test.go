package cldr

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadScriptDirections(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-core", "scriptMetadata.json"), `{
		"scriptMetadata": {
			"Arab": {"rtl": "YES"},
			"Latn": {"rtl": "NO"},
			"Zinh": {},
			"Zzzz": {"rtl": "UNKNOWN"}
		}
	}`)

	got, err := loadScriptDirections(root)
	if err != nil {
		t.Fatalf("loadScriptDirections() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("loadScriptDirections() = %#v, want two known directions", got)
	}
	if rtl, ok := got["Arab"]; !ok || !rtl {
		t.Errorf("Arab = %t, %t; want true, true", rtl, ok)
	}
	if rtl, ok := got["Latn"]; !ok || rtl {
		t.Errorf("Latn = %t, %t; want false, true", rtl, ok)
	}
	for _, script := range []string{"Zinh", "Zzzz"} {
		if _, ok := got[script]; ok {
			t.Errorf("%s present, want unknown direction omitted", script)
		}
	}
}

func TestLoadScriptDirectionsRejectsInvalidSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "malformed JSON", raw: `{`, want: "parse"},
		{name: "missing metadata", raw: `{}`, want: "missing scriptMetadata"},
		{name: "invalid enum", raw: `{"scriptMetadata":{"Arab":{"rtl":"MAYBE"}}}`, want: "invalid rtl value"},
		{name: "invalid script", raw: `{"scriptMetadata":{"arab":{"rtl":"YES"}}}`, want: "invalid script code"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-core", "scriptMetadata.json"), tc.raw)
			_, err := loadScriptDirections(root)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("loadScriptDirections() error = %v, want containing %q", err, tc.want)
			}
		})
	}
}
