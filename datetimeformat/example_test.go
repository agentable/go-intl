package datetimeformat_test

import (
	"fmt"
	"time"

	gointl "github.com/agentable/go-intl"
	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/locale"
)

var (
	_ func(*datetimeformat.DateTimeFormat, time.Time, time.Time) (string, error)                     = (*datetimeformat.DateTimeFormat).FormatRange
	_ func(*datetimeformat.DateTimeFormat, time.Time, time.Time) ([]datetimeformat.RangePart, error) = (*datetimeformat.DateTimeFormat).FormatRangeToParts
	_ func(*gointl.DateTimeFormat, time.Time, time.Time) (string, error)                             = (*gointl.DateTimeFormat).FormatRange
	_ func(*gointl.DateTimeFormat, time.Time, time.Time) ([]datetimeformat.RangePart, error)         = (*gointl.DateTimeFormat).FormatRangeToParts
)

// Example demonstrates Intl.DateTimeFormat.prototype.format from ECMA-402.
func Example() {
	format, err := datetimeformat.New(mustLocaleList("en-US"), datetimeformat.Options{
		TimeZone: gointl.String("UTC"),
	})
	if err != nil {
		panic(err)
	}

	t := time.Date(2020, time.May, 14, 15, 4, 0, 0, time.UTC)
	out, err := format.Format(t)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)

	// Output:
	// 5/14/2020
}

// Example_options demonstrates Intl.DateTimeFormat constructor options from ECMA-402.
func Example_options() {
	format, err := datetimeformat.New(mustLocaleList("en-US"), datetimeformat.Options{
		DateStyle: gointl.String(string(datetimeformat.LongStyle)),
		TimeZone:  gointl.String("UTC"),
	})
	if err != nil {
		panic(err)
	}

	t := time.Date(2020, time.May, 14, 15, 4, 0, 0, time.UTC)
	out, err := format.Format(t)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)

	// Output:
	// May 14, 2020
}

// ExampleDateTimeFormat_FormatToParts demonstrates Intl.DateTimeFormat.prototype.formatToParts from ECMA-402.
func ExampleDateTimeFormat_FormatToParts() {
	format, err := datetimeformat.New(mustLocaleList("en-US"), datetimeformat.Options{
		Year:     gointl.String(string(datetimeformat.NumericFieldStyle)),
		Month:    gointl.String(string(datetimeformat.ShortMonthStyle)),
		Day:      gointl.String(string(datetimeformat.NumericFieldStyle)),
		TimeZone: gointl.String("UTC"),
	})
	if err != nil {
		panic(err)
	}

	t := time.Date(2020, time.May, 14, 15, 4, 0, 0, time.UTC)
	parts, err := format.FormatToParts(t)
	if err != nil {
		panic(err)
	}
	for _, part := range parts {
		fmt.Printf("%s=%q\n", part.Type, part.Value)
	}

	// Output:
	// month="May"
	// literal=" "
	// day="14"
	// literal=", "
	// year="2020"
}

// ExampleDateTimeFormat_FormatRange demonstrates direct DateTimeFormat range results.
func ExampleDateTimeFormat_FormatRange() {
	format, err := datetimeformat.New(mustLocaleList("en-US"), datetimeformat.Options{
		Year:     gointl.String(string(datetimeformat.NumericFieldStyle)),
		Month:    gointl.String(string(datetimeformat.ShortMonthStyle)),
		Day:      gointl.String(string(datetimeformat.NumericFieldStyle)),
		TimeZone: gointl.String("UTC"),
	})
	if err != nil {
		panic(err)
	}

	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	out, err := format.FormatRange(date, date)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
	parts, err := format.FormatRangeToParts(date, date)
	if err != nil {
		panic(err)
	}
	for _, part := range parts {
		fmt.Print(part.Value)
	}
	fmt.Println()

	// Output:
	// May 8, 2026
	// May 8, 2026
}

func mustLocaleList(tags ...string) locale.List {
	locales, err := locale.ParseList(tags...)
	if err != nil {
		panic(err)
	}
	return locales
}
