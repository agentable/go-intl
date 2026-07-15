# 46 — Segmenter

Status: active
Owns: `segmenter` package public API, mapping of granularity to UAX #29 segmentation via `github.com/rivo/uniseg`, and explicit UTF-16 code-unit / UTF-8 byte-offset bridge.

References:
- ECMA-402 Segmenter constructor, segment records, iterator, and `containing` behavior.
- Native-engine witness fixtures define locale-tailoring expectations the active backend cannot yet satisfy.
- Underlying engine: `github.com/rivo/uniseg` v0.4+ (grapheme, word, sentence)

---

## 0. Backend Dependency Boundary

Current public MIT releases keep the Segmenter backend license-compatible with
go-intl distribution. go-segment may become the segmentation substrate only
after one of these is true:

- go-segment is available under a license compatible with go-intl's public MIT
  distribution;
- the integration is explicitly internal or optional and does not become a
  mandatory public `go.mod` dependency;
- the shared boundary substrate moves to a separately licensed module.

go-intl owns ECMA-402 locale resolution, option validation, `intlerr` error
shape, JSON segment record policy, conformance fixtures, and
`SupportedLocalesOf`. go-segment must not import go-intl. go-intl must not add a
mandatory `github.com/agentable/go-segment` dependency while go-segment remains
commercial-only.

Backend parity work may use go-segment's `boundary` package as evidence or in
internal builds, but the public Segmenter API must not lose `All`,
`Containing`, `ContainingByte`, UTF-16 code-unit indexes, byte indexes, invalid
UTF-8 behavior, empty-input behavior, word-like presence semantics, or the
tailored-locale support gate.

## 1. Public Surface

```go
package segmenter

type Granularity string
const (
    GraphemeGranularity Granularity = "grapheme"
    WordGranularity     Granularity = "word"
    SentenceGranularity Granularity = "sentence"
)

type Options struct {
    LocaleMatcher *string
    Granularity   *string
}

type Segment struct {
    Segment       string
    CodeUnitIndex int // UTF-16 code-unit offset, matching ECMA-402 index
    ByteIndex     int // UTF-8 byte offset for Go string callers
    Input         string
    IsWordLike    bool // populated only when Granularity == WordGranularity
}

type Segmenter struct{ /* immutable resolved options + locale */ }
type Segments  struct{ /* segmentation view over one input string */ }

func New(locales locale.List, opts Options) (*Segmenter, error)
func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error)
func (s *Segmenter) Segment(input string) *Segments
func (s *Segments)  All() iter.Seq[Segment]
func (s *Segments)  Containing(index int) (Segment, bool)
func (s *Segments)  ContainingByte(index int) (Segment, bool)
func (s *Segmenter) ResolvedOptions() ResolvedOptions
```

MUST rules:

1. `Granularity` defaults to `GraphemeGranularity` when omitted.
2. `LocaleMatcher` and `Granularity` are presence-aware: `nil` means omitted, while a non-nil pointer is an explicit option value. Explicit empty strings are invalid option values.
3. `Segment.CodeUnitIndex` is the ECMA-402 UTF-16 code-unit offset in the original input string.
4. `All` is a Go 1.23 range-over-func iterator, mirroring the JavaScript iterable contract.
5. `Segment.ByteIndex` is the UTF-8 byte offset for callers that need Go string slicing.
6. `Containing(index)` returns the segment whose half-open UTF-16 code-unit range contains `index`. Out-of-range returns `(Segment{}, false)`.
7. `ContainingByte(index)` returns the segment whose half-open byte range `[ByteIndex, ByteIndex+len(Segment))` contains `index`.
8. `IsWordLike` is meaningful only when `Granularity == WordGranularity`. For other granularities the Go field is `false` and JSON omits `isWordLike`, matching ECMA-402's conditional segment data property.
9. `Segmenter` and `Segments` are immutable after construction; both are safe for concurrent use as long as no goroutine mutates the underlying input string.

---

## 2. Granularity Implementation

| Granularity | Implementation |
|-------------|----------------|
| `grapheme`  | `uniseg.FirstGraphemeClusterInString` loop |
| `word`      | `uniseg.FirstWordInString` loop with `IsWordLike` derived from Unicode Letter/Number categories |
| `sentence`  | `uniseg.FirstSentenceInString` loop |

Locale-specific overrides for Thai, Khmer, Lao, Myanmar (dictionary-based word breaking), CJK word segmentation, and Japanese sentence-boundary nuances are NOT implemented. Those locales must not be advertised by `SupportedLocalesOf` until real tailoring lands; constructor locale negotiation may still fall back to the default locale. In the current surface, examples such as `ja`, `th`, and `zh-Hant` must remain unsupported by the static supported-locales API even though their BCP 47 tags parse successfully.

Current tier: **narrowed implementation gap**.

