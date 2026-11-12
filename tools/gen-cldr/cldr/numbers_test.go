package cldr

import (
	"encoding/json/jsontext"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestLoadCurrencies(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-numbers-full", "main", "en", "currencies.json"), `{
		"main": {
			"en": {
				"numbers": {
					"currencies": {
						"EUR": {
							"displayName": "Euro",
							"symbol": "€",
							"pattern": "¤#,##0.00"
						},
						"USD": {
							"displayName": "US Dollar",
							"displayName-count-zero": "US dollars",
							"displayName-count-one": "US dollar",
							"displayName-count-two": "US dollars",
							"displayName-count-few": "US dollars",
							"displayName-count-many": "US dollars",
							"displayName-count-other": "US dollars",
							"symbol": "$",
							"symbol-alt-narrow": "$",
							"pattern": "¤ #,##0.00",
							"decimal": ".",
							"group": ","
						}
					}
				}
			}
		}
	}`)

	got, err := loadCurrencies(root, []string{undefinedLocale, "en", "missing"})
	if err != nil {
		t.Fatalf("loadCurrencies() error = %v", err)
	}
	want := map[string]Currencies{
		"en": {
			"EUR": {
				Display:   map[string]string{"other": "Euro"},
				Canonical: "Euro",
				Symbol:    "€",
				Pattern:   "¤#,##0.00",
			},
			"USD": {
				Display: map[string]string{
					"zero":  "US dollars",
					"one":   "US dollar",
					"two":   "US dollars",
					"few":   "US dollars",
					"many":  "US dollars",
					"other": "US dollars",
				},
				Canonical: "US Dollar",
				Symbol:    "$",
				Narrow:    "$",
				Pattern:   "¤ #,##0.00",
				Decimal:   ".",
				Group:     ",",
			},
		},
	}
	assertCurrenciesByLocale(t, "loadCurrencies()", got, want)
}

func TestLoadCurrenciesRejectsInvalidPluralCategory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-numbers-full", "main", "en", "currencies.json"), `{
		"main": {
			"en": {
				"numbers": {
					"currencies": {
						"USD": {
							"displayName": "US Dollar",
							"displayName-count-invalid": "bad"
						}
					}
				}
			}
		}
	}`)

	if _, err := loadCurrencies(root, []string{"en"}); err == nil {
		t.Fatal("loadCurrencies() succeeded, want error")
	}
}

func TestLoadCurrenciesRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-numbers-full", "main", "en", "currencies.json"), `{`)

	if _, err := loadCurrencies(root, []string{"en"}); err == nil {
		t.Fatal("loadCurrencies() succeeded, want error")
	}
}

func TestLoadCurrenciesRejectsInvalidShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		doc  string
	}{
		{
			name: "missing locale body",
			doc: `{
				"main": {}
			}`,
		},
		{
			name: "missing currencies",
			doc: `{
				"main": {
					"en": {
						"numbers": {}
					}
				}
			}`,
		},
		{
			name: "null currencies",
			doc: `{
				"main": {
					"en": {
						"numbers": {
							"currencies": null
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
			mustWriteFile(t, filepath.Join(root, "cldr-numbers-full", "main", "en", "currencies.json"), tc.doc)

			if _, err := loadCurrencies(root, []string{"en"}); err == nil {
				t.Fatal("loadCurrencies() succeeded, want error")
			}
		})
	}
}

func TestLoadNumbersRejectsInvalidNestedJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-numbers-full", "main", "en", "numbers.json"), `{
		"main": {
			"en": {
				"numbers": {
					"defaultNumberingSystem": "latn",
					"symbols-numberSystem-latn": {
						"decimal": "."
					},
					"decimalFormats-numberSystem-latn": "{"
				}
			}
		}
	}`)

	if _, err := loadNumbers(root, []string{"en"}); err == nil {
		t.Fatal("loadNumbers() succeeded, want error")
	}
}

func TestLoadNumbersPreservesNonDefaultSymbols(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "cldr-numbers-full", "main", "en", "numbers.json")
	mustWriteFile(t, path, numbersDocument(`{
		"defaultNumberingSystem": "latn",
		"symbols-numberSystem-latn": `+minimalNumberSymbolsJSON+`,
		"symbols-numberSystem-arab": {"decimal":"٫","group":"٬","nan":"local NaN","exponential":"local exponent","timeSeparator":"٫"},
		"miscPatterns-numberSystem-arab": {"range":"{0} ~ {1}"},
		"decimalFormats-numberSystem-latn": {"standard":"#,##0.###"},
		"percentFormats-numberSystem-latn": {"standard":"#,##0%"},
		"scientificFormats-numberSystem-latn": {"standard":"#E0"},
		"currencyFormats-numberSystem-latn": {"standard":"¤#,##0.00","unitPattern-count-other":"{0} {1}"}
	}`))

	data, err := loadNumbers(root, []string{"en"})
	if err != nil {
		t.Fatal(err)
	}
	numbers := data["en"]
	want := NumberSymbols{Decimal: "٫", Group: "٬", NaN: "local NaN", Exponential: "local exponent", TimeSeparator: "٫", RangeSign: " ~ "}
	if got := numbers.Symbols["arab"]; got != want {
		t.Fatalf("non-default symbols = %+v, want %+v", got, want)
	}
	if got := numbers.Symbols["latn"].Decimal; got != "." {
		t.Errorf("latn decimal = %q, want .", got)
	}
	if got := numbers.DecimalPatterns["latn"]; got != "#,##0.###" {
		t.Errorf("latn decimal pattern = %q, want #,##0.###", got)
	}
	if _, ok := numbers.DecimalPatterns["arab"]; ok {
		t.Fatal("symbols-only row acquired a decimal pattern")
	}
}

func TestLoadNumbersPreservesPinnedNonDefaultFormats(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", ".cldr-json", "node_modules")
	path := filepath.Join(root, "cldr-numbers-full", "main", "gu", "numbers.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("pinned CLDR checkout missing; run task data:fetch:cldr")
	}
	data, err := loadNumbers(root, []string{"gu"})
	if err != nil {
		t.Fatal(err)
	}
	numbers := data["gu"]
	for _, tc := range []struct {
		system, decimal, percent, scientific, currency string
	}{
		{"latn", "#,##,##0.###", "#,##,##0%", "[#E0]", "¤#,##,##0.00"},
		{"gujr", "#,##,##0.###", "#,##0%", "#E0", "¤#,##0.00"},
	} {
		for _, field := range []struct{ name, got, want string }{
			{"decimal", numbers.DecimalPatterns[tc.system], tc.decimal},
			{"percent", numbers.PercentPatterns[tc.system], tc.percent},
			{"scientific", numbers.ScientificPatterns[tc.system], tc.scientific},
			{"currency", numbers.CurrencyPatterns[tc.system]["standard"], tc.currency},
		} {
			if field.got != field.want {
				t.Errorf("gu/%s %s = %q, want pinned %q", tc.system, field.name, field.got, field.want)
			}
		}
	}
	for _, field := range []struct{ name, got, want string }{
		{"decimal compact", numbers.CompactPatterns["gujr"]["short"][4]["other"], "00\u00a0હજાર"},
		{"currency compact", numbers.CurrencyCompactPatterns["gujr"]["short"][4]["other"], "¤00\u00a0હજાર"},
		{"currency name", numbers.CurrencyNamePatterns["gujr"]["other"], "{0} {1}"},
	} {
		if field.got != field.want {
			t.Errorf("gu/gujr %s = %q, want pinned %q", field.name, field.got, field.want)
		}
	}
	if got, want := numbers.CurrencySpacing["gujr"], (CurrencySpacing{BeforeCurrency: "\u00a0", AfterCurrency: "\u00a0"}); got != want {
		t.Errorf("gu/gujr currency spacing = %+v, want pinned %+v", got, want)
	}
}

func TestLoadNumbersPreservesIndividualNonDefaultFormats(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ family, raw, pattern string }{
		{"decimalFormats", `{"standard":"#0.###","short":{"decimalFormat":{"10000-count-other":"00K"}}}`, "#0.###"},
		{"percentFormats", `{"standard":"#0%"}`, "#0%"},
		{"scientificFormats", `{"standard":"0E0"}`, "0E0"},
		{"currencyFormats", `{"standard":"¤#0.00","unitPattern-count-other":"{0} {1}","short":{"standard":{"10000-count-other":"¤00K"}},"currencySpacing":{"beforeCurrency":{"currencyMatch":"[[:^S:]&[:^Z:]]","surroundingMatch":"[:digit:]","insertBetween":" "}}}`, "¤#0.00"},
	} {
		t.Run(tc.family, func(t *testing.T) {
			t.Parallel()
			root := writeNonDefaultNumberFormat(t, tc.family, tc.raw)
			data, err := loadNumbers(root, []string{"en"})
			if err != nil {
				t.Fatal(err)
			}
			numbers := data["en"]
			for family, patterns := range map[string]map[string]string{
				"decimalFormats":    numbers.DecimalPatterns,
				"percentFormats":    numbers.PercentPatterns,
				"scientificFormats": numbers.ScientificPatterns,
				"currencyFormats":   {"latn": numbers.CurrencyPatterns["latn"]["standard"]},
			} {
				if family == "currencyFormats" && numbers.CurrencyPatterns["arab"] != nil {
					patterns["arab"] = numbers.CurrencyPatterns["arab"]["standard"]
				}
				got, present := patterns["arab"]
				if family == tc.family {
					if !present || got != tc.pattern {
						t.Errorf("%s arab = %q, present %t, want %q", family, got, present, tc.pattern)
					}
				} else if present {
					t.Errorf("absent %s acquired arab row %q", family, got)
				}
			}
			if tc.family == "decimalFormats" && numbers.CompactPatterns["arab"]["short"][4]["other"] != "00K" {
				t.Errorf("decimal compact = %v, want explicit 00K", numbers.CompactPatterns["arab"])
			}
			if tc.family == "currencyFormats" {
				if got := numbers.CurrencyCompactPatterns["arab"]["short"][4]["other"]; got != "¤00K" {
					t.Errorf("currency compact = %q, want explicit ¤00K", got)
				}
				if got := numbers.CurrencyNamePatterns["arab"]["other"]; got != "{0} {1}" {
					t.Errorf("currency name = %q, want explicit {0} {1}", got)
				}
				if got := numbers.CurrencySpacing["arab"].BeforeCurrency; got != " " {
					t.Errorf("currency spacing = %q, want explicit space", got)
				}
			}
		})
	}
}

func TestLoadNumbersRejectsMalformedNonDefaultFormats(t *testing.T) {
	t.Parallel()
	for _, family := range []string{"decimalFormats", "percentFormats", "scientificFormats", "currencyFormats"} {
		for _, raw := range []string{`null`, `12`, `"bad"`, `{}`, `{"standard":null}`, `{"standard":12}`} {
			t.Run(family+"/"+raw, func(t *testing.T) {
				t.Parallel()
				root := writeNonDefaultNumberFormat(t, family, raw)
				_, err := loadNumbers(root, []string{"en"})
				if err == nil {
					t.Fatal("loadNumbers accepted a malformed present family")
				}
				for _, context := range []string{"en", "arab", family} {
					if !strings.Contains(err.Error(), context) {
						t.Errorf("error %q missing %q", err, context)
					}
				}
			})
		}
	}
	for _, tc := range []struct{ family, raw, field string }{
		{"decimalFormats", `{"standard":"#0.###","short":{"decimalFormat":{"10000-count-other":"00K'"}}}`, "short"},
		{"currencyFormats", `{"standard":"¤#0.00","short":{"standard":{"10000-count-other":"¤00K'"}}}`, "short"},
		{"currencyFormats", `{"standard":"¤#0.00","unitPattern-count-other":"{0}"}`, "unitPattern-count-other"},
		{"currencyFormats", `{"standard":"¤#0.00","currencySpacing":{"beforeCurrency":{"currencyMatch":"[:letter:]"}}}`, "currencyMatch"},
	} {
		t.Run(tc.family+"/"+tc.field, func(t *testing.T) {
			t.Parallel()
			root := writeNonDefaultNumberFormat(t, tc.family, tc.raw)
			_, err := loadNumbers(root, []string{"en"})
			if err == nil {
				t.Fatal("loadNumbers accepted malformed supported subfields")
			}
			for _, context := range []string{"en", "arab", tc.family, tc.field} {
				if !strings.Contains(err.Error(), context) {
					t.Errorf("error %q missing %q", err, context)
				}
			}
		})
	}
}

func writeNonDefaultNumberFormat(t *testing.T, family, raw string) string {
	t.Helper()
	root := t.TempDir()
	fields := `{"defaultNumberingSystem":"latn","symbols-numberSystem-latn":` + minimalNumberSymbolsJSON + `,
		"decimalFormats-numberSystem-latn":{"standard":"#,##0.###"},
		"percentFormats-numberSystem-latn":{"standard":"#,##0%"},
		"scientificFormats-numberSystem-latn":{"standard":"#E0"},
		"currencyFormats-numberSystem-latn":{"standard":"¤#,##0.00","unitPattern-count-other":"{0} {1}"},` +
		strconv.Quote(family+"-numberSystem-arab") + `:` + raw + `}`
	mustWriteFile(t, filepath.Join(root, "cldr-numbers-full", "main", "en", "numbers.json"), numbersDocument(fields))
	return root
}

func TestLoadNumbersPreservesPinnedSymbolOverrides(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", ".cldr-json", "node_modules")
	if _, err := os.Stat(filepath.Join(root, "cldr-numbers-full", "main", "ar", "numbers.json")); os.IsNotExist(err) {
		t.Skip("pinned CLDR checkout missing; run task data:fetch:cldr")
	}
	data, err := loadNumbers(root, []string{"ar", "ur"})
	if err != nil {
		t.Fatal(err)
	}
	// CLDR 48.1.0 explicit rows, independent of the runtime projection.
	if got := data["ar"].Symbols["arab"]; got.NaN != "ليس\u00a0رقمًا" || got.Exponential != "أس" {
		t.Errorf("ar/arab override = %+v, want local NaN and exponent", got)
	}
	if got := data["ur"].Symbols["arabext"]; got.Decimal != "٫" || got.Group != "٬" || got.TimeSeparator != "٫" {
		t.Errorf("ur/arabext override = %+v, want U+066B/U+066C and U+066B time separator", got)
	}
}

func TestLoadNumbersRejectsMalformedNonDefaultSymbols(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`null`, `12`, `{"decimal":12}`, `{"group":","}`} {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			path := filepath.Join(root, "cldr-numbers-full", "main", "en", "numbers.json")
			mustWriteFile(t, path, numbersDocument(`{
				"defaultNumberingSystem":"latn",
				"symbols-numberSystem-latn":`+minimalNumberSymbolsJSON+`,
				"symbols-numberSystem-arab":`+raw+`,
				"decimalFormats-numberSystem-latn":{"standard":"#,##0.###"},
				"percentFormats-numberSystem-latn":{"standard":"#,##0%"},
				"scientificFormats-numberSystem-latn":{"standard":"#E0"},
				"currencyFormats-numberSystem-latn":{"standard":"¤#,##0.00","unitPattern-count-other":"{0} {1}"}
			}`))
			_, err := loadNumbers(root, []string{"en"})
			if err == nil {
				t.Fatal("loadNumbers ignored a malformed non-default symbols row")
			}
			for _, context := range []string{path, "symbols-numberSystem-arab"} {
				if !strings.Contains(err.Error(), context) {
					t.Errorf("error %q missing %q", err, context)
				}
			}
		})
	}
}

func TestLoadNumbersRejectsInvalidShape(t *testing.T) {
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
			name: "missing numbers",
			doc: `{
				"main": {
					"en": {}
				}
			}`,
		},
		{
			name: "null numbers",
			doc: `{
					"main": {
						"en": {
							"numbers": null
						}
					}
				}`,
		},
		{
			name: "empty numbers",
			doc: `{
					"main": {
						"en": {
							"numbers": {}
						}
					}
				}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-numbers-full", "main", "en", "numbers.json"), tc.doc)

			if _, err := loadNumbers(root, []string{"en"}); err == nil {
				t.Fatal("loadNumbers() succeeded, want error")
			}
		})
	}
}

