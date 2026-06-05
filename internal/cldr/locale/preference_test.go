package cldrlocale

import (
	"reflect"
	"testing"
	"time"
)

func TestPreferenceAccessors(t *testing.T) {
	t.Parallel()

	if got, want := HourCyclePreference("US"), []string{"h12", "h23"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HourCyclePreference(US) = %#v, want %#v", got, want)
	}
	if !HasHourCyclePreference("US") || HasHourCyclePreference("ZZ") {
		t.Fatalf("HasHourCyclePreference(US/ZZ) = %v/%v, want true/false", HasHourCyclePreference("US"), HasHourCyclePreference("ZZ"))
	}
	if got, want := HourCyclePreference("ZZ"), []string{"h23", "h12"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HourCyclePreference(ZZ) = %#v, want %#v", got, want)
	}
	if got, want := FirstDayOfWeek("US"), time.Sunday; got != want {
		t.Fatalf("FirstDayOfWeek(US) = %v, want %v", got, want)
	}
	if got, want := FirstDayOfWeek("DE"), time.Monday; got != want {
		t.Fatalf("FirstDayOfWeek(DE) = %v, want %v", got, want)
	}
	start, end := Weekend("US")
	if start != time.Saturday || end != time.Sunday {
		t.Fatalf("Weekend(US) = %v, %v", start, end)
	}
	if got, want := MinimalDaysInFirstWeek("DE"), 4; got != want {
		t.Fatalf("MinimalDaysInFirstWeek(DE) = %d, want %d", got, want)
	}
	if got, want := CalendarPreference("TH"), []string{"buddhist", "gregorian"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("CalendarPreference(TH) = %#v, want %#v", got, want)
	}
	if !HasCalendarPreference("TH") || HasCalendarPreference("ZZ") {
		t.Fatalf("HasCalendarPreference(TH/ZZ) = %v/%v, want true/false", HasCalendarPreference("TH"), HasCalendarPreference("ZZ"))
	}
	if got, want := CalendarPreference("ZZ"), []string{"gregorian"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("CalendarPreference(ZZ) = %#v, want %#v", got, want)
	}
	if !HasWeekPreference("US") || HasWeekPreference("ZZ") {
		t.Fatalf("HasWeekPreference(US/ZZ) = %v/%v, want true/false", HasWeekPreference("US"), HasWeekPreference("ZZ"))
	}
}
