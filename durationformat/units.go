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

const durationCalendarUnitMaxAbsExclusive uint64 = 1 << 32

const (
	nanosecondsPerDay         int64 = 86_400_000_000_000
	nanosecondsPerHour        int64 = 3_600_000_000_000
	nanosecondsPerMinute      int64 = 60_000_000_000
	nanosecondsPerSecond      int64 = 1_000_000_000
	nanosecondsPerMillisecond int64 = 1_000_000
	nanosecondsPerMicrosecond int64 = 1_000
	nanosecondsPerNanosecond  int64 = 1
)

type durationUnitSpec struct {
	index              unitIndex
	unit               string
	formatUnit         Unit
	styleSet           durationUnitStyleSet
	digitalDefault     UnitStyle
	maxAbsExclusive    uint64
	nanosecondsPerUnit int64
	fractional         bool
	clockUnit          bool
	clockContinuation  bool
	fractionalChild    unitIndex
	hasFractionalChild bool
	trackPrevious      bool
}

type durationUnitStyleSet uint8

const (
	durationDateUnitStyleSet durationUnitStyleSet = iota
	durationTimeUnitStyleSet
	durationSubsecondUnitStyleSet
)

var durationUnitSpecs = [...]durationUnitSpec{
	{index: yearsIndex, unit: "years", formatUnit: Year, styleSet: durationDateUnitStyleSet, digitalDefault: ShortUnitStyle, maxAbsExclusive: durationCalendarUnitMaxAbsExclusive},
	{index: monthsIndex, unit: "months", formatUnit: Month, styleSet: durationDateUnitStyleSet, digitalDefault: ShortUnitStyle, maxAbsExclusive: durationCalendarUnitMaxAbsExclusive},
	{index: weeksIndex, unit: "weeks", formatUnit: Week, styleSet: durationDateUnitStyleSet, digitalDefault: ShortUnitStyle, maxAbsExclusive: durationCalendarUnitMaxAbsExclusive},
	{index: daysIndex, unit: "days", formatUnit: Day, styleSet: durationDateUnitStyleSet, digitalDefault: ShortUnitStyle, nanosecondsPerUnit: nanosecondsPerDay},
	{index: hoursIndex, unit: "hours", formatUnit: Hour, styleSet: durationTimeUnitStyleSet, digitalDefault: NumericUnitStyle, nanosecondsPerUnit: nanosecondsPerHour, clockUnit: true, trackPrevious: true},
	{index: minutesIndex, unit: "minutes", formatUnit: Minute, styleSet: durationTimeUnitStyleSet, digitalDefault: NumericUnitStyle, nanosecondsPerUnit: nanosecondsPerMinute, clockUnit: true, clockContinuation: true, trackPrevious: true},
	{index: secondsIndex, unit: "seconds", formatUnit: Second, styleSet: durationTimeUnitStyleSet, digitalDefault: NumericUnitStyle, nanosecondsPerUnit: nanosecondsPerSecond, clockUnit: true, clockContinuation: true, fractionalChild: millisecondsIndex, hasFractionalChild: true, trackPrevious: true},
	{index: millisecondsIndex, unit: "milliseconds", formatUnit: Millisecond, styleSet: durationSubsecondUnitStyleSet, digitalDefault: NumericUnitStyle, nanosecondsPerUnit: nanosecondsPerMillisecond, fractional: true, fractionalChild: microsecondsIndex, hasFractionalChild: true, trackPrevious: true},
	{index: microsecondsIndex, unit: "microseconds", formatUnit: Microsecond, styleSet: durationSubsecondUnitStyleSet, digitalDefault: NumericUnitStyle, nanosecondsPerUnit: nanosecondsPerMicrosecond, fractional: true, fractionalChild: nanosecondsIndex, hasFractionalChild: true, trackPrevious: true},
	{index: nanosecondsIndex, unit: "nanoseconds", formatUnit: Nanosecond, styleSet: durationSubsecondUnitStyleSet, digitalDefault: NumericUnitStyle, nanosecondsPerUnit: nanosecondsPerNanosecond, fractional: true},
}
