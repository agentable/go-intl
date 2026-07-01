package cldr

import (
	"maps"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDisplayNamesMapsAndInherits(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteDisplayNamesFixture(t, root, "en")
	mustWriteFile(t, filepath.Join(root, "cldr-localenames-full", "main", "fr", "territories.json"), `{
		"main": {
			"fr": {
				"localeDisplayNames": {
					"territories": {
						"FR": "France"
					}
				}
			}
		}
	}`)

	got, err := loadDisplayNames(root, []string{undefinedLocale, "en-US", "en", "fr", "missing"})
	if err != nil {
		t.Fatalf("loadDisplayNames() error = %v", err)
	}

	en := DisplayNames{
		Languages: LanguageDisplay{
			Dialect: StyledNames{
				Long: map[string]string{
					"en":    "English",
					"en-US": "American English",
					"fr":    "French",
				},
				Short: map[string]string{
					"en-US": "US English",
					"fr":    "Fr.",
				},
				Narrow: map[string]string{
					"en-US": "US Eng.",
				},
			},
			Standard: StyledNames{
				Long: map[string]string{
					"en":    "English",
					"en-US": "English (United States)",
					"fr":    "French",
				},
				Short: map[string]string{
					"en-US": "English (US)",
					"fr":    "Fr.",
				},
				Narrow: map[string]string{
					"en-US": "English (U.S.)",
				},
			},
		},
		Territories: StyledNames{
			Long: map[string]string{
				"FR": "France",
				"US": "United States",
			},
			Short: map[string]string{
				"US": "US",
			},
			Narrow: map[string]string{
				"US": "U.S.",
			},
		},
		Scripts: StyledNames{
			Long: map[string]string{
				"Latn": "Latin",
			},
			Short: map[string]string{
				"Latn": "Latn",
			},
			Narrow: map[string]string{
				"Cyrl": "C",
			},
		},
		Calendars: StyledNames{
			Long: map[string]string{
				"gregorian": "Gregorian Calendar",
			},
			Short: map[string]string{
				"buddhist": "Buddhist",
			},
			Narrow: map[string]string{
				"islamic": "Islamic",
			},
		},
		DateTimeFields: StyledNames{
			Long: map[string]string{
				"month":        "month",
				"timeZoneName": "time zone",
				"weekOfYear":   "week",
			},
			Short: map[string]string{
				"month":      "mo.",
				"weekOfYear": "wk.",
			},
			Narrow: map[string]string{
				"second":       "sec.",
				"timeZoneName": "tz",
			},
		},
		LocalePattern: "{0} ({1})",
	}
	want := map[string]DisplayNames{
		"en":    en,
		"en-US": en,
	}
	assertDisplayNamesByLocale(t, "loadDisplayNames()", got, want)
}

func TestLoadDisplayNamesRejectsInvalidLanguagesJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-localenames-full", "main", "en", "languages.json"), `{`)

	if _, err := loadDisplayNames(root, []string{"en"}); err == nil {
		t.Fatal("loadDisplayNames() succeeded, want error")
	}
}

