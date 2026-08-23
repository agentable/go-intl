package durationformat

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestDurationFormatPreservesIntegralNumberBoundaries(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(DigitalStyle)})
	if err != nil {
		t.Fatalf("New(en, digital) error = %v", err)
	}

	tests := []struct {
		name        string
		nanoseconds float64
		want        string
	}{
		{name: "2^53-1", nanoseconds: float64(1<<53 - 1), want: "0:00:9007199.254740991"},
		{name: "2^53", nanoseconds: float64(1 << 53), want: "0:00:9007199.254740992"},
		{name: "2^53+2", nanoseconds: float64(1<<53) + 2, want: "0:00:9007199.254740994"},
		{name: "above int64", nanoseconds: float64(1e20), want: "0:00:100000000000"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := format.Format(Duration{Nanoseconds: tc.nanoseconds})
			if err != nil {
				t.Fatalf("Format(nanoseconds=%g) error = %v", tc.nanoseconds, err)
			}
			if got != tc.want {
				t.Fatalf("Format(nanoseconds=%g) = %q, want %q", tc.nanoseconds, got, tc.want)
			}
			parts, err := format.FormatToParts(Duration{Nanoseconds: tc.nanoseconds})
			if err != nil {
				t.Fatalf("FormatToParts(nanoseconds=%g) error = %v", tc.nanoseconds, err)
			}
			if joined := durationPartsText(parts); joined != got {
				t.Fatalf("joined FormatToParts(nanoseconds=%g) = %q, want Format() %q", tc.nanoseconds, joined, got)
			}
		})
	}
}

func TestDurationFormatRejectsNonIntegralNumbers(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	tests := []struct {
		name  string
		value float64
		text  string
	}{
		{name: "NaN", value: math.NaN(), text: "NaN"},
		{name: "positive infinity", value: math.Inf(1), text: "Infinity"},
		{name: "negative infinity", value: math.Inf(-1), text: "-Infinity"},
		{name: "fraction", value: 1.5, text: "1.5"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			duration := Duration{Nanoseconds: tc.value}
			_, err := format.Format(duration)
			if err == nil {
				t.Fatalf("Format(nanoseconds=%s) error = nil, want invalid value", tc.text)
			}
			assertDurationInvalidValue(t, err, "nanoseconds", tc.text, "en", expectedDurationIntegerValue)

			_, err = format.FormatToParts(duration)
			if err == nil {
				t.Fatalf("FormatToParts(nanoseconds=%s) error = nil, want invalid value", tc.text)
			}
			assertDurationInvalidValue(t, err, "nanoseconds", tc.text, "en", expectedDurationIntegerValue)
		})
	}
}

func TestDurationFormatNormalizesNegativeZero(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(DigitalStyle)})
	if err != nil {
		t.Fatalf("New(en, digital) error = %v", err)
	}

	negativeZero := math.Copysign(0, -1)
	got, err := format.Format(Duration{Nanoseconds: negativeZero})
	if err != nil {
		t.Fatalf("Format(nanoseconds=-0) error = %v", err)
	}
	if want := "0:00:00"; got != want {
		t.Fatalf("Format(nanoseconds=-0) = %q, want %q", got, want)
	}
}

func TestDurationFormatSubsecondMathValueWitnesses(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Style: stringPtr(DigitalStyle)})
	if err != nil {
		t.Fatalf("New(en-US, digital) error = %v", err)
	}

	tests := []struct {
		name     string
		duration Duration
		want     string
	}{
		{name: "milliseconds", duration: Duration{Seconds: 2, Milliseconds: 712}, want: "0:00:02.712"},
		{name: "seconds nanoseconds", duration: Duration{Seconds: 17_179_869_194, Nanoseconds: 712}, want: "0:00:17179869194.000000712"},
		{name: "seconds microseconds", duration: Duration{Seconds: 17_179_869_194, Microseconds: 712}, want: "0:00:17179869194.000712"},
		{name: "seconds milliseconds", duration: Duration{Seconds: 17_179_869_194, Milliseconds: 712}, want: "0:00:17179869194.712"},
		{name: "nanoseconds", duration: Duration{Nanoseconds: 17_179_869_194_712}, want: "0:00:17179.869194712"},
		{name: "unsafe nanoseconds", duration: Duration{Nanoseconds: 9_007_199_254_740_992}, want: "0:00:9007199.254740992"},
		{name: "unsafe seconds", duration: Duration{Seconds: 9_007_199_254_740_991, Nanoseconds: 712}, want: "0:00:9007199254740991.000000712"},
		{name: "microseconds into milliseconds", duration: Duration{Milliseconds: 2, Microseconds: 712}, want: "0:00:00.002712"},
		{name: "milliseconds nanoseconds", duration: Duration{Milliseconds: 17_179_869_194, Nanoseconds: 712}, want: "0:00:17179869.194000712"},
		{name: "milliseconds microseconds", duration: Duration{Milliseconds: 17_179_869_194, Microseconds: 712}, want: "0:00:17179869.194712"},
		{name: "microseconds nanoseconds", duration: Duration{Microseconds: 17_179_869_194, Nanoseconds: 712}, want: "0:00:17179.869194712"},
		{name: "negative milliseconds", duration: Duration{Seconds: -2, Milliseconds: -712}, want: "-0:00:02.712"},
		{name: "negative seconds nanoseconds", duration: Duration{Seconds: -17_179_869_194, Nanoseconds: -712}, want: "-0:00:17179869194.000000712"},
		{name: "negative seconds microseconds", duration: Duration{Seconds: -17_179_869_194, Microseconds: -712}, want: "-0:00:17179869194.000712"},
		{name: "negative seconds milliseconds", duration: Duration{Seconds: -17_179_869_194, Milliseconds: -712}, want: "-0:00:17179869194.712"},
		{name: "negative microseconds into milliseconds", duration: Duration{Milliseconds: -2, Microseconds: -712}, want: "-0:00:00.002712"},
		{name: "negative milliseconds nanoseconds", duration: Duration{Milliseconds: -17_179_869_194, Nanoseconds: -712}, want: "-0:00:17179869.194000712"},
		{name: "negative milliseconds microseconds", duration: Duration{Milliseconds: -17_179_869_194, Microseconds: -712}, want: "-0:00:17179869.194712"},
		{name: "negative microseconds nanoseconds", duration: Duration{Microseconds: -17_179_869_194, Nanoseconds: -712}, want: "-0:00:17179.869194712"},
		{name: "negative nanoseconds", duration: Duration{Nanoseconds: -17_179_869_194_712}, want: "-0:00:17179.869194712"},
		{name: "negative unsafe nanoseconds", duration: Duration{Nanoseconds: -9_007_199_254_740_992}, want: "-0:00:9007199.254740992"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := format.Format(tc.duration)
			if err != nil {
				t.Fatalf("Format(%#v) error = %v", tc.duration, err)
			}
			if got != tc.want {
				t.Fatalf("Format(%#v) = %q, want %q", tc.duration, got, tc.want)
			}
		})
	}
}

