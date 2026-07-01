# SPEC 11 — Locale Matching

> **Status:** Revised (2026-05-31)
> **Priority:** High (locale parsing layer that all formatters must pass through; blocking SPEC 20 / 30 / 40 / 60)
> **Authority:** ECMA-402 `.references/ecma402/spec/negotiation.html` is the normative source. This SPEC documents the current Go contract for `internal/ecma402.CanonicalLocaleList`, the `internal/localematcher` package, `LookupMatchingLocaleByPrefix`, `LookupMatchingLocaleByBestFit`, `ResolveOptions`, `ResolveLocale`, `FilterLocales`, compiled matcher indexes, and the active bounded distance model.

---

## Overview

ECMA-402 locale negotiation is a locale selection algorithm that all formatters (`NumberFormat` / `DateTimeFormat` / `PluralRules`) must execute during the initialization phase. It normalizes the `locales` parameters to the Language Priority List, uses `options.localeMatcher` to select a locale from the set of locales supported by the implementation, and merges the `-u-` extended keys (`ca` / `nu` / `hc` / ...) with options override into the final resolved field.

This SPEC decision:

1. **Do not** reuse `golang.org/x/text/language.Matcher` (output `Confidence` instead of CLDR distance, not comparable to ECMA-402 conformance).
2. In `internal/localematcher/` **self-implemented** ECMA-402 lookup and best-fit locale matching. Best-fit keeps the Generated reference three-tier shape while using the active bounded Go distance model; a full CLDR `languageMatching` codegen table is not part of the current runtime contract.
3. `LookupMatchingLocaleByBestFit` is an enhancement of `LookupMatchingLocaleByPrefix`; both share the subtag truncation and canonicalization subroutines.
4. Formatter constructors use `internal/ecma402.ResolveConstructorLocale` as the narrow Go typed wrapper over requested-locale preparation, `localeMatcher` algorithm selection, default-locale fallback, and `ResolveLocale`.
5. `FilterLocales` is the only semantic source of constructor `supportedLocalesOf`.

This SPEC **does not** define the `Locale` type itself (SPEC 10), the field mapping of the `-u-` extended key (SPEC 10 §1), the CLDR data format and generator (SPEC 50), and the internal slots of the formatter (SPEC 12).

## 0. ECMA-402 Alignment

Go APIs do not accept arbitrary JavaScript values, but the semantic pipeline must match ECMA-402:

