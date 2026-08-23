package cldr

import (
	"encoding/json/jsontext"
	"maps"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestLoadDatesLoadsGregorianCalendar(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteDayPeriods(t, root)
	mustWriteFile(t, filepath.Join(root, "cldr-dates-full", "main", "en", gregorianCalendarFile), `{
		"main": {
			"en": {
				"dates": {
					"calendars": {
							"gregorian": {
								"dateFormats": {
									"full": "EEEE, MMMM d, y",
									"long": "MMMM d, y",
									"medium": "MMM d, y",
									"short": "M/d/yy"
								},
								"timeFormats": {
									"full": "h:mm:ss a zzzz",
									"long": "h:mm:ss a z",
									"medium": "h:mm:ss a",
									"short": "h:mm a"
								},
								"dateTimeFormats": {
									"full": "{1} 'at' {0}",
									"long": "{1} 'at' {0}",
									"medium": "{1}, {0}",
									"short": "{1}, {0}"
								}
							}
						}
				}
			}
		}
	}`)

	got, err := loadDates(root, []string{undefinedLocale, "missing", "en"})
	if err != nil {
		t.Fatalf("loadDates() error = %v", err)
	}
	if _, ok := got[undefinedLocale]; ok {
		t.Fatalf("loadDates() retained %q locale", undefinedLocale)
	}
	if _, ok := got["missing"]; ok {
		t.Fatal("loadDates() retained locale without date file")
	}
	calendar, ok := got["en"].Calendars[gregorianCalendarType]
	if !ok {
		t.Fatalf("loadDates()[en].Calendars missing %q calendar", gregorianCalendarType)
	}
	assertStringMap(t, "gregorian date formats", calendar.DateFormats, map[string]string{
		"full":   "EEEE, MMMM d, y",
		"long":   "MMMM d, y",
		"medium": "MMM d, y",
		"short":  "M/d/yy",
	})
	assertDayPeriodRanges(t, "day period rules", got["en"].DayPeriodRules["en"], []DayPeriodRange{
		{From: 0, To: 12 * time.Hour, Type: "am"},
	})
}

func TestLoadDatesRejectsInvalidShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		doc  string
	}{
		{
			name: "missing main",
			doc:  `{}`,
		},
		{
			name: "missing locale body",
			doc: `{
				"main": {}
			}`,
		},
		{
			name: "missing dates",
			doc: `{
				"main": {
					"en": {}
				}
			}`,
		},
		{
			name: "missing calendars",
			doc: `{
				"main": {
					"en": {
						"dates": {}
					}
				}
			}`,
		},
		{
			name: "null calendars",
			doc: `{
				"main": {
					"en": {
						"dates": {
							"calendars": null
						}
					}
				}
			}`,
		},
		{
			name: "missing gregorian",
			doc: `{
				"main": {
					"en": {
						"dates": {
							"calendars": {}
						}
					}
				}
			}`,
		},
		{
			name: "null gregorian",
			doc: `{
				"main": {
					"en": {
						"dates": {
							"calendars": {
								"gregorian": null
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
			mustWriteDayPeriods(t, root)
			mustWriteFile(t, filepath.Join(root, "cldr-dates-full", "main", "en", gregorianCalendarFile), tc.doc)

			if _, err := loadDates(root, []string{"en"}); err == nil {
				t.Fatal("loadDates() succeeded, want error")
			}
		})
	}
}

func TestLoadDayPeriodRulesRejectsInvalidShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		doc  string
	}{
		{
			name: "missing supplemental",
			doc:  `{}`,
		},
		{
			name: "missing rule set",
			doc:  `{"supplemental":{}}`,
		},
		{
			name: "null rule set",
			doc:  `{"supplemental":{"dayPeriodRuleSet":null}}`,
		},
		{
			name: "wrong rule set type",
			doc:  `{"supplemental":{"dayPeriodRuleSet":"en"}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "dayPeriods.json"), tc.doc)

			if _, err := loadDayPeriodRules(root); err == nil {
				t.Fatal("loadDayPeriodRules() succeeded, want error")
			}
		})
	}
}

