package durationformat

type unitIndex int

const (
	yearsIndex unitIndex = iota
	monthsIndex
	weeksIndex
	daysIndex
	hoursIndex
	minutesIndex
	secondsIndex
	millisecondsIndex
	microsecondsIndex
	nanosecondsIndex
	unitCount
)

type durationUnitSpec struct {
	index          unitIndex
	unit           string
	formatUnit     Unit
	styles         []UnitStyle
	digitalDefault UnitStyle
	fractional     bool
	trackPrevious  bool
}

var (
	durationDateUnitStyles      = []UnitStyle{LongUnitStyle, ShortUnitStyle, NarrowUnitStyle}
	durationTimeUnitStyles      = []UnitStyle{LongUnitStyle, ShortUnitStyle, NarrowUnitStyle, NumericUnitStyle, TwoDigitUnitStyle}
	durationSubsecondUnitStyles = []UnitStyle{LongUnitStyle, ShortUnitStyle, NarrowUnitStyle, NumericUnitStyle}
	durationUnitSpecs           = []durationUnitSpec{
		{index: yearsIndex, unit: "years", formatUnit: Year, styles: durationDateUnitStyles, digitalDefault: ShortUnitStyle},
		{index: monthsIndex, unit: "months", formatUnit: Month, styles: durationDateUnitStyles, digitalDefault: ShortUnitStyle},
		{index: weeksIndex, unit: "weeks", formatUnit: Week, styles: durationDateUnitStyles, digitalDefault: ShortUnitStyle},
		{index: daysIndex, unit: "days", formatUnit: Day, styles: durationDateUnitStyles, digitalDefault: ShortUnitStyle},
		{index: hoursIndex, unit: "hours", formatUnit: Hour, styles: durationTimeUnitStyles, digitalDefault: NumericUnitStyle, trackPrevious: true},
		{index: minutesIndex, unit: "minutes", formatUnit: Minute, styles: durationTimeUnitStyles, digitalDefault: NumericUnitStyle, trackPrevious: true},
		{index: secondsIndex, unit: "seconds", formatUnit: Second, styles: durationTimeUnitStyles, digitalDefault: NumericUnitStyle, trackPrevious: true},
		{index: millisecondsIndex, unit: "milliseconds", formatUnit: Millisecond, styles: durationSubsecondUnitStyles, digitalDefault: NumericUnitStyle, fractional: true, trackPrevious: true},
		{index: microsecondsIndex, unit: "microseconds", formatUnit: Microsecond, styles: durationSubsecondUnitStyles, digitalDefault: NumericUnitStyle, fractional: true, trackPrevious: true},
		{index: nanosecondsIndex, unit: "nanoseconds", formatUnit: Nanosecond, styles: durationSubsecondUnitStyles, digitalDefault: NumericUnitStyle, fractional: true},
	}
)
