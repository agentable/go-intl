package segmenter

import "github.com/agentable/go-intl/locale"

type ResolvedOptions struct {
	Locale      locale.Locale `json:"locale"`
	Granularity Granularity   `json:"granularity"`
}

func (f *Segmenter) ResolvedOptions() ResolvedOptions {
	return f.resolved
}
