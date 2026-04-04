# 对照端 对标 — 迭代轮次 01

**日期**: 2026-03-30  
**依据流程**: [`功能对标流程.md`](./功能对标流程.md)  
**总览路线图**: [`development-roadmap-parity.md`](./development-roadmap-parity.md)  
**差异基线**: [`orchestra-vs-reference-parity-analysis.md`](./orchestra-vs-reference-parity-analysis.md) v2.0

---

## 1. 本轮范围

在 **不改动** 服务端工作区模型、REST/WS 契约的前提下，对齐 对照端 习惯的 **i18n 覆盖面**（成员管理页、登录页为入口），并补齐 **可重复执行的 E2E 登录验收**，便于后续轮次在稳定基线上继续对标。

---

## 2. 当前 Orchestra 状态简述

- v2.0 周期已交付：聊天/终端核心、设置侧 i18n 骨架、`vue-i18n` 与 `locale` 联动。
- **Members**、**Login** 仍为大量硬编码英文，与 对照端 多语言产品形态不一致。
- E2E 仅有 `/login` 可见性与 `/health`，缺少「占位认证 → 工作区列表路由」的回归。

---

## 3. 差距列表

| # | 对照端 期望 | Orchestra 现状 | 保留差异？ |
|---|--------------|------------------|------------|
| G1 | 成员列表、邀请入口等多语言 | `MembersPage` 英文硬编码 | 否 — 本轮用 i18n key 对齐 |
| G2 | 登录与错误提示可本地化 | `LoginPage` 英文硬编码 | 否 — 本轮用 i18n key 对齐 |
| G3 | 关键用户路径有自动化回归 | 无登录成功路径 E2E | 否 — 本轮增加 Playwright 用例 |
| G4 | Tauri 级全局未读事件流 | 仍无 WS 推送未读 | **是** — Web 架构保留 HTTP 为主，后续轮次再议 SSE/WS |

---

## 4. 本轮可交付项

- [ ] **R01-1**：扩展 `src/i18n/locales/{en,zh}.json`，覆盖 `MembersPage` 主文案（标题、按钮、加载/空态、角色分区标题）。
- [ ] **R01-2**：`LoginPage.vue` 使用 `vue-i18n`（标题、表单标签、按钮、错误、默认账号提示）。
- [ ] **R01-3**：新增 `frontend/e2e/auth-flow.spec.ts`：使用默认账号 `orchestra` / `orchestra` 登录后 URL 含 `/workspaces`。
- [ ] **R01-4**：总览 [`development-roadmap-parity.md`](./development-roadmap-parity.md) 增加指向本文件的「轮次索引」链接。

---

## 5. 后续轮次 backlog（不在本轮）

- `MemberRow` / `AddMemberModal` / `EditMemberModal` 全量 i18n。
- 未读/消息状态的 Web 推送（SSE 或 WS）与 `applyUnreadSync` 自动订阅。
- `pendingTerminalMessages` / `readThrough` 与 对照端 深度对齐（REQ-206）。

---

## 6. 验证方式

- `cd frontend && pnpm exec vue-tsc --noEmit`
- `cd frontend && pnpm run build && pnpm exec playwright test e2e/auth-flow.spec.ts e2e/smoke.spec.ts`：`auth-flow` 必过；`smoke` 中 **backend health** 在本地未起 Go 服务时会 **skip**（非失败）。
- 手工：切换设置语言为中文，打开 Members 与 Login（需临时退出登录）核对文案。

---

## 7. 本轮结论（Step 3 填写）

- [x] R01-1 — Members 页 i18n
- [x] R01-2 — Login 页 i18n
- [x] R01-3 — E2E 登录流
- [x] R01-4 — 总览链轮次 01

**判定**: 本轮目标范围内 **已对齐**。Git 提交信息建议：`feat(i18n): members and login locales; test(e2e): auth flow to workspaces`
