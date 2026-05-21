# SPEC 11 — Locale Matching

> **Status:** Draft (2026-05-08)
> **Priority:** High (locale parsing layer that all formatters must pass through; blocking SPEC 20 / 30 / 40 / 60)
> **Authority:** ECMA-402 `.references/ecma402/spec/negotiation.html` is the normative source. This SPEC documents the current Go contract for the `internal/localematcher` package, `CanonicalizeLocaleList`, `LookupMatchingLocaleByPrefix`, `LookupMatchingLocaleByBestFit`, `ResolveOptions`, `ResolveLocale`, `FilterLocales`, UTS #35 distance table and three-level matching algorithm.

---

## Overview

ECMA-402 locale negotiation is a locale selection algorithm that all formatters (`NumberFormat` / `DateTimeFormat` / `PluralRules`) must execute during the initialization phase. It normalizes the `locales` parameters to the Language Priority List, uses `options.localeMatcher` to select a locale from the set of locales supported by the implementation, and merges the `-u-` extended keys (`ca` / `nu` / `hc` / ...) with options override into the final resolved field.

This SPEC decision:

1. **Do not** reuse `golang.org/x/text/language.Matcher` (output `Confidence` instead of CLDR distance, not comparable to ECMA-402 conformance).
2. In `internal/localematcher/` **self-implemented** ECMA-402 lookup, and transplant the three-layer matching algorithm of FormatJS `intl-localematcher` + UTS #35 distance table in the implementation-defined part of best-fit.
3. `LookupMatchingLocaleByBestFit` is an enhancement of `LookupMatchingLocaleByPrefix`; both share the subtag truncation and canonicalization subroutines.
4. `ResolveOptions` reads `locales` and `options.localeMatcher`, calls `ResolveLocale`, and returns the Go equivalent of options object, resolved locale record, and resolution option record.
5. `FilterLocales` is the only semantic source of constructor `supportedLocalesOf`.

This SPEC **does not** define the `Locale` type itself (SPEC 10), the field mapping of the `-u-` extended key (SPEC 10 §1), the CLDR data format and generator (SPEC 50), and the internal slots of the formatter (SPEC 12).

## 0. ECMA-402 Alignment

Go APIs do not accept arbitrary JavaScript values, but the semantic pipeline must match ECMA-402:

| ECMA-402 operation | Go responsibility |
|--------------------|-------------------|
| `CanonicalizeLocaleList(locales)` | Parse/canonicalize/deduplicate requested `locale.Locale` values while preserving order |
| `ResolveOptions(constructor, localeData, locales, options, ...)` | Resolve locale/options for formatter constructors from typed `Options` |
| `LookupMatchingLocaleByPrefix` | Implement RFC 4647 lookup ignoring Unicode extension sequences |
| `LookupMatchingLocaleByBestFit` | Implementation-defined best-fit, using the selected formatjs/UTS #35 distance algorithm |
| `ResolveLocale` | Merge matched locale data, relevant extension keys, Unicode extension requests, and explicit option overrides |
| `FilterLocales` | Implement `Intl.<Constructor>.supportedLocalesOf` |

Rules:

1. `localeMatcher` accepts exactly `"lookup"` and `"best fit"`, defaulting to `"best fit"`.
2. Unicode extension keys influence resolved locale only when they are in the constructor's relevant extension key list.
3. Explicit options override Unicode extension values when both are supported.
4. Returned supported locale lists preserve requested order.
5. Unrecognized but well-formed requested locales are ignored for `supportedLocalesOf` and fall back through `ResolveLocale` for constructors.
6. `[[AvailableLocales]]` is derived from formatter payload support, then expanded with ECMA-402 fallback tags. A data-backed locale such as `zh-Hant-HK` therefore also makes `zh-HK` available, while `Result.DataLocale` still points at the concrete payload locale.

---

## 1. Package Layout

### 1.1 Decision: Private package `internal/localematcher`

```text
internal/localematcher/
├── match.go ← Public entrance Match / ResolveLocale and Algorithm enumeration
├── lookup.go       ← LookupMatcher(ECMA-402 §9.2.6) + BestAvailableLocale
├── best_fit.go ← BestFitMatcher three-layer algorithm (Tier 1 / Tier 2 / Tier 3)
├── distance.go ← findMatchingDistance + UTS #35 Table query + sync.Map memoize
├── resolve.go ← ResolveLocale(ECMA-402 §9.2.7)+ relevantExtensionKeys processing
├── ucanonicalize.go← UnicodeExtensionValue / InsertUnicodeExtensionAndCanonicalize
├── canonicalize.go ← CanonicalizeLocaleList(ECMA-402 §6.2.1)
└── data.go ← languageMatching / paradigmLocales / matchVariables data accessor wrapper
```

> **Why**:
> 1. **`internal/`** - matcher is the implementation details of formatter entry and is not exposed to users; SPEC 00 §3 has defined `internal/localematcher`.
> 2. **Files are split according to ECMA-402 abstract operation** - one file corresponds to a section of spec (`LookupMatcher.ts` / `BestFitMatcher.ts` / `ResolveLocale.ts`), mirroring FormatJS 1:1 to reduce migration costs.
> 3. **Data accesses `internal/cldr` indirectly through `data.go`** - matcher does not access `import internal/cldr` directly, but through the interface provided by `data.go` (SPEC 50 §6 Data Access API); avoid the circular dependency between SPEC 11 and SPEC 50.
>
> **Rejected**:
> - **Public `localematcher` package** - users never construct matcher directly; `internal/` enforces hidden implementation.
> - **Single file `match.go` fully plugged** - spec The three core operators (Lookup / BestFit / Resolve) each have migration costs and test fixtures; splitting by file facilitates fixture alignment.

### 1.2 Public entrance signature

