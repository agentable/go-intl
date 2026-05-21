package listformat_test

import (
	"fmt"

	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/locale"
)

// Example demonstrates Intl.ListFormat.prototype.format from ECMA-402.
func Example() {
	format, err := listformat.New(locale.MustParseList("en-US"), listformat.Options{})
	if err != nil {
		panic(err)
	}

	fmt.Println(format.Format([]string{"apples", "bananas", "cherries"}))

	// Output:
	// apples, bananas, and cherries
}

// Example_options demonstrates Intl.ListFormat constructor options from ECMA-402.
func Example_options() {
	format, err := listformat.New(locale.MustParseList("en-US"), listformat.Options{
		Type: listformat.Disjunction,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(format.Format([]string{"red", "green", "blue"}))

	// Output:
	// red, green, or blue
}

// ExampleListFormat_FormatToParts demonstrates Intl.ListFormat.prototype.formatToParts from ECMA-402.
func ExampleListFormat_FormatToParts() {
	format, err := listformat.New(locale.MustParseList("en-US"), listformat.Options{})
	if err != nil {
		panic(err)
	}

	for _, part := range format.FormatToParts([]string{"apples", "bananas"}) {
		fmt.Printf("%s=%q\n", part.Type, part.Value)
	}

	// Output:
	// element="apples"
	// literal=" and "
	// element="bananas"
}
