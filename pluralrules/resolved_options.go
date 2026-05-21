package pluralrules

import (
	"slices"

	"github.com/agentable/go-intl/internal/cldr/plural"
	"github.com/agentable/go-intl/locale"
)

type ResolvedOptions struct {
	Locale                   locale.Locale       `json:"locale"`
	Type                     Type                `json:"type"`
	MinimumIntegerDigits     int                 `json:"minimumIntegerDigits"`
	MinimumFractionDigits    int                 `json:"minimumFractionDigits"`
	MaximumFractionDigits    int                 `json:"maximumFractionDigits"`
	MinimumSignificantDigits int                 `json:"minimumSignificantDigits,omitempty"`
	MaximumSignificantDigits int                 `json:"maximumSignificantDigits,omitempty"`
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
	minFracDigits, maxFracDigits := f.cfg.minFracDigits, f.cfg.maxFracDigits
	if f.cfg.roundingPriority == "auto" && (f.cfg.hasMinSigDigits || f.cfg.hasMaxSigDigits) {
		minFracDigits, maxFracDigits = 0, 0
	}
	return ResolvedOptions{
		Locale:                   f.loc,
		Type:                     f.cfg.typ,
		MinimumIntegerDigits:     f.cfg.minIntDigits,
		MinimumFractionDigits:    minFracDigits,
		MaximumFractionDigits:    maxFracDigits,
		MinimumSignificantDigits: f.cfg.minSigDigits,
		MaximumSignificantDigits: f.cfg.maxSigDigits,
		PluralCategories:         slices.Clone(categories),
		Notation:                 Notation(f.cfg.notation),
		CompactDisplay:           CompactDisplay(f.cfg.compactDisplay),
		RoundingIncrement:        f.cfg.roundingIncrement,
		RoundingMode:             RoundingMode(f.cfg.roundingMode),
		RoundingPriority:         RoundingPriority(f.cfg.roundingPriority),
		TrailingZeroDisplay:      TrailingZeroDisplay(f.cfg.trailingZeroDisplay),
	}
}
