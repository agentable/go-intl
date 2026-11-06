package listformat

import (
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/locale"
)

func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error) {
	return ecma402.SupportedLocalesOf(ecma402.SupportedLocalesOptions{
		Owner:         listFormatOwner,
		Matcher:       listLocaleMatcher(),
		Requested:     locales,
		LocaleMatcher: opts.LocaleMatcher,
	})
}
