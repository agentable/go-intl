package cldr

import (
	"maps"
	"path/filepath"
	"slices"
	"testing"
)

func TestLoadRelativeTimeFieldsMapsAndInherits(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-dates-full", "main", "en", "dateFields.json"), `{
		"main": {
			"en": {
				"dates": {
					"fields": {
						"second": {
							"relativeTime-type-future": {
								"relativeTimePattern-count-one": "in {0} second",
								"relativeTimePattern-count-other": "in {0} seconds",
								"ignored-count-other": "ignored"
							},
							"relativeTime-type-past": {
								"relativeTimePattern-count-one": "{0} second ago",
								"relativeTimePattern-count-other": "{0} seconds ago"
							},
							"relative-type-0": "now",
							"ignored": "ignored"
						},
						"minute-short": {
							"relativeTime-type-future": {
								"relativeTimePattern-count-other": "in {0} min."
							},
							"relativeTime-type-past": {
								"relativeTimePattern-count-other": "{0} min. ago"
							}
						},
						"year-narrow": {
							"relative-type-0": "this yr."
						},
						"era": {
							"relative-type-0": "ignored"
						}
					}
				}
			}
		}
	}`)

	got, err := loadRelativeTimeFields(root, []string{undefinedLocale, "en-US", "en", "missing"})
	if err != nil {
		t.Fatalf("loadRelativeTimeFields() error = %v", err)
	}
	want := map[string]RelativeTimeFields{
		"en":    wantRelativeTimeFields(),
		"en-US": wantRelativeTimeFields(),
	}
	assertRelativeTimeFieldsByLocale(t, "loadRelativeTimeFields()", got, want)
}

func TestLoadRelativeTimeFieldsRejectsInvalidShape(t *testing.T) {
	t.Parallel()

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
			name: "missing dates",
			raw:  `{"main":{"en":{}}}`,
		},
		{
			name: "missing fields",
			raw:  `{"main":{"en":{"dates":{}}}}`,
		},
		{
			name: "null fields",
			raw:  `{"main":{"en":{"dates":{"fields":null}}}}`,
		},
		{
			name: "wrong fields type",
			raw:  `{"main":{"en":{"dates":{"fields":"second"}}}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-dates-full", "main", "en", "dateFields.json"), tc.raw)

			if _, err := loadRelativeTimeFields(root, []string{"en"}); err == nil {
				t.Fatal("loadRelativeTimeFields() succeeded, want error")
			}
		})
	}
}

