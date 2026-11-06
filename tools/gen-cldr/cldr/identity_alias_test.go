package cldr

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadIdentityAliases(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, "cldr-core/supplemental/aliases.json")
	const valid = `{"supplemental":{"metadata":{"alias":{
 "languageAlias":{"iw":{"_replacement":"he"}},
 "scriptAlias":{"Qaai":{"_replacement":"Zinh"}},
 "territoryAlias":{"SU":{"_replacement":"RU AM"}},
 "variantAlias":{"heploc":{"_replacement":"alalc97"}},
 "subdivisionAlias":{"fi01":{"_replacement":"AX"}}
 }}}}`
	mustWriteFile(t, path, valid)
	got, err := loadIdentityAliases(root)
	if err != nil {
		t.Fatal(err)
	}
	if got["languageAlias"]["iw"] != "he" || got["territoryAlias"]["SU"] != "RU AM" || got["subdivisionAlias"]["fi01"] != "AX" {
		t.Fatalf("lost alias values/order: %v", got)
	}
	for _, input := range []string{"{", `{"supplemental":{}}`, strings.Replace(valid, `"he"`, `""`, 1)} {
		mustWriteFile(t, path, input)
		if _, err := loadIdentityAliases(root); err == nil {
			t.Fatal("accepted missing/malformed alias source")
		}
	}
}
