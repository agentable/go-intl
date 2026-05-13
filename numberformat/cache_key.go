package numberformat

import "github.com/agentable/go-intl/internal/cachekey"

func CanonicalKey(opts ...Options) string {
	cfg := defaultConfig()
	for _, opt := range opts {
		applyOptions(&cfg, opt)
	}
	parts := make([]string, 0, 19)
	def := defaultConfig()
	parts = cachekey.AppendString(parts, "compactDisplay", cfg.compactDisplay, def.compactDisplay)
	parts = cachekey.AppendString(parts, "currency", cfg.currency, def.currency)
	parts = cachekey.AppendString(parts, "currencyDisplay", cfg.currencyDisplay, def.currencyDisplay)
	parts = cachekey.AppendString(parts, "currencySign", cfg.currencySign, def.currencySign)
	parts = cachekey.AppendString(parts, "localeMatcher", cfg.localeMatcher, def.localeMatcher)
	parts = cachekey.AppendInt(parts, "maximumFractionDigits", cfg.maxFracDigits, def.maxFracDigits)
	if cfg.hasMaxSigDigits {
		parts = cachekey.AppendInt(parts, "maximumSignificantDigits", cfg.maxSigDigits, def.maxSigDigits)
	}
	parts = cachekey.AppendInt(parts, "minimumFractionDigits", cfg.minFracDigits, def.minFracDigits)
	parts = cachekey.AppendInt(parts, "minimumIntegerDigits", cfg.minIntDigits, def.minIntDigits)
	if cfg.hasMinSigDigits {
		parts = cachekey.AppendInt(parts, "minimumSignificantDigits", cfg.minSigDigits, def.minSigDigits)
	}
	parts = cachekey.AppendString(parts, "notation", cfg.notation, def.notation)
	parts = cachekey.AppendString(parts, "numberingSystem", cfg.numberingSystem, def.numberingSystem)
	parts = cachekey.AppendInt(parts, "roundingIncrement", cfg.roundingIncrement, def.roundingIncrement)
	parts = cachekey.AppendString(parts, "roundingMode", cfg.roundingMode, def.roundingMode)
	parts = cachekey.AppendString(parts, "roundingPriority", cfg.roundingPriority, def.roundingPriority)
	parts = cachekey.AppendString(parts, "signDisplay", cfg.signDisplay, def.signDisplay)
	parts = cachekey.AppendString(parts, "style", cfg.style, def.style)
	parts = cachekey.AppendString(parts, "trailingZeroDisplay", cfg.trailingZeroDisplay, def.trailingZeroDisplay)
	parts = cachekey.AppendString(parts, "unit", cfg.unit, def.unit)
	parts = cachekey.AppendString(parts, "unitDisplay", cfg.unitDisplay, def.unitDisplay)
	parts = cachekey.AppendString(parts, "useGrouping", cfg.useGrouping, def.useGrouping)
	return cachekey.Join(parts)
}
