package collator

import (
	"sync"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrcoll "github.com/agentable/go-intl/internal/collation"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

// Collator compares strings according to the resolved locale and options.
// A Collator is safe for concurrent use; the embedded x/text/collate.Collator
// is not, so each Compare call holds a private clone.
type Collator struct {
	resolved ResolvedOptions
	options  []collate.Option
	tag      language.Tag
}

var collatorLocaleMatcher = sync.OnceValue(func() *localematcher.Matcher {
	return localematcher.NewMatcher(cldrcoll.SupportedLocales(), cldrlocale.Maximize)
})

// New constructs a Collator for the requested locale and options.
func New(locales locale.List, opts Options) (*Collator, error) {
	validationLocale := ecma402.ValidationLocale(locales)
	cfg := defaultConfig()
	applyOptions(&cfg, opts)
	if err := cfg.validate(validationLocale); err != nil {
		return nil, err
	}

	resolvedLocale, dataLocale, cfg, err := resolveLocale(locales, validationLocale, cfg)
	if err != nil {
		return nil, err
	}

	tag := collateTag(dataLocale, cfg)
	collOpts := buildCollateOptions(cfg)

	sensitivity := cfg.sensitivity
	if sensitivity == "" {
		sensitivity = string(VariantSensitivity)
	}
	caseFirst := cfg.caseFirst
	if caseFirst == "" {
		caseFirst = string(FalseCaseFirst)
	}

	f := &Collator{
		options: collOpts,
		tag:     tag,
		resolved: ResolvedOptions{
			Locale:            resolvedLocale,
			Usage:             Usage(cfg.usage),
			Sensitivity:       Sensitivity(sensitivity),
			CaseFirst:         CaseFirst(caseFirst),
			Collation:         "default",
			Numeric:           cfg.numeric,
			IgnorePunctuation: cfg.ignorePunctuation,
		},
	}
	return f, nil
}

// Compare returns a negative number when x sorts before y, zero when equal,
// and positive when x sorts after y. The JS bridge for
// `Intl.Collator.prototype.compare`.
func (f *Collator) Compare(x, y string) int {
	return f.newCollator().CompareString(x, y)
}

func resolveLocale(locales locale.List, fallback locale.Locale, cfg config) (locale.Locale, string, config, error) {
	matcher, _ := ecma402.LocaleMatcherAlgorithm(cfg.localeMatcher)
	requested := ecma402.RequestedLocaleStrings(locales)
	defaultLocale := ecma402.DefaultLocale()
	matched := collatorLocaleMatcher().Match(requested, defaultLocale, matcher)
	if err := validateMatchedExtension(matched, cfg, fallback); err != nil {
		return locale.Locale{}, "", cfg, err
	}
	resolved := localematcher.ResolveLocale(localematcher.ResolveOptions{
		Algorithm:             matcher,
		Matcher:               collatorLocaleMatcher(),
		Requested:             requested,
		DefaultLocale:         defaultLocale,
		RelevantExtensionKeys: []string{"kn", "kf"},
		OptionValues:          collatorOptionValues(cfg),
		LocaleData:            collatorLocaleData{},
	})

	dataLocale := resolved.DataLocale
	if dataLocale == "" {
		dataLocale = defaultLocale
	}
	resolvedLocale, err := locale.Parse(resolved.Locale)
	if err != nil {
		resolvedLocale = fallback
	}
	cfg.numeric = resolved.Extensions["kn"] == "true"
	cfg.caseFirst = resolved.Extensions["kf"]
	if cfg.caseFirst == "" {
		cfg.caseFirst = string(FalseCaseFirst)
	}
	return resolvedLocale, dataLocale, cfg, nil
}

func validateMatchedExtension(matched localematcher.Result, cfg config, fallback locale.Locale) error {
	if matched.Extension == "" {
		return nil
	}
	loc := fallback
	if parsed, err := locale.Parse(matched.Locale + matched.Extension); err == nil {
		loc = parsed
	}
	if value := localematcher.UnicodeExtensionValue(matched.Extension, "co"); value != "" && cfg.collation == "" && !isDefaultCollation(value) {
		return unsupportedOption("collation", value, loc)
	}
	if value := localematcher.UnicodeExtensionValue(matched.Extension, "kf"); value != "" && cfg.caseFirst == "" && value != string(FalseCaseFirst) {
		return unsupportedOption("caseFirst", value, loc)
	}
	return nil
}

func collatorOptionValues(cfg config) []localematcher.Option {
	var out []localematcher.Option
	if cfg.numericSet {
		value := "false"
		if cfg.numeric {
			value = "true"
		}
		out = append(out, localematcher.Option{Key: "kn", Value: value})
	}
	if cfg.caseFirst != "" {
		out = append(out, localematcher.Option{Key: "kf", Value: cfg.caseFirst})
	}
	return out
}

type collatorLocaleData struct{}

func (collatorLocaleData) For(_, key string) []string {
	switch key {
	case "kn":
		return []string{"false", "true"}
	case "kf":
		return []string{"false"}
	default:
		return nil
	}
}

func (f *Collator) newCollator() *collate.Collator {
	return collate.New(f.tag, f.options...)
}

func collateTag(dataLocale string, cfg config) language.Tag {
	tag, _ := language.Parse(dataLocale)
	if !cfg.ignorePunctuation {
		return tag
	}
	// x/text exposes UCA alternate-shifted handling through the BCP 47 "ka" key.
	shifted, err := tag.SetTypeForKey("ka", "shifted")
	if err != nil {
		return tag
	}
	return shifted
}

func buildCollateOptions(cfg config) []collate.Option {
	opts := make([]collate.Option, 0, 4)
	switch cfg.sensitivity {
	case string(BaseSensitivity):
		opts = append(opts, collate.IgnoreCase, collate.IgnoreDiacritics)
	case string(AccentSensitivity):
		opts = append(opts, collate.IgnoreCase)
	case string(CaseSensitivity):
		opts = append(opts, collate.IgnoreDiacritics)
	}
	if cfg.numeric {
		opts = append(opts, collate.Numeric)
	}
	return opts
}