| ECMA-402 operation | Go responsibility |
|--------------------|-------------------|
| `CanonicalizeLocaleList(locales)` | Deduplicate already-parsed requested `locale.Locale` values while preserving order through `internal/ecma402.CanonicalLocaleList` |
| `ResolveOptions(constructor, localeData, locales, options, ...)` | Resolve locale/options for formatter constructors from typed `Options` |
| `LookupMatchingLocaleByPrefix` | Implement RFC 4647 lookup ignoring Unicode extension sequences |
| `LookupMatchingLocaleByBestFit` | Implementation-defined best-fit, using the selected Generated reference three-tier shape and active bounded distance model |
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
├── available.go     ← ECMA-402 Available Locales expansion and data-locale mapping
├── compiled.go      ← compiled Matcher indexes for repeated constructor negotiation
├── match.go         ← Match / MatchWithMaximizer and Algorithm enumeration
├── lookup.go        ← LookupMatcher(ECMA-402 §9.2.6) + BestAvailableLocale
├── best_fit.go      ← BestFitMatcher three-layer algorithm (Tier 1 / Tier 2 / Tier 3)
├── distance.go      ← bounded matching distance + sync.Map memoize
├── filter.go        ← FilterLocales / FilterLocalesWithMaximizer
├── resolve.go       ← ResolveLocale(ECMA-402 §9.2.7) + relevantExtensionKeys processing
└── ucanonicalize.go ← UnicodeExtensionValue / InsertUnicodeExtensionAndCanonicalize
```

Unicode extension parsing and insertion are delegated to `internal/localeid`
so `locale.Parse`, `ResolveLocale`, and supported-locales filtering share the
same ECMA-402 first-wins keyword behavior and private-use insertion rule.

Locale-list canonicalization for public Go APIs is deliberately outside
`internal/localematcher`: raw strings are parsed at `locale.Parse` /
`locale.ParseList`, root `GetCanonicalLocales` exposes the public
`Intl.getCanonicalLocales` bridge, and formatter constructors use
`internal/ecma402.CanonicalLocaleList` / `RequestedLocaleStrings` before they
enter matching.

> **Why**:
> 1. **`internal/`** - matcher is the implementation details of formatter entry and is not exposed to users; SPEC 00 §3 has defined `internal/localematcher`.
> 2. **Files are split according to ECMA-402 abstract operation** - one file corresponds to a section of spec (`LookupMatcher.ts` / `BestFitMatcher.ts` / `ResolveLocale.ts`), mirroring Generated reference 1:1 to reduce migration costs.
> 3. **Data is injected by callers** - matcher does not import `internal/cldr`; formatter constructors pass generated supported-locale lists, maximizers, and relevant-extension-key data lookups.
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
AlgorithmBestFit // sec-BestFitMatcher (three tiers + bounded distance model)
)

// Result is an internal product of matcher, not ResolvedLocale; the latter is composed of ResolveLocale.
type Result struct {
Locale string // Hit supported locale string (canonical BCP 47)
DataLocale string // Locale used for CLDR data lookup.
Extension string // Unicode extension sequence from the matched request, when present.
Distance int // 0 = completely equivalent; >= threshold is considered a miss
}

// Match selects the supported locales that best matches the requested list.
// alg = AlgorithmLookup goes to LookupMatcher; alg = AlgorithmBestFit goes to BestFitMatcher.
// Either algorithm does not return an error (Result{Locale: defaultLocale} when there is no match).
func Match(requested, supported []string, defaultLocale string, alg Algorithm) Result

// DefaultMatchingThreshold is the BestFitMatcher Tier 3 distance threshold (Generated reference DEFAULT_MATCHING_THRESHOLD).
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
| Locale-only operations | `internal/cldr/locale.AvailableLocales()` |

Formatter constructors **MUST** pass the appropriate generated supported-locale accessor into `internal/ecma402.ResolveConstructorLocale`. The shared ECMA-402 entry owns requested-locale preparation, matcher dispatch, default fallback, and relevant-extension merging. After that shared step, each formatter owns its own CLDR data fallback and payload selection; there is no separate `internal/cldrmatch` platform.

`internal/localematcher.NewMatcher` compiles that raw supported list into the ECMA-402 Available Locales List used by lookup, best-fit, `ResolveLocale`, and `FilterLocales`. The compiled list is duplicate-free, lacks Unicode extension sequences, preserves concrete data locales, and adds only spec-mandated less-narrow fallbacks:

- ordinary truncation fallbacks such as `de` for `de-DE` when the raw payload has no separate `de` record;
- language-region aliases such as `az-AZ` for `az-Latn-AZ` and `zh-HK` for `zh-Hant-HK`.

Derived fallbacks are matching aliases, not new payload records. `Result.Locale` may be `zh-HK`; `Result.DataLocale` remains `zh-Hant-HK` so constructor locale data reads stay honest.
The matcher owns only the alias expansion rule; script and region subtag shape
checks come from `internal/localeid`, the same grammar owner used by
`Intl.Locale` and DisplayNames code canonicalization.

> **Why `Algorithm` is an int enumeration instead of a string**: Go iota is cheaper than string verification; ECMA-402 spec text is `"lookup"` / `"best fit"` (including spaces), `parseAlgorithm(string) (Algorithm, error)` is used at the boundary to convert, and all internals are int.
>
> **Why `DefaultMatchingThreshold = 838` instead of spec verbatim value**:generated-reference `intl-localematcher/abstract/utils.ts` defines `DEFAULT_MATCHING_THRESHOLD = 838`;sec-BestFitMatcher does not specify a specific value. We continue to use Generated reference values to ensure conformance tests are consistent.
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
| Generated reference three-tier algorithm shape (Tier 1 exact / Tier 2 maximize+truncation / Tier 3 bounded distance) | ✅ Selected | There is a public test baseline; aligned with the conformance goal without shipping an unused generated distance table |
| Reuse `golang.org/x/text/language.Matcher` | ❌ Reject | Output `Confidence`(No/Low/High/Exact) is not CLDR distance; tie-breaking is determined by `x/text`, inconsistent with Generated reference; `x/text`'s CLDR data version is not synchronized with `internal/cldr` |
| ICU-only simplified heuristic | ❌ Reject | No fixture exposed; cannot reverse verify from conformance test |

> **Why Generated reference algorithm**:
> 1. **conformance is a hard constraint** - SPEC 70 requires byte-equality to pass generated-reference `intl-localematcher/tests/conformance.test.ts`; only the same algorithm can be passed.
> 2. **CLDR distance is a meaningful scalar** - `Confidence` is level 4 discrete and cannot distinguish fine-grained differences such as "es-MX vs es-419"; CLDR distance (0-840+) can.
> 3. **Can be maintained independently** - The three layers of `internal/localematcher` are ~500 LOC, and the fixtures are all in Generated reference; when upgrading CLDR, you only need to rerun the fixture to find the regression.
>
> **Rejected `language.Matcher`**:
> - ❌ `language.Matcher`'s tie-breaking uses `Confidence` + `x/text` internal implementation details, **invisible** to the caller, unable to byte-match Generated reference.
> - ❌ The data version of `x/text` does not form the same conformance baseline as the ECMA-402/formatjs pinned version (SPEC 50 §1 locking CLDR 48.1.0).
> - ❌ Introduce a second CLDR data source to destroy data record consistency.
>
> **Fallout**: consumer-driven expansion can expose the `localematcher.WithLanguageMatcher()` option to allow users to switch to `x/text` (applicable to the "I only want the behavior of `x/text`" scenario); active scope is not implemented.

### 3.2 Three-layer algorithm

`findBestMatch(requested, supported, threshold)`(generated-reference `intl-localematcher/abstract/utils.ts` §findBestMatch):

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

# === TIER 3 — Bounded distance calculation ===
lowestDistance = +Inf # Reset: Tier 2's "position penalty" is not comparable to distance output
for i, desired in requested:
    for k, candidate in supported:
        d = matchingDistance(desired, candidate, maximize(desired), maximize(candidate))
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

> **Why Tier 3 reset `lowestDistance`**: The distance of Tier 2 is the "subtag removal position × 10 + request order × 40" heuristic, which is a different scale from the bounded matching distance used in Tier 3. Mixing the two turns an implementation detail of fallback order into a false semantic distance.
>
> **Why position penalty `i*40`**:Generated reference verbatim;ECMA-402 does not support `Accept-Language`'s `q=0.1` weighting, but preserves request order as weak priority.

### 3.3 Go signature

```go
// internal/localematcher/best_fit.go (signature)

