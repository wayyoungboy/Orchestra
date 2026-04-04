# Orchestra vs 对照端 — Gap analysis (summary)

**Date**: 2026-03-30  
**Status**: Superseded for detail by **[`dev_doc/orchestra-vs-reference-parity-analysis.md`](../dev_doc/orchestra-vs-reference-parity-analysis.md) (v2.1)** and **[`dev_doc/development-roadmap-parity.md`](../dev_doc/development-roadmap-parity.md)**.

This file is kept as a **short orientation**. The v1-era tables below mixed "missing" items that have since been implemented (read receipts, terminal search, attach, member row actions, i18n shell, diagnostics, etc.). **Do not use the legacy sections as a current backlog without cross-checking v2.1.**

---

## Summary (current)

| Area | Orchestra (Web) | vs 对照端 (Tauri) |
|------|-----------------|---------------------|
| Chat core | REST + SQLite；已读游标；频道成员；流式终端消息经 WS | IPC + 本地 DB；事件总线推送更全 |
| Terminal | Go PTY + WS；Search；重连；按成员 attach；服务端会话列表轮询 | 更深 VT/快照/编排 |
| Members / UI | `MemberRow`、侧栏动作、终端角标（WS + 轮询）、`InviteMenu` 邀请入口、`secretary`（秘书）角色 | 更多桌面向组件；Tauri 邀请仍多实例/沙箱等字段 |
| Cross-cutting | `vue-i18n` 管线（部分文案）、快捷键注册表导出、Toast、诊断日志、Skills 占位页 | 插件市场、原生监控等未搬 |
| Deliberately not Web | — | 多窗口、Tauri 插件加载、快照审计全量（见 v2.1 **不做清单**） |

---

## Legacy sections (historical)

The following blocks were written in **2026-03-29** against an older Orchestra tree. They are **not maintained line-by-line**. For an actionable list, see **v2.1 修订说明** and roadmap **§8 实施进度**.

<details>
<summary>Original long tables (collapsed)</summary>

The previous version listed P1/P2/P3 gaps (InviteAssistantModal parity, MemberRow missing, streaming missing, etc.). Many entries are now **done** or **partially done** — refer to `dev_doc/orchestra-vs-reference-parity-analysis.md` v2.1.

</details>

---

## Technical notes (still valid)

- Web: terminal streaming via **WebSocket**; clipboard via **Web Clipboard API**; notifications via **in-app Toast** / optional Web Push later.
- Architecture: 对照端 uses **Tauri IPC** where Orchestra uses **HTTP**; parity targets **behavior**, not identical transports.

---

## Conclusion

Orchestra matches a **web-appropriate subset** of 对照端 per roadmap Phases 1–4. Remaining differences are documented in **orchestra-vs-reference-parity-analysis.md v2.1** (partial rows, N/A Web, and the **不做清单**).
