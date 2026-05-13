package cldr

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type Numbers struct {
	DefaultNumberingSystem string
	Symbols                map[string]NumberSymbols
	DecimalPatterns        map[string]string
	PercentPatterns        map[string]string
	ScientificPatterns     map[string]string
	CurrencyPatterns       map[string]map[string]string
	CompactPatterns        map[string]map[string]map[int]map[string]string
}

type NumberSymbols struct {
	Decimal, Group, Percent, Plus, Minus, NaN, Infinity, ApproxSign, PerMille, Exponential, SuperscriptingExponent string
}

type Currencies map[string]CurrencyNames

type CurrencyNames struct {
	Display map[string]string
	Symbol  string
	Narrow  string
}

type CurrencyFraction struct {
	Digits     int
	CashDigits int
	Rounding   int
}

func loadNumbers(root string, locales []string) (map[string]Numbers, error) {
	out := make(map[string]Numbers)
	for _, locale := range locales {
		if locale == "und" {
			continue
		}
		path := filepath.Join(root, "cldr-numbers-full", "main", locale, "numbers.json")
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var doc map[string]map[string]map[string]json.RawMessage
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		body, ok := doc["main"][locale]["numbers"]
		if !ok {
			return nil, fmt.Errorf("numbers body missing for %s", locale)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(body, &fields); err != nil {
			return nil, fmt.Errorf("parse numbers fields for %s: %w", locale, err)
		}
		num := Numbers{
			Symbols:            make(map[string]NumberSymbols),
			DecimalPatterns:    make(map[string]string),
			PercentPatterns:    make(map[string]string),
			ScientificPatterns: make(map[string]string),
			CurrencyPatterns:   make(map[string]map[string]string),
			CompactPatterns:    make(map[string]map[string]map[int]map[string]string),
		}
		json.Unmarshal(fields["defaultNumberingSystem"], &num.DefaultNumberingSystem)
		if num.DefaultNumberingSystem == "" {
			num.DefaultNumberingSystem = "latn"
		}
		for _, ns := range []string{num.DefaultNumberingSystem, "latn"} {
			if raw, ok := fields["symbols-numberSystem-"+ns]; ok {
				num.Symbols[ns] = parseNumberSymbols(raw)
			}
			if raw, ok := fields["decimalFormats-numberSystem-"+ns]; ok {
				num.DecimalPatterns[ns] = parseStandard(raw)
				if patterns := parseCompactPatterns(raw); len(patterns) > 0 {
					num.CompactPatterns[ns] = patterns
				}
			}
			if raw, ok := fields["percentFormats-numberSystem-"+ns]; ok {
				num.PercentPatterns[ns] = parseStandard(raw)
			}
			if raw, ok := fields["scientificFormats-numberSystem-"+ns]; ok {
				num.ScientificPatterns[ns] = parseStandard(raw)
			}
			if raw, ok := fields["currencyFormats-numberSystem-"+ns]; ok {
				num.CurrencyPatterns[ns] = parseCurrencyPatterns(raw)
			}
		}
		out[locale] = num
	}
	return out, nil
}

func parseNumberSymbols(raw json.RawMessage) NumberSymbols {
	var doc struct {
		Decimal                string `json:"decimal"`
		Group                  string `json:"group"`
		Percent                string `json:"percentSign"`
		Plus                   string `json:"plusSign"`
		Minus                  string `json:"minusSign"`
		NaN                    string `json:"nan"`
		Infinity               string `json:"infinity"`
		ApproxSign             string `json:"approximatelySign"`
		PerMille               string `json:"perMille"`
		Exponential            string `json:"exponential"`
		SuperscriptingExponent string `json:"superscriptingExponent"`
	}
	json.Unmarshal(raw, &doc)
	return NumberSymbols(doc)
}

func parseStandard(raw json.RawMessage) string {
	var doc struct {
		Standard string `json:"standard"`
	}
	json.Unmarshal(raw, &doc)
	return doc.Standard
}

func parseCurrencyPatterns(raw json.RawMessage) map[string]string {
	var doc struct {
		Standard   string `json:"standard"`
		Accounting string `json:"accounting"`
	}
	json.Unmarshal(raw, &doc)
	out := map[string]string{"standard": doc.Standard}
	if doc.Accounting != "" {
		out["accounting"] = doc.Accounting
	}
	return out
}

func parseCompactPatterns(raw json.RawMessage) map[string]map[int]map[string]string {
	var doc struct {
		Short compactFormat `json:"short"`
		Long  compactFormat `json:"long"`
	}
	json.Unmarshal(raw, &doc)
	out := make(map[string]map[int]map[string]string)
	if patterns := parseCompactDisplayPatterns(doc.Short.DecimalFormat); len(patterns) > 0 {
		out["short"] = patterns
	}
	if patterns := parseCompactDisplayPatterns(doc.Long.DecimalFormat); len(patterns) > 0 {
		out["long"] = patterns
	}
	return out
}

type compactFormat struct {
	DecimalFormat map[string]string `json:"decimalFormat"`
}

func parseCompactDisplayPatterns(raw map[string]string) map[int]map[string]string {
	out := make(map[int]map[string]string)
	for _, key := range slices.Sorted(maps.Keys(raw)) {
		pattern := raw[key]
		magnitude, plural, ok := strings.Cut(key, "-count-")
		if !ok || pattern == "" {
			continue
		}
		exp := len(magnitude) - 1
		if exp <= 0 {
			continue
		}
		if _, err := strconv.Atoi(magnitude); err != nil {
			continue
		}
		if out[exp] == nil {
			out[exp] = make(map[string]string)
		}
		out[exp][plural] = pattern
	}
	return out
}

func loadCurrencies(root string, locales []string) (map[string]Currencies, error) {
	out := make(map[string]Currencies)
	for _, locale := range locales {
		if locale == "und" {
			continue
		}
		path := filepath.Join(root, "cldr-numbers-full", "main", locale, "currencies.json")
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var doc map[string]map[string]struct {
			Numbers struct {
				Currencies map[string]map[string]string `json:"currencies"`
			} `json:"numbers"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		cur := make(Currencies)
		for _, code := range slices.Sorted(maps.Keys(doc["main"][locale].Numbers.Currencies)) {
			fields := doc["main"][locale].Numbers.Currencies[code]
			names := CurrencyNames{Display: make(map[string]string)}
			for _, key := range slices.Sorted(maps.Keys(fields)) {
				value := fields[key]
				switch key {
				case "displayName":
					names.Display["other"] = value
				case "displayName-count-zero":
					names.Display["zero"] = value
				case "displayName-count-one":
					names.Display["one"] = value
				case "displayName-count-two":
					names.Display["two"] = value
				case "displayName-count-few":
					names.Display["few"] = value
				case "displayName-count-many":
					names.Display["many"] = value
				case "displayName-count-other":
					names.Display["other"] = value
				case "symbol":
					names.Symbol = value
				case "symbol-alt-narrow":
					names.Narrow = value
				}
			}
			cur[code] = names
		}
		out[locale] = cur
	}
	return out, nil
}

func loadCurrencyFractions(root string) (map[string]CurrencyFraction, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "currencyData.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read currencyData.json: %w", err)
	}
	var doc struct {
		Supplemental struct {
			CurrencyData struct {
				Fractions map[string]struct {
					Digits     int `json:"_digits,string"`
					CashDigits int `json:"_cashDigits,string"`
					Rounding   int `json:"_rounding,string"`
				} `json:"fractions"`
			} `json:"currencyData"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse currencyData.json: %w", err)
	}
	out := make(map[string]CurrencyFraction, len(doc.Supplemental.CurrencyData.Fractions))
	for _, code := range slices.Sorted(maps.Keys(doc.Supplemental.CurrencyData.Fractions)) {
		f := doc.Supplemental.CurrencyData.Fractions[code]
		cash := f.CashDigits
		if cash == 0 {
			cash = f.Digits
		}
		out[code] = CurrencyFraction{Digits: f.Digits, CashDigits: cash, Rounding: f.Rounding}
	}
	return out, nil
}