// BestFitMatcher implements ECMA-402 §9.2.5 (Generated reference three-layer algorithm).
func BestFitMatcher(requested, supported []string, defaultLocale string) Result

// BestFitMatcherWithMaximizer lets tests and constructors inject the same
// maximizer used by Locale maximize/minimize.
func BestFitMatcherWithMaximizer(requested, supported []string, defaultLocale string, maximizer Maximizer) Result

// findBestMatch is the core three-tier entrance on a compiled Matcher.
type bestMatchResult struct {
    matchedDesired   string
    matchedSupported string
    distance         int
}

func (m *Matcher) findBestMatch(requested []string, threshold int) bestMatchResult

// getFallbackCandidates truncate the maximize results according to the right-to-left subtag to generate candidates.
// Example: "zh-Hant-TW" → ["zh-Hant-TW", "zh-Hant", "zh"]
func getFallbackCandidates(maximized string) []string

// matchingDistance returns the active bounded best-fit distance (memoized via sync.Map).
func matchingDistance(desired, supported, maximizedDesired, maximizedSupported string) int
```

> **Why `findBestMatch` is not exported**: It is a compiled matcher implementation detail. Public callers only need the selected `Result`; exposing the candidate-distance internals would freeze a non-ECMA-402 detail.
>
> **Why matching distance memoizes with `sync.Map`**: Repeated calculation of the same requested/supported/maximized tuple appears during constructor-heavy workloads and supported-locale filtering. The cache is process-local, bounded by locale-pair variety, and removes duplicate distance work without adding formatter-level cache controls.

### 3.4 Active bounded distance model

`matchingDistance` is intentionally small and local to the active ECMA-402
surface. It is not a generated CLDR `languageMatching.json` table.

| Case | Distance |
|------|----------|
| Requested and supported tags are identical | `0` |
| Maximized requested and supported tags are identical | `0` |
| Fixture-backed generated-reference-sensitive pairs, such as `en-CA` vs `en-US` or `es-KY` vs `es-419` | Recorded fixture distance |
| Same language without a fixture override | `40` |
| Different language | `840` |

`Matcher` adds request-order penalties and derived-fallback penalties around this
base distance, then compares the result against `DefaultMatchingThreshold`.

> **Why no generated languageMatching table**: ECMA-402 leaves best-fit matching implementation-defined. The active product profile needs stable, auditable behavior for the supported locale set, not an unused generated table that suggests broader ICU parity than the runtime can verify.

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
    Algorithm             Algorithm        // "lookup" | "best fit"
    Matcher               *Matcher         // optional compiled supported-locale index
    Requested             []string         // internal/ecma402.RequestedLocaleStrings output
    Supported             []string         // available locales when Matcher is nil
    DefaultLocale         string           // matcher fallback
    RelevantExtensionKeys []string         // Example: ["nu"] (NumberFormat)
    OptionValues          []Option         // explicit option overrides in stable order
    Options               map[string]string // map bridge for callers that already build keyed option data
    LocaleData            LocaleDataLookup // Inject from internal/cldr
    Maximizer             Maximizer        // optional maximize function when Matcher is nil
}

// LocaleDataLookup is a subset of the SPEC 12 §5 / SPEC 50 §6 data provider.
type LocaleDataLookup interface {
// For returns a list of legal values for a locale key. The first item is the locale default.
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

Formatter constructors do not build this input ad hoc. They pass typed
constructor state through `internal/ecma402.ResolveConstructorLocale`, which
owns only the shared negotiation wrapper:

```go
// internal/ecma402/constructor_locale.go(signature)
type ConstructorLocaleOptions struct {
    Locales               locale.List
    Fallback              locale.Locale
    LocaleMatcher         string
    Matcher               *localematcher.Matcher
    RelevantExtensionKeys []string
    OptionValues          []localematcher.Option
    LocaleData            localematcher.LocaleDataLookup
}

