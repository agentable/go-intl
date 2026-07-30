package cldr

import (
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	pluralop "github.com/agentable/go-intl/internal/plural"
)

type Numbers struct {
	DefaultNumberingSystem string
	Symbols                map[string]NumberSymbols
	DecimalPatterns        map[string]string
	PercentPatterns        map[string]string
	ScientificPatterns     map[string]string
	CurrencyPatterns       map[string]map[string]string
	CurrencyNamePatterns   map[string]map[string]string
	CompactPatterns        map[string]map[string]map[int]map[string]string
}

type NumberSymbols struct {
	Decimal, Group, Percent, Plus, Minus, NaN, Infinity, ApproxSign, RangeSign, PerMille, Exponential, SuperscriptingExponent, TimeSeparator string
}

type Currencies map[string]CurrencyNames

type CurrencyNames struct {
	Display   map[string]string
	Canonical string
	Symbol    string
	Narrow    string
}

type CurrencyFraction struct {
	Digits     int
	CashDigits int
	Rounding   int
}

func loadNumbers(root string, locales []string) (map[string]Numbers, error) {
	out := make(map[string]Numbers)
	for _, locale := range locales {
		if locale == undefinedLocale {
			continue
		}
		path := filepath.Join(root, "cldr-numbers-full", "main", locale, "numbers.json")
		raw, ok, err := readOptionalFile(path)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		var doc map[string]map[string]map[string]json.RawMessage
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		main, ok := doc["main"]
		if !ok {
			return nil, fmt.Errorf("numbers body missing for %s", locale)
		}
		localeBody, ok := main[locale]
		if !ok {
			return nil, fmt.Errorf("numbers body missing for %s", locale)
		}
		body, ok := localeBody["numbers"]
		if !ok {
			return nil, fmt.Errorf("numbers body missing for %s", locale)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(body, &fields); err != nil {
			return nil, fmt.Errorf("parse numbers fields for %s: %w", locale, err)
		}
		if len(fields) == 0 {
			return nil, fmt.Errorf("numbers data missing for %s", locale)
		}
		num := Numbers{
			Symbols:              make(map[string]NumberSymbols),
			DecimalPatterns:      make(map[string]string),
			PercentPatterns:      make(map[string]string),
			ScientificPatterns:   make(map[string]string),
			CurrencyPatterns:     make(map[string]map[string]string),
			CurrencyNamePatterns: make(map[string]map[string]string),
			CompactPatterns:      make(map[string]map[string]map[int]map[string]string),
		}
		rawDefault, ok := fields["defaultNumberingSystem"]
		if !ok {
			return nil, fmt.Errorf("defaultNumberingSystem missing for %s", locale)
		}
		if err := json.Unmarshal(rawDefault, &num.DefaultNumberingSystem); err != nil {
			return nil, fmt.Errorf("parse %s defaultNumberingSystem: %w", path, err)
		}
		if num.DefaultNumberingSystem == "" {
			return nil, fmt.Errorf("defaultNumberingSystem missing for %s", locale)
		}
		for _, ns := range numberSystemLoadOrder(num.DefaultNumberingSystem) {
			if err := loadNumberSystemFields(path, locale, fields, ns, &num); err != nil {
				return nil, err
			}
		}
		out[locale] = num
	}
	return out, nil
}

func numberSystemLoadOrder(defaultNumberingSystem string) []string {
	if defaultNumberingSystem == "latn" {
		return []string{"latn"}
	}
	return []string{defaultNumberingSystem, "latn"}
}

func loadNumberSystemFields(path, locale string, fields map[string]json.RawMessage, ns string, num *Numbers) error {
	raw, err := requiredNumberSystemField(fields, locale, ns, "symbols")
	if err != nil {
		return err
	}
	symbols, err := parseNumberSymbols(raw)
	if err != nil {
		return fmt.Errorf("parse %s symbols-numberSystem-%s: %w", path, ns, err)
	}
	if symbols.Decimal == "" {
		return fmt.Errorf("symbols-numberSystem-%s decimal missing for %s", ns, locale)
	}
	if raw, ok := fields["miscPatterns-numberSystem-"+ns]; ok {
		rangeSign, err := parseRangeSign(raw)
		if err != nil {
			return fmt.Errorf("parse %s miscPatterns-numberSystem-%s: %w", path, ns, err)
		}
		symbols.RangeSign = rangeSign
	}
	if symbols.RangeSign == "" {
		symbols.RangeSign = "–"
	}
	num.Symbols[ns] = symbols

	decimal, err := requiredStandardNumberPattern(path, locale, fields, ns, "decimalFormats")
	if err != nil {
		return err
	}
	num.DecimalPatterns[ns] = decimal
	raw = fields["decimalFormats-numberSystem-"+ns]
	patterns, err := parseCompactPatterns(raw)
	if err != nil {
		return fmt.Errorf("parse %s decimalFormats-numberSystem-%s compact: %w", path, ns, err)
	}
	if len(patterns) > 0 {
		num.CompactPatterns[ns] = patterns
	}

	percent, err := requiredStandardNumberPattern(path, locale, fields, ns, "percentFormats")
	if err != nil {
		return err
	}
	num.PercentPatterns[ns] = percent

	scientific, err := requiredStandardNumberPattern(path, locale, fields, ns, "scientificFormats")
	if err != nil {
		return err
	}
	num.ScientificPatterns[ns] = scientific

	raw, err = requiredNumberSystemField(fields, locale, ns, "currencyFormats")
	if err != nil {
		return err
	}
	currency, err := parseCurrencyPatterns(raw)
	if err != nil {
		return fmt.Errorf("parse %s currencyFormats-numberSystem-%s: %w", path, ns, err)
	}
	if currency.sign["standard"] == "" {
		return fmt.Errorf("currencyFormats-numberSystem-%s standard pattern missing for %s", ns, locale)
	}
	if currency.name[pluralop.Other.String()] == "" && ns == num.DefaultNumberingSystem {
		return fmt.Errorf("currencyFormats-numberSystem-%s unitPattern-count-other missing for %s", ns, locale)
	}
	for _, plural := range slices.Sorted(maps.Keys(currency.name)) {
		pattern := currency.name[plural]
		field := "unitPattern-count-" + plural
		if err := validatePatternPlaceholders(field, pattern, "{0}", "{1}"); err != nil {
			return fmt.Errorf("parse %s currencyFormats-numberSystem-%s: %w", path, ns, err)
		}
	}
	if len(currency.name) > 0 {
		num.CurrencyNamePatterns[ns] = currency.name
	}
	num.CurrencyPatterns[ns] = currency.sign
	return nil
}

func requiredStandardNumberPattern(path, locale string, fields map[string]json.RawMessage, ns, prefix string) (string, error) {
	raw, err := requiredNumberSystemField(fields, locale, ns, prefix)
	if err != nil {
		return "", err
	}
	pattern, err := parseStandard(raw)
	if err != nil {
		return "", fmt.Errorf("parse %s %s-numberSystem-%s: %w", path, prefix, ns, err)
	}
	if pattern == "" {
		return "", fmt.Errorf("%s-numberSystem-%s standard pattern missing for %s", prefix, ns, locale)
	}
	return pattern, nil
}

func requiredNumberSystemField(fields map[string]json.RawMessage, locale, ns, prefix string) (json.RawMessage, error) {
	key := prefix + "-numberSystem-" + ns
	raw, ok := fields[key]
	if !ok || len(raw) == 0 {
		return nil, fmt.Errorf("%s missing for %s", key, locale)
	}
	return raw, nil
}

func parseNumberSymbols(raw json.RawMessage) (NumberSymbols, error) {
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
		TimeSeparator          string `json:"timeSeparator"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return NumberSymbols{}, err
	}
	return NumberSymbols{
		Decimal:                doc.Decimal,
		Group:                  doc.Group,
		Percent:                doc.Percent,
		Plus:                   doc.Plus,
		Minus:                  doc.Minus,
		NaN:                    doc.NaN,
		Infinity:               doc.Infinity,
		ApproxSign:             doc.ApproxSign,
		PerMille:               doc.PerMille,
		Exponential:            doc.Exponential,
		SuperscriptingExponent: doc.SuperscriptingExponent,
		TimeSeparator:          doc.TimeSeparator,
	}, nil
}

func parseRangeSign(raw json.RawMessage) (string, error) {
	var doc struct {
		Range string `json:"range"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return "", err
	}
	rest := strings.NewReplacer("{0}", "", "{1}", "").Replace(doc.Range)
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return "", nil
	}
	r, _ := utf8.DecodeRuneInString(rest)
	if r == utf8.RuneError {
		return "", nil
	}
	return string(r), nil
}

