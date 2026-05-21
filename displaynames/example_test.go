package displaynames_test

import (
	"fmt"

	"github.com/agentable/go-intl/displaynames"
	"github.com/agentable/go-intl/locale"
)

// Example demonstrates Intl.DisplayNames.prototype.of from ECMA-402.
func Example() {
	names, err := displaynames.New(locale.MustParseList("en-US"), displaynames.Options{
		Type: displaynames.Language,
	})
	if err != nil {
		panic(err)
	}

	name, ok, err := names.Of("fr")
	if err != nil {
		panic(err)
	}
	fmt.Println(name, ok)

	// Output:
	// French true
}

// Example_options demonstrates Intl.DisplayNames constructor options from ECMA-402.
func Example_options() {
	names, err := displaynames.New(locale.MustParseList("en-US"), displaynames.Options{
		Type:  displaynames.Region,
		Style: displaynames.ShortStyle,
	})
	if err != nil {
		panic(err)
	}

	name, ok, err := names.Of("US")
	if err != nil {
		panic(err)
	}
	fmt.Println(name, ok)

	// Output:
	// US true
}

// ExampleDisplayNames_Of demonstrates Intl.DisplayNames.prototype.of from ECMA-402.
func ExampleDisplayNames_Of() {
	names, err := displaynames.New(locale.MustParseList("en-US"), displaynames.Options{
		Type:     displaynames.Language,
		Fallback: displaynames.NoneFallback,
	})
	if err != nil {
		panic(err)
	}

	name, ok, err := names.Of("zz")
	if err != nil {
		panic(err)
	}
	fmt.Println(name, ok)

	// Output:
	//  false
}