type ConstructorLocaleResolution struct {
    Locale     locale.Locale
    DataLocale string
    Extensions map[string]string
}

func ResolveConstructorLocale(opts ConstructorLocaleOptions) ConstructorLocaleResolution
```

The wrapper must not absorb formatter-owned behavior: CLDR data fallback,
unsupported-option errors, pattern selection, calendar/hour-cycle defaults,
numbering-system fallback, plural-rule lookup, time-zone handling, and embedded
formatter construction remain in the constructor package.

`internal/ecma402.ResolveConstructorLocale` is the only production entrypoint
from formatter construction into `localematcher.ResolveLocale`. Formatter
packages may prepare matcher inputs, validate formatter-owned unsupported
states, and inspect selected Unicode extension values, but they must not call
`ResolveLocale` directly or copy its negotiation sequence.

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
> **Why `LocaleData` is an interface rather than a specific type**: Same as §3.4 - Decoupling the implementation details of SPEC 11 and SPEC 50. formatter passes in specific types such as `internal/cldr/number.NumberLocaleData{}` or `internal/cldr/date.DateLocaleData{}` when calling `New()`.

### 4.3 InsertUnicodeExtensionAndCanonicalize

`InsertUnicodeExtensionAndCanonicalize(locale, extension)`(generated-reference `intl-localematcher/abstract/InsertUnicodeExtensionAndCanonicalize.ts`):

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

ECMA-402 §6.2.1 `CanonicalizeLocaleList(locales)` accepts JavaScript dynamic
values, validates strings / `Intl.Locale` instances, canonicalizes each entry,
and returns the first occurrence of every canonical locale identifier.

Go splits that operation at a typed boundary:

1. Raw strings enter through `locale.Parse`, `locale.ParseList`, or `locale.New`.
2. Parsed `locale.Locale` values already hold canonical Unicode locale IDs.
3. `internal/ecma402.CanonicalLocaleList` removes duplicates by `Locale.String()` while preserving first-seen order.
4. Root `gointl.GetCanonicalLocales` is the only public ECMA-402 bridge for this operation.
5. Formatter constructors and `SupportedLocalesOf` use the internal helper; formatter-independent public canonicalize-list helpers and `any`-accepting public helpers are not part of the surface.

### 5.2 Go signature

```go
// internal/ecma402/locale_list.go(signature)

