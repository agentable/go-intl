package datetimeformat

import (
	"testing"

	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func intPtr(v int) *int {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

func stringPtr[T ~string](v T) *string {
	value := string(v)
	return &value
}

func TestOptionsPointerValuesCopiedDuringConstruction(t *testing.T) {
	t.Parallel()

	formatMatcher := string(BasicFormatMatcher)
	timeZoneName := string(ShortTimeZoneName)
	weekday := string(ShortFieldStyle)
	era := string(ShortFieldStyle)
	year := string(NumericFieldStyle)
	month := string(ShortMonthStyle)
	day := string(NumericFieldStyle)
	dayPeriod := string(ShortFieldStyle)
	hour := string(NumericFieldStyle)
	minute := string(TwoDigitFieldStyle)
	second := string(TwoDigitFieldStyle)
	hourCycle := string(H23HourCycle)
	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
		FormatMatcher: &formatMatcher,
		TimeZoneName:  &timeZoneName,
		Weekday:       &weekday,
		Era:           &era,
		Year:          &year,
		Month:         &month,
		Day:           &day,
		DayPeriod:     &dayPeriod,
		Hour:          &hour,
		Minute:        &minute,
		Second:        &second,
		HourCycle:     &hourCycle,
	})
	if err != nil {
		t.Fatal(err)
	}

	timeZoneName = string(LongTimeZoneName)
	weekday = string(LongFieldStyle)
	era = string(LongFieldStyle)
	year = string(TwoDigitFieldStyle)
	month = string(LongMonthStyle)
	day = string(TwoDigitFieldStyle)
	dayPeriod = string(LongFieldStyle)
	hour = string(TwoDigitFieldStyle)
	minute = string(NumericFieldStyle)
	second = string(NumericFieldStyle)
	hourCycle = string(H11HourCycle)

	resolved := format.ResolvedOptions()
	if got, want := ecma402.ResolvedScalarValue(resolved.TimeZoneName), ShortTimeZoneName; got != want {
		t.Fatalf("ResolvedOptions().TimeZoneName = %q, want %q", got, want)
	}
	if got, want := ecma402.ResolvedScalarValue(resolved.Weekday), ShortFieldStyle; got != want {
		t.Fatalf("ResolvedOptions().Weekday = %q, want %q", got, want)
	}
	if got, want := ecma402.ResolvedScalarValue(resolved.Era), ShortFieldStyle; got != want {
		t.Fatalf("ResolvedOptions().Era = %q, want %q", got, want)
	}
	if got, want := ecma402.ResolvedScalarValue(resolved.Year), NumericFieldStyle; got != want {
		t.Fatalf("ResolvedOptions().Year = %q, want %q", got, want)
	}
	if got, want := ecma402.ResolvedScalarValue(resolved.Month), ShortMonthStyle; got != want {
		t.Fatalf("ResolvedOptions().Month = %q, want %q", got, want)
	}
	if got, want := ecma402.ResolvedScalarValue(resolved.Day), NumericFieldStyle; got != want {
		t.Fatalf("ResolvedOptions().Day = %q, want %q", got, want)
	}
	if got, want := ecma402.ResolvedScalarValue(resolved.DayPeriod), ShortFieldStyle; got != want {
		t.Fatalf("ResolvedOptions().DayPeriod = %q, want %q", got, want)
	}
	if got, want := ecma402.ResolvedScalarValue(resolved.Hour), NumericFieldStyle; got != want {
		t.Fatalf("ResolvedOptions().Hour = %q, want %q", got, want)
	}
	if got, want := ecma402.ResolvedScalarValue(resolved.Minute), TwoDigitFieldStyle; got != want {
		t.Fatalf("ResolvedOptions().Minute = %q, want %q", got, want)
	}
	if got, want := ecma402.ResolvedScalarValue(resolved.Second), TwoDigitFieldStyle; got != want {
		t.Fatalf("ResolvedOptions().Second = %q, want %q", got, want)
	}
	if got, want := ecma402.ResolvedScalarValue(resolved.HourCycle), H23HourCycle; got != want {
		t.Fatalf("ResolvedOptions().HourCycle = %q, want %q", got, want)
	}
}

func TestOptionsStylePointerValuesCopiedDuringConstruction(t *testing.T) {
	t.Parallel()

	dateStyle := string(MediumDateTimeStyle)
	timeStyle := string(ShortDateTimeStyle)
	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
		DateStyle: &dateStyle,
		TimeStyle: &timeStyle,
	})
	if err != nil {
		t.Fatal(err)
	}
	dateStyle = string(FullDateTimeStyle)
	timeStyle = string(LongDateTimeStyle)

	resolved := format.ResolvedOptions()
	if got, want := ecma402.ResolvedScalarValue(resolved.DateStyle), MediumDateTimeStyle; got != want {
		t.Fatalf("ResolvedOptions().DateStyle = %q, want %q", got, want)
	}
	if got, want := ecma402.ResolvedScalarValue(resolved.TimeStyle), ShortDateTimeStyle; got != want {
		t.Fatalf("ResolvedOptions().TimeStyle = %q, want %q", got, want)
	}
}

func TestOptionsBoolPointerValuesCopiedDuringConstruction(t *testing.T) {
	t.Parallel()

	hour12 := false
	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
		Hour:   stringPtr(NumericFieldStyle),
		Hour12: &hour12,
	})
	if err != nil {
		t.Fatal(err)
	}
	hour12 = true

	if got := format.ResolvedOptions().Hour12; got == nil || *got {
		t.Fatalf("ResolvedOptions().Hour12 = %v, want false", got)
	}
}
