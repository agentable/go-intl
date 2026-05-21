package segmenter

import (
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intlerr"
	cldrseg "github.com/agentable/go-intl/internal/segmentation"
	"github.com/agentable/go-intl/locale"
)

func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error) {
	return ecma402.SupportedLocalesOf("segmenter", cldrseg.SupportedLocales(), locales, string(opts.LocaleMatcher), cldrlocale.Maximize, intlerr.ErrInvalidOption)
}
