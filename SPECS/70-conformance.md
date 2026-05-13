# SPEC 70 — Conformance Test Strategy

> **Status:** Draft (2026-05-08)
> **Type:** Flow + Schema + Rule — defines how go-intl proves ECMA-402 observable behavior through FormatJS-derived fixtures.
> **Authority:** This spec is SSOT for conformance fixture format, fixture sources, divergence handling, the three-gate CI policy, and XFAIL discipline. Per-formatter SPECS (10/20/30/40) own their own option semantics; this spec owns how those semantics are *verified* against the reference implementations.

---

## Overview

go-intl active scope 的"正确性"由 **ECMA-402 规范 + fixture-driven conformance tests** 定义:规范决定语义边界,fixtures 从 FormatJS 测试机械或人工抽取 input/output 对,在 Go 端 harness 内逐条断言。本 SPEC 覆盖:

1. **fixture 格式 SSOT**:统一 JSON schema,跨 formatter 通用。
2. **fixture 来源**:FormatJS tests 和少量手写 ECMA-402 边界用例。
3. **divergence 流程**:已知偏差登记 `<package>/testdata/divergences.md`,**禁止** silently skip。
4. **CI 三 gate**:conformance(block)+ divergence audit(block)+ performance(PR warn / main block)。
5. **XFAIL 时效**:每条 XFAIL 必须有过期日期,过期自动 fail。

> **Why**: ECMA-402 一一对齐是 SPECS/00 §1 的核心承诺;fixture 是承诺的执行机制。SPEC 70 把 R08 调研结论固化为可执行规则。
> **Rejected**: 仅靠手写 unit test 覆盖 ECMA-402 行为——formatjs 单一 polyfill 已积累数千条 fixture,手写 0% 可能赶上覆盖率。

---

## 1. Fixture Format SSOT

### 1.1 JSON Schema

每条 fixture **必须**符合下述 schema(JSON object):

```json
{
  "id": "intl-numberformat/percent/en/notation=compact_signDisplay=always_long",
  "source": "formatjs:packages/intl-numberformat/tests/percent/__snapshots__/en.test.ts.snap",
  "locale": "en",
  "options": {"style": "percent", "notation": "compact"},
  "input": 1.0,
  "expected": "+1 million",
  "expectedParts": [{"type": "plusSign", "value": "+"}, {"type": "integer", "value": "1"}]
}
```

### 1.2 字段契约

| 字段 | 必填 | 类型 | 含义 |
|------|------|------|------|
| `id` | 必填 | string | 全局唯一;稳定 slug,建议包含 formatter、source、feature / case index;divergences.md 通过此 id 引用 |
| `source` | 必填 | string | `<source>:<path>` —— `formatjs:` 或 `manual:` |
| `locale` | 必填 | string | BCP 47 locale tag,Go 端在 fixture 边界用 `locale.Parse(string)` 解析 |
| `options` | 必填 | object | ECMA-402 选项;键名按 spec 原文(camelCase) |
| `input` | 必填 | number / string / object | 数字字面量、ISO-8601 字符串(DateTimeFormat)、`{start, end}` 对象(Range) |
| `expected` | 可选 | string | `Format` 输出;FormatToParts-only fixture 可省略 |
| `expectedParts` | 可选 | array | `FormatToParts` 输出;省略表示不校验 parts |
| `expectedRange` | 可选 | string | `FormatRange` 输出 |
| `expectedRangeParts` | 可选 | array | `FormatRangeToParts` 输出 |
| `errorCode` | 可选 | string | 错误用例(替代 `expected`),对应 go-intl sentinel 名称 |

**必须遵守**:

