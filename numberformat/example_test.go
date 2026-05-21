package numberformat_test

import (
	"fmt"

	gointl "github.com/agentable/go-intl"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
)

// Example demonstrates Intl.NumberFormat.prototype.format with the 0.5
// rounding example from ECMA-402 §15.5.
func Example() {
	format, err := numberformat.New(locale.MustParseList("en-US"), numberformat.Options{
		MaximumFractionDigits: gointl.Int(0),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(format.Format(numberformat.Float(0.5)))

	// Output:
	// 1
}

// Example_options demonstrates the "floor" rounding-mode row from ECMA-402 §15.5.
func Example_options() {
	format, err := numberformat.New(locale.MustParseList("en-US"), numberformat.Options{
		MaximumFractionDigits: gointl.Int(0),
		RoundingMode:          numberformat.FloorRoundingMode,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(format.Format(numberformat.Float(-1.5)))

	// Output:
	// -2
}

// ExampleNumberFormat_FormatToParts demonstrates Intl.NumberFormat.prototype.formatToParts from ECMA-402.
func ExampleNumberFormat_FormatToParts() {
	format, err := numberformat.New(locale.MustParseList("en-US"), numberformat.Options{})
	if err != nil {
		panic(err)
	}

	for _, part := range format.FormatToParts(numberformat.Float(-1.5)) {
		fmt.Printf("%s=%q\n", part.Type, part.Value)
	}

	// Output:
	// minusSign="-"
	// integer="1"
	// decimal="."
	// fraction="5"
}