| Field | Value |
|-------|-------|
| Current behavior | Locale-sensitive dictionary/CJK tailoring is not advertised by `SupportedLocalesOf`. The advertised set is an explicit 11-locale snapshot (`ar de en en-GB en-US es fr hi it pt ru`) whose UAX #29 word/sentence boundaries are backed by native-engine witness fixtures. |
| Rationale | Advertising a locale as supported is a **capability + witness** claim: the active `uniseg` backend produces trustworthy word and sentence boundaries for that locale and a native-engine fixture proves it, not merely that the tag parses as BCP 47. The rationale is capability, not "the locales we happen to conformance-test": witnessed non-dictionary locales that are not yet in the snapshot (see future work) are withheld only until their fixtures land. |
| Guardrail | `internal/segmentation.SupportedLocales()` is an explicit allowlist that returns a snapshot. It must not be generated from CLDR locale-profile data, must not be auto-derived as "all CLDR minus dictionary/CJK" (V8 and FormatJS disagree on that boundary), and must not be exposed as mutable package storage. |
| review_after | 2026-09-30 or the next segmentation backend evaluation, whichever comes first. |
| Removal path | Add or select a segmentation backend with dictionary/CJK tailoring, generate native engine fixtures for affected locales, then expand `internal/segmentation.SupportedLocales()`. |

**Future work (D2) — widen the non-dictionary snapshot.** `uniseg` segments the non-dictionary locales `nl`, `pl`, and `tr` byte-for-byte as native does, but they are withheld from the snapshot until word and sentence native-engine witness fixtures are generated for them. Widening is fixture-gated expansion of the allowlist, not auto-derivation. The dictionary/CJK guardrail below is separate and stays: `ja`, `th`, and `zh-Hant` must not be advertised until tailored segmentation lands.

**Divergence — `ResolvedOptions().Locale` for withheld locales.** When a constructor request resolves to a withheld locale, `ResolvedOptions().Locale` falls back to `en` rather than echoing the requested tag. Native V8 reflects the requested locale here even when its segmentation is untailored, so this is an accepted, currently unrecorded V8 divergence; it resolves as the snapshot widens and tailored locales are added.

This gap is not an accepted divergence from ECMA-402. It is an honest supported-locale boundary until go-intl can implement locale-tailored segmentation.
The dependency evidence lives in `reports/github.com-rivo-uniseg.md`. native-engine
tailored-locale fixtures under `segmenter/testdata/conformance/node-v26/` must
remain XFAIL until the backend can match those boundaries and the supported
locale allowlist expands in the same change.

The withheld-locale set is part of the product contract. Manual supported-locale fixtures must keep dictionary and CJK locales out of `SupportedLocalesOf`, while native-engine tailored-locale fixtures keep the target behavior visible for the eventual backend upgrade.

---

## 3. Resolved Options

```go
type ResolvedOptions struct {
    Locale      locale.Locale
    Granularity Granularity
}
```

`Locale` is the resolved data locale after locale matching, not the request locale verbatim. ResolvedOptions and segment-record JSON field names follow [SPEC 73 §JSON Shape Policy](./73-json-records.md#1-json-shape-policy), [SPEC 73 §Other Constructors](./73-json-records.md#other-constructors), and [SPEC 73 §Part and Locale Info Records](./73-json-records.md#3-part-and-locale-info-records).

---

## 4. Errors

- `gointl.ErrInvalidOption`: invalid `LocaleMatcher` or `Granularity`.

Constructor and `SupportedLocalesOf` failures expose `*gointl.Error` and follow SPEC 12's `expected ...; got ...` text rule. `Segment`, `All`, and `Containing` do not return errors.

---

## 5. Static Supported Locales

```go
func SupportedLocalesOf(locales locale.List, opts Options) (locale.List, error)
```

MUST rules:

1. Use `internal/segmentation.SupportedLocales()` as the supported set. The list contains locales whose active boundaries do not require dictionary or locale-specific tailoring beyond the UAX #29 defaults. The package lives outside `internal/cldr/` because the boundary algorithm comes from `github.com/rivo/uniseg`, not CLDR, and it must not inherit the CLDR locale profile automatically.
2. `internal/segmentation.SupportedLocales()` MUST return a fresh slice and be covered by an exact allowlist test. Adding any locale requires word and sentence native-engine fixtures before the public supported-locale API may advertise it.
2a. The manual `SupportedLocalesOf` fixture must continue to request known tailored locales and expect only the actively supported set; deleting that fixture weakens the capability boundary.
3. Validate the `localeMatcher` option through the shared ECMA-402 helper, then require lookup containment in the explicit capability allowlist. Best-fit may choose among genuinely available locale data for ordinary formatters, but it must not turn a cross-language fallback such as `km → en` into a false claim that dictionary segmentation is implemented.
4. Accept one `Options` value; `Options{}` represents omitted static-method options.
5. Read only `LocaleMatcher`; `nil` means omitted and an explicit empty string is invalid.

---

## 6. Typed Bridges and Gap Boundaries

| JavaScript | Go | Class |
|------------|----|-------|
| `Segments` iterable yielding `{ segment, index, input, isWordLike? }` | `Segments.All() iter.Seq[Segment]` | Typed bridge — Go 1.23 range-over-func. |
| `index` measured in UTF-16 code units | `Segment.CodeUnitIndex` | Direct mapping. |
| Go string byte offset | `Segment.ByteIndex` / `ContainingByte(index)` | Typed bridge for Go string slicing. |
| `Segments.containing(index)` returns object or `undefined` | `Containing(index) (Segment, bool)` | Typed bridge. |
| Dictionary/CJK locale tailoring | Not advertised by `SupportedLocalesOf` until implemented | Narrowed implementation gap; see §2. |
