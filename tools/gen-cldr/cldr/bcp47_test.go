package cldr

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestLoadUnicodeTypeAliases(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-bcp47", "bcp47", "calendar.json"), `{
		"keyword": {
			"u": {
				"ca": {
					"islamic-civil": {"_description":"canonical"},
					"islamicc": {"_deprecated":true,"_alias":"islamic-civil","_preferred":"islamic-civil"}
				},
				"co": {
					"islamic-civil": {"_description":"unrelated collation"}
				}
			},
			"t": {
				"m0": {
					"prprname": {"_alias":"names"}
				}
			}
		}
	}`)
	mustWriteFile(t, filepath.Join(root, "cldr-bcp47", "bcp47", "measure.json"), `{
		"keyword": {
			"u": {
				"ms": {
					"uksystem": {"_alias":"imperial"},
					"ussystem": {}
				}
			}
		}
	}`)

	got, err := loadUnicodeTypeAliases(root)
	if err != nil {
		t.Fatalf("loadUnicodeTypeAliases() error = %v", err)
	}
	want := []UnicodeTypeAlias{
		{Key: "ca", Alias: "islamicc", Canonical: "islamic-civil"},
		{Key: "ms", Alias: "imperial", Canonical: "uksystem"},
		{Key: "t:m0", Alias: "names", Canonical: "prprname"},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("loadUnicodeTypeAliases() = %#v, want %#v", got, want)
	}
}

func TestLoadUnicodeTypeAliasesRejectsInvalidGraphs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		types string
		want  string
	}{
		{
			name:  "conflicting alias",
			types: `"alpha":{"_alias":"shared"},"bravo":{"_alias":"shared"}`,
			want:  "conflicting alias",
		},
		{
			name:  "cycle",
			types: `"alpha":{"_alias":"bravo"},"bravo":{"_alias":"alpha"}`,
			want:  "cycle",
		},
		{
			name:  "missing preferred target",
			types: `"alpha":{"_deprecated":true,"_preferred":"missing"}`,
			want:  "preferred target",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			raw := `{"keyword":{"u":{"aa":{` + tc.types + `}}}}`
			mustWriteFile(t, filepath.Join(root, "cldr-bcp47", "bcp47", "fixture.json"), raw)
			if _, err := loadUnicodeTypeAliases(root); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("loadUnicodeTypeAliases() error = %v, want containing %q", err, tc.want)
			}
		})
	}
}

func TestLoadUnicodeTypeAliasesKeyGrammar(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		key   string
		valid bool
	}{{"00", false}, {"a0", false}, {"0a", true}, {"aa", true}} {
		t.Run(tc.key, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			path := filepath.Join(root, "cldr-bcp47", "bcp47", "keys.json")
			mustWriteFile(t, path, fmt.Sprintf(`{"keyword":{"u":{%q:{"foobar":{}}}}}`, tc.key))
			_, err := loadUnicodeTypeAliases(root)
			if tc.valid {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), "invalid key") {
				t.Fatalf("loadUnicodeTypeAliases error = %v, want source and invalid key", err)
			}
		})
	}
}