func TestLoadNumbersRejectsMissingRequiredNumberSystemData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		doc  string
	}{
		{
			name: "missing default numbering system",
			doc: numbersDocument(`{
				"symbols-numberSystem-latn": ` + minimalNumberSymbolsJSON + `,
				"decimalFormats-numberSystem-latn": {"standard":"#,##0.###"},
				"percentFormats-numberSystem-latn": {"standard":"#,##0%"},
				"scientificFormats-numberSystem-latn": {"standard":"#E0"},
				"currencyFormats-numberSystem-latn": {"standard":"¤#,##0.00"}
			}`),
		},
		{
			name: "empty default numbering system",
			doc: numbersDocument(`{
				"defaultNumberingSystem": "",
				"symbols-numberSystem-latn": ` + minimalNumberSymbolsJSON + `,
				"decimalFormats-numberSystem-latn": {"standard":"#,##0.###"},
				"percentFormats-numberSystem-latn": {"standard":"#,##0%"},
				"scientificFormats-numberSystem-latn": {"standard":"#E0"},
				"currencyFormats-numberSystem-latn": {"standard":"¤#,##0.00"}
			}`),
		},
		{
			name: "missing symbols",
			doc: numbersDocument(`{
				"defaultNumberingSystem": "latn",
				"decimalFormats-numberSystem-latn": {"standard":"#,##0.###"},
				"percentFormats-numberSystem-latn": {"standard":"#,##0%"},
				"scientificFormats-numberSystem-latn": {"standard":"#E0"},
				"currencyFormats-numberSystem-latn": {"standard":"¤#,##0.00"}
			}`),
		},
		{
			name: "null symbols",
			doc: numbersDocument(`{
				"defaultNumberingSystem": "latn",
				"symbols-numberSystem-latn": null,
				"decimalFormats-numberSystem-latn": {"standard":"#,##0.###"},
				"percentFormats-numberSystem-latn": {"standard":"#,##0%"},
				"scientificFormats-numberSystem-latn": {"standard":"#E0"},
				"currencyFormats-numberSystem-latn": {"standard":"¤#,##0.00"}
			}`),
		},
		{
			name: "missing decimal",
			doc: numbersDocument(`{
				"defaultNumberingSystem": "latn",
				"symbols-numberSystem-latn": ` + minimalNumberSymbolsJSON + `,
				"percentFormats-numberSystem-latn": {"standard":"#,##0%"},
				"scientificFormats-numberSystem-latn": {"standard":"#E0"},
				"currencyFormats-numberSystem-latn": {"standard":"¤#,##0.00"}
			}`),
		},
		{
			name: "missing decimal standard",
			doc: numbersDocument(`{
				"defaultNumberingSystem": "latn",
				"symbols-numberSystem-latn": ` + minimalNumberSymbolsJSON + `,
				"decimalFormats-numberSystem-latn": {},
				"percentFormats-numberSystem-latn": {"standard":"#,##0%"},
				"scientificFormats-numberSystem-latn": {"standard":"#E0"},
				"currencyFormats-numberSystem-latn": {"standard":"¤#,##0.00"}
			}`),
		},
		{
			name: "missing percent",
			doc: numbersDocument(`{
				"defaultNumberingSystem": "latn",
				"symbols-numberSystem-latn": ` + minimalNumberSymbolsJSON + `,
				"decimalFormats-numberSystem-latn": {"standard":"#,##0.###"},
				"scientificFormats-numberSystem-latn": {"standard":"#E0"},
				"currencyFormats-numberSystem-latn": {"standard":"¤#,##0.00"}
			}`),
		},
		{
			name: "missing scientific",
			doc: numbersDocument(`{
				"defaultNumberingSystem": "latn",
				"symbols-numberSystem-latn": ` + minimalNumberSymbolsJSON + `,
				"decimalFormats-numberSystem-latn": {"standard":"#,##0.###"},
				"percentFormats-numberSystem-latn": {"standard":"#,##0%"},
				"currencyFormats-numberSystem-latn": {"standard":"¤#,##0.00"}
			}`),
		},
		{
			name: "missing currency",
			doc: numbersDocument(`{
				"defaultNumberingSystem": "latn",
				"symbols-numberSystem-latn": ` + minimalNumberSymbolsJSON + `,
				"decimalFormats-numberSystem-latn": {"standard":"#,##0.###"},
				"percentFormats-numberSystem-latn": {"standard":"#,##0%"},
				"scientificFormats-numberSystem-latn": {"standard":"#E0"}
			}`),
		},
		{
			name: "missing latn fallback for non-latn default",
			doc: numbersDocument(`{
				"defaultNumberingSystem": "arab",
				"symbols-numberSystem-arab": ` + minimalNumberSymbolsJSON + `,
				"decimalFormats-numberSystem-arab": {"standard":"#,##0.###"},
				"percentFormats-numberSystem-arab": {"standard":"#,##0%"},
				"scientificFormats-numberSystem-arab": {"standard":"#E0"},
				"currencyFormats-numberSystem-arab": {"standard":"¤#,##0.00"}
			}`),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-numbers-full", "main", "en", "numbers.json"), tc.doc)

			if _, err := loadNumbers(root, []string{"en"}); err == nil {
				t.Fatal("loadNumbers() succeeded, want error")
			}
		})
	}
}

