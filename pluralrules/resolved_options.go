package pluralrules

import (
	"slices"

	"github.com/agentable/go-intl/internal/cldr/plural"
	"github.com/agentable/go-intl/locale"
)

type ResolvedOptions struct {
	Locale                   locale.Locale
	Type                     Type
	MinimumIntegerDigits     int
	MinimumFractionDigits    int
	MaximumFractionDigits    int
	MinimumSignificantDigits int
	MaximumSignificantDigits int
	PluralCategories         []Category
	Notation                 Notation
	CompactDisplay           CompactDisplay
	RoundingIncrement        int
	RoundingMode             RoundingMode
	RoundingPriority         RoundingPriority
	TrailingZeroDisplay      TrailingZeroDisplay
}

func (r *PluralRules) ResolvedOptions() ResolvedOptions {
	categories := plural.Categories(r.loc.Tag().String(), r.cfg.typ.String())
	minFracDigits, maxFracDigits := r.cfg.minFracDigits, r.cfg.maxFracDigits
	if r.cfg.roundingPriority == "auto" && (r.cfg.hasMinSigDigits || r.cfg.hasMaxSigDigits) {
		minFracDigits, maxFracDigits = 0, 0
	}
	return ResolvedOptions{
		Locale:                   r.loc,
		Type:                     r.cfg.typ,
		MinimumIntegerDigits:     r.cfg.minIntDigits,
		MinimumFractionDigits:    minFracDigits,
		MaximumFractionDigits:    maxFracDigits,
		MinimumSignificantDigits: r.cfg.minSigDigits,
		MaximumSignificantDigits: r.cfg.maxSigDigits,
		PluralCategories:         slices.Clone(categories),
		Notation:                 Notation(r.cfg.notation),
		CompactDisplay:           CompactDisplay(r.cfg.compactDisplay),
		RoundingIncrement:        r.cfg.roundingIncrement,
		RoundingMode:             RoundingMode(r.cfg.roundingMode),
		RoundingPriority:         RoundingPriority(r.cfg.roundingPriority),
		TrailingZeroDisplay:      TrailingZeroDisplay(r.cfg.trailingZeroDisplay),
	}
}
