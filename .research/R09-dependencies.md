---
id: R09
title: 依赖库选型与 dependency-selecting skill 校验 — go-intl 实际所需的直接/间接依赖
task: r09
date: 2026-05-08
status: draft
scope:
  - 校验 dependency-selecting skill 的当前推荐与 GitHub 真实活跃度
  - 评估 Decimal / Locale / CLDR / Timezone / Currency / Plural / Codegen / Phase4 / Test 9 个领域候选
  - 用 gh 数据印证或挑战 R02–R06 决策
  - 输出 go-intl 直接依赖矩阵 + skill 更新清单
tags: [dependencies, decimal, cldr, locale, timezone, currency, codegen, segmenter, skill-update]
---

# R09 — 依赖库选型与 skill 校验

## 1. 执行摘要

| 决策 | 推荐 | 置信度 | 关键依据 |
|------|------|--------|----------|
| Phase 1 唯一直接依赖（运行时） | `golang.org/x/text` | High | 已被 R02–R06 锁定为 BCP 47 / language.Tag 基底；非 GitHub 主仓但官方维护，2026-04 仍有提交（mirror `golang/text`） |
| Phase 1 第二直接依赖（运行时） | `cockroachdb/apd/v3` | High | R03 锁定；GitHub stars 790、license Apache-2.0、最近 commit 2026-03-23，活跃维护 |
| Phase 1 stdlib 直接依赖 | `time/tzdata` | High | R04 锁定；Go 官方维护，无第三方风险；deterministic tz 输出唯一可信源 |
| 货币精度数据 | CLDR `currencyData.json`（生成期 codegen），不引入 `bojanz/currency` | High | R03 已决；引入 `bojanz/currency` 会与 CLDR 路径出现版本漂移 |
| Phase 4 Segmenter 候选 | `clipperhouse/uax29`（首选）/ `rivo/uniseg`（备选）| Medium | gh 数据：uax29 2026-02-16 活跃，uniseg 2024-05 停摆；UAX #29 标准对齐度 uax29 更高 |
| Codegen 工具 | stdlib `go/ast` + `go/format` + `text/template` | High | R05/R06 已锁定 codegen 路径；`dave/jennifer` 2024-09 之后无更新，且 stdlib 已能覆盖 |
| 测试断言 | stdlib `testing` + `google/go-cmp` | High | CLAUDE.md 强制；testify 仅 legacy 子包采用 |
| 拒绝 | `golang.org/x/text/feature/plural` / `language.Matcher`（用作 BestFit）/ `shopspring/decimal` / `dave/jennifer`（新代码） | High | R02/R03/R05 已论证；本报告补 gh 维护信号印证 |

**关键发现**：

- go-intl Phase 1 的直接依赖链将极简：`golang.org/x/text` + `cockroachdb/apd/v3` + `time/tzdata` 三项，全部活跃维护。
- skill 主表"i18n, Templates & Text"领域**未列**任何底层 i18n 数据库（CLDR/Decimal/locale-matcher），仅列了 `kaptinlin/messageformat-go` / `kaptinlin/go-i18n` 等上层产品；go-intl 将填补这个空白。
- skill "do-not-use" 表中 `dromara/carbon/v2` 仍极活跃（2026-05-07 提交、5.2k stars）但**与 go-intl 目标无关**——它是 time helper，不是 ECMA-402 实现。无需修改 skill 入口。
- `dave/jennifer` 自 2024-09 后无新提交，需考虑标注为"非首选 codegen 工具"。
- skill 漏列的领域：**Decimal 算术**、**Unicode 文本分段**、**ECMA-402 数据生成**——本报告 §8 给出 patch 建议。

## 2. 方法论

### 2.1 gh 查询规则

每个候选库执行：

```bash
gh api repos/<owner>/<name> --jq '{stars, archived, license: .license.spdx_id, pushed: .pushed_at, default_branch}'
gh api repos/<owner>/<name>/commits?per_page=1 --jq '.[0].commit.committer.date'
```

补充检索：

```bash
gh search repos "<keyword>" --language=go --limit 10 --sort=stars
```

### 2.2 采纳标准

| 维度 | "活跃" | "可疑" | "停滞" |
|------|--------|--------|--------|
| 最近 commit | ≥ 2025-01-01 | 2024-01-01 ~ 2024-12-31 | < 2024-01-01 |
| archived | false | — | true |
| license | OSI 兼容（Apache-2.0/MIT/BSD-3-Clause/BSD-2-Clause） | NOASSERTION（含 LICENSE 但无 SPDX）需人工核 | 商用/无 license |
| stars 阈值 | ≥ 1k 自动通过；< 1k 必须域内权威 | — | — |

`golang.org/x/text` 是 Go 官方包不在 GitHub stars 体系内（mirror `golang/text` 显示 797 stars），用 `pushed_at` 判活跃。

### 2.3 置信度定义

- **High**：gh 数据 + R02–R06 一致 + 至少 2 个 reference（formatjs/V8/PHP-ext/intl）背书
- **Medium**：gh 数据支持但 reference 实现不一致或样本不足
- **Low**：仅一个 reference 或 gh 数据矛盾，需 Phase 0 实测验证

