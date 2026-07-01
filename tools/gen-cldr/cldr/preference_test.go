package cldr

import (
	"maps"
	"path/filepath"
	"slices"
	"testing"
)

func TestLoadPreferenceData(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWritePreferenceFixtures(t, root)

	got, err := loadPreferenceData(root)
	if err != nil {
		t.Fatalf("loadPreferenceData() error = %v", err)
	}
	want := PreferenceData{
		HourCycle: map[string][]string{
			"001":      {"h23", "h12", "h11", "h24"},
			"US-POSIX": {"h12", "h23"},
			"ZZ":       {},
		},
		Week: map[string]WeekData{
			"001": {
				FirstDay:     "mon",
				WeekendStart: "sat",
				WeekendEnd:   "sun",
				MinDays:      1,
			},
			"AE": {
				FirstDay:     "mon",
				WeekendStart: "fri",
				WeekendEnd:   "sat",
				MinDays:      1,
			},
			"GB": {
				FirstDay:     "mon",
				WeekendStart: "sat",
				WeekendEnd:   "sun",
				MinDays:      4,
			},
			"US": {
				FirstDay:     "sun",
				WeekendStart: "sat",
				WeekendEnd:   "sun",
				MinDays:      1,
			},
		},
		Calendar: map[string][]string{
			"001":      {"gregorian"},
			"IR":       {"persian", "gregorian"},
			"US-POSIX": {"gregorian"},
		},
	}
	assertPreferenceData(t, "loadPreferenceData()", got, want)
}

func TestLoadPreferenceDataRejectsInvalidTimeDataJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "timeData.json"), `{`)

	if _, err := loadPreferenceData(root); err == nil {
		t.Fatal("loadPreferenceData() succeeded, want error")
	}
}

func TestLoadHourCyclePreferenceRejectsInvalidShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "missing timeData", raw: `{}`},
		{
			name: "null timeData",
			raw:  `{"supplemental":{"timeData":null}}`,
		},
		{
			name: "missing world default",
			raw:  `{"supplemental":{"timeData":{"US":{"_allowed":"h"}}}}`,
		},
		{
			name: "empty world default",
			raw:  `{"supplemental":{"timeData":{"001":{}}}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "timeData.json"), tc.raw)

			if _, err := loadHourCyclePreference(root); err == nil {
				t.Fatal("loadHourCyclePreference() succeeded, want error")
			}
		})
	}
}

func TestLoadHourCyclePreferenceRejectsInvalidAllowedToken(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "timeData.json"), `{
		"supplemental": {
			"timeData": {
				"001": {"_allowed": "H x"}
			}
		}
	}`)

	if _, err := loadHourCyclePreference(root); err == nil {
		t.Fatal("loadHourCyclePreference() succeeded, want error")
	}
}

func TestLoadWeekDataRejectsInvalidMinDays(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "weekData.json"), `{
		"supplemental": {
			"weekData": {
				"firstDay": {"001": "mon"},
				"weekendStart": {"001": "sat"},
				"weekendEnd": {"001": "sun"},
				"minDays": {"001": "many"}
			}
		}
	}`)

	if _, err := loadWeekData(root); err == nil {
		t.Fatal("loadWeekData() succeeded, want error")
	}
}

func TestLoadWeekDataRejectsInvalidShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "missing weekData", raw: `{}`},
		{
			name: "missing firstDay map",
			raw: `{
				"supplemental": {
					"weekData": {
						"weekendStart": {"001": "sat"},
						"weekendEnd": {"001": "sun"},
						"minDays": {"001": "1"}
					}
				}
			}`,
		},
		{
			name: "null weekendStart map",
			raw: `{
				"supplemental": {
					"weekData": {
						"firstDay": {"001": "mon"},
						"weekendStart": null,
						"weekendEnd": {"001": "sun"},
						"minDays": {"001": "1"}
					}
				}
			}`,
		},
		{
			name: "missing firstDay world default",
			raw: `{
				"supplemental": {
					"weekData": {
						"firstDay": {"US": "sun"},
						"weekendStart": {"001": "sat"},
						"weekendEnd": {"001": "sun"},
						"minDays": {"001": "1"}
					}
				}
			}`,
		},
		{
			name: "missing minDays world default",
			raw: `{
				"supplemental": {
					"weekData": {
						"firstDay": {"001": "mon"},
						"weekendStart": {"001": "sat"},
						"weekendEnd": {"001": "sun"},
						"minDays": {"US": "1"}
					}
				}
			}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "weekData.json"), tc.raw)

			if _, err := loadWeekData(root); err == nil {
				t.Fatal("loadWeekData() succeeded, want error")
			}
		})
	}
}

