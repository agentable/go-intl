package ecma402

import (
	"regexp"
	"strconv"
	"strings"
)

// TimeZoneNameSet is the data-injection contract used by IsValidTimeZoneName.
// Concrete implementations live alongside the IANA tz database (SPEC 31);
// internal/ecma402 consumes this interface to avoid coupling to that data.
//
// Contains receives an uppercase IANA zone identifier (link aliases included)
// and returns true if it is recognised.
type TimeZoneNameSet interface {
	Contains(upperName string) bool
}

// utcOffsetRegex matches the UTC offset forms accepted by ECMA-262
// IsTimeZoneOffsetString: ±HH, ±HHMM, ±HH:MM, ±HH:MM:SS, ±HH:MM:SS.sss.
var utcOffsetRegex = regexp.MustCompile(
	`^([+-])(\d{2})(?::?(\d{2}))?(?::?(\d{2}))?(?:\.(\d{1,9}))?$`,
)

// isTimeZoneOffsetString validates a UTC-offset time-zone string per
// ECMA-262 sec-istimezoneoffsetstring.
func isTimeZoneOffsetString(s string) bool {
	if len(s) == 0 || (s[0] != '+' && s[0] != '-') {
		return false
	}
	m := utcOffsetRegex.FindStringSubmatch(s)
	if m == nil {
		return false
	}
	hours, _ := strconv.Atoi(m[2])
	minutes := 0
	if m[3] != "" {
		minutes, _ = strconv.Atoi(m[3])
	}
	seconds := 0
	if m[4] != "" {
		seconds, _ = strconv.Atoi(m[4])
	}
	return hours <= 23 && minutes <= 59 && seconds <= 59
}

// IsValidTimeZoneName mirrors ECMA-402 sec-isvalidtimezonename (extended per
// PR #788 to accept UTC offsets). A name is valid if it is a UTC-offset form
// or if its uppercase form is recognised by the supplied TimeZoneNameSet.
func IsValidTimeZoneName(tz string, zones TimeZoneNameSet) bool {
	if isTimeZoneOffsetString(tz) {
		return true
	}
	if zones == nil {
		return false
	}
	return zones.Contains(strings.ToUpper(tz))
}
