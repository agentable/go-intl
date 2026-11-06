# Segmenter feasibility — issue #10

2026-11-06. **Decision: keep #10 open.** Neither tested pure Go backend is a
complete replacement for the issue's Node/ICU word-navigation target. This is a
bounded evaluation of two candidates, not evidence that every pure Go approach
is impossible. No public Segmenter, partial supported-locale claim, runtime
Node/cgo dependency, or root-module dependency is introduced.

## Reproduce

Run `NODE_BIN=/path/to/node-v26.0.0 GO_BIN=/path/to/go1.27.3 GOTOOLCHAIN=local bash reports/segmenter-feasibility/probe.sh` from the repository.
The command is one shell line. It creates and removes a temporary module,
downloads the two pinned candidates, prints native records and candidate
boundaries/offsets, then reports mismatches for seven explicit locales. Set
`NODE_BIN` to the actual binary; the script rejects other Node versions.
For example, `npm exec --yes --package=node@26.0.0 -- which node` locates a
versioned npm-provided binary without changing the host Node installation.

Native target: **Node v26.0.0, V8 14.6.202.33-node.19, ICU 78.3,
Unicode 17.0, CLDR 48.0** (the active repository witness version).
Probe environment: macOS arm64, Go 1.27.3, default unstripped Go build.
This diagnostic does not publish conformance goldens or modify the existing
witness baseline.

## Candidates and data provenance

| Candidate | Pinned source | Unicode / license | Outcome |
| --- | --- | --- | --- |
| `github.com/clipperhouse/uax29/v2` v2.7.0 | `b03477d1fbba89df95a6b55da0a222b2e2228610` | Unicode 17 UAX #29; MIT | Does not supply the dictionary/CJK word behavior needed here |
| `github.com/rivo/uniseg` v0.4.7 | `03509a98a092b522b2ff0de13e53513d18b3b837` | Unicode 15 UAX #29; MIT | Same observed boundary gaps; older Unicode target as well |

Module content checksums, verified through the Go module download path:

```text
uax29/v2 v2.7.0 h1:+gs4oBZ2gPfVrKPthwbMzWZDaAFPGYK72F0NJv2v7Vk=
uniseg v0.4.7 h1:WUdvkW8uEhrYfLC4ZzdpI2ztxP1I582+49Oc5Mq64VQ=
```

Candidate README/license files and the pinned word APIs were inspected. They
are useful UAX #29 implementations; this evaluation does not assert that their
boundary differences violate Unicode or ECMA-402. No dictionary/model data was
selected for redistribution. An ICU port would need a separate inventory and
license review of its rules, dictionaries, models and generation tooling;
these candidate licenses do not qualify ICU data by implication.

## Observed boundaries and offsets

Both candidates produced the same boundaries on these 16 inputs. For **each**
of `en`, `th`, `zh`, `ja`, `km`, `lo`, `my`, each candidate differed from native
on **13/16** complete boundary/index/byte-index records. The three matching
inputs were the emoji/combining-mark index cases. Native locale resolution
returned the explicitly requested locale for all seven.

