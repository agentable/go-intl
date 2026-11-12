package datetimeformat

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/intltest"
)

func FuzzDateTimeFormatPartition(f *testing.F) {
	for _, millis := range []int64{0, -1, 1580745845123, 8640000000000000, -8640000000000000, 8640000000000001, -8640000000000001} {
		for scenario := uint8(0); scenario < 8; scenario++ {
			f.Add(millis, int64(3600000), scenario)
		}
	}
	f.Fuzz(func(t *testing.T, millis, delta int64, scenario uint8) {
		// Preserve both TimeClip boundaries while excluding int64/time overflow.
		millis %= 8640000000000002
		delta %= 3 * 86400000
		start, end := time.UnixMilli(millis), time.UnixMilli(millis+delta)
		options := []Options{
			{Minute: new("numeric")}, {Second: new("numeric"), FractionalSecondDigits: new(3)},
			{DayPeriod: new("long")}, {Weekday: new("long"), Hour: new("numeric")},
			{Year: new("numeric"), Month: new("long"), Day: new("numeric"), Hour: new("numeric"), Minute: new("2-digit")},
			{DateStyle: new("full"), TimeStyle: new("short")}, {Era: new("short"), Hour: new("numeric")},
			{Hour: new("2-digit"), Minute: new("2-digit"), TimeZoneName: new("shortOffset")},
		}
		opts := options[int(scenario)%len(options)]
		opts.TimeZone = new("UTC")
		opts.HourCycle = new([]string{"h11", "h12", "h23", "h24"}[int(scenario)%4])
		tag := []string{"en", "zh", "ja", "ru", "ar", "en-u-nu-arab-x-test"}[int(scenario)%6]
		format, err := New(intltest.LocaleList(t, tag), opts)
		if err != nil {
			t.Fatal(err)
		}
		text, err := format.Format(start)
		parts, partsErr := format.FormatToParts(start)
		if millis < -8640000000000000 || millis > 8640000000000000 {
			if !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(partsErr, intlerr.ErrInvalidValue) {
				t.Fatalf("TimeClip errors=%v, %v", err, partsErr)
			}
		} else {
			if err != nil || partsErr != nil {
				t.Fatalf("valid instant errors=%v, %v", err, partsErr)
			}
			var joined strings.Builder
			for _, part := range parts {
				joined.WriteString(part.Value)
			}
			if text != joined.String() {
				t.Fatalf("text %q != parts %q", text, joined.String())
			}
		}
		text, err = format.FormatRange(start, end)
		rangeParts, partsErr := format.FormatRangeToParts(start, end)
		endMillis := millis + delta
		if millis < -8640000000000000 || millis > 8640000000000000 || endMillis < -8640000000000000 || endMillis > 8640000000000000 {
			if !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(partsErr, intlerr.ErrInvalidValue) {
				t.Fatalf("TimeClip range errors=%v, %v", err, partsErr)
			}
			return
		}
		if err != nil || partsErr != nil {
			t.Fatalf("valid range errors=%v, %v", err, partsErr)
		}
		var joined strings.Builder
		for _, part := range rangeParts {
			if part.Source != SourceStartRange && part.Source != SourceEndRange && part.Source != SourceShared {
				t.Fatalf("invalid source %q", part.Source)
			}
			joined.WriteString(part.Value)
		}
		if text != joined.String() {
			t.Fatalf("range %q != parts %q", text, joined.String())
		}
	})
}
