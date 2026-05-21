package locale_test

import (
	"fmt"

	"github.com/agentable/go-intl/locale"
)

// ExampleParse demonstrates Intl.Locale canonicalization from ECMA-402.
func ExampleParse() {
	loc := locale.MustParse("en-us-u-ca-gregorian-hc-h23")
	fmt.Println(loc.String())

	// Output:
	// en-US-u-ca-gregory-hc-h23
}

// ExampleNew_options demonstrates Intl.Locale constructor options from ECMA-402.
func ExampleNew_options() {
	loc, err := locale.New("en", locale.Options{
		Region:   "GB",
		Calendar: "gregory",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(loc.String())

	// Output:
	// en-GB-u-ca-gregory
}

// ExampleLocale_GetWeekInfo demonstrates Intl.Locale.prototype.getWeekInfo from ECMA-402.
func ExampleLocale_GetWeekInfo() {
	loc := locale.MustParse("en-GB")
	info := loc.GetWeekInfo()
	fmt.Println(info.FirstDay)
	fmt.Println(info.Weekend)

	// Output:
	// Monday
	// [Saturday Sunday]
}