| Input | Node v26.0.0 word segments (`en`) | Both candidates |
| --- | --- | --- |
| `ไทยภาษา` | ไทย / ภาษา | ไ / ท / ย / ภ / า / ษ / า |
| `ภาษาไทยx` | ภาษา / ไทยx | ภ / า / ษ / า / ไ / ท / ย / x |
| `中文测试` | 中文 / 测试 | 中 / 文 / 测 / 试 |
| `日本語の文章` | 日本語 / の / 文章 | 日 / 本 / 語 / の / 文 / 章 |
| `ខ្មែរ` | ខ្មែរ | ខ្ / មែ / រ |
| `ພາສາລາວ` | ພາສາ / ລາວ | ພ / າ / ສ / າ / ລ / າ / ວ |
| `မြန်မာစာ` | မြန်မာ / စာ | မြ / န် / မာ / စာ |
| `😀x` | 😀 / x | 😀 / x |
| `áb` | áb | áb |
| `👩‍💻x` | 👩‍💻 / x | 👩‍💻 / x |
| `ภาษาไทยภาษาไทย` | ภาษา / ไทย / ภาษา / ไทย | ภ / า / ษ / า / ไ / ท / ย / ภ / า / ษ / า / ไ / ท / ย |
| `研究生命起源` | 研究 / 生命 / 起源 | 研 / 究 / 生 / 命 / 起 / 源 |
| `東京都に行く` | 東京 / 都 / に / 行く | 東 / 京 / 都 / に / 行 / く |
| `English ไทยภาษา 中文测试` | English /   / ไทย / ภาษา /   / 中文 / 测试 | English /   / ไ / ท / ย / ภ / า / ษ / า /   / 中 / 文 / 测 / 试 |
| `ไทยxyzภาษา` | ไทยxyzภาษา | ไ / ท / ย / xyz / ภ / า / ษ / า |
| `กำลังทดสอบระบบ` | กำลัง / ทดสอบ / ระบบ | ก / ำ / ลั / ง / ท / ด / ส / อ / บ / ร / ะ / บ / บ |

The probe prints native `isWordLike`, UTF-16 `index` and UTF-8 `byteIndex` for
every segment; candidate comparison checks segment text and both offsets.
Neither candidate word API provides ICU's word-status classification, so this
probe does **not** claim candidate `isWordLike`, locale negotiation,
`Containing`, or grapheme/sentence compatibility.
For `😀x`, native `x` starts at UTF-16 2 / byte 4; for `👩‍💻x`, at 5 / 11.
Emoji records are non-word-like. A Go rune index cannot stand in for either
public offset contract.

The 200,704-byte mixed-script long-input probe materialized 73,728 candidate
records. One run measured approximately 5.47 ms / 73,755 allocations for uax29
and 9.85 ms / 73,755 allocations for uniseg. These numbers include the
prototype's record collection and UTF-16 conversion, and are **not** isolated
engine benchmarks. The combined unstripped probe binary was 4,076,498 bytes;
that is not a production backend size delta. Performance is not the rejection
reason: the required word boundaries already fail.

## Standard and implementation evidence

ECMA-402 `.references/ecma402/spec/segmenter.html`, `FindBoundary` and
`CreateSegmentDataObject`, defines API/record/index semantics while allowing
implementation-defined boundaries and locale tailoring. Node parity is a
separate consumer acceptance target, not a claim that one ICU dictionary is
normatively mandated.

The vendored Node implementation uses ICU BreakIterator in
`deps/v8/src/objects/js-segmenter.cc`; `js-segments.cc` derives `isWordLike`
from ICU rule-status ranges. ICU's
`deps/icu-small/source/common/brkeng.cpp:ICULanguageBreakFactory::loadEngineFor`
selects dictionary/LSTM engines for complex scripts and a CJK engine;
`dictbe.cpp` includes normalization and original-offset handling. A longest-word
lookup or a letter-category heuristic is not established equivalent behavior.
FormatJS `packages/intl-segmenter/segmenter.ts` supplies useful API/rule evidence,
not proof that these candidates match ICU's language engines.
These paths are relative to `.references/node` or `.references/formatjs`.
Pinned reference revisions are recorded in the root `improve.md`.

## Follow-up gate

Before implementation, qualify a maintainable pure Go dictionary/model backend
or a bounded ICU algorithm/data port with reproducible pins and redistribution
rights. It must pass the issue corpus plus ambiguous/unknown words and mixed
scripts under the fixed native target. Then implement and verify all three
granularities, negotiation, word status, UTF-16/byte containment, invalid UTF-8
handling and independent iterators as planned in M04/M05 of `improve.md`.
This investigation stops M03 at a documented backend gap; M04/M05 remain
conditional and #10 is not resolved by the LocaleMatcher contribution.
