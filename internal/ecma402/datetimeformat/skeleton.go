package ecma402dtf

import "strings"

// FieldStyle is an ECMA-402 textual date-time field width.
type FieldStyle string

// NumericStyle is an ECMA-402 numeric date-time field width.
type NumericStyle string

// HourCycle is an ECMA-402 hour cycle identifier.
type HourCycle string

// TimeZoneName is an ECMA-402 time zone name style.
type TimeZoneName string

// FormatMatcher selects the DateTimeFormat matching algorithm.
type FormatMatcher string

const (
	FieldNarrow  FieldStyle = "narrow"
	FieldShort   FieldStyle = "short"
	FieldLong    FieldStyle = "long"
	FieldNumeric FieldStyle = "numeric"
	Field2Digit  FieldStyle = "2-digit"

	NumericNumeric NumericStyle = "numeric"
	Numeric2Digit  NumericStyle = "2-digit"

	HourCycleH11 HourCycle = "h11"
	HourCycleH12 HourCycle = "h12"
	HourCycleH23 HourCycle = "h23"
	HourCycleH24 HourCycle = "h24"

	TimeZoneNameShort        TimeZoneName = "short"
	TimeZoneNameLong         TimeZoneName = "long"
	TimeZoneNameShortOffset  TimeZoneName = "shortOffset"
	TimeZoneNameLongOffset   TimeZoneName = "longOffset"
	TimeZoneNameShortGeneric TimeZoneName = "shortGeneric"
	TimeZoneNameLongGeneric  TimeZoneName = "longGeneric"

	MatcherBasic   FormatMatcher = "basic"
	MatcherBestFit FormatMatcher = "best fit"
)

// Options describes requested date-time skeleton fields.
type Options struct {
	Weekday                FieldStyle
	Era                    FieldStyle
	Year                   NumericStyle
	Month                  FieldStyle
	Day                    NumericStyle
	Hour                   NumericStyle
	HourCycle              HourCycle
	Minute                 NumericStyle
	Second                 NumericStyle
	DayPeriod              FieldStyle
	FractionalSecondDigits int
	TimeZoneName           TimeZoneName
	Hour12                 *bool
}

// Formats describes a parsed CLDR skeleton candidate.
type Formats struct {
	Pattern                string
	Pattern12              string
	Skeleton               string
	PatternHasTimeZoneName bool
	Era                    FieldStyle
	Year                   NumericStyle
	Month                  FieldStyle
	Day                    NumericStyle
	Weekday                FieldStyle
	Hour                   NumericStyle
	HourCycle              HourCycle
	Minute                 NumericStyle
	Second                 NumericStyle
	DayPeriod              FieldStyle
	FractionalSecondDigits int
	TimeZoneName           TimeZoneName
}

// Parse converts an LDML skeleton and pattern into a matcher candidate.
func Parse(skeleton string, pattern string, hour12 *bool, hourCycle HourCycle) Formats {
	format := Formats{Skeleton: skeleton, Pattern: pattern, PatternHasTimeZoneName: patternHasTimeZoneName(pattern)}
	for i := 0; i < len(skeleton); {
		if skeleton[i] == '\'' {
			i = skipQuotedLiteral(skeleton, i)
			continue
		}
		j := i + 1
		for j < len(skeleton) && skeleton[j] == skeleton[i] {
			j++
		}
		applySkeletonToken(&format, skeleton[i], j-i)
		i = j
	}
	if hourCycle != "" {
		format.HourCycle = hourCycle
	}
	return format
}

func patternHasTimeZoneName(pattern string) bool {
	for i := 0; i < len(pattern); {
		if pattern[i] == '\'' {
			i = skipQuotedLiteral(pattern, i)
			continue
		}
		if strings.IndexByte("zZOvVxX", pattern[i]) >= 0 {
			return true
		}
		j := i + 1
		for j < len(pattern) && pattern[j] == pattern[i] {
			j++
		}
		i = j
	}
	return false
}

