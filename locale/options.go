package locale

import (
	"slices"
	"strings"

	"golang.org/x/text/language"
)

type Options struct {
	Language        string
	Script          string
	Region          string
	Variants        []string
	Calendar        string
	Collation       string
	HourCycle       string
	CaseFirst       string
	Numeric         *bool
	NumberingSystem string
	FirstDayOfWeek  string
}

func applyLanguageOptions(loc *Locale, opts Options) error {
	if opts.Language == "" && opts.Script == "" && opts.Region == "" && opts.Variants == nil {
		return nil
	}
	lang, script, region := tagParts(loc.tag)
	if opts.Language != "" {
		lang = strings.ToLower(opts.Language)
	}
	if opts.Script != "" {
		script = canonicalScript(opts.Script)
	}
	if opts.Region != "" {
		region = strings.ToUpper(opts.Region)
	}
	variants := loc.Variants()
	if opts.Variants != nil {
		variants = slices.Clone(opts.Variants)
	}
	base, err := buildLanguageTag(lang, script, region, variants)
	if err != nil {
		return err
	}
	loc.tag = base
	return nil
}

func applyOptions(loc *Locale, opts Options) {
	if opts.Calendar != "" {
		loc.ext.calendar = strings.ToLower(opts.Calendar)
	}
	if opts.Collation != "" {
		loc.ext.collation = strings.ToLower(opts.Collation)
	}
	if opts.HourCycle != "" {
		loc.ext.hourCycle = strings.ToLower(opts.HourCycle)
	}
	if opts.CaseFirst != "" {
		loc.ext.caseFirst = strings.ToLower(opts.CaseFirst)
	}
	if opts.Numeric != nil {
		loc.ext.numericSet = true
		loc.ext.numeric = *opts.Numeric
		if *opts.Numeric {
			loc.ext.numericValue = ""
		} else {
			loc.ext.numericValue = "false"
		}
	}
	if opts.NumberingSystem != "" {
		loc.ext.numberingSystem = strings.ToLower(opts.NumberingSystem)
	}
	if opts.FirstDayOfWeek != "" {
		loc.ext.firstDayOfWeek = strings.ToLower(opts.FirstDayOfWeek)
	}
}

func buildLanguageTag(lang, script, region string, variants []string) (language.Tag, error) {
	if err := validateVariants(variants); err != nil {
		return language.Tag{}, err
	}
	parts := []string{lang}
	if script != "" {
		parts = append(parts, script)
	}
	if region != "" {
		parts = append(parts, region)
	}
	for _, variant := range variants {
		parts = append(parts, strings.ToLower(variant))
	}
	tag := strings.Join(parts, "-")
	parsed, err := language.Parse(tag)
	if err != nil {
		return language.Tag{}, invalidLocaleOption("languageIdentifier", tag, err)
	}
	return parsed, nil
}

func validateVariants(variants []string) error {
	seen := map[string]bool{}
	for _, variant := range variants {
		v := strings.ToLower(variant)
		if seen[v] || !isVariantSubtag(v) {
			return invalidLocaleOption("variants", variant, nil)
		}
		seen[v] = true
	}
	return nil
}

func isVariantSubtag(v string) bool {
	if len(v) >= 5 && len(v) <= 8 {
		return asciiAlnum(v)
	}
	return len(v) == 4 && v[0] >= '0' && v[0] <= '9' && asciiAlnum(v)
}

func asciiAlnum(s string) bool {
	for i := range len(s) {
		c := s[i]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') {
			continue
		}
		return false
	}
	return true
}

func canonicalScript(script string) string {
	script = strings.ToLower(script)
	if script == "" {
		return ""
	}
	return strings.ToUpper(script[:1]) + script[1:]
}