func TestParseCalendarExtractsGregorianData(t *testing.T) {
	t.Parallel()

	calendar := mustParseCalendar(t, `{
		"eras": {
			"eraNames": {"0": "Before Common Era", "1": "Common Era"},
			"eraAbbr": {"0": "BCE", "1": "CE"},
			"eraNarrow": {"0": "B", "1": "A"}
		},
		"months": {
			"format": {
				"wide": {"1": "January", "2": "February"},
				"abbreviated": {"1": "Jan", "2": "Feb"}
			},
			"stand-alone": {
				"abbreviated": {"1": "JanSA", "2": "FebSA"}
			}
		},
		"days": {
			"format": {
				"abbreviated": {"sun": "Sun", "mon": "Mon"}
			}
		},
		"quarters": {
			"format": {
				"wide": {"1": "1st quarter", "2": "2nd quarter"}
			}
		},
			"dayPeriods": {
				"format": {
					"wide": {
						"midnight": "midnight",
						"am": "AM",
						"noon": "noon",
						"pm": "PM"
					}
				}
			},
			"dateFormats": {
				"full": "EEEE, MMMM d, y",
				"long": "MMMM d, y",
				"medium": "MMM d, y",
				"short": {"_value": "M/d/yy"}
			},
			"timeFormats": {
				"full": "h:mm:ss a zzzz",
				"long": "h:mm:ss a z",
				"medium": "h:mm:ss a",
				"short": "h:mm a"
			},
			"dateTimeFormats": {
				"full": "{1} 'at' {0}",
				"long": "{1} 'at' {0}",
				"medium": "{1}, {0}",
				"short": "{1}, {0}",
				"availableFormats": {
					"Hms": {"_value": "HH:mm:ss"},
					"yMd": "M/d/y",
					"yMd-alt-variant": "ignored"
				},
				"intervalFormats": {
					"intervalFormatFallback": "{0} - {1}",
					"yMd": {
						"d": "M/d/y - M/d/y",
						"M-alt-variant": "ignored"
					},
					"yMMMd-alt-variant": {
						"d": "ignored"
					}
				},
				"appendItems": {
					"Day": "{0} ({2}: {1})"
				}
			},
		"dateTimeFormats-atTime": {
			"standard": {
				"long": "{1} at {0}"
			}
		}
	}`)

	formatWide := calendar.Names[CalendarNameKey{Width: "wide", Context: "format"}]
	assertStrings(t, "wide format eras", formatWide.Eras, []string{"Before Common Era", "Common Era"})
	assertStrings(t, "wide format months", formatWide.Months, namedSlots(12, map[int]string{0: "January", 1: "February"}))
	assertStrings(t, "wide format quarters", formatWide.Quarters, namedSlots(4, map[int]string{0: "1st quarter", 1: "2nd quarter"}))
	assertStrings(t, "wide format day periods", formatWide.DayPeriods[:4], []string{"midnight", "AM", "noon", "PM"})

	formatAbbr := calendar.Names[CalendarNameKey{Width: "abbreviated", Context: "format"}]
	assertStrings(t, "abbreviated format eras", formatAbbr.Eras, []string{"BCE", "CE"})
	assertStrings(t, "abbreviated format weekdays", formatAbbr.Weekdays, namedSlots(7, map[int]string{0: "Sun", 1: "Mon"}))

	formatNarrow := calendar.Names[CalendarNameKey{Width: "narrow", Context: "format"}]
	assertStrings(t, "narrow format eras", formatNarrow.Eras, []string{"B", "A"})

	standaloneWide := calendar.Names[CalendarNameKey{Width: "wide", Context: "stand-alone"}]
	assertStrings(t, "wide stand-alone months fall back to format", standaloneWide.Months, namedSlots(12, map[int]string{0: "January", 1: "February"}))

	standaloneAbbr := calendar.Names[CalendarNameKey{Width: "abbreviated", Context: "stand-alone"}]
	assertStrings(t, "abbreviated stand-alone months", standaloneAbbr.Months, namedSlots(12, map[int]string{0: "JanSA", 1: "FebSA"}))

	assertStringMap(t, "date formats", calendar.DateFormats, map[string]string{
		"full":   "EEEE, MMMM d, y",
		"long":   "MMMM d, y",
		"medium": "MMM d, y",
		"short":  "M/d/yy",
	})
	assertStringMap(t, "time formats", calendar.TimeFormats, map[string]string{
		"full":   "h:mm:ss a zzzz",
		"long":   "h:mm:ss a z",
		"medium": "h:mm:ss a",
		"short":  "h:mm a",
	})
	assertStringMap(t, "date time formats", calendar.DateTimeFormats, map[string]string{
		"full":   "{1} 'at' {0}",
		"long":   "{1} 'at' {0}",
		"medium": "{1}, {0}",
		"short":  "{1}, {0}",
	})
	assertStringMap(t, "date time at formats", calendar.DateTimeAtFormats, map[string]string{
		"long": "{1} at {0}",
	})
	assertStringMap(t, "available formats", calendar.AvailableFormats, map[string]string{
		"Hms": "HH:mm:ss",
		"yMd": "M/d/y",
	})

	if got, want := calendar.IntervalFormats.FallbackPattern, "{0} - {1}"; got != want {
		t.Fatalf("interval fallback = %q, want %q", got, want)
	}
	assertStringMap(t, "interval yMd formats", calendar.IntervalFormats.BySkeleton["yMd"], map[string]string{
		"d": "M/d/y - M/d/y",
	})
	if _, ok := calendar.IntervalFormats.BySkeleton["yMMMd-alt-variant"]; ok {
		t.Fatal("interval formats retained alt-variant skeleton")
	}

	assertStringMap(t, "append items", calendar.AppendItems, map[string]string{
		"Day": "{0} ({2}: {1})",
	})
}