func TestLoadNumbersPreservesCurrencyNamePlacementPatterns(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-numbers-full", "main", "sw", "numbers.json"), numbersDocumentForLocale("sw", `{
		"defaultNumberingSystem": "arab",
		"symbols-numberSystem-arab": `+minimalNumberSymbolsJSON+`,
		"decimalFormats-numberSystem-arab": {"standard":"#,##0.###"},
		"percentFormats-numberSystem-arab": {"standard":"#,##0%"},
		"scientificFormats-numberSystem-arab": {"standard":"#E0"},
		"currencyFormats-numberSystem-arab": {"standard":"¤#,##0.00","unitPattern-count-one":"{0} {1}","unitPattern-count-other":"{1} {0}"},
		"symbols-numberSystem-latn": `+minimalNumberSymbolsJSON+`,
		"decimalFormats-numberSystem-latn": {"standard":"#,##0.###"},
		"percentFormats-numberSystem-latn": {"standard":"#,##0%"},
		"scientificFormats-numberSystem-latn": {"standard":"#E0"},
		"currencyFormats-numberSystem-latn": {"standard":"¤#,##0.00","unitPattern-count-one":"{0} then {1}"}
	}`))

	got, err := loadNumbers(root, []string{"sw"})
	if err != nil {
		t.Fatalf("loadNumbers() error = %v", err)
	}
	want := map[string]map[string]string{
		"arab": {"one": "{0} {1}", "other": "{1} {0}"},
		"latn": {"one": "{0} then {1}"},
	}
	if !maps.EqualFunc(got["sw"].CurrencyNamePatterns, want, maps.Equal) {
		t.Errorf("loadNumbers() currency-name placement patterns = %#v, want %#v", got["sw"].CurrencyNamePatterns, want)
	}
}

func TestLoadNumbersRejectsInvalidCurrencyNamePlacementPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pattern string
		field   bool
	}{
		{name: "missing number placeholder", pattern: "{1}", field: true},
		{name: "missing currency placeholder", pattern: "{0}", field: true},
		{name: "duplicate placeholder", pattern: "{0} {1} {1}", field: true},
		{name: "missing default row"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			currencyField := `"currencyFormats-numberSystem-latn": {"standard":"¤#,##0.00"}`
			if tc.field {
				currencyField = `"currencyFormats-numberSystem-latn": {"standard":"¤#,##0.00","unitPattern-count-other":` + strconv.Quote(tc.pattern) + `}`
			}
			doc := numbersDocument(`{
				"defaultNumberingSystem": "latn",
				"symbols-numberSystem-latn": ` + minimalNumberSymbolsJSON + `,
				"decimalFormats-numberSystem-latn": {"standard":"#,##0.###"},
				"percentFormats-numberSystem-latn": {"standard":"#,##0%"},
				"scientificFormats-numberSystem-latn": {"standard":"#E0"},
				` + currencyField + `
			}`)
			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-numbers-full", "main", "en", "numbers.json"), doc)

			_, err := loadNumbers(root, []string{"en"})
			if err == nil {
				t.Fatal("loadNumbers() succeeded, want error")
			}
			for _, want := range []string{"en", "latn"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("loadNumbers() error = %q, want locale and numbering system context containing %q", err, want)
				}
			}
		})
	}
}

func TestLoadCurrencyFractions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "currencyData.json"), `{
		"supplemental": {
			"currencyData": {
				"fractions": {
					"DEFAULT": {"_digits": "2", "_cashDigits": "0", "_rounding": "0"},
					"JPY": {"_digits": "0", "_cashDigits": "0", "_rounding": "0"},
					"CHF": {"_digits": "2", "_cashDigits": "2", "_rounding": "5"}
				}
			}
		}
	}`)

	got, err := loadCurrencyFractions(root)
	if err != nil {
		t.Fatalf("loadCurrencyFractions() error = %v", err)
	}
	want := map[string]CurrencyFraction{
		"DEFAULT": {Digits: 2},
		"JPY":     {Digits: 0},
		"CHF":     {Digits: 2},
	}
	if !maps.Equal(got, want) {
		t.Fatalf("loadCurrencyFractions() = %#v, want %#v", got, want)
	}
}

func TestLoadCurrencyFractionsRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "currencyData.json"), `{`)

	if _, err := loadCurrencyFractions(root); err == nil {
		t.Fatal("loadCurrencyFractions() succeeded, want error")
	}
}

func TestLoadCurrencyFractionsRejectsInvalidShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		doc  string
	}{
		{
			name: "missing fractions",
			doc: `{
				"supplemental": {
					"currencyData": {}
				}
			}`,
		},
		{
			name: "null fractions",
			doc:  `{"supplemental":{"currencyData":{"fractions":null}}}`,
		},
		{
			name: "empty fractions",
			doc:  `{"supplemental":{"currencyData":{"fractions":{}}}}`,
		},
		{
			name: "missing default fraction",
			doc:  `{"supplemental":{"currencyData":{"fractions":{"JPY":{"_digits":"0","_cashDigits":"0","_rounding":"0"}}}}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "currencyData.json"), tc.doc)

			if _, err := loadCurrencyFractions(root); err == nil {
				t.Fatal("loadCurrencyFractions() succeeded, want error")
			}
		})
	}
}

