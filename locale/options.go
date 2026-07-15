package locale

import "github.com/agentable/go-intl/internal/localeid"

type Options struct {
	Language        *string
	Script          *string
	Region          *string
	Calendar        *string
	Collation       *string
	HourCycle       *string
	CaseFirst       *string
	Numeric         *bool
	NumberingSystem *string
	FirstDayOfWeek  *string
}

func applyLanguageOptions(loc *Locale, opts Options) error {
	if opts.Language == nil && opts.Script == nil && opts.Region == nil {
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
	tag, err := localeid.ReplaceLanguageSubtags(loc.tag, lang, script, region)
	if err != nil {
		return invalidLocaleOptionExpected("languageIdentifier", localeid.Join(lang, script, region), localeLanguageIdentifierExpected, err)
	}
	loc.tag = tag
	return nil
}

func applyOptions(loc *Locale, opts Options) error {
	if err := applyStringOption(&loc.ext.calendar, opts.Calendar, "calendar", localeUnicodeTypeExpected); err != nil {
		return err
	}
	if err := applyStringOption(&loc.ext.collation, opts.Collation, "collation", localeUnicodeTypeExpected); err != nil {
		return err
	}
	if err := applyStringOption(&loc.ext.hourCycle, opts.HourCycle, "hourCycle", localeHourCycleExpected); err != nil {
		return err
	}
	if err := applyStringOption(&loc.ext.caseFirst, opts.CaseFirst, "caseFirst", localeCaseFirstExpected); err != nil {
		return err
	}
	if opts.Numeric != nil {
		loc.ext.setNumericOption(*opts.Numeric)
	}
	if err := applyStringOption(&loc.ext.numberingSystem, opts.NumberingSystem, "numberingSystem", localeUnicodeTypeExpected); err != nil {
		return err
	}
	return applyStringOption(&loc.ext.firstDayOfWeek, opts.FirstDayOfWeek, "firstDayOfWeek", localeFirstDayExpected)
}

func applyStringOption(dst *string, value *string, name, expected string) error {
	if value == nil {
		return nil
	}
	if *value == "" {
		return invalidLocaleOptionExpected(name, *value, expected, nil)
	}
	*dst = *value
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
