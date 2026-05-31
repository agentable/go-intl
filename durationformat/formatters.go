package durationformat

import (
	"fmt"

	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
)

type durationFormatters struct {
	list            *listformat.ListFormat
	unit            [unitCount]durationNumberFormatters
	unitFraction    [unitCount]durationNumberFormatters
	numeric         [unitCount]durationNumberFormatters
	numericFraction [unitCount]durationNumberFormatters
}

type durationNumberFormatters struct {
	signVisible *numberformat.NumberFormat
	signHidden  *numberformat.NumberFormat
}

func (f durationNumberFormatters) formatter(signDisplayed bool) *numberformat.NumberFormat {
	if !signDisplayed {
		return f.signHidden
	}
	return f.signVisible
}

func buildDurationFormatters(resolved ResolvedOptions, unitOptions [unitCount]resolvedUnitConfig) (durationFormatters, error) {
	var out durationFormatters
	locales := locale.List{resolved.Locale}

	listStyle := listformat.Style(resolved.Style)
	if resolved.Style == DigitalStyle {
		listStyle = listformat.ShortStyle
	}
	list, err := listformat.New(locales, listformat.Options{Type: listformat.Unit, Style: listStyle})
	if err != nil {
		return out, fmt.Errorf("durationformat: construct list formatter: %w", err)
	}
	out.list = list

	for _, spec := range durationUnitSpecs {
		opt := unitOptions[spec.index]
		switch opt.style {
		case LongUnitStyle, ShortUnitStyle, NarrowUnitStyle:
			pair, err := newDurationUnitFormatters(locales, resolved, spec, opt, false)
			if err != nil {
				return out, err
			}
			out.unit[spec.index] = pair
			if durationNextUnitFractional(unitOptions, spec.index) {
				pair, err := newDurationUnitFormatters(locales, resolved, spec, opt, true)
				if err != nil {
					return out, err
				}
				out.unitFraction[spec.index] = pair
			}
		case NumericUnitStyle, TwoDigitUnitStyle:
			pair, err := newDurationNumericFormatters(locales, resolved, opt, false)
			if err != nil {
				return out, err
			}
			out.numeric[spec.index] = pair
			if spec.index == secondsIndex {
				pair, err := newDurationNumericFormatters(locales, resolved, opt, true)
				if err != nil {
					return out, err
				}
				out.numericFraction[spec.index] = pair
			}
		case fractionalUnitStyle:
		}
	}
	return out, nil
}

func newDurationUnitFormatters(locales locale.List, resolved ResolvedOptions, spec durationUnitSpec, opt resolvedUnitConfig, fractional bool) (durationNumberFormatters, error) {
	options := numberformat.Options{
		Style:           numberformat.UnitStyle,
		Unit:            numberformat.UnitIdentifier(string(spec.formatUnit)),
		UnitDisplay:     numberformat.UnitDisplay(publicUnitStyle(opt.style)),
		NumberingSystem: resolved.NumberingSystem,
	}
	if fractional {
		setDurationFractionDigits(&options, resolved)
		options.RoundingMode = numberformat.TruncRoundingMode
	}
	return newDurationNumberFormatters(locales, options, spec.unit)
}

func newDurationNumericFormatters(locales locale.List, resolved ResolvedOptions, opt resolvedUnitConfig, fractional bool) (durationNumberFormatters, error) {
	options := numberformat.Options{
		NumberingSystem: resolved.NumberingSystem,
		UseGrouping:     numberformat.UseGroupingFalse,
	}
	if opt.style == TwoDigitUnitStyle {
		minimumIntegerDigits := 2
		options.MinimumIntegerDigits = &minimumIntegerDigits
	}
	if fractional {
		setDurationFractionDigits(&options, resolved)
		options.RoundingMode = numberformat.TruncRoundingMode
	}
	return newDurationNumberFormatters(locales, options, "numeric")
}

func newDurationNumberFormatters(locales locale.List, options numberformat.Options, name string) (durationNumberFormatters, error) {
	signVisible, err := numberformat.New(locales, options)
	if err != nil {
		return durationNumberFormatters{}, fmt.Errorf("durationformat: construct number formatter for %s: %w", name, err)
	}

	options.SignDisplay = numberformat.NeverSignDisplay
	signHidden, err := numberformat.New(locales, options)
	if err != nil {
		return durationNumberFormatters{}, fmt.Errorf("durationformat: construct sign-hidden number formatter for %s: %w", name, err)
	}
	return durationNumberFormatters{signVisible: signVisible, signHidden: signHidden}, nil
}

func setDurationFractionDigits(options *numberformat.Options, resolved ResolvedOptions) {
	minimumFractionDigits := 0
	maximumFractionDigits := 9
	if resolved.FractionalDigits != nil {
		minimumFractionDigits = *resolved.FractionalDigits
		maximumFractionDigits = *resolved.FractionalDigits
	}
	options.MinimumFractionDigits = &minimumFractionDigits
	options.MaximumFractionDigits = &maximumFractionDigits
}