func TestParseNumberSymbols(t *testing.T) {
	t.Parallel()

	got, err := parseNumberSymbols(jsontext.Value(`{
		"decimal": ".",
		"group": ",",
		"percentSign": "%",
		"plusSign": "+",
		"minusSign": "-",
		"nan": "NaN",
		"infinity": "∞",
		"approximatelySign": "~",
		"perMille": "‰",
		"exponential": "E",
		"superscriptingExponent": "×",
		"timeSeparator": ":"
	}`))
	if err != nil {
		t.Fatalf("parseNumberSymbols() error = %v", err)
	}
	want := NumberSymbols{
		Decimal:                ".",
		Group:                  ",",
		Percent:                "%",
		Plus:                   "+",
		Minus:                  "-",
		NaN:                    "NaN",
		Infinity:               "∞",
		ApproxSign:             "~",
		PerMille:               "‰",
		Exponential:            "E",
		SuperscriptingExponent: "×",
		TimeSeparator:          ":",
	}
	if got != want {
		t.Fatalf("parseNumberSymbols() = %#v, want %#v", got, want)
	}
}

func TestParseRangeSign(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "en dash", raw: `{"range":"{0}–{1}"}`, want: "–"},
		{name: "hyphen", raw: `{"range":"{0}-{1}"}`, want: "-"},
		{name: "wave dash", raw: `{"range":"{0}～{1}"}`, want: "～"},
		{name: "spaces", raw: `{"range":"{0} – {1}"}`, want: " – "},
		{name: "leading space", raw: `{"range":"{0} –{1}"}`, want: " –"},
		{name: "trailing space", raw: `{"range":"{0}– {1}"}`, want: "– "},
		{name: "multiple characters", raw: `{"range":"{0} ⇔ to {1}"}`, want: " ⇔ to "},
		{name: "missing field", raw: `{}`, want: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseRangeSign(jsontext.Value(tc.raw))
			if err != nil {
				t.Fatalf("parseRangeSign(%s) error = %v", tc.raw, err)
			}
			if got != tc.want {
				t.Fatalf("parseRangeSign(%s) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestParseRangeSignRejectsInvalidPattern(t *testing.T) {
	t.Parallel()

	for _, pattern := range []string{"", "{0}{1}", "{0}–", "–{1}", "{1}–{0}", "{0}{0}–{1}", "{0}–{1}{1}", "{0}–{2}{1}", "prefix{0}–{1}", "{0}–{1}suffix"} {
		t.Run(pattern, func(t *testing.T) {
			t.Parallel()
			_, err := parseRangeSign(jsontext.Value(`{"range":` + strconv.Quote(pattern) + `}`))
			if err == nil {
				t.Fatalf("parseRangeSign(%q) succeeded, want invalid pattern error", pattern)
			}
		})
	}
}

func TestParseNumberSymbolsMonetaryOverrides(t *testing.T) {
	t.Parallel()

	// Source values: pinned CLDR 48.1 de-AT/fr-CH symbols-numberSystem-latn.
	for _, tc := range []struct {
		name string
		raw  string
		want NumberSymbols
	}{
		{"de-AT", `{"decimal":",","group":"\u00a0","currencyGroup":"."}`, NumberSymbols{Decimal: ",", Group: "\u00a0", CurrencyGroup: "."}},
		{"fr-CH", `{"decimal":",","group":"\u202f","currencyDecimal":"."}`, NumberSymbols{Decimal: ",", Group: "\u202f", CurrencyDecimal: "."}},
		{"omitted", `{"decimal":".","group":","}`, NumberSymbols{Decimal: ".", Group: ","}},
		{"explicit empty", `{"currencyDecimal":"","currencyGroup":""}`, NumberSymbols{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseNumberSymbols(jsontext.Value(tc.raw))
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("parseNumberSymbols() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestParseStandard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "standard pattern", raw: `{"standard":"#,##0.###"}`, want: "#,##0.###"},
		{name: "missing standard", raw: `{"accounting":"¤#,##0.00"}`, want: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseStandard(jsontext.Value(tc.raw))
			if err != nil {
				t.Fatalf("parseStandard(%s) error = %v", tc.raw, err)
			}
			if got != tc.want {
				t.Fatalf("parseStandard(%s) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestParseCurrencyPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want parsedCurrencyPatterns
	}{
		{
			name: "standard and accounting",
			raw:  `{"standard":"¤#,##0.00","accounting":"¤#,##0.00;(¤#,##0.00)"}`,
			want: parsedCurrencyPatterns{
				sign: map[string]string{
					"standard":   "¤#,##0.00",
					"accounting": "¤#,##0.00;(¤#,##0.00)",
				},
			},
		},
		{
			name: "standard only",
			raw:  `{"standard":"¤#,##0.00"}`,
			want: parsedCurrencyPatterns{sign: map[string]string{"standard": "¤#,##0.00"}},
		},
		{
			name: "currency name placement",
			raw:  `{"standard":"¤#,##0.00","unitPattern-count-one":"{0} {1}","unitPattern-count-other":"{1} {0}"}`,
			want: parsedCurrencyPatterns{
				sign: map[string]string{"standard": "¤#,##0.00"},
				name: map[string]string{"one": "{0} {1}", "other": "{1} {0}"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseCurrencyPatterns(jsontext.Value(tc.raw))
			if err != nil {
				t.Fatalf("parseCurrencyPatterns(%s) error = %v", tc.raw, err)
			}
			assertStringMap(t, "parseCurrencyPatterns("+tc.name+").sign", got.sign, tc.want.sign)
			assertStringMap(t, "parseCurrencyPatterns("+tc.name+").name", got.name, tc.want.name)
		})
	}
}

func TestParseCurrencyPatternsRejectsInvalidPluralCategory(t *testing.T) {
	t.Parallel()

	_, err := parseCurrencyPatterns(jsontext.Value(`{"standard":"¤#,##0.00","unitPattern-count-invalid":"{0} {1}","unitPattern-count-other":"{0} {1}"}`))
	if err == nil {
		t.Fatal("parseCurrencyPatterns() succeeded, want invalid plural category error")
	}
}

func TestParseCompactPatterns(t *testing.T) {
	t.Parallel()

	got, err := parseCompactPatterns(jsontext.Value(`{
		"short": {
			"decimalFormat": {
				"1000-count-one": "0K",
				"1000-count-1": "0 thousand",
				"1000-count-other": "0K",
				"10000-count-other": "00K",
				"1000-count-zero": ""
			}
		},
		"long": {
			"decimalFormat": {
				"1000000-count-one": "0 million",
				"1000000-count-other": "0 million"
			}
		}
	}`))
	if err != nil {
		t.Fatalf("parseCompactPatterns() error = %v", err)
	}
	want := map[string]map[int]map[string]string{
		"short": {
			3: {"1": "0 thousand", "one": "0K", "other": "0K"},
			4: {"other": "00K"},
		},
		"long": {
			6: {"one": "0 million", "other": "0 million"},
		},
	}
	assertCompactPatterns(t, "parseCompactPatterns()", got, want)
}

func TestParseCompactPatternsRejectsInvalidKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
	}{
		{name: "missing count separator", key: "1000-other"},
		{name: "non power magnitude", key: "2000-count-other"},
		{name: "small magnitude", key: "1-count-other"},
		{name: "invalid count", key: "1000-count-invalid"},
		{name: "alt count", key: "1000-count-one-alt-alphaNextToNumber"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			raw := jsontext.Value(`{
				"short": {
					"decimalFormat": {
						"` + tc.key + `": "0K"
					}
				}
			}`)
			if _, err := parseCompactPatterns(raw); err == nil {
				t.Fatal("parseCompactPatterns() succeeded, want error")
			}
		})
	}
}

func TestNumberParsersRejectInvalidJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		parse func(jsontext.Value) error
	}{
		{
			name: "symbols",
			parse: func(raw jsontext.Value) error {
				_, err := parseNumberSymbols(raw)
				return err
			},
		},
		{
			name: "range sign",
			parse: func(raw jsontext.Value) error {
				_, err := parseRangeSign(raw)
				return err
			},
		},
		{
			name: "standard",
			parse: func(raw jsontext.Value) error {
				_, err := parseStandard(raw)
				return err
			},
		},
		{
			name: "currency patterns",
			parse: func(raw jsontext.Value) error {
				_, err := parseCurrencyPatterns(raw)
				return err
			},
		},
		{
			name: "compact patterns",
			parse: func(raw jsontext.Value) error {
				_, err := parseCompactPatterns(raw)
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if err := tc.parse(jsontext.Value(`{`)); err == nil {
				t.Fatal("parse succeeded, want error")
			}
		})
	}
}

const minimalNumberSymbolsJSON = `{
	"decimal": ".",
	"group": ",",
	"percentSign": "%",
	"plusSign": "+",
	"minusSign": "-",
	"nan": "NaN",
	"infinity": "∞"
}`

func numbersDocument(fields string) string {
	return numbersDocumentForLocale("en", fields)
}

func numbersDocumentForLocale(locale, fields string) string {
	return `{
		"main": {
			"` + locale + `": {
				"numbers": ` + fields + `
			}
		}
	}`
}

func assertCompactPatterns(t *testing.T, name string, got, want map[string]map[int]map[string]string) {
	t.Helper()

	if !maps.EqualFunc(got, want, func(gotPatterns, wantPatterns map[int]map[string]string) bool {
		return maps.EqualFunc(gotPatterns, wantPatterns, maps.Equal)
	}) {
		t.Fatalf("%s = %#v, want %#v", name, got, want)
	}
}

func mustWriteFile(t *testing.T, path, data string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatalf("create fixture dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(data), 0o666); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
}

func assertCurrenciesByLocale(t *testing.T, name string, got, want map[string]Currencies) {
	t.Helper()

	if !maps.EqualFunc(got, want, currenciesEqual) {
		t.Fatalf("%s = %#v, want %#v", name, got, want)
	}
}

func currenciesEqual(got, want Currencies) bool {
	return maps.EqualFunc(got, want, currencyNamesEqual)
}

func currencyNamesEqual(got, want CurrencyNames) bool {
	return got.Canonical == want.Canonical &&
		got.Symbol == want.Symbol &&
		got.Narrow == want.Narrow &&
		got.Pattern == want.Pattern && got.Decimal == want.Decimal && got.Group == want.Group &&
		maps.Equal(got.Display, want.Display)
}

func TestLoadMinimumGroupingDigits(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		raw  string
		want int
	}{{"", 1}, {`"1"`, 1}, {`"2"`, 2}, {`"3"`, 3}, {`"0"`, 0}, {`"bad"`, 0}, {`2`, 0}} {
		t.Run(tc.raw, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			field := ""
			if tc.raw != "" {
				field = `"minimumGroupingDigits":` + tc.raw + `,`
			}
			path := filepath.Join(root, "cldr-numbers-full", "main", "en", "numbers.json")
			mustWriteFile(t, path, numbersDocument(`{`+field+`"defaultNumberingSystem":"latn","symbols-numberSystem-latn":`+minimalNumberSymbolsJSON+`,"decimalFormats-numberSystem-latn":{"standard":"#,##0.###"},"percentFormats-numberSystem-latn":{"standard":"#,##0%"},"scientificFormats-numberSystem-latn":{"standard":"#E0"},"currencyFormats-numberSystem-latn":{"standard":"¤#,##0.00","unitPattern-count-other":"{0} {1}"}}`))
			got, err := loadNumbers(root, []string{"en"})
			if tc.want == 0 {
				if err == nil || !strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), "minimumGroupingDigits") {
					t.Fatalf("error = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got["en"].MinimumGroupingDigits != tc.want {
				t.Errorf("minimum = %d, want %d", got["en"].MinimumGroupingDigits, tc.want)
			}
		})
	}
}

func TestLoadNumbersRejectsUnclosedCompactQuote(t *testing.T) {
	t.Parallel()
	for _, pattern := range []string{"00K'", "'0;00K", "00K;'-00K"} {
		t.Run(pattern, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			path := filepath.Join(root, "cldr-numbers-full", "main", "en", "numbers.json")
			mustWriteFile(t, path, numbersDocument(`{"defaultNumberingSystem":"latn","symbols-numberSystem-latn":`+minimalNumberSymbolsJSON+`,"decimalFormats-numberSystem-latn":{"standard":"#,##0.###","short":{"decimalFormat":{"10000-count-other":`+strconv.Quote(pattern)+`}}},"percentFormats-numberSystem-latn":{"standard":"#,##0%"},"scientificFormats-numberSystem-latn":{"standard":"#E0"},"currencyFormats-numberSystem-latn":{"standard":"¤#,##0.00","unitPattern-count-other":"{0} {1}"}}`))
			_, err := loadNumbers(root, []string{"en"})
			if err == nil {
				t.Fatal("loadNumbers accepted an unclosed compact quote")
			}
			for _, context := range []string{path, "short", "10000-count-other", "unclosed quote"} {
				if !strings.Contains(err.Error(), context) {
					t.Errorf("error %q missing %q", err, context)
				}
			}
		})
	}
}

func TestLoadNumbersCurrencyCompactPatterns(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", ".cldr-json", "node_modules")
	path := filepath.Join(root, "cldr-numbers-full", "main", "en", "numbers.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("pinned CLDR checkout missing; run task data:fetch:cldr")
	}
	data, err := loadNumbers(root, []string{"en"})
	if err != nil {
		t.Fatal(err)
	}
	numbers := data["en"]
	for _, tc := range []struct{ count, want string }{
		{"one", "¤00K"},
		{"other", "¤00K"},
		{"one-alt-alphaNextToNumber", "¤\u00a000K"},
		{"other-alt-alphaNextToNumber", "¤\u00a000K"},
	} {
		if got := numbers.CurrencyCompactPatterns["latn"]["short"][4][tc.count]; got != tc.want {
			t.Errorf("currency short 10000-count-%s = %q, want %q", tc.count, got, tc.want)
		}
	}
	if got := numbers.CurrencyCompactPatterns["latn"]["long"]; len(got) != 0 {
		t.Errorf("currency long = %v, want absent", got)
	}
	if got := numbers.CompactPatterns["latn"]["short"][4]["other"]; got != "00K" {
		t.Errorf("decimal short = %q, want 00K", got)
	}
	if got := numbers.CurrencyNamePatterns["latn"]["other"]; got != "{0} {1}" {
		t.Errorf("currency name placement = %q, want {0} {1}", got)
	}
}

func TestLoadNumbersRejectsUnsupportedCurrencySpacing(t *testing.T) {
	t.Parallel()
	path := filepath.Join("..", ".cldr-json", "node_modules", "cldr-numbers-full", "main", "en", "numbers.json")
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Skip("pinned CLDR checkout missing; run task data:fetch:cldr")
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, old, replacement, field string }{
		{"currency class", `"[[:^S:]&[:^Z:]]"`, `"[:letter:]"`, "currencyMatch"},
		{"number class", `"[:digit:]"`, `"[:letter:]"`, "surroundingMatch"},
		{"insertion type", `"insertBetween": " "`, `"insertBetween": 12`, "insertBetween"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			doc := strings.ReplaceAll(string(raw), tc.old, tc.replacement)
			if doc == string(raw) {
				t.Fatalf("source does not contain %s", tc.old)
			}
			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-numbers-full", "main", "en", "numbers.json"), doc)
			_, err := loadNumbers(root, []string{"en"})
			if err == nil {
				t.Fatal("loadNumbers accepted unsupported currencySpacing")
			}
			for _, context := range []string{"en", "latn", "currencySpacing", tc.field} {
				if !strings.Contains(err.Error(), context) {
					t.Errorf("error %q missing context %q", err, context)
				}
			}
		})
	}
}

