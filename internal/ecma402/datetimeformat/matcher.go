package ecma402dtf

import "strings"

// Skeleton matcher penalties mirror ICU DateTimePatternGenerator
// icu4c/source/i18n/dtptngen.cpp computeDistance.
const (
	removalPenalty              = 120 // Penalizes a missing requested field.
	additionPenalty             = 20  // Penalizes an extra field in the candidate pattern.
	longMorePenalty             = 6   // Penalizes a text field that is much wider than requested.
	shortMorePenalty            = 3   // Penalizes a text field that is slightly wider than requested.
	longLessPenalty             = 8   // Penalizes a text field that is much narrower than requested.
	shortLessPenalty            = 6   // Penalizes a text field that is slightly narrower than requested.
	differentNumericTypePenalty = 15  // Penalizes numeric versus two-digit field mismatches.
	offsetPenalty               = 1   // Penalizes a non-exact pattern width offset.
)

// Match selects the best-fit format and adjusts field widths to options.
func Match(opts Options, formats []Formats) Formats {
	return AdjustFieldTypes(MatchBasic(opts, formats), opts)
}

// MatchBasic selects the highest-scoring format without field-width adjustment.
func MatchBasic(opts Options, formats []Formats) Formats {
	if len(formats) == 0 {
		return Formats{}
	}
	bestFormat := formats[0]
	bestScore := scoreFormat(opts, formats[0])
	for _, format := range formats[1:] {
		score := scoreFormat(opts, format)
		if score > bestScore {
			bestScore = score
			bestFormat = format
		}
	}
	return bestFormat
}

// AdjustFieldTypes rewrites selected pattern field widths to match options.
func AdjustFieldTypes(format Formats, opts Options) Formats {
	adjustNumericField(&format, &format.Year, opts.Year, 'y')
	if opts.Month != "" && format.Month != opts.Month {
		if canAdjustFieldStyle(format.Pattern, monthPatternFields, opts.Month) {
			format.Pattern = adjustPatternFields(format.Pattern, monthPatternFields, fieldPatternWidth('M', opts.Month))
			format.Month = opts.Month
		}
	}
	adjustNumericField(&format, &format.Day, opts.Day, 'd')
	if opts.Weekday != "" && format.Weekday != opts.Weekday {
		format.Pattern = adjustPatternFields(format.Pattern, weekdayPatternFields, fieldPatternWidth('E', opts.Weekday))
		format.Weekday = opts.Weekday
	}
	if opts.Era != "" && format.Era != opts.Era {
		format.Pattern = adjustPatternFields(format.Pattern, "G", fieldPatternWidth('G', opts.Era))
		format.Era = opts.Era
	}
	if opts.Hour != "" && format.Hour != opts.Hour {
		format.Pattern = adjustPatternFields(format.Pattern, hourPatternFields, numericPatternWidthFor(hourPatternChar(format.HourCycle), opts.Hour))
		format.Hour = opts.Hour
	}
	// Minute and second widths are deliberately not adjusted: the reference
	// best-fit matcher skips them (FormatJS BestFitFormatMatcher.ts:109-112,
	// "Don't mess with minute/second"). Rewriting them corrupts the interval
	// path, which parses the resolved pattern as its own skeleton.
	if opts.DayPeriod != "" && format.DayPeriod != opts.DayPeriod {
		format.Pattern = adjustPatternFields(format.Pattern, dayPeriodPatternFields, fieldPatternWidth('B', opts.DayPeriod))
		format.DayPeriod = opts.DayPeriod
	}
	if opts.FractionalSecondDigits != 0 && format.FractionalSecondDigits != opts.FractionalSecondDigits {
		format.Pattern = adjustPatternFields(format.Pattern, "S", repeatedField('S', opts.FractionalSecondDigits))
		format.FractionalSecondDigits = opts.FractionalSecondDigits
	}
	if opts.TimeZoneName != "" && format.TimeZoneName != opts.TimeZoneName {
		format.Pattern = adjustPatternFields(format.Pattern, timeZoneNamePatternFields, TimeZonePatternField(opts.TimeZoneName))
		format.TimeZoneName = opts.TimeZoneName
	}
	return format
}

func adjustNumericField(format *Formats, current *NumericStyle, requested NumericStyle, char byte) {
	if requested == "" || *current == requested {
		return
	}
	format.Pattern = adjustPatternFields(format.Pattern, string(char), numericPatternWidthFor(char, requested))
	*current = requested
}

func adjustPatternFields(pattern string, chars string, replacement string) string {
	out := make([]byte, 0, len(pattern)+len(replacement))
	for i := 0; i < len(pattern); {
		run := nextPatternRun(pattern, i)
		if !run.quoted && strings.IndexByte(chars, run.char) >= 0 {
			out = append(out, replacement...)
		} else {
			out = append(out, pattern[i:run.end]...)
		}
		i = run.end
	}
	return string(out)
}

func numericPatternWidthFor(char byte, style NumericStyle) string {
	if style == Numeric2Digit {
		return repeatedField(char, 2)
	}
	return string(char)
}

func fieldPatternWidth(char byte, style FieldStyle) string {
	return repeatedField(char, fieldStyleWidth(style))
}

func repeatedField(char byte, length int) string {
	return strings.Repeat(string(char), length)
}

func canAdjustFieldStyle(pattern string, chars string, requested FieldStyle) bool {
	width, ok := firstPatternFieldWidth(pattern, chars)
	if !ok {
		return true
	}
	have := fieldWidthStyle(width)
	return !isNumericFieldStyle(have) || isNumericFieldStyle(requested)
}

