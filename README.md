# go-intl

[![Go Reference](https://pkg.go.dev/badge/github.com/agentable/go-intl.svg)](https://pkg.go.dev/github.com/agentable/go-intl)
[![Go Version](https://img.shields.io/badge/Go-1.27%2B-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

A Go implementation of the active ECMA-402 `Intl` API with typed constructors and generated CLDR data.

## Features

- **Native Intl alignment**: Public packages map to the active `Intl` constructors: `Locale`, `NumberFormat`, `DateTimeFormat`, `PluralRules`, `ListFormat`, `RelativeTimeFormat`, `DurationFormat`, and `DisplayNames`.
- **Typed Go bridge**: Parse locale strings into `locale.List`, then pass `time.Time`, typed option structs, and typed numeric values while preserving ECMA-402 behavior.
- **Root namespace**: The root package represents the JavaScript `Intl` namespace as an aggregate facade; production services that need one formatter should import that constructor package directly.
- **Reusable formatters**: Construct once and reuse; constructors resolve locale, options, and data so repeated formatting stays on the cached path.
- **Host-friendly records**: Resolved options, parts, ranges, locale info, and durations marshal with ECMA-402 JSON field names for API and JS-host boundaries.
- **Structured errors**: Root sentinels work with `errors.Is`, and `gointl.Error` exposes stable kind, owner, option, value, locale, and expected-value guidance.
- **CLDR-backed data**: Ship generated CLDR data as Go source; formatter output needs no runtime CLDR JSON loading or ICU engine.
- **Reference fixtures**: Verify formatter output against ECMA-402-derived FormatJS fixtures and native Intl snapshots.

## Installation

```bash
go get github.com/agentable/go-intl
```

Requires **Go 1.27.0**.

## Quick Start

Construct the formatter that matches the JavaScript `Intl` constructor you would use:

```go
package main

import (
	"fmt"
	"log"
	"time"

	gointl "github.com/agentable/go-intl"
	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
)

func main() {
	locales, err := locale.ParseList("en-US")
	if err != nil {
		log.Fatal(err)
	}

	priceFormat, err := numberformat.New(locales, numberformat.Options{
		Style:    gointl.String(numberformat.CurrencyStyle),
		Currency: gointl.String("USD"),
	})
	if err != nil {
		log.Fatal(err)
	}
	price := priceFormat.Format(numberformat.Float(1234.5))

	dateFormat, err := datetimeformat.New(locales, datetimeformat.Options{
		DateStyle: gointl.String(datetimeformat.LongDateTimeStyle),
		TimeZone:  gointl.String("America/New_York"),
	})
	if err != nil {
		log.Fatal(err)
	}
	date, err := dateFormat.Format(time.Date(2026, time.May, 8, 14, 30, 0, 0, time.UTC))
	if err != nil {
		log.Fatal(err)
	}

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
| `github.com/agentable/go-intl` | Root `Intl` namespace helpers, active constructor type aliases, supported values, and structured error categories. |
| `github.com/agentable/go-intl/locale` | BCP 47 parsing, request lists, canonical locale identity, maximize/minimize, and locale info getters. |
| `github.com/agentable/go-intl/numberformat` | Decimal, percent, currency, unit, compact, scientific, engineering, parts, and range formatting. |
| `github.com/agentable/go-intl/datetimeformat` | Date/time styles, field-based formatting, time zones, parts, and date/time ranges. |
| `github.com/agentable/go-intl/pluralrules` | Cardinal and ordinal plural category selection, including range selection. |
| `github.com/agentable/go-intl/listformat` | Locale-sensitive list formatting and list parts. |
| `github.com/agentable/go-intl/relativetimeformat` | Locale-sensitive relative-time formatting and parts. |
| `github.com/agentable/go-intl/durationformat` | Locale-sensitive duration formatting and parts. |
| `github.com/agentable/go-intl/displaynames` | Localized names for languages, regions, scripts, currencies, calendars, and date-time fields. |
| `github.com/agentable/go-intl/option` | Zero-dependency `Int`/`Bool`/`String` pointer helpers for optional scalar options; usable without importing the aggregate root. |

Prefer constructor subpackages in services that need one formatter. Importing `github.com/agentable/go-intl` is an aggregate facade: it mirrors the JavaScript `Intl` namespace and therefore imports every active constructor surface for namespace helpers and type aliases. Use the root package for `GetCanonicalLocales`, root supported-value accessors, or when you intentionally want the full `Intl` namespace shape.

See the [Go package documentation](https://pkg.go.dev/github.com/agentable/go-intl) for the full API reference.

## Experimental Modules

`github.com/agentable/go-intl/x` is the experimental satellite module: Go bridges for in-flight TC39 proposals and non-surface exposure of finished-spec machinery, governed by [SPEC 80](SPECS/80-x-modules.md). It stays **v0 with no compatibility promise** — pin an exact version.

| Package | Use |
|---------|-----|
| `github.com/agentable/go-intl/x/localematcher` | Match requested locales against your available locales with the same ECMA-402 lookup/best-fit algorithms and CLDR data that back the `Intl` constructors (Stage 1 `Intl.LocaleMatcher` shape). |

```bash
go get github.com/agentable/go-intl/x/localematcher
```

```go
result := localematcher.Match(
    []string{"zh-HK"},
    []string{"en", "zh-Hans", "zh-Hant", "ja"},
    "en",
    localematcher.AlgorithmBestFit,
)
fmt.Println(result.Locale) // zh-Hant
```

See [x/README.md](x/README.md) for the module charter, versioning, and graduation policy.

## Native Intl Mapping

The Go packages follow the ownership of the native JavaScript `Intl` API:

| JavaScript | Go |
|------------|----|
| `Intl.getCanonicalLocales(locales)` | `gointl.GetCanonicalLocales(locales)` |
| `Intl.supportedValuesOf("calendar")` | `gointl.SupportedCalendars()` |
| `Intl.supportedValuesOf("currency")` | `gointl.SupportedCurrencies()` |
| `Intl.supportedValuesOf("numberingSystem")` | `gointl.SupportedNumberingSystems()` |
| `Intl.supportedValuesOf("timeZone")` | `gointl.SupportedTimeZones()` |
| `Intl.supportedValuesOf("unit")` | `gointl.SupportedUnits()` |
| `new Intl.Locale(tag, options)` | `locale.Parse(tag)`, `locale.New(tag, options)`, or `locale.FromTag(tag, options)` |
| `new Intl.NumberFormat(locales, options)` | `numberformat.New(locales, options)` |
| `Intl.NumberFormat.supportedLocalesOf(locales, options)` | `numberformat.SupportedLocalesOf(locales, options)` |
| `new Intl.DateTimeFormat(locales, options)` | `datetimeformat.New(locales, options)` |
| `Intl.DateTimeFormat.supportedLocalesOf(locales, options)` | `datetimeformat.SupportedLocalesOf(locales, options)` |
| `new Intl.PluralRules(locales, options)` | `pluralrules.New(locales, options)` |
| `Intl.PluralRules.supportedLocalesOf(locales, options)` | `pluralrules.SupportedLocalesOf(locales, options)` |
| `new Intl.ListFormat(locales, options)` | `listformat.New(locales, options)` |
| `Intl.ListFormat.supportedLocalesOf(locales, options)` | `listformat.SupportedLocalesOf(locales, options)` |
| `new Intl.RelativeTimeFormat(locales, options)` | `relativetimeformat.New(locales, options)` |
| `Intl.RelativeTimeFormat.supportedLocalesOf(locales, options)` | `relativetimeformat.SupportedLocalesOf(locales, options)` |
| `new Intl.DurationFormat(locales, options)` | `durationformat.New(locales, options)` |
| `Intl.DurationFormat.supportedLocalesOf(locales, options)` | `durationformat.SupportedLocalesOf(locales, options)` |
| `new Intl.DisplayNames(locales, options)` | `displaynames.New(locales, options)` |
| `Intl.DisplayNames.supportedLocalesOf(locales, options)` | `displaynames.SupportedLocalesOf(locales, options)` |

## Usage

### Parse Locales Once

Parse BCP 47 tags at your application boundary. Formatter constructors receive a `locale.List`: use `nil` or `locale.List{}` for omitted locales, and `locale.ParseList` for request lists that can fail.

Use the root `gointl.GetCanonicalLocales` helper when you need the ECMA-402 `Intl.getCanonicalLocales` operation. The lower-level locale-list canonicalization step is internal to constructors and `SupportedLocalesOf`.

```go
locales, err := locale.ParseList("zh-Hant-TW-u-nu-hanidec")
if err != nil {
	return err
}
loc := locales[0]

fmt.Println(loc.String())
fmt.Println(loc.Maximize().String())
fmt.Println(loc.GetWeekInfo().FirstDay)
if direction := loc.GetTextInfo().Direction; direction != nil {
	fmt.Println(*direction)
}
```

`GetTextInfo().Direction` is present only when the generated CLDR script
metadata has a known `ltr` or `rtl` value. Unknown direction remains `nil` and
is omitted from JSON instead of being guessed.

Canonicalization follows CLDR language aliases: `twi` becomes `ak`, while
`no` and `nb` retain their distinct canonical identities.

Short examples below use a local `mustLocaleList` helper for brevity; production code should call `locale.ParseList` and handle the returned error.

### Use Constructor Packages Directly

Use constructor packages directly in production services that need one `Intl` constructor, or when calls share locale, options, parts, ranges, or resolved options. This keeps the dependency graph tied to the formatter you use; the root package is measured as aggregate facade cost.

```go
locales := mustLocaleList("en-US")

compactFormat, err := numberformat.New(locales, numberformat.Options{
	Notation: gointl.String(numberformat.CompactNotation),
})
if err != nil {
	return err
}

rules, err := pluralrules.New(locales, pluralrules.Options{})
if err != nil {
	return err
}

category := rules.Select(pluralrules.Int(1))

fmt.Println(compactFormat.Format(numberformat.Int(1200)))
fmt.Println(category)
```

Formatter-specific options stay in their formatter packages. The root package does not re-export `numberformat`, `datetimeformat`, or `pluralrules` option names.

### Filter Supported Locales

Constructor static methods stay with their packages, just like native `Intl.<Constructor>.supportedLocalesOf`:

```go
requested := mustLocaleList("de-DE", "en-US-u-nu-latn", "zh-Hans-CN")

supported, err := numberformat.SupportedLocalesOf(requested, numberformat.Options{
	LocaleMatcher: gointl.String(numberformat.LookupLocaleMatcher),
})
if err != nil {
	return err
}

fmt.Println(supported)
```

The result preserves the requested locale order and returns requested locale values, including Unicode extensions.

### Construct Formatters for Repeated Work

Use formatter packages directly when you need resolved options, parts, ranges, or repeated calls in a hot path. Constructors do the locale, option, and data setup; keep the formatter and call its methods instead of rebuilding it per value.

```go
format, err := numberformat.New(mustLocaleList("en-US"), numberformat.Options{
	Style:       gointl.String(numberformat.UnitStyle),
	Unit:        gointl.String("kilometer-per-hour"),
	UnitDisplay: gointl.String(numberformat.ShortUnitDisplay),
})
if err != nil {
	return err
}

fmt.Println(format.Format(numberformat.Int(88)))
if unit := format.ResolvedOptions().Unit; unit != nil {
	fmt.Println(*unit)
}
```

Output:

```text
88 km/h
kilometer-per-hour
```

Unit identifiers follow native `Intl.NumberFormat`: use canonical lowercase ECMA-402 identifiers such as `meter`, `microsecond`, or `kilometer-per-hour`.

### Format Localized Currency Names

Use `CurrencyDisplayName` when the caller needs localized words instead of a
symbol. The formatter owns the locale-specific name placement and sign spacing;
both the name and its position may change with the rounded value's plural
category, so do not concatenate the currency name around formatted output.
Visible fraction digits introduced by digit options and scientific,
engineering, or compact exponents participate in that selection exactly as
they do in native `Intl`.

```go
format, err := numberformat.New(mustLocaleList("sw"), numberformat.Options{
	Style:           gointl.String(numberformat.CurrencyStyle),
	Currency:        gointl.String("USD"),
	CurrencyDisplay: gointl.String(numberformat.CurrencyDisplayName),
})
if err != nil {
	return err
}

fmt.Println(format.Format(numberformat.Int(123)))
```

Output:

```text
dola za Marekani 123.00
```

### Use Root Namespace Helpers

Import the root package when you need native `Intl` namespace functions or pointer helpers for optional scalar options:

```go
import (
	"fmt"

	gointl "github.com/agentable/go-intl"
	"github.com/agentable/go-intl/locale"
)

locales := gointl.GetCanonicalLocales(mustLocaleList("en-US", "en-US"))
units := gointl.SupportedUnits()

fmt.Println(locales[0])
fmt.Println(units[:3])
```

Root supported-value accessors cover calendars, currencies, numbering systems,
time zones, and units.

Use `gointl.Int`, `gointl.Bool`, and `gointl.String` for optional option fields where ECMA-402 distinguishes omitted from an explicit zero, false, or empty value.

These helpers also live in the zero-dependency leaf package
`github.com/agentable/go-intl/option` as `option.Int`, `option.Bool`, and
`option.String`. A service that uses a single formatter package can set its
optional scalar options through `option` without importing the aggregate root,
which pulls in every constructor package. The root re-exports the same helpers
as `gointl.Int`/`gointl.Bool`/`gointl.String` for namespace fidelity, so the
root-namespace examples above stay the primary documented style.

### Format Exact Decimals

Use `Decimal` when binary `float64` cannot represent the value you need to
format. It accepts Intl numeric strings, including surrounding ECMAScript
whitespace, empty strings (zero), and unsigned `0b`/`0o`/`0x` integers.
Malformed strings return an error. Finite nonzero values retain their exact
precision; strings beyond the Number rounding range become signed infinity or
signed zero. Use `BigInt` for integers without that range normalization:

```go
format, err := numberformat.New(mustLocaleList("en-US"), numberformat.Options{
	Style:                 gointl.String(numberformat.PercentStyle),
	MinimumFractionDigits: gointl.Int(2),
	MaximumFractionDigits: gointl.Int(2),
})
if err != nil {
	return err
}

value, err := numberformat.Decimal("0.075")
if err != nil {
	return err
}

out := format.Format(value)

fmt.Println(out)
```

### Inspect Parts

Use parts APIs when you need to style or transform individual formatted tokens:

```go
format, err := numberformat.New(mustLocaleList("en-US"), numberformat.Options{
	Style:    gointl.String(numberformat.CurrencyStyle),
	Currency: gointl.String("USD"),
})
if err != nil {
	return err
}

for _, part := range format.FormatToParts(numberformat.Float(1234.5)) {
	fmt.Printf("%s: %q\n", part.Type, part.Value)
}
```

Keep literal parts when rebuilding the output. They include spacing and
invisible direction marks needed by Arabic and Persian text; concatenating all
part values reproduces `Format`. Range parts additionally report `startRange`,
`endRange`, or `shared` ownership.

### Format Number Ranges

`FormatRange` and `FormatRangeToParts` preserve input order and use native Intl
range semantics. When both endpoints render to the same visible text, the result
uses the locale approximate sign; when notation or suffixes render differently,
the endpoints remain a real range.

```go
format, err := numberformat.New(mustLocaleList("en-US"), numberformat.Options{
	MaximumFractionDigits: gointl.Int(0),
})
if err != nil {
	return err
}

text, err := format.FormatRange(numberformat.Float(1.1), numberformat.Float(1.2))
if err != nil {
	return err
}

fmt.Println(text)
```

Output:

```text
~1
```

### Marshal Records for Host Boundaries

Public ECMA-402 records use camelCase JSON keys, so API adapters and JavaScript
hosts can pass them through without rebuilding field maps:

```go
format, err := numberformat.New(mustLocaleList("en-US"), numberformat.Options{
	Style:    gointl.String(numberformat.CurrencyStyle),
	Currency: gointl.String("USD"),
})
if err != nil {
	return err
}

data, err := json.Marshal(format.ResolvedOptions())
if err != nil {
	return err
}

fmt.Println(string(data))
```

`locale.Locale` marshals as its canonical BCP 47 string. Parts, range parts,
`durationformat.Duration` and locale `WeekInfo` / `TextInfo` also use ECMA-402
field names.

Inactive resolved-option properties are omitted. Explicit zero and false values
remain present: a disabled NumberFormat grouping mode marshals as the JSON
boolean `"useGrouping": false`, and a resolved 24-hour clock includes
`"hour12": false`. Date/time styles report their style options without exposing
the individual component fields. See [JSON records](SPECS/73-json-records.md)
for the complete presence policy.

### Format Dates and Ranges

`datetimeformat` accepts `time.Time` and supports style-based or field-based formatting:

```go
format, err := datetimeformat.New(mustLocaleList("en-US"), datetimeformat.Options{
	DateStyle: gointl.String(datetimeformat.MediumDateTimeStyle),
	TimeStyle: gointl.String(datetimeformat.ShortDateTimeStyle),
	TimeZone:  gointl.String("America/New_York"),
})
if err != nil {
	return err
}

start := time.Date(2026, time.May, 8, 14, 30, 0, 0, time.UTC)
end := start.Add(2 * time.Hour)

text, err := format.Format(start)
if err != nil {
	return err
}
fmt.Println(text)
rangeText, err := format.FormatRange(start, end)
if err != nil {
	return err
}
fmt.Println(rangeText)
```

Omitting `TimeZone` uses the host default. A Go `TZ` file path such as
`/usr/share/zoneinfo/America/New_York` may lack a usable Intl time-zone
identifier, so `New` can return `gointl.ErrUnsupportedOption` even when Go can
load that file. Set `TimeZone: gointl.String("America/New_York")`, as above,
to select a named zone explicitly.

### Choose Hour Cycles and Fractional Seconds

Set `HourCycle` when the clock's numeric range matters. `h11` uses 0–11,
`h12` uses 1–12, `h23` uses 0–23, and `h24` uses 1–24. These choices apply to
components, time styles, and ranges. `Hour12`, when supplied, takes precedence
and chooses the locale's preferred cycle in the requested 12- or 24-hour family.

Use field options to show milliseconds with the locale's decimal separator:

```go
format, err := datetimeformat.New(mustLocaleList("fr-FR"), datetimeformat.Options{
	Hour:                   gointl.String(datetimeformat.TwoDigitFieldStyle),
	Minute:                 gointl.String(datetimeformat.TwoDigitFieldStyle),
	Second:                 gointl.String(datetimeformat.TwoDigitFieldStyle),
	FractionalSecondDigits: gointl.Int(3),
	HourCycle:              gointl.String(datetimeformat.H24HourCycle),
	TimeZone:               gointl.String("UTC"),
})
if err != nil {
	return err
}

instant := time.Date(2026, time.May, 8, 0, 0, 7, 987_000_000, time.UTC)
text, err := format.Format(instant)
if err != nil {
	return err
}
fmt.Println(text)
```

Output:

```text
24:00:07,987
```

`FractionalSecondDigits` accepts 1, 2, or 3. Fractional-second parts contain
digits; their decimal separator is a literal part.

### Select Plural Categories

Use `pluralrules` to select CLDR plural categories for message selection:

```go
rules, err := pluralrules.New(mustLocaleList("en"), pluralrules.Options{
	Type: gointl.String(pluralrules.Ordinal),
})
if err != nil {
	return err
}

for _, n := range []int64{1, 2, 3, 4} {
	category := rules.Select(pluralrules.Int(n))
	fmt.Println(category)
}
```

`New` resolves and validates the locale's generated rule family once, so
`Select` is total and returns a category directly. `SelectRange` still returns
an error for invalid runtime endpoints such as `NaN`.

`SelectRange` follows the same digit-option formatting path as `Select`: if two
decimal endpoints format to different strings, they fall through to CLDR plural
range data even when their mathematical rounded values compare equal. Cardinal
ranges use the locale's explicit category-pair result when present and return
`other` when that pair is not defined.

Compact plural selection follows native Intl behavior. Public `PluralRules`
compact notation selects from the source decimal string plus the selected
compact exponent. `NumberFormat` separately chooses the compact suffix from
the visible compact mantissa, while localized currency names and unit phrases
use the rounded value at its original magnitude.

Output:

```text
one
two
few
other
```

### Format Lists and Relative Time

Use `listformat` and `relativetimeformat` for native list and relative-time phrasing:

```go
locales := mustLocaleList("en")

list, err := listformat.New(locales, listformat.Options{
	Type:  gointl.String(listformat.Conjunction),
	Style: gointl.String(listformat.ShortStyle),
})
if err != nil {
	return err
}

relative, err := relativetimeformat.New(locales, relativetimeformat.Options{
	Numeric: gointl.String(relativetimeformat.NumericAuto),
})
if err != nil {
	return err
}

out, err := relative.Format(relativetimeformat.Int(-1), relativetimeformat.Day)
if err != nil {
	return err
}

fmt.Println(list.Format([]string{"red", "green", "blue"}))
fmt.Println(out)
```

ListFormat selects contextual conjunctions from the original element text:
Spanish changes `y` to `e` or `o` to `u` for the matching prefixes, and Hebrew
adds a dash before non-Hebrew text. The same literals appear in `FormatToParts`.

```go
spanish, err := listformat.New(mustLocaleList("es"), listformat.Options{})
if err != nil {
	return err
}
fmt.Println(spanish.Format([]string{"madre", "hijo"})) // madre e hijo

hebrew, err := listformat.New(mustLocaleList("he"), listformat.Options{})
if err != nil {
	return err
}
fmt.Println(hebrew.Format([]string{"א", "Go"})) // א ו-Go
```

Relative-time values always use ECMAScript Number semantics. `Int` and `Uint`
are convenience conversions through `float64`, so integers beyond `2^53` round
exactly as native `Intl.RelativeTimeFormat`; `Float` preserves negative zero.

### Format Durations

Use `durationformat` for `Intl.DurationFormat` semantics, including digital time formatting and fractional sub-second rollup. Reuse the formatter for repeated duration output; it resolves its embedded number and list formatters at construction time.

```go
format, err := durationformat.New(mustLocaleList("en"), durationformat.Options{
	Style: gointl.String(durationformat.DigitalStyle),
})
if err != nil {
	return err
}

out, err := format.Format(durationformat.Duration{
	Hours:   1,
	Minutes: 2,
	Seconds: 3,
})
if err != nil {
	return err
}

fmt.Println(out)
```

Output:

```text
1:02:03
```

Duration fields are `float64` because they mirror ECMAScript Number, but every
field must be finite and integral. Formatting projects each field to its exact
integer before validation and rollup, so represented values such as `1e20`
nanoseconds are not narrowed to `int64`; fractions, NaN, and infinities return
`gointl.ErrInvalidValue`.

### Name Codes

`SupportedCurrencies()` enumerates canonical codes with a name in at least one
selected CLDR locale; numeric precision exceptions do not define membership.
DisplayNames and currency-name formatting retain all names in the selected
profile, including AED, SAR, BDT, historic DEM and XXX. An unknown well-formed
code remains constructible and falls back according to the formatter option.

Use `displaynames` for localized names of language, region, script, currency,
calendar, and date-time field codes:

```go
locales := mustLocaleList("en")

names, err := displaynames.New(locales, displaynames.Options{
	Type: gointl.String(displaynames.Region),
})
if err != nil {
	return err
}
region, ok, err := names.Of("FR")
if err != nil {
	return err
}
if ok {
	fmt.Println(region)
}
```

Choose `StandardLanguageDisplay` for a language followed by its script, region,
and variant names. The default dialect mode can use names such as
`American English` for `en-US`.

```go
names, err := displaynames.New(mustLocaleList("en"), displaynames.Options{
	Type:            gointl.String(displaynames.Language),
	LanguageDisplay: gointl.String(displaynames.StandardLanguageDisplay),
})
if err != nil {
	return err
}
name, ok, err := names.Of("en-Cyrl-US")
if err != nil {
	return err
}
if ok {
	fmt.Println(name) // English (Cyrillic, United States)
}
```

Names use the resolved locale and its parents. With `NoneFallback`, a missing
component returns `ok == false`; with `CodeFallback`, the canonical code appears
in its place.

## Supported Data

`tools/locale-profile.json` defines the CLDR locale profile used by generated
number, date, plural, list, relative time, duration, display-name, unit,
currency, and time-zone display data. Constructor `SupportedLocalesOf` methods
derive support from the payload family they use, so a locale is advertised only
when the backing data can support it.

The default data profile is a curated product shape, not a hidden compatibility
matrix: 104 locale tags, CLDR 48.1.0, ICU 78, and IANA tzdata 2025b. Active
CLDR-backed constructors share that generated profile while deriving their
supported locales from the payloads they actually consume.

To broaden generated CLDR coverage, add tags to `tools/locale-profile.json` and
run `task data`. Any profile expansion must also include behavior evidence
(`task data:contract` and `task conformance:verify`), binary-size evidence
(`task build:size`), and cold-build evidence (`task build:size:cold`). Selectable
build profiles are deliberately absent until repeated host demand proves that a
single curated default cannot serve the product.

To check runtime support, call the constructor package's `SupportedLocalesOf`.

Locale parsing accepts BCP 47 tags through `golang.org/x/text/language`.
Root supported-value accessors return ECMA-402 values: calendars are `gregory`
and `iso8601`, numbering systems include the simple digit systems from
ECMA-402, units come from the sanctioned unit list, currencies come from CLDR,
and time zones come from the generated IANA/CLDR identifier registry.

DateTimeFormat currently formats Gregorian/ISO calendar data. Well-formed but
unsupported calendar requests participate in locale negotiation and fall back to
the generated calendar data; malformed calendar identifiers return constructor
errors.

DateTimeFormat accepts every named Zone or Link in the pinned IANA registry and
matches identifiers using ECMA-402 ASCII-case-insensitive rules. Links such as
`US/Eastern`, `Atlantic/Jan_Mayen`, and `Pacific/Truk` resolve to stable primary
identifiers. `SupportedTimeZones` returns the complete primary projection, and
`Locale.GetTimeZones` returns the `zone.tab` primary identifiers for the
locale's explicit region. A primary identifier can use the display names of its
CLDR alias without changing `ResolvedOptions().TimeZone`. Go's
`time.LoadLocation` supplies transitions from `ZONEINFO`, host paths, GOROOT,
then embedded `time/tzdata` as fallback. Historical GMT offsets retain seconds
when the transition data contains them.

`DateTimeFormat.FormatRange` and `FormatRangeToParts` preserve caller-provided
endpoint order. A later first argument is valid and remains `startRange`; the
methods do not silently sort the range.

After `datetimeformat.New` succeeds, `Format`, `FormatToParts`, `FormatRange`,
and `FormatRangeToParts` accept typed `time.Time` values and return `(result, error)`.
An instant outside ±8,640,000,000,000,000 epoch milliseconds returns
`ErrInvalidValue` before millisecond truncation. Locale, option, and time-zone
failures remain construction errors.

## Known Divergences

`go-intl` targets observable ECMA-402 behavior. Typed Go bridges and differences
from reference output are documented separately:

- **Typed bridges** turn JavaScript dynamic shapes into idiomatic Go signatures. They are intentional and stable.
- **Pinned data** can produce different preferences and symbols from a host's ICU version, as described below.
- **Conformance divergences** are accepted reference mismatches; each one is enumerated and audited by `task conformance:verify`.

### Typed bridges

| JavaScript shape | Go shape | Why |
|------------------|----------|-----|
| `displayNames.of(code)` returns `string \| undefined` or throws `RangeError` | `DisplayNames.Of(code) (string, bool, error)` | `ok` distinguishes missing data; `error` distinguishes invalid code shape. |
| `locale.getTextInfo().direction` may be unavailable when script metadata is unknown | `locale.TextInfo.Direction *string` | `nil` preserves absence instead of inventing an LTR default. |
| JS `new Intl.X(locales, options?)` | `New(locales, opts Options)` accepting a `locale.List` plus exactly one typed `Options` value | Callers express omitted locales with `nil` / `locale.List{}` and use `Options{}` for the empty or omitted JS options object. |
| `format(value)` accepting `Number \| BigInt \| string` | Opaque `numberformat.Value` constructors plus `Format`, `FormatToParts`, `FormatRange`, and `FormatRangeToParts` | Preserves type safety without a public `any` hot path. |
| Resolved option properties that JS omits when inactive (e.g. PluralRules `compactDisplay`, DateTimeFormat component fields, or DisplayNames `languageDisplay`) | Pointer fields on `ResolvedOptions` that are `nil` when the spec hides them | Distinguishes "not set" from "explicitly zero" without ambiguity. |

### Data-dependent behavior

Pinned CLDR data can differ from a host's ICU data. Two current DateTimeFormat
boundaries matter when comparing output:

- `Hour12` set to `gointl.Bool(true)` for `ja` and `ja-JP` selects the Japanese
  `h11` preference, so midnight uses hour 0. Node 26.10.0 selects `h12` for plain `ja` and `h11`
  for `ja-JP`; set `HourCycle` explicitly to choose a numeric clock range.
- The French data has no `arab` number-symbol row. French fractional seconds with
  `NumberingSystem` set to `gointl.String("arab")` use Arabic digits and the
  French decimal comma; native ICU can use the Arabic decimal separator. Arabic-Egyptian `latn` and
  `arab` requests have separate symbol rows.

See [hour-cycle selection](SPECS/30-datetimeformat.md#22-hourcycle-linkage-13111)
and [fractional-second separators](SPECS/31-datetimeformat-skeleton.md) for the
data policy.

### Conformance divergences

Per-package accepted mismatches against FormatJS/Node reference output live in
`<package>/testdata/divergences.md` only when that package has an active or
resolved divergence history. Do not create or retain empty placeholder
`divergences.md` files. The current active divergence audit is:

- [`datetimeformat/testdata/divergences.md`](datetimeformat/testdata/divergences.md)

Each entry uses the strict `id`, `source`, `owner`, `status`, `reason`, optional
DateTimeFormat `native_witness`, `review_after`, and `removal_path` fields.
Unknown, duplicate, malformed, or incomplete fields fail
`task conformance:verify`; resolved records are validated as history before
being excluded from active fixture matching.

## Error Handling

Constructors and formatter methods with reachable caller-fixable failures
return errors that work with the root sentinels in
`github.com/agentable/go-intl`. Most such failures also carry `*gointl.Error`,
which is useful for config UIs, API errors, and host bindings that need stable
machine-readable context. The human error string uses an
`expected ...; got ...` shape and omits internal ECMA-402 abstract-operation
names.

| Error | Meaning |
|-------|---------|
| `gointl.ErrInvalidOption` | A constructor or `SupportedLocalesOf` received an invalid option. |
| `gointl.ErrUnsupportedOption` | A valid ECMA-402 option is not backed by active implementation behavior. |
| `gointl.ErrInvalidValue` | A runtime formatting value is malformed, non-finite, or otherwise invalid. |
| `gointl.ErrInvalidCode` | `DisplayNames.Of` received an invalid code for its resolved type. |

`gointl.ErrUnsupportedOption` also matches `errors.ErrUnsupported`.

```go
_, err := numberformat.New(mustLocaleList("en-US"), numberformat.Options{
	Style: gointl.String(numberformat.CurrencyStyle),
})
if err != nil {
	if errors.Is(err, gointl.ErrInvalidOption) {
		if detail, ok := errors.AsType[*gointl.Error](err); ok {
			return fmt.Errorf("fix %s %s=%q (expected %s): %w",
				detail.Owner, detail.Name, detail.Value, detail.Expected, err)
		}
		return fmt.Errorf("fix formatter options: %w", err)
	}
	return err
}
```

`Error.Kind` is one of `invalidOption`, `unsupportedOption`, `invalidValue`, or
`invalidCode`. `Error.Expected` is optional;
when it is empty, `Error()` still derives generic expected-value guidance from
the error kind and field name. The wrapped sentinel remains the source of truth
for branching with `errors.Is`.

## Development

Development guidance for coding agents lives in [`CLAUDE.md`](CLAUDE.md), with
[`AGENTS.md`](AGENTS.md) kept as the same entrypoint.

```bash
task deps                 # Download modules and tidy go.mod/go.sum
task fmt                  # Format Go code
task vet                  # Run go vet
task test                 # Run go test -race -p 1 ./...
task modules:verify       # Test, vet, and check tidy diff in all four release modules
task lint                 # Run go mod tidy check and golangci-lint
task codegraph:source        # Build a source-only CodeGraph mirror under .tmp/codegraph-source
task codegraph:source:status # Verify mirror/worktree sync, then show CodeGraph index status
task codegraph:source:sync-check # Verify source-only mirror matches the current worktree
task conformance:verify   # Validate fixtures, skip-list, coverage, Node witness matrix, and divergence audit
task conformance:witness  # Refresh generated Node Intl witness fixtures with the active node binary
task data                 # Regenerate CLDR data into internal/cldr/ (writes back)
task data:check           # Regenerate/compare CLDR data and run gen-cldr test/vet
task data:contract        # Verify generated CLDR data contracts
task build:size           # Report root, formatter, and CLDR binary size deltas
task build:size:cold      # Report the same size table after clearing Go's build cache
task bench:run            # Run one-shot benchmark telemetry
task bench                # Produce a non-blocking benchmark report, optionally with BASELINE=<file>
task vuln                 # Run govulncheck
task verify               # Run deps, fmt, vet, lint, race/module tests, conformance, data contract, and vuln
```

Run a targeted package while developing:

```bash
go test -race ./numberformat/...
go test -race -run TestPluralRules_Cardinal/en ./pluralrules/
(cd tools/gen-cldr && go test ./...)
(cd tools/gen-cldr && go vet ./...)
(cd tools/gen-plural-rules && go test ./...)
(cd tools/gen-plural-rules && go vet ./...)
(cd tools/gen-fixtures-from-formatjs && go test ./...)
(cd tools/gen-fixtures-from-formatjs && go vet ./...)
go test ./tools/conformance ./tools/check-conformance
```

## Documentation

- [SPECS](./SPECS/) record public contracts, formatter behavior, data layout, and conformance rules.
- [CLAUDE.md](./CLAUDE.md) defines development workflow and repository conventions for AI coding agents.

## Contributing

Open an issue before changing public behavior. Keep README changes focused on installation, usage, examples, and development commands; put contracts in `SPECS/` and agent workflow rules in `CLAUDE.md` and the `AGENTS.md` symlink.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
