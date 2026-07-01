package collator

import (
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrcoll "github.com/agentable/go-intl/internal/collation"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/locale"
)

func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error) {
	return ecma402.SupportedLocalesOf(ecma402.SupportedLocalesOptions{
		Owner:         collatorOwner,
		Supported:     cldrcoll.SupportedLocales(),
		Requested:     locales,
		LocaleMatcher: opts.LocaleMatcher,
		Maximizer:     cldrlocale.Maximize,
	})
}
