package cldr

import (
	"encoding/json"
	"maps"
	"path/filepath"
	"slices"
	"testing"
)

func TestListPatternKeysCoverTypesAndStyles(t *testing.T) {
	t.Parallel()

	want := []listPatternKey{
		{cldr: "listPattern-type-standard", typ: "conjunction", style: "long"},
		{cldr: "listPattern-type-standard-short", typ: "conjunction", style: "short"},
		{cldr: "listPattern-type-standard-narrow", typ: "conjunction", style: "narrow"},
		{cldr: "listPattern-type-or", typ: "disjunction", style: "long"},
		{cldr: "listPattern-type-or-short", typ: "disjunction", style: "short"},
		{cldr: "listPattern-type-or-narrow", typ: "disjunction", style: "narrow"},
		{cldr: "listPattern-type-unit", typ: "unit", style: "long"},
		{cldr: "listPattern-type-unit-short", typ: "unit", style: "short"},
		{cldr: "listPattern-type-unit-narrow", typ: "unit", style: "narrow"},
	}
	if !slices.Equal(listPatternKeys[:], want) {
		t.Fatalf("listPatternKeys = %#v, want %#v", listPatternKeys, want)
	}
}

func TestLoadListPatternsMapsAndInherits(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	rows := completeListPatternRows()
	rows["listPattern-type-standard-ignored"] = ListPattern{Pair: "ignored"}
	mustWriteFile(t, filepath.Join(root, "cldr-misc-full", "main", "en", "listPatterns.json"), completeListPatternsDocument(t, rows))

	got, err := loadListPatterns(root, []string{undefinedLocale, "en-US", "en", "missing"})
	if err != nil {
		t.Fatalf("loadListPatterns() error = %v", err)
	}
	want := map[string]ListPatterns{
		"en":    wantListPatterns(),
		"en-US": wantListPatterns(),
	}
	assertListPatternsByLocale(t, "loadListPatterns()", got, want)
}

