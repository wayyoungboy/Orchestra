# 对照端 对标 — 迭代轮次 02

**日期**: 2026-03-30  
**依据流程**: [`功能对标流程.md`](./功能对标流程.md)  
**总览**: [`development-roadmap-parity.md`](./development-roadmap-parity.md)  
**上轮**: [`development-roadmap-parity-01.md`](./development-roadmap-parity-01.md)

---

## 1. 本轮范围

延续轮次 01 的 **i18n 纵深**：覆盖成员行（`MemberRow`）菜单/右键/状态/终端角标文案、主壳未选工作区提示、聊天侧栏分区标题，使与 对照端 一致的多语言体验在 **高频路径** 上连贯。

---

## 2. 当前状态简述

- 轮次 01：`MembersPage`、`LoginPage`、登录 E2E 已交付。
- `MemberRow`、`WorkspaceMain` 空态、`ChatSidebar` 仍为硬编码英文，与设置中切换中文不一致。

---

## 3. 差距列表

| # | 对照端 期望 | Orchestra 现状 | 保留差异？ |
|---|--------------|----------------|------------|
| G1 | 成员行菜单与状态多语言 | `MemberRow` 全英文 | 否 |
| G2 | 全局壳层提示多语言 | `WorkspaceMain` 重复英文空态 | 否 |
| G3 | 会话侧栏分区标题多语言 | `ChatSidebar` Channels/DM 英文 | 否 |

---

## 4. 本轮可交付项

- [ ] **R02-1**：`MemberRow.vue` — `vue-i18n`（下拉、状态区、角标 title、右键菜单、`displayRole`）。
- [ ] **R02-2**：`WorkspaceMain.vue` — 未选工作区时的标题与说明。
- [ ] **R02-3**：`ChatSidebar.vue` — Channels / Direct Messages / 默认工作区名。
- [ ] **R02-4**：扩展 `locales/en.json`、`zh.json`；总览路线图增加 **轮次 02** 链接。

---

## 5. 后续轮次 backlog

- ~~`AddMemberModal` / `EditMemberModal` / `ChatHeader` / `ChatInput` 全量 i18n~~；~~`WorkspaceSelection` / `WorkspaceSwitcher` 文案~~ → 已在 **[轮次 03](./development-roadmap-parity-03.md)** 关闭。

---

## 6. 验证

- `pnpm exec vue-tsc --noEmit`
- `pnpm run build && pnpm exec playwright test e2e/auth-flow.spec.ts e2e/smoke.spec.ts`

---

## 7. 本轮结论（Step 3）

- [x] R02-1 — MemberRow i18n
- [x] R02-2 — WorkspaceMain 空态
- [x] R02-3 — ChatSidebar
- [x] R02-4 — 文案与总览链接

**判定**: **已对齐**（本轮目标范围内）。提交信息建议：`feat(i18n): member row, workspace empty state, chat sidebar (parity round 02)`