// CanonicalLocaleList returns the first occurrence of each canonical locale
// while preserving request order.
func CanonicalLocaleList(locales locale.List) locale.List

// RequestedLocaleStrings returns canonical requested locale identifiers for
// ResolveLocale. nil means the locales argument was omitted or empty.
func RequestedLocaleStrings(locales locale.List) []string

// ValidationLocale returns a concrete locale for constructor option error
// context before locale negotiation has resolved.
func ValidationLocale(locales locale.List) locale.Locale
```

> **Why no public `any` API**: ECMA-402 needs `any` because JavaScript accepts strings, arrays, and `Intl.Locale` objects at runtime. Go callers should parse raw strings once and then pass `locale.List`. Simulating JavaScript's dynamic boundary in public Go API would make every formatter call less clear without improving conformance.
>
> **Why no error return**: Canonical locale-list dedupe is error-free after parsing. Invalid BCP 47 input remains the responsibility of `locale.Parse` / `locale.ParseList`.

### 5.3 Input normalization

| Input type | Processing |
|---------|------|
| `nil` / empty `locale.List` | Return an empty `locale.List`; `RequestedLocaleStrings` returns `nil` |
| `locale.List` | Keep the first locale for each canonical `loc.String()` |
| Raw strings | Parse first with `locale.Parse` / `locale.ParseList`; do not pass raw strings to internal canonicalization |

This keeps ECMA-402 behavior while preserving the repository rule that public
exported symbols map to native owners, not abstract operations.

---

## 6. FilterLocales and LookupSupportedLocales

### 6.1 Purpose of FilterLocales

`Intl.NumberFormat.supportedLocalesOf(locales, options)` etc. The semantic source of spec static methods. Go typed bridge receives values already parsed into `locale.Locale`; it then deduplicates by canonical locale string and applies `options.localeMatcher` to choose between `lookup` and `best fit`. The result is the subset of requested locales that can be matched by the supported list, preserving the requested locale value, relative order, and Unicode extensions.

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
    locale.List{mustLocale("en-US-u-nu-latn"), mustLocale("fr-FR")},
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
    []string{"zh-TW", "fr-FR"},
)
fmt.Println(out) // For example ["zh-TW", "fr-FR"]
```

---

## 7. Error handling

### 7.1 Sentinel Error

