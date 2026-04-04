# Orchestra 后续工作规划（基于与 对照端 的差异）

**版本**: 1.1  
**日期**: 2026-03-30  
**依据**: [orchestra-vs-reference-parity-analysis.md](./orchestra-vs-reference-parity-analysis.md) **v2.1**、[development-roadmap-parity.md](./development-roadmap-parity.md) 及 [轮次 01–03](./development-roadmap-parity-03.md) 结论。

本文回答：**在已交付的 Web 子集之上，接下来建议做什么、按什么顺序做、什么刻意不做。** 实施时仍按 [功能对标流程.md](./功能对标流程.md) 可继续开 `development-roadmap-parity-NN.md` 做增量轮次。

---

## 1. 原则

1. **不削弱 Orchestra 自有目标**：服务端工作区、REST 浏览、SQLite 权威数据源等（见 `CLAUDE.md` / 设计规格）。
2. **对齐的是「用户可感知行为」**，不是 Tauri API 名称或桌面专属能力。
3. **优先低风险、可验收条目**：i18n 收尾、E2E、契约清晰的中台能力，再做大功能（好友、插件市场）。
4. **每批改动对应可运行验证**：`vue-tsc`、`build`、现有 e2e；需要后端的用例在 CI 或文档中说明启动方式。

### 1.1 已落实（轮次 03 / 2026-03-30）

- **A / B / C** 已在 [development-roadmap-parity-03.md](./development-roadmap-parity-03.md) 闭环（i18n 扫尾、E2E `workspaces`、`clientTraceId` 与 §5 表更新）。  
- **成员邀请**：`InviteMenu` + 侧栏 `header-action`，与 对照端 单入口菜单模式对齐；`AddMemberModal` 仍无 Tauri 级实例/沙箱字段（见 v2.1 §3.1）。  
- **秘书角色**：`roleType: secretary`，PTY 转发与助手并列；产品文案「秘书」对应 对照端 式监工语义。

---

## 2. 近期建议（高优先级，1～3 个迭代内）

| 方向 | 内容 | 对应差距（v2.1） | 状态 |
|------|------|------------------|------|
| **A. i18n 扫尾** | 见上轮 backlog 所列组件 | §3.4 | **已完成**（轮次 03） |
| **B. E2E 加固** | `workspaces` + health；`make reset-data` | 质量基线 | **已完成**（轮次 03） |
| **C. 消息与发送契约** | `clientTraceId` 可选字段 | §5 | **已完成**（轮次 03；mentions 全量文档仍可补） |

---

## 3. 中期建议（产品体验与 对照端 事件模型靠拢）

| 方向 | 内容 | 对应差距 | 备注 |
|------|------|----------|------|
| **D. 全局未读 / 消息状态推送** | 增加 **工作区级 SSE 或 WebSocket**（或短轮询升级），服务端在相关事件时推送；前端在入口订阅并调用现有 `applyUnreadSync` / `applyMessageStatus` | §3.1.1「无全局自动订阅」 | 比单纯轮询聊天列表更省、更接近 Tauri 事件总线语义 |
| **E. 终端流与 chatStore 深度对齐** | `pendingTerminalMessages`、`readThrough` 等与 对照端 行为逐项对照实现（可分阶段） | REQ-206 / §3.1.1「部分」 | 改动集中在前端 store + 必要时 WS payload 约定 |
| **F. 邀请 / 成员弹窗能力** | `InviteMenu` 入口与 CLI 选型已有；仍缺 **实例数 / 沙箱 / unlimitedAccess** 等与 `inviteProjectMembers` 对齐的 API 字段 | §3.1 Invite*「部分」 | 需产品裁剪 + `MemberCreate` 扩展 |
| **G. 消息类型与展示** | 附件、富文本或 对照端 已有消息类型的子集；`MessagesList` + API | §3.1 MessagesList「部分」 | 与存储迁移、上传 API 绑定，宜单独里程碑 |
| **H. 认证与安全** | 替换占位 `authStore` 为可部署方案（会话/JWT/OAuth 择一） | §3.5 / §5 认证 | 与「多用户」路线图强相关 |

---

## 4. 长期 / 可选（投入大或依赖产品决策）

| 方向 | 内容 | 说明 |
|------|------|------|
| **I. 好友与通讯录** | `contactsStore`、`FriendsView`、邀请好友链路 | 对照端 完整社交图；Web 是否要做需产品定调 |
| **J. Skills / 插件** | HTTP 插件清单、安装、与运行隔离；替代 Tauri `skills_*` | 当前仅为 [Skills 占位页](../frontend/src/features/skills/SkillsPlaceholder.vue) |
| **K. 终端进阶** | `terminal_list_environments`、更强编排、**快照审计**（REQ-304） | 与 Go `terminal_engine` 级能力差距大；快照已标延期 |
| **L. 分析报告迭代** | 大版本发布后升级 `orchestra-vs-reference-parity-analysis.md` 至 v2.1/v3.0 | 复用 R5 流程 |

---

## 5. 刻意不做或长期不对齐（Web / 成本边界）

以下在 v2.1 **不做清单** 或 **N/A Web** 中已说明，**不应**作为「必须赶上 对照端」的 backlog：

- Tauri **多窗口**、**原生通知窗口**、**平台更新/激活**。
- **全库** `chat_clear_all_messages` / repair 等运维命令（除非做内部运维工具）。
- **VT 语义引擎 + 快照审计全量** 与桌面一致（除非单独立项服务端终端产品）。
- **插件市场实装** 无明确 Web 商业模式前保持占位 + 文档即可。

---

## 6. 与现有文档的索引关系

| 文档 | 用途 |
|------|------|
| [orchestra-vs-reference-parity-analysis.md](./orchestra-vs-reference-parity-analysis.md) | **差异事实源**（有/无/部分） |
| [development-roadmap-parity.md](./development-roadmap-parity.md) | REQ Phase 与 R1–R5 总览 |
| [development-roadmap-parity-01/02/03](./development-roadmap-parity-03.md) | 已完成的增量轮次 |
| [功能对标流程.md](./功能对标流程.md) | 下一轮 `parity-NN` 的写作与评审步骤 |
| 本文 | **从差距推导的优先行动清单**（可随季度修订） |

---

## 7. 建议的下一里程碑命名（供轮次文档引用）

- **轮次 03**：**已完成** — 见 [development-roadmap-parity-03.md](./development-roadmap-parity-03.md)。  
- **轮次 04（建议）**：**D 推送通道**（未读/消息状态）或 **E** `pendingTerminalMessages` 子集；同步可将差异报告勘误为 v2.2（若有大量 API 变更）。  
- **里程碑 M2**：**E** 或 **G** 二选一深做，避免并行拖长周期。

---

*文档结束。修订时请更新版本号与日期，并在 `development-roadmap-parity.md` 文首保留指向本文的链接。*