func TestLoadNumbersCurrencySpacing(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", ".cldr-json", "node_modules")
	path := filepath.Join(root, "cldr-numbers-full", "main", "en", "numbers.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("pinned CLDR checkout missing; run task data:fetch:cldr")
	}
	data, err := loadNumbers(root, []string{"en"})
	if err != nil {
		t.Fatal(err)
	}
	want := CurrencySpacing{BeforeCurrency: "\u00a0", AfterCurrency: "\u00a0"}
	if got := data["en"].CurrencySpacing["latn"]; got != want {
		t.Fatalf("pinned en spacing = %+v, want %+v", got, want)
	}
}

func TestCurrencySpacingOmission(t *testing.T) {
	t.Parallel()
	rule := `{"currencyMatch":"[[:^S:]&[:^Z:]]","surroundingMatch":"[:digit:]","insertBetween":" "}`
	for _, tc := range []struct {
		name, raw string
		want      CurrencySpacing
	}{
		{"absent", `{}`, CurrencySpacing{}},
		{"before only", `{"currencySpacing":{"beforeCurrency":` + rule + `}}`, CurrencySpacing{BeforeCurrency: " "}},
		{"after only", `{"currencySpacing":{"afterCurrency":` + rule + `}}`, CurrencySpacing{AfterCurrency: " "}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseCurrencyPatterns(jsontext.Value(tc.raw))
			if err != nil {
				t.Fatal(err)
			}
			if got.spacing != tc.want {
				t.Errorf("spacing = %+v, want %+v", got.spacing, tc.want)
			}
		})
	}
}

