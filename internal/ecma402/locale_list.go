package ecma402

import (
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

// CanonicalLocaleList returns the first occurrence of each canonical locale
// while preserving request order. Raw string parsing has already happened at
// the public Go boundary.
func CanonicalLocaleList(locales locale.List) locale.List {
	seen := map[string]bool{}
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
	canonical := CanonicalLocaleList(locales)
	if len(canonical) == 0 {
		return nil
	}
	out := make([]string, len(canonical))
	for i, loc := range canonical {
		out[i] = loc.String()
	}
	return out
}

// ValidationLocale returns the locale used when option validation needs a
// concrete locale for error context before locale negotiation has resolved.
func ValidationLocale(locales locale.List) locale.Locale {
	canonical := CanonicalLocaleList(locales)
	if len(canonical) > 0 {
		return canonical[0]
	}
	loc, err := locale.Parse(DefaultLocale())
	if err != nil {
		return locale.Locale{}
	}
	return loc
}

// SupportedLocales canonicalizes and filters requested locales against a
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
func SupportedLocalesOf(
	owner string,
	supported []string,
	requested locale.List,
	localeMatcher string,
	maximizer localematcher.Maximizer,
	invalidOption error,
) (locale.List, error) {
	matcher, ok := LocaleMatcherAlgorithm(localeMatcher)
	if !ok {
		return nil, InvalidOptionError(owner, "localeMatcher", localeMatcher, ValidationLocale(requested).String(), invalidOption)
	}
	return SupportedLocales(supported, requested, matcher, maximizer), nil
}
