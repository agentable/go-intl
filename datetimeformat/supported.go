package datetimeformat

import (
	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/locale"
)

func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error) {
	supported, err := ecma402.SupportedLocalesOf("datetimeformat", cldrdate.SupportedLocales(), locales, string(opts.LocaleMatcher), cldrlocale.Maximize, intlerr.ErrInvalidOption)
	if err != nil {
		return nil, err
	}
	return filterSupportedLocaleCalendars(supported), nil
}

func filterSupportedLocaleCalendars(locales locale.List) locale.List {
	out := make(locale.List, 0, len(locales))
	for _, loc := range locales {
		if calendar := loc.Calendar(); calendar != "" && !isSupportedCalendar(calendar) {
			continue
		}
		out = append(out, loc)
	}
	return out
}
