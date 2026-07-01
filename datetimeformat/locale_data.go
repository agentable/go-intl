package datetimeformat

import (
	"slices"
	"sync"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
)

type dateLocaleData struct{}

type dateLocaleDataCacheKey struct {
	locale string
	key    string
}

var dateLocaleDataCache sync.Map

func (dateLocaleData) For(locale, key string) []string {
	cacheKey := dateLocaleDataCacheKey{locale: locale, key: key}
	if data, ok := dateLocaleDataCache.Load(cacheKey); ok {
		return slices.Clone(data.([]string))
	}
	data := cldrdate.DateLocaleData{}.For(locale, key)
	actual, _ := dateLocaleDataCache.LoadOrStore(cacheKey, data)
	return slices.Clone(actual.([]string))
}