1. fixture 文件 **必须**是 JSON 数组(每文件多条 fixture),路径 `<package>/testdata/conformance/<source>/<file>.json`。
2. `id` **必须**全局唯一;违反时 CI 直接 block(`tools/check-fixtures/` 校验)。
3. `input` 为日期时 **必须**用 ISO-8601 with timezone offset 字符串(如 `"2020-01-01T00:00:00Z"`),Go harness 通过 `time.Parse(time.RFC3339, ...)` 反序列化。**禁止**使用 Unix epoch ms 数字(formatjs `Date.now()` 风格),歧义太大。
4. 错误用例 **必须**走 `errorCode` 字段单独存档于 `<package>/testdata/conformance/<source>/errors.json`,与正向用例分文件;Go harness 走 `errors.Is(err, sentinel)` 校验。
5. `options` 字段 **必须**保持 ECMA-402 spec 原文命名(`maximumFractionDigits` 而非 `MaximumFractionDigits`);Go 端 harness 在加载时映射为 typed `Options` 值。
6. **禁止**在 fixture 中嵌入 JS 函数、回调、Date 字面量;不可机械抽取的部分按 SPEC §2.4 流程归类。

> **Why**: 统一 schema 跨 formatter 通用,harness 可以共享;一份 schema 也是 messageformat-go 未来集成测试的契约。
> **Rejected**: 每个 formatter 自定义 schema(NumberFormat `"value"` vs DateTimeFormat `"date"`)——4 套 harness、4 套加载器,DRY 违反。

### 1.3 Manifest

`<package>/testdata/conformance/<source>/MANIFEST.json` **应当**记录每个 fixture 文件的:

| 字段 | 含义 |
|------|------|
| `extractor_version` | `tools/gen-fixtures-from-formatjs/` 的版本 tag |
| `extracted_from` | formatjs 提交 SHA + 测试文件路径 |
| `extracted_at` | ISO 8601 时间戳 |
| `count` | 该文件中 fixture 条数 |

**应当**(非 active scope 强制):manifest 由 PR-CI lint 校验缺失。active scope 可通过 git blame 替代;consumer-driven expansion 触发条件 = fixture PR 数量 > 5/月。

> **Why**: manifest 是审计轨迹——formatjs 升级导致 fixture 漂移时,manifest 让 reviewer 一眼定位"哪个 PR 引入"。
> **Rejected**: active scope 强制 manifest——R08 §6.3 指出初版可用 git blame 替代,YAGNI。

---

## 2. Fixture Sources

### 2.1 Fixture source matrix

| Source | 角色 | active scope 要求 |
|--------|------|-------------|
| **formatjs Vitest** | **主 fixture source** | 已生成 fixture 必过;未生成 source 必须进入 `.skip-list.json` 审计 |
| **manual ECMA-402 edge cases** | 补充 fixture source | 仅用于 FormatJS 无法机械抽取或 spec 明确要求的边界 |

**必须遵守**:

1. 主 fixture **必须**来自 formatjs Vitest,通过 `tools/gen-fixtures-from-formatjs/` 转译;normative 解释权仍归 ECMA-402 spec。
2. Manual fixture **必须**说明对应 ECMA-402 section 或本地 SPEC rationale,并用 `source: "manual:<topic>"` 标记。
3. fixture 来源 **必须**通过 `source` 字段标记;**禁止**混合来源到同一 JSON 文件。
4. formatjs 版本 **必须**钉死在 `tools/.gen-versions`;当前值为本地 `.references/formatjs` 引用。升级 formatjs 或刷新 submodule 后必须重跑 extractor,并审查生成 fixture 与 `.skip-list.json` diff。

> **Why**: FormatJS is the maintained TypeScript ECMA-402 polyfill and the only vendored implementation reference. Keeping fixture provenance narrow reduces license and toolchain surface.

### 2.2 抽取工具

`tools/gen-fixtures-from-formatjs/` **必须**:

