package durationformat

import (
	"fmt"

	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
)

type durationNumberFormatters struct {
	signVisible *numberformat.NumberFormat
	signHidden  *numberformat.NumberFormat
}

func buildDurationFormatters(format *DurationFormat) error {
	locales := locale.List{format.resolved.Locale}

	listStyle := listformat.Style(format.resolved.Style)
	if format.resolved.Style == DigitalStyle {
		listStyle = listformat.ShortStyle
	}
	list, err := listformat.New(locales, listformat.Options{
		Type:  optionString(listformat.Unit),
		Style: optionString(listStyle),
	})
	if err != nil {
		return fmt.Errorf("durationformat: construct list formatter: %w", err)
	}
	format.listFormatter = list

	for _, spec := range durationUnitSpecs[:] {
		opt := format.unitOptions[spec.index]
		switch opt.style {
		case LongUnitStyle, ShortUnitStyle, NarrowUnitStyle:
			options := durationUnitNumberOptions(format.resolved, spec, opt)
			pair, err := newDurationNumberFormatters(locales, options, spec.unit)
			if err != nil {
				return err
			}
			format.unitFormatters[spec.index] = pair
			if durationNextUnitFractional(format.unitOptions, spec) {
				pair, err := newDurationNumberFormatters(locales, durationFractionalNumberOptions(options, format.resolved), spec.unit)
				if err != nil {
					return err
				}
				format.unitFractionFormatters[spec.index] = pair
			}
		case NumericUnitStyle, TwoDigitUnitStyle:
			options := durationNumericNumberOptions(format.resolved, opt)
			pair, err := newDurationNumberFormatters(locales, options, "numeric")
			if err != nil {
				return err
			}
			format.numericFormatters[spec.index] = pair
			if spec.index == secondsIndex {
				pair, err := newDurationNumberFormatters(locales, durationFractionalNumberOptions(options, format.resolved), "numeric")
				if err != nil {
					return err
				}
				format.secondsNumericFractionFormatter = pair
			}
		case fractionalUnitStyle:
		}
	}
	return nil
}

func durationUnitNumberOptions(resolved ResolvedOptions, spec durationUnitSpec, opt resolvedUnitConfig) numberformat.Options {
	return numberformat.Options{
		Style:           optionString(numberformat.UnitStyle),
		Unit:            optionString(spec.formatUnit),
		UnitDisplay:     optionString(numberformat.UnitDisplay(publicUnitStyle(opt.style))),
		NumberingSystem: optionString(resolved.NumberingSystem),
	}
}

func durationNumericNumberOptions(resolved ResolvedOptions, opt resolvedUnitConfig) numberformat.Options {
	options := numberformat.Options{
		NumberingSystem: optionString(resolved.NumberingSystem),
		UseGrouping:     optionString(numberformat.UseGroupingFalse),
	}
	if opt.style == TwoDigitUnitStyle {
		minimumIntegerDigits := 2
		options.MinimumIntegerDigits = &minimumIntegerDigits
	}
	return options
}

func newDurationNumberFormatters(locales locale.List, options numberformat.Options, name string) (durationNumberFormatters, error) {
	signVisible, err := numberformat.New(locales, options)
	if err != nil {
		return durationNumberFormatters{}, fmt.Errorf("durationformat: construct number formatter for %s: %w", name, err)
	}

	signHiddenOptions := options
	signHiddenOptions.SignDisplay = optionString(numberformat.NeverSignDisplay)
	signHidden, err := numberformat.New(locales, signHiddenOptions)
	if err != nil {
		return durationNumberFormatters{}, fmt.Errorf("durationformat: construct sign-hidden number formatter for %s: %w", name, err)
	}
	return durationNumberFormatters{signVisible: signVisible, signHidden: signHidden}, nil
}

func durationFractionalNumberOptions(options numberformat.Options, resolved ResolvedOptions) numberformat.Options {
	minimumFractionDigits := 0
	maximumFractionDigits := 9
	if resolved.FractionalDigits != nil {
		minimumFractionDigits = *resolved.FractionalDigits
		maximumFractionDigits = *resolved.FractionalDigits
	}
	options.MinimumFractionDigits = &minimumFractionDigits
	options.MaximumFractionDigits = &maximumFractionDigits
	options.RoundingMode = optionString(numberformat.TruncRoundingMode)
	return options
}

func optionString[T ~string](value T) *string {
	out := string(value)
	return &out
}
