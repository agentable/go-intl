package numberformat_test

import (
	"fmt"

	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
)

func ExampleNumberFormat_FormatFloat64() {
	format, err := numberformat.New(locale.MustParse("en-US"), numberformat.Options{
		Style:    numberformat.CurrencyStyle,
		Currency: numberformat.CurrencyCode("USD"),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(format.FormatFloat64(1234.5))

	// Output:
	// $1,234.50
}