1. 是独立 Go module(独立 `go.mod`),与主 module 解耦。
2. 输入:formatjs `packages/<polyfill>/tests/` 全部 `.test.ts` 与 `__snapshots__/*.snap`。
3. 输出:`<package>/testdata/conformance/formatjs/<source-slug>.json`;每个 JSON 文件 **必须**只包含一个 `source` 值,防止 `tools/check-fixtures` 的 mixed-source gate 失效。
4. 机械抽取 **必须**只覆盖能无损确定 `{locale, options, input, expected}` 的断言形态。当前 active extractor 支持:
   - `const nf = new NumberFormat("en", {...}); expect(nf.format(42)).toBe("42")`
   - `expect(new Intl.NumberFormat("en", {...}).format(42)).toEqual("42")`
   - `NumberFormat` 的 `formatToParts` / `formatRange` / `formatRangeToParts` 直接字符串或 parts array 断言。
   - `const pr = new PluralRules("en", {...}); expect(pr.select(1)).toBe("one")`
   - `expect(new Intl.PluralRules("fr").select(1000000n)).toBe("many")`
   - `PluralRules.selectRange(start, end)` 的静态数值 / BigInt 字面量断言。
   - `const dtf = new DateTimeFormat("en-US", {...}); expect(dtf.format(date)).toBe("...")`
   - `DateTimeFormat.formatRange` / `formatRangeToParts` 中可静态还原为 RFC3339 字符串的 `Date` 输入。
   - `Intl.Locale` / `Intl.getCanonicalLocales` 的 `toString` / `maximize` / `minimize` / canonicalization 字符串断言。
   - 简单 JS object literal options:string / number / boolean 值。
   - PluralRules BigInt 输入以十进制字符串写入 fixture,避免 Go 端用 float64 承载整数语义。
5. 下列来源 **必须**写入 `.skip-list.json`,每条包含 `source`、`category` 与 `reason`,禁止静默丢弃:
   - `__snapshots__/*.snap` 但缺少 source test input mapping。
   - `it.each(table)` / `tests` 数组 / 回调 / 变量期望值等无法无损静态还原的 Vitest shape。
   - 已能机械发现但超出当前 generated fixture gate 的断言(例如 locale/unit/compact/currency-name/selectRange 行为尚未进入该 gate)。
   - 同一 source 内只有部分断言可无损抽取时,该 source 仍必须进入 `.skip-list.json`,reason 必须说明剩余断言为何未进入 gate。
6. 错误用例(`expect(...).toThrow(...)`)**必须**写入 `errors.json`,与正向用例分文件。

`.skip-list.json` category 值 **必须**来自下列集合:

| category | 含义 |
|----------|------|
| `unsupported-extractor-shape` | 源文件含 `expect(...)`,但当前 extractor 不能无损静态还原 |
| `partial-extraction` | 同一 source 中部分断言已生成 fixture,剩余断言仍未覆盖 |
| `snapshot-source` | snapshot 需要反查输入 mapping,不能单独生成 fixture |
| `scope-exclusion` | FormatJS source 属于 active ECMA-402 surface 之外 |
| `pending-implementation-gap` | 源断言可抽取,但 go-intl 尚未实现对应行为 |
| `accepted-divergence` | 已生成或可生成 fixture,但项目接受与引用实现不同;必须同时填写 `divergenceId` |
| `missing-reference` | 本地 `.references/formatjs` 或对应 tests path 缺失 |

**必须遵守**:

1. 抽取 **必须**幂等——同 formatjs commit 跑两次产出 byte-identical。
2. 抽取产物 **禁止**人工编辑;手动 fixture 走 `<package>/testdata/conformance/manual/<file>.json`,**不**与 formatjs/ 目录混淆。
3. `.skip-list.json` 是 extraction audit,不是 test skip 机制。已生成但失败的 fixture **必须**走 divergences.md 或 xfail.json;未生成的 source 才能出现在 skip-list。
4. 不可机械化的 fixture(R08 §1.2 的 Date 字面量、回调、错误)**必须**走人工迁移;**禁止** silently skip。

> **Why**: 抽取脚本是连接 formatjs 与 go-intl 的唯一可信桥梁;幂等性保证升级 formatjs 时 diff 可读。
> **Rejected**: 盲目 AST 全量移植——不完整 AST 规则会生成看似正式但输入/选项错误的 fixture。宁可把复杂来源写入 source/reason skip-list,也不生成不可信 fixture。

---

