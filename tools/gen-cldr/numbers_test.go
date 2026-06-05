package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGeneratesNumbersAndCurrencies(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o777); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	versionPath := filepath.Join(out, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := Run(context.Background(), Config{CLDRDir: root, OutDir: out, VersionFile: versionPath, ProfileFile: writeLocaleProfileFixture(t, dir)}, log); err != nil {
		t.Fatalf("Run: %v", err)
	}

	for _, name := range []string{filepath.Join("number", "data.go"), filepath.Join("currency", "data.go")} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatalf("expected generated %s: %v", name, err)
		}
	}
	// Number data is now a const-only domain payload: a private _data string
	// table plus the main and supported blobs. The patterns live in that
	// domain-private _data, not the shared strings.go table.
	number, err := os.ReadFile(filepath.Join(out, "number", "data.go"))
	if err != nil {
		t.Fatalf("read number/data.go: %v", err)
	}
	if !containsAll(string(number), "package number", "const _data", "_numberBlob", "_numberSupportedBlob") {
		t.Fatalf("number/data.go missing expected generated content:\n%s", number)
	}
	// Search the decoded const payload so a chunk boundary cannot split a pattern.
	numberData := readGeneratedStringTable(t, filepath.Join(out, "number", "data.go"))
	if !containsAll(numberData, "#,##0.###", "¤#,##0.00", "¤#,##0.00;(¤#,##0.00)", "0K", "0 thousand") {
		t.Fatalf("number/data.go missing expected number pattern strings:\n%s", numberData)
	}
	// Currency data is now a const-only domain payload: a private _data string
	// table plus the fraction/names/supported blobs. The display names live in
	// that domain-private _data, not the shared strings.go table.
	currency, err := os.ReadFile(filepath.Join(out, "currency", "data.go"))
	if err != nil {
		t.Fatalf("read currency/data.go: %v", err)
	}
	if !containsAll(string(currency), "package currency", "const _data", "_currencyFractionBlob", "_currencyNamesBlob", "_currencySupportedBlob") {
		t.Fatalf("currency/data.go missing expected generated content:\n%s", currency)
	}
	// Search the decoded const payload so a chunk boundary cannot split a name.
	currencyData := readGeneratedStringTable(t, filepath.Join(out, "currency", "data.go"))
	if !containsAll(currencyData, "US dollar", "Japanese yen") {
		t.Fatalf("currency/data.go missing expected currency display names:\n%s", currencyData)
	}
}

func writeNumberCLDRFixture(t *testing.T, root string) {
	t.Helper()
	locales := []string{"en", "en-US", "zh", "zh-Hans", "zh-Hans-CN"}
	for _, loc := range locales {
		dir := filepath.Join(root, "cldr-numbers-full", "main", loc)
		if err := os.MkdirAll(dir, 0o777); err != nil {
			t.Fatalf("mkdir numbers %s: %v", loc, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "numbers.json"), []byte(numbersJSON(loc)), 0o666); err != nil {
			t.Fatalf("write numbers %s: %v", loc, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "currencies.json"), []byte(currenciesJSON(loc)), 0o666); err != nil {
			t.Fatalf("write currencies %s: %v", loc, err)
		}
	}
	supp := filepath.Join(root, "cldr-core", "supplemental")
	currencyData := `{"supplemental":{"currencyData":{"fractions":{"DEFAULT":{"_digits":"2","_rounding":"0"},"JPY":{"_digits":"0","_rounding":"0"},"USD":{"_digits":"2","_rounding":"0"}}}}}`
	if err := os.WriteFile(filepath.Join(supp, "currencyData.json"), []byte(currencyData), 0o666); err != nil {
		t.Fatalf("write currencyData: %v", err)
	}
}

func numbersJSON(locale string) string {
	return strings.ReplaceAll(`{
		"main":{"LOCALE":{"numbers":{
			"defaultNumberingSystem":"latn",
			"symbols-numberSystem-latn":{"decimal":".","group":",","list":";","percentSign":"%","plusSign":"+","minusSign":"-","approximatelySign":"~","exponential":"E","superscriptingExponent":"×","perMille":"‰","infinity":"∞","nan":"NaN","timeSeparator":":"},
			"miscPatterns-numberSystem-latn":{"range":"{0}–{1}"},
			"decimalFormats-numberSystem-latn":{"standard":"#,##0.###","long":{"decimalFormat":{"1000-count-one":"0 thousand","1000-count-other":"0 thousand"}},"short":{"decimalFormat":{"1000-count-one":"0K","1000-count-other":"0K"}}},
			"percentFormats-numberSystem-latn":{"standard":"#,##0%"},
			"scientificFormats-numberSystem-latn":{"standard":"#E0"},
			"currencyFormats-numberSystem-latn":{"standard":"¤#,##0.00","accounting":"¤#,##0.00;(¤#,##0.00)","unitPattern-count-other":"{0} {1}","currencySpacing":{"beforeCurrency":{"insertBetween":" "},"afterCurrency":{"insertBetween":" "}},"short":{"standard":{"1000-count-one":"¤0K","1000-count-other":"¤0K"}}}
		}}}
	}`, "LOCALE", locale)
}

func currenciesJSON(locale string) string {
	usd := "US dollar"
	jpy := "Japanese yen"
	if strings.HasPrefix(locale, "zh") {
		usd = "美元"
		jpy = "日元"
	}
	return `{"main":{"` + locale + `":{"numbers":{"currencies":{"USD":{"displayName":"` + usd + `","displayName-count-one":"` + usd + `","displayName-count-other":"` + usd + `","symbol":"$","symbol-alt-narrow":"$"},"JPY":{"displayName":"` + jpy + `","displayName-count-other":"` + jpy + `","symbol":"¥","symbol-alt-narrow":"¥"}}}}}}`
}
