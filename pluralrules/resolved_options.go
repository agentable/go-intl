package pluralrules

import (
	"slices"

	"github.com/agentable/go-intl/internal/cldr/plural"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
	"github.com/agentable/go-intl/locale"
)

type ResolvedOptions struct {
	Locale                   locale.Locale       `json:"locale"`
	Type                     Type                `json:"type"`
	MinimumIntegerDigits     int                 `json:"minimumIntegerDigits"`
	MinimumFractionDigits    *int                `json:"minimumFractionDigits,omitempty"`
	MaximumFractionDigits    *int                `json:"maximumFractionDigits,omitempty"`
	MinimumSignificantDigits *int                `json:"minimumSignificantDigits,omitempty"`
	MaximumSignificantDigits *int                `json:"maximumSignificantDigits,omitempty"`
	PluralCategories         []Category          `json:"pluralCategories"`
	Notation                 Notation            `json:"notation"`
	CompactDisplay           CompactDisplay      `json:"compactDisplay"`
	RoundingIncrement        int                 `json:"roundingIncrement"`
	RoundingMode             RoundingMode        `json:"roundingMode"`
	RoundingPriority         RoundingPriority    `json:"roundingPriority"`
	TrailingZeroDisplay      TrailingZeroDisplay `json:"trailingZeroDisplay"`
}

func (f *PluralRules) ResolvedOptions() ResolvedOptions {
	categories := plural.Categories(f.loc.Tag().String(), f.cfg.typ.String())
	resolved := ResolvedOptions{
		Locale:               f.loc,
		Type:                 f.cfg.typ,
		MinimumIntegerDigits: f.cfg.minIntDigits,
		PluralCategories:     slices.Clone(categories),
		Notation:             Notation(f.cfg.notation),
		CompactDisplay:       CompactDisplay(f.cfg.compactDisplay),
		RoundingIncrement:    f.cfg.roundingIncrement,
		RoundingMode:         RoundingMode(f.cfg.roundingMode),
		RoundingPriority:     RoundingPriority(f.cfg.roundingPriority),
		TrailingZeroDisplay:  TrailingZeroDisplay(f.cfg.trailingZeroDisplay),
	}
	switch f.cfg.roundingType {
	case ecma402nf.RoundingTypeFractionDigits:
		resolved.MinimumFractionDigits = resolvedInt(f.cfg.minFracDigits)
		resolved.MaximumFractionDigits = resolvedInt(f.cfg.maxFracDigits)
	case ecma402nf.RoundingTypeSignificantDigits:
		resolved.MinimumSignificantDigits = resolvedInt(f.cfg.minSigDigits)
		resolved.MaximumSignificantDigits = resolvedInt(f.cfg.maxSigDigits)
	default:
		resolved.MinimumFractionDigits = resolvedInt(f.cfg.minFracDigits)
		resolved.MaximumFractionDigits = resolvedInt(f.cfg.maxFracDigits)
		resolved.MinimumSignificantDigits = resolvedInt(f.cfg.minSigDigits)
		resolved.MaximumSignificantDigits = resolvedInt(f.cfg.maxSigDigits)
	}
	return resolved
}

func resolvedInt(v int) *int {
	return &v
}
