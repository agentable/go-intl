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
	return offsetMilliseconds(sign, hour, minute), nil
}

func CanonicalOffsetString(s string) (string, error) {
	sign, hour, minute, err := parseOffsetString(s)
	if err != nil {
		return "", err
	}
	return canonicalOffsetString(sign, hour, minute), nil
}

func parseFixedOffset(s string) (int64, string, error) {
	sign, hour, minute, err := parseOffsetString(s)
	if err != nil {
		return 0, "", err
	}
	return offsetMilliseconds(sign, hour, minute), canonicalOffsetString(sign, hour, minute), nil
}

func isOffsetName(s string) bool {
	return s != "" && (s[0] == '+' || s[0] == '-')
}

func offsetMilliseconds(sign int64, hour, minute int) int64 {
	return sign * int64(hour*3600*1000+minute*60*1000)
}

func canonicalOffsetString(sign int64, hour, minute int) string {
	signRune := '+'
	if sign < 0 && (hour != 0 || minute != 0) {
		signRune = '-'
	}
	return fmt.Sprintf("%c%02d:%02d", signRune, hour, minute)
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

func asciiDigits(s string) bool {
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
