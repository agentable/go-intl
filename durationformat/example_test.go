package durationformat_test

import (
	"fmt"

	gointl "github.com/agentable/go-intl"
	"github.com/agentable/go-intl/durationformat"
	"github.com/agentable/go-intl/locale"
)

// Example demonstrates Intl.DurationFormat.prototype.format from ECMA-402.
func Example() {
	format, err := durationformat.New(mustLocaleList("en"), durationformat.Options{
		Style: gointl.String(durationformat.DigitalStyle),
	})
	if err != nil {
		panic(err)
	}

	out, err := format.Format(durationformat.Duration{
		Hours:   1,
		Minutes: 2,
		Seconds: 3,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(out)

	// Output:
	// 1:02:03
}

// Example_options demonstrates Intl.DurationFormat constructor options from ECMA-402.
func Example_options() {
	format, err := durationformat.New(mustLocaleList("en"), durationformat.Options{
		Style: gointl.String(durationformat.LongStyle),
	})
	if err != nil {
		panic(err)
	}

	out, err := format.Format(durationformat.Duration{
		Hours:   1,
		Minutes: 2,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(out)

	// Output:
	// 1 hour, 2 minutes
}

// ExampleDurationFormat_FormatToParts demonstrates Intl.DurationFormat.prototype.formatToParts from ECMA-402.
func ExampleDurationFormat_FormatToParts() {
	format, err := durationformat.New(mustLocaleList("en"), durationformat.Options{
		Style: gointl.String(durationformat.DigitalStyle),
	})
	if err != nil {
		panic(err)
	}

	parts, err := format.FormatToParts(durationformat.Duration{
		Hours:   1,
		Minutes: 2,
		Seconds: 3,
	})
	if err != nil {
		panic(err)
	}
	for _, part := range parts {
		fmt.Printf("%s=%q unit=%q\n", part.Type, part.Value, part.Unit)
	}

	// Output:
	// integer="1" unit="hour"
	// literal=":" unit=""
	// integer="02" unit="minute"
	// literal=":" unit=""
	// integer="03" unit="second"
}

func ExampleDurationFormat_Format_wideNumber() {
	format, err := durationformat.New(mustLocaleList("en"), durationformat.Options{
		Style: gointl.String(durationformat.DigitalStyle),
	})
	if err != nil {
		panic(err)
	}

	out, err := format.Format(durationformat.Duration{Nanoseconds: 1e20})
	if err != nil {
		panic(err)
	}

	fmt.Println(out)

	// Output:
	// 0:00:100000000000
}

func mustLocaleList(tags ...string) locale.List {
	locales, err := locale.ParseList(tags...)
	if err != nil {
		panic(err)
	}
	return locales
}
