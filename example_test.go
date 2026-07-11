package gointl_test

import (
	"errors"
	"fmt"
	"time"

	gointl "github.com/agentable/go-intl"
	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
)

// Example_errorKind classifies a constructor error by category through the
// exported Error.Kind field and the root ErrorKind constants — the struct-based
// alternative to errors.Is against a sentinel.
func Example_errorKind() {
	_, err := numberformat.New(mustLocaleList("en-US"), numberformat.Options{
		Style:    gointl.String(numberformat.CurrencyStyle),
		Currency: gointl.String("US"), // not a 3-letter ISO 4217 code
	})

	var intlErr *gointl.Error
	if errors.As(err, &intlErr) {
		switch intlErr.Kind {
		case gointl.InvalidOption:
			fmt.Println("invalid option:", intlErr.Name)
		default:
			fmt.Println("other:", intlErr.Kind)
		}
	}

	// Output:
	// invalid option: currency
}

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
		Style:    gointl.String(numberformat.CurrencyStyle),
		Currency: gointl.String("USD"),
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
		Year:     gointl.String(string(datetimeformat.NumericFieldStyle)),
		Month:    gointl.String(string(datetimeformat.ShortMonthStyle)),
		Day:      gointl.String(string(datetimeformat.NumericFieldStyle)),
		TimeZone: gointl.String("UTC"),
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
		Style:    gointl.String(numberformat.CurrencyStyle),
		Currency: gointl.String("EUR"),
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
