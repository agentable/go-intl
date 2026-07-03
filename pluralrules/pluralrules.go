package pluralrules

import (
	"strings"
	"sync"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/cldr/plural"
	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
	ecma402pr "github.com/agentable/go-intl/internal/ecma402/pluralrules"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

// Category is one ECMA-402 plural category returned by Select and SelectRange.
type Category uint8

const (
	Zero Category = iota
	One
	Two
	Few
	Many
	Other
)

func (c Category) String() string {
	return ecma402pr.Category(c).String()
}

func (c Category) MarshalText() ([]byte, error) {
	return ecma402pr.Category(c).MarshalText()
}

type PluralRules struct {
	dataLocale      string
	digitOptions    ecma402nf.DigitOptions
	compact         compactExponentSet
	integerOperands bool
	rule            pluralRuleFunc
	resolved        ResolvedOptions
}

type pluralRuleFunc func(ecma402pr.OperandsRecord) ecma402pr.Category

var pluralRulesLocaleMatcher = sync.OnceValue(func() *localematcher.Matcher {
	return localematcher.NewMatcher(plural.SupportedLocales(), cldrlocale.Maximize)
})

func New(locales locale.List, opts Options) (*PluralRules, error) {
	cfg := configFromOptions(opts)
	validationLocale := ecma402.ValidationLocale(locales)
	validationLocaleName := validationLocale.String()
	if err := cfg.validate(validationLocaleName); err != nil {
		return nil, err
	}
	resolvedDigits, invalid, ok := ecma402nf.SetNumberFormatDigitOptions(cfg.digits, 0, 3, cfg.notation)
	if ok {
		return nil, ecma402nf.InvalidDigitOptionError(pluralRulesOwner, invalid, validationLocaleName)
	}
	resolution := ecma402.ResolveConstructorLocale(ecma402.ConstructorLocaleOptions{
		Locales:       locales,
		Fallback:      validationLocale,
		LocaleMatcher: cfg.localeMatcher,
		Matcher:       pluralRulesLocaleMatcher(),
	})
	dataLocale := ecma402.ResolveDataLocaleTag(resolution)
	rule := plural.CardinalRuleOrDefault(dataLocale)
	if cfg.typ == string(Ordinal) {
		if ordinalRule, ok := plural.OrdinalRule(dataLocale); ok {
			rule = ordinalRule
		}
	}
	return &PluralRules{
		dataLocale:      dataLocale,
		digitOptions:    resolvedDigits.DigitOptions,
		compact:         compactExponentsForPluralRules(ecma402.ResolveDataLocale(resolution, cldrnumber.ResolveLocale), cfg),
		integerOperands: resolvedDigits.CanUseIntegerOperands(cfg.notation),
		rule:            rule,
		resolved:        resolvedOptionsForPluralRules(resolution.Locale, cfg, resolvedDigits, dataLocale),
	}, nil
}

// Select returns the plural category for a numeric value.
func (f *PluralRules) Select(v Value) (Category, error) {
	numeric := v.numeric
	if f.integerOperands {
		switch numeric.Kind {
		case ecma402.NumericValueInt64:
			return selectInteger(numeric.Int64, f.rule), nil
		case ecma402.NumericValueUint64:
			return selectUnsignedInteger(numeric.Uint64, f.rule), nil
		case ecma402.NumericValueDecimal:
		}
	}
	if err := ecma402.RequireFiniteDecimalInput(numeric.Decimal); err != nil {
		return Other, invalidValue("value", numeric.Decimal.String(), f.resolved.Locale.String(), err)
	}
	_, _, category := resolveDecimal(numeric.Decimal, f.resolved.Notation, f.digitOptions, f.compact, f.rule)
	return category, nil
}

// SelectRange returns the plural category for a numeric range.
func (f *PluralRules) SelectRange(start, end Value) (Category, error) {
	startNumeric := start.numeric
	endNumeric := end.numeric
	if f.integerOperands {
		switch {
		case startNumeric.Kind == ecma402.NumericValueInt64 && endNumeric.Kind == ecma402.NumericValueInt64:
			startCategory := selectInteger(startNumeric.Int64, f.rule)
			if startNumeric.Int64 == endNumeric.Int64 {
				return startCategory, nil
			}
			return selectRangeCategories(startCategory, selectInteger(endNumeric.Int64, f.rule), f), nil
		case startNumeric.Kind == ecma402.NumericValueUint64 && endNumeric.Kind == ecma402.NumericValueUint64:
			startCategory := selectUnsignedInteger(startNumeric.Uint64, f.rule)
			if startNumeric.Uint64 == endNumeric.Uint64 {
				return startCategory, nil
			}
			return selectRangeCategories(startCategory, selectUnsignedInteger(endNumeric.Uint64, f.rule), f), nil
		}
	}
	if err := ecma402.RequireFiniteDecimalInput(startNumeric.Decimal); err != nil {
		return Other, invalidValue("start", startNumeric.Decimal.String(), f.resolved.Locale.String(), err)
	}
	if err := ecma402.RequireFiniteDecimalInput(endNumeric.Decimal); err != nil {
		return Other, invalidValue("end", endNumeric.Decimal.String(), f.resolved.Locale.String(), err)
	}
	return selectRangeDecimal(startNumeric.Decimal, endNumeric.Decimal, f), nil
}

func selectRangeDecimal(start, end decimal.Decimal, f *PluralRules) Category {
	notation := f.resolved.Notation
	digitOptions := f.digitOptions
	compact := f.compact
	rule := f.rule
	startFormatted, _, startCategory := resolveDecimal(start, notation, digitOptions, compact, rule)
	endFormatted, _, endCategory := resolveDecimal(end, notation, digitOptions, compact, rule)
	return selectRangeResolved(startFormatted, startCategory, endFormatted, endCategory, f)
}

func selectRangeResolved(startFormatted string, startCategory Category, endFormatted string, endCategory Category, f *PluralRules) Category {
	if startFormatted == endFormatted {
		return startCategory
	}
	return selectRangeCategories(startCategory, endCategory, f)
}

func selectRangeCategories(startCategory Category, endCategory Category, f *PluralRules) Category {
	if f.resolved.Type != Cardinal {
		return endCategory
	}
	if category, ok := plural.CardinalRange(
		f.dataLocale,
		ecma402pr.Category(startCategory),
		ecma402pr.Category(endCategory),
	); ok {
		return Category(category)
	}
	return endCategory
}

func selectInteger(n int64, rule pluralRuleFunc) Category {
	return Category(rule(ecma402pr.GetIntegerOperands(n)))
}

func selectUnsignedInteger(n uint64, rule pluralRuleFunc) Category {
	return Category(rule(ecma402pr.GetUnsignedIntegerOperands(n)))
}

func resolveDecimal(d decimal.Decimal, notation Notation, digitOptions ecma402nf.DigitOptions, compact compactExponentSet, rule pluralRuleFunc) (string, decimal.Decimal, Category) {
	exponent := 0
	if notation == CompactNotation {
		if _, _, compactExponent, ok := ecma402nf.ResolveCompactMagnitude(d, digitOptions, compact.exponentForMagnitude); ok {
			exponent = compactExponent
		}
	} else if exponent = pluralExponent(d, notation); exponent != 0 {
		d = decimal.Scale10(d, -int32(exponent)) // #nosec G115 -- exponent is derived from decimal.Log10Floor int32.
	}
	result := ecma402nf.FormatNumericToString(d, digitOptions)
	formatted := strings.TrimPrefix(result.Formatted, "-")
	ops := ecma402pr.GetOperands(formatted, exponent)
	return formatted, result.Rounded, Category(rule(ops))
}

func pluralExponent(d decimal.Decimal, notation Notation) int {
	switch notation {
	case ScientificNotation:
		exponent, _ := ecma402nf.ScientificExponent(d, false)
		return exponent
	case EngineeringNotation:
		exponent, _ := ecma402nf.ScientificExponent(d, true)
		return exponent
	default:
		return 0
	}
}

func invalidValue(name, value, loc string, err error) error {
	return ecma402.InvalidFiniteNumericValueError(pluralRulesOwner, name, value, loc, err)
}

type compactExponentSet struct {
	entries []compactExponentEntry
}

type compactExponentEntry struct {
	magnitude int
	exponent  int
}

func compactExponentsForPluralRules(loc cldrnumber.Locale, cfg config) compactExponentSet {
	if cfg.notation != string(CompactNotation) {
		return compactExponentSet{}
	}
	entries := make([]compactExponentEntry, 0, ecma402nf.MaxCompactMagnitude-ecma402nf.MinCompactMagnitude+1)
	for exponent := ecma402nf.MaxCompactMagnitude; exponent >= ecma402nf.MinCompactMagnitude; exponent-- {
		other := loc.CompactPattern("", cfg.compactDisplay, exponent, "other")
		if other == "" {
			continue
		}
		entries = append(entries, compactExponentEntry{
			magnitude: exponent,
			exponent:  ecma402nf.CompactExponentForPattern(exponent, other),
		})
	}
	return compactExponentSet{entries: entries}
}

func (p compactExponentSet) exponentForMagnitude(magnitude int) (int, bool) {
	for _, entry := range p.entries {
		if magnitude >= entry.magnitude {
			return entry.exponent, true
		}
	}
	return 0, false
}
