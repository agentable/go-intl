package tz

import (
	"errors"
	"testing"
)

func TestParseOffsetStringValidOffsets(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in   string
		want int64
	}{
		{in: "+01", want: 1 * 3600 * 1000},
		{in: "-05", want: -5 * 3600 * 1000},
		{in: "+0100", want: 1 * 3600 * 1000},
		{in: "-0530", want: -(5*3600*1000 + 30*60*1000)},
		{in: "+05:30", want: 5*3600*1000 + 30*60*1000},
		{in: "-08:00", want: -8 * 3600 * 1000},
		{in: "+23:59", want: 23*3600*1000 + 59*60*1000},
		{in: "-23:59", want: -(23*3600*1000 + 59*60*1000)},
	} {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()

			got, err := ParseOffsetString(tc.in)
			if err != nil {
				t.Fatalf("ParseOffsetString(%q) error = %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("ParseOffsetString(%q) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseOffsetStringRejectsInvalidOffsets(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"+24:00", "+01:60", "+01:30:45", "+1:00", "01:00"} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()

			_, err := ParseOffsetString(in)
			if !errors.Is(err, ErrUnsupportedTimeZone) {
				t.Fatalf("ParseOffsetString(%q) error = %v, want ErrUnsupportedTimeZone", in, err)
			}
		})
	}
}

func TestCanonicalOffsetString(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in   string
		want string
	}{
		{in: "+01", want: "+01:00"},
		{in: "+0100", want: "+01:00"},
		{in: "-0530", want: "-05:30"},
		{in: "+00", want: "+00:00"},
		{in: "-00:00", want: "-00:00"},
	} {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()

			got, err := CanonicalOffsetString(tc.in)
			if err != nil {
				t.Fatalf("CanonicalOffsetString(%q) error = %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("CanonicalOffsetString(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