## 3. Divergence Handling

### 3.1 divergences.md 文件

每个 formatter 包 `<package>/testdata/divergences.md` **必须**存在(可初始为空)。每条 divergence 条目 **必须**包含:

| 字段 | 含义 |
|------|------|
| `id` | 与 fixture `id` 字段一一对应 |
| `source` | `formatjs:` 或 `manual:` |
| `our` | go-intl 实际输出(转义不可见字符) |
| `reference` | 对照实现输出 |
| `category` | R08 §3.1 的类别(compact-threshold / unit-plural / tz-abbr / canonicalization / range-plural / hour12-default / cldr-version-pin / etc.) |
| `rationale` | 为何接受此分歧(spec 实现定义 / CLDR 数据版本 / FormatJS 差异) |
| `review_after` | 下次评审锚点(CLDR 升级 / Go 1.27 / 季度回顾日期) |

### 3.2 处置流程

**必须遵守**:

1. CI fixture runner **必须**读取 divergences.md 的 `id` 列表,命中即跳过断言。
2. 任何不在 divergences.md 列表的失败 **必须** block `task verify`。
3. divergences.md 修改 **必须** 经 PR 显式审批(reviewer ≥ 1 maintainer);**禁止** 自动登记。
4. divergences.md 中的 `id` **必须**能在 fixture 文件中找到对应条目;CI lint(`tools/check-divergences/`)校验完整性。
5. divergence 条目 **禁止** 删除;过时条目改 `status: resolved` 标注 + 保留审计轨迹。

> **Why**: divergences.md 是 spec 解释权的人审通道;自动登记会让"未通过 fixture"事故被静默掩盖。
> **Rejected**: silently skip(将 fixture 移到 `testdata/skipped/`)——本 SPEC §6 forbid。

### 3.3 已知 divergence 类别

下述类别 **必须**在 divergences.md 中可见(出现时):

| 类别 | 示例 |
|------|------|
| compact-threshold | formatjs `1.2K` 起算点与某 locale ICU 阈值不同 |
| canonicalization | `en-arab-US` ⇒ `en-Arab-US` 大小写规则(本类强制对齐,出现即 bug) |
| range-plural | `(1, 1.5)` ⇒ `'one'` / `'few'` 取决 CLDR `pluralRanges.json` |
| hour12-default | `en-IN` 默认 12h vs 24h locale 边界 |
| cldr-version-pin | go-intl pin CLDR 48.1.0 vs fixture source data baseline |

---

## 4. CI Three Gates

`task verify` **必须**串成下述三 gate;每 gate 行为如表:

| Gate | 作用 | 触发命令 | PR 行为 | main 行为 |
|------|------|---------|---------|----------|
| **Gate 1: Conformance** | 所有非 divergence fixture 必过 | `task test`(`go test -race -p 1 ./...`) | block | block |
| **Gate 2: Divergence Audit** | divergences.md 与 fixture 一致;无 silently skip | `tools/check-divergences/` | block | block |
| **Gate 3: Performance** | hot-path benchmarks 不超过 SPEC 71 阈值 | SPEC 71 §3 详定 | warn(评论 PR) | block |

**必须遵守**:

1. Gate 1 + Gate 2 **必须**在所有 PR 上阻塞;**禁止** override(除非 maintainer 显式批准并修订 SPEC)。
2. Gate 3 在 PR 上 **必须**仅 warn(评论性能 diff);main 分支 nightly job 才阻塞。
3. `task verify` **必须**串行三 gate(失败短路);**禁止** Gate 1 失败但 Gate 2/3 仍报告(噪音掩盖)。
4. 三 gate 任一失败 **必须**返回非零 exit code,CI 据此 block。

> **Why**: 三 gate 把 "正确性 / spec 解释 / 性能" 三类回归分层管理。Gate 3 在 PR 仅 warn 是因为单 PR 性能微抖动 < 5% 是噪音,main nightly 用 benchstat 跑统计才有信号。
> **Rejected**: 三 gate 全部 PR 阻塞——会把所有 noisy benchmark 变成 PR 摩擦,反而退化为"reviewer 习惯性 force-merge",规则失效。