func parseStandard(raw json.RawMessage) (string, error) {
	var doc struct {
		Standard string `json:"standard"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return "", err
	}
	return doc.Standard, nil
}

type parsedCurrencyPatterns struct {
	sign map[string]string
	name map[string]string
}

func parseCurrencyPatterns(raw json.RawMessage) (parsedCurrencyPatterns, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return parsedCurrencyPatterns{}, err
	}
	out := parsedCurrencyPatterns{
		sign: make(map[string]string),
		name: make(map[string]string),
	}
	for _, key := range slices.Sorted(maps.Keys(fields)) {
		var value string
		switch key {
		case "standard", "accounting":
			if err := json.Unmarshal(fields[key], &value); err != nil {
				return parsedCurrencyPatterns{}, fmt.Errorf("parse %s: %w", key, err)
			}
			if value != "" {
				out.sign[key] = value
			}
		default:
			plural, ok, err := pluralCategoryFromField(key, "unitPattern-count-")
			if err != nil {
				return parsedCurrencyPatterns{}, fmt.Errorf("parse %s: %w", key, err)
			}
			if !ok {
				continue
			}
			if err := json.Unmarshal(fields[key], &value); err != nil {
				return parsedCurrencyPatterns{}, fmt.Errorf("parse %s: %w", key, err)
			}
			out.name[plural] = value
		}
	}
	return out, nil
}

func parseCompactPatterns(raw json.RawMessage) (map[string]map[int]map[string]string, error) {
	var doc struct {
		Short compactFormat `json:"short"`
		Long  compactFormat `json:"long"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	out := make(map[string]map[int]map[string]string)
	patterns, err := parseCompactDisplayPatterns(doc.Short.DecimalFormat)
	if err != nil {
		return nil, fmt.Errorf("short decimalFormat: %w", err)
	}
	if len(patterns) > 0 {
		out["short"] = patterns
	}
	patterns, err = parseCompactDisplayPatterns(doc.Long.DecimalFormat)
	if err != nil {
		return nil, fmt.Errorf("long decimalFormat: %w", err)
	}
	if len(patterns) > 0 {
		out["long"] = patterns
	}
	return out, nil
}

type compactFormat struct {
	DecimalFormat map[string]string `json:"decimalFormat"`
}

func parseCompactDisplayPatterns(raw map[string]string) (map[int]map[string]string, error) {
	out := make(map[int]map[string]string)
	for _, key := range slices.Sorted(maps.Keys(raw)) {
		exp, count, err := compactPatternKey(key)
		if err != nil {
			return nil, err
		}
		pattern := raw[key]
		if pattern == "" {
			continue
		}
		if out[exp] == nil {
			out[exp] = make(map[string]string)
		}
		out[exp][count] = pattern
	}
	return out, nil
}

func compactPatternKey(key string) (int, string, error) {
	magnitude, count, ok := strings.Cut(key, "-count-")
	if !ok {
		return 0, "", fmt.Errorf("invalid compact pattern key %q: expected <magnitude>-count-<count>", key)
	}
	if !validCompactMagnitude(magnitude) {
		return 0, "", fmt.Errorf("invalid compact pattern magnitude %q", magnitude)
	}
	if !validCompactCount(count) {
		return 0, "", fmt.Errorf("invalid compact pattern count %q", count)
	}
	return len(magnitude) - 1, count, nil
}

func validCompactMagnitude(magnitude string) bool {
	if len(magnitude) < 4 || magnitude[0] != '1' {
		return false
	}
	for _, digit := range magnitude[1:] {
		if digit != '0' {
			return false
		}
	}
	return true
}

func validCompactCount(count string) bool {
	if _, ok := pluralop.ParseCategory(count); ok {
		return true
	}
	return asciiDigits(count)
}

func asciiDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func loadCurrencies(root string, locales []string) (map[string]Currencies, error) {
	out := make(map[string]Currencies)
	for _, locale := range locales {
		if locale == undefinedLocale {
			continue
		}
		path := filepath.Join(root, "cldr-numbers-full", "main", locale, "currencies.json")
		raw, ok, err := readOptionalFile(path)
		if err != nil {
			return nil, err
		}
		if !ok {
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
		main, ok := doc["main"]
		if !ok {
			return nil, fmt.Errorf("currencies body missing for %s", locale)
		}
		body, ok := main[locale]
		if !ok {
			return nil, fmt.Errorf("currencies body missing for %s", locale)
		}
		if body.Numbers.Currencies == nil {
			return nil, fmt.Errorf("currencies data missing for %s", locale)
		}
		cur := make(Currencies)
		for _, code := range slices.Sorted(maps.Keys(body.Numbers.Currencies)) {
			fields := body.Numbers.Currencies[code]
			names, err := parseCurrencyNames(fields)
			if err != nil {
				return nil, fmt.Errorf("parse %s currency %s: %w", path, code, err)
			}
			cur[code] = names
		}
		out[locale] = cur
	}
	return out, nil
}

func parseCurrencyNames(fields map[string]string) (CurrencyNames, error) {
	names := CurrencyNames{Display: make(map[string]string)}
	for _, key := range slices.Sorted(maps.Keys(fields)) {
		value := fields[key]
		switch key {
		case "displayName":
			names.Canonical = value
			other := pluralop.Other.String()
			if _, ok := names.Display[other]; !ok {
				names.Display[other] = value
			}
		case "symbol":
			names.Symbol = value
		case "symbol-alt-narrow":
			names.Narrow = value
		default:
			plural, ok, err := pluralCategoryFromField(key, "displayName-count-")
			if err != nil {
				return CurrencyNames{}, fmt.Errorf("parse %s: %w", key, err)
			}
			if ok {
				names.Display[plural] = value
			}
		}
	}
	return names, nil
}

func loadCurrencyFractions(root string) (map[string]CurrencyFraction, error) {
	path := filepath.Join(root, "cldr-core", "supplemental", "currencyData.json")
	raw, err := readRequiredFile(path)
	if err != nil {
		return nil, err
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
	if len(doc.Supplemental.CurrencyData.Fractions) == 0 {
		return nil, fmt.Errorf("expected supplemental currencyData fractions map")
	}
	if _, ok := doc.Supplemental.CurrencyData.Fractions["DEFAULT"]; !ok {
		return nil, fmt.Errorf("expected supplemental currencyData fractions DEFAULT row")
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
