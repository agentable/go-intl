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
	locales := mustLocaleList("en-us", "en-US", "zh-Hans-CN-u-nu-latn")

	for _, loc := range gointl.GetCanonicalLocales(locales) {
		fmt.Println(loc.String())
	}

	// Output:
	// en-US
	// zh-Hans-CN-u-nu-latn
}

func Example_numberFormat() {
	format, err := numberformat.New(mustLocaleList("en-US"), numberformat.Options{
		Style:    numberformat.CurrencyStyle,
		Currency: numberformat.Currency("USD"),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(format.Format(numberformat.Float(1234.5)))

	// Output:
	// $1,234.50
}

func Example_dateTimeFormat() {
	format, err := datetimeformat.New(mustLocaleList("en-US"), datetimeformat.Options{
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
	format, err := numberformat.New(mustLocaleList("en-US"), numberformat.Options{
		Style:    numberformat.CurrencyStyle,
		Currency: numberformat.Currency("EUR"),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(format.Format(numberformat.Float(1234.5)))
}

func mustLocaleList(tags ...string) locale.List {
	locales, err := locale.ParseList(tags...)
	if err != nil {
		panic(err)
	}
	return locales
}
