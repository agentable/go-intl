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

func gregoryTimeInLocation(t time.Time, loc *time.Location) (time.Time, localTime) {
	if loc != nil {
		t = t.In(loc)
	}
	return t, gregoryLocalTime(t)
}

func (t localTime) displayYear() int {
	if t.Year <= 0 {
		return 1 - t.Year
	}
	return t.Year
}
