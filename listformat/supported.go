package listformat

import (
	cldrlist "github.com/agentable/go-intl/internal/cldr/list"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/locale"
)

func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error) {
	return ecma402.SupportedLocalesOf(ecma402.SupportedLocalesOptions{
		Owner:         listFormatOwner,
		Supported:     cldrlist.SupportedLocales(),
		Requested:     locales,
		LocaleMatcher: opts.LocaleMatcher,
		Maximizer:     cldrlist.Maximize,
	})
}
