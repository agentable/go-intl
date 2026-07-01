package numberformat

import (
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/locale"
)

func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error) {
	return ecma402.SupportedLocalesOf(ecma402.SupportedLocalesOptions{
		Owner:         numberFormatOwner,
		Supported:     cldrnumber.SupportedLocales(),
		Requested:     locales,
		LocaleMatcher: opts.LocaleMatcher,
		Maximizer:     cldrlocale.Maximize,
	})
}
