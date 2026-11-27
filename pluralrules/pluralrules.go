package pluralrules

import (
	"fmt"
	"strings"
	"sync"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/cldr/plural"
	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
	"github.com/agentable/go-intl/internal/localematcher"
	cldrpattern "github.com/agentable/go-intl/internal/pattern"
	pluralop "github.com/agentable/go-intl/internal/plural"
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
	return pluralop.Category(c).String()
}

func (c Category) MarshalText() ([]byte, error) {
	return pluralop.Category(c).MarshalText()
}

type PluralRules struct {
	dataLocale      string
	digitOptions    ecma402nf.ResolvedDigitOptions
	compact         compactExponentSet
	integerOperands bool
	rule            pluralRuleFunc
	resolved        ResolvedOptions
}

type pluralRuleFunc func(pluralop.OperandsRecord) pluralop.Category

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
	rule, err := plural.Rule(dataLocale, cfg.typ)
	if err != nil {
		return nil, err
	}
	compact, err := compactExponentsForPluralRules(ecma402.ResolveDataLocale(resolution, cldrnumber.ResolveLocale), cfg)
	if err != nil {
		return nil, err
	}
	return &PluralRules{
		dataLocale:      dataLocale,
		digitOptions:    resolvedDigits,
		compact:         compact,
		integerOperands: resolvedDigits.CanUseIntegerOperands(cfg.notation),
		rule:            rule,
		resolved:        resolvedOptionsForPluralRules(resolution.Locale, cfg, resolvedDigits, dataLocale),
	}, nil
}

// Select returns the plural category for a numeric value.
func (f *PluralRules) Select(v Value) Category {
	numeric := v.numeric
	if f.integerOperands {
		switch numeric.Kind {
		case ecma402.NumericValueInt64:
			return selectInteger(numeric.Int64, f.rule)
		case ecma402.NumericValueUint64:
			return selectUnsignedInteger(numeric.Uint64, f.rule)
		case ecma402.NumericValueDecimal:
		}
	}
	_, category := f.resolveDecimal(numeric.Decimal)
	return category
}

func selectInteger(n int64, rule pluralRuleFunc) Category {
	return Category(rule(pluralop.GetIntegerOperands(n)))
}

func selectUnsignedInteger(n uint64, rule pluralRuleFunc) Category {
	return Category(rule(pluralop.GetUnsignedIntegerOperands(n)))
}

func (f *PluralRules) resolveDecimal(d decimal.Decimal) (string, Category) {
	if !d.IsFinite() {
		return d.String(), Other
	}
	result := ecma402nf.FormatNumericToString(d, f.digitOptions)
	formatted := strings.TrimPrefix(result.Formatted, "-")
	exponent := 0
	notation := f.resolved.Notation
	switch notation {
	case CompactNotation:
		integer, _, _ := strings.Cut(formatted, ".")
		magnitude := len(strings.TrimLeft(integer, "0")) - 1
		exponent = f.compact.exponentForMagnitude(magnitude)
	case ScientificNotation, EngineeringNotation:
		exponent, _ = ecma402nf.ScientificExponent(result.Rounded, f.digitOptions, notation == EngineeringNotation)
	case StandardNotation:
	}
	ops := pluralop.GetOperands(formatted, exponent)
	return formatted, Category(f.rule(ops))
}

type compactExponentSet []compactExponentEntry

type compactExponentEntry struct {
	magnitude int
	exponent  int
}

func compactExponentsForPluralRules(loc cldrnumber.Locale, cfg config) (compactExponentSet, error) {
	if cfg.notation != string(CompactNotation) {
		return nil, nil
	}
	entries := make(compactExponentSet, 0, ecma402nf.MaxCompactMagnitude-ecma402nf.MinCompactMagnitude+1)
	for exponent := ecma402nf.MaxCompactMagnitude; exponent >= ecma402nf.MinCompactMagnitude; exponent-- {
		other := loc.CompactPattern("", cfg.compactDisplay, exponent, "other")
		if other == "" {
			continue
		}
		parsed, err := cldrpattern.ParseCompact(exponent, other)
		if err != nil {
			return nil, fmt.Errorf("pluralrules: compact pattern display %s magnitude %d: %w", cfg.compactDisplay, exponent, err)
		}
		entries = append(entries, compactExponentEntry{
			magnitude: exponent,
			exponent:  parsed.Exponent,
		})
	}
	return entries, nil
}

func (p compactExponentSet) exponentForMagnitude(magnitude int) int {
	for _, entry := range p {
		if magnitude >= entry.magnitude {
			return entry.exponent
		}
	}
	return 0
}
