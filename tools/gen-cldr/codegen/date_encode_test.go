package codegen

import (
	"testing"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

func TestAppendDayPeriodRange(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	var e blobEncoder
	appendDayPeriodRange(&e, cldr.DayPeriodRange{
		From: 2,
		To:   5,
		Type: "am",
	}, table)

	assertBytesEqual(t, "appendDayPeriodRange() bytes", e.bytes(), []byte{2, 5, 0, 2})
}

func TestEncodeCalendarNames(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	var e blobEncoder
	encodeCalendarNames(&e, cldr.CalendarNames{
		Eras:       []string{"e"},
		Months:     []string{"m"},
		Weekdays:   []string{"w"},
		DayPeriods: []string{"d"},
	}, table)

	assertBytesEqual(t, "encodeCalendarNames() bytes", e.bytes(), []byte{1, 0, 1, 1, 1, 1, 1, 2, 1, 1, 3, 1})
}

func TestEncodeIntervalSkeletons(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	var e blobEncoder
	encodeIntervalSkeletons(&e, map[string]map[string]string{
		"yMd": {"d": "a"},
		"Hm":  {"H": "bb"},
	}, table)

	want := []byte{
		2,
		0, 2, 1, 2, 1, 3, 2,
		5, 3, 1, 8, 1, 9, 1,
	}
	assertBytesEqual(t, "encodeIntervalSkeletons() bytes", e.bytes(), want)
}
