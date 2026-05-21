package pluralrules

import (
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/cldr/plural"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/locale"
)

func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error) {
	return ecma402.SupportedLocalesOf("pluralrules", plural.SupportedLocales(), locales, string(opts.LocaleMatcher), cldrlocale.Maximize, intlerr.ErrInvalidOption)
}