func firstPatternFieldWidth(pattern string, chars string) (int, bool) {
	for i := 0; i < len(pattern); {
		run := nextPatternRun(pattern, i)
		if !run.quoted && strings.IndexByte(chars, run.char) >= 0 {
			return run.width, true
		}
		i = run.end
	}
	return 0, false
}

func hourPatternChar(hourCycle HourCycle) byte {
	switch hourCycle {
	case HourCycleH11:
		return 'K'
	case HourCycleH12:
		return 'h'
	case HourCycleH23:
		return 'H'
	case HourCycleH24:
		return 'k'
	default:
		return 'h'
	}
}

func scoreFormat(opts Options, format Formats) int {
	score := 0
	score += scoreHourCycleField(opts.Hour12, format.HourCycle)
	score += scoreStyleField(opts.Weekday, format.Weekday)
	score += scoreStyleField(opts.Era, format.Era)
	score += scoreNumericField(opts.Year, format.Year)
	score += scoreStyleField(opts.Month, format.Month)
	score += scoreNumericField(opts.Day, format.Day)
	score += scoreStyleField(opts.DayPeriod, format.DayPeriod)
	score += scoreNumericField(opts.Hour, format.Hour)
	score += scoreNumericField(opts.Minute, format.Minute)
	score += scoreNumericField(opts.Second, format.Second)
	score += scoreFractionalSecondDigits(opts.FractionalSecondDigits, format.FractionalSecondDigits)
	score += scoreTimeZoneNameField(opts.TimeZoneName, format.TimeZoneName)
	return score
}

func scoreHourCycleField(want12 *bool, have HourCycle) int {
	if want12 == nil || have == "" || *want12 == hourCycleIs12Hour(have) {
		return 0
	}
	return -removalPenalty
}

func hourCycleIs12Hour(hourCycle HourCycle) bool {
	return hourCycle == HourCycleH11 || hourCycle == HourCycleH12
}

func scoreStyleField(want, have FieldStyle) int {
	if fieldScore := scoreField(want != "", have != ""); fieldScore != 0 || want == "" || have == "" {
		return fieldScore
	}
	if want == have {
		return 0
	}
	if isNumericFieldStyle(want) != isNumericFieldStyle(have) {
		return -differentNumericTypePenalty
	}
	return -stylePenalty(styleIndex(want), styleIndex(have))
}

func scoreNumericField(want, have NumericStyle) int {
	if fieldScore := scoreField(want != "", have != ""); fieldScore != 0 || want == "" || have == "" {
		return fieldScore
	}
	if want == have {
		return 0
	}
	return -differentNumericTypePenalty
}

func scoreTimeZoneNameField(want, have TimeZoneName) int {
	if fieldScore := scoreField(want != "", have != ""); fieldScore != 0 || want == "" || have == "" {
		return fieldScore
	}
	if want == have {
		return 0
	}
	switch want {
	case TimeZoneNameShort, TimeZoneNameShortGeneric:
		switch {
		case have == TimeZoneNameShortOffset:
			return -offsetPenalty
		case have == TimeZoneNameLongOffset:
			return -(offsetPenalty + shortMorePenalty)
		case want == TimeZoneNameShort && have == TimeZoneNameLong:
			return -shortMorePenalty
		case want == TimeZoneNameShortGeneric && have == TimeZoneNameLongGeneric:
			return -shortMorePenalty
		default:
			return -removalPenalty
		}
	case TimeZoneNameShortOffset:
		if have == TimeZoneNameLongOffset {
			return -shortMorePenalty
		}
	case TimeZoneNameLong, TimeZoneNameLongGeneric:
		switch {
		case have == TimeZoneNameLongOffset:
			return -offsetPenalty
		case have == TimeZoneNameShortOffset:
			return -(offsetPenalty + longLessPenalty)
		case want == TimeZoneNameLong && have == TimeZoneNameShort:
			return -longLessPenalty
		case want == TimeZoneNameLongGeneric && have == TimeZoneNameShortGeneric:
			return -longLessPenalty
		default:
			return -removalPenalty
		}
	case TimeZoneNameLongOffset:
		if have == TimeZoneNameShortOffset {
			return -longLessPenalty
		}
	}
	return -removalPenalty
}

func scoreFractionalSecondDigits(want, have int) int {
	if fieldScore := scoreField(want != 0, have != 0); fieldScore != 0 || want == 0 || have == 0 {
		return fieldScore
	}
	if want == have {
		return 0
	}
	return -stylePenalty(want-1, have-1)
}

func scoreField(want, have bool) int {
	switch {
	case want && !have:
		return -removalPenalty
	case !want && have:
		return -additionPenalty
	default:
		return 0
	}
}

func isNumericFieldStyle(style FieldStyle) bool {
	return style == FieldNumeric || style == Field2Digit
}

func styleIndex(style FieldStyle) int {
	switch style {
	case Field2Digit:
		return 0
	case FieldNumeric:
		return 1
	case FieldNarrow:
		return 2
	case FieldShort:
		return 3
	case FieldLong:
		return 4
	default:
		return 0
	}
}

func stylePenalty(wantIndex, haveIndex int) int {
	delta := wantIndex - haveIndex
	if delta < 0 {
		if delta == -1 {
			return shortMorePenalty
		}
		return longMorePenalty
	}
	if delta == 1 {
		return shortLessPenalty
	}
	return longLessPenalty
}