Locale parsing failures classify through root `gointl.ErrInvalidValue` before
values enter matcher code.

> **Why there is no matcher’s own sentinel**: matcher internal `Match` / `ResolveLocale` **does not return an error** (spec requires always falling back to `defaultLocale`); only the locale input parsing phase will fail and be uniformly mapped to the root category.

### 7.2 Reconciling with `locale` package errors

Locale parsing errors classify through `gointl.ErrInvalidValue`; `internal/localematcher` does not define or re-export a public sentinel.

> **Why root category**: When users catch errors via the formatter package, `errors.Is(err, gointl.ErrInvalidValue)` should match all locale parsing failures; multiple independent sentinels will break this equivalence.

---

## 8. Boundary to SPEC 12 / SPEC 50

### 8.1 Relationship with SPEC 12

| SPEC | Provide | Consumption |
|------|------|------|
| SPEC 11 (this) | `ResolveLocale` returns `ResolvedLocale` and documents the locale-list abstract operation contract | receives requested-locale strings and option data from `internal/ecma402.ResolveConstructorLocale` |
| SPEC 12 | Formatter-independent wrappers / validators / pattern / decimal boundary for production path reuse | wraps SPEC 11 algorithms without importing generated CLDR data |

> **DECISION**: "locale-shape abstract operations" such as `BestAvailableLocale` and `ResolveLocale` are documented in **SPEC 11**. Reusable typed constructor-entry helpers such as canonical locale-list dedupe and `ResolveConstructorLocale` live in `internal/ecma402` because formatter constructors need a Go typed boundary, while the actual matching algorithms stay in `internal/localematcher`.

### 8.2 Relationship with SPEC 50

| SPEC | Provide | Consumption |
|------|------|------|
| SPEC 11 (this) | `Matcher`, `Maximizer`, and `LocaleDataLookup` interfaces | receives supported-locale slices, maximizers, and locale-data lookups from formatter constructors |
| SPEC 50 | generated supported-locale accessors, locale maximizer, and relevant-extension-key data providers | Not consumed directly by SPEC 11 |

Dependency direction: formatter packages inject SPEC 50 data into SPEC 11 algorithms. `internal/localematcher` must not import `internal/cldr` directly.

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

> **Why**: `Confidence`(No/Low/High/Exact) Level 4 discrete, unable to byte-match generated-reference `distance` value; tie-breaking is determined internally by `x/text` and is not visible.

### 9.2 ❌ Do not use `import internal/cldr` directly in matcher

```go
// ❌ Error: SPEC 11 and SPEC 50 become tightly coupled.
import "github.com/agentable/go-intl/internal/cldr"
func NewMatcherForNumbers() *Matcher {
    return NewMatcher(cldr.NumberSupportedLocales(), cldr.Maximize)
}

// ✅ Correct: formatter packages inject generated data.
matcher := localematcher.NewMatcher(cldrnumber.SupportedLocales(), cldrlocale.Maximize)
```

> **Why**: CLDR data updates (SPEC 50) should not force matcher package changes. Formatter packages already own the constructor/data boundary and can inject exactly the data family they use.

### 9.3 ❌ Do not return error in `Match` / `ResolveLocale`

```go
// ❌ Error: Violation of ECMA-402 spec("matcher always succeeds")
func Match(...) (Result, error)

// ✅ Correct: Return defaultLocale when there is no match
func Match(requested, supported []string, defaultLocale string, alg Algorithm) Result
```

> **Why**: ECMA-402 §9.2.5 / §9.2.6 both stipulate that "no match returns defaultLocale"; making the caller handle a matcher error adds cognitive load and contradicts the spec. User-input errors belong at the raw locale parsing boundary before matcher code runs.

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

> **Why**: Tier 2 distance is heuristic (subtag position × 10), while Tier 3 uses the active bounded matching distance plus request-order and derived-fallback penalties. The two are not the same scale.

### 9.7 ❌ Don’t skip memoization in matching distance

