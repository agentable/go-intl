# SPEC 73 — JSON Record Field Names

> **Status:** Revised (2026-06-15)
> **Priority:** Medium (host-boundary stability for resolved options, parts, and locale info records)
> **Authority:** ECMA-402 resolvedOptions objects, part records, segment records, and locale info records define observable JSON field names. Go struct fields are typed bridges only.

---

## 1. JSON Shape Policy

Go record types that mirror ECMA-402 objects use `encoding/json` tags as the host-boundary contract. Field presence follows ECMA-402, not Go zero values:

1. A property that ECMA-402 always reports has no `omitempty`.
2. A branch-only property whose zero value is meaningful uses `*T` plus `omitempty`; `nil` means the JavaScript property is absent.
3. A branch-only string or enum property may use value type plus `omitempty` only when the empty string is not a legal ECMA-402 value.
4. Internal bridge fields use `json:"-"`; they must never appear in JSON output.
5. `locale.Locale` marshals as the canonical BCP 47 string through `encoding.TextMarshaler`.

`record_json_test.go` is the project-wide guard for this policy. Formatter-specific conformance tests remain responsible for comparing values against Generated reference / native fixtures.

The guard must exercise host-boundary records through real constructors or public accessors wherever possible, not only by marshaling hand-built structs. It must cover every record family branch whose JSON presence can change without a Go type change:

- NumberFormat and PluralRules fraction-digit, significant-digit, and precision branches.
- DateTimeFormat style shortcuts versus granular date/time fields.
- DurationFormat duration records and `fractionalDigits` option presence.
- DisplayNames language-only `languageDisplay`.
- Collator `collation` default and backend-specialized values.
- Segmenter word records, including `isWordLike` and omitted `ByteIndex`.
- Locale info records, including week information and text direction.

Adding a public record field or changing `omitempty` behavior without updating this guard is a SPEC violation.

Part records keep ECMA-402 method vocabulary distinct from option vocabulary. In particular, DateTimeFormat's `fractionalSecondDigits` option and resolved property produce `Part` / `RangePart` records with JSON `"type":"fractionalSecond"`; `"fractionalSecondDigits"` is never a part type.

---

## 2. ResolvedOptions Field Names

### Intl.NumberFormat

| Go field | ECMA-402 / JSON field | Presence |
|----------|------------------------|----------|
| `Locale` | `locale` | Always |
| `NumberingSystem` | `numberingSystem` | Always |
| `Style` | `style` | Always |
| `Currency` | `currency` | Currency style only |
| `CurrencyDisplay` | `currencyDisplay` | Currency style only |
| `CurrencySign` | `currencySign` | Currency style only |
| `Unit` | `unit` | Unit style only |
| `UnitDisplay` | `unitDisplay` | Unit style only |
| `MinimumIntegerDigits` | `minimumIntegerDigits` | Always |
| `MinimumFractionDigits` | `minimumFractionDigits` | Fraction or precision branch |
| `MaximumFractionDigits` | `maximumFractionDigits` | Fraction or precision branch |
| `MinimumSignificantDigits` | `minimumSignificantDigits` | Significant or precision branch |
| `MaximumSignificantDigits` | `maximumSignificantDigits` | Significant or precision branch |
| `UseGrouping` | `useGrouping` | Always |
| `Notation` | `notation` | Always |
| `CompactDisplay` | `compactDisplay` | Compact notation only |
| `SignDisplay` | `signDisplay` | Always |
| `RoundingIncrement` | `roundingIncrement` | Always |
| `RoundingMode` | `roundingMode` | Always |
| `RoundingPriority` | `roundingPriority` | Always |
| `TrailingZeroDisplay` | `trailingZeroDisplay` | Always |

NumberFormat branch-only resolved options (`currency`, `currencyDisplay`, `currencySign`, `unit`, `unitDisplay`, and `compactDisplay`) use pointer fields plus `omitempty`; `nil` is the Go bridge for an absent ECMA-402 property.

### Intl.DateTimeFormat

| Go field | ECMA-402 / JSON field | Presence |
|----------|------------------------|----------|
| `Locale` | `locale` | Always |
| `Calendar` | `calendar` | Always |
| `NumberingSystem` | `numberingSystem` | Always |
| `TimeZone` | `timeZone` | Always |
| `HourCycle` | `hourCycle` | When an hour field is present |
| `Hour12` | `hour12` | When an hour field is present |
| `Weekday` | `weekday` | When requested/resolved |
| `Era` | `era` | When requested/resolved |
| `Year` | `year` | When requested/resolved |
| `Month` | `month` | When requested/resolved |
| `Day` | `day` | When requested/resolved |
| `DayPeriod` | `dayPeriod` | When requested/resolved |
| `Hour` | `hour` | When requested/resolved |
| `Minute` | `minute` | When requested/resolved |
| `Second` | `second` | When requested/resolved |
| `FractionalSecondDigits` | `fractionalSecondDigits` | When requested/resolved |
| `TimeZoneName` | `timeZoneName` | When requested/resolved |
| `DateStyle` | `dateStyle` | When style shortcut is used |
| `TimeStyle` | `timeStyle` | When style shortcut is used |

