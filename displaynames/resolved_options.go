package displaynames

import "github.com/agentable/go-intl/locale"

type ResolvedOptions struct {
	Locale   locale.Locale `json:"locale"`
	Style    Style         `json:"style"`
	Type     Type          `json:"type"`
	Fallback Fallback      `json:"fallback"`
	// LanguageDisplay is nil when Type != Language. ECMA-402 §12.4.2
	// (resolvedOptions) only writes this property for Language-typed
	// DisplayNames instances; the pointer makes the omission unambiguous.
	LanguageDisplay *LanguageDisplay `json:"languageDisplay,omitempty"`
}

func (d *DisplayNames) ResolvedOptions() ResolvedOptions {
	resolved := d.resolved
	if resolved.LanguageDisplay != nil {
		languageDisplay := *resolved.LanguageDisplay
		resolved.LanguageDisplay = &languageDisplay
	}
	return resolved
}
