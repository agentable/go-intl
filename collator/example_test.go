package collator_test

import (
	"fmt"
	"sort"

	gointl "github.com/agentable/go-intl"
	"github.com/agentable/go-intl/collator"
	"github.com/agentable/go-intl/locale"
)

// Example demonstrates Intl.Collator.prototype.compare from ECMA-402.
func Example() {
	compare, err := collator.New(mustLocaleList("en-US"), collator.Options{})
	if err != nil {
		panic(err)
	}

	fmt.Println(compare.Compare("a", "b") < 0)

	// Output:
	// true
}

// Example_options demonstrates Intl.Collator constructor options from ECMA-402.
func Example_options() {
	compare, err := collator.New(mustLocaleList("en-US"), collator.Options{
		Numeric: gointl.Bool(true),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(compare.Compare("2", "10") < 0)

	// Output:
	// true
}

// ExampleCollator_Compare demonstrates Intl.Collator.prototype.compare from ECMA-402.
func ExampleCollator_Compare() {
	compare, err := collator.New(mustLocaleList("en-US"), collator.Options{
		Numeric: gointl.Bool(true),
	})
	if err != nil {
		panic(err)
	}

	values := []string{"10", "2", "1"}
	sort.Slice(values, func(i, j int) bool {
		return compare.Compare(values[i], values[j]) < 0
	})
	fmt.Println(values)

	// Output:
	// [1 2 10]
}

func mustLocaleList(tags ...string) locale.List {
	locales, err := locale.ParseList(tags...)
	if err != nil {
		panic(err)
	}
	return locales
}