func TestDurationFormatRejectsNormalizedSecondsLimit(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(DigitalStyle)})
	if err != nil {
		t.Fatalf("New(en, digital) error = %v", err)
	}

	_, err = format.Format(Duration{Seconds: float64(1 << 53), Nanoseconds: 712})
	if err == nil {
		t.Fatal("Format(seconds=2^53) error = nil, want invalid value")
	}
	assertDurationInvalidValue(t, err, "duration", "normalized seconds", "en", expectedDurationNormalizedSeconds)
}

func TestDurationFormatPreservesExactValuesAcrossUnitStyles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		options  Options
		duration Duration
		want     string
		wantPart []Part
	}{
		{
			name:     "wide negative year",
			options:  Options{Style: stringPtr(LongStyle)},
			duration: Duration{Years: -4_294_967_295},
			want:     "-4,294,967,295 years",
			wantPart: []Part{{Type: PartMinusSign, Value: "-", Unit: Year}, {Type: PartInteger, Value: "4", Unit: Year}, {Type: PartGroup, Value: ",", Unit: Year}, {Type: PartInteger, Value: "294", Unit: Year}, {Type: PartGroup, Value: ",", Unit: Year}, {Type: PartInteger, Value: "967", Unit: Year}, {Type: PartGroup, Value: ",", Unit: Year}, {Type: PartInteger, Value: "295", Unit: Year}, {Type: PartLiteral, Value: " ", Unit: Year}, {Type: PartUnit, Value: "years", Unit: Year}},
		},
		{
			name: "fractional seconds include nanoseconds",
			options: Options{
				Style:        stringPtr(NarrowStyle),
				Milliseconds: stringPtr(NumericUnitStyle),
			},
			duration: Duration{Seconds: -1, Nanoseconds: -2},
			want:     "-1.000000002s",
			wantPart: []Part{{Type: PartMinusSign, Value: "-", Unit: Second}, {Type: PartInteger, Value: "1", Unit: Second}, {Type: PartDecimal, Value: ".", Unit: Second}, {Type: PartFraction, Value: "000000002", Unit: Second}, {Type: PartUnit, Value: "s", Unit: Second}},
		},
		{
			name: "numeric seconds retain one sign",
			options: Options{
				Hours:   stringPtr(NumericUnitStyle),
				Minutes: stringPtr(TwoDigitUnitStyle),
				Seconds: stringPtr(TwoDigitUnitStyle),
			},
			duration: Duration{Minutes: -2, Seconds: -3},
			want:     "-0:02:03",
			wantPart: []Part{{Type: PartMinusSign, Value: "-", Unit: Hour}, {Type: PartInteger, Value: "0", Unit: Hour}, {Type: PartLiteral, Value: ":"}, {Type: PartInteger, Value: "02", Unit: Minute}, {Type: PartLiteral, Value: ":"}, {Type: PartInteger, Value: "03", Unit: Second}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, "en-US")}, tc.options)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			got, err := format.Format(tc.duration)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("Format() = %q, want %q", got, tc.want)
			}
			parts, err := format.FormatToParts(tc.duration)
			if err != nil {
				t.Fatalf("FormatToParts() error = %v", err)
			}
			if !reflect.DeepEqual(parts, tc.wantPart) {
				t.Fatalf("FormatToParts() = %#v, want %#v", parts, tc.wantPart)
			}
		})
	}
}

func TestDurationFormatRejectsEveryUnitBeyondItsPublicLimit(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}
	tests := []struct {
		name     string
		duration Duration
		value    string
	}{
		{name: "years", duration: Duration{Years: 4_294_967_296}, value: "4294967296"},
		{name: "months", duration: Duration{Months: -4_294_967_296}, value: "-4294967296"},
		{name: "weeks", duration: Duration{Weeks: 4_294_967_296}, value: "4294967296"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := format.Format(tc.duration)
			if got != "" || !errors.Is(err, intlerr.ErrInvalidValue) {
				t.Fatalf("Format() = %q, error %v, want empty output and ErrInvalidValue", got, err)
			}
			assertDurationInvalidValue(t, err, tc.name, tc.value, "en", "an absolute value less than 2^32")
		})
	}
}
