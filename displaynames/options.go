package displaynames

import (
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/locale"
)

type LocaleMatcher string
type Type string
type Style string
type Fallback string
type LanguageDisplay string

const (
	LookupLocaleMatcher  LocaleMatcher = "lookup"
	BestFitLocaleMatcher LocaleMatcher = "best fit"

	Language      Type = "language"
	Region        Type = "region"
	Script        Type = "script"
	Currency      Type = "currency"
	Calendar      Type = "calendar"
	DateTimeField Type = "dateTimeField"

	LongStyle   Style = "long"
	ShortStyle  Style = "short"
	NarrowStyle Style = "narrow"

	CodeFallback Fallback = "code"
	NoneFallback Fallback = "none"

	DialectLanguageDisplay  LanguageDisplay = "dialect"
	StandardLanguageDisplay LanguageDisplay = "standard"
)

type Options struct {
	LocaleMatcher   LocaleMatcher
	Type            Type
	Style           Style
	Fallback        Fallback
	LanguageDisplay LanguageDisplay
}

type config struct {
	localeMatcher   string
	typ             string
	style           string
	fallback        string
	languageDisplay string
}

func defaultConfig() config {
	return config{
		localeMatcher:   string(BestFitLocaleMatcher),
		style:           string(LongStyle),
		fallback:        string(CodeFallback),
		languageDisplay: string(DialectLanguageDisplay),
	}
}

func applyOptions(cfg *config, opts Options) {
	if opts.LocaleMatcher != "" {
		cfg.localeMatcher = string(opts.LocaleMatcher)
	}
	if opts.Type != "" {
		cfg.typ = string(opts.Type)
	}
	if opts.Style != "" {
		cfg.style = string(opts.Style)
	}
	if opts.Fallback != "" {
		cfg.fallback = string(opts.Fallback)
	}
	if opts.LanguageDisplay != "" {
		cfg.languageDisplay = string(opts.LanguageDisplay)
	}
}

func (cfg config) validate(loc locale.Locale) error {
	if check, ok := ecma402.InvalidStringOption(ecma402.LocaleMatcherOption(cfg.localeMatcher)); ok {
		return invalidOption(check.Name, check.Value, loc)
	}
	switch cfg.typ {
	case "":
		return invalidOption("type", "", loc)
	case string(Language), string(Region), string(Script), string(Currency), string(Calendar), string(DateTimeField):
	default:
		return invalidOption("type", cfg.typ, loc)
	}
	if check, ok := ecma402.InvalidStringOption(
		ecma402.RequiredStringOption("style", cfg.style, string(LongStyle), string(ShortStyle), string(NarrowStyle)),
		ecma402.RequiredStringOption("fallback", cfg.fallback, string(CodeFallback), string(NoneFallback)),
		ecma402.RequiredStringOption("languageDisplay", cfg.languageDisplay, string(DialectLanguageDisplay), string(StandardLanguageDisplay)),
	); ok {
		return invalidOption(check.Name, check.Value, loc)
	}
	return nil
}

func invalidOption(name, value string, loc locale.Locale) error {
	return ecma402.InvalidOptionError("displaynames", name, value, loc.String(), intlerr.ErrInvalidOption)
}
