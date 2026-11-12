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
		f.Add(millis, int64(1580745845123), uint8(0), uint8(0), uint8(255), uint8(2))
	}
	for i, instant := range []string{"2024-03-10T07:00:00Z", "2024-04-06T15:00:00Z", "2024-09-28T14:00:00Z", "2024-03-10T02:00:00Z"} {
		at, err := time.Parse(time.RFC3339, instant)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(at.UnixMilli()-1, at.UnixMilli(), uint8(i+1), uint8(0), uint8(71), uint8(2))
		f.Add(at.UnixMilli(), at.UnixMilli()-3600000, uint8(i+1), uint8(4), uint8(27), uint8(1))
	}
	f.Fuzz(func(t *testing.T, millis, endMillis int64, zone, language, components, cycle uint8) {
		// Independent bounded endpoints retain both sides of TimeClip.
		millis %= 8640000000000002
		endMillis %= 8640000000000002
		start, end := time.UnixMilli(millis), time.UnixMilli(endMillis)
		opts := Options{TimeZone: new([]string{"UTC", "America/New_York", "Australia/Lord_Howe", "Pacific/Chatham", "Africa/Casablanca", "+05:30"}[int(zone)%6])}
		opts.HourCycle = new([]string{"h11", "h12", "h23", "h24", "invalid"}[int(cycle)%5])
		if components&1 != 0 {
			opts.Hour = new("2-digit")
		}
		if components&2 != 0 {
			opts.Minute = new("2-digit")
		}
		if components&4 != 0 {
			opts.Second = new("2-digit")
		}
		if components&8 != 0 {
			opts.DayPeriod = new("long")
		}
		if components&16 != 0 {
			opts.FractionalSecondDigits = new(int(components>>5)%3 + 1)
		}
		if components&32 != 0 {
			opts.Year, opts.Month, opts.Day = new("numeric"), new("short"), new("numeric")
		}
		if components&64 != 0 {
			opts.TimeZoneName = new([]string{"short", "long", "shortOffset", "longOffset", "shortGeneric", "longGeneric"}[int(components>>1)%6])
		}
		if components&128 != 0 && components&127 == 0 {
			opts.DateStyle, opts.TimeStyle = new("full"), new("short")
		}
		tag := []string{"en", "zh", "ja", "ru", "ar", "en-u-nu-arab-x-test"}[int(language)%6]
		format, err := New(intltest.LocaleList(t, tag), opts)
		if cycle%5 == 4 {
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("invalid hourCycle error = %v", err)
			}
			return
		}
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