DateTimeFormat branch-only resolved options (`hourCycle`, `hour12`, component fields, `fractionalSecondDigits`, `timeZoneName`, `dateStyle`, and `timeStyle`) use pointer fields plus `omitempty`; `nil` is the Go bridge for an absent ECMA-402 property. `hour12` is present whenever `hourCycle` is present and is derived from the resolved hour cycle rather than from user option presence.

### Other Constructors

| Go type | Go field | ECMA-402 / JSON field | Presence |
|---------|----------|------------------------|----------|
| `pluralrules.ResolvedOptions` | `Locale` | `locale` | Always |
| `pluralrules.ResolvedOptions` | `Type` | `type` | Always |
| `pluralrules.ResolvedOptions` | `MinimumIntegerDigits` | `minimumIntegerDigits` | Always |
| `pluralrules.ResolvedOptions` | `MinimumFractionDigits` | `minimumFractionDigits` | Fraction / precision branch |
| `pluralrules.ResolvedOptions` | `MaximumFractionDigits` | `maximumFractionDigits` | Fraction / precision branch |
| `pluralrules.ResolvedOptions` | `MinimumSignificantDigits` | `minimumSignificantDigits` | Significant / precision branch |
| `pluralrules.ResolvedOptions` | `MaximumSignificantDigits` | `maximumSignificantDigits` | Significant / precision branch |
| `pluralrules.ResolvedOptions` | `PluralCategories` | `pluralCategories` | Always |
| `pluralrules.ResolvedOptions` | `Notation` | `notation` | Always |
| `pluralrules.ResolvedOptions` | `CompactDisplay` | `compactDisplay` | Compact notation only |
| `pluralrules.ResolvedOptions` | `RoundingIncrement` | `roundingIncrement` | Always |
| `pluralrules.ResolvedOptions` | `RoundingMode` | `roundingMode` | Always |
| `pluralrules.ResolvedOptions` | `RoundingPriority` | `roundingPriority` | Always |
| `pluralrules.ResolvedOptions` | `TrailingZeroDisplay` | `trailingZeroDisplay` | Always |
| `listformat.ResolvedOptions` | `Locale` | `locale` | Always |
| `listformat.ResolvedOptions` | `Type` | `type` | Always |
| `listformat.ResolvedOptions` | `Style` | `style` | Always |
| `relativetimeformat.ResolvedOptions` | `Locale` | `locale` | Always |
| `relativetimeformat.ResolvedOptions` | `Style` | `style` | Always |
| `relativetimeformat.ResolvedOptions` | `Numeric` | `numeric` | Always |
| `relativetimeformat.ResolvedOptions` | `NumberingSystem` | `numberingSystem` | Always |
| `durationformat.ResolvedOptions` | `Locale` | `locale` | Always |
| `durationformat.ResolvedOptions` | `NumberingSystem` | `numberingSystem` | Always |
| `durationformat.ResolvedOptions` | `Style` | `style` | Always |
| `durationformat.ResolvedOptions` | `<unit>` / `<unit>Display` | camelCase unit fields | Always |
| `durationformat.ResolvedOptions` | `FractionalDigits` | `fractionalDigits` | Option set only |
| `displaynames.ResolvedOptions` | `Locale` | `locale` | Always |
| `displaynames.ResolvedOptions` | `Style` | `style` | Always |
| `displaynames.ResolvedOptions` | `Type` | `type` | Always |
| `displaynames.ResolvedOptions` | `Fallback` | `fallback` | Always |
| `displaynames.ResolvedOptions` | `LanguageDisplay` | `languageDisplay` | Language type only |
| `collator.ResolvedOptions` | `Locale` | `locale` | Always |
| `collator.ResolvedOptions` | `Usage` | `usage` | Always |
| `collator.ResolvedOptions` | `Sensitivity` | `sensitivity` | Always |
| `collator.ResolvedOptions` | `CaseFirst` | `caseFirst` | Always |
| `collator.ResolvedOptions` | `Collation` | `collation` | Always |
| `collator.ResolvedOptions` | `Numeric` | `numeric` | Always |
| `collator.ResolvedOptions` | `IgnorePunctuation` | `ignorePunctuation` | Always |
| `segmenter.ResolvedOptions` | `Locale` | `locale` | Always |
| `segmenter.ResolvedOptions` | `Granularity` | `granularity` | Always |

---

## 3. Part and Locale Info Records

| Go type | Go field | ECMA-402 / JSON field | Presence |
|---------|----------|------------------------|----------|
| Formatter `Part` | `Type` | `type` | Always |
| Formatter `Part` | `Value` | `value` | Always |
| Range `Part` | `Source` | `source` | Range parts only |
| `relativetimeformat.Part` | `Unit` | `unit` | Non-literal numeric parts |
| `durationformat.Part` | `Unit` | `unit` | Non-literal unit parts |
| `segmenter.Segment` | `Segment` | `segment` | Always |
| `segmenter.Segment` | `CodeUnitIndex` | `index` | Always |
| `segmenter.Segment` | `ByteIndex` | omitted | Internal Go bridge only |
| `segmenter.Segment` | `Input` | `input` | Always |
| `segmenter.Segment` | `IsWordLike` | `isWordLike` | Word granularity records only |
| `locale.WeekInfo` | `FirstDay` | `firstDay` | Always |
| `locale.WeekInfo` | `Weekend` | `weekend` | Always |
| `locale.TextInfo` | `Direction` | `direction` | Always |
