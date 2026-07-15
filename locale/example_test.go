package locale_test

import (
	"fmt"

	"github.com/agentable/go-intl/locale"
)

// ExampleParse demonstrates Intl.Locale canonicalization from ECMA-402.
func ExampleParse() {
	loc := mustLocale("en-us-u-ca-islamicc-hc-h23")
	fmt.Println(loc.String())

	// Output:
	// en-US-u-ca-islamic-civil-hc-h23
}

// ExampleNew_options demonstrates Intl.Locale constructor options from ECMA-402.
func ExampleNew_options() {
	region := "GB"
	calendar := "gregory"
	loc, err := locale.New("en", locale.Options{
		Region:   &region,
		Calendar: &calendar,
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
	loc := mustLocale("en-GB")
	info := loc.GetWeekInfo()
	fmt.Println(info.FirstDay)
	fmt.Println(info.Weekend)

	// Output:
	// Monday
	// [Saturday Sunday]
}

func mustLocale(tag string) locale.Locale {
	loc, err := locale.Parse(tag)
	if err != nil {
		panic(err)
	}
	return loc
}