func TestCalendarNameRoutesCoverContextsAndWidths(t *testing.T) {
	t.Parallel()

	got := make([]CalendarNameKey, len(calendarNameContexts)*len(calendarNameWidths))
	i := 0
	for _, context := range calendarNameContexts {
		for _, width := range calendarNameWidths {
			got[i] = CalendarNameKey{Width: width, Context: context}
			i++
		}
	}
	want := []CalendarNameKey{
		{Width: "wide", Context: "format"},
		{Width: "abbreviated", Context: "format"},
		{Width: "narrow", Context: "format"},
		{Width: "wide", Context: "stand-alone"},
		{Width: "abbreviated", Context: "stand-alone"},
		{Width: "narrow", Context: "stand-alone"},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("calendar name routes = %#v, want %#v", got, want)
	}
}

func TestCalendarValueOrders(t *testing.T) {
	t.Parallel()

	assertStrings(t, "calendar style order", calendarStyleOrder[:], []string{"full", "long", "medium", "short"})
	assertStrings(t, "weekday name order", weekdayNameOrder[:], []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"})
	assertStrings(t, "day period name order", dayPeriodNameOrder[:], []string{
		"midnight", "am", "noon", "pm",
		"morning1", "morning2", "afternoon1", "afternoon2",
		"evening1", "evening2", "night1", "night2",
	})
}

func TestCalendarNameOrdersPreserveSlots(t *testing.T) {
	t.Parallel()

	assertStrings(t, "era names", orderedEraValues(map[string]string{"1": "Common Era"}), []string{"", "Common Era"})
	assertStrings(t, "numbered names", orderedNumberValues(map[string]string{"2": "February"}, 12), namedSlots(12, map[int]string{1: "February"}))
	assertStrings(t, "weekday names", orderedWeekdayValues(map[string]string{"mon": "Mon"}), namedSlots(7, map[int]string{1: "Mon"}))
}

func TestCalendarAltVariantSuffix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key  string
		want bool
	}{
		{key: "yMd-alt-variant", want: true},
		{key: "M-alt-variant", want: true},
		{key: "intervalFormatFallback", want: false},
		{key: "yMd", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.key, func(t *testing.T) {
			t.Parallel()

			if got := isCalendarAltVariant(tc.key); got != tc.want {
				t.Fatalf("isCalendarAltVariant(%q) = %v, want %v", tc.key, got, tc.want)
			}
		})
	}
}

func TestParseCalendarFallsBackDateTimeAtFormats(t *testing.T) {
	t.Parallel()

	calendar := mustParseCalendar(t, `{
		"dateFormats": {
			"full": "EEEE, MMMM d, y",
			"long": "MMMM d, y",
			"medium": "MMM d, y",
			"short": "M/d/yy"
		},
		"timeFormats": {
			"full": "h:mm:ss a zzzz",
			"long": "h:mm:ss a z",
			"medium": "h:mm:ss a",
			"short": "h:mm a"
		},
		"dateTimeFormats": {
			"full": "{1} at {0}",
			"long": "{1} at {0}",
			"medium": "{1}, {0}",
			"short": "{1}, {0}"
		}
	}`)

	if !maps.Equal(calendar.DateTimeAtFormats, calendar.DateTimeFormats) {
		t.Fatalf("DateTimeAtFormats = %#v, want DateTimeFormats %#v", calendar.DateTimeAtFormats, calendar.DateTimeFormats)
	}
}

func TestRequiredStyleFormatsRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values map[string]jsontext.Value
	}{
		{
			name: "missing style",
			values: map[string]jsontext.Value{
				"long":   jsontext.Value(`"MMMM d, y"`),
				"medium": jsontext.Value(`"MMM d, y"`),
				"short":  jsontext.Value(`"M/d/yy"`),
			},
		},
		{
			name: "empty style",
			values: map[string]jsontext.Value{
				"full":   jsontext.Value(`""`),
				"long":   jsontext.Value(`"MMMM d, y"`),
				"medium": jsontext.Value(`"MMM d, y"`),
				"short":  jsontext.Value(`"M/d/yy"`),
			},
		},
		{
			name: "invalid style",
			values: map[string]jsontext.Value{
				"full":   jsontext.Value(`["bad"]`),
				"long":   jsontext.Value(`"MMMM d, y"`),
				"medium": jsontext.Value(`"MMM d, y"`),
				"short":  jsontext.Value(`"M/d/yy"`),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := requiredStyleFormats(tc.values); err == nil {
				t.Fatal("requiredStyleFormats() succeeded, want error")
			}
		})
	}
}

func TestIntervalFormatsRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  jsontext.Value
	}{
		{
			name: "missing fallback",
			raw:  jsontext.Value(`{"yMd":{"d":"M/d/y - M/d/y"}}`),
		},
		{
			name: "empty fallback",
			raw:  jsontext.Value(`{"intervalFormatFallback":""}`),
		},
		{
			name: "invalid fallback placeholders",
			raw:  jsontext.Value(`{"intervalFormatFallback":"{0} -"}`),
		},
		{
			name: "empty field pattern",
			raw:  jsontext.Value(`{"intervalFormatFallback":"{0} - {1}","yMd":{"d":""}}`),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := intervalFormats(tc.raw); err == nil {
				t.Fatal("intervalFormats() succeeded, want error")
			}
		})
	}
}

func TestAppendItemsRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  jsontext.Value
	}{
		{
			name: "empty pattern",
			raw:  jsontext.Value(`{"Timezone":""}`),
		},
		{
			name: "missing second placeholder",
			raw:  jsontext.Value(`{"Timezone":"{0}"}`),
		},
		{
			name: "duplicate first placeholder",
			raw:  jsontext.Value(`{"Timezone":"{0} {0} {1}"}`),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := appendItems(tc.raw); err == nil {
				t.Fatal("appendItems() succeeded, want error")
			}
		})
	}
}

func TestParseCalendarRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	if _, err := parseCalendar(jsontext.Value(`{`)); err == nil {
		t.Fatal("parseCalendar({) succeeded, want error")
	}
	if _, err := parseCalendar(jsontext.Value(`null`)); err == nil {
		t.Fatal("parseCalendar(null) succeeded, want error")
	}
}

