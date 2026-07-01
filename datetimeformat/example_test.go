package datetimeformat_test

import (
	"fmt"
	"time"

	gointl "github.com/agentable/go-intl"
	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/locale"
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
	fmt.Println(format.Format(t))

	// Output:
	// 5/14/2020
}

// Example_options demonstrates Intl.DateTimeFormat constructor options from ECMA-402.
func Example_options() {
	format, err := datetimeformat.New(mustLocaleList("en-US"), datetimeformat.Options{
		DateStyle: gointl.String(string(datetimeformat.LongDateTimeStyle)),
		TimeZone:  gointl.String("UTC"),
	})
	if err != nil {
		panic(err)
	}

	t := time.Date(2020, time.May, 14, 15, 4, 0, 0, time.UTC)
	fmt.Println(format.Format(t))

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
	for _, part := range format.FormatToParts(t) {
		fmt.Printf("%s=%q\n", part.Type, part.Value)
	}

	// Output:
	// month="May"
	// literal=" "
	// day="14"
	// literal=", "
	// year="2020"
}

func mustLocaleList(tags ...string) locale.List {
	locales, err := locale.ParseList(tags...)
	if err != nil {
		panic(err)
	}
	return locales
}
