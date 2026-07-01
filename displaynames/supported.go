package displaynames

import (
	cldrdn "github.com/agentable/go-intl/internal/cldr/displaynames"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/locale"
)

func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error) {
	return ecma402.SupportedLocalesOf(ecma402.SupportedLocalesOptions{
		Owner:         displayNamesOwner,
		Supported:     cldrdn.SupportedLocales(),
		Requested:     locales,
		LocaleMatcher: opts.LocaleMatcher,
		Maximizer:     cldrlocale.Maximize,
	})
}
