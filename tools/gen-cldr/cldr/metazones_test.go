package cldr

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestLoadMetazonesComposesZoneHistoryAndLocalizedNames(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteMetazoneFixture(t, root)
	mustWriteTimeZoneNamesFixture(t, root, "en")

	got, err := loadMetazones(root, []string{undefinedLocale, "en", "missing"})
	if err != nil {
		t.Fatalf("loadMetazones() error = %v", err)
	}
	want := Metazones{
		ZoneToMetazones: map[string][]MetazonePeriod{
			"America/Argentina/La_Rioja": {
				{Metazone: "Argentina", Start: openMetazoneStart, End: unixMillis(1991, time.March, 1, 2, 0)},
				{Metazone: "Argentina_Western", Start: unixMillis(1991, time.March, 1, 2, 0), End: openMetazoneEnd},
			},
			"America/Chicago": {
				{Metazone: "America_Central", Start: openMetazoneStart, End: openMetazoneEnd},
				{Metazone: "America_Central_Daylight", Start: unixMillis(2000, time.January, 1, 0, 0), End: openMetazoneEnd},
			},
			"America/New_York": {
				{Metazone: "America_Eastern", Start: unixMillis(1970, time.January, 1, 0, 0), End: unixMillis(2007, time.March, 11, 7, 0)},
			},
			"Europe/London": {
				{Metazone: "GMT", Start: openMetazoneStart, End: unixMillis(1996, time.March, 31, 1, 0)},
			},
		},
		Names: map[string]map[string]MetazoneNames{
			"en": {
				"America_Eastern": {
					LongGeneric:   "Eastern Time",
					LongStandard:  "Eastern Standard Time",
					LongDaylight:  "Eastern Daylight Time",
					ShortGeneric:  "ET",
					ShortStandard: "EST",
					ShortDaylight: "EDT",
				},
			},
		},
		ZoneNames: map[string]map[string]MetazoneNames{
			"en": {
				"America/New_York": {
					LongGeneric:   "New York Time",
					LongStandard:  "New York Standard Time",
					LongDaylight:  "New York Daylight Time",
					ShortStandard: "NYST",
				},
			},
		},
		ExemplarCities: map[string]map[string]string{
			"en": {
				"America/Chicago":  "Chicago",
				"America/New_York": "New York",
			},
		},
		Formats: map[string]TimeZoneFormats{
			"en": {GMTFormat: "GMT{0}", GMTZeroFormat: "GMT", HourFormat: "+HH:mm;-HH:mm", RegionFormat: "{0} Time"},
		},
	}
	assertMetazones(t, "loadMetazones()", got, want)
}

func TestLoadZoneToMetazoneRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "invalid json", raw: `{`},
		{name: "missing supplemental", raw: `{}`},
		{
			name: "missing metaZones",
			raw:  `{"supplemental":{}}`,
		},
		{
			name: "missing metazoneInfo",
			raw: `{
				"supplemental": {
					"metaZones": {}
				}
			}`,
		},
		{
			name: "missing timezone",
			raw: `{
				"supplemental": {
					"metaZones": {
						"metazoneInfo": {}
					}
				}
			}`,
		},
		{
			name: "null timezone",
			raw: `{
				"supplemental": {
					"metaZones": {
						"metazoneInfo": {
							"timezone": null
						}
					}
				}
			}`,
		},
		{
			name: "wrong timezone type",
			raw: `{
				"supplemental": {
					"metaZones": {
						"metazoneInfo": {
							"timezone": "America"
						}
					}
				}
			}`,
		},
		{
			name: "invalid boundary",
			raw: `{
					"supplemental": {
						"metaZones": {
							"metazoneInfo": {
								"timezone": {
								"America": {
									"New_York": {
										"usesMetazone": {
											"_mzone": "America_Eastern",
											"_from": "soon"
										}
									}
								}
							}
							}
						}
					}
				}`,
		},
		{
			name: "missing metazone",
			raw: `{
					"supplemental": {
						"metaZones": {
							"metazoneInfo": {
								"timezone": {
									"America": {
										"New_York": {
											"usesMetazone": {
												"_from": "2000-01-01 00:00"
											}
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
			mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "metaZones.json"), tc.raw)

			if _, err := loadZoneToMetazone(root); err == nil {
				t.Fatal("loadZoneToMetazone() succeeded, want error")
			}
		})
	}
}

func TestLoadTimeZoneNamesRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "invalid json", raw: `{`},
		{
			name: "missing body",
			raw: `{
					"main": {
						"fr": {
						"dates": {
							"timeZoneNames": {}
						}
					}
					}
				}`,
		},
		{
			name: "missing dates",
			raw: `{
					"main": {
						"en": {}
					}
				}`,
		},
		{
			name: "null dates",
			raw: `{
					"main": {
						"en": {
							"dates": null
						}
					}
				}`,
		},
		{
			name: "missing timeZoneNames",
			raw: `{
					"main": {
						"en": {
							"dates": {}
						}
					}
				}`,
		},
		{
			name: "null timeZoneNames",
			raw: `{
					"main": {
						"en": {
							"dates": {
								"timeZoneNames": null
							}
						}
					}
				}`,
		},
		{
			name: "missing gmtFormat",
			raw: `{
					"main": {
						"en": {
							"dates": {
								"timeZoneNames": {
									"gmtZeroFormat": "GMT",
									"hourFormat": "+HH:mm;-HH:mm"
								}
							}
						}
					}
				}`,
		},
		{
			name: "missing gmtZeroFormat",
			raw: `{
					"main": {
						"en": {
							"dates": {
								"timeZoneNames": {
									"gmtFormat": "GMT{0}",
									"hourFormat": "+HH:mm;-HH:mm"
								}
							}
						}
					}
				}`,
		},
		{
			name: "missing hourFormat",
			raw: `{
					"main": {
						"en": {
							"dates": {
								"timeZoneNames": {
									"gmtFormat": "GMT{0}",
									"gmtZeroFormat": "GMT"
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
			mustWriteFile(t, filepath.Join(root, "cldr-dates-full", "main", "en", "timeZoneNames.json"), tc.raw)

			if _, _, _, _, err := loadTimeZoneNames(root, []string{"en"}); err == nil {
				t.Fatal("loadTimeZoneNames() succeeded, want error")
			}
		})
	}
}

func TestLoadTimeZoneNamesRejectsInvalidRegionFormat(t *testing.T) {
	t.Parallel()

	for _, regionFormat := range []string{"Location time", "{0} / {0}", "{1} Time"} {
		t.Run(regionFormat, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			raw := `{"main":{"en":{"dates":{"timeZoneNames":{"gmtFormat":"GMT{0}","gmtZeroFormat":"GMT","hourFormat":"+HH:mm;-HH:mm","regionFormat":` + fmt.Sprintf("%q", regionFormat) + `}}}}}`
			mustWriteFile(t, filepath.Join(root, "cldr-dates-full", "main", "en", "timeZoneNames.json"), raw)

			_, _, _, _, err := loadTimeZoneNames(root, []string{"en"})
			if err == nil {
				t.Fatal("loadTimeZoneNames() succeeded, want error")
			}
			for _, text := range []string{"en", "regionFormat", regionFormat} {
				if !strings.Contains(err.Error(), text) {
					t.Fatalf("loadTimeZoneNames() error = %q, want %q", err, text)
				}
			}
		})
	}
}

func TestLoadTimeZoneNamesUsesRootRegionFormatWhenMissing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	raw := `{"main":{"en":{"dates":{"timeZoneNames":{"gmtFormat":"GMT{0}","gmtZeroFormat":"GMT","hourFormat":"+HH:mm;-HH:mm"}}}}}`
	mustWriteFile(t, filepath.Join(root, "cldr-dates-full", "main", "en", "timeZoneNames.json"), raw)

	_, _, _, formats, err := loadTimeZoneNames(root, []string{"en"})
	if err != nil {
		t.Fatalf("loadTimeZoneNames() error = %v", err)
	}
	if got, want := formats["en"].RegionFormat, "{0}"; got != want {
		t.Fatalf("RegionFormat = %q, want CLDR root %q", got, want)
	}
}

func mustWriteMetazoneFixture(t *testing.T, root string) {
	t.Helper()

	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "metaZones.json"), `{
		"supplemental": {
			"metaZones": {
				"metazoneInfo": {
					"timezone": {
						"America": {
							"Argentina": {
								"La_Rioja": [
									{
										"usesMetazone": {
											"_mzone": "Argentina",
											"_to": "1991-03-01 02:00"
										}
									},
									{
										"usesMetazone": {
											"_mzone": "Argentina_Western",
											"_from": "1991-03-01 02:00"
										}
									}
								]
							},
							"Chicago": [
								{
									"usesMetazone": {
										"_mzone": "America_Central"
									}
								},
								{
									"usesMetazone": {
										"_mzone": "America_Central_Daylight",
										"_from": "2000-01-01 00:00"
									}
								}
							],
							"New_York": {
								"usesMetazone": {
									"_mzone": "America_Eastern",
									"_from": "1970-01-01 00:00",
									"_to": "2007-03-11 07:00"
								}
							}
						},
							"Europe": {
								"London": {
									"usesMetazone": {
										"_mzone": "GMT",
										"_to": "1996-03-31 01:00"
									}
								}
							}
						}
					}
				}
			}
		}`)
}

func mustWriteTimeZoneNamesFixture(t *testing.T, root, locale string) {
	t.Helper()

	mustWriteFile(t, filepath.Join(root, "cldr-dates-full", "main", locale, "timeZoneNames.json"), `{
		"main": {
			"`+locale+`": {
				"dates": {
					"timeZoneNames": {
						"gmtFormat": "GMT{0}",
						"gmtZeroFormat": "GMT",
						"hourFormat": "+HH:mm;-HH:mm",
						"regionFormat": "{0} Time",
						"metazone": {
							"America_Eastern": {
								"long": {
									"generic": "Eastern Time",
									"standard": "Eastern Standard Time",
									"daylight": "Eastern Daylight Time"
								},
								"short": {
									"generic": "ET",
									"standard": "EST",
									"daylight": "EDT"
								}
							}
						},
						"zone": {
							"America": {
								"Chicago": {
									"exemplarCity": "Chicago"
								},
								"New_York": {
									"exemplarCity": "New York",
									"long": {
										"generic": "New York Time",
										"standard": "New York Standard Time",
										"daylight": "New York Daylight Time"
									},
									"short": {
										"standard": "NYST"
									}
								}
							},
							"Etc": {
								"UTC": {}
							}
						}
					}
				}
			}
		}
	}`)
}

func unixMillis(year int, month time.Month, day, hour, minute int) int64 {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC).UnixMilli()
}

func assertMetazones(t *testing.T, name string, got, want Metazones) {
	t.Helper()

	if !maps.EqualFunc(got.ZoneToMetazones, want.ZoneToMetazones, slices.Equal) ||
		!maps.EqualFunc(got.Names, want.Names, maps.Equal) ||
		!maps.EqualFunc(got.ZoneNames, want.ZoneNames, maps.Equal) ||
		!maps.EqualFunc(got.ExemplarCities, want.ExemplarCities, maps.Equal) ||
		!maps.Equal(got.Formats, want.Formats) {
		t.Fatalf("%s = %#v, want %#v", name, got, want)
	}
}
