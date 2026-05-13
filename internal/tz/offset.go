package tz

import (
	"fmt"
	"strconv"
)

func ParseOffsetString(s string) (int64, error) {
	sign, hour, minute, err := parseOffsetString(s)
	if err != nil {
		return 0, err
	}
	return sign * int64(hour*3600*1000+minute*60*1000), nil
}

func CanonicalOffsetString(s string) (string, error) {
	_, hour, minute, err := parseOffsetString(s)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%c%02d:%02d", s[0], hour, minute), nil
}

func parseOffsetString(s string) (int64, int, int, error) {
	if len(s) == 0 || s[0] != '+' && s[0] != '-' {
		return 0, 0, 0, invalidOffset(s)
	}
	sign := int64(1)
	if s[0] == '-' {
		sign = -1
	}
	var hourText, minuteText string
	switch len(s) {
	case len("+05"):
		hourText = s[1:3]
		minuteText = "00"
	case len("+0530"):
		hourText = s[1:3]
		minuteText = s[3:5]
	case len("+05:30"):
		if s[3] != ':' {
			return 0, 0, 0, invalidOffset(s)
		}
		hourText = s[1:3]
		minuteText = s[4:6]
	default:
		return 0, 0, 0, invalidOffset(s)
	}
	if !asciiDigits(hourText) || !asciiDigits(minuteText) {
		return 0, 0, 0, invalidOffset(s)
	}
	hour, err := strconv.Atoi(hourText)
	if err != nil {
		return 0, 0, 0, invalidOffset(s)
	}
	minute, err := strconv.Atoi(minuteText)
	if err != nil {
		return 0, 0, 0, invalidOffset(s)
	}
	if hour > 23 || minute > 59 {
		return 0, 0, 0, invalidOffset(s)
	}
	return sign, hour, minute, nil
}

func invalidOffset(s string) error {
	return fmt.Errorf("tz: invalid offset %q: %w", s, ErrUnsupportedTimeZone)
}

func asciiDigits(s string) bool {
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
