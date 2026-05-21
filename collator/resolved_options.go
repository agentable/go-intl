package collator

import "github.com/agentable/go-intl/locale"

type ResolvedOptions struct {
	Locale            locale.Locale `json:"locale"`
	Usage             Usage         `json:"usage"`
	Sensitivity       Sensitivity   `json:"sensitivity"`
	CaseFirst         CaseFirst     `json:"caseFirst"`
	Collation         string        `json:"collation,omitempty"`
	Numeric           bool          `json:"numeric"`
	IgnorePunctuation bool          `json:"ignorePunctuation"`
}

func (f *Collator) ResolvedOptions() ResolvedOptions {
	return f.resolved
}