func TestLoadRelativeTimeFieldsRejectsInvalidPatternMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "wrong pattern map type",
			raw: `{
				"main": {
					"en": {
						"dates": {
							"fields": {
								"second": {
									"relativeTime-type-future": "bad"
								}
							}
						}
					}
				}
			}`,
		},
		{
			name: "empty plural pattern",
			raw: `{
				"main": {
					"en": {
						"dates": {
							"fields": {
								"second": {
									"relativeTime-type-future": {
										"relativeTimePattern-count-other": ""
									}
								}
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
			mustWriteFile(t, filepath.Join(root, "cldr-dates-full", "main", "en", "dateFields.json"), tc.raw)

			if _, err := loadRelativeTimeFields(root, []string{"en"}); err == nil {
				t.Fatal("loadRelativeTimeFields() succeeded, want error")
			}
		})
	}
}

func TestLoadRelativeTimeFieldsRejectsInvalidPluralCategory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-dates-full", "main", "en", "dateFields.json"), `{
		"main": {
			"en": {
				"dates": {
					"fields": {
						"second": {
							"relativeTime-type-future": {
								"relativeTimePattern-count-invalid": "in {0} seconds"
							}
						}
					}
				}
			}
		}
	}`)

	if _, err := loadRelativeTimeFields(root, []string{"en"}); err == nil {
		t.Fatal("loadRelativeTimeFields() succeeded, want error")
	}
}

func TestLoadRelativeTimeFieldsRejectsInvalidRelativeLiteralKey(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-dates-full", "main", "en", "dateFields.json"), `{
		"main": {
			"en": {
				"dates": {
					"fields": {
						"second": {
							"relative-type-today": "now"
						}
					}
				}
			}
		}
	}`)

	if _, err := loadRelativeTimeFields(root, []string{"en"}); err == nil {
		t.Fatal("loadRelativeTimeFields() succeeded, want error")
	}
}

func TestRelativeTimeFieldKeysCoverUnitsAndStyles(t *testing.T) {
	t.Parallel()

	want := []relativeTimeFieldKey{
		{cldr: "second", unit: "second", style: "long"},
		{cldr: "second-short", unit: "second", style: "short"},
		{cldr: "second-narrow", unit: "second", style: "narrow"},
		{cldr: "minute", unit: "minute", style: "long"},
		{cldr: "minute-short", unit: "minute", style: "short"},
		{cldr: "minute-narrow", unit: "minute", style: "narrow"},
		{cldr: "hour", unit: "hour", style: "long"},
		{cldr: "hour-short", unit: "hour", style: "short"},
		{cldr: "hour-narrow", unit: "hour", style: "narrow"},
		{cldr: "day", unit: "day", style: "long"},
		{cldr: "day-short", unit: "day", style: "short"},
		{cldr: "day-narrow", unit: "day", style: "narrow"},
		{cldr: "week", unit: "week", style: "long"},
		{cldr: "week-short", unit: "week", style: "short"},
		{cldr: "week-narrow", unit: "week", style: "narrow"},
		{cldr: "month", unit: "month", style: "long"},
		{cldr: "month-short", unit: "month", style: "short"},
		{cldr: "month-narrow", unit: "month", style: "narrow"},
		{cldr: "quarter", unit: "quarter", style: "long"},
		{cldr: "quarter-short", unit: "quarter", style: "short"},
		{cldr: "quarter-narrow", unit: "quarter", style: "narrow"},
		{cldr: "year", unit: "year", style: "long"},
		{cldr: "year-short", unit: "year", style: "short"},
		{cldr: "year-narrow", unit: "year", style: "narrow"},
	}
	if !slices.Equal(relativeTimeFieldKeys[:], want) {
		t.Fatalf("relativeTimeFieldKeys = %#v, want %#v", relativeTimeFieldKeys, want)
	}
}

func wantRelativeTimeFields() RelativeTimeFields {
	return RelativeTimeFields{
		"minute": {
			"short": {
				Future: map[string]string{"other": "in {0} min."},
				Past:   map[string]string{"other": "{0} min. ago"},
			},
		},
		"second": {
			"long": {
				Future: map[string]string{
					"one":   "in {0} second",
					"other": "in {0} seconds",
				},
				Past: map[string]string{
					"one":   "{0} second ago",
					"other": "{0} seconds ago",
				},
				Relative: map[string]string{"0": "now"},
			},
		},
		"year": {
			"narrow": {
				Relative: map[string]string{"0": "this yr."},
			},
		},
	}
}

func assertRelativeTimeFieldsByLocale(t *testing.T, name string, got, want map[string]RelativeTimeFields) {
	t.Helper()

	if !maps.EqualFunc(got, want, relativeTimeFieldsEqual) {
		t.Fatalf("%s = %#v, want %#v", name, got, want)
	}
}

func relativeTimeFieldsEqual(got, want RelativeTimeFields) bool {
	return maps.EqualFunc(got, want, func(gotStyles, wantStyles map[string]RelativeTimeField) bool {
		return maps.EqualFunc(gotStyles, wantStyles, relativeTimeFieldEqual)
	})
}

func relativeTimeFieldEqual(got, want RelativeTimeField) bool {
	return maps.Equal(got.Future, want.Future) &&
		maps.Equal(got.Past, want.Past) &&
		maps.Equal(got.Relative, want.Relative)
}