```go
// ❌ Error: recompute the same distance tuple repeatedly.
func cachedMatchingDistance(d, s, md, ms string) int {
    return matchingDistance(d, s, md, ms)
}

// ✅ Correct: sync.Map memoize.
var distanceCache sync.Map // map[[4]string]int
func cachedMatchingDistance(d, s, md, ms string) int {
    key := [4]string{d, s, md, ms}
    if v, ok := distanceCache.Load(key); ok { return v.(int) }
    dist := matchingDistance(d, s, md, ms)
    distanceCache.Store(key, dist)
    return dist
}
```

> **Why**: Tier 3 compares every requested locale with every supported locale. Memoization removes repeated tuple work across constructors without creating public cache controls.

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

- [ ] `internal/localematcher/` subdirectory is divided into files `available.go` / `compiled.go` / `match.go` / `lookup.go` / `best_fit.go` / `distance.go` / `filter.go` / `resolve.go` / `ucanonicalize.go`.
- [ ] `internal/localematcher` is not directly accessed `import "github.com/agentable/go-intl/internal/cldr"`; callers inject generated supported-locale slices, maximizers, and locale-data lookups.
- [ ] `internal/localematcher` is not directly `import "github.com/agentable/go-intl/internal/ecma402"` (SPEC 12 is option-shape; this SPEC is locale-shape; there is no loop between the two).

### LookupMatcher

- [ ] `LookupMatcher(requested, supported, defaultLocale) Result` implements ECMA-402 §9.2.6 verbatim.
- [ ] `BestAvailableLocale(supported, locale) string` implements ECMA-402 §9.2.4 (single-character subtag skips position -2).
- [ ] generated-reference `intl-localematcher/tests/LookupMatcher.test.ts` All fixtures pass in `internal/localematcher/lookup_test.go`.

### BestFitMatcher

- [ ] `BestFitMatcher(requested, supported, defaultLocale) Result` implements the three-tier algorithm (Tier 1 exact / Tier 2 maximize+truncation / Tier 3 bounded distance).
- [ ] `findBestMatch` resets `lowestDistance = +Inf` on Tier 3 entry (not mixed with Tier 2 heuristics).
- [ ] `getFallbackCandidates(maximized)` output `["zh-Hant-TW","zh-Hant","zh"]` (right-to-left subtag truncation).
- [ ] `cachedMatchingDistance` uses `sync.Map` memoize for the same desired/supported/maximized tuple.
- [ ] `DefaultMatchingThreshold = 838`(Generated reference verbatim).
- [ ] generated-reference `intl-localematcher/tests/BestFitMatcher.test.ts` and `tests/conformance.test.ts` all fixtures pass in `internal/localematcher/best_fit_test.go`.

### ResolveLocale

- [ ] `ResolveLocale(opts) ResolvedLocale` implements ECMA-402 §9.2.7 (option > extension > localeData default priority).
- [ ] `relevantExtensionKeys` is a dynamic parameter (NumberFormat uses `["nu"]`, DateTimeFormat uses `["ca","nu","hc"]`).
- [ ] `InsertUnicodeExtensionAndCanonicalize` outputs `-u-` keys in lexicographic order (ca < co < fw < hc < kf < kn < nu).
- [ ] `UnicodeExtensionValue` returns `"true"` in the default type (`-u-kn` writes the scene separately).
- [ ] generated-reference `intl-localematcher/tests/ResolveLocale.test.ts` All fixtures pass.

### CanonicalizeLocaleList

- [ ] `internal/ecma402.CanonicalLocaleList(locale.List) locale.List` accepts `nil` / empty `locale.List` / ordered locale request list.
- [ ] raw string locale only occurs at `locale.Parse` boundaries; locale-list canonicalization does not create impossible errors.
- [ ] Keep the order of first appearance after deduplication.
- [ ] No public formatter-independent canonicalize-list helper or `any`-accepting canonical locale-list helper exists.

