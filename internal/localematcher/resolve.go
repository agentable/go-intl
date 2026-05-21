package localematcher

import "slices"

type LocaleDataLookup interface {
	For(locale, key string) []string
}

// Option is a resolved option value that can override a Unicode extension key.
type Option struct {
	Key   string
	Value string
}

type ResolveOptions struct {
	Algorithm             Algorithm
	Matcher               *Matcher
	Requested             []string
	Supported             []string
	DefaultLocale         string
	RelevantExtensionKeys []string
	OptionValues          []Option
	Options               map[string]string
	LocaleData            LocaleDataLookup
	Maximizer             Maximizer
}

type ResolvedLocale struct {
	Locale     string
	DataLocale string
	Extensions map[string]string
}

func ResolveLocale(opts ResolveOptions) ResolvedLocale {
	matched := matchForResolve(opts)
	foundLocale := matched.Locale
	result := ResolvedLocale{Locale: foundLocale, DataLocale: matched.DataLocale}
	if len(opts.RelevantExtensionKeys) > 0 {
		result.Extensions = map[string]string{}
	}
	supportedKeywords := []keyword{}
	for _, key := range opts.RelevantExtensionKeys {
		keyLocaleData := localeDataFor(opts.LocaleData, matched.DataLocale, key)
		value := ""
		if len(keyLocaleData) > 0 {
			value = keyLocaleData[0]
		}
		requestedValue := UnicodeExtensionValue(matched.Extension, key)
		if requestedValue != "" && slices.Contains(keyLocaleData, requestedValue) {
			value = requestedValue
			supportedKeywords = append(supportedKeywords, keyword{key: key, value: requestedValue})
		}
		if optionsValue := optionValueFor(opts, key); optionsValue != "" && slices.Contains(keyLocaleData, optionsValue) {
			value = optionsValue
			supportedKeywords = removeKeyword(supportedKeywords, key)
		}
		result.Extensions[key] = value
	}
	if len(supportedKeywords) > 0 {
		result.Locale = InsertUnicodeExtensionAndCanonicalize(foundLocale, supportedKeywords)
	}
	return result
}

func matchForResolve(opts ResolveOptions) Result {
	if opts.Matcher != nil {
		return opts.Matcher.Match(opts.Requested, opts.DefaultLocale, opts.Algorithm)
	}
	return MatchWithMaximizer(opts.Requested, opts.Supported, opts.DefaultLocale, opts.Algorithm, opts.Maximizer)
}

func optionValueFor(opts ResolveOptions, key string) string {
	for _, option := range opts.OptionValues {
		if option.Key == key {
			return option.Value
		}
	}
	return opts.Options[key]
}

func localeDataFor(data LocaleDataLookup, loc, key string) []string {
	if data == nil {
		return nil
	}
	return data.For(loc, key)
}

func removeKeyword(keywords []keyword, key string) []keyword {
	out := keywords[:0]
	for _, kw := range keywords {
		if kw.key != key {
			out = append(out, kw)
		}
	}
	return out
}