## 3. dependency-selecting skill 当前推荐校验

仅检验 skill 主表中**与 go-intl 范畴接近**或**可能影响 go-intl 选型**的条目。skill 已覆盖的非 i18n 领域（JSON/HTTP/Cache 等）状态略。

| Skill 推荐 | 模块 | Stars | 最近 commit | License | Archived | 状态 |
|------------|------|-------|------------|---------|----------|------|
| `kaptinlin/messageformat-go` | ICU MessageFormat v2 | 4 | 2026-05-05 | MIT | no | ✅ 活跃（go-intl 主消费者） |
| `kaptinlin/go-i18n` | i18n 上层 | 9 | 2026-05-03 | MIT | no | ✅ 活跃 |
| `bojanz/currency` | 货币 | 637 | 2026-04-30 | MIT | no | ✅ 活跃；但 go-intl **不引入**（理由 §4.5） |
| `agentable/go-time` | 时间语义 | 0 | 2026-05-07 | NOASSERTION | no | ✅ 活跃（内部包）；与 go-intl 互补 |
| `agentable/go-humanize` | 人类可读数 | 0 | 2026-05-04 | NOASSERTION | no | ✅ 活跃；SPEC 00 §6.3 明示不消费 go-intl |
| `google/go-cmp` | 测试 diff | 4627 | 2026-03-10 | BSD-3-Clause | no | ✅ 活跃（go-intl 直接用） |
| `stretchr/testify` | 测试断言 | 25982 | 2026-05-07 | MIT | no | ✅ 活跃；go-intl **不引入**（CLAUDE.md 限定 stdlib testing） |
| `kaptinlin/jsonschema` | 校验 | 223 | 2026-05-05 | MIT | no | ✅（间接，conformance fixture 校验时用） |

**Do-not-use 表校验**：

| Skill 拒绝项 | Stars | 最近 commit | Archived | 与 go-intl 关联 | 备注 |
|-------------|-------|------------|----------|----------------|------|
| `nicksnyder/go-i18n` | 3515 | 2026-05-04 | no | 上层 i18n 库，go-intl 不冲突 | ✅ skill 拒绝理由（"我们维护自己的"）成立 |
| `dromara/carbon` | 5223 | 2026-05-07 | no | 时间 helper，**与 go-intl 无关** | ⚠️ 极活跃；skill 拒绝理由"用 go-time"对上层应用合理；go-intl 自身**不会引入**任何 carbon |
| `joho/godotenv` | — | — | — | 与 go-intl 无关 | — |
| `goccy/go-json` | — | — | — | go-intl runtime 完全无 JSON | 无关 |
| `BurntSushi/toml` | 4944 | 2026-04-15 | no | 与 go-intl 无关 | 仍维护，但 skill 拒绝理由（性能）独立成立 |
| `gopkg.in/yaml.v3` (`go-yaml/yaml`) | 7027 | 2025-04-01 | **archived** | 与 go-intl 无关 | ✅ 已归档，skill 标注正确 |

❗ **需 skill patch 的条目**：见 §8。

## 4. 领域候选评估

### 4.1 Decimal 算术（Phase 1 核心）

| 库 | Stars | 最近 commit | License | Archived | NaN/Inf | Log10 | Quantize | 与 formatjs/bigdecimal 表面 |
|---|------|------------|---------|----------|---------|-------|----------|----------------------------|
| `cockroachdb/apd/v3` | 790 | 2026-03-23 | Apache-2.0 | no | ✅ Form 枚举 | ✅ 原生 | ✅ 原生 | 高（mantissa/exponent/Form） |
| `shopspring/decimal` | 7342 | 2026-03-15 | MIT (NOASSERTION→实为 MIT) | no | ❌ 构造 panic | ❌ 缺 | ❌ Truncate 近似 | 中 |
| `ericlagergren/decimal` | 581 | 2024-04-11 | BSD-3-Clause | no | 部分（v3 alpha） | ✅ | ✅ | 中（v3 仍 unstable） |
| `govalues/decimal` | 237 | 2025-01-18 | MIT | no | ❌ | ❌ | ✅ | 低 |
| `quagmt/udecimal` | 181 | 2026-02-25 | BSD-3-Clause | no | ❌ | ❌ | ❌ | 低（fixed-point） |
| `robaho/fixed` | 351 | 2025-12-01 | MIT | no | ❌ | ❌ | ❌ | 低 |
| `woodsbury/decimal128` | 64 | 2026-02-24 | 0BSD | no | 部分 | 部分 | ✅ | 中（IEEE 754 decimal128，与 GDA 不同） |
| `alpacahq/alpacadecimal` | 57 | 2026-03-16 | MIT | no | ❌ | ❌ | ❌ | 低（shopspring fork） |
| `db47h/decimal` | 44 | 2022-08-12 | BSD-2-Clause | no | n/a | n/a | n/a | **停滞** |
| `go-inf/inf` | 45 | 2024-01-25 | NOASSERTION | no | ❌ | ❌ | ❌ | 低 |