### LookupSupportedLocales

- [ ] `FilterLocales(supported, requested, matcher) locale.List` implements ECMA-402 `FilterLocales`.
- [ ] `FilterLocales` canonical-deduplicates requested locale values before filtering.
- [ ] `FilterLocales` preserves requested locale order and Unicode extensions for matched locales.
- [ ] Constructor package `SupportedLocalesOf` methods call `internal/ecma402.SupportedLocalesOf` rather than duplicating matcher loops.
- [ ] `LookupSupportedLocales(supported, requested) []string` implements ECMA-402 §9.2.8.
- [ ] Output stripped of `-u-` extension.

### Error

- [ ] Raw locale-list parsing errors are owned by `locale.Parse` / `locale.ParseList`; canonical locale-list dedupe is error-free after parsing.
- [ ] `Match` / `ResolveLocale` does not return error (always falls back to `defaultLocale`).
- [ ] Formatter constructors use `internal/ecma402.ResolveConstructorLocale` for common negotiation; production code outside `internal/ecma402/constructor_locale.go` does not call `localematcher.ResolveLocale` directly.
- [ ] There is **no** `panic` call in the package (the test covers various abnormal inputs).

### Performance

- [ ] Tier 1 hit path benchmark < 200 ns(`go test -bench=BenchmarkTier1Match`).
- [ ] Tier 3 steady state (memoize hit) 100-locale supported list benchmark < 100 µs.
- [ ] `sync.Map` memoize has no race (`-race` passes) under concurrent 10 goroutines.

### Boundary with `language.Matcher`

- [ ] `internal/localematcher` package does not import `golang.org/x/text/language.Matcher` / `MatchStrings` (`grep -r "language.Matcher" internal/localematcher/` should be empty).
- [ ] `Locale.Maximize` / `Minimize` (SPEC 10) are reused within Tier 2 (do not reimplement the likelySubtags query).

### Test

- [ ] generated-reference `intl-localematcher/tests/locale-match-fixtures.json` is ported to `internal/localematcher/testdata/match-fixtures.json` and consumed in `match_test.go` table driver.
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
- `.references/formatjs/packages/intl-localematcher/abstract/languageMatching.ts` —— Reference-only CLDR matching table for future best-fit expansion; generated CLDR accessors are not consumed by the active matcher.
- `.references/formatjs/packages/intl-localematcher/tests/conformance.test.ts` —— ICU4J alignment fixture
- `.references/formatjs/packages/intl-localematcher/tests/locale-match-fixtures.json` —— table-driven fixture
- `.references/ext/src/ecma402/locale.cpp` —— PHP `ecma402_bestAvailableLocale` via ICU (same as spec but different path)

### Cross-SPEC

- [SPEC 00 §8 Q5 — BestFit Matcher Implementation Selection](./00-vision-and-scope.md#8-open-questions)(This SPEC is closed)
- [SPEC 10 §4 — Maximize & Minimize](./10-locale.md#4-maximize--minimize) — Tier 2 fallback
- [SPEC 10 §1 — Locale structure](./10-locale.md#1-locale-structure) — `Locale.String()` is the matcher input
- [SPEC 12 §3 — Option Validation](./12-abstract-operations.md#3-option-validation) —— matcher does not implement formatter option validation repeatedly
- [SPEC 12 §5 — Internal Slots](./12-abstract-operations.md#5-internal-slots) —— `[[Locale]]` slot holds `ResolvedLocale.Locale`
- [SPEC 50 §6 — Data Access API](./50-cldr-data.md#6-data-access-api) —— generated supported-locale accessors, maximizers, and `LocaleDataLookup` providers
- [SPEC 70 §Conformance](./70-conformance.md) —— matcher fixture is part of conformance test


---

> This SPEC is a maintenance record for `internal/localematcher`. New ECMA-402 matcher subroutines or a deliberate move to generated CLDR `languageMatching.json` data trigger this SPEC revision before code changes.
