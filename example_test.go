package gointl_test

import (
	"fmt"
	"time"

	gointl "github.com/agentable/go-intl"
	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
)

func Example_getCanonicalLocales() {
	locales := []locale.Locale{
		locale.MustParse("en-us"),
		locale.MustParse("en-US"),
		locale.MustParse("zh-Hans-CN-u-nu-latn"),
	}

	for _, loc := range gointl.GetCanonicalLocales(locales...) {
		fmt.Println(loc.String())
	}

	// Output:
	// en-US
	// zh-Hans-CN-u-nu-latn
}

func Example_numberFormat() {
	loc := locale.MustParse("en-US")
	format, err := numberformat.New(loc, numberformat.Options{
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

func Example_dateTimeFormat() {
	loc := locale.MustParse("en-US")
	format, err := datetimeformat.New(loc, datetimeformat.Options{
		Year:  datetimeformat.NumericFieldStyle,
		Month: datetimeformat.ShortMonthStyle,
		Day:   datetimeformat.NumericFieldStyle,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(format.Format(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)))

	// Output:
	// May 8, 2026
}

func Example_directFormatter() {
	loc := locale.MustParse("en-US")
	format, err := numberformat.New(loc, numberformat.Options{
		Style:    numberformat.CurrencyStyle,
		Currency: numberformat.CurrencyCode("EUR"),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(format.FormatFloat64(1234.5))

	// Output:
	// EUR1,234.50
}