func TestLoadWeekDataCanonicalizesRegions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "weekData.json"), `{
		"supplemental": {
			"weekData": {
				"firstDay": {"001": "mon", "US_POSIX": "sun"},
				"weekendStart": {"001": "sat"},
				"weekendEnd": {"001": "sun"},
				"minDays": {"001": "1", "US_POSIX": "4"}
			}
		}
	}`)

	got, err := loadWeekData(root)
	if err != nil {
		t.Fatalf("loadWeekData() error = %v", err)
	}
	want := map[string]WeekData{
		"001": {
			FirstDay:     "mon",
			WeekendStart: "sat",
			WeekendEnd:   "sun",
			MinDays:      1,
		},
		"US-POSIX": {
			FirstDay:     "sun",
			WeekendStart: "sat",
			WeekendEnd:   "sun",
			MinDays:      4,
		},
	}
	if !maps.Equal(got, want) {
		t.Fatalf("loadWeekData() = %#v, want %#v", got, want)
	}
}

func TestLoadCalendarPreferenceCanonicalizesRegions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "calendarPreferenceData.json"), `{
		"supplemental": {
			"calendarPreferenceData": {
				"001": ["gregorian"],
				"US_POSIX": ["gregorian"]
			}
		}
	}`)

	got, err := loadCalendarPreference(root)
	if err != nil {
		t.Fatalf("loadCalendarPreference() error = %v", err)
	}
	want := map[string][]string{
		"001":      {"gregorian"},
		"US-POSIX": {"gregorian"},
	}
	if !maps.EqualFunc(got, want, slices.Equal) {
		t.Fatalf("loadCalendarPreference() = %#v, want %#v", got, want)
	}
}

func TestLoadCalendarPreferenceRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "calendarPreferenceData.json"), `{`)

	if _, err := loadCalendarPreference(root); err == nil {
		t.Fatal("loadCalendarPreference() succeeded, want error")
	}
}

func TestLoadCalendarPreferenceRejectsInvalidShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "missing calendarPreferenceData", raw: `{}`},
		{
			name: "null calendarPreferenceData",
			raw:  `{"supplemental":{"calendarPreferenceData":null}}`,
		},
		{
			name: "missing world default",
			raw:  `{"supplemental":{"calendarPreferenceData":{"US":["gregorian"]}}}`,
		},
		{
			name: "empty world default",
			raw:  `{"supplemental":{"calendarPreferenceData":{"001":[]}}}`,
		},
		{
			name: "wrong region shape",
			raw:  `{"supplemental":{"calendarPreferenceData":{"001":"gregorian"}}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "calendarPreferenceData.json"), tc.raw)

			if _, err := loadCalendarPreference(root); err == nil {
				t.Fatal("loadCalendarPreference() succeeded, want error")
			}
		})
	}
}

func mustWritePreferenceFixtures(t *testing.T, root string) {
	t.Helper()

	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "timeData.json"), `{
		"supplemental": {
			"timeData": {
				"001": {"_allowed": "H h K k hB hb"},
				"US_POSIX": {"_allowed": "h H"},
				"ZZ": {"_allowed": ""}
			}
		}
	}`)
	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "weekData.json"), `{
		"supplemental": {
			"weekData": {
				"firstDay": {"001": "mon", "US": "sun"},
				"weekendStart": {"001": "sat", "AE": "fri"},
				"weekendEnd": {"001": "sun", "AE": "sat"},
				"minDays": {"001": "1", "GB": "4"}
			}
		}
	}`)
	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "calendarPreferenceData.json"), `{
		"supplemental": {
				"calendarPreferenceData": {
					"001": ["gregorian"],
					"IR": ["persian", "gregorian"],
					"US_POSIX": ["gregorian"]
				}
			}
		}`)
}

func assertPreferenceData(t *testing.T, name string, got, want PreferenceData) {
	t.Helper()

	if !maps.EqualFunc(got.HourCycle, want.HourCycle, slices.Equal) ||
		!maps.Equal(got.Week, want.Week) ||
		!maps.EqualFunc(got.Calendar, want.Calendar, slices.Equal) {
		t.Fatalf("%s = %#v, want %#v", name, got, want)
	}
}