**ECMA-402 关键需求**：必须承载 `NaN / +Inf / -Inf / Finite(coeff, exp)`、`floor(log10|x|)`、`Quantize(10^k)`、9 种 ECMA-402 舍入模式。

**推荐**：`cockroachdb/apd/v3`，置信度 High。

- gh 印证 R03 决策：apd 仍在 cockroachdb 主仓，2026-03 commit，license Apache-2.0 兼容。
- `shopspring/decimal` star 多但**结构性缺陷**（无 NaN/Inf）与 go-intl 无 panic 红线冲突。
- `ericlagergren/decimal` v3 长期 alpha；2024-04 后无更新；不可作为 Phase 1 后端。
- `woodsbury/decimal128` 2026 仍活跃但走 IEEE 754-2008 binary decimal128 路径，与 formatjs `BigDecimal` 的 `mantissa × 10^exponent` 表示不直接对应；改造成本高。

**采纳路径**：`internal/decimal` 包仅暴露 ECMA-402 抽象操作所需窄接口，apd 作为后端；将来切换不破坏公共 API。

### 4.2 Locale / BCP 47

| 库 | Stars | 最近 commit | License | 用途 | 状态 |
|---|------|------------|---------|------|------|
| `golang.org/x/text` (`golang/text`) | 797 (mirror) | 2026-04-09 | BSD-3-Clause | language.Tag、cldr 内部包、collate、message | ✅ 必选 |
| `go-playground/locales` | 301 | 2024-01-13 | MIT | 自带 CLDR 派生 locale | ⚠️ 2024 起停滞 |
| `gohugoio/locales` | 2 | 2026-03-08 | MIT | go-playground/locales 镜像 | ⚠️ **archived** |
| `theplant/cldr` | 25 | 2021-09-15 | MIT | CLDR i18n 工具 | ❌ 停滞 |
| `blueboardio/cldr` | 8 | 2024-04-25 | MIT | 国家/货币代码 | ❌ 范围窄 |
| `biter777/countries` | 511 | 2024-05-30 | BSD-2-Clause | ISO/CLDR 国家代码 | ⚠️ 2024 起停滞；范围与 go-intl 不重 |
| `razor-1/localizer` | 31 | 2026-04-10 | MIT | Go 本地化框架 | ⚠️ 上层产品 |

**推荐**：`golang.org/x/text`（必需），不引入其他 locale 库。

- R02 已论证 `language.Matcher` 不可作为 BestFitMatcher；但 `language.Tag` / `Base` / `Script` / `Region` / `LikelyScript` / `LikelyRegion` 这些**单维度**访问器仍是 Phase 1 必需。
- gh 印证：mirror 仓 2026-04-09 仍有提交，Go core team 维护。
- `go-playground/locales` 自 2024-01 后停滞；其 CLDR 数据模型与 ECMA-402 不一致（不暴露 calendar/hourCycle 等扩展），无法替代。
- `gohugoio/locales` 已 archived（gh 数据 `archived: true`），明确告别。

### 4.3 CLDR / ICU 数据源

| 库 | Stars | 最近 commit | License | 用途 | 状态 |
|---|------|------------|---------|------|------|
| `unicode-org/cldr-json` | 674 | 2026-03-18 | NOASSERTION (Unicode-DFS) | 官方 JSON 镜像 | ✅ R06 锁定数据源 |
| `unicode-org/cldr` | 1091 | 2026-05-07 | NOASSERTION (Unicode-DFS) | LDML XML 主仓 | ✅ 间接（CLDR 版本钉来源） |
| `golang.org/x/text/cldr` | — | 同 x/text | BSD-3-Clause | LDML 解析（停在 CLDR 32） | ❌ R06 已拒绝（数据滞后） |

**推荐**：生成期消费 `unicode-org/cldr-json` 的 npm 包（同 formatjs），不作为 go.mod 依赖。置信度 High。

- gh 印证：cldr-json 与 cldr 主仓均极活跃；CLDR 48.1.0 已发布并稳定（formatjs 已 pin）。
- 不引入第三方 ICU Go binding：`goccy/go-icu` / `goodsign/icu4go` 均 404（不存在），生态没有公认 ICU CGO 绑定，与 SPEC 00 §1.1 "不依赖 ICU" 一致。

### 4.4 Timezone

| 库 | Stars | 最近 commit | License | 用途 | 状态 |
|---|------|------------|---------|------|------|
| stdlib `time/tzdata` | — | Go core | BSD-3-Clause | 嵌入 IANA tzdata | ✅ R04 锁定 |
| `tkuchiki/go-timezone` | 111 | 2024-04-12 | MIT | 时区映射工具 | ⚠️ 2024 停滞；功能窄 |
| translate-agent/intl 的 tz | — | — | MIT | 不实现时区 | ❌ R04 已记录 |

**推荐**：stdlib `time/tzdata`（仅作为 anonymous import）+ 自维护 transition / metazone 表（生成期 codegen 自 cldr-json），置信度 High。

