package pluralrules

import (
	"errors"
	"strings"
	"sync"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/cldr/plural"
	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
	ecma402pr "github.com/agentable/go-intl/internal/ecma402/pluralrules"
	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

type Category = ecma402pr.Category

const (
	Zero  = ecma402pr.Zero
	One   = ecma402pr.One
	Two   = ecma402pr.Two
	Few   = ecma402pr.Few
	Many  = ecma402pr.Many
	Other = ecma402pr.Other
)

type PluralRules struct {
	loc         locale.Locale
	rangeLocale string
	cfg         config
	rule        func(ecma402pr.OperandsRecord) ecma402pr.Category
}

var pluralRulesLocaleMatcher = sync.OnceValue(func() *localematcher.Matcher {
	return localematcher.NewMatcher(plural.SupportedLocales(), cldrlocale.Maximize)
})

func New(locales locale.List, opts Options) (*PluralRules, error) {
	cfg := configFromOptions(opts)
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	cfg = resolveDigitDefaults(cfg)
	loc := ecma402.ValidationLocale(locales)
	defaultLocale := ecma402.DefaultLocale()
	dataLocale := defaultLocale
	alg := localematcher.AlgorithmBestFit
	if cfg.localeMatcher == string(LookupLocaleMatcher) {
		alg = localematcher.AlgorithmLookup
	}
	matched := pluralRulesLocaleMatcher().Match(ecma402.RequestedLocaleStrings(locales), defaultLocale, alg)
	if matched.DataLocale != "" {
		dataLocale = matched.DataLocale
		if matchedLoc, err := locale.Parse(matched.Locale); err == nil {
			loc = matchedLoc
		}
	}
	rule, ok := plural.CardinalRule(dataLocale)
	if cfg.typ == Ordinal {
		rule, ok = plural.OrdinalRule(dataLocale)
	}
	if !ok {
		rule, _ = plural.CardinalRule("en")
	}
	return &PluralRules{loc: loc, rangeLocale: loc.Tag().String(), cfg: cfg, rule: rule}, nil
}

// Select returns the plural category for a numeric value.
func (f *PluralRules) Select(v Value) (Category, error) {
	if f.cfg.canUseIntegerOperands() {
		switch v.kind {
		case valueInt64:
			return f.selectInteger(v.int64), nil
		case valueUint64:
			return f.selectUnsignedInteger(v.uint64), nil
		case valueDecimal:
		}
	}
	if err := ecma402.RequireFiniteDecimalInput(v.decimal); err != nil {
		return Other, invalidValue("value", v.decimal.String(), err)
	}
	_, _, category := f.resolveDecimal(v.decimal)
	return category, nil
}

func parseFiniteDecimalValue(name, value string) (decimal.Decimal, error) {
	d, err := ecma402.ParseFiniteDecimalInput(value)
	if err != nil {
		return decimal.Decimal{}, invalidValue(name, value, err)
	}
	return d, nil
}

// SelectRange returns the plural category for a numeric range.
func (f *PluralRules) SelectRange(start, end Value) (Category, error) {
	if f.cfg.canUseIntegerOperands() {
		switch {
		case start.kind == valueInt64 && end.kind == valueInt64:
			startCategory := f.selectInteger(start.int64)
			if start.int64 == end.int64 {
				return startCategory, nil
			}
			return f.selectRangeCategories(startCategory, f.selectInteger(end.int64)), nil
		case start.kind == valueUint64 && end.kind == valueUint64:
			startCategory := f.selectUnsignedInteger(start.uint64)
			if start.uint64 == end.uint64 {
				return startCategory, nil
			}
			return f.selectRangeCategories(startCategory, f.selectUnsignedInteger(end.uint64)), nil
		}
	}
	if err := ecma402.RequireFiniteDecimalInput(start.decimal); err != nil {
		return Other, invalidValue("start", start.decimal.String(), err)
	}
	if err := ecma402.RequireFiniteDecimalInput(end.decimal); err != nil {
		return Other, invalidValue("end", end.decimal.String(), err)
	}
	return f.selectRangeDecimal(start.decimal, end.decimal), nil
}

func (f *PluralRules) selectRangeDecimal(start, end decimal.Decimal) Category {
	_, startRounded, startCategory := f.resolveDecimal(start)
	_, endRounded, endCategory := f.resolveDecimal(end)
	return f.selectRangeResolved(startRounded, startCategory, endRounded, endCategory)
}

func (f *PluralRules) selectRangeResolved(startRounded decimal.Decimal, startCategory Category, endRounded decimal.Decimal, endCategory Category) Category {
	if startRounded.Cmp(endRounded) == 0 {
		return startCategory
	}
	return f.selectRangeCategories(startCategory, endCategory)
}

func (f *PluralRules) selectRangeCategories(startCategory Category, endCategory Category) Category {
	if category, ok := plural.Range(f.rangeLocale, f.cfg.typ.String(), startCategory, endCategory); ok {
		return category
	}
	return endCategory
}

func (f *PluralRules) selectInteger(n int64) Category {
	return f.rule(ecma402pr.GetIntegerOperands(n))
}

func (f *PluralRules) selectUnsignedInteger(n uint64) Category {
	return f.rule(ecma402pr.GetUnsignedIntegerOperands(n))
}

func (f *PluralRules) resolveDecimal(d decimal.Decimal) (string, decimal.Decimal, Category) {
	exponent := f.computeExponent(d)
	if exponent != 0 {
		d = decimal.Scale10(d, -int32(exponent)) // #nosec G115 -- exponent is derived from decimal.Log10Floor int32.
	}
	result := formatDecimal(d, f.cfg)
	formatted := strings.TrimPrefix(result.Formatted, "-")
	ops := ecma402pr.GetOperands(formatted, exponent)
	return formatted, result.Rounded, f.rule(ops)
}

func (f *PluralRules) computeExponent(d decimal.Decimal) int {
	switch f.cfg.notation {
	case string(ScientificNotation):
		return scientificExponent(d, false)
	case string(EngineeringNotation):
		return scientificExponent(d, true)
	default:
		// Compact notation is not exposed via the plural operand c/e: ICU and V8
		// derive plural operands from the source value (e.g. 1_200_000) rather
		// than the compact significand (1.2). See pluralrules-formatjs-index-test-ts-{039..046}.
		return 0
	}
}

func scientificExponent(d decimal.Decimal, engineering bool) int {
	abs := decimal.Abs(d)
	if abs.IsZero() {
		return 0
	}
	magnitude, err := decimal.Log10Floor(abs)
	if err != nil {
		return 0
	}
	exponent := int(magnitude)
	if engineering {
		exponent -= positiveMod(exponent, 3)
	}
	return exponent
}

func positiveMod(n, mod int) int {
	out := n % mod
	if out < 0 {
		out += mod
	}
	return out
}

func formatDecimal(d decimal.Decimal, cfg config) ecma402nf.FormattedNumeric {
	return ecma402nf.FormatNumericToString(d, cfg.digitOptions())
}

func resolveDigitDefaults(cfg config) config {
	if !cfg.hasMinSigDigits && !cfg.hasMaxSigDigits && cfg.roundingPriority == "auto" {
		return cfg
	}
	if !cfg.hasMinSigDigits {
		cfg.minSigDigits = 1
	}
	if !cfg.hasMaxSigDigits {
		cfg.maxSigDigits = 21
	}
	return cfg
}

func invalidOption(name, value string) error {
	return ecma402.InvalidOptionError("pluralrules", name, value, "", intlerr.ErrInvalidOption)
}

func invalidValue(name, value string, err error) error {
	cause := intlerr.ErrInvalidValue
	if err != nil {
		cause = errors.Join(intlerr.ErrInvalidValue, err)
	}
	return intlerr.New(intlerr.InvalidValue, "pluralrules", name, value, "", cause)
}