---

## 5. XFAIL Discipline

XFAIL(expected failure)= fixture 已知 fail 但被允许通过 CI。**必须**有过期日期。

### 5.1 XFAIL 登记

XFAIL 条目 **必须**位于 `<package>/testdata/xfail.json`,字段:

| 字段 | 含义 |
|------|------|
| `id` | fixture id |
| `reason` | 失败原因(missing implementation / upstream bug / 等待 CLDR 升级) |
| `expires_at` | ISO 8601 日期;过期后 CI **必须** fail this fixture |
| `tracking_issue` | go-intl issue URL |

**必须遵守**:

1. 每条 XFAIL **必须**有 `expires_at`;**禁止** 永久 XFAIL(用 divergences.md 替代)。
2. `expires_at` 默认 ≤ 90 天;延期 **必须** 经 PR 显式批准。
3. XFAIL 过期后 CI **必须**自动 fail——用 fixture runner 的"日期 > expires_at 强制断言"实现。
4. XFAIL **禁止** 用于"我懒得修"——只允许:upstream bug 待修(FormatJS/CLDR)、本仓 issue 已开但优先级低、依赖 SPEC 修订。

> **Why**: XFAIL 时效是反熵机制——没有过期日期的 XFAIL 会变成"永远不修的 TODO 注释",一年后 CI 满屏黄色。
> **Rejected**: 无限期 XFAIL——这会把临时豁免变成长期未验证行为。

### 5.2 XFAIL 与 divergence 区别

| 维度 | XFAIL | Divergence |
|------|-------|-----------|
| 性质 | 已知 bug / 未实现 | 已知行为差异 |
| 解决方式 | 修复后删除 | 长期共存 |
| 时效 | 必须有过期日期 | 永久(但需 review_after 锚点) |
| CI 行为 | 跳过断言 | 跳过断言 |
| 来源 | go-intl 自身代码缺陷 | spec 实现定义 / CLDR 版本差异 |

---

## 6. Forbidden

- **禁止** silently skip 未通过的 fixture(必须走 divergences.md 或 xfail.json,二者均需 PR 审批)。
- **禁止** silently skip 未抽取的 FormatJS source(必须写入 `.skip-list.json` 的 `source` + `category` + `reason`)。
- **禁止** fixture 无 `id` 或 `source` 字段。
- **禁止** 在同一 JSON 文件内混合 `formatjs:` / `node:` / `icu4j:` 来源。
- **禁止** 用 Unix epoch ms 数字作为 DateTime 输入(必须 ISO-8601 字符串)。
- **禁止** 删除 divergences.md 历史条目(过时条目改 `status: resolved` 保留)。
- **禁止** XFAIL 无 `expires_at`。
- **禁止** ICU4J 进入 CI 工具链(Java 开销与 ROI 不匹配)。
- **禁止** 引入 `stretchr/testify` 等断言库;测试栈限定 stdlib `testing` + `google/go-cmp`。
- **禁止** 引入 snapshot 框架(`gkampitakis/go-snaps` / `bradleyjkemp/cupaloy` 等)替代 fixture JSON——R09 §4.9 已论证 fixture JSON 足够。
- **禁止** 抽取脚本人工编辑产物(手动 fixture 走 `manual/` 目录)。
- **禁止** `.skip-list.json` 被 fixture runner 当成通过依据;它只审计 extraction coverage。
- **禁止** Gate 3 在 PR 阻塞(防 noisy 摩擦);**禁止** Gate 1/2 在 main 不阻塞(必须 block)。
- **禁止** 在 hot-path Go 代码中调用 fixture loader(loader 仅在 `_test.go` 文件)。

---

## 7. Acceptance Criteria

### Fixture Schema

- [ ] `<package>/testdata/conformance/<source>/*.json` 存在;每文件是 JSON 数组,符合 SPEC §1.1 schema。
- [ ] `tools/check-fixtures/` 校验 `id` 全局唯一;违反 block CI。
- [ ] 错误用例独立位于 `errors.json`,通过 `errors.Is` 校验。