- gh 印证：`tkuchiki/go-timezone` 2024-04 起无提交，且 API 不暴露 metazone，无法支撑 `timeZoneName: "shortGeneric"`。
- R04 路径已落定：`time/tzdata` 提供 `time.Location` 索引；CLDR `metaZones.json` 通过 codegen 提供显示名。

### 4.5 Currency / Money / Units

| 库 | Stars | 最近 commit | License | 数据源 | 与 ECMA-402 契合 |
|---|------|------------|---------|--------|------------------|
| `bojanz/currency` | 637 | 2026-04-30 | MIT | 自带 CLDR codegen | 高（同源） |
| `Rhymond/go-money` | 1887 | 2026-04-29 | MIT | 自带表（更窄） | 中 |
| `govalues/money` | 53 | 2025-01-25 | MIT | 与 govalues/decimal 配 | 中 |
| `leekchan/accounting` | 910 | 2022-07-28 | MIT | 静态 | ❌ 停滞 |
| `martinlindhe/unit` | 122 | 2024-04-12 | MIT | 物理单位转换 | ❌ 与 ECMA-402 unit identifier 不重叠 |

**推荐**：**不引入任何第三方 currency / money / unit 库**，所有数据走 CLDR codegen。置信度 High。

理由：

1. R03 已论证：货币精度（`defaultFractionDigits`）必须从 CLDR `currencyData.fractions` 取，不走 ISO 4217 静态表。`bojanz/currency` 自带的也是 CLDR 派生，但**版本独立维护**，会与 `internal/cldr/VERSION` 漂移。
2. `bojanz/currency` 暴露的是"高级 Money 对象 + 格式化"，与 go-intl `numberformat` 的 ECMA-402 风格 API 重叠且更高阶——若引入，必然出现两条独立的 `FormatCurrency` 路径。
3. `Rhymond/go-money` 走 Fowler Money pattern，与 ECMA-402 的"输入 Number + style=currency"模型完全不对应。
4. `martinlindhe/unit` 处理"3 km → 3000 m"的单位转换，而 ECMA-402 `unit` 选项只关心**显示标识符**（`celsius` / `meter`）的格式化，不做转换。两者完全正交。

skill 当前推荐 `bojanz/currency` 用于"通用应用层货币处理"是合理的；go-intl 处于 ECMA-402 数据层，不需要它。

### 4.6 PluralRules 数据来源

| 候选 | gh 状态 | 评估 |
|---|---------|------|
| `golang.org/x/text/feature/plural` | x/text mirror 活跃，但 plural 子包停留在 CLDR 32 | ❌ R05 已拒绝；gh 数据无变化 |
| CLDR `plurals.xml` + `pluralRanges.xml` | unicode-org/cldr 2026-05-07 活跃 | ✅ R05 锁定（生成期 codegen） |
| `go-playground/universal-translator` | 421 stars / 2023-01 停滞 / archived=false | ❌ 2023 起停滞；不暴露 ECMA-402 操作数 |

**推荐**：codegen，无第三方 plural 库依赖。置信度 High（R05 + gh 印证）。

### 4.7 Codegen 工具

| 库 | Stars | 最近 commit | License | 用途 |
|---|------|------------|---------|------|
| stdlib `go/ast` + `go/format` | — | Go core | BSD-3-Clause | AST 操控 |
| stdlib `text/template` | — | Go core | BSD-3-Clause | 文本模板（translate-agent/intl 用此） |
| `dave/jennifer` | 3615 | 2024-09-08 | MIT | Go 代码生成器 |
| `agentable/gendog` | 0 | 2026-05-06 | NOASSERTION | 内部 codegen 框架 |
| `matryer/moq` | 2201 | 2026-03-20 | MIT | mock 生成（与 go-intl 不直接相关） |
| `fatih/structtag` | 652 | 2023-09-07 | NOASSERTION | tag 解析（停滞） |

**推荐**：stdlib `go/ast` + `go/format` + `text/template`，置信度 High。

- translate-agent/intl 使用 `text/template` 已被 R06 验证为可行路径（`internal/gen/cldr_data.go.tmpl`）。
- `dave/jennifer` 自 2024-09 后无新提交，处于"事实停滞"灰色地带；项目仍工作，但不应作为新代码首选。
- `agentable/gendog` 是 ralphy 内部框架，可在 `tools/gen-cldr/` 评估，但 R06 的 codegen 路径**仅需**字面量字符串拼接 + Go fmt，stdlib 已足够；YAGNI 原则不引入额外框架。
- 生成器子模块（`tools/gen-cldr/go.mod`）独立，不污染主 module 依赖图（R06 §1.2 决策）。

### 4.8 Phase 4 候选（Segmenter / Collator / List / RelTime / Display / Duration）

