package datetimeformat

import "time"

func (f *DateTimeFormat) FormatToParts(t time.Time) ([]Part, error) {
	t, err := normalizeInstant(t, "value")
	if err != nil {
		return nil, err
	}
	_, local := gregoryTimeInLocation(t, f.location)
	return f.pattern.parts(f, local), nil
}
