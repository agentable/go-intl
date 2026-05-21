package segmenter

import "github.com/agentable/go-intl/locale"

type ResolvedOptions struct {
	// Locale is the resolved locale. Mirrors Intl.Segmenter resolved option "locale".
	Locale locale.Locale `json:"locale"`
	// Granularity is the resolved segmentation granularity. Mirrors Intl.Segmenter resolved option "granularity".
	Granularity Granularity `json:"granularity"`
}

func (f *Segmenter) ResolvedOptions() ResolvedOptions {
	return f.resolved
}
