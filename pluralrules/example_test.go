package pluralrules_test

import (
	"fmt"

	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/pluralrules"
)

func ExamplePluralRules_SelectInt() {
	rules, err := pluralrules.New(locale.MustParse("en"))
	if err != nil {
		panic(err)
	}

	fmt.Println(rules.SelectInt(1))
	fmt.Println(rules.SelectInt(2))

	// Output:
	// one
	// other
}