```go
// internal/localematcher/match.go(signature)
package localematcher

// Algorithm is a LocaleMatcher option of ECMA-402 §9.2.5.
type Algorithm int

const (
AlgorithmLookup Algorithm = iota // sec-LookupMatcher(accurate + subtag truncation)
AlgorithmBestFit // sec-BestFitMatcher (three layers + UTS #35 distance)
)

// Result is an internal product of matcher, not ResolvedLocale; the latter is composed of ResolveLocale.
type Result struct {
Locale string // Hit supported locale string (canonical BCP 47)
DataLocale string // locale used to check CLDR data (remove -u- expansion)
Distance int // 0 = completely equivalent; >= threshold is considered a miss
}

// Match selects the supported locales that best matches the requested list.
// alg = AlgorithmLookup goes to LookupMatcher; alg = AlgorithmBestFit goes to BestFitMatcher.
// Either algorithm does not return an error (Result{Locale: defaultLocale} when there is no match).
func Match(requested, supported []string, defaultLocale string, alg Algorithm) Result

// DefaultMatchingThreshold is the BestFitMatcher Tier 3 distance threshold (FormatJS DEFAULT_MATCHING_THRESHOLD).
// Distance >= this threshold is regarded as "out of the acceptable range" and falls back to defaultLocale.
const DefaultMatchingThreshold = 838
```

Call example:

```go
res := localematcher.Match(
    []string{"zh-TW"},
    []string{"zh-Hans", "zh-Hant", "en"},
    "en", localematcher.AlgorithmBestFit,
)
fmt.Println(res.Locale) // "zh-Hant"(zh-TW maximize to zh-Hant-TW; subtag truncation fallback)
fmt.Println(res.Distance) // 0
```

Formatter callers **MUST** pass in a supported list aligned with the formatter payload:

| Formatter | Supported list source |
|-----------|-----------------------|
| NumberFormat | `internal/cldr/number.SupportedLocales()` |
| DateTimeFormat | `internal/cldr/date.SupportedLocales()` |
| PluralRules | `internal/cldr/plural.SupportedLocales()` |
| Locale-only operations | `internal/cldr.AvailableLocales()` |

`internal/cldrmatch` is the only encapsulation point for this mapping; the public formatter package is prohibited from maintaining its own supported-locale list.

`internal/localematcher.NewMatcher` compiles that raw supported list into the ECMA-402 Available Locales List used by lookup, best-fit, `ResolveLocale`, and `FilterLocales`. The compiled list is duplicate-free, lacks Unicode extension sequences, preserves concrete data locales, and adds only spec-mandated less-narrow fallbacks:

- ordinary truncation fallbacks such as `de` for `de-DE` when the raw payload has no separate `de` record;
- language-region aliases such as `az-AZ` for `az-Latn-AZ` and `zh-HK` for `zh-Hant-HK`.

Derived fallbacks are matching aliases, not new payload records. `Result.Locale` may be `zh-HK`; `Result.DataLocale` remains `zh-Hant-HK` so constructor locale data reads stay honest.

> **Why `Algorithm` is an int enumeration instead of a string**: Go iota is cheaper than string verification; ECMA-402 spec text is `"lookup"` / `"best fit"` (including spaces), `parseAlgorithm(string) (Algorithm, error)` is used at the boundary to convert, and all internals are int.
>
> **Why `DefaultMatchingThreshold = 838` instead of spec verbatim value**:FormatJS `intl-localematcher/abstract/utils.ts` defines `DEFAULT_MATCHING_THRESHOLD = 838`;sec-BestFitMatcher does not specify a specific value. We continue to use FormatJS values to ensure conformance tests are consistent.
>
> **Why formatter-specific supported lists**: CLDR `availableLocales` is the complete data set, not the formatter payload set. Using the actual generated payload to derive the supported list can prevent the matcher from hitting a locale that does not have numbers/date data.

---

## 2. LookupMatcher(ECMA-402 §9.2.6)

### 2.1 Algorithm

ECMA-402 §9.2.6 `LookupMatcher`:

```text
1. For each locale L in the requestedLocales list:
a. noExtensionsLocale := Remove -u- / -t- extensions of L
   b. availableLocale   := BestAvailableLocale(supportedLocales, noExtensionsLocale)
c. If availableLocale is not undefined, return { locale: availableLocale, extension: -u- segment in L }
2. Otherwise return { locale: defaultLocale, extension: undefined }
```

`BestAvailableLocale(availableLocales, locale)`(§9.2.4):

```text
1. candidate := locale
2. Loop:
a. If candidate ∈ availableLocales, return candidate
b. pos := The subscript of the last '-' in candidate; if it does not exist, return undefined
c. If pos >= 2 and candidate[pos-2] = '-', pos -= 2 (remove single-character subtag, such as 'en-x')
   d. candidate := candidate[:pos]
```

### 2.2 Go signature

```go
// internal/localematcher/lookup.go (signature)

// LookupMatcher implements ECMA-402 §9.2.6.
func LookupMatcher(requested, supported []string, defaultLocale string) Result

// BestAvailableLocale implements ECMA-402 §9.2.4.
// Return an empty string to indicate no match.
func BestAvailableLocale(supported []string, locale string) string
```

> **Why subtag truncation is a loop rather than a recursion**: ECMA-402 uses pseudo code to describe loops; Go style also prefers loops; and BCP 47 tag depth is bounded (actual ≤ 5 subtag), and there is no risk of stack overflow.
>
> **Why export `BestAvailableLocale`** separately: `LookupSupportedLocales`(§9.2.8), `ResolveLocale`(§9.2.7), `BestFitMatcher` Tier 2 all require subtag truncation subroutine; DRY.

### 2.3 Boundary situations