func TestLoadListPatternsRejectsInvalidShape(t *testing.T) {
	t.Parallel()

	missingPatternRow := completeListPatternRows()
	delete(missingPatternRow, "listPattern-type-unit-narrow")

	tests := []struct {
		name string
		raw  string
	}{
		{name: "invalid json", raw: `{`},
		{name: "missing main", raw: `{}`},
		{
			name: "missing locale body",
			raw:  `{"main":{}}`,
		},
		{
			name: "missing listPatterns",
			raw:  `{"main":{"en":{}}}`,
		},
		{
			name: "null listPatterns",
			raw:  `{"main":{"en":{"listPatterns":null}}}`,
		},
		{
			name: "wrong listPatterns type",
			raw:  `{"main":{"en":{"listPatterns":"standard"}}}`,
		},
		{
			name: "missing pattern row",
			raw:  completeListPatternsDocument(t, missingPatternRow),
		},
		{
			name: "missing template field",
			raw: `{
					"main": {
						"en": {
							"listPatterns": {
								"listPattern-type-standard": {
									"2": "{0} and {1}",
									"start": "{0}, {1}",
									"middle": "{0}, {1}"
								}
							}
						}
					}
				}`,
		},
		{
			name: "empty template field",
			raw: `{
					"main": {
						"en": {
							"listPatterns": {
								"listPattern-type-standard": {
									"2": "{0} and {1}",
									"start": "",
									"middle": "{0}, {1}",
									"end": "{0}, and {1}"
								}
							}
						}
					}
				}`,
		},
		{
			name: "missing placeholder",
			raw: `{
					"main": {
						"en": {
							"listPatterns": {
								"listPattern-type-standard": {
									"2": "{0} and {1}",
									"start": "{0}, {1}",
									"middle": "{0}, {1}",
									"end": "{0}, and"
								}
							}
						}
					}
				}`,
		},
		{
			name: "duplicate placeholder",
			raw: `{
						"main": {
							"en": {
							"listPatterns": {
								"listPattern-type-standard": {
									"2": "{0} and {1}",
									"start": "{0}, {1}",
									"middle": "{0}, {1}",
									"end": "{0}, {0}, and {1}"
								}
							}
						}
						}
					}`,
		},
		{
			name: "malformed placeholder syntax",
			raw: `{
						"main": {
							"en": {
								"listPatterns": {
									"listPattern-type-standard": {
										"2": "{0} and {1}",
										"start": "{0}, {1}",
										"middle": "{0}, {1} {",
										"end": "{0}, and {1}"
									}
								}
							}
						}
					}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-misc-full", "main", "en", "listPatterns.json"), tc.raw)

			if _, err := loadListPatterns(root, []string{"en"}); err == nil {
				t.Fatal("loadListPatterns() succeeded, want error")
			}
		})
	}
}

func TestValidateListPatternReportsFieldsInStableOrder(t *testing.T) {
	t.Parallel()

	err := validateListPattern("listPattern-type-standard", ListPattern{
		Pair:   "{0}",
		Start:  "{0}",
		Middle: "{0}",
		End:    "{0}",
	})
	if err == nil {
		t.Fatal("validateListPattern() succeeded, want error")
	}
	if got, want := err.Error(), `listPattern-type-standard.2 invalid: expected one {0} and one {1}, got "{0}"`; got != want {
		t.Fatalf("validateListPattern() error = %q, want %q", got, want)
	}
}

func completeListPatternsDocument(t *testing.T, rows map[string]ListPattern) string {
	t.Helper()

	doc := struct {
		Main map[string]struct {
			ListPatterns map[string]ListPattern `json:"listPatterns"`
		} `json:"main"`
	}{
		Main: map[string]struct {
			ListPatterns map[string]ListPattern `json:"listPatterns"`
		}{
			"en": {ListPatterns: rows},
		},
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return string(raw)
}

func completeListPatternRows() map[string]ListPattern {
	return map[string]ListPattern{
		"listPattern-type-standard": {
			Pair:   "{0} and {1}",
			Start:  "{0}, {1}",
			Middle: "{0}, {1}",
			End:    "{0}, and {1}",
		},
		"listPattern-type-standard-short": {
			Pair:   "{0} & {1}",
			Start:  "{0}, {1}",
			Middle: "{0}, {1}",
			End:    "{0}, & {1}",
		},
		"listPattern-type-standard-narrow": {
			Pair:   "{0} + {1}",
			Start:  "{0} + {1}",
			Middle: "{0} + {1}",
			End:    "{0} + {1}",
		},
		"listPattern-type-or": {
			Pair:   "{0} or {1}",
			Start:  "{0}, {1}",
			Middle: "{0}, {1}",
			End:    "{0}, or {1}",
		},
		"listPattern-type-or-short": {
			Pair:   "{0} or {1}",
			Start:  "{0}, {1}",
			Middle: "{0}, {1}",
			End:    "{0}, or {1}",
		},
		"listPattern-type-or-narrow": {
			Pair:   "{0}/{1}",
			Start:  "{0}/{1}",
			Middle: "{0}/{1}",
			End:    "{0}/{1}",
		},
		"listPattern-type-unit": {
			Pair:   "{0}, {1}",
			Start:  "{0}, {1}",
			Middle: "{0}, {1}",
			End:    "{0}, {1}",
		},
		"listPattern-type-unit-short": {
			Pair:   "{0}, {1}",
			Start:  "{0}, {1}",
			Middle: "{0}, {1}",
			End:    "{0}, {1}",
		},
		"listPattern-type-unit-narrow": {
			Pair:   "{0} {1}",
			Start:  "{0} {1}",
			Middle: "{0} {1}",
			End:    "{0} {1}",
		},
	}
}

func wantListPatterns() ListPatterns {
	rows := completeListPatternRows()
	return ListPatterns{
		"conjunction": {
			"long":   rows["listPattern-type-standard"],
			"short":  rows["listPattern-type-standard-short"],
			"narrow": rows["listPattern-type-standard-narrow"],
		},
		"disjunction": {
			"long":   rows["listPattern-type-or"],
			"short":  rows["listPattern-type-or-short"],
			"narrow": rows["listPattern-type-or-narrow"],
		},
		"unit": {
			"long":   rows["listPattern-type-unit"],
			"short":  rows["listPattern-type-unit-short"],
			"narrow": rows["listPattern-type-unit-narrow"],
		},
	}
}

func assertListPatternsByLocale(t *testing.T, name string, got, want map[string]ListPatterns) {
	t.Helper()

	if !maps.EqualFunc(got, want, listPatternsEqual) {
		t.Fatalf("%s = %#v, want %#v", name, got, want)
	}
}

func listPatternsEqual(got, want ListPatterns) bool {
	return maps.EqualFunc(got, want, func(gotStyles, wantStyles map[string]ListPattern) bool {
		return maps.Equal(gotStyles, wantStyles)
	})
}