func TestLoadDisplayNamesRejectsInvalidLocaleNamesShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		file          string
		withLanguages bool
		body          string
	}{
		{
			name: "languages missing main",
			file: "languages.json",
			body: `{}`,
		},
		{
			name: "languages missing locale body",
			file: "languages.json",
			body: `{
				"main": {
					"fr": {
						"localeDisplayNames": {
							"languages": {
								"fr": "French"
							}
						}
					}
				}
			}`,
		},
		{
			name: "languages missing localeDisplayNames",
			file: "languages.json",
			body: `{
				"main": {
					"en": {}
				}
			}`,
		},
		{
			name: "languages missing field",
			file: "languages.json",
			body: `{
				"main": {
					"en": {
						"localeDisplayNames": {}
					}
				}
			}`,
		},
		{
			name: "languages null field",
			file: "languages.json",
			body: `{
				"main": {
					"en": {
						"localeDisplayNames": {
							"languages": null
						}
					}
				}
			}`,
		},
		{
			name: "languages empty field",
			file: "languages.json",
			body: `{
				"main": {
					"en": {
						"localeDisplayNames": {
							"languages": {}
						}
					}
				}
			}`,
		},
		{
			name:          "territories missing field",
			file:          "territories.json",
			withLanguages: true,
			body: `{
				"main": {
					"en": {
						"localeDisplayNames": {}
					}
				}
			}`,
		},
		{
			name:          "scripts null field",
			file:          "scripts.json",
			withLanguages: true,
			body: `{
				"main": {
					"en": {
						"localeDisplayNames": {
							"scripts": null
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
			if tc.withLanguages {
				mustWriteMinimalLanguages(t, root, "en")
			}
			mustWriteFile(t, filepath.Join(root, "cldr-localenames-full", "main", "en", tc.file), tc.body)

			if _, err := loadDisplayNames(root, []string{"en"}); err == nil {
				t.Fatal("loadDisplayNames() succeeded, want error")
			}
		})
	}
}

func TestLoadDisplayNamesRejectsInvalidLocaleDisplayNamesShape(t *testing.T) {
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
			}`,
		},
		{
			name: "missing localeDisplayNames",
			body: `{
				"main": {
					"en": {}
				}
			}`,
		},
		{
			name: "missing localeDisplayPattern",
			body: `{
				"main": {
					"en": {
						"localeDisplayNames": {
							"types": {
								"calendar": {
									"gregorian": "Gregorian Calendar"
								}
							}
						}
					}
				}
			}`,
		},
		{
			name: "missing localePattern",
			body: `{
				"main": {
					"en": {
						"localeDisplayNames": {
							"localeDisplayPattern": {},
							"types": {
								"calendar": {
									"gregorian": "Gregorian Calendar"
								}
							}
						}
					}
				}
			}`,
		},
		{
			name: "empty localePattern",
			body: `{
				"main": {
					"en": {
						"localeDisplayNames": {
							"localeDisplayPattern": {
								"localePattern": ""
							},
							"types": {
								"calendar": {
									"gregorian": "Gregorian Calendar"
								}
							}
						}
					}
				}
			}`,
		},
		{
			name: "missing types",
			body: `{
				"main": {
					"en": {
						"localeDisplayNames": {
							"localeDisplayPattern": {
								"localePattern": "{0} ({1})"
							}
						}
					}
				}
			}`,
		},
		{
			name: "missing calendar",
			body: `{
				"main": {
					"en": {
						"localeDisplayNames": {
							"localeDisplayPattern": {
								"localePattern": "{0} ({1})"
							},
							"types": {}
						}
					}
				}
			}`,
		},
		{
			name: "null calendar",
			body: `{
				"main": {
					"en": {
						"localeDisplayNames": {
							"localeDisplayPattern": {
								"localePattern": "{0} ({1})"
							},
							"types": {
								"calendar": null
							}
						}
					}
				}
			}`,
		},
		{
			name: "empty calendar",
			body: `{
				"main": {
					"en": {
						"localeDisplayNames": {
							"localeDisplayPattern": {
								"localePattern": "{0} ({1})"
							},
							"types": {
								"calendar": {}
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
			mustWriteMinimalLanguages(t, root, "en")
			mustWriteFile(t, filepath.Join(root, "cldr-localenames-full", "main", "en", "localeDisplayNames.json"), tc.body)

			if _, err := loadDisplayNames(root, []string{"en"}); err == nil {
				t.Fatal("loadDisplayNames() succeeded, want error")
			}
		})
	}
}

func TestLoadDisplayNamesRejectsInvalidDateTimeFieldsJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteMinimalLanguages(t, root, "en")
	mustWriteFile(t, filepath.Join(root, "cldr-dates-full", "main", "en", "dateFields.json"), `{`)

	if _, err := loadDisplayNames(root, []string{"en"}); err == nil {
		t.Fatal("loadDisplayNames() succeeded, want error")
	}
}

func TestLoadDisplayNamesRejectsInvalidDateTimeFieldsShape(t *testing.T) {
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
						"dates": {
							"fields": {
								"year": {"displayName": "year"}
							}
						}
					}
				}
			}`,
		},
		{
			name: "missing dates",
			body: `{
				"main": {
					"en": {}
				}
			}`,
		},
		{
			name: "missing fields",
			body: `{
				"main": {
					"en": {
						"dates": {}
					}
				}
			}`,
		},
		{
			name: "null fields",
			body: `{
				"main": {
					"en": {
						"dates": {
							"fields": null
						}
					}
				}
			}`,
		},
		{
			name: "empty fields",
			body: `{
				"main": {
					"en": {
						"dates": {
							"fields": {}
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
			mustWriteMinimalLanguages(t, root, "en")
			mustWriteFile(t, filepath.Join(root, "cldr-dates-full", "main", "en", "dateFields.json"), tc.body)

			if _, err := loadDisplayNames(root, []string{"en"}); err == nil {
				t.Fatal("loadDisplayNames() succeeded, want error")
			}
		})
	}
}

func TestLoadDisplayNamesRejectsLanguagesReadError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "cldr-localenames-full", "main", "en", "languages.json")
	if err := os.MkdirAll(path, 0o777); err != nil {
		t.Fatalf("create fixture dir: %v", err)
	}
	if _, err := loadDisplayNames(root, []string{"en"}); err == nil {
		t.Fatal("loadDisplayNames() succeeded, want error")
	}
}

func TestLoadDisplayNamesRejectsLocaleDisplayNamesReadError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteMinimalLanguages(t, root, "en")
	path := filepath.Join(root, "cldr-localenames-full", "main", "en", "localeDisplayNames.json")
	if err := os.MkdirAll(path, 0o777); err != nil {
		t.Fatalf("create fixture dir: %v", err)
	}
	if _, err := loadDisplayNames(root, []string{"en"}); err == nil {
		t.Fatal("loadDisplayNames() succeeded, want error")
	}
}

func TestLoadDisplayNamesRejectsDateTimeFieldsReadError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteMinimalLanguages(t, root, "en")
	path := filepath.Join(root, "cldr-dates-full", "main", "en", "dateFields.json")
	if err := os.MkdirAll(path, 0o777); err != nil {
		t.Fatalf("create fixture dir: %v", err)
	}
	if _, err := loadDisplayNames(root, []string{"en"}); err == nil {
		t.Fatal("loadDisplayNames() succeeded, want error")
	}
}

func mustWriteMinimalLanguages(t *testing.T, root, locale string) {
	t.Helper()

	mustWriteFile(t, filepath.Join(root, "cldr-localenames-full", "main", locale, "languages.json"), `{
		"main": {
			"`+locale+`": {
				"localeDisplayNames": {
					"languages": {
						"`+locale+`": "`+locale+`"
					}
				}
			}
		}
	}`)
}

func mustWriteDisplayNamesFixture(t *testing.T, root, locale string) {
	t.Helper()

	mustWriteFile(t, filepath.Join(root, "cldr-localenames-full", "main", locale, "languages.json"), `{
		"main": {
			"`+locale+`": {
				"localeDisplayNames": {
					"languages": {
						"en": "English",
						"en-US": "American English",
						"en-US-alt-short": "US English",
						"en-US-alt-narrow": "US Eng.",
						"fr": "French",
						"fr-alt-short": "Fr.",
						"de-alt-variant": "ignored",
						"zz": ""
					}
				}
			}
		}
	}`)
	mustWriteFile(t, filepath.Join(root, "cldr-localenames-full", "main", locale, "territories.json"), `{
		"main": {
			"`+locale+`": {
				"localeDisplayNames": {
					"territories": {
						"US": "United States",
						"US-alt-short": "US",
						"US-alt-narrow": "U.S.",
						"FR": "France"
					}
				}
			}
		}
	}`)
	mustWriteFile(t, filepath.Join(root, "cldr-localenames-full", "main", locale, "scripts.json"), `{
		"main": {
			"`+locale+`": {
				"localeDisplayNames": {
					"scripts": {
						"Latn": "Latin",
						"Latn-alt-short": "Latn",
						"Cyrl-alt-narrow": "C",
						"Arab-alt-variant": "ignored"
					}
				}
			}
		}
	}`)
	mustWriteFile(t, filepath.Join(root, "cldr-localenames-full", "main", locale, "localeDisplayNames.json"), `{
		"main": {
			"`+locale+`": {
				"localeDisplayNames": {
					"localeDisplayPattern": {
						"localePattern": "{0} ({1})"
					},
					"types": {
						"calendar": {
							"gregorian": "Gregorian Calendar",
							"buddhist-alt-short": "Buddhist",
							"islamic-alt-narrow": "Islamic"
						}
					}
				}
			}
		}
	}`)
	mustWriteFile(t, filepath.Join(root, "cldr-dates-full", "main", locale, "dateFields.json"), `{
		"main": {
			"`+locale+`": {
				"dates": {
					"fields": {
						"week": {"displayName": "week"},
						"week-short": {"displayName": "wk."},
						"zone": {"displayName": "time zone"},
						"zone-narrow": {"displayName": "tz"},
						"month": {"displayName": "month"},
						"month-short": {"displayName": "mo."},
						"second-narrow": {"displayName": "sec."},
						"era-variant": {"displayName": "ignored"},
						"ignored": {"displayName": ""}
					}
				}
			}
		}
	}`)
}

func assertDisplayNamesByLocale(t *testing.T, name string, got, want map[string]DisplayNames) {
	t.Helper()

	if !maps.EqualFunc(got, want, displayNamesEqual) {
		t.Fatalf("%s = %#v, want %#v", name, got, want)
	}
}

func displayNamesEqual(got, want DisplayNames) bool {
	return languageDisplayEqual(got.Languages, want.Languages) &&
		styledNamesEqual(got.Territories, want.Territories) &&
		styledNamesEqual(got.Scripts, want.Scripts) &&
		styledNamesEqual(got.Calendars, want.Calendars) &&
		styledNamesEqual(got.DateTimeFields, want.DateTimeFields) &&
		got.LocalePattern == want.LocalePattern
}

func languageDisplayEqual(got, want LanguageDisplay) bool {
	return styledNamesEqual(got.Dialect, want.Dialect) &&
		styledNamesEqual(got.Standard, want.Standard)
}

func styledNamesEqual(got, want StyledNames) bool {
	return maps.Equal(got.Long, want.Long) &&
		maps.Equal(got.Short, want.Short) &&
		maps.Equal(got.Narrow, want.Narrow)
}
