package datetimeformat

import "github.com/agentable/go-intl/internal/cachekey"

func CanonicalKey(opts ...Options) string {
	cfg := defaultConfig()
	for _, opt := range opts {
		applyOptions(&cfg, opt)
	}
	def := defaultConfig()
	parts := make([]string, 0, 18)
	parts = cachekey.AppendNonEmptyString(parts, "calendar", cfg.calendar)
	parts = cachekey.AppendNonEmptyString(parts, "dateStyle", cfg.dateStyle)
	parts = cachekey.AppendNonEmptyString(parts, "day", cfg.day)
	parts = cachekey.AppendNonEmptyString(parts, "dayPeriod", cfg.dayPeriod)
	parts = cachekey.AppendNonEmptyString(parts, "era", cfg.era)
	if cfg.hasFractionalSecondDigits {
		parts = cachekey.AppendInt(parts, "fractionalSecondDigits", cfg.fractionalSecondDigits, 0)
	}
	parts = cachekey.AppendString(parts, "formatMatcher", cfg.formatMatcher, def.formatMatcher)
	parts = cachekey.AppendNonEmptyString(parts, "hour", cfg.hour)
	if cfg.hasHour12 {
		parts = cachekey.AppendBool(parts, "hour12", cfg.hour12)
	}
	parts = cachekey.AppendNonEmptyString(parts, "hourCycle", cfg.hourCycle)
	parts = cachekey.AppendString(parts, "localeMatcher", cfg.localeMatcher, def.localeMatcher)
	parts = cachekey.AppendNonEmptyString(parts, "minute", cfg.minute)
	parts = cachekey.AppendNonEmptyString(parts, "month", cfg.month)
	parts = cachekey.AppendNonEmptyString(parts, "numberingSystem", cfg.numberingSystem)
	parts = cachekey.AppendNonEmptyString(parts, "second", cfg.second)
	parts = cachekey.AppendNonEmptyString(parts, "timeStyle", cfg.timeStyle)
	parts = cachekey.AppendNonEmptyString(parts, "timeZone", cfg.timeZone)
	parts = cachekey.AppendNonEmptyString(parts, "timeZoneName", cfg.timeZoneName)
	parts = cachekey.AppendNonEmptyString(parts, "weekday", cfg.weekday)
	parts = cachekey.AppendNonEmptyString(parts, "year", cfg.year)
	return cachekey.Join(parts)
}
