# go-intl

[![Go Reference](https://pkg.go.dev/badge/github.com/agentable/go-intl.svg)](https://pkg.go.dev/github.com/agentable/go-intl)
[![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

A Go implementation of the active ECMA-402 `Intl` API with typed constructors and generated CLDR data.

## Features

- **Native Intl alignment**: Public packages map to the active `Intl` constructors: `Locale`, `NumberFormat`, `DateTimeFormat`, `PluralRules`, `ListFormat`, `RelativeTimeFormat`, `DurationFormat`, `DisplayNames`, `Collator`, and `Segmenter`.
- **Typed Go bridge**: Parse locale strings into `locale.List`, then pass `time.Time`, typed option structs, and typed numeric values while preserving ECMA-402 behavior.
- **Root namespace**: The root package represents the JavaScript `Intl` namespace as an aggregate facade; production services that need one formatter should import that constructor package directly.
- **Reusable formatters**: Construct once and reuse; constructors resolve locale, options, and data so repeated formatting stays on the cached path.
- **Host-friendly records**: Resolved options, parts, ranges, locale info, segments, and durations marshal with ECMA-402 JSON field names for API and JS-host boundaries.
- **Structured errors**: Root sentinels work with `errors.Is`, and `gointl.Error` exposes stable kind, owner, option, value, locale, and expected-value guidance.
- **CLDR-backed data**: Ship generated CLDR data as Go source; applications do not load JSON, ICU, or time-zone data files at runtime.
- **Reference fixtures**: Verify formatter output against ECMA-402-derived FormatJS fixtures and native Intl snapshots.

## Installation

```bash
go get github.com/agentable/go-intl
```

Requires **Go 1.26+**.

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
| `github.com/agentable/go-intl` | Root `Intl` namespace helpers, active constructor type aliases, supported values, and structured error categories. |
| `github.com/agentable/go-intl/locale` | BCP 47 parsing, request lists, canonical locale identity, maximize/minimize, and locale info getters. |
| `github.com/agentable/go-intl/numberformat` | Decimal, percent, currency, unit, compact, scientific, engineering, parts, and range formatting. |
| `github.com/agentable/go-intl/datetimeformat` | Date/time styles, field-based formatting, time zones, parts, and date/time ranges. |
| `github.com/agentable/go-intl/pluralrules` | Cardinal and ordinal plural category selection, including range selection. |
| `github.com/agentable/go-intl/listformat` | Locale-sensitive list formatting and list parts. |
| `github.com/agentable/go-intl/relativetimeformat` | Locale-sensitive relative-time formatting and parts. |
| `github.com/agentable/go-intl/durationformat` | Locale-sensitive duration formatting and parts. |
| `github.com/agentable/go-intl/displaynames` | Localized names for languages, regions, scripts, currencies, calendars, and date-time fields. |
| `github.com/agentable/go-intl/collator` | Locale-sensitive string comparison for sorting. |
| `github.com/agentable/go-intl/segmenter` | Unicode grapheme, word, and sentence segmentation for supported locales. |
| `github.com/agentable/go-intl/option` | Zero-dependency `Int`/`Bool`/`String` pointer helpers for optional scalar options; usable without importing the aggregate root. |

Prefer constructor subpackages in services that need one formatter. Importing `github.com/agentable/go-intl` is an aggregate facade: it mirrors the JavaScript `Intl` namespace and therefore imports every active constructor surface for namespace helpers and type aliases. Use the root package for `GetCanonicalLocales`, root supported-value accessors, or when you intentionally want the full `Intl` namespace shape.

See the [Go package documentation](https://pkg.go.dev/github.com/agentable/go-intl) for the full API reference.

## Native Intl Mapping

The Go packages follow the ownership of the native JavaScript `Intl` API:

| JavaScript | Go |
|------------|----|
| `Intl.getCanonicalLocales(locales)` | `gointl.GetCanonicalLocales(locales)` |
| `Intl.supportedValuesOf("calendar")` | `gointl.SupportedCalendars()` |
| `Intl.supportedValuesOf("collation")` | `gointl.SupportedCollations()` |
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
| `new Intl.Collator(locales, options)` | `collator.New(locales, options)` |
| `Intl.Collator.supportedLocalesOf(locales, options)` | `collator.SupportedLocalesOf(locales, options)` |
| `new Intl.Segmenter(locales, options)` | `segmenter.New(locales, options)` |
| `Intl.Segmenter.supportedLocalesOf(locales, options)` | `segmenter.SupportedLocalesOf(locales, options)` |

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
```

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

category, err := rules.Select(pluralrules.Int(1))
if err != nil {
	return err
}

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

Unit identifiers follow native `Intl.NumberFormat`: use canonical lowercase ECMA-402 identifiers such as `meter`, `microsecond`, or `kilometer-per-hour`.

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

Root supported-value accessors cover calendars, collations, currencies, numbering systems, time zones, and units. Collation currently reports only identifiers the active backend can apply truthfully.

Use `gointl.Int`, `gointl.Bool`, and `gointl.String` for optional option fields where ECMA-402 distinguishes omitted from an explicit zero, false, or empty value.

These helpers also live in the zero-dependency leaf package
`github.com/agentable/go-intl/option` as `option.Int`, `option.Bool`, and
`option.String`. A service that uses a single formatter package can set its
optional scalar options through `option` without importing the aggregate root,
which pulls in every constructor package. The root re-exports the same helpers
as `gointl.Int`/`gointl.Bool`/`gointl.String` for namespace fidelity, so the
root-namespace examples above stay the primary documented style.

### Format Exact Decimals

Use decimal-string methods when binary `float64` cannot represent the value you need to format:

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
`durationformat.Duration`, locale `WeekInfo` / `TextInfo`, and
`segmenter.Segment` also use ECMA-402 field names.

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

fmt.Println(format.Format(start))
rangeText, err := format.FormatRange(start, end)
if err != nil {
	return err
}
fmt.Println(rangeText)
```

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
	category, err := rules.Select(pluralrules.Int(n))
	if err != nil {
		return err
	}
	fmt.Println(category)
}
```

`SelectRange` follows the same digit-option formatting path as `Select`: if two
decimal endpoints format to different strings, they fall through to CLDR plural
range data even when their mathematical rounded values compare equal.

Compact plural selection follows native Intl behavior. Public `PluralRules`
compact notation selects from the source decimal string plus the selected
compact exponent; `NumberFormat` compact formatting separately chooses its
compact suffix from the visible compact display decimal plus that exponent.

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
	Type:  listformat.Conjunction,
	Style: listformat.ShortStyle,
})
if err != nil {
	return err
}

relative, err := relativetimeformat.New(locales, relativetimeformat.Options{
	Numeric: relativetimeformat.NumericAuto,
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

### Name Codes, Sort Text, and Segment Strings

Use the remaining constructor packages when you need display names, collation, or text boundaries:

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

coll, err := collator.New(locales, collator.Options{})
if err != nil {
	return err
}
words := []string{"z", "a", "ä"}
slices.SortFunc(words, coll.Compare)
fmt.Println(words)

seg, err := segmenter.New(locales, segmenter.Options{
	Granularity: segmenter.WordGranularity,
})
if err != nil {
	return err
}
for part := range seg.Segment("Hello, world.").All() {
	if part.IsWordLike {
		fmt.Println(part.Segment)
	}
}
```

## Supported Data

`tools/locale-profile.json` defines the CLDR locale profile used by generated
number, date, plural, list, relative time, duration, display-name, unit,
currency, and time-zone display data. Constructor `SupportedLocalesOf` methods
derive support from the payload family they use, so a locale is advertised only
when the backing data or engine can support it.

The default data profile is a curated product shape, not a hidden compatibility
matrix: 104 locale tags, CLDR 48.1.0, ICU 78, and tzdata 2025b. Most
CLDR-backed constructors share that generated profile. `Collator` and
`Segmenter` are intentionally engine-specific: `Collator` follows
`golang.org/x/text/collate`, and `Segmenter` currently reports only locales
whose UAX #29 boundaries do not require dictionary or CJK tailoring. For example,
`ja`, `th`, and `zh-Hant` are not advertised by `segmenter.SupportedLocalesOf`
until tailored segmentation lands.

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
ECMA-402, units come from the sanctioned unit list, currencies and time zones
come from generated CLDR / tz data, and collations come from the active
collation backend.

DateTimeFormat currently formats Gregorian/ISO calendar data. Well-formed but
unsupported calendar requests participate in locale negotiation and fall back to
the generated calendar data; malformed calendar identifiers return constructor
errors.

DateTimeFormat accepts canonical IANA time zones and generated alias links such
as `America/Montreal`, resolving them to canonical names from the pinned CLDR
alias data plus documented legacy IANA links. `SupportedTimeZones` advertises
canonical names, not aliases.

## Known Divergences

`go-intl` matches observable ECMA-402 output where it can and documents every accepted difference. Two categories exist:

- **Typed bridges** turn JavaScript dynamic shapes into idiomatic Go signatures. They are intentional and stable.
- **Conformance divergences** are accepted reference mismatches; each one is enumerated and audited by `task conformance:verify`.

### Typed bridges

| JavaScript shape | Go shape | Why |
|------------------|----------|-----|
| `collator.compare(x, y)` returns `-1 \| 0 \| 1` | `Collator.Compare(x, y) int` returning standard `-`/`0`/`+` | Matches `slices.SortFunc` and `bytes.Compare`. |
| `displayNames.of(code)` returns `string \| undefined` or throws `RangeError` | `DisplayNames.Of(code) (string, bool, error)` | `ok` distinguishes missing data; `error` distinguishes invalid code shape. |
| `segmenter.segment(s)` returns a JS iterable of `{ segment, index, isWordLike }` | `Segments.All() iter.Seq[Segment]` with `Segment.CodeUnitIndex` and `Segment.ByteIndex` | Mirrors ECMA-402 while still supporting Go string offsets. |
| JS `new Intl.X(locales, options?)` | `New(locales, opts Options)` accepting a `locale.List` plus exactly one typed `Options` value | Callers express omitted locales with `nil` / `locale.List{}` and use `Options{}` for the empty or omitted JS options object. |
| `format(value)` accepting `Number \| BigInt \| string` | Opaque `numberformat.Value` constructors plus `Format`, `FormatToParts`, `FormatRange`, and `FormatRangeToParts` | Preserves type safety without a public `any` hot path. |
| Resolved option properties that JS omits when inactive (e.g. PluralRules `compactDisplay`, DateTimeFormat component fields, or DisplayNames `languageDisplay`) | Pointer fields on `ResolvedOptions` that are `nil` when the spec hides them | Distinguishes "not set" from "explicitly zero" without ambiguity. |

### Conformance divergences

Per-package accepted mismatches against FormatJS/Node reference output live in
`<package>/testdata/divergences.md` only when that package has an active or
resolved divergence history. Do not create or retain empty placeholder
`divergences.md` files. The current active divergence audit is:

- [`datetimeformat/testdata/divergences.md`](datetimeformat/testdata/divergences.md)

Each active entry includes a divergence ID, source fixture, owner, reason, review date, and removal path. `task conformance:verify` enforces that every accepted divergence is still referenced.

## Error Handling

Constructors and formatter methods return errors that work with the root
sentinels in `github.com/agentable/go-intl`. Most caller-fixable failures also
carry `*gointl.Error`, which is useful for config UIs, API errors, and host
bindings that need stable machine-readable context. The human error string uses
an `expected ...; got ...` shape and omits internal ECMA-402 abstract-operation
names.

| Error | Meaning |
|-------|---------|
| `gointl.ErrInvalidOption` | A constructor or `SupportedLocalesOf` received an invalid option. |
| `gointl.ErrUnsupportedOption` | A valid ECMA-402 option is not backed by active implementation behavior. |
| `gointl.ErrInvalidValue` | A runtime formatting value is malformed, non-finite, or otherwise invalid. |
| `gointl.ErrInvalidCode` | `DisplayNames.Of` received an invalid code for its resolved type. |
| `gointl.ErrInvalidKey` | A keyed root namespace operation received an unsupported key. |
| `gointl.ErrUnsupportedLocale` | A locale request is outside the active data set. |
| `gointl.ErrUnsupportedBackend` | Required implementation support is unavailable. |

The three `ErrUnsupported*` categories also match `errors.ErrUnsupported`.

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

`Error.Kind` returns values such as `invalidOption`, `unsupportedOption`,
`invalidValue`, `invalidCode`, and `invalidKey`. `Error.Expected` is optional;
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
task lint                 # Run go mod tidy check and golangci-lint
task codegraph:source        # Build a source-only CodeGraph mirror under .tmp/codegraph-source
task codegraph:source:status # Verify mirror/worktree sync, then show CodeGraph index status
task codegraph:source:sync-check # Verify source-only mirror matches the current worktree
task conformance:verify   # Validate fixtures, skip-list, coverage, Node witness matrix, and divergence audit
task conformance:witness  # Refresh generated Node Intl witness fixtures with the active node binary
task data                 # Regenerate CLDR data into internal/cldr/ (writes back)
task data:check           # Regenerate CLDR data into a temp tree and compare byte-for-byte
task data:contract        # Verify generated CLDR data contracts
task build:size           # Report root, formatter, and CLDR binary size deltas
task build:size:cold      # Report the same size table after clearing Go's build cache
task bench:run            # Run one-shot benchmark telemetry
task bench                # Produce a non-blocking benchmark report, optionally with BASELINE=<file>
task vuln                 # Run govulncheck
task verify               # Run deps, fmt, vet, lint, test, conformance, data contract, and vuln checks
```

The host consumer profile is exercised by `go test ./...`; it protects
supported-set boundaries and reversed range behavior that host integrations
depend on.

Use `task codegraph:source` before structural exploration. The generated mirror
excludes `.references/` and lives under ignored `.tmp/codegraph-source`, so
current-source questions do not accidentally traverse the vendored reference
trees. `task codegraph:source:status` checks that the mirror still matches the
working tree before reporting the CodeGraph index status; use
`task codegraph:source:sync-check` when you only need the mirror freshness
check. Reference projects keep their own CodeGraph indexes when a comparison
needs implementation evidence.

Measure aggregate root facade cost separately from per-surface formatter cost.
Treat `go list -deps .` as root aggregate evidence only; use direct subpackage
commands for single-formatter dependency measurements.

```bash
go list -deps . | wc -l
go list -deps ./numberformat | wc -l
go list -deps ./datetimeformat | wc -l
task build:size  # CLDR linker smoke for generated data changes
task build:size:cold  # CLDR cold-build smoke for profile changes
```

For binary-size checks, label root facade harnesses separately from formatter-only harnesses.

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
