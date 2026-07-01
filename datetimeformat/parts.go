package datetimeformat

import "time"

func (f *DateTimeFormat) FormatToParts(t time.Time) []Part {
	pattern := f.pattern
	_, local := gregoryTimeInLocation(t.Round(0), f.location)
	return pattern.parts(f, local)
}