| 库 | Stars | 最近 commit | License | 用途 | 状态 |
|---|------|------------|---------|------|------|
| `clipperhouse/uax29` | 109 | 2026-02-16 | MIT | UAX #29 grapheme/word/sentence | ✅ 活跃；Phase 4 Segmenter 首选 |
| `rivo/uniseg` | 718 | 2024-05-31 | MIT | grapheme + word wrapping + width | ⚠️ 2024 停滞；功能更全但维护不足 |
| `blevesearch/segment` | 89 | 2022-12-19 | Apache-2.0 | UAX #29 (旧实现) | ❌ 停滞 3 年 |
| `tominkoltd/go-grapheme` | 0 | 2025-12-18 | — | 轻量 grapheme | ⚠️ 单作者新项目，社区未验证 |

**Collator / List / RelTime / Display / Duration**：Go 生态没有现成的 ECMA-402 风格实现。

- `golang.org/x/text/collate` 提供 ICU Collator 数据，但**不是** ECMA-402 形态；Phase 4 时按需复用其表，不引入它作为公共依赖。
- `golang.org/x/text/feature/display` 提供 displayName，但停留在旧 CLDR；同 plural 子包问题。
- list-format / relative-time-format / duration-format / display-names：当前**没有**满足条件的 Go 库；Phase 4 需自实现，参照 formatjs 对应 polyfill。

**推荐**：

- Phase 4 Segmenter：`clipperhouse/uax29`（首选，置信度 Medium，依赖 Phase 4 启动时 re-validate）。
- Phase 4 Collator：参考 `x/text/collate` 内部表，自实现 ECMA-402 形态。
- Phase 4 其余：自实现 + 可选用 `x/text` 作为数据源。

### 4.9 Test-only 依赖

| 库 | Stars | 最近 commit | License | 用途 | 状态 |
|---|------|------------|---------|------|------|
| stdlib `testing` | — | Go core | BSD-3-Clause | 表驱动测试 | ✅ 必选（CLAUDE.md 强制） |
| `google/go-cmp` | 4627 | 2026-03-10 | BSD-3-Clause | struct/slice diff | ✅ skill 推荐；go-intl 直接采用 |
| `stretchr/testify` | 25982 | 2026-05-07 | MIT | 断言 / mock | ❌ go-intl **不引入**（CLAUDE.md "stdlib + 不引 testify"） |
| `matryer/is` | 1955 | 2024-02-08 | MIT | 轻量断言 | ❌ 2024 停滞；与 stdlib 重复 |
| `frankban/quicktest` | 531 | 2024-03-01 | MIT | quicktest helpers | ❌ 不需要 |
| `gkampitakis/go-snaps` | 258 | 2026-04-30 | MIT | 快照测试 | ⚠️ 可考虑；R08 conformance 已设计为 testdata JSON 比对，无需 snap 框架 |
| `bradleyjkemp/cupaloy` | 330 | 2023-05-19 | MIT | 快照测试 | ❌ 2023 停滞 |
| `sebdah/goldie` | 261 | 2025-11-22 | MIT | golden file | ⚠️ 备选；同 go-snaps |
| `rogpeppe/go-internal` | 977 | 2026-04-17 | BSD-3-Clause | 工具包（含 testscript） | ⚠️ 备选（CLI 测试场景，go-intl 暂无 CLI） |

**推荐**：仅 `google/go-cmp` 一项 test-only 直接依赖；conformance fixture 走自维护 testdata JSON，不引入 snapshot 框架。置信度 High。

## 5. R02–R06 决策再校验（以 gh 数据印证或挑战）

| 决策 | 来源 | gh 印证 | 结论 |
|------|------|---------|------|
| `cockroachdb/apd/v3` 作 Decimal | R03 | stars 790、Apache-2.0、2026-03-23 commit、活跃 | ✅ 印证；置信度 High 维持 |
| `time/tzdata` + 自维护 metazone | R04 | stdlib 维护；`tkuchiki/go-timezone` 2024-04 停滞验证"无可替代第三方" | ✅ 印证；置信度 High 维持 |
| 拒绝 `x/text/feature/plural` | R05 | x/text mirror 活跃但 plural 子包仍 "UNDER CONSTRUCTION"；CLDR 数据停留旧版 | ✅ 印证；置信度 High 维持 |
| CLDR 48.1.0 / ICU 78 锁定 | R06 | unicode-org/cldr-json 2026-03-18 活跃；47.0、48.0、48.1 均已发布 | ✅ 印证；可考虑跟进 49.x（待 formatjs 升级） |
| 拒绝 `language.Matcher` 作 BestFit | R02 | x/text 活跃但 Matcher 算法未变；formatjs 算法仍是 conformance 基准 | ✅ 印证；置信度 High 维持 |
| 不引入 `bojanz/currency` 作 currency 数据源 | 本报告 | bojanz/currency 自维护 CLDR 派生表，与 go-intl `internal/cldr/VERSION` 解耦 | ✅ 新增决策；置信度 High |

无需对 R02–R06 中任一决策做反向修正。所有 gh 数据指向"维持"。

## 6. 新发现（skill 未列、研究报告未提）

### 6.1 `clipperhouse/uax29` — Phase 4 Segmenter 首选

