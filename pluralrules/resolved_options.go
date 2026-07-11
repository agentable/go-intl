package pluralrules

import (
	"slices"

	"github.com/agentable/go-intl/internal/cldr/plural"
	"github.com/agentable/go-intl/internal/ecma402"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
	pluralop "github.com/agentable/go-intl/internal/plural"
	"github.com/agentable/go-intl/locale"
)

// ResolvedOptions mirrors the ECMA-402 §16.4.5 "Resolved Options of PluralRules
// Instances" table; field declaration order is the observable JSON key order and
// must match that table.
type ResolvedOptions struct {
	Locale                   locale.Locale       `json:"locale"`
	Type                     Type                `json:"type"`
	Notation                 Notation            `json:"notation"`
	CompactDisplay           *CompactDisplay     `json:"compactDisplay,omitempty"`
	MinimumIntegerDigits     int                 `json:"minimumIntegerDigits"`
	MinimumFractionDigits    *int                `json:"minimumFractionDigits,omitempty"`
	MaximumFractionDigits    *int                `json:"maximumFractionDigits,omitempty"`
	MinimumSignificantDigits *int                `json:"minimumSignificantDigits,omitempty"`
	MaximumSignificantDigits *int                `json:"maximumSignificantDigits,omitempty"`
	PluralCategories         []Category          `json:"pluralCategories"`
	RoundingIncrement        int                 `json:"roundingIncrement"`
	RoundingMode             RoundingMode        `json:"roundingMode"`
	RoundingPriority         RoundingPriority    `json:"roundingPriority"`
	TrailingZeroDisplay      TrailingZeroDisplay `json:"trailingZeroDisplay"`
}

func (f *PluralRules) ResolvedOptions() ResolvedOptions {
	resolved := f.resolved
	resolved.MinimumFractionDigits = ecma402.CloneResolvedScalar(resolved.MinimumFractionDigits)
	resolved.MaximumFractionDigits = ecma402.CloneResolvedScalar(resolved.MaximumFractionDigits)
	resolved.MinimumSignificantDigits = ecma402.CloneResolvedScalar(resolved.MinimumSignificantDigits)
	resolved.MaximumSignificantDigits = ecma402.CloneResolvedScalar(resolved.MaximumSignificantDigits)
	resolved.PluralCategories = slices.Clone(resolved.PluralCategories)
	resolved.CompactDisplay = ecma402.CloneResolvedScalar(resolved.CompactDisplay)
	return resolved
}

func resolvedOptionsForPluralRules(loc locale.Locale, cfg config, digits ecma402nf.ResolvedDigitOptions, dataLocale string) ResolvedOptions {
	categories := publicCategories(plural.Categories(dataLocale, cfg.typ))
	digitProperties := digits.ResolvedPluralRulesProperties()
	return ResolvedOptions{
		Locale:                   loc,
		Type:                     Type(cfg.typ),
		Notation:                 Notation(cfg.notation),
		CompactDisplay:           resolvedCompactDisplay(cfg),
		MinimumIntegerDigits:     digits.MinimumIntegerDigits,
		MinimumFractionDigits:    digitProperties.MinimumFractionDigits,
		MaximumFractionDigits:    digitProperties.MaximumFractionDigits,
		MinimumSignificantDigits: digitProperties.MinimumSignificantDigits,
		MaximumSignificantDigits: digitProperties.MaximumSignificantDigits,
		PluralCategories:         categories,
		RoundingIncrement:        digits.RoundingIncrement,
		RoundingMode:             RoundingMode(digits.RoundingMode),
		RoundingPriority:         RoundingPriority(digits.RoundingPriority),
		TrailingZeroDisplay:      TrailingZeroDisplay(digits.TrailingZeroDisplay),
	}
}

func resolvedCompactDisplay(cfg config) *CompactDisplay {
	if cfg.notation != string(CompactNotation) {
		return nil
	}
	return ecma402.ResolvedScalar(CompactDisplay(cfg.compactDisplay))
}

func publicCategories(categories []pluralop.Category) []Category {
	out := make([]Category, len(categories))
	for i, category := range categories {
		out[i] = Category(category)
	}
	return out
}
