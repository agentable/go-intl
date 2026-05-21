package relativetimeformat_test

import (
	"fmt"

	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/relativetimeformat"
)

// Example demonstrates Intl.RelativeTimeFormat.prototype.format from ECMA-402.
func Example() {
	format, err := relativetimeformat.New(locale.MustParseList("en-US"), relativetimeformat.Options{})
	if err != nil {
		panic(err)
	}

	out, err := format.FormatInt(3, relativetimeformat.Day)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)

	// Output:
	// in 3 days
}

// Example_options demonstrates Intl.RelativeTimeFormat constructor options from ECMA-402.
func Example_options() {
	format, err := relativetimeformat.New(locale.MustParseList("en-US"), relativetimeformat.Options{
		Numeric: relativetimeformat.NumericAuto,
	})
	if err != nil {
		panic(err)
	}

	out, err := format.FormatInt(-1, relativetimeformat.Day)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)

	// Output:
	// yesterday
}

// ExampleRelativeTimeFormat_FormatIntToParts demonstrates Intl.RelativeTimeFormat.prototype.formatToParts from ECMA-402.
func ExampleRelativeTimeFormat_FormatIntToParts() {
	format, err := relativetimeformat.New(locale.MustParseList("en-US"), relativetimeformat.Options{})
	if err != nil {
		panic(err)
	}

	parts, err := format.FormatIntToParts(-2, relativetimeformat.Day)
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
