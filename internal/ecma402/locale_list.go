package ecma402

import (
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

// CanonicalLocaleList returns the first occurrence of each canonical locale
// while preserving request order. Raw string parsing has already happened at
// the public Go boundary.
func CanonicalLocaleList(locales locale.List) locale.List {
	seen := make(map[string]bool, len(locales))
	out := make(locale.List, 0, len(locales))
	for _, loc := range locales {
		key := loc.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, loc)
	}
	return out
}

// RequestedLocaleStrings returns the canonical requested locale identifiers.
//
// A nil result represents omitted locales and lets ResolveLocale select the
// default locale through the environment provider.
func RequestedLocaleStrings(locales locale.List) []string {
	if len(locales) == 0 {
		return nil
	}
	canonical := CanonicalLocaleList(locales)
	out := make([]string, len(canonical))
	for i, loc := range canonical {
		out[i] = loc.String()
	}
	return out
}

// ValidationLocale returns the locale used when option validation needs a
// concrete locale for error context before locale negotiation has resolved.
func ValidationLocale(locales locale.List) locale.Locale {
	if len(locales) > 0 {
		return locales[0]
	}
	return defaultLocaleValue()
}

// SupportedLocales canonicalizes requested locales, then filters them against a
// generated supported-locale set while preserving requested order.
func SupportedLocales(
	supported []string,
	requested locale.List,
	matcher localematcher.Algorithm,
	maximizer localematcher.Maximizer,
) locale.List {
	canonical := CanonicalLocaleList(requested)
	return localematcher.FilterLocalesWithMaximizer(supported, canonical, matcher, maximizer)
}

// SupportedLocalesOf applies the ECMA-402 supportedLocalesOf localeMatcher
// option to a generated supported-locale set. Invalid localeMatcher values
// return an OptionError owned by the calling formatter package.
func SupportedLocalesOf(opts SupportedLocalesOptions) (locale.List, error) {
	check := LocaleMatcherOptionInput("", false)
	if opts.LocaleMatcher != nil {
		check = LocaleMatcherOptionInput(*opts.LocaleMatcher, true)
	}
	if check, ok := InvalidStringOption(check); ok {
		loc := ValidationLocale(opts.Requested).String()
		return nil, InvalidStringOptionError(opts.Owner, check, loc)
	}
	matcher, _ := LocaleMatcherAlgorithm(check.Value)
	supported := SupportedLocales(opts.Supported, opts.Requested, matcher, opts.Maximizer)
	return supported, nil
}

// SupportedLocalesOptions carries the formatter-owned inputs to SupportedLocalesOf.
type SupportedLocalesOptions struct {
	Owner     string
	Supported []string
	Requested locale.List
	// LocaleMatcher is the pointer-backed supportedLocalesOf option; nil means omitted.
	LocaleMatcher *string
	Maximizer     localematcher.Maximizer
}
