package collator

import (
	"slices"
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
// A Collator is safe for concurrent use: each Compare borrows a backend from a
// pool, because an x/text collator mutates private iterators while comparing and
// cannot be shared across goroutines. Pooling gives lock-free parallel compares.
type Collator struct {
	state *collatorState
}

type collatorState struct {
	resolved ResolvedOptions
	backends sync.Pool // of *collate.Collator
}

var collatorLocaleMatcher = sync.OnceValue(func() *localematcher.Matcher {
	return localematcher.NewMatcher(cldrcoll.SupportedLocales(), cldrlocale.Maximize)
})

var (
	supportedCollatorNumericValues   = [...]string{"false", "true"}
	supportedCollatorCaseFirstValues = [...]string{string(FalseCaseFirst)}
)

// New constructs a Collator for the requested locale and options.
func New(locales locale.List, opts Options) (*Collator, error) {
	validationLocale := ecma402.ValidationLocale(locales)
	validationLocaleName := validationLocale.String()
	cfg := defaultConfig()
	applyOptions(&cfg, opts)
	if err := cfg.validate(validationLocaleName); err != nil {
		return nil, err
	}

	resolvedLocale, dataLocale, cfg := resolveLocale(locales, validationLocale, cfg)

	tag := collateTag(dataLocale, cfg)
	collOpts := buildCollateOptions(cfg)

	sensitivity := cfg.sensitivity
	if sensitivity == "" {
		sensitivity = string(VariantSensitivity)
	}

	f := &Collator{state: &collatorState{
		resolved: ResolvedOptions{
			Locale:            resolvedLocale,
			Usage:             Usage(cfg.usage),
			Sensitivity:       Sensitivity(sensitivity),
			CaseFirst:         CaseFirst(cfg.caseFirst),
			Collation:         cfg.collation,
			Numeric:           cfg.numeric,
			IgnorePunctuation: cfg.ignorePunctuation,
		},
		backends: sync.Pool{New: func() any { return collate.New(tag, collOpts...) }},
	}}
	return f, nil
}

// Compare returns a negative number when x sorts before y, zero when equal,
// and positive when x sorts after y. The JS bridge for
// `Intl.Collator.prototype.compare`.
func (f *Collator) Compare(x, y string) int {
	backend := f.state.backends.Get().(*collate.Collator)
	defer f.state.backends.Put(backend)
	return backend.CompareString(x, y)
}

func resolveLocale(locales locale.List, fallback locale.Locale, cfg config) (locale.Locale, string, config) {
	resolution := ecma402.ResolveConstructorLocale(ecma402.ConstructorLocaleOptions{
		Locales:               locales,
		Fallback:              fallback,
		LocaleMatcher:         cfg.localeMatcher,
		Matcher:               collatorLocaleMatcher(),
		RelevantExtensionKeys: []ecma402.UnicodeExtensionKey{ecma402.UnicodeExtensionKeyCollation, ecma402.UnicodeExtensionKeyNumeric, ecma402.UnicodeExtensionKeyCaseFirst},
		OptionValues:          collatorOptionValues(cfg),
		LocaleData:            collatorLocaleData{},
	})

	dataLocale := ecma402.ResolveDataLocaleTag(resolution)
	cfg.collation = resolution.Extensions[ecma402.UnicodeExtensionKeyCollation]
	if cfg.collation == "" {
		cfg.collation = "default"
	}
	cfg.numeric = resolution.Extensions[ecma402.UnicodeExtensionKeyNumeric] == "true"
	cfg.caseFirst = resolution.Extensions[ecma402.UnicodeExtensionKeyCaseFirst]
	if cfg.caseFirst == "" {
		cfg.caseFirst = string(FalseCaseFirst)
	}
	return resolution.Locale, dataLocale, cfg
}

func collatorOptionValues(cfg config) []ecma402.UnicodeExtensionOption {
	var out []ecma402.UnicodeExtensionOption
	if cfg.collation != "" {
		out = append(out, ecma402.UnicodeExtensionOption{Key: ecma402.UnicodeExtensionKeyCollation, Value: cfg.collation})
	}
	if cfg.numericSet {
		value := "false"
		if cfg.numeric {
			value = "true"
		}
		out = append(out, ecma402.UnicodeExtensionOption{Key: ecma402.UnicodeExtensionKeyNumeric, Value: value})
	}
	if cfg.caseFirst != "" {
		out = append(out, ecma402.UnicodeExtensionOption{Key: ecma402.UnicodeExtensionKeyCaseFirst, Value: cfg.caseFirst})
	}
	return out
}

type collatorLocaleData struct{}

func (collatorLocaleData) For(locale, key string) []string {
	switch key {
	case string(ecma402.UnicodeExtensionKeyCollation):
		return cldrcoll.SupportedCollationsForLocale(locale)
	case string(ecma402.UnicodeExtensionKeyNumeric):
		return slices.Clone(supportedCollatorNumericValues[:])
	case string(ecma402.UnicodeExtensionKeyCaseFirst):
		return slices.Clone(supportedCollatorCaseFirstValues[:])
	default:
		return nil
	}
}

func collateTag(dataLocale string, cfg config) language.Tag {
	tag, _ := language.Parse(dataLocale)
	if cfg.collation != "" && cfg.collation != "default" {
		if collated, err := tag.SetTypeForKey("co", cfg.collation); err == nil {
			tag = collated
		}
	}
	return tag
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
