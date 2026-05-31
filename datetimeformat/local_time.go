package datetimeformat

import "time"

type localTime struct {
	Time       time.Time
	Weekday    time.Weekday
	Era        string
	Year       int
	Month      time.Month
	Day        int
	Hour       int
	Minute     int
	Second     int
	Nanosecond int
}

func (f *DateTimeFormat) localTime(t time.Time) localTime {
	switch f.resolved.Calendar {
	case "gregory", "iso8601":
		return gregoryLocalTime(t)
	default:
		return gregoryLocalTime(t)
	}
}

func gregoryLocalTime(t time.Time) localTime {
	year := t.Year()
	era := "AD"
	if year <= 0 {
		era = "BC"
	}
	return localTime{
		Time:       t,
		Weekday:    t.Weekday(),
		Era:        era,
		Year:       year,
		Month:      t.Month(),
		Day:        t.Day(),
		Hour:       t.Hour(),
		Minute:     t.Minute(),
		Second:     t.Second(),
		Nanosecond: t.Nanosecond(),
	}
}

func (t localTime) displayYear() int {
	if t.Year <= 0 {
		return 1 - t.Year
	}
	return t.Year
}
