package cldr

import (
	"maps"
	"path/filepath"
	"slices"
	"testing"
)

func TestUnitLoaderWidthsCoverCLDRInputOrder(t *testing.T) {
	t.Parallel()

	want := []string{"long", "short", "narrow"}
	if !slices.Equal(unitLoaderWidths[:], want) {
		t.Fatalf("unitLoaderWidths = %v, want %v", unitLoaderWidths, want)
	}
}

func TestLoadUnitsExtractsPatternsAndCompound(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-units-full", "main", "en", "units.json"), `{
		"main": {
			"en": {
				"units": {
					"long": {
						"per": {
							"compoundUnitPattern": "{0} per {1}"
						},
							"length-meter": {
								"unitPattern-count-one": "{0} meter",
								"unitPattern-count-other": "{0} meters",
								"displayName": "meters"
							},
							"temperature-celsius": {
								"unitPattern-count-other": "{0} degrees Celsius"
							},
							"power2-meter": {
								"unitPattern-count-other": "{0} square meters"
							},
							"unknown-meter": {
								"unitPattern-count-other": "{0} unknown meters"
							},
							"invalid": {
								"unitPattern-count-other": "ignored"
							}
					},
					"short": {
						"per": {
							"compoundUnitPattern": "{0}/{1}"
						},
						"length-meter": {
							"unitPattern-count-other": "{0} m"
						}
					},
					"narrow": {
						"per": {
							"compoundUnitPattern": "{0}/{1}"
						},
						"length-meter": {
							"unitPattern-count-other": "{0}m"
						}
					}
				}
			}
		}
	}`)

	got, err := loadUnits(root, []string{undefinedLocale, "en", "missing"})
	if err != nil {
		t.Fatalf("loadUnits() error = %v", err)
	}
	want := map[string]Units{
		"en": {
			"celsius": {
				Patterns: map[string]map[string]map[string]string{
					"long": {
						"celsius": {"other": "{0} degrees Celsius"},
					},
				},
				Compound: wantUnitCompound(),
			},
			"meter": {
				Patterns: map[string]map[string]map[string]string{
					"long": {
						"meter": {"one": "{0} meter", "other": "{0} meters"},
					},
					"short": {
						"meter": {"other": "{0} m"},
					},
					"narrow": {
						"meter": {"other": "{0}m"},
					},
				},
				Compound: wantUnitCompound(),
			},
		},
	}
	assertUnitsByLocale(t, "loadUnits()", got, want)
}

func TestLoadUnitsRejectsInvalidPluralCategory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-units-full", "main", "en", "units.json"), `{
		"main": {
						"en": {
							"units": {
								"long": {
									"per": {
										"compoundUnitPattern": "{0} per {1}"
									},
									"length-meter": {
										"unitPattern-count-invalid": "bad"
									}
								}
				}
			}
		}
	}`)

	if _, err := loadUnits(root, []string{"en"}); err == nil {
		t.Fatal("loadUnits() succeeded, want error")
	}
}

func TestLoadUnitsRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-units-full", "main", "en", "units.json"), `{`)

	if _, err := loadUnits(root, []string{"en"}); err == nil {
		t.Fatal("loadUnits() succeeded, want error")
	}
}

func TestLoadUnitsRejectsInvalidUnitDataShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing main",
			body: `{}`,
		},
		{
			name: "missing locale body",
			body: `{
				"main": {
					"fr": {
						"units": {
							"long": {},
							"short": {},
							"narrow": {}
						}
					}
				}
			}`,
		},
		{
			name: "missing units",
			body: `{
				"main": {
					"en": {}
				}
			}`,
		},
		{
			name: "null units",
			body: `{
				"main": {
					"en": {
						"units": null
					}
				}
			}`,
		},
		{
			name: "empty units",
			body: `{
				"main": {
					"en": {
						"units": {}
					}
				}
			}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-units-full", "main", "en", "units.json"), tc.body)

			if _, err := loadUnits(root, []string{"en"}); err == nil {
				t.Fatal("loadUnits() succeeded, want error")
			}
		})
	}
}

