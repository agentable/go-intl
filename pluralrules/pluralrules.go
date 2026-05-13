package pluralrules

import (
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"

	cldrdata "github.com/agentable/go-intl/internal/cldr"
	"github.com/agentable/go-intl/internal/cldr/plural"
	"github.com/agentable/go-intl/internal/decimal"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
	ecma402pr "github.com/agentable/go-intl/internal/ecma402/pluralrules"
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
	loc  locale.Locale
	cfg  config
	rule func(ecma402pr.OperandsRecord) ecma402pr.Category
}

func New(loc locale.Locale, opts ...Options) (*PluralRules, error) {
	cfg := defaultConfig()
	if len(opts) > 1 {
		return nil, invalidOption("options", "multiple")
	}
	if len(opts) > 0 {
		cfg = configFromOptions(opts[0])
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	cfg = resolveDigitDefaults(cfg)
	dataLocale := loc.Tag().String()
	alg := localematcher.AlgorithmBestFit
	if cfg.localeMatcher == string(LookupLocaleMatcher) {
		alg = localematcher.AlgorithmLookup
	}
	matched := localematcher.Match([]string{loc.String()}, plural.SupportedLocales(), "en", alg)
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
	return &PluralRules{loc: loc, cfg: cfg, rule: rule}, nil
}

func (r *PluralRules) SelectInt(n int) Category {
	return r.SelectInt64(int64(n))
}

func (r *PluralRules) SelectInt64(n int64) Category {
	_, _, category := r.resolveDecimal(decimal.FromInt64(n))
	return category
}

func (r *PluralRules) SelectUint(n uint) Category {
	return r.SelectUint64(uint64(n))
}

func (r *PluralRules) SelectUint64(n uint64) Category {
	_, _, category := r.resolveDecimal(decimal.New(false, new(big.Int).SetUint64(n), 0))
	return category
}

func (r *PluralRules) SelectFloat64(n float64) (Category, error) {
	return r.selectFinite(decimal.FromFloat64(n), fmt.Sprint(n))
}

func (r *PluralRules) SelectDecimal(s string) (Category, error) {
	d, err := decimal.ParseString(s)
	if err != nil {
		return Other, invalidValue("decimal", s, err)
	}
	return r.selectFinite(d, s)
}

func (r *PluralRules) SelectRangeInt(start, end int) Category {
	return r.SelectRangeInt64(int64(start), int64(end))
}

func (r *PluralRules) SelectRangeInt64(start, end int64) Category {
	return r.selectRangeDecimal(decimal.FromInt64(start), decimal.FromInt64(end))
}

func (r *PluralRules) SelectRangeUint(start, end uint) Category {
	return r.SelectRangeUint64(uint64(start), uint64(end))
}

func (r *PluralRules) SelectRangeUint64(start, end uint64) Category {
	return r.selectRangeDecimal(
		decimal.New(false, new(big.Int).SetUint64(start), 0),
		decimal.New(false, new(big.Int).SetUint64(end), 0),
	)
}

func (r *PluralRules) SelectRangeFloat64(start, end float64) (Category, error) {
	startDecimal := decimal.FromFloat64(start)
	endDecimal := decimal.FromFloat64(end)
	if !startDecimal.IsFinite() {
		return Other, invalidValue("start", fmt.Sprint(start), nil)
	}
	if !endDecimal.IsFinite() {
		return Other, invalidValue("end", fmt.Sprint(end), nil)
	}
	return r.selectRangeDecimal(startDecimal, endDecimal), nil
}

func (r *PluralRules) SelectRangeDecimal(start, end string) (Category, error) {
	startDecimal, err := decimal.ParseString(start)
	if err != nil {
		return Other, invalidValue("start", start, err)
	}
	endDecimal, err := decimal.ParseString(end)
	if err != nil {
		return Other, invalidValue("end", end, err)
	}
	if !startDecimal.IsFinite() {
		return Other, invalidValue("start", start, nil)
	}
	if !endDecimal.IsFinite() {
		return Other, invalidValue("end", end, nil)
	}
	return r.selectRangeDecimal(startDecimal, endDecimal), nil
}

func (r *PluralRules) selectFinite(d decimal.Decimal, value string) (Category, error) {
	if !d.IsFinite() {
		return Other, invalidValue("value", value, nil)
	}
	_, _, category := r.resolveDecimal(d)
	return category, nil
}

func (r *PluralRules) selectRangeDecimal(start, end decimal.Decimal) Category {
	_, startRounded, startCategory := r.resolveDecimal(start)
	_, endRounded, endCategory := r.resolveDecimal(end)
	return r.selectRangeResolved(startRounded, startCategory, endRounded, endCategory)
}

func (r *PluralRules) selectRangeResolved(startRounded decimal.Decimal, startCategory Category, endRounded decimal.Decimal, endCategory Category) Category {
	if startRounded.Cmp(endRounded) == 0 {
		return startCategory
	}
	if category, ok := plural.Range(r.loc.Tag().String(), r.cfg.typ.String(), startCategory, endCategory); ok {
		return category
	}
	return endCategory
}

func (r *PluralRules) resolveDecimal(d decimal.Decimal) (string, decimal.Decimal, Category) {
	exponent := r.computeExponent(d)
	if exponent != 0 {
		d = decimal.Scale10(d, -int32(exponent)) // #nosec G115 -- exponent is derived from decimal.Log10Floor int32.
	}
	result := formatDecimal(d, r.cfg)
	formatted := strings.TrimPrefix(result.Formatted, "-")
	ops := ecma402pr.GetOperands(formatted, exponent)
	return formatted, result.Rounded, r.rule(ops)
}

func (r *PluralRules) computeExponent(d decimal.Decimal) int {
	switch r.cfg.notation {
	case string(ScientificNotation):
		return scientificExponent(d, false)
	case string(EngineeringNotation):
		return scientificExponent(d, true)
	case string(CompactNotation):
		if !slices.Contains(cldrdata.NumberSupportedLocales(), r.loc.String()) {
			return 0
		}
		return compactExponent(d)
	default:
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

func compactExponent(d decimal.Decimal) int {
	abs := decimal.Abs(d)
	if abs.IsZero() {
		return 0
	}
	magnitude, err := decimal.Log10Floor(abs)
	if err != nil || magnitude < 3 {
		return 0
	}
	return int(magnitude)
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
	return fmt.Errorf("pluralrules: invalid %s %q: %w", name, value, ErrInvalidOption)
}

func invalidValue(name, value string, err error) error {
	if err != nil {
		return fmt.Errorf("pluralrules: invalid %s %q: %w", name, value, errors.Join(ErrInvalidValue, err))
	}
	return fmt.Errorf("pluralrules: invalid %s %q: %w", name, value, ErrInvalidValue)
}