func TestParseCalendarRejectsInvalidSubfields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "date format value",
			raw: `{
					"dateFormats": {
						"short": ["bad"]
					}
				}`,
		},
		{
			name: "time format value",
			raw: `{
					"timeFormats": {
						"short": ["bad"]
					}
				}`,
		},
		{
			name: "date time format value",
			raw: `{
					"dateTimeFormats": {
						"short": ["bad"]
					}
				}`,
		},
		{
			name: "available formats",
			raw: `{
					"dateTimeFormats": {
						"availableFormats": "bad"
				}
			}`,
		},
		{
			name: "available format value",
			raw: `{
						"dateTimeFormats": {
						"availableFormats": {
							"yMd": ["bad"]
						}
						}
					}`,
		},
		{
			name: "empty available format value",
			raw: `{
						"dateTimeFormats": {
							"availableFormats": {
								"yMd": ""
							}
						}
					}`,
		},
		{
			name: "interval formats",
			raw: `{
					"dateTimeFormats": {
						"intervalFormats": "bad"
					}
			}`,
		},
		{
			name: "interval fallback value",
			raw: `{
					"dateTimeFormats": {
						"intervalFormats": {
							"intervalFormatFallback": ["bad"]
						}
					}
				}`,
		},
		{
			name: "interval skeleton",
			raw: `{
					"dateTimeFormats": {
						"intervalFormats": {
							"yMd": "bad"
					}
				}
				}`,
		},
		{
			name: "interval field value",
			raw: `{
					"dateTimeFormats": {
						"intervalFormats": {
							"yMd": {
								"d": ["bad"]
							}
						}
					}
				}`,
		},
		{
			name: "append items",
			raw: `{
					"dateTimeFormats": {
						"appendItems": "bad"
				}
			}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := parseCalendar(jsontext.Value(tc.raw)); err == nil {
				t.Fatal("parseCalendar() succeeded, want error")
			}
		})
	}
}

func TestParseDayPeriodRuleSetSupportsCLDRShapes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "wrapped",
			raw: `{
				"dayPeriodRules": {
					"am": {"_from": "00:00", "_before": "12:00"},
					"noon": {"_at": "12:00"},
					"pm": {"_from": "12:00", "_before": "24:00"}
				}
			}`,
		},
		{
			name: "flat",
			raw: `{
				"am": {"_from": "00:00", "_before": "12:00"},
				"noon": {"_at": "12:00"},
				"pm": {"_from": "12:00", "_before": "24:00"}
			}`,
		},
	}
	want := []DayPeriodRange{
		{From: 0, To: 12 * time.Hour, Type: "am"},
		{From: 12 * time.Hour, To: 12 * time.Hour, Type: "noon"},
		{From: 12 * time.Hour, To: 24 * time.Hour, Type: "pm"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseDayPeriodRuleSet(jsontext.Value(tc.raw))
			if err != nil {
				t.Fatalf("parseDayPeriodRuleSet() error = %v", err)
			}
			assertDayPeriodRanges(t, tc.name, got, want)
		})
	}
}

func TestParseDayPeriodRuleSetRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "invalid json", raw: `{`},
		{name: "invalid at", raw: `{"dayPeriodRules": {"noon": {"_at": "bad"}}}`},
		{name: "invalid from", raw: `{"dayPeriodRules": {"am": {"_from": "bad", "_before": "12:00"}}}`},
		{name: "invalid before", raw: `{"dayPeriodRules": {"am": {"_from": "00:00", "_before": "bad"}}}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := parseDayPeriodRuleSet(jsontext.Value(tc.raw)); err == nil {
				t.Fatal("parseDayPeriodRuleSet() succeeded, want error")
			}
		})
	}
}

func TestParseDayPeriodClock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value string
		want  time.Duration
	}{
		{value: "00:00", want: 0},
		{value: "06:30", want: 6*time.Hour + 30*time.Minute},
		{value: "24:00", want: 24 * time.Hour},
	}
	for _, tc := range tests {
		t.Run(tc.value, func(t *testing.T) {
			t.Parallel()

			got, err := parseDayPeriodClock(tc.value)
			if err != nil {
				t.Fatalf("parseDayPeriodClock(%q) error = %v", tc.value, err)
			}
			if got != tc.want {
				t.Fatalf("parseDayPeriodClock(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestParseDayPeriodClockRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"6", "aa:00", "06:aa", "-1:00", "12:-1", "12:60", "24:01", "25:00"} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			if _, err := parseDayPeriodClock(value); err == nil {
				t.Fatalf("parseDayPeriodClock(%q) succeeded, want error", value)
			}
		})
	}
}

func mustParseCalendar(t *testing.T, raw string) Calendar {
	t.Helper()

	calendar, err := parseCalendar(jsontext.Value(raw))
	if err != nil {
		t.Fatalf("parseCalendar() error = %v", err)
	}
	return calendar
}

func mustWriteDayPeriods(t *testing.T, root string) {
	t.Helper()

	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "dayPeriods.json"), `{
		"supplemental": {
			"dayPeriodRuleSet": {
				"en": {
					"dayPeriodRules": {
						"am": {"_from": "00:00", "_before": "12:00"}
					}
				}
			}
		}
	}`)
}

func assertStrings(t *testing.T, name string, got, want []string) {
	t.Helper()

	if !slices.Equal(got, want) {
		t.Fatalf("%s = %#v, want %#v", name, got, want)
	}
}

func namedSlots(count int, values map[int]string) []string {
	out := make([]string, count)
	for index, value := range values {
		out[index] = value
	}
	return out
}

func assertStringMap(t *testing.T, name string, got, want map[string]string) {
	t.Helper()

	if !maps.Equal(got, want) {
		t.Fatalf("%s = %#v, want %#v", name, got, want)
	}
}

func assertDayPeriodRanges(t *testing.T, name string, got, want []DayPeriodRange) {
	t.Helper()

	if !slices.Equal(got, want) {
		t.Fatalf("%s = %#v, want %#v", name, got, want)
	}
}
