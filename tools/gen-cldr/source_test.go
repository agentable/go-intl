package main

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

func TestLoadAllBuildsTypedSource(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	writeListPatternCLDRFixture(t, root)
	writeRelativeTimeCLDRFixture(t, root)
	writeDisplayNamesCLDRFixture(t, root)

	versions := cldr.Versions{CLDR: "48.1.0", ICU: "78", TZData: "2025b"}
	profile := []string{"en", "en-US"}
	source, err := cldr.LoadAll(context.Background(), root, versions, profile)
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	if source.Root != root {
		t.Fatalf("LoadAll().Root = %q, want %q", source.Root, root)
	}
	if want := []string{"und", "en", "en-US"}; !slices.Equal(source.Available, want) {
		t.Fatalf("LoadAll().Available = %#v, want %#v", source.Available, want)
	}
	if source.LikelySubtags["en"] != "en_Latn_US" {
		t.Fatalf("LoadAll().LikelySubtags[en] = %q, want en_Latn_US", source.LikelySubtags["en"])
	}
	if rtl, ok := source.ScriptDirections["Latn"]; !ok || rtl {
		t.Fatalf("LoadAll().ScriptDirections[Latn] = %t, %t; want false, true", rtl, ok)
	}
	requireSourceEntry(t, "Numbers", source.Numbers, "en")
	currencies := requireSourceEntry(t, "Currencies", source.Currencies, "en")
	if currencies["USD"].Canonical != "US dollar" {
		t.Fatalf("LoadAll().Currencies[en][USD].Canonical = %q, want US dollar", currencies["USD"].Canonical)
	}
	if source.CurrencyFractions["USD"].Digits != 2 {
		t.Fatalf("LoadAll().CurrencyFractions[USD].Digits = %d, want 2", source.CurrencyFractions["USD"].Digits)
	}

	requireSourceEntry(t, "Dates", source.Dates, "en")
	if got := source.Preference.HourCycle["US"]; !slices.Equal(got, []string{"h12", "h23"}) {
		t.Fatalf("LoadAll().Preference.HourCycle[US] = %#v, want [h12 h23]", got)
	}
	if len(source.Metazones.ZoneToMetazones) == 0 {
		t.Fatal("LoadAll().Metazones.ZoneToMetazones is empty, want fixture zones")
	}
	requireSourceEntry(t, "Units", source.Units, "en")
	requireSourceEntry(t, "ListPatterns", source.ListPatterns, "en")
	requireSourceEntry(t, "RelativeTime", source.RelativeTime, "en")
	displayNames := requireSourceEntry(t, "DisplayNames", source.DisplayNames, "en")
	if displayNames.LocalePattern != "{0} ({1})" {
		t.Fatalf("LoadAll().DisplayNames[en].LocalePattern = %q, want {0} ({1})", displayNames.LocalePattern)
	}
}

func requireSourceEntry[K comparable, V any](t *testing.T, name string, values map[K]V, key K) V {
	t.Helper()

	value, ok := values[key]
	if !ok {
		t.Fatalf("LoadAll().%s missing key %#v", name, key)
	}
	return value
}

func writeDisplayNamesCLDRFixture(t *testing.T, root string) {
	t.Helper()

	locale := "en"
	base := filepath.Join(root, "cldr-localenames-full", "main", locale)
	mustWriteGenCLDRFile(t, filepath.Join(base, "languages.json"), `{
		"main": {
			"en": {
				"localeDisplayNames": {
					"languages": {
						"en": "English",
						"en-US": "American English",
						"fr": "French"
					}
				}
			}
		}
	}`)
	mustWriteGenCLDRFile(t, filepath.Join(base, "territories.json"), `{
		"main": {
			"en": {
				"localeDisplayNames": {
					"territories": {
						"US": "United States",
						"FR": "France"
					}
				}
			}
		}
	}`)
	mustWriteGenCLDRFile(t, filepath.Join(base, "scripts.json"), `{
		"main": {
			"en": {
				"localeDisplayNames": {
					"scripts": {
						"Latn": "Latin"
					}
				}
			}
		}
	}`)
	mustWriteGenCLDRFile(t, filepath.Join(base, "localeDisplayNames.json"), `{
		"main": {
			"en": {
				"localeDisplayNames": {
					"localeDisplayPattern": {
						"localePattern": "{0} ({1})"
					},
					"types": {
						"calendar": {
							"gregorian": "Gregorian Calendar"
						}
					}
				}
			}
		}
	}`)
}

func mustWriteGenCLDRFile(t *testing.T, path, data string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatalf("create fixture dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(data), 0o666); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
}
