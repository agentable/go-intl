package localematcher

import (
	"slices"

	"github.com/agentable/go-intl/internal/localeid"
)

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
	LocaleData            LocaleDataLookup
	Maximizer             Maximizer
}

type ResolvedLocale struct {
	Locale     string
	DataLocale string
	Extension  string
	Extensions map[string]string
}

func ResolveLocale(opts ResolveOptions) ResolvedLocale {
	var matched Result
	if opts.Matcher != nil {
		matched = opts.Matcher.Match(opts.Requested, opts.DefaultLocale, opts.Algorithm)
	} else {
		matched = MatchWithMaximizer(opts.Requested, opts.Supported, opts.DefaultLocale, opts.Algorithm, opts.Maximizer)
	}
	foundLocale := matched.Locale
	result := ResolvedLocale{Locale: foundLocale, DataLocale: matched.DataLocale, Extension: matched.Extension}
	if len(opts.RelevantExtensionKeys) > 0 {
		result.Extensions = map[string]string{}
	}
	supportedKeywords := []localeid.UnicodeKeyword{}
	for _, key := range opts.RelevantExtensionKeys {
		var keyLocaleData []string
		if opts.LocaleData != nil {
			keyLocaleData = opts.LocaleData.For(matched.DataLocale, key)
		}
		value := ""
		if len(keyLocaleData) > 0 {
			value = keyLocaleData[0]
		}
		requestedValue := UnicodeExtensionValue(matched.Extension, key)
		supportedKeywordValue := ""
		if requestedValue != "" && slices.Contains(keyLocaleData, requestedValue) {
			value = requestedValue
			supportedKeywordValue = requestedValue
		}
		for _, option := range opts.OptionValues {
			if option.Key != key {
				continue
			}
			if option.Value != "" && slices.Contains(keyLocaleData, option.Value) {
				value = option.Value
				supportedKeywordValue = ""
			}
			break
		}
		if supportedKeywordValue != "" {
			supportedKeywords = append(supportedKeywords, localeid.UnicodeKeyword{Key: key, Value: supportedKeywordValue})
		}
		result.Extensions[key] = value
	}
	if len(supportedKeywords) > 0 {
		result.Locale = InsertUnicodeExtensionAndCanonicalize(foundLocale, supportedKeywords)
	}
	return result
}
