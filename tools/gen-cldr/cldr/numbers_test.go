package cldr

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
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
							"symbol": "€"
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
							"symbol-alt-narrow": "$"
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
		"DEFAULT": {Digits: 2, CashDigits: 2, Rounding: 0},
		"JPY":     {Digits: 0, CashDigits: 0, Rounding: 0},
		"CHF":     {Digits: 2, CashDigits: 2, Rounding: 5},
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

	got, err := parseNumberSymbols(json.RawMessage(`{
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
		{name: "spaces are pattern glue", raw: `{"range":"{0} – {1}"}`, want: "–"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseRangeSign(json.RawMessage(tc.raw))
			if err != nil {
				t.Fatalf("parseRangeSign(%s) error = %v", tc.raw, err)
			}
			if got != tc.want {
				t.Fatalf("parseRangeSign(%s) = %q, want %q", tc.raw, got, tc.want)
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

			got, err := parseStandard(json.RawMessage(tc.raw))
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
		want map[string]string
	}{
		{
			name: "standard and accounting",
			raw:  `{"standard":"¤#,##0.00","accounting":"¤#,##0.00;(¤#,##0.00)"}`,
			want: map[string]string{
				"standard":   "¤#,##0.00",
				"accounting": "¤#,##0.00;(¤#,##0.00)",
			},
		},
		{
			name: "standard only",
			raw:  `{"standard":"¤#,##0.00"}`,
			want: map[string]string{"standard": "¤#,##0.00"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseCurrencyPatterns(json.RawMessage(tc.raw))
			if err != nil {
				t.Fatalf("parseCurrencyPatterns(%s) error = %v", tc.raw, err)
			}
			assertStringMap(t, "parseCurrencyPatterns("+tc.name+")", got, tc.want)
		})
	}
}

func TestParseCompactPatterns(t *testing.T) {
	t.Parallel()

	got, err := parseCompactPatterns(json.RawMessage(`{
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

			raw := json.RawMessage(`{
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
		parse func(json.RawMessage) error
	}{
		{
			name: "symbols",
			parse: func(raw json.RawMessage) error {
				_, err := parseNumberSymbols(raw)
				return err
			},
		},
		{
			name: "range sign",
			parse: func(raw json.RawMessage) error {
				_, err := parseRangeSign(raw)
				return err
			},
		},
		{
			name: "standard",
			parse: func(raw json.RawMessage) error {
				_, err := parseStandard(raw)
				return err
			},
		},
		{
			name: "currency patterns",
			parse: func(raw json.RawMessage) error {
				_, err := parseCurrencyPatterns(raw)
				return err
			},
		},
		{
			name: "compact patterns",
			parse: func(raw json.RawMessage) error {
				_, err := parseCompactPatterns(raw)
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if err := tc.parse(json.RawMessage(`{`)); err == nil {
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
	return `{
		"main": {
			"en": {
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
		maps.Equal(got.Display, want.Display)
}
