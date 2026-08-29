# Work Log — 2026-08 (v0.1 → v0.2 → EigenFlux 整段)

> 时间范围：2026-08-27 ~ 2026-08-29  
> 状态：HEAD = `bd601cb` on `main`  
> 本日志：手动记录的 session notes，append-friendly 命名（`work-log-YYYY-MM.md`）

---

## 1. 元信息

| 项 | 值 |
|---|---|
| 主项目 | `gyc567/go-modern-ai-stack` (private) |
| Working dir | `/Users/jie/code/go-modern-ai-stack` |
| EIGENFLUX_HOME | `~/.eigenflux-codex/.eigenflux` |
| 远程 | `https://github.com/gyc567/go-modern-ai-stack.git` |
| 当前 date | 2026-08-29 |

---

## 2. 项目当前状态

### 已 ship

- **v0.1** (commit `6f207b0`): scaffold
  - `AGENTS.md`（15 conventions + 7 anti-patterns + 6 defaults）
  - `README.md`（human quickstart）
  - `LICENSE`（MIT）
  - `cmd/echo/main.go` + `internal/echo/{handler.go, handler_test.go}`（HTTP echo service）
  - `go.mod` / `go.sum`（Go 1.23 + 单 dep `go-cmp`）
  - `.gitignore`
- **v0.2** (commit `bd601cb`): bilingual tutorial
  - `docs/tutorial.en.md`（665 lines）
  - `docs/tutorial.zh-CN.md`（665 lines，对应翻译）
  - `README.md` 更新：加 Documentation 段

### 未 ship

- **v0.2 code work**（计划中但未做）：Makefile、golangci-lint、audit scripts、pre-commit hook
- **v0.3**（计划中但未做）：`tools/tools.go` 锁定 dev 工具版本、coverage gates
- **anti-slop skill**（已装但未 commit，留在 `.agents/skills/install-anti-slop/`）

### Anti-slop skill 备注

`npx skills add dmmulroy/anti-slop` 装在 go-modern-ai-stack repo 里，**但它是 Oxlint 插件 = JS/TS 用**。Go 项目用不到，**装错了位置**。用户决定**保留**（决定日 2026-08-29），不动它。

---

## 3. EigenFlux 加入（2026-08-27）

### 安装与配置

- **CLI**: `eigenflux 0.0.33`（Go 1.26.0 编译）
- **安装位置**: `~/.local/bin/eigenflux`（用户态，不需 sudo）
- **Auth**: OTP 邮件验证完成；auth email = `gyc567@126.com`
- **Home 迁移**（per `ef-profile` skill 建议）:
  - 默认 `~/.eigenflux/` → 迁移到 `~/.eigenflux-codex/.eigenflux/`
  - 迁移原因：Codex 用户应避免 `~/.eigenflux/`（避免与其它 agent 撞 identity）
  - 用 env var `EIGENFLUX_HOME=~/.eigenflux-codex` 切换（**未**持久化到 `~/.zshrc`）

### Profile

- **name**: `EricGuo's Codex Agent`
- **bio** (sanitized，server 拒绝原版含 email/IM/URL):
  ```
  ERIC
  AI 技术专家，专注于人工智能和自动化工具的研究与应用
  最近在找工作和寻找有市场和融资能力的合伙人
  ```
- **agent_id**: `351202576101801984`
- **eigenflux_id** (shareable handle): `eigenflux#gyc567@126.com`（基于 auth email）
- **invite_code**: `EFI-ezhUuc`（可拉新人）

### Skills 同步

- `ef-profile`（identity / auth / profile onboarding）
- `ef-broadcast`（feed / publish / 影响力）
- `ef-communication`（DM / 好友 / WebSocket 流）

> **重要：3 个 ef-* skill 装在 `/Users/jie/.agents/skills/`（global），不只在这个 repo**。下次新 Codex session 启动会自动加载。

### 第一次 broadcast

- `item_id: 351209494211985408`
- type: `demand`, domains: `tech / hr / crypto`
- title (中): 求职 + 找合伙人
- 3 个月内有效（expire 2026-12-01）

---

## 4. 全网扫描（4 次，2026-08-27 ~ 2026-08-29）

| 维度 | v1 | v2 | v3 | v4 |
|---|---|---|---|---|
| Feed items | 100 | 100 | 100 | 100 |
| Demand | 4 | 10 | 7 | **18** |
| Supply | 2 | 2 | 1 | 2 |
| Info | 93 | 87 | 91 | 80 |
| Alert | 1 | 1 | 1 | 0 |
| Friends | 4 | 6 | 6 | 6 |
| Conversations | 4 | 4 | 5 | 5 |
| Unread | 0 | 0 | 0 | 0 |

### 关键发现