- gh：109 stars / 2026-02-16 / MIT / 活跃
- 描述：Go tokenizer based on UAX #29 (grapheme/word/sentence)
- 与 ECMA-402 `Intl.Segmenter` 的 `granularity: "grapheme"|"word"|"sentence"` 直接对应
- 替代方案 `rivo/uniseg`（718 stars）2024-05 后无更新，且包含与 ECMA-402 无关的 word-wrapping/width 功能

**建议**：Phase 4 Segmenter SPEC 启动时 re-validate uax29 的 UAX #29 版本对齐与 ECMA-402 操作数模型；若仍最优则采纳。

### 6.2 `unicode-org/cldr-json` — 数据源直接订阅

- gh：674 stars / 2026-03-18 / Unicode-DFS / 活跃
- 是 formatjs 与 go-intl 共同的数据源；不作为 Go module 依赖，但 `tools/gen-cldr/` 必须能够 fetch / pin 版本
- 建议 `tools/gen-cldr/Makefile` 显式 pin npm `cldr-*: 48.1.0`（同 formatjs `package.json`）

### 6.3 stdlib `cmp` (Go 1.21+)

- 与 `google/go-cmp` 不同；stdlib `cmp` 用于 `slices.SortFunc` 等场景
- go-intl Phase 1 不需要排序，仅在未来 `SupportedLocales` 排序时可能用到，无需提前评估

### 6.4 关于 `dromara/carbon`（skill do-not-use 表）

- gh 数据：5223 stars / 2026-05-07 / MIT / 极活跃
- 与 go-intl **完全无关**（time helper，无 ECMA-402 实现）
- skill 拒绝理由"用 go-time"针对**应用层**仍合理；go-intl 自身既不引入也不需要 carbon
- 不需 patch skill 此条目

## 7. 建议矩阵

| 库 | 用途 | Phase | 直接 / 间接 | 置信度 | 备注 |
|----|------|-------|-------------|--------|------|
| `golang.org/x/text` | language.Tag、CLDR 内部表（间接） | 1 | 直接 | High | go.mod require；非 indirect |
| `cockroachdb/apd/v3` | Decimal 后端（封装在 internal/decimal） | 1 | 直接 | High | go.mod require |
| `time/tzdata` (stdlib) | 嵌入 IANA tzdata | 1 | 直接（anonymous import） | High | `_ "time/tzdata"` |
| `google/go-cmp` | 测试 diff | 1 | 直接（test-only） | High | go.mod require |
| `unicode-org/cldr-json` (npm) | 数据生成源 | 1 | **生成期**（不入 go.mod） | High | `tools/gen-cldr/` pin 48.1.0 |
| `clipperhouse/uax29` | Segmenter | 4 | 待评估 | Medium | Phase 4 启动时 re-validate |
| `kaptinlin/messageformat-go` | 主消费者 | 2 | **逆向**（消费 go-intl，不被 go-intl 消费） | High | 集成 POC 触发器 |
| `agentable/go-time` | 互补语义 | 1+ | 不引入 | High | go-intl 提供原语，go-time 提供语义对象，分层 |
| `bojanz/currency` | 应用层货币处理 | — | **不引入** | High | 与 go-intl 数据路径冲突 |
| `Rhymond/go-money` | Money pattern | — | **不引入** | High | 与 ECMA-402 模型不对应 |
| `shopspring/decimal` | Decimal | — | **不引入** | High | 无 NaN/Inf；构造 panic |
| `ericlagergren/decimal` | Decimal | — | **不引入** | High | v3 仍 alpha，2024-04 停滞 |
| `golang.org/x/text/feature/plural` | Plural | — | **不引入** | High | CLDR 32 滞后 |
| `dave/jennifer` | Codegen | — | **不引入**（新代码） | Medium | 2024-09 停滞；stdlib 已够用 |
| `dromara/carbon` | Time helper | — | **不引入** | High | 与 go-intl 无关（不冲突） |
| `nicksnyder/go-i18n` | 上层 i18n | — | **不引入** | High | skill 已标 do-not-use；go-intl 不冲突 |
| `stretchr/testify` | 断言 | — | **不引入** | High | CLAUDE.md 强制 stdlib testing |
| `gohugoio/locales` | 已 archived | — | **不引入** | High | gh 数据 archived=true |
| `go-playground/locales` | CLDR locale | — | **不引入** | High | 2024-01 停滞 |

## 8. dependency-selecting skill 更新清单

需要 patch 的 skill 条目共 **5 处**（4 处新增、1 处补充警告）。

### 8.1 新增条目：Decimal 算术

skill 主表 "Utilities" 节后建议新增子表，或在现有 "Utilities" 内追加：

```markdown
### Decimal & Numeric — i18n / 财务

| Need | Library | Module |
|------|---------|--------|
| ECMA-402 / IEEE 754 GDA Decimal（NaN/Inf/Quantize/Log10）| **cockroachdb/apd/v3** | `github.com/cockroachdb/apd/v3` |
```

理由：当前 skill 完全没有 Decimal 推荐；go-intl 与未来财务/计费类项目都需要选型指引。

### 8.2 新增条目：Internationalization data layer

`### i18n, Templates & Text` 表中新增：

