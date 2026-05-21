package relativetimeformat

import (
	cldrrelativetime "github.com/agentable/go-intl/internal/cldr/relativetime"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
)

type RelativeTimeFormat struct {
	resolved ResolvedOptions
	number   *numberformat.NumberFormat
	plural   *pluralrules.PluralRules
	fields   cldrrelativetime.RelativeTimeFields
}

func New(locales locale.List, opts Options) (*RelativeTimeFormat, error) {
	validationLocale := ecma402.ValidationLocale(locales)
	cfg := defaultConfig()
	applyOptions(&cfg, opts)
	if err := cfg.validate(validationLocale); err != nil {
		return nil, err
	}
	resolvedLocale, cldrLoc, numberingSystem := resolveLocale(locales, validationLocale, cfg)
	number, err := numberformat.New(locale.List{resolvedLocale}, numberformat.Options{NumberingSystem: numberingSystem})
	if err != nil {
		return nil, invalidOption("numberingSystem", numberingSystem, validationLocale)
	}
	plural, err := pluralrules.New(locale.List{resolvedLocale}, pluralrules.Options{})
	if err != nil {
		return nil, invalidOption("locale", resolvedLocale.String(), validationLocale)
	}
	numberOptions := number.ResolvedOptions()
	return &RelativeTimeFormat{
		number: number,
		plural: plural,
		fields: cldrrelativetime.FieldsFor(cldrLoc),
		resolved: ResolvedOptions{
			Locale:          resolvedLocale,
			Style:           Style(cfg.style),
			Numeric:         Numeric(cfg.numeric),
			NumberingSystem: numberOptions.NumberingSystem,
		},
	}, nil
}
