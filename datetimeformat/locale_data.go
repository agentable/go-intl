package datetimeformat

import (
	"sync"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
)

type dateLocaleData struct{}

var dateLocaleDataCache sync.Map

func (dateLocaleData) For(locale, key string) []string {
	cacheKey := [2]string{locale, key}
	if data, ok := dateLocaleDataCache.Load(cacheKey); ok {
		return data.([]string)
	}
	data := cldrdate.DateLocaleData{}.For(locale, key)
	actual, _ := dateLocaleDataCache.LoadOrStore(cacheKey, data)
	return actual.([]string)
}