func TestLoadNumbersRejectsInvalidCurrencyCompactPattern(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ key, pattern string }{
		{"10000-count-other", "¤00K'"},
		{"10000-count-other-alt-alphaNextToNumber", "¤\u00a000K'"},
		{"10000-count-other-alt-unknown", "¤00K"},
	} {
		t.Run(tc.key, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			path := filepath.Join(root, "cldr-numbers-full", "main", "en", "numbers.json")
			mustWriteFile(t, path, numbersDocument(`{"defaultNumberingSystem":"latn","symbols-numberSystem-latn":`+minimalNumberSymbolsJSON+`,"decimalFormats-numberSystem-latn":{"standard":"#,##0.###"},"percentFormats-numberSystem-latn":{"standard":"#,##0%"},"scientificFormats-numberSystem-latn":{"standard":"#E0"},"currencyFormats-numberSystem-latn":{"standard":"¤#,##0.00","unitPattern-count-other":"{0} {1}","short":{"standard":{`+strconv.Quote(tc.key)+`:`+strconv.Quote(tc.pattern)+`}}}}`))
			_, err := loadNumbers(root, []string{"en"})
			if err == nil {
				t.Fatal("loadNumbers accepted an invalid currency compact pattern")
			}
			for _, context := range []string{path, "currencyFormats-numberSystem-latn", "short", tc.key} {
				if !strings.Contains(err.Error(), context) {
					t.Errorf("error %q missing %q", err, context)
				}
			}
		})
	}
}
