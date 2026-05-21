package collator

import (
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrcoll "github.com/agentable/go-intl/internal/collation"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/locale"
)

func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error) {
	supported, err := ecma402.SupportedLocalesOf("collator", cldrcoll.SupportedLocales(), locales, string(opts.LocaleMatcher), cldrlocale.Maximize, intlerr.ErrInvalidOption)
	if err != nil {
		return nil, err
	}
	return filterSupportedLocaleExtensions(supported), nil
}

func filterSupportedLocaleExtensions(locales locale.List) locale.List {
	out := make(locale.List, 0, len(locales))
	for _, loc := range locales {
		if !supportedLocaleExtensions(loc) {
			continue
		}
		out = append(out, loc)
	}
	return out
}

func supportedLocaleExtensions(loc locale.Locale) bool {
	if collation := loc.Collation(); collation != "" && !isDefaultCollation(collation) {
		return false
	}
	if caseFirst := loc.CaseFirst(); caseFirst != "" && caseFirst != string(FalseCaseFirst) {
		return false
	}
	return true
}
