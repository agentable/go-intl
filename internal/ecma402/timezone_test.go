package ecma402_test

import (
	"strings"
	"testing"

	"github.com/agentable/go-intl/internal/ecma402"
)

// stubZones is a minimal TimeZoneNameSet for tests — uppercased names only.
type stubZones map[string]struct{}

func (s stubZones) Contains(upperName string) bool {
	_, ok := s[upperName]
	return ok
}

func newZones(names ...string) stubZones {
	out := make(stubZones, len(names))
	for _, n := range names {
		out[strings.ToUpper(n)] = struct{}{}
	}
	return out
}

func TestIsValidTimeZoneName_Named(t *testing.T) {
	t.Parallel()
	zones := newZones(
		"America/Los_Angeles",
		"America/New_York",
		"America/Indiana/Indianapolis",
		"America/Indianapolis", // link alias
	)

	tests := []struct {
		in   string
		want bool
	}{
		{"America/Los_Angeles", true},
		{"america/los_angeles", true}, // case-insensitive
		{"America/New_York", true},
		{"America/Indiana/Indianapolis", true},
		{"America/Indianapolis", true}, // alias resolves through zones
		{"America/Mars", false},
		{"", false},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := ecma402.IsValidTimeZoneName(tc.in, zones); got != tc.want {
				t.Errorf("IsValidTimeZoneName(%q) = %v, want %v",
					tc.in, got, tc.want)
			}
		})
	}
}

func TestIsValidTimeZoneName_UTCOffset(t *testing.T) {
	t.Parallel()
	zones := newZones("America/New_York")

	valid := []string{
		"+00:00", "+01:00", "-05:00", "+05:30", "-12:00", "+23:59",
		"+0100", "-0500", "+0530",
		"+01", "-05", "+00",
		"+01:30:45", "-05:00:00",
		"+01:30:45.123", "-05:00:00.999999999",
	}
	for _, tz := range valid {
		t.Run("valid/"+tz, func(t *testing.T) {
			t.Parallel()
			if !ecma402.IsValidTimeZoneName(tz, zones) {
				t.Errorf("IsValidTimeZoneName(%q) = false, want true", tz)
			}
		})
	}

	invalid := []string{
		"+24:00", "+25:00", "-24:00",
		"+01:60", "+01:99",
		"+01:30:60",
		"+1:00", "+01:0", "+1:0",
		"01:00", "0100",
		"+01:00:00:00", "+01:00abc",
	}
	for _, tz := range invalid {
		t.Run("invalid/"+tz, func(t *testing.T) {
			t.Parallel()
			if ecma402.IsValidTimeZoneName(tz, zones) {
				t.Errorf("IsValidTimeZoneName(%q) = true, want false", tz)
			}
		})
	}
}

func TestIsValidTimeZoneName_NilSet(t *testing.T) {
	t.Parallel()
	// Nil zones still permits valid UTC offsets.
	if !ecma402.IsValidTimeZoneName("+05:00", nil) {
		t.Error("UTC offset must be valid even with nil zones")
	}
	// Nil zones rejects named identifiers.
	if ecma402.IsValidTimeZoneName("America/New_York", nil) {
		t.Error("named zone must be invalid with nil zones")
	}
}
