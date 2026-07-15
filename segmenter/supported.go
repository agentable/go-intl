package segmenter

import (
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/ecma402"
	cldrseg "github.com/agentable/go-intl/internal/segmentation"
	"github.com/agentable/go-intl/locale"
)

func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error) {
	return ecma402.SupportedLocalesOf(ecma402.SupportedLocalesOptions{
		Owner:         segmenterOwner,
		Supported:     cldrseg.SupportedLocales(),
		Requested:     locales,
		LocaleMatcher: opts.LocaleMatcher,
		Maximizer:     cldrlocale.Maximize,
		// A best-fit fallback to an allowlisted language does not prove that the
		// requested language's dictionary/CJK boundaries are implemented.
		RequireLookupMatch: true,
	})
}
