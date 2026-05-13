package datetimeformat

func HasTimeZoneOption(opts ...Options) bool {
	cfg := defaultConfig()
	for _, opt := range opts {
		applyOptions(&cfg, opt)
	}
	return cfg.timeZone != ""
}
