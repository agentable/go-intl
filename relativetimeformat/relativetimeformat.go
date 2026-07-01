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
	fields   map[Unit]relativeTimeField
}

func New(locales locale.List, opts Options) (*RelativeTimeFormat, error) {
	validationLocale := ecma402.ValidationLocale(locales)
	validationLocaleName := validationLocale.String()
	cfg := defaultConfig()
	applyOptions(&cfg, opts)
	if err := cfg.validate(validationLocaleName); err != nil {
		return nil, err
	}
	resolvedLocale, cldrLoc, numberingSystem := resolveLocale(locales, validationLocale, cfg)
	number, err := numberformat.New(locale.List{resolvedLocale}, numberformat.Options{NumberingSystem: &numberingSystem})
	if err != nil {
		return nil, ecma402.InvalidOptionErrorExpected(
			relativeTimeFormatOwner,
			"numberingSystem",
			numberingSystem,
			validationLocaleName,
			"a numbering system supported by relative-time number formatting",
			err,
		)
	}
	plural, err := pluralrules.New(locale.List{resolvedLocale}, pluralrules.Options{})
	if err != nil {
		return nil, ecma402.InvalidOptionErrorExpected(
			relativeTimeFormatOwner,
			"locale",
			resolvedLocale.String(),
			validationLocaleName,
			"a locale supported by relative-time plural rules",
			err,
		)
	}
	fields, err := compileRelativeTimeFields(cldrrelativetime.FieldsFor(cldrLoc), Style(cfg.style))
	if err != nil {
		return nil, err
	}
	numberOptions := number.ResolvedOptions()
	resolved := ResolvedOptions{
		Locale:          resolvedLocale,
		Style:           Style(cfg.style),
		Numeric:         Numeric(cfg.numeric),
		NumberingSystem: numberOptions.NumberingSystem,
	}
	return &RelativeTimeFormat{
		resolved: resolved,
		number:   number,
		plural:   plural,
		fields:   fields,
	}, nil
}