| Input | Expected Behavior |
|------|---------|
| `requested = []` | Return `{Locale: defaultLocale, Distance: 0}` |
| `requested = ["xx-INVALID"]`,`supported = ["en"]` | No match after truncation, fallback to `defaultLocale` |
| `requested = ["en-US-u-ca-buddhist"]`,`supported = ["en-US"]` | `noExtensionsLocale = "en-US"` hits; extension `-u-ca-buddhist` is followed by `ResolveLocale` |
| `requested = ["en-x-private"]`,`supported = ["en"]` | subtag truncation skip single character `x` directly remove `x-private`, hit `en` |

---

## 3. BestFitMatcher(ECMA-402 §9.2.5)

### 3.1 ECMA-402 text

ECMA-402 §9.2.5 `BestFitMatcher`:

> The algorithm used to determine the best fit is implementation-defined. Conforming implementations are recommended to satisfy the following criteria: [...]

That is, spec **does not force** a specific algorithm. This gives three options for this SPEC decision and rejection list:

| Candidate | Decision | Reason |
|------|------|------|
| Transplanted FormatJS three-tier algorithm (Tier 1 accurate / Tier 2 maximize+truncation / Tier 3 UTS #35) | ✅ Selected | There is a public test baseline; aligned with the conformance goal |
| Reuse `golang.org/x/text/language.Matcher` | ❌ Reject | Output `Confidence`(No/Low/High/Exact) is not CLDR distance; tie-breaking is determined by `x/text`, inconsistent with FormatJS; `x/text`'s CLDR data version is not synchronized with `internal/cldr` |
| ICU-only simplified heuristic | ❌ Reject | No fixture exposed; cannot reverse verify from conformance test |

> **Why FormatJS algorithm**:
> 1. **conformance is a hard constraint** - SPEC 70 requires byte-equality to pass FormatJS `intl-localematcher/tests/conformance.test.ts`; only the same algorithm can be passed.
> 2. **CLDR distance is a meaningful scalar** - `Confidence` is level 4 discrete and cannot distinguish fine-grained differences such as "es-MX vs es-419"; CLDR distance (0-840+) can.
> 3. **Can be maintained independently** - The three layers of `internal/localematcher` are ~500 LOC, and the fixtures are all in FormatJS; when upgrading CLDR, you only need to rerun the fixture to find the regression.
>
> **Rejected `language.Matcher`**:
> - ❌ `language.Matcher`'s tie-breaking uses `Confidence` + `x/text` internal implementation details, **invisible** to the caller, unable to byte-match FormatJS.
> - ❌ The data version of `x/text` does not form the same conformance baseline as the ECMA-402/formatjs pinned version (SPEC 50 §1 locking CLDR 48.1.0).
> - ❌ Introduce a second CLDR data source to destroy data record consistency.
>
> **Fallout**: consumer-driven expansion can expose the `localematcher.WithLanguageMatcher()` option to allow users to switch to `x/text` (applicable to the "I only want the behavior of `x/text`" scenario); active scope is not implemented.

### 3.2 Three-layer algorithm

`findBestMatch(requested, supported, threshold)`(FormatJS `intl-localematcher/abstract/utils.ts` §findBestMatch):

```text
let lowestDistance = +Inf
let result = { matchedDesired: "", matchedSupported: "", distances: {} }

# === TIER 1 — Exact match fast path ===
for i, desired in requested:
    if desired ∈ supportedSet:
        distance = 0 + i*40
        result.distances[desired] = { desired: distance }
        if distance < lowestDistance:
            lowestDistance = distance
            result.matchedDesired = desired
            result.matchedSupported = desired
        if i == 0:
return result # The first requested hit, return immediately

# === TIER 2 — maximize + suffix truncation ===
for i, desired in requested:
    maximized = Locale(desired).maximize().toString()
    if maximized != desired:
        candidates = getFallbackCandidates(maximized)  # ["zh-Hant-TW","zh-Hant","zh"]
        for j, candidate in candidates:
if candidate == desired: continue # Tier 1 Checked
            if candidate ∈ supportedSet:
# Compare whether the maximize results of candidate and desired are consistent
                candidateMax = Locale(candidate).maximize().toString()
                distance = (candidateMax == maximized) ? (0 + i*40): (j*10 + i*40)
                if distance < lowestDistance:
                    lowestDistance = distance
                    result.matchedDesired   = desired
                    result.matchedSupported = candidate
                break

if result.matchedSupported != "" and lowestDistance == 0:
return result # Tier 2 finds distance=0 and returns directly

# === TIER 3 — Complete UTS #35 Distance Calculation ===
lowestDistance = +Inf # Reset: Tier 2's "position penalty" is not comparable to CLDR distance
for i, desired in requested:
    for k, candidate in supported:
        d = findMatchingDistance(desired, candidate)
        finalDistance = d + i*40
        result.distances[desired][candidate] = finalDistance
        if finalDistance < lowestDistance:
            lowestDistance = finalDistance
            result.matchedDesired   = desired
            result.matchedSupported = candidate

if lowestDistance >= threshold:
    result.matchedDesired   = ""
    result.matchedSupported = ""
return result
```

> **Why Tier 3 reset `lowestDistance`**: The distance of Tier 2 is the "subtag removal position × 10 + request order × 40" heuristic, which is a different scale from the CLDR distance of Tier 3 (based on paradigmLocales / matchVariables table). Mixing will result in "Tier 2 gives es→es-MX = 20, Tier 3 gives es-419 = 39", misjudgment that es is better than es-419 (the latter is closer in the CLDR table).
>
> **Why position penalty `i*40`**:FormatJS verbatim;ECMA-402 does not support `Accept-Language`'s `q=0.1` weighting, but preserves request order as weak priority.

### 3.3 Go signature

```go
// internal/localematcher/best_fit.go (signature)

// BestFitMatcher implements ECMA-402 §9.2.5 (FormatJS three-layer algorithm).
func BestFitMatcher(requested, supported []string, defaultLocale string) Result

// findBestMatch is the core three-layer entrance, wrapped by BestFitMatcher.
type bestMatchResult struct {
    matchedDesired   string
    matchedSupported string
    distances        map[string]map[string]int
}

func findBestMatch(requested, supported []string, threshold int) bestMatchResult

// getFallbackCandidates truncate the maximize results according to the right-to-left subtag to generate candidates.
// Example: "zh-Hant-TW" → ["zh-Hant-TW", "zh-Hant", "zh"]
func getFallbackCandidates(maximized string) []string

// findMatchingDistance goes to the UTS #35 distance table (memoized via sync.Map).
func findMatchingDistance(desired, supported string) int
```

> **Why `findBestMatch` is not exported**: It returns `distances` map for debugging/fixture comparison, but the production code only uses `BestFitMatcher`’s `Result`; not exported to avoid ABI exposure.
>
> **Why `findMatchingDistance` memoize with `sync.Map`**: Repeated calculation of the same pair of (desired, supported) distances in formatRange / large batch request scenarios; `sync.Map` has zero lock overhead for concurrent reading and writing.

### 3.4 UTS #35 Distance Table

`findMatchingDistance` depends on CLDR `languageMatching.json` derived table:

| Data item | Value range | Source |
|-------|-------|------|
| `paradigmLocales` | Collection: `{en, en-GB, es, es-419, pt-BR, pt-PT}` | CLDR `cldr-core/supplemental/languageMatching.json` |
| `matchVariables` | `$enUS` / `$cnsar` / `$americas` / `$maghreb` and other regional macros | Same as above |
| Distance weight | language=80 / script=20-50 / region=4-50 | Same as above (depends on paradigm and variable hits) |

The data is output to `internal/cldr/locale_matching.go` by SPEC 50 §6 codegen, and this SPEC is accessed indirectly through `data.go` accessor.

```go
// internal/localematcher/data.go (signature)

// Indirect accessor - do not directly import internal/cldr to avoid SPEC 11 ↔ SPEC 50 circular references.
type LanguageMatchingData interface {
    ParadigmLocales() map[string]struct{}
    MatchVariables(name string) []string
    DistanceFor(desiredLSR, supportedLSR string) int
}

func data() LanguageMatchingData // Singleton; injected by internal/cldr
```

> **Why interface boundary**: SPEC 50 codegen directly generates `internal/cldr` to implement the concrete type of `LanguageMatchingData`; `internal/localematcher` only relies on the interface. In this way, SPEC 50 data update (CLDR upgrade) does not force SPEC 11 to change the code.

### 3.5 Performance Targets

| Scenario | Target |
|------|------|
| Tier 1 hit | < 200 ns(supportedSet is `map[string]struct{}`) |
| Tier 2 hit | < 5 µs / single match (including maximize + truncation + lookup) |
| Tier 3 complete (supported list = 100) | < 100 µs (steady state after memoize hit) |
| Tier 3 cold(memoize miss) | < 1 ms / single match |

> **Why these numbers**: SPEC 71 §benchmark lists matcher as a fixed overhead in NumberFormat / DateTimeFormat construction; < 10 µs once constructed is a reasonable goal.

---

## 4. ResolveLocale(ECMA-402 §9.2.7)

### 4.1 Algorithm

ECMA-402 §9.2.7 `ResolveLocale(availableLocales, requestedLocales, options, relevantExtensionKeys, localeData)`:

```text
1. matcher := options["localeMatcher"]               # "lookup" | "best fit"
2. r := (matcher == "lookup") ? LookupMatcher(...): BestFitMatcher(...)
3. foundLocale := r.locale
4. result := { dataLocale: foundLocale }
5. supportedExtension := "-u"
6. for each key ∈ relevantExtensionKeys:
   a. foundLocaleData := localeData[foundLocale][key]
b. value := foundLocaleData[0] # locale default
   c. supportedExtensionAddition := ""
d. If r.extension is not undefined:
        i.  requestedValue := UnicodeExtensionValue(r.extension, key)
ii. If requestedValue ∈ keyLocaleData, and (key does not need to be normalized or requestedValue has been normalized),
            value := requestedValue;supportedExtensionAddition := "-" + key + "-" + value
e. If options[key] is not undefined and ∈ keyLocaleData,
override value(options priority > extension)
   f. result[key]:= value
   g. supportedExtension += supportedExtensionAddition
7. If supportedExtension != "-u":
     foundLocale := InsertUnicodeExtensionAndCanonicalize(foundLocale, supportedExtension)
8. result.locale := foundLocale
9. return result
```

### 4.2 Go signature

```go
// internal/localematcher/resolve.go(signature)

// ResolveOptions is the input (options + context) of ResolveLocale.
type ResolveOptions struct {
    Algorithm             Algorithm                  // "lookup" | "best fit"
Requested []string // CanonicalizeLocaleList output
    Supported             []string                   // available locales
    DefaultLocale         string                     // matcher fallback
RelevantExtensionKeys []string // Example: `["nu"]`(NumberFormat)
    Options               map[string]string          // user-provided -u- overrides
LocaleData LocaleDataLookup // Inject from internal/cldr
}

// LocaleDataLookup is a subset of the SPEC 12 §5 / SPEC 50 §6 data provider.
type LocaleDataLookup interface {
// For returns a list of legal values of locale on key. The first item is the default value of locale.
    For(locale, key string) []string
}

// ResolvedLocale is the final product of ResolveLocale; the formatter writes it into the internal slot.
type ResolvedLocale struct {
Locale string // canonical BCP 47 with supportedExtension
DataLocale string // Used to check CLDR data (without -u-)
Extensions map[string]string //The final value of relevantExtensionKeys (option > extension > default)
}

// ResolveLocale implements ECMA-402 §9.2.7.
func ResolveLocale(opts ResolveOptions) ResolvedLocale
```

Call example:

```go
res := localematcher.ResolveLocale(localematcher.ResolveOptions{
    Algorithm:             localematcher.AlgorithmBestFit,
    Requested:             []string{"zh-TW-u-nu-hanidec"},
    Supported:             []string{"zh-Hant", "en"},
    DefaultLocale:         "en",
    RelevantExtensionKeys: []string{"nu"},
    LocaleData:            cldr.NumberFormatLocaleData(),
})
fmt.Println(res.Locale)            // "zh-Hant-u-nu-hanidec"
fmt.Println(res.DataLocale)        // "zh-Hant"
fmt.Println(res.Extensions["nu"])  // "hanidec"
```

> **Why `Extensions` is `map[string]string`**: `relevantExtensionKeys` is dynamic (NumberFormat uses `["nu"]`, DateTimeFormat uses `["ca","nu","hc"]`, PluralRules uses `["nu"]`); hard-coded fields will be repeatedly defined.
>
> **Why `LocaleData` is an interface rather than a specific type**: Same as §3.4 - Decoupling the implementation details of SPEC 11 and SPEC 50. formatter passes in specific types such as `internal/cldr.NumberFormatLocaleData()` when calling `New()`.

### 4.3 InsertUnicodeExtensionAndCanonicalize

`InsertUnicodeExtensionAndCanonicalize(locale, extension)`(FormatJS `intl-localematcher/abstract/InsertUnicodeExtensionAndCanonicalize.ts`):

```go
// internal/localematcher/ucanonicalize.go(signature)

// InsertUnicodeExtensionAndCanonicalize supportsExtension(such as "-u-nu-hanidec")
// Insert locale and execute ECMA-402 §6.2.3 CanonicalizeUnicodeLocaleId.
//
// Key points:
// - If locale already has -u-, merge (option takes precedence; subsequent -u-key- replaces the previous one)
// - If the locale has -t- and other extensions, the position of -u- must be after them (BCP 47 sorting)
// - After canonicalize, -u- internal keys are in lexicographic order (ca < co < fw < hc < kf < kn < nu)
func InsertUnicodeExtensionAndCanonicalize(locale, extension string) string

// UnicodeExtensionValue reads the type value of key from the -u- section (ECMA-402 §6.2.2).
// Example: UnicodeExtensionValue("-u-ca-buddhist-hc-h23", "ca") = "buddhist"
// The default type(`-u-kn` written alone) returns "true".
func UnicodeExtensionValue(extension, key string) string
```

> **Why packaging `language.Tag.SetTypeForKey`**: `x/text`'s `SetTypeForKey` accepts a single key and does not enforce dict-order; both BCP 47 spec and ECMA-402 require keys to be in dictionary order; we do the final sorting in `InsertUnicodeExtensionAndCanonicalize`.
>
> **Rejected**: Directly adjust `language.Tag.SetTypeForKey` for multiple splicing - dict-order is not guaranteed and violates spec.

---

## 5. CanonicalizeLocaleList(ECMA-402 §6.2.1)

### 5.1 Algorithm

ECMA-402 §6.2.1 `CanonicalizeLocaleList(locales)`:

```text
1. If locales = undefined, return []
2. seen := []
3. If typeof locales = "string" or Locale instance, O := [locales]; otherwise O := ToObject(locales)
4. for k from 0 to O.length-1:
   a. tag := O[k]
b. If tag is not string / Locale, RangeError
c. If tag is Locale, canonicalizedTag := tag.toString(); otherwise canonicalizedTag := CanonicalizeUnicodeLocaleId(tag)
d. If canonicalizedTag ∉ seen,seen.push(canonicalizedTag)
5. Return seen
```

### 5.2 Go signature

```go
// internal/localematcher/canonicalize.go(signature)

// CanonicalizeLocaleList implements ECMA-402 §6.2.1.
// Accepts multiple forms of input such as []string / locale.List / locale.Locale / string, etc.
// Return the canonical BCP 47 string slice after deduplication.
func CanonicalizeLocaleList(locales any) ([]string, error)
```

> **Why `any`**: ECMA-402 accepts polymorphic input; Go has no union types, `any` + reflection / type switch is the only solution.
>
> **Why returns error instead of panic**: Invalid BCP 47 tag is user error, and errors matching `gointl.ErrInvalidValue` are uniformly returned at the boundary; CLAUDE.md red line "no panic in production".

### 5.3 Input normalization

| Input type | Processing |
|---------|------|
| `nil` | Returns an empty slice |
| `string` | Single element slice, normalized by `language.Parse` |
| `locale.Locale` | Single element slice, use `loc.String()` |
| `[]string` | Item-by-item `language.Parse` Normalization |
| `locale.List` | Item by item `loc.String()` |
| Others | Return errors matching `gointl.ErrInvalidValue` |

> **Why `[]any`** is not accepted: Go style favors named slice types; mixed type slices are extremely rare in Go. To mix, first turn to `[]string`.

---

## 6. FilterLocales and LookupSupportedLocales

### 6.1 Purpose of FilterLocales

`Intl.NumberFormat.supportedLocalesOf(locales, options)` etc. The semantic source of spec static methods. Go typed bridge receives the value that has been parse into `locale.Locale`, but it still must be deduplicated according to the canonical locale string, and then press `options.localeMatcher` to choose between `lookup` and `best fit`. Returns a subset of the requested locale that can be matched by the supported list, preserving the requested locale itself, relative order, and Unicode extensions.

### 6.2 Go signature

```go
// internal/localematcher/filter.go(signature)

// FilterLocalesWithMaximizer canonical-deduplicates requested locales and implements
// ECMA-402 FilterLocales with CLDR-backed best-fit maximization supplied by callers.
func FilterLocalesWithMaximizer[T localeIdentifier](supported []string, requested []T, matcher Algorithm, maximizer Maximizer) []T
```

```go
// Call example (exposed by formatter package)
out := localematcher.FilterLocalesWithMaximizer(
    cldrnumber.SupportedLocales(),
    locale.List{locale.MustParse("en-US-u-nu-latn"), locale.MustParse("fr-FR")},
    localematcher.AlgorithmBestFit,
    cldrlocale.Maximize,
)
fmt.Println(out) // For example [en-US-u-nu-latn fr-FR]
```

> **Why returns requested locale**:ECMA-402 `FilterLocales` Explicitly append the hit `_locale_` to the result instead of appending the data locale. The constructor `supportedLocalesOf` is a capability detection API. The caller is concerned about which locales it requests can be satisfied by the current implementation.

### 6.3 LookupSupportedLocales(ECMA-402 legacy helper)

Low-level lookup-only helper. Used for testing and algorithm decomposition; constructor `SupportedLocalesOf` must not bypass `FilterLocales` and call it directly.

```go
// internal/localematcher/lookup.go(signature)

// LookupSupportedLocales implements ECMA-402 §9.2.8.
// Go to BestAvailableLocale for each requestedLocale, and keep it if it hits (remove -u- expansion).
func LookupSupportedLocales(supported, requested []string) []string
```

```go
// Call example (exposed by formatter package)
out := localematcher.LookupSupportedLocales(
    cldrnumber.SupportedLocales(),
    []string{"zh-TW", "fr-FR", "xx-INVALID"},
)
fmt.Println(out) // For example ["zh-TW", "fr-FR"]
```

---

## 7. Error handling

### 7.1 Sentinel Error

```go
Locale parsing failures classify through root `gointl.ErrInvalidValue`.
```

> **Why there is no matcher’s own sentinel**: matcher internal `Match` / `ResolveLocale` **does not return an error** (spec requires always falling back to `defaultLocale`); only the locale input parsing phase will fail and be uniformly mapped to the root category.

### 7.2 Reconciling with `locale` package errors

Locale parsing errors classify through `gointl.ErrInvalidValue`; `internal/localematcher` does not define or re-export a public sentinel.

> **Why root category**: When users catch errors via the formatter package, `errors.Is(err, gointl.ErrInvalidValue)` should match all locale parsing failures; multiple independent sentinels will break this equivalence.

---

## 8. Boundary to SPEC 12 / SPEC 50

### 8.1 Relationship with SPEC 12

| SPEC | Provide | Consumption |
|------|------|------|
| SPEC 11 (this) | `ResolveLocale` returns `ResolvedLocale` | calls SPEC 12 §3 `CanonicalizeLocaleList` etc. abstract op? **No** —— `CanonicalizeLocaleList` is owned by this SPEC (ECMA-402 §6.2.1) |
| SPEC 12 | Validators / pattern / decimal boundary for production path reuse | No consumption SPEC 11 |

> **DECISION**: "locale-shape abstract operations" such as `CanonicalizeLocaleList` and `BestAvailableLocale` are currently documented in **SPEC 11**(this). typed formatter constructors own option validation; SPEC 12 no longer carries the JS options-object pipeline.

### 8.2 Relationship with SPEC 50

| SPEC | Provide | Consumption |
|------|------|------|
| SPEC 11 (this) | `LanguageMatchingData` / `LocaleDataLookup` interface | indirect reading through `data.go` accessor |
| SPEC 50 | Concrete type that implements the interface (output to `internal/cldr/locale_matching.go` / `internal/cldr/likely_subtags.go` by codegen) | Not consumed SPEC 11 |

Dependency direction: **SPEC 11 → SPEC 50 interface** + **SPEC 50 implementation injection**. No loops.

---

## 9. Forbidden

### 9.1 ❌ Do not reuse `language.Matcher`

```go
// ❌ Error: Output Confidence is not CLDR distance
matcher := language.NewMatcher([]language.Tag{language.AmericanEnglish, language.Japanese})
tag, _, conf := matcher.Match(language.MustParse("en-GB"))

// ✅ Correct: use internal/localematcher three-layer algorithm
res := localematcher.Match(
    []string{"en-GB"}, []string{"en-US", "ja"}, "en-US",
    localematcher.AlgorithmBestFit,
)
```

> **Why**: `Confidence`(No/Low/High/Exact) Level 4 discrete, unable to byte-match FormatJS `distance` value; tie-breaking is determined internally by `x/text` and is not visible.

### 9.2 ❌ Do not use `import internal/cldr` directly in matcher

```go
// ❌ Error: Circular dependency risk + SPEC 11 and SPEC 50 are tightly coupled
import "github.com/agentable/go-intl/internal/cldr"
func findMatchingDistance(...) int {
    return cldr.LanguageMatching.Distance(...)
}

// ✅ Correct: injected through the LanguageMatchingData interface
func findMatchingDistance(d, s string) int {
    return data().DistanceFor(toLSR(d), toLSR(s))
}
```

> **Why**: CLDR data updates (SPEC 50) should not force matcher code changes; interface isolation allows SPEC 50 to evolve independently.

### 9.3 ❌ Do not return error in `Match` / `ResolveLocale`

```go
// ❌ Error: Violation of ECMA-402 spec("matcher always succeeds")
func Match(...) (Result, error)

// ✅ Correct: Return defaultLocale when there is no match
func Match(requested, supported []string, defaultLocale string, alg Algorithm) Result
```

> **Why**: ECMA-402 §9.2.5 / §9.2.6 both stipulate that "no match returns defaultLocale"; let the caller handle the error, which increases the cognitive load and is inconsistent with the spec. `CanonicalizeLocaleList` is the only possible boundary for errors (user input parsing).

### 9.4 ❌ Don’t panic

```go
// ❌ Error: Violation of CLAUDE.md "no panic in production"
func BestAvailableLocale(supported []string, locale string) string {
    if locale == "" {
        panic("empty locale")
    }
    // ...
}

// ✅ Correct: Returns an empty string (spec semantics "no match")
func BestAvailableLocale(supported []string, locale string) string {
    if locale == "" {
        return ""
    }
    // ...
}
```

### 9.5 ❌ Do not implement maximize repeatedly in BestFitMatcher

```go
// ❌ Error: Handwritten likelySubtags table query in best_fit.go
func tier2Maximize(loc string) string { /* Check the table by yourself */ }

// ✅ Correct: reuse locale.Locale.Maximize()(SPEC 10 §4)
func tier2Maximize(loc string) string {
    l, err := locale.Parse(loc)
    if err != nil { return loc }
    return l.Maximize().String()
}
```

> **Why**: The maximum algorithm is recorded in SPEC 10 / SPEC 50 (`MaximizeSubtags` table); matcher reuse avoids data desynchronization.

### 9.6 ❌ Do not compare Tier 2 and Tier 3 distances together

```go
// ❌ Error: lowestDistance is not reset when Tier 2 skips Tier 3
if tier2Distance < threshold { return tier2Result }
// Tier 3 uses the lowestDistance of tier2 as the initial value - misjudgment

// ✅ Correct: Tier 3 entrance reset lowestDistance = +Inf
lowestDistance = math.MaxInt
for ... { /* Tier 3 */ }
```

> **Why**: Tier 2 distance is heuristic (subtag position × 10), Tier 3 is CLDR measured distance (0-840); the two are not in the same scale. FormatJS `utils.ts` comment explicitly states that.

### 9.7 ❌ Don’t memoize in `findMatchingDistance`

```go
// ❌ Error: Check the CLDR table again for each match
func findMatchingDistance(d, s string) int {
return data().DistanceFor(toLSR(d), toLSR(s)) // Call 100 times = Check 100 times
}

// ✅ Correct: sync.Map memoize
var distanceCache sync.Map  // map[[2]string]int
func findMatchingDistance(d, s string) int {
    key := [2]string{d, s}
    if v, ok := distanceCache.Load(key); ok { return v.(int) }
    dist := data().DistanceFor(toLSR(d), toLSR(s))
    distanceCache.Store(key, dist)
    return dist
}
```

> **Why**: Tier 3 requires 10,000 distance calculations in a 100-locale scenario; memoize reduces the steady state to 0 µs/time.

### 9.8 ❌ Do not compare strings for equivalence `Result`

```go
// ❌ Error: The Result.Distance field also participates in comparison, but the user only cares about Locale
if res1 == res2 { /* ... */ }

// ✅ Correct: Use Result.Locale string comparison
if res1.Locale == res2.Locale { /* ... */ }
```

> **Why**: `Distance` may be different in fixture tests (different CLDR versions); using `Locale` is more stable.

---

## 10. Acceptance Criteria

### Package structure

- [ ] `internal/localematcher/` subdirectory is divided into files `match.go` / `lookup.go` / `best_fit.go` / `distance.go` / `resolve.go` / `ucanonicalize.go` / `canonicalize.go` / `data.go` / `errors.go`.
- [ ] `internal/localematcher` is not directly accessed `import "github.com/agentable/go-intl/internal/cldr"`; indirectly accessed through the `data.go` interface.
- [ ] `internal/localematcher` is not directly `import "github.com/agentable/go-intl/internal/ecma402"` (SPEC 12 is option-shape; this SPEC is locale-shape; there is no loop between the two).

### LookupMatcher

- [ ] `LookupMatcher(requested, supported, defaultLocale) Result` implements ECMA-402 §9.2.6 verbatim.
- [ ] `BestAvailableLocale(supported, locale) string` implements ECMA-402 §9.2.4 (single-character subtag skips position -2).
- [ ] FormatJS `intl-localematcher/tests/LookupMatcher.test.ts` All fixtures pass in `internal/localematcher/lookup_test.go`.

### BestFitMatcher

- [ ] `BestFitMatcher(requested, supported, defaultLocale) Result` implements the three-tier algorithm (Tier 1 accurate / Tier 2 maximize+truncation / Tier 3 UTS #35).
- [ ] `findBestMatch` resets `lowestDistance = +Inf` on Tier 3 entry (not mixed with Tier 2 heuristics).
- [ ] `getFallbackCandidates(maximized)` output `["zh-Hant-TW","zh-Hant","zh"]` (right-to-left subtag truncation).
- [ ] `findMatchingDistance` uses `sync.Map` memoize (the same (desired, supported) pair is only looked up once).
- [ ] `DefaultMatchingThreshold = 838`(FormatJS verbatim).
- [ ] FormatJS `intl-localematcher/tests/BestFitMatcher.test.ts` and `tests/conformance.test.ts` all fixtures pass in `internal/localematcher/best_fit_test.go`.

### ResolveLocale

- [ ] `ResolveLocale(opts) ResolvedLocale` implements ECMA-402 §9.2.7 (option > extension > localeData default priority).
- [ ] `relevantExtensionKeys` is a dynamic parameter (NumberFormat uses `["nu"]`, DateTimeFormat uses `["ca","nu","hc"]`).
- [ ] `InsertUnicodeExtensionAndCanonicalize` outputs `-u-` keys in lexicographic order (ca < co < fw < hc < kf < kn < nu).
- [ ] `UnicodeExtensionValue` returns `"true"` in the default type (`-u-kn` writes the scene separately).
- [ ] FormatJS `intl-localematcher/tests/ResolveLocale.test.ts` All fixtures pass.

### CanonicalizeLocaleList

- [ ] `locale.CanonicalizeList(locale.List) locale.List` accepts `nil` / empty `locale.List` / ordered locale request list.
- [ ] raw string locale only occurs at `locale.Parse` boundaries; locale-list canonicalization does not create impossible errors.
- [ ] Keep the order of first appearance after deduplication.

### LookupSupportedLocales

- [ ] `FilterLocales(supported, requested, matcher) locale.List` implements ECMA-402 `FilterLocales`.
- [ ] `FilterLocales` canonical-deduplicates requested locale values before filtering.
- [ ] `FilterLocales` preserves requested locale order and Unicode extensions for matched locales.
- [ ] Constructor package `SupportedLocalesOf` methods call `internal/ecma402.SupportedLocalesOf` rather than duplicating matcher loops.
- [ ] `LookupSupportedLocales(supported, requested) []string` implements ECMA-402 §9.2.8.
- [ ] Output stripped of `-u-` extension.

### Error

- [ ] `errors.Is(err, gointl.ErrInvalidValue)` returns true if locale list parsing fails.
- [ ] `Match` / `ResolveLocale` does not return error (always falls back to `defaultLocale`).
- [ ] There is **no** `panic` call in the package (the test covers various abnormal inputs).

### Performance

- [ ] Tier 1 hit path benchmark < 200 ns(`go test -bench=BenchmarkTier1Match`).
- [ ] Tier 3 steady state (memoize hit) 100-locale supported list benchmark < 100 µs.
- [ ] `sync.Map` memoize has no race (`-race` passes) under concurrent 10 goroutines.

### Boundary with `language.Matcher`

- [ ] `internal/localematcher` package does not import `golang.org/x/text/language.Matcher` / `MatchStrings` (`grep -r "language.Matcher" internal/localematcher/` should be empty).
- [ ] `Locale.Maximize` / `Minimize` (SPEC 10) are reused within Tier 2 (do not reimplement the likelySubtags query).

### Test

- [ ] FormatJS `intl-localematcher/tests/locale-match-fixtures.json` is ported to `internal/localematcher/testdata/match-fixtures.json` and consumed in `match_test.go` table driver.
- [ ] Use `t.Parallel()` for all tests.
- [ ] At least 1 `Example*` function demonstrating `Match` + `ResolveLocale` concatenation.

---

## References

### Specification

- [ECMA-402 §6.2 — Language Tags](https://tc39.es/ecma402/#sec-language-tags)
- [ECMA-402 §9.2 — Locale Resolution](https://tc39.es/ecma402/#sec-locale-resolution)
- [ECMA-402 §9.2.4 — BestAvailableLocale](https://tc39.es/ecma402/#sec-bestavailablelocale)
- [ECMA-402 §9.2.5 — BestFitMatcher](https://tc39.es/ecma402/#sec-bestfitmatcher)
- [ECMA-402 §9.2.6 — LookupMatcher](https://tc39.es/ecma402/#sec-lookupmatcher)
- [ECMA-402 §9.2.7 — ResolveLocale](https://tc39.es/ecma402/#sec-resolvelocale)
- [ECMA-402 §9.2.8 — LookupSupportedLocales](https://tc39.es/ecma402/#sec-lookupsupportedlocales)
- [UTS #35 §EnhancedLanguageMatching](https://unicode.org/reports/tr35/#EnhancedLanguageMatching)

### Reference implementations

- `.references/formatjs/packages/intl-localematcher/abstract/utils.ts` —— Three-layer `findBestMatch` implementation + `findMatchingDistance` + `DEFAULT_MATCHING_THRESHOLD = 838` (key file)
- `.references/formatjs/packages/intl-localematcher/abstract/BestFitMatcher.ts` —— `BestFitMatcher` packaging
- `.references/formatjs/packages/intl-localematcher/abstract/LookupMatcher.ts` - ECMA-402 §9.2.6 implementation
- `.references/formatjs/packages/intl-localematcher/abstract/BestAvailableLocale.ts` —— subtag truncation
- `.references/formatjs/packages/intl-localematcher/abstract/ResolveLocale.ts` —— Complete resolution + relevantExtensionKeys processing
- `.references/formatjs/packages/intl-localematcher/abstract/InsertUnicodeExtensionAndCanonicalize.ts`
- `.references/formatjs/packages/intl-localematcher/abstract/UnicodeExtensionValue.ts`
- `.references/formatjs/packages/intl-localematcher/abstract/CanonicalizeLocaleList.ts`
- `.references/formatjs/packages/intl-localematcher/abstract/LookupSupportedLocales.ts`
- `.references/formatjs/packages/intl-localematcher/abstract/languageMatching.ts` —— CLDR matching data table (paradigmLocales: en, en-GB, es, es-419, pt-BR, pt-PT; matchVariables: $enUS, $cnsar, $americas, $maghreb; distance weight)
- `.references/formatjs/packages/intl-localematcher/tests/conformance.test.ts` —— ICU4J alignment fixture
- `.references/formatjs/packages/intl-localematcher/tests/locale-match-fixtures.json` —— table-driven fixture
- `.references/ext/src/ecma402/locale.cpp` —— PHP `ecma402_bestAvailableLocale` via ICU (same as spec but different path)

### Cross-SPEC

- [SPEC 00 §8 Q5 — BestFit Matcher Implementation Selection](./00-vision-and-scope.md#8-open-questions)(This SPEC is closed)
- [SPEC 10 §4 — Maximize & Minimize](./10-locale.md#4-maximize--minimize) — Tier 2 fallback
- [SPEC 10 §1 — Locale structure](./10-locale.md#1-locale-structure) — `Locale.String()` is the matcher input
- [SPEC 12 §3 — Option Validation](./12-abstract-operations.md#3-option-validation) —— matcher does not implement formatter option validation repeatedly
- [SPEC 12 §5 — Internal Slots](./12-abstract-operations.md#5-internal-slots) —— `[[Locale]]` slot holds `ResolvedLocale.Locale`
- [SPEC 50 §6 — Data Access API](./50-cldr-data.md#6-data-access-api) —— `LanguageMatchingData` / `LocaleDataLookup` implements injection
- [SPEC 70 §Conformance](./70-conformance.md) —— matcher fixture is part of conformance test


---

> This SPEC is a maintenance record for `internal/localematcher`. New ECMA-402 matcher subroutine (rare) or CLDR `languageMatching.json` data structure changes trigger this SPEC revision; specific data updates for the UTS #35 distance table are completed by the SPEC 50 codegen.