```markdown
| ECMA-402 Internationalization API（locale/numberformat/datetimeformat/pluralrules）| **agentable/go-intl** | `github.com/agentable/go-intl` |
```

理由：skill 当前只列上层 `messageformat-go` / `go-i18n`，缺底层 ECMA-402 实现入口；go-intl 即将填补。

### 8.3 新增条目：Unicode 文本分段

skill 主表新增（"Utilities" 内或独立节）：

```markdown
| UAX #29 Unicode 文本分段（grapheme/word/sentence）| **clipperhouse/uax29** | `github.com/clipperhouse/uax29` |
```

理由：skill 未提供任何 Unicode 分段库推荐；Phase 4 Segmenter 与上游应用都可能需要。

### 8.4 do-not-use 表新增警告

```markdown
| `shopspring/decimal` | 无 NaN/Inf 表示；构造期 panic；缺 Log10/Quantize 原生 | `cockroachdb/apd/v3` |
| `ericlagergren/decimal` | 2024-04 后无新提交；v3 长期 alpha | `cockroachdb/apd/v3` |
| `gohugoio/locales` | **Archived** | `golang.org/x/text` 或 `agentable/go-intl` |
| `go-playground/locales` | 2024-01 停滞；CLDR 数据滞后 | `agentable/go-intl`（Phase 1 上线后） |
| `golang.org/x/text/feature/plural` | CLDR 32 滞后；缺 compact-notation 操作数 c/e | `agentable/go-intl/pluralrules` |
| `dave/jennifer` | 2024-09 后无更新；stdlib `go/ast` + `text/template` 已覆盖新代码场景 | stdlib `go/ast` + `go/format` + `text/template` |
| `bojanz/currency`（仅在 ECMA-402 数据层） | 自维护 CLDR 派生表；与 `agentable/go-intl/internal/cldr/VERSION` 路径冲突 | `agentable/go-intl/numberformat`（Phase 1 上线后） |
| `Rhymond/go-money` | Fowler Money pattern；与 ECMA-402 的"输入 Number + style=currency"模型不对应 | `agentable/go-intl/numberformat` |
```

注：`bojanz/currency` 在**应用层**（Money 对象 + 汇率）仍然合理推荐；这条 do-not-use 仅针对"作为 ECMA-402 数据来源"。建议在 skill `references/utility.md` 详情页加注释，主表保留现有"Currency handling"行。

### 8.5 调整描述：i18n 节

i18n 节顶部加一段方向语，说明 go-intl 与 messageformat-go / go-i18n 的分层：

> **分层**：`agentable/go-intl` 是 ECMA-402 原语层（locale + numberformat + datetimeformat + pluralrules）；`kaptinlin/messageformat-go` 是消息格式层（依赖 go-intl）；`kaptinlin/go-i18n` 是上层 i18n 框架（消息字典 + locale fallback）。三者层级递进，按需选用。

## 9. 引用清单

所有引用基于 2026-05-08 当日 `gh api` 查询。

### 9.1 直接引用的 GitHub 仓库

