package locale

import (
	"slices"
	"strings"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/localeid"
)

type Options struct {
	Language        *string
	Script          *string
	Region          *string
	Variants        *string
	Calendar        *string
	Collation       *string
	HourCycle       *string
	CaseFirst       *string
	Numeric         *bool
	NumberingSystem *string
	FirstDayOfWeek  *string
}

func applyLanguageOptions(loc *Locale, opts Options) error {
	if opts.Language == nil && opts.Script == nil && opts.Region == nil && opts.Variants == nil {
		return nil
	}
	lang, script, region := localeid.Parts(loc.tag)
	if err := applySubtagOption(&lang, opts.Language, "language", localeLanguageExpected, localeid.CanonicalUnicodeLanguageSubtag); err != nil {
		return err
	}
	if err := applySubtagOption(&script, opts.Script, "script", localeScriptExpected, localeid.CanonicalUnicodeScriptSubtag); err != nil {
		return err
	}
	if err := applySubtagOption(&region, opts.Region, "region", localeRegionExpected, localeid.CanonicalUnicodeRegionSubtag); err != nil {
		return err
	}
	variants := localeid.VariantSubtags(loc.tag)
	if opts.Variants != nil {
		variants = nil
		for _, value := range strings.Split(*opts.Variants, "-") {
			canonical, ok := localeid.CanonicalUnicodeVariantSubtag(value)
			if !ok || slices.Contains(variants, canonical) {
				return invalidLocaleOptionExpected("variants", *opts.Variants, "one or more distinct Unicode variant subtags", nil)
			}
			variants = append(variants, canonical)
		}
	}
	tag, err := localeid.ReplaceLanguageID(loc.tag, lang, script, region, variants, cldrlocale.MaximizeSubtags)
	if err != nil {
		return invalidLocaleOptionExpected("languageIdentifier", localeid.Join(lang, script, region), localeLanguageIdentifierExpected, err)
	}
	loc.tag = tag
	return nil
}

func applyOptions(loc *Locale, opts Options) error {
	for _, opt := range []struct {
		key, name, expected string
		value               *string
		normalize           func(string) (string, error)
	}{
		{"ca", "calendar", localeUnicodeTypeExpected, opts.Calendar, normalizeUnicodeTypeForKey("ca")},
		{"co", "collation", localeUnicodeTypeExpected, opts.Collation, normalizeUnicodeTypeForKey("co")},
		{"fw", "firstDayOfWeek", localeFirstDayExpected, opts.FirstDayOfWeek, normalizeFirstDayOfWeek},
		{"hc", "hourCycle", localeHourCycleExpected, opts.HourCycle, normalizeHourCycle},
		{"kf", "caseFirst", localeCaseFirstExpected, opts.CaseFirst, normalizeCaseFirst},
		{"nu", "numberingSystem", localeUnicodeTypeExpected, opts.NumberingSystem, normalizeUnicodeTypeForKey("nu")},
	} {
		if opt.value == nil {
			continue
		}
		if *opt.value == "" {
			return invalidLocaleOptionExpected(opt.name, *opt.value, opt.expected, nil)
		}
		value, err := opt.normalize(*opt.value)
		if err != nil {
			return invalidLocaleOptionExpected(opt.name, *opt.value, opt.expected, err)
		}
		loc.ext.setKeyword(opt.key, value)
	}
	if opts.Numeric != nil {
		value := "false"
		if *opts.Numeric {
			value = "true"
		}
		loc.ext.setKeyword("kn", value)
	}
	return nil
}

func applySubtagOption(dst *string, value *string, name, expected string, canonicalize func(string) (string, bool)) error {
	if value == nil {
		return nil
	}
	canonical, ok := canonicalize(*value)
	if !ok {
		return invalidLocaleOptionExpected(name, *value, expected, nil)
	}
	*dst = canonical
	return nil
}
