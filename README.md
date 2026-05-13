# go-intl

[![Go Reference](https://pkg.go.dev/badge/github.com/agentable/go-intl.svg)](https://pkg.go.dev/github.com/agentable/go-intl)
[![Go Version](https://img.shields.io/badge/Go-1.26.2%2B-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Agentable%20Commercial-blue)](./LICENSE)

A Go implementation of the active ECMA-402 `Intl` constructors for supported locales and options

## Features

- **Native Intl alignment**: Public packages map to `Intl.Locale`, `Intl.NumberFormat`, `Intl.DateTimeFormat`, and `Intl.PluralRules`.
- **Typed Go bridge**: Pass `locale.Locale`, `time.Time`, typed option structs, and typed numeric methods while preserving ECMA-402 behavior.
- **Root namespace**: The root package represents the JavaScript `Intl` namespace; formatter construction stays with constructor packages.
- **Reusable formatters**: Construct package-level formatters when you need repeated formatting, parts, ranges, or resolved options.
- **CLDR-backed data**: Ship generated CLDR data as Go source; applications do not load JSON, ICU, or time-zone data files at runtime.
- **Reference fixtures**: Check formatter output against ECMA-402-derived FormatJS fixtures and `.references/node/` native Intl snapshots.

## Installation

```bash
go get github.com/agentable/go-intl
```

Requires **Go 1.26.2+**.

## Quick Start

Construct the formatter that matches the JavaScript `Intl` constructor you would use:

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
)

func main() {
	loc, err := locale.Parse("en-US")
	if err != nil {
		log.Fatal(err)
	}

	priceFormat, err := numberformat.New(loc, numberformat.Options{
		Style:    numberformat.CurrencyStyle,
		Currency: numberformat.CurrencyCode("USD"),
	})
	if err != nil {
		log.Fatal(err)
	}
	price := priceFormat.FormatFloat64(1234.5)

	dateFormat, err := datetimeformat.New(loc, datetimeformat.Options{
		DateStyle: datetimeformat.LongDateTimeStyle,
		TimeZone:  "America/New_York",
	})
	if err != nil {
		log.Fatal(err)
	}
	date := dateFormat.Format(time.Date(2026, time.May, 8, 14, 30, 0, 0, time.UTC))

	fmt.Println(price)
	fmt.Println(date)
}
```

Output:

```text
$1,234.50
May 8, 2026
```

## Packages

| Package | Use |
|---------|-----|
| `github.com/agentable/go-intl` | Root `Intl` namespace helpers and active constructor type aliases. |
| `github.com/agentable/go-intl/locale` | BCP 47 parsing, canonicalization, maximize/minimize, and locale info getters. |
| `github.com/agentable/go-intl/numberformat` | Decimal, percent, currency, unit, compact, scientific, engineering, parts, and range formatting. |
| `github.com/agentable/go-intl/datetimeformat` | Date/time styles, field-based formatting, time zones, parts, and date/time ranges. |
| `github.com/agentable/go-intl/pluralrules` | Cardinal and ordinal plural category selection, including range selection. |

See the [Go package documentation](https://pkg.go.dev/github.com/agentable/go-intl) for the full API reference.

## Native Intl Mapping

The Go packages follow the ownership of the native JavaScript `Intl` API:

| JavaScript | Go |
|------------|----|
| `Intl.getCanonicalLocales(locales)` | `gointl.GetCanonicalLocales(locales...)` |
| `Intl.supportedValuesOf(key)` | `gointl.SupportedValuesOf(key)` |
| `new Intl.Locale(tag, options)` | `locale.Parse(tag)` or `locale.New(tag, options)` |
| `new Intl.NumberFormat(locales, options)` | `numberformat.New(loc, options)` |
| `Intl.NumberFormat.supportedLocalesOf(locales, options)` | `numberformat.SupportedLocalesOf(locales, options)` |
| `new Intl.DateTimeFormat(locales, options)` | `datetimeformat.New(loc, options)` |
| `Intl.DateTimeFormat.supportedLocalesOf(locales, options)` | `datetimeformat.SupportedLocalesOf(locales, options)` |
| `new Intl.PluralRules(locales, options)` | `pluralrules.New(loc, options)` |
| `Intl.PluralRules.supportedLocalesOf(locales, options)` | `pluralrules.SupportedLocalesOf(locales, options)` |

## Usage

### Parse Locales Once

Parse BCP 47 tags at your application boundary, then pass `locale.Locale` through your formatting code:

```go
loc, err := locale.Parse("zh-Hant-TW-u-nu-hanidec")
if err != nil {
	return err
}

fmt.Println(loc.String())
fmt.Println(loc.Maximize().String())
fmt.Println(loc.GetWeekInfo().FirstDay)
```

Use `locale.MustParse` in tests and package-level examples where a hard-coded locale must be valid.

### Use Root Namespace Helpers

Import the root package when you need native `Intl` namespace functions:

```go
import (
	"fmt"

	gointl "github.com/agentable/go-intl"
	"github.com/agentable/go-intl/locale"
)

locales := gointl.GetCanonicalLocales(
	locale.MustParse("en-US"),
	locale.MustParse("en-US"),
)

units, err := gointl.SupportedValuesOf(gointl.SupportedValueUnit)
if err != nil {
	return err
}

fmt.Println(locales[0])
fmt.Println(units[:3])
```

`SupportedValuesOf` supports the ECMA-402 keys `calendar`, `collation`, `currency`, `numberingSystem`, `timeZone`, and `unit`.

### Use Constructor Packages Directly

Use constructor packages directly when calls share locale, options, parts, ranges, or resolved options:

```go
compactFormat, err := numberformat.New(locale.MustParse("en-US"), numberformat.Options{
	Notation: numberformat.CompactNotation,
})
if err != nil {
	return err
}

rules, err := pluralrules.New(locale.MustParse("en-US"))
if err != nil {
	return err
}

fmt.Println(compactFormat.FormatInt(1200))
fmt.Println(rules.SelectInt(1))
```

Formatter-specific options stay in their formatter packages. The root package does not re-export `numberformat`, `datetimeformat`, or `pluralrules` option names.

### Filter Supported Locales

Constructor static methods stay with their packages, just like native `Intl.<Constructor>.supportedLocalesOf`:

```go
requested := []locale.Locale{
	locale.MustParse("de-DE"),
	locale.MustParse("en-US-u-nu-latn"),
	locale.MustParse("zh-Hans-CN"),
}

supported, err := numberformat.SupportedLocalesOf(requested, numberformat.Options{
	LocaleMatcher: numberformat.LookupLocaleMatcher,
})
if err != nil {
	return err
}

fmt.Println(supported)
```

The result preserves the requested locale order and returns requested locale values, including Unicode extensions.

### Construct Formatters for Repeated Work

Use formatter packages directly when you need resolved options, parts, ranges, or repeated calls in a hot path:

```go
format, err := numberformat.New(locale.MustParse("en-US"), numberformat.Options{
	Style:       numberformat.UnitStyle,
	Unit:        numberformat.UnitIdentifier("kilometer-per-hour"),
	UnitDisplay: numberformat.ShortUnitDisplay,
})
if err != nil {
	return err
}

fmt.Println(format.FormatInt(88))
fmt.Println(format.ResolvedOptions().Unit)
```

Unit identifiers follow native `Intl.NumberFormat`: use canonical lowercase ECMA-402 identifiers such as `meter`, `microsecond`, or `kilometer-per-hour`.

### Format Exact Decimals

Use decimal-string methods when binary `float64` cannot represent the value you need to format:

```go
format, err := numberformat.New(locale.MustParse("en-US"), numberformat.Options{
	Style: numberformat.PercentStyle,
	FractionDigits: numberformat.FractionDigits(2, 2),
})
if err != nil {
	return err
}

out, err := format.FormatDecimal("0.075")
if err != nil {
	return err
}

fmt.Println(out)
```

### Inspect Parts

Use parts APIs when you need to style or transform individual formatted tokens:

```go
format, err := numberformat.New(locale.MustParse("en-US"), numberformat.Options{
	Style:    numberformat.CurrencyStyle,
	Currency: numberformat.CurrencyCode("USD"),
})
if err != nil {
	return err
}

for _, part := range format.FormatFloat64ToParts(1234.5) {
	fmt.Printf("%s: %q\n", part.Type, part.Value)
}
```

### Format Dates and Ranges

`datetimeformat` accepts `time.Time` and supports style-based or field-based formatting:

```go
format, err := datetimeformat.New(locale.MustParse("en-US"), datetimeformat.Options{
	DateStyle: datetimeformat.MediumDateTimeStyle,
	TimeStyle: datetimeformat.ShortDateTimeStyle,
	TimeZone:  "America/New_York",
})
if err != nil {
	return err
}

start := time.Date(2026, time.May, 8, 14, 30, 0, 0, time.UTC)
end := start.Add(2 * time.Hour)

fmt.Println(format.Format(start))
fmt.Println(format.FormatRange(start, end))
```

### Select Plural Categories

Use `pluralrules` to select CLDR plural categories for message selection:

```go
rules, err := pluralrules.New(locale.MustParse("en"), pluralrules.Options{
	Type: pluralrules.Ordinal,
})
if err != nil {
	return err
}

fmt.Println(rules.SelectInt(1))
fmt.Println(rules.SelectInt(2))
fmt.Println(rules.SelectInt(3))
fmt.Println(rules.SelectInt(4))
```

Output:

```text
one
two
few
other
```

## Supported Data

This repository currently ships a focused generated data profile:

| Surface | Supported locales in generated formatter data |
|---------|-----------------------------------------------|
| Locale info and likely-subtag data | See `tools/locale-profile.json`; 104 locale tags are included. |
| NumberFormat and DateTimeFormat | Profile-limited generated CLDR payloads; use `SupportedLocalesOf` to inspect exact constructor support. |
| PluralRules | `ar`, `en`, `en-US`, `fr`, `hi`, `pl`, `zh`, `zh-Hans`, `zh-Hans-CN`. |

Locale parsing accepts BCP 47 tags through `golang.org/x/text/language`; generated locale info and formatter data are profile-limited. Locale negotiation follows ECMA-402 against the generated supported-locale data. Expand `tools/locale-profile.json` and regenerate CLDR data when your application needs a broader locale set.

`SupportedValuesOf` returns ECMA-402 root values: calendars include `iso8601`, numbering systems include the simple digit systems from ECMA-402, units come from the sanctioned unit list, and currencies/time zones/collations come from generated CLDR/tz data.

## Error Handling

Constructors return package sentinels that work with `errors.Is`:

| Error | Meaning |
|-------|---------|
| `locale.ErrInvalidLocale` | Locale parsing rejected the input locale. |
| `numberformat.ErrInvalidOption` | NumberFormat received an invalid option. |
| `datetimeformat.ErrInvalidOption` | DateTimeFormat received an invalid option. |
| `datetimeformat.ErrUnsupportedTimeZone` | DateTimeFormat received an unsupported time zone. |
| `pluralrules.ErrInvalidOption` | PluralRules received an invalid option. |

```go
_, err := numberformat.New(locale.MustParse("en-US"), numberformat.Options{
	Style: numberformat.CurrencyStyle,
})
if err != nil {
	if errors.Is(err, numberformat.ErrInvalidOption) {
		return fmt.Errorf("fix formatter options: %w", err)
	}
	return err
}
```

## Development

```bash
task deps                 # Download modules and tidy go.mod/go.sum
task test                 # Run go test -race -p 1 ./...
task lint                 # Run go mod tidy check and golangci-lint
task conformance:verify   # Validate fixture schema and divergence IDs
task data:verify          # Regenerate CLDR data into a temp tree and compare byte-for-byte
task data:contract        # Verify generated CLDR data contracts
task build:size           # Check CLDR binary size budget
task bench:gate           # Run tagged benchmark budget tests
task verify               # Run deps, fmt, vet, lint, test, conformance, data contract, and vuln checks
```

Run a targeted package while developing:

```bash
go test -race ./numberformat/...
go test -race -run TestPluralRules_Cardinal/en ./pluralrules/
(cd tools/gen-fixtures-from-formatjs && go test ./...)
```

## Documentation

- [SPECS](./SPECS/) define public contracts, formatter behavior, data layout, and conformance rules.
- [AGENTS.md](./AGENTS.md) defines development workflow and repository conventions for AI coding agents.

## Contributing

Open an issue before changing public behavior. Keep README changes focused on installation, usage, examples, and development commands; put contracts in `SPECS/` and agent workflow rules in `AGENTS.md`.

## License

This software is licensed under the **Agentable Commercial License**, exclusively for use with Agentable platform services and their direct integrations.
See the [LICENSE](./LICENSE) file for full terms.