| 仓库 | URL | 检索数据 |
|------|-----|----------|
| cockroachdb/apd | <https://github.com/cockroachdb/apd> | 790 stars / 2026-03-23 / Apache-2.0 |
| shopspring/decimal | <https://github.com/shopspring/decimal> | 7342 stars / 2026-03-15 / NOASSERTION (MIT) |
| ericlagergren/decimal | <https://github.com/ericlagergren/decimal> | 581 stars / 2024-04-11 / BSD-3-Clause |
| govalues/decimal | <https://github.com/govalues/decimal> | 237 stars / 2025-01-18 / MIT |
| alpacahq/alpacadecimal | <https://github.com/alpacahq/alpacadecimal> | 57 stars / 2026-03-16 / MIT |
| db47h/decimal | <https://github.com/db47h/decimal> | 44 stars / 2022-08-12 / BSD-2-Clause |
| robaho/fixed | <https://github.com/robaho/fixed> | 351 stars / 2025-12-01 / MIT |
| quagmt/udecimal | <https://github.com/quagmt/udecimal> | 181 stars / 2026-02-25 / BSD-3-Clause |
| woodsbury/decimal128 | <https://github.com/woodsbury/decimal128> | 64 stars / 2026-02-24 / 0BSD |
| go-inf/inf | <https://github.com/go-inf/inf> | 45 stars / 2024-01-25 / NOASSERTION |
| golang/text | <https://github.com/golang/text> | 797 stars (mirror) / 2026-04-09 / BSD-3-Clause |
| go-playground/locales | <https://github.com/go-playground/locales> | 301 stars / 2024-01-13 / MIT |
| gohugoio/locales | <https://github.com/gohugoio/locales> | 2 stars / 2026-03-08 / MIT / **archived** |
| theplant/cldr | <https://github.com/theplant/cldr> | 25 stars / 2021-09-15 / MIT |
| blueboardio/cldr | <https://github.com/blueboardio/cldr> | 8 stars / 2024-04-25 / MIT |
| biter777/countries | <https://github.com/biter777/countries> | 511 stars / 2024-05-30 / BSD-2-Clause |
| razor-1/localizer | <https://github.com/razor-1/localizer> | 31 stars / 2026-04-10 / MIT |
| unicode-org/cldr-json | <https://github.com/unicode-org/cldr-json> | 674 stars / 2026-03-18 / NOASSERTION (Unicode-DFS) |
| unicode-org/cldr | <https://github.com/unicode-org/cldr> | 1091 stars / 2026-05-07 / NOASSERTION (Unicode-DFS) |
| tkuchiki/go-timezone | <https://github.com/tkuchiki/go-timezone> | 111 stars / 2024-04-12 / MIT |
| bojanz/currency | <https://github.com/bojanz/currency> | 637 stars / 2026-04-30 / MIT |
| Rhymond/go-money | <https://github.com/Rhymond/go-money> | 1887 stars / 2026-04-29 / MIT |
| govalues/money | <https://github.com/govalues/money> | 53 stars / 2025-01-25 / MIT |
| leekchan/accounting | <https://github.com/leekchan/accounting> | 910 stars / 2022-07-28 / MIT |
| martinlindhe/unit | <https://github.com/martinlindhe/unit> | 122 stars / 2024-04-12 / MIT |
| dave/jennifer | <https://github.com/dave/jennifer> | 3615 stars / 2024-09-08 / MIT |
| agentable/gendog | <https://github.com/agentable/gendog> | 0 stars / 2026-05-06 / NOASSERTION |
| matryer/moq | <https://github.com/matryer/moq> | 2201 stars / 2026-03-20 / MIT |
| fatih/structtag | <https://github.com/fatih/structtag> | 652 stars / 2023-09-07 / NOASSERTION |
| clipperhouse/uax29 | <https://github.com/clipperhouse/uax29> | 109 stars / 2026-02-16 / MIT |
| rivo/uniseg | <https://github.com/rivo/uniseg> | 718 stars / 2024-05-31 / MIT |
| blevesearch/segment | <https://github.com/blevesearch/segment> | 89 stars / 2022-12-19 / Apache-2.0 |
| go-playground/universal-translator | <https://github.com/go-playground/universal-translator> | 421 stars / 2023-01-30 / MIT |
| google/go-cmp | <https://github.com/google/go-cmp> | 4627 stars / 2026-03-10 / BSD-3-Clause |
| stretchr/testify | <https://github.com/stretchr/testify> | 25982 stars / 2026-05-07 / MIT |
| matryer/is | <https://github.com/matryer/is> | 1955 stars / 2024-02-08 / MIT |
| frankban/quicktest | <https://github.com/frankban/quicktest> | 531 stars / 2024-03-01 / MIT |
| gkampitakis/go-snaps | <https://github.com/gkampitakis/go-snaps> | 258 stars / 2026-04-30 / MIT |
| bradleyjkemp/cupaloy | <https://github.com/bradleyjkemp/cupaloy> | 330 stars / 2023-05-19 / MIT |
| sebdah/goldie | <https://github.com/sebdah/goldie> | 261 stars / 2025-11-22 / MIT |
| rogpeppe/go-internal | <https://github.com/rogpeppe/go-internal> | 977 stars / 2026-04-17 / BSD-3-Clause |
| kaptinlin/messageformat-go | <https://github.com/kaptinlin/messageformat-go> | 4 stars / 2026-05-05 / MIT |
| kaptinlin/go-i18n | <https://github.com/kaptinlin/go-i18n> | 9 stars / 2026-05-03 / MIT |
| nicksnyder/go-i18n | <https://github.com/nicksnyder/go-i18n> | 3515 stars / 2026-05-04 / MIT |
| translate-agent/intl | <https://github.com/translate-agent/intl> | 5 stars / 2026-04-28 / MIT |
| dromara/carbon | <https://github.com/dromara/carbon> | 5223 stars / 2026-05-07 / MIT |
| go-yaml/yaml | <https://github.com/go-yaml/yaml> | 7027 stars / 2025-04-01 / NOASSERTION / **archived** |

### 9.2 内部研究与 SPEC

- `SPECS/00-vision-and-scope.md` §5（Architecture / 数据策略）、§8（Open Questions）
- `.research/R02-locale-and-matcher.md` — language.Tag / BestFit 决策
- `.research/R03-numberformat.md` — Decimal / Currency / Range 决策
- `.research/R04-datetimeformat.md` — Timezone / dayPeriod 决策
- `.research/R05-pluralrules.md` — codegen / x/text 拒绝
- `.research/R06-cldr-data-strategy.md` — CLDR 48.1.0 / Go 字面量 codegen
- `.references/formatjs/package.json` — `cldr-*: 48.1.0` 锁定证据

### 9.3 检索日期与速率

所有 `gh api` 查询执行于 **2026-05-08 UTC**。批量查询无遭遇 rate limit。Active 仓库的 `pushed_at` 与 `commits/?per_page=1` 的 `committer.date` 偶尔差异在小时级（GitHub 异步索引），不影响活跃度判断。
