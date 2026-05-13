package datetimeformat

import (
	"fmt"

	"github.com/agentable/go-intl/internal/cldr"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

func SupportedLocalesOf(locales []locale.Locale, opts ...Options) ([]locale.Locale, error) {
	matcher, err := supportedLocaleMatcher(opts...)
	if err != nil {
		return nil, err
	}
	return localematcher.FilterLocales(cldr.DateSupportedLocales(), locales, matcher), nil
}

func supportedLocaleMatcher(opts ...Options) (localematcher.Algorithm, error) {
	if len(opts) > 1 {
		return 0, fmt.Errorf("datetimeformat: multiple options: %w", ErrInvalidOption)
	}
	if len(opts) == 0 || opts[0].LocaleMatcher == "" || opts[0].LocaleMatcher == BestFitLocaleMatcher {
		return localematcher.AlgorithmBestFit, nil
	}
	if opts[0].LocaleMatcher == LookupLocaleMatcher {
		return localematcher.AlgorithmLookup, nil
	}
	return 0, invalidOption("localeMatcher", string(opts[0].LocaleMatcher), locale.Locale{})
}
