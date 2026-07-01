package durationformat

import (
	"sync"

	cldrlist "github.com/agentable/go-intl/internal/cldr/list"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	cldrplural "github.com/agentable/go-intl/internal/cldr/plural"
	cldrunit "github.com/agentable/go-intl/internal/cldr/unit"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/locale"
)

func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error) {
	return ecma402.SupportedLocalesOf(ecma402.SupportedLocalesOptions{
		Owner:         durationFormatOwner,
		Supported:     supportedLocales(),
		Requested:     locales,
		LocaleMatcher: opts.LocaleMatcher,
		Maximizer:     cldrlocale.Maximize,
	})
}

var supportedLocales = sync.OnceValue(func() []string {
	return cldrlocale.IntersectSupportedLocales(
		cldrnumber.SupportedLocales(),
		cldrlist.SupportedLocales(),
		cldrplural.SupportedLocales(),
		cldrunit.SupportedLocales(),
	)
})
