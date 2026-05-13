package locale_test

import (
	"fmt"

	"github.com/agentable/go-intl/locale"
)

func ExampleParse() {
	loc := locale.MustParse("en-us-u-ca-gregorian-hc-h23")
	fmt.Println(loc.String())
	// Output: en-US-u-ca-gregory-hc-h23
}
