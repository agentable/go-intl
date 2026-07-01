package relativetimeformat_test

import (
	"fmt"

	gointl "github.com/agentable/go-intl"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/relativetimeformat"
)

// Example demonstrates Intl.RelativeTimeFormat.prototype.format from ECMA-402.
func Example() {
	format, err := relativetimeformat.New(mustLocaleList("en-US"), relativetimeformat.Options{})
	if err != nil {
		panic(err)
	}

	out, err := format.Format(relativetimeformat.Int(3), relativetimeformat.Day)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)

	// Output:
	// in 3 days
}

// Example_options demonstrates Intl.RelativeTimeFormat constructor options from ECMA-402.
func Example_options() {
	format, err := relativetimeformat.New(mustLocaleList("en-US"), relativetimeformat.Options{
		Numeric: gointl.String(relativetimeformat.NumericAuto),
	})
	if err != nil {
		panic(err)
	}

	out, err := format.Format(relativetimeformat.Int(-1), relativetimeformat.Day)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)

	// Output:
	// yesterday
}

// ExampleRelativeTimeFormat_FormatToParts demonstrates Intl.RelativeTimeFormat.prototype.formatToParts from ECMA-402.
func ExampleRelativeTimeFormat_FormatToParts() {
	format, err := relativetimeformat.New(mustLocaleList("en-US"), relativetimeformat.Options{})
	if err != nil {
		panic(err)
	}

	parts, err := format.FormatToParts(relativetimeformat.Int(-2), relativetimeformat.Day)
	if err != nil {
		panic(err)
	}
	for _, part := range parts {
		fmt.Printf("%s=%q unit=%q\n", part.Type, part.Value, part.Unit)
	}

	// Output:
	// integer="2" unit="day"
	// literal=" days ago" unit=""
}

func mustLocaleList(tags ...string) locale.List {
	locales, err := locale.ParseList(tags...)
	if err != nil {
		panic(err)
	}
	return locales
}
