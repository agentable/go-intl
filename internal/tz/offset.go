package tz

func ParseOffsetString(s string) (int64, error) {
	minutes, err := parseOffsetMinutes(s)
	return int64(minutes) * 60_000, err
}

func CanonicalOffsetString(s string) (string, error) {
	minutes, err := parseOffsetMinutes(s)
	if err != nil {
		return "", err
	}
	return canonicalOffsetString(minutes), nil
}

func isOffsetName(s string) bool {
	return s != "" && (s[0] == '+' || s[0] == '-')
}

func canonicalOffsetString(minutes int) string {
	sign := byte('+')
	if minutes < 0 {
		sign = '-'
		minutes = -minutes
	}
	const digits = "0123456789"
	hour, minute := minutes/60, minutes%60
	return string([]byte{sign, digits[hour/10], digits[hour%10], ':', digits[minute/10], digits[minute%10]})
}

func parseOffsetMinutes(s string) (int, error) {
	if !isOffsetName(s) {
		return 0, invalidOffset(s)
	}
	minuteText := "00"
	switch len(s) {
	case len("+05"):
	case len("+0530"):
		minuteText = s[3:5]
	case len("+05:30"):
		if s[3] != ':' {
			return 0, invalidOffset(s)
		}
		minuteText = s[4:6]
	default:
		return 0, invalidOffset(s)
	}
	hour, hourOK := parseOffsetPair(s[1:3])
	minute, minuteOK := parseOffsetPair(minuteText)
	if !hourOK || !minuteOK || hour > 23 || minute > 59 {
		return 0, invalidOffset(s)
	}
	minutes := hour*60 + minute
	if s[0] == '-' {
		minutes = -minutes
	}
	return minutes, nil
}

func parseOffsetPair(s string) (int, bool) {
	tens, ones := s[0]-'0', s[1]-'0'
	if tens > 9 || ones > 9 {
		return 0, false
	}
	return int(tens)*10 + int(ones), true
}