func TestLoadUnitsRejectsMissingWidth(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-units-full", "main", "en", "units.json"), `{
		"main": {
			"en": {
					"units": {
						"long": {
							"per": {
								"compoundUnitPattern": "{0} per {1}"
							},
							"length-meter": {
								"unitPattern-count-other": "{0} meters"
							}
						},
						"short": {
							"per": {
								"compoundUnitPattern": "{0}/{1}"
							},
							"length-meter": {
								"unitPattern-count-other": "{0} m"
							}
						}
					}
				}
			}
	}`)

	if _, err := loadUnits(root, []string{"en"}); err == nil {
		t.Fatal("loadUnits() succeeded, want error")
	}
}

func TestLoadUnitsRejectsInvalidWidthShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width string
	}{
		{
			name:  "string",
			width: `"bad"`,
		},
		{
			name:  "null",
			width: `null`,
		},
		{
			name:  "empty map",
			width: `{}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-units-full", "main", "en", "units.json"), `{
				"main": {
					"en": {
						"units": {
							"long": `+tc.width+`
						}
					}
				}
			}`)

			if _, err := loadUnits(root, []string{"en"}); err == nil {
				t.Fatal("loadUnits() succeeded, want error")
			}
		})
	}
}

func TestLoadUnitsRejectsMissingCompoundPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		long string
	}{
		{
			name: "missing per",
			long: `{
				"length-meter": {
					"unitPattern-count-other": "{0} meters"
				}
			}`,
		},
		{
			name: "null per",
			long: `{
				"per": null,
				"length-meter": {
					"unitPattern-count-other": "{0} meters"
				}
			}`,
		},
		{
			name: "missing compoundUnitPattern",
			long: `{
				"per": {},
				"length-meter": {
					"unitPattern-count-other": "{0} meters"
				}
			}`,
		},
		{
			name: "empty compoundUnitPattern",
			long: `{
				"per": {
					"compoundUnitPattern": ""
				},
				"length-meter": {
					"unitPattern-count-other": "{0} meters"
				}
			}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-units-full", "main", "en", "units.json"), `{
				"main": {
					"en": {
						"units": {
							"long": `+tc.long+`
						}
					}
				}
			}`)

			if _, err := loadUnits(root, []string{"en"}); err == nil {
				t.Fatal("loadUnits() succeeded, want error")
			}
		})
	}
}

func TestLoadUnitsRejectsMissingUnitPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		unit string
	}{
		{
			name: "null unit",
			unit: `null`,
		},
		{
			name: "empty unit",
			unit: `{}`,
		},
		{
			name: "only display name",
			unit: `{
				"displayName": "meters"
			}`,
		},
		{
			name: "empty plural pattern",
			unit: `{
				"unitPattern-count-other": ""
			}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-units-full", "main", "en", "units.json"), `{
				"main": {
					"en": {
						"units": {
							"long": {
								"per": {
									"compoundUnitPattern": "{0} per {1}"
								},
								"length-meter": `+tc.unit+`
							}
						}
					}
				}
			}`)

			if _, err := loadUnits(root, []string{"en"}); err == nil {
				t.Fatal("loadUnits() succeeded, want error")
			}
		})
	}
}

func TestUnitIdentifierFromCLDRKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
		want string
		ok   bool
	}{
		{name: "simple", key: "length-meter", want: "meter", ok: true},
		{name: "hyphenated unit", key: "length-mile-scandinavian", want: "mile-scandinavian", ok: true},
		{name: "missing namespace", key: "meter", ok: false},
		{name: "empty unit", key: "length-", ok: false},
		{name: "unknown namespace", key: "unknown-meter", ok: false},
		{name: "compound namespace", key: "power2-meter", ok: false},
		{name: "unsanctioned CLDR unit", key: "duration-day-person", ok: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := unitIdentifierFromCLDRKey(tc.key)
			if got != tc.want || ok != tc.ok {
				t.Fatalf("unitIdentifierFromCLDRKey(%q) = %q, %t, want %q, %t", tc.key, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func wantUnitCompound() map[string]string {
	return map[string]string{
		"long":   "{0} per {1}",
		"short":  "{0}/{1}",
		"narrow": "{0}/{1}",
	}
}

func assertUnitsByLocale(t *testing.T, name string, got, want map[string]Units) {
	t.Helper()

	if !maps.EqualFunc(got, want, unitsEqual) {
		t.Fatalf("%s = %#v, want %#v", name, got, want)
	}
}

func unitsEqual(got, want Units) bool {
	return maps.EqualFunc(got, want, unitDataEqual)
}

func unitDataEqual(got, want UnitData) bool {
	return maps.Equal(got.Compound, want.Compound) &&
		maps.EqualFunc(got.Patterns, want.Patterns, unitWidthPatternsEqual)
}

func unitWidthPatternsEqual(got, want map[string]map[string]string) bool {
	return maps.EqualFunc(got, want, maps.Equal)
}