func skipQuotedLiteral(s string, start int) int {
	for i := start + 1; i < len(s); i++ {
		if s[i] != '\'' {
			continue
		}
		if i+1 < len(s) && s[i+1] == '\'' {
			i++
			continue
		}
		return i + 1
	}
	return len(s)
}

func applySkeletonToken(format *Formats, char byte, length int) {
	switch char {
	case 'G':
		format.Era = textStyle(length)
	case 'y', 'Y', 'u', 'U', 'r':
		// LDML §2.6.1: y/Y calendar year, u extended year, U cyclic year name
		// (Chinese/Japanese), r related Gregorian year. ECMA-402 surfaces all
		// of them through the `year` option; the active Gregorian path renders
		// only the numeric year value.
		format.Year = numericStyle(length)
	case 'M', 'L':
		format.Month = fieldWidthStyle(length)
	case 'd':
		format.Day = numericStyle(length)
	case 'E', 'e', 'c':
		format.Weekday = weekdayStyle(length)
	case 'h':
		format.Hour = numericStyle(length)
		format.HourCycle = HourCycleH12
	case 'H':
		format.Hour = numericStyle(length)
		format.HourCycle = HourCycleH23
	case 'K':
		format.Hour = numericStyle(length)
		format.HourCycle = HourCycleH11
	case 'k':
		format.Hour = numericStyle(length)
		format.HourCycle = HourCycleH24
	case 'm':
		format.Minute = numericStyle(length)
	case 's':
		format.Second = numericStyle(length)
	case 'S':
		format.FractionalSecondDigits = length
	case 'a', 'b', 'B':
		format.DayPeriod = textStyle(length)
	case 'z':
		format.TimeZoneName = timeZoneNameStyle(length)
	case 'Z', 'O', 'X', 'x':
		format.TimeZoneName = offsetTimeZoneNameStyle(length)
	case 'v':
		format.TimeZoneName = genericTimeZoneNameStyle(length)
	case 'V':
		// LDML §2.6.1: V is a zone identifier / exemplar city variant of the
		// generic non-location format. Mapped to the same TimeZoneName field
		// so the best-fit matcher does not treat skeletons with V as missing
		// time-zone information.
		format.TimeZoneName = genericTimeZoneNameStyle(length)
	case 'Q', 'q':
		// LDML §2.6.1: quarter (formatting / standalone). Not surfaced by
		// ECMA-402 options today. Recognized here so the parser does not
		// silently treat a quarter-bearing skeleton as token-less; future
		// matcher work can score this on its own Formats field.
	}
}

func numericStyle(length int) NumericStyle {
	if length == 2 {
		return Numeric2Digit
	}
	return NumericNumeric
}

func fieldWidthStyle(length int) FieldStyle {
	switch length {
	case 1:
		return FieldNumeric
	case 2:
		return Field2Digit
	case 3:
		return FieldShort
	case 4:
		return FieldLong
	default:
		return FieldNarrow
	}
}

func textStyle(length int) FieldStyle {
	switch length {
	case 4:
		return FieldLong
	case 5:
		return FieldNarrow
	default:
		return FieldShort
	}
}

func weekdayStyle(length int) FieldStyle {
	switch length {
	case 4:
		return FieldLong
	case 5:
		return FieldNarrow
	default:
		return FieldShort
	}
}

func timeZoneNameStyle(length int) TimeZoneName {
	if length == 4 {
		return TimeZoneNameLong
	}
	return TimeZoneNameShort
}

func offsetTimeZoneNameStyle(length int) TimeZoneName {
	if length == 4 {
		return TimeZoneNameLongOffset
	}
	return TimeZoneNameShortOffset
}

func genericTimeZoneNameStyle(length int) TimeZoneName {
	if length == 4 {
		return TimeZoneNameLongGeneric
	}
	return TimeZoneNameShortGeneric
}
