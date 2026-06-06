# Orchestra 下阶段实施计划

> **Status update 2026-06-07:** This plan is retained as reference material. New actionable work should be derived from `docs/superpowers/specs/2026-06-07-orchestra-mvp-product-loop-design.md` and should prioritize the MVP loop over broad parity coverage: `Workspace -> Members -> Chat mention/DM -> Dispatch -> Agent session -> Output -> Chat/Task state`.

**Date**: 2026-05-04
**Baseline**: agent.Registry 重构已完成（C1 派发队列, C2 状态轮询, H1 @mention 解析）

---

## Phase 5: 核心可靠性（Critical + Quick Wins）

目标：消息投递有重试保障，前端状态实时准确。

| # | 任务 | 来源 | 工作量 | 说明 |
|---|------|------|--------|------|
| 5.1 | 接入 Outbox | C3 | M | outbox/worker.go 已实现，需要在 conversation handler 的 SendMessage 和 InternalChatSend 中写入 outbox 表，启动 worker goroutine |
| 5.2 | unread_sync 广播 | H3 | S | mark-read 后通过 ChatHub 广播 EventUnreadSync，前端 Pinia store 监听更新 |
| 5.3 | ChatDispatch 合并 | H2 | S | agent/dispatch.go 已有队列，添加时间窗口合并（连发消息 300ms 内合并为一条） |

---

## Phase 6: 终端引擎补全

目标：agent 输出可靠解析，大量输出不崩。

| # | 任务 | 来源 | 工作量 | 说明 |
|---|------|------|--------|------|
| 6.1 | 语义分析 Worker | H4 | L | 从终端 JSON 输出提取结构化内容（assistant_message / tool_use / result），替代当前的原始转发 |
| 6.2 | WebSocket 流控 | H5 | M | 输出节流 16ms/64KB + ACK 背压，防大量输出淹没前端 |
| 6.3 | 时间常量补全 | H6 | S | agent/constants.go 已有部分，补齐剩余（redraw 抑制、shell 检测超时等） |

---

## Phase 7: 编排层增强

目标：多 agent 协作更智能。

| # | 任务 | 来源 | 工作量 | 说明 |
|---|------|------|--------|------|
| 7.1 | DND/muted 过滤 | M1 | S | 派发时检查 member.muted，跳过 DND 成员 |
| 7.2 | DM 自动创建 | M2 | S | POST .../direct 端点，按 user+target 查找或创建 DM 会话 |
| 7.3 | last_message_preview | M3 | S | 消息写入时更新 conversations.last_message_preview（截断 120 字） |
| 7.4 | 通知 API 基础版 | M4 | M | badge counts + ignore-all + 浏览器 Notification |

---

## Phase 8: SDK Provider 模式

目标：支持通过 Anthropic API 直接驱动 agent，不依赖本地 CLI。

| # | 任务 | 来源 | 工作量 | 说明 |
|---|------|------|--------|------|
| 8.1 | 后端 model + migration | — | S | members 表加 provider_mode / sdk_model 字段 |
| 8.2 | SDK session 实现 | — | L | 新的 session 类型，通过 Anthropic Messages API 交互，实现与 tmux session 相同的接口 |
| 8.3 | 前端 provider mode UI | — | S | 恢复已撤回的 AddMember/EditMember modal 改动 |
| 8.4 | API key 管理 | — | M | Settings 页面管理 Anthropic API key，加密存储 |

---

## 工作量说明

- S = 小（< 2h）
- M = 中（2-6h）
- L = 大（6h+）

## 建议执行顺序

Phase 5 → Phase 6 → Phase 7 → Phase 8

Phase 5 最紧急（可靠性），Phase 8 最后做（新能力扩展）。
Phase 6 和 7 之间相对独立，可根据实际需要调整顺序。
