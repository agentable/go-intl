package datetimeformat

import "time"

func (f *DateTimeFormat) FormatToParts(t time.Time) []Part {
	t = t.Round(0)
	if f.location != nil {
		t = t.In(f.location)
	}
	if parts, ok := f.pattern.parts(f, t); ok {
		return parts
	}
	return nil
}
