# 对照端 对标 — 迭代轮次 03

**日期**: 2026-03-30  
**依据流程**: [`功能对标流程.md`](./功能对标流程.md)  
**总览**: [`development-roadmap-parity.md`](./development-roadmap-parity.md)  
**上轮**: [`development-roadmap-parity-02.md`](./development-roadmap-parity-02.md)

---

## 1. 本轮范围

落实 [orchestra-follow-up-from-parity-gaps.md](./orchestra-follow-up-from-parity-gaps.md) 中 **近期高优先级** 条目，并补齐成员邀请交互与 **秘书（`secretary`）** 角色；同步刷新差异分析 **v2.1**。

---

## 2. 可交付项（已对齐验收）

| 代号 | 内容 |
|------|------|
| **R03-1** | i18n：`AddMemberModal`、`EditMemberModal`、`ChatHeader`、`ChatInput`、`WorkspaceSelection`、`WorkspaceSwitcher` 等（轮次 02 backlog 扫尾）。 |
| **R03-2** | E2E：`GET /api/workspaces` 用例；`make reset-data` / `clean` 清除 SQLite WAL；Playwright 文档注释。 |
| **R03-3** | 消息契约：请求体可选 `clientTraceId`（Go 接受、暂不落库）；空列表 `GET /api/workspaces` 返回 `[]`。 |
| **R03-4** | 邀请 UI：`InviteMenu.vue`；`MembersSidebar` `#header-action`；`MembersPage` / `ChatInterface` 单入口 + 菜单（对齐 对照端 模式）。 |
| **R03-5** | 角色 **`secretary`**：前后端类型、分组排序、终端自动启动与频道用户输入 **PTY 转发**（与 `assistant` 并列）；中文产品名 **秘书**（对应 对照端 式监工语义）。 |

---

## 3. 本轮结论（Step 3）

- [x] R03-1 — i18n 扫尾  
- [x] R03-2 — E2E / 本地数据清理目标  
- [x] R03-3 — `clientTraceId` + 工作区列表 JSON  
- [x] R03-4 — 邀请菜单与侧栏槽位  
- [x] R03-5 — `secretary` 全链路  

**判定**: **已对齐**（本轮目标范围内）。差异分析报告升级为 [**orchestra-vs-reference-parity-analysis.md v2.1**](./orchestra-vs-reference-parity-analysis.md)。

---

## 4. 仍为后续 backlog（未在本轮关闭）

- 对照端 级 `inviteProjectMembers`：**实例数、沙箱、unlimitedAccess** 等与 Go `MemberCreate` 扩展。  
- `pendingTerminalMessages` / `readThrough`、全局未读推送、好友通讯录等 — 见 [后续工作规划](./orchestra-follow-up-from-parity-gaps.md) 中期/长期表。