### Fixture Sources

- [ ] `tools/gen-fixtures-from-formatjs/` 是独立 Go module;当前 generated gate 至少输出 NumberFormat format / parts / range,DateTimeFormat format / range,PluralRules select / selectRange,以及 Locale canonicalization fixtures。
- [ ] 根目录 `.skip-list.json` 存在,每条记录包含 `source`、`category` 与 `reason`,并覆盖无法机械抽取或超出当前 generated gate 的 FormatJS sources / partial sources;`accepted-divergence` 记录必须包含 `divergenceId`。
- [ ] formatjs 引用钉死在 `tools/.gen-versions`;CI 校验存在。

### Divergences

- [ ] 每 formatter 包 `<package>/testdata/divergences.md` 存在(可空)。
- [ ] `tools/check-divergences/` 校验 divergence `id` 在 fixture 中可定位;违反 block CI。
- [ ] divergences.md 修改 PR 在 GitHub 上要求 ≥ 1 maintainer review。

### CI Gates

- [ ] `task verify` 串行执行 Gate 1 → Gate 2 → Gate 3,失败短路。
- [ ] Gate 1 + Gate 2 在 PR 与 main 均 block。
- [ ] Gate 3 在 PR 仅 warn(评论);main nightly job block。
- [ ] `tools/gen-fixtures-from-formatjs/` 输出的 testdata 在对应 formatter 包通过 100%(已知 divergence 与 XFAIL 除外)。

### XFAIL

- [ ] `<package>/testdata/xfail.json` schema 校验通过(每条含 `id`、`reason`、`expires_at`、`tracking_issue`)。
- [ ] CI 在 `expires_at` 过期后自动 fail 对应 fixture。
- [ ] XFAIL 总数在 main 分支不增长(每月 review;active scope 接近完成时应 → 0)。

### SPEC / Code Drift Checklist

- [ ] 新增或修改任何 exported API 前,先在 ECMA-402 中定位 owner(`Intl` / `Intl.Locale` / `Intl.NumberFormat` / `Intl.DateTimeFormat` / `Intl.PluralRules`)或明确为 Go typed bridge;无 owner 则不得进入 public surface。
- [ ] 新增或修改 option / resolved option / part type / range source 时,同步核对 ECMA-402 字段名、允许值、默认值和错误边界。
- [ ] 本地 SPEC 与 ECMA-402 冲突时,先改 SPEC 再改代码或 fixture。
- [ ] FormatJS 中可机械抽取的引用用例必须进入 fixture;不可抽取的来源必须进入 `.skip-list.json` 并带 category。
- [ ] 已生成 fixture 失败时,只能通过实现修复、`testdata/divergences.md` 或 `testdata/xfail.json` 处理,不得移出 fixture 或写入 skip-list。

### 工具链约束

- [ ] `go.mod` 不含 `stretchr/testify`、`gkampitakis/go-snaps`、`bradleyjkemp/cupaloy`、`sebdah/goldie`。
- [ ] `go.mod` 测试依赖仅 `google/go-cmp`(SPEC 70 唯一 test-only 直接依赖)。

---

## References

- SPECS/00 §2.1(test fixture policy)、§2.2(reference hygiene)
- SPEC 60(Acceptance Criteria 引用本 SPEC fixtures)
- SPEC 71(Benchmark Strategy & Performance Gates,Gate 3 owner)
- `.research/R08-conformance-and-benchmarks.md` §1–§7
- `.research/R09-dependencies.md` §4.9(test-only deps)
- `.references/formatjs/packages/intl-numberformat/tests/`(主抽取来源)
- `.references/formatjs/packages/intl-datetimeformat/tests/`(主抽取来源)
- `.references/formatjs/packages/intl-pluralrules/tests/`(主抽取来源)
- `.references/formatjs/packages/intl-locale/tests/`(主抽取来源)
