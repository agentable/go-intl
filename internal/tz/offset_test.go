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
		{in: "+14:00", want: 14 * 3600 * 1000},
		{in: "-14:00", want: -14 * 3600 * 1000},
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

	for _, in := range []string{"", "+", "+14:01", "-14:01", "+23:59", "-23:59", "+24:00", "+01:60", "+01:30:45", "+1:00", "01:00", "+0a:00", "+\uff10\uff11:00"} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()

			_, err := ParseOffsetString(in)
			if !errors.Is(err, ErrUnsupportedTimeZone) {
				t.Fatalf("ParseOffsetString(%q) error = %v, want ErrUnsupportedTimeZone", in, err)
			}
			if !errors.Is(err, errors.ErrUnsupported) {
				t.Fatalf("ParseOffsetString(%q) error = %v, want errors.ErrUnsupported", in, err)
			}
		})
	}
}

func TestParseOffsetStringPreservesNegativeZero(t *testing.T) {
	t.Parallel()

	got, err := ParseOffsetString("-00:00")
	if err != nil {
		t.Fatalf("ParseOffsetString(-00:00) error = %v", err)
	}
	if got != 0 {
		t.Fatalf("ParseOffsetString(-00:00) = %d, want 0", got)
	}

	canonical, err := CanonicalOffsetString("-00")
	if err != nil {
		t.Fatalf("CanonicalOffsetString(-00) error = %v", err)
	}
	if canonical != "-00:00" {
		t.Fatalf("CanonicalOffsetString(-00) = %q, want -00:00", canonical)
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

func TestCanonicalOffsetStringRejectsInvalidOffset(t *testing.T) {
	t.Parallel()

	_, err := CanonicalOffsetString("+24:00")
	if !errors.Is(err, ErrUnsupportedTimeZone) {
		t.Fatalf("CanonicalOffsetString(+24:00) error = %v, want ErrUnsupportedTimeZone", err)
	}
	if !errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("CanonicalOffsetString(+24:00) error = %v, want errors.ErrUnsupported", err)
	}
}
