package durationformat

import (
	"slices"
	"sync"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/cldr"
	cldrlist "github.com/agentable/go-intl/internal/cldr/list"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	cldrplural "github.com/agentable/go-intl/internal/cldr/plural"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/locale"
)

func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error) {
	return ecma402.SupportedLocalesOf("durationformat", supportedLocales(), locales, string(opts.LocaleMatcher), cldrlocale.Maximize, intlerr.ErrInvalidOption)
}

var supportedLocales = sync.OnceValue(func() []string {
	lists := cldrlist.SupportedLocales()
	plurals := cldrplural.SupportedLocales()
	units := cldr.UnitSupportedLocales()

	out := slices.Clone(cldrnumber.SupportedLocales())
	return slices.DeleteFunc(out, func(loc string) bool {
		return !slices.Contains(lists, loc) ||
			!slices.Contains(plurals, loc) ||
			!slices.Contains(units, loc)
	})
})
