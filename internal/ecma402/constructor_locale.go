package ecma402

import (
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

// UnicodeExtensionKey is a relevant Unicode extension key used by ECMA-402
// constructor locale resolution.
type UnicodeExtensionKey string

const (
	UnicodeExtensionKeyCollation       UnicodeExtensionKey = "co"
	UnicodeExtensionKeyCalendar        UnicodeExtensionKey = "ca"
	UnicodeExtensionKeyHourCycle       UnicodeExtensionKey = "hc"
	UnicodeExtensionKeyNumberingSystem UnicodeExtensionKey = "nu"
	UnicodeExtensionKeyNumeric         UnicodeExtensionKey = "kn"
	UnicodeExtensionKeyCaseFirst       UnicodeExtensionKey = "kf"
)

// NumberingSystemExtensionKeys returns the constructor locale keys for services
// whose only relevant Unicode extension key is "nu".
func NumberingSystemExtensionKeys() []UnicodeExtensionKey {
	return []UnicodeExtensionKey{UnicodeExtensionKeyNumberingSystem}
}

// UnicodeExtensionOption is a constructor option value that overrides a
// relevant Unicode extension key during ECMA-402 locale resolution.
type UnicodeExtensionOption struct {
	Key   UnicodeExtensionKey
	Value string
}

// NumberingSystemExtensionOptions returns the constructor locale option value
// for a formatter-owned numberingSystem option.
func NumberingSystemExtensionOptions(value string) []UnicodeExtensionOption {
	return []UnicodeExtensionOption{{Key: UnicodeExtensionKeyNumberingSystem, Value: value}}
}

// ConstructorLocaleOptions carries the shared ECMA-402 locale negotiation inputs
// every Intl service constructor prepares before formatter-specific data lookup.
type ConstructorLocaleOptions struct {
	Locales               locale.List
	Fallback              locale.Locale
	LocaleMatcher         string
	Matcher               *localematcher.Matcher
	RelevantExtensionKeys []UnicodeExtensionKey
	OptionValues          []UnicodeExtensionOption
	LocaleData            localematcher.LocaleDataLookup
}

// ConstructorLocaleResolution is the shared constructor locale result. Formatter
// packages still own CLDR data fallback, option-derived fields, and errors.
type ConstructorLocaleResolution struct {
	Locale     locale.Locale
	DataLocale string
	Extensions map[UnicodeExtensionKey]string
}

// ResolveConstructorLocale applies the ECMA-402 ResolveLocale path used by Intl
// service constructors after each formatter has validated its typed options.
func ResolveConstructorLocale(opts ConstructorLocaleOptions) ConstructorLocaleResolution {
	algorithm, _ := LocaleMatcherAlgorithm(opts.LocaleMatcher)
	result := localematcher.ResolveLocale(localematcher.ResolveOptions{
		Algorithm:             algorithm,
		Matcher:               opts.Matcher,
		Requested:             RequestedLocaleStrings(opts.Locales),
		DefaultLocale:         DefaultLocale(),
		RelevantExtensionKeys: extensionKeyStrings(opts.RelevantExtensionKeys),
		OptionValues:          extensionOptions(opts.OptionValues),
		LocaleData:            opts.LocaleData,
	})
	resolvedLocale, err := locale.Parse(result.Locale)
	if err != nil {
		resolvedLocale = opts.Fallback
	}
	return ConstructorLocaleResolution{
		Locale:     resolvedLocale,
		DataLocale: result.DataLocale,
		Extensions: extensionValues(result.Extensions),
	}
}

// ResolveDataLocale resolves a formatter-owned data locale from a constructor
// locale result, falling back to the implementation default locale.
func ResolveDataLocale[T any](resolution ConstructorLocaleResolution, resolve func(string) (T, bool)) T {
	loc, ok := resolve(ResolveDataLocaleTag(resolution))
	if ok {
		return loc
	}
	loc, _ = resolve(DefaultLocale())
	return loc
}

// ResolveDataLocaleTag returns the constructor data-locale tag, falling back to
// the implementation default locale when locale matching produced none.
func ResolveDataLocaleTag(resolution ConstructorLocaleResolution) string {
	if resolution.DataLocale != "" {
		return resolution.DataLocale
	}
	return DefaultLocale()
}

func extensionKeyStrings(keys []UnicodeExtensionKey) []string {
	if len(keys) == 0 {
		return nil
	}
	out := make([]string, len(keys))
	for i, key := range keys {
		out[i] = string(key)
	}
	return out
}

func extensionOptions(options []UnicodeExtensionOption) []localematcher.Option {
	if len(options) == 0 {
		return nil
	}
	out := make([]localematcher.Option, len(options))
	for i, option := range options {
		out[i] = localematcher.Option{Key: string(option.Key), Value: option.Value}
	}
	return out
}

func extensionValues(values map[string]string) map[UnicodeExtensionKey]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[UnicodeExtensionKey]string, len(values))
	for key, value := range values {
		out[UnicodeExtensionKey(key)] = value
	}
	return out
}