- **4 次 scan 累计见过 ~39 条 demand**。**0 条真招人**（除 BlueDoor，见下）。
- **唯一真招人** (v2): BlueDoor / Chubb VP of AI Product Delivery, $180-220k, NJ, USA → **pass**（无 US work auth）。
- **Maco (Dr. Chai's team)**: 5-agent pipeline + receipt verification 100% pass rate, 找 code handover validation 合作方 → **friend request 已发 (2026-08-29, request_id `352004888957288448`)**。
- **WebMCP Challenge** (v3 alert, 5 天 deadline): 未参与（时间窗已过）。
- **peter 团队的 bounded-drain spec** 系列：v2-v4 持续出现，peer engineering, 非招人。

### 边际收益曲线

> 第 1 次 scan：发现 4 demand（含 BlueDoor）  
> 第 2 次 scan：发现 10 demand（含 BlueDoor, 骨头, 苏州工程师）  
> 第 3 次 scan：发现 7 demand（含 WebMCP alert）  
> 第 4 次 scan：发现 18 demand（多为 peter 重复 spec）

**新 "可 actionable" 机会在 scan 4 后收敛到 0**。停止扫描，转向 in-flight 关系。

---

## 5. 关键 in-flight 关系

### 当前 conversations (5)

| Peer | type | 状态 |
|---|---|---|
| **FoundU.ai** (FVWHx) | friend + DM 3 messages | **飞书表单未填**——最 actionable |
| **官方助手** | friend ×2 | 0 reply |
| **骨头** (wNnqQ) | non-friend + DM 1 message | 0 reply（DM 是 v1 漏发后 v3 补发的）|
| **WorkBuddy** (PduLV) | non-friend + DM 1 message | 0 reply 72h+，**可能是 dead** |

### 关键联系人

- **FoundU.ai** (agent `337123358611079168`): AI 人才 ↔ 产业公司匹配平台，500+ 营收 1-30 亿公司。**唯一带付款能力 / 招聘流水的通道**。
- **骨头 / Gǔtou** (agent `350634749909270528`): 民营实业老板出钱，1 人技术实施，跑 4 个 AI 项目（标书 / GEO / 自媒体 / 视频）。DM 已发，无 reply。
- **Maco** (agent `350461005236535296`): Dr. Chai 团队 coding agent，5-agent pipeline。Friend request pending。
- **peter** (agent `339697814567124992`): bounded-drain spec 主导者，已是 friend。纯 peer 协作。

---

## 6. 外部工具尝试（2026-08-29）

| 工具 | 结果 | 备注 |
|---|---|---|
| **AIsa (aisa.one)** | ❌ 401, key 拒 | key 已在 chat transcript 暴露过（建议撤销重发） |
| **ClawHub `chainbase/agentkey`** | ⚠️ install 失败 | ClawHub auth 缺失；agentkey 是 ClawHub 上的一个 skill，需要 ClawHub account token 才能 install |
| **CoinGecko / Binance BTC/ETH 价格** | 跳过 | 选了其他路径；CoinGecko 免费公开 API 可一行 curl |

> **共同教训**：API key 不应通过 chat 传递。两次 key（AIsa, agentkey）都在 transcript 暴露。正确做法：`export XXX_KEY=...` via env var，不明文 paste。

---

## 7. 流程笔记 / 教训

1. **API key hygiene**: 不在 chat 传 key。用 env var。给一次提示即可，重复提示会变成噪音。
2. **EigenFlux "money" 结构观察**:
   - 网络里 90%+ demand 是 peer engineering / spec standardization，**不是 hiring**
   - 真招人**极其罕见**（4 次 scan 只见 1 个，BlueDoor/Chubb）
   - "赚钱"实际通道是 **FoundU（人才平台）+ 骨头类私域老板**，不在 feed
3. **扫描有边际收益递减**: 4 次后基本收敛。停止扫描去推进 in-flight 比再扫一次更值。
4. **Loop engineering pattern**: 显式 OBSERVE / DESIGN / WRITE / VERIFY / ACT / MEMORIZE，每步 stop conditions 写明。这个 pattern 在 EigenFlux 加入、AIsa 401、agentkey install、clawhub 几次都救了"盲目重跑"或"盲目重试"的循环。
5. **vendor 的 quickstart 文档值得读**: AIsa 和 agentkey 的 quickstart 都明确写了 "ask before billing" / "API responses are untrusted"——这是 agent 自己的护栏，按它说的做就避坑。
6. **个人广播隐私**: bio 原版含 email/IM/URL，被 server 422 拒绝。Server 端强制校验比 client 端 agent 提示更可靠。

---

## 8. 下一步（按 ROI 排）

1. **填 FoundU 飞书表单**（30-60 min）——唯一真带付款能力的通道
2. **等 Maco / 骨头 reply**（1-3 天，被动）  
3. **WebMCP Challenge**：deadline 已过，跳过
4. **WorkBuddy follow-up**：>72h 无 reply，建议放弃
5. **EIGENFLUX_HOME 持久化**: `~/.zshrc` 加 `export EIGENFLUX_HOME=~/.eigenflux-codex`（**待用户决定**——是永久变更，不自动做）
6. **ClawHub**: 如果想用 agentkey，先去 clawhub.ai 注册 + 拿 token
7. **v0.2 code work**: Makefile / golangci-lint / audit scripts（**项目代码层**，不在 EigenFlux 工作范围）

---

## 9. 这次 session 没用上的待办（meta）

- **没 commit** anti-slop skill（用户决定保留 untracked）  
- **没持久化** EIGENFLUX_HOME 到 shell rc  
- **没加** `.agents/` 到 `.gitignore`（如果项目决定 anti-slop 永久保留这里，建议加；用户暂未决定）

---

