package displaynames

import (
	"github.com/agentable/go-intl/internal/ecma402"
)

type LocaleMatcher string
type Type string
type Style string
type Fallback string
type LanguageDisplay string

const displayNamesOwner = "displaynames"

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

var (
	displayNameTypeValues = [...]string{
		string(Language),
		string(Region),
		string(Script),
		string(Currency),
		string(Calendar),
		string(DateTimeField),
	}
	displayNameStyleValues = [...]string{
		string(LongStyle),
		string(ShortStyle),
		string(NarrowStyle),
	}
	displayNameFallbackValues = [...]string{
		string(CodeFallback),
		string(NoneFallback),
	}
	displayNameLanguageDisplayValues = [...]string{
		string(DialectLanguageDisplay),
		string(StandardLanguageDisplay),
	}
)

type Options struct {
	LocaleMatcher   *string
	Type            *string
	Style           *string
	Fallback        *string
	LanguageDisplay *string
}

type config struct {
	localeMatcher    string
	hasLocaleMatcher bool
	typ              string
	style            string
	fallback         string
	languageDisplay  string
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
	ecma402.ApplyOptionInput(&cfg.localeMatcher, &cfg.hasLocaleMatcher, opts.LocaleMatcher)
	ecma402.ApplyOption(&cfg.typ, opts.Type)
	ecma402.ApplyOption(&cfg.style, opts.Style)
	ecma402.ApplyOption(&cfg.fallback, opts.Fallback)
	ecma402.ApplyOption(&cfg.languageDisplay, opts.LanguageDisplay)
}

func (cfg config) validate(locName string) error {
	return ecma402.ValidateStringOptions(
		displayNamesOwner,
		locName,
		ecma402.LocaleMatcherOptionInput(cfg.localeMatcher, cfg.hasLocaleMatcher),
		displayNameTypeOption(cfg.typ),
		displayNameStyleOption(cfg.style),
		displayNameFallbackOption(cfg.fallback),
		displayNameLanguageDisplayOption(cfg.languageDisplay),
	)
}

func displayNameTypeOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("type", value, displayNameTypeValues[:]...)
}

func displayNameStyleOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("style", value, displayNameStyleValues[:]...)
}

func displayNameFallbackOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("fallback", value, displayNameFallbackValues[:]...)
}

func displayNameLanguageDisplayOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("languageDisplay", value, displayNameLanguageDisplayValues[:]...)
}
