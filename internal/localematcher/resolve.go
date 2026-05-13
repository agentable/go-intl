package localematcher

import "slices"

type LocaleDataLookup interface {
	For(locale, key string) []string
}

type ResolveOptions struct {
	Algorithm             Algorithm
	Requested             []string
	Supported             []string
	DefaultLocale         string
	RelevantExtensionKeys []string
	Options               map[string]string
	LocaleData            LocaleDataLookup
}

type ResolvedLocale struct {
	Locale     string
	DataLocale string
	Extensions map[string]string
}

func ResolveLocale(opts ResolveOptions) ResolvedLocale {
	matched := Match(opts.Requested, opts.Supported, opts.DefaultLocale, opts.Algorithm)
	foundLocale := matched.Locale
	result := ResolvedLocale{Locale: foundLocale, DataLocale: matched.DataLocale, Extensions: map[string]string{}}
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
		if optionsValue := opts.Options[key]; optionsValue != "" && slices.Contains(keyLocaleData, optionsValue) {
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
