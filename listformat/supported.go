package listformat

import (
	cldrlist "github.com/agentable/go-intl/internal/cldr/list"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/locale"
)

func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error) {
	return ecma402.SupportedLocalesOf("listformat", cldrlist.SupportedLocales(), locales, string(opts.LocaleMatcher), cldrlist.Maximize, intlerr.ErrInvalidOption)
}
