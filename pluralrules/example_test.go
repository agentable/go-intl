package pluralrules_test

import (
	"fmt"

	gointl "github.com/agentable/go-intl"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/pluralrules"
)

// Example demonstrates Intl.PluralRules.prototype.select from ECMA-402.
func Example() {
	rules, err := pluralrules.New(mustLocaleList("en"), pluralrules.Options{})
	if err != nil {
		panic(err)
	}

	one, err := rules.Select(pluralrules.Int(1))
	if err != nil {
		panic(err)
	}
	other, err := rules.Select(pluralrules.Int(2))
	if err != nil {
		panic(err)
	}
	fmt.Println(one)
	fmt.Println(other)

	// Output:
	// one
	// other
}

// Example_options demonstrates Intl.PluralRules constructor options from ECMA-402.
func Example_options() {
	rules, err := pluralrules.New(mustLocaleList("en"), pluralrules.Options{
		Type: gointl.String(pluralrules.Ordinal),
	})
	if err != nil {
		panic(err)
	}

	for _, n := range []int64{1, 2, 3, 4} {
		category, err := rules.Select(pluralrules.Int(n))
		if err != nil {
			panic(err)
		}
		fmt.Printf("%d: %s\n", n, category)
	}

	// Output:
	// 1: one
	// 2: two
	// 3: few
	// 4: other
}

// ExamplePluralRules_SelectRange demonstrates Intl.PluralRules.prototype.selectRange from ECMA-402.
func ExamplePluralRules_SelectRange() {
	rules, err := pluralrules.New(mustLocaleList("en"), pluralrules.Options{})
	if err != nil {
		panic(err)
	}

	category, err := rules.SelectRange(pluralrules.Int(1), pluralrules.Int(2))
	if err != nil {
		panic(err)
	}
	fmt.Println(category)

	// Output:
	// other
}

func mustLocaleList(tags ...string) locale.List {
	locales, err := locale.ParseList(tags...)
	if err != nil {
		panic(err)
	}
	return locales
}
