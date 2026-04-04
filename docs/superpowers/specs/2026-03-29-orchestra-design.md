# Orchestra - 多Agent协作Web系统设计文档

**版本**: 1.0
**日期**: 2026-03-29
**状态**: 待审核

---

## 一、项目概述

### 1.1 背景

Orchestra 是对**对照用桌面参考实现**（基于 Tauri）的 Web 版复刻，支持多 Agent（Claude Code、Gemini CLI、Codex 等）并行执行和编排。

### 1.2 目标

将对照用桌面参考实现的核心能力迁移到 Web 平台，同时保留：
- 多 CLI 工具并行执行
- Agent 编排与工作流
- 实时终端交互
- 聊天消息流

### 1.3 设计定位（项目初衷）

1. **服务化 + Web 使用** — 把对照用桌面参考实现做成**可由浏览器访问的后端服务**：CLI 与 PTY 跑在 Orchestra 所在机器上，前端为 Web（Vue），不再依赖本机 Tauri 与桌面 IPC。
2. **相对对照用桌面参考实现的刻意扩展** — **工作区 / 工作目录切换**：在服务端浏览与绑定路径，按工作区隔离，突破「单一本机工程根」的限制。
3. **其余能力** — 在行为、数据契约与交互上**尽量直接对齐对照用桌面参考实现**，以**迁移/搬运**为主（映射到 HTTP、WebSocket 与服务端路径），避免平行发明另一套产品语义。实现技术栈（Go 单体 vs Rust/Tauri）可以不同，**用户心智与功能边界**应与对照端一致、可对照。

### 1.4 开源协议

Apache 2.0

> 注：上文「1.3 设计定位」与仓库根目录 `CLAUDE.md` 中 **Project Charter** 互为中英文对照说明。

---

## 二、设计决策

| 决策项 | 选择 | 说明 |
|--------|------|------|
| 目标场景 | 单用户优先，预留多用户 | 用户认证作为可选模块 |
| CLI集成 | 后端代理模式 | 服务器运行 CLI 进程，前端通过 WebSocket 交互 |
| 工作目录 | 服务端路径浏览 | 用户指定服务器上的工作路径 |
| 前端技术栈 | Vue 3 + Pinia + Tailwind | 沿用对照端技术选型，降低迁移成本 |
| 后端技术栈 | Go 单体服务 | 高并发、进程管理能力强 |
| 终端隔离 | 初期无隔离 | 后续多用户时引入容器化 |
| 结构化存储 | SQLite | 预留 MySQL 迁移接口 |
| 文件存储 | 本地文件系统 | 用户工作空间目录 |
| 进程管理 | 进程池 | 限制 CLI 进程数量，防止资源耗尽 |

---

## 三、系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                     Vue 3 Frontend                          │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐│
│  │ Chat Module │ │ Terminal UI │ │ Agent Orchestrator View ││
│  │ (WebSocket) │ │ (WebSocket) │ │ (REST API)              ││
│  └─────────────┘ └─────────────┘ └─────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
                          │
                    HTTPS / WebSocket
                          │
┌─────────────────────────────────────────────────────────────┐
│                     Go Backend (单体)                        │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌───────────┐ │
│  │ HTTP API   │ │ WebSocket  │ │ Process    │ │ Agent     │ │
│  │ Handler    │ │ Gateway    │ │ Pool       │ │ Scheduler │ │
│  └────────────┘ └────────────┘ └────────────┘ └───────────┘ │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌───────────┐ │
│  │ Storage    │ │ File       │ │ Workspace  │ │ Security  │ │
│  │ Service    │ │ Manager    │ │ Manager    │ │ Module    │ │
│  └────────────┘ └────────────┘ └────────────┘ └───────────┘ │
└─────────────────────────────────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
  ┌─────▼─────┐    ┌─────▼─────┐    ┌─────▼─────┐
  │   SQLite  │    │  Local FS │    │CLI Processes│
  │  (消息/配置)│    │(工作目录) │    │(Claude/Gemini)│
  └───────────┘    └───────────┘    └───────────┘
```

---

## 四、前端架构

### 4.1 目录结构

```
src/
├── app/
│   ├── App.vue
│   └── router.ts
│
├── features/
│   ├── chat/
│   │   ├── ChatInterface.vue
│   │   ├── chatStore.ts
│   │   ├── chatApi.ts           # HTTP API 调用
│   │   ├── chatSocket.ts        # WebSocket 实时消息
│   │   ├── contactsStore.ts
│   │   ├── components/
│   │   │   ├── ChatHeader.vue
│   │   │   ├── ChatInput.vue
│   │   │   ├── ChatSidebar.vue
│   │   │   ├── MessagesList.vue
│   │   │   └── MembersSidebar.vue
│   │   └── modals/
│   │
│   ├── terminal/
│   │   ├── TerminalPane.vue
│   │   ├── TerminalWorkspace.vue
│   │   ├── terminalStore.ts
│   │   ├── terminalSocket.ts    # WebSocket 终端流
│   │   ├── terminalMemberStore.ts
│   │   ├── terminalEvents.ts
│   │   ├── components/
│   │   └── modals/
│   │
│   ├── workspace/
│   │   ├── WorkspaceSelection.vue
│   │   ├── PathBrowser.vue      # 服务端路径浏览
│   │   ├── workspaceStore.ts
│   │   ├── projectStore.ts
│   │   └── workspaceApi.ts
│   │
│   ├── agent/
│   │   ├── AgentPanel.vue
│   │   ├── SkillStore.vue
│   │   ├── skillLibrary.ts
│   │   ├── skillsBridge.ts
│   │   └── agentStore.ts
│   │
│   ├── settings/
│   │   ├── Settings.vue
│   │   ├── PluginMarketplace.vue
│   │   ├── settingsStore.ts
│   │   ├── themeStore.ts
│   │   └── theme.ts
│   │
│   ├── notifications/
│   │   ├── NotificationPreview.vue
│   │   └── notificationStore.ts
│   │
│   └── global/
│       ├── globalStore.ts
│       └── layoutStore.ts
│
├── shared/
│   ├── components/
│   ├── keyboard/
│   ├── context-menu/
│   ├── monitoring/
│   ├── utils/
│   └── types/
│
└── main.ts
```

### 4.2 功能模块对照

| 对照端模块 | Orchestra 模块 | 变化说明 |
|--------------|----------------|----------|
| chat/* | chat/* | Tauri 调用改为 HTTP/WebSocket |
| terminal/* | terminal/* | 多窗口改为单页面标签页 |
| workspace/* | workspace/* | 新增 PathBrowser 服务端路径浏览 |
| skills/* | agent/* | 重命名，语义更清晰 |
| global/* | global/* | 基本不变 |

---

## 五、后端架构

### 5.1 目录结构

```
internal/
├── api/
│   ├── router.go
│   ├── handlers/
│   │   ├── chat.go
│   │   ├── agent.go
│   │   ├── workspace.go
│   │   ├── settings.go
│   │   └── auth.go
│   └── middleware/
│       ├── auth.go
│       ├── ratelimit.go
│       └── logger.go
│
├── ws/
│   ├── gateway.go
│   ├── terminal.go
│   └── chat.go
│
├── terminal/
│   ├── pool.go              # 进程池
│   ├── session.go           # CLI 会话
│   ├── pty.go               # PTY 处理
│   └── process.go           # 进程管理
│
├── agent/
│   ├── scheduler.go
│   ├── dispatcher.go
│   ├── workflow.go
│   └── session.go
│
├── storage/
│   ├── database.go
│   ├── repository/
│   │   ├── interface.go     # 存储接口抽象
│   │   ├── conversation.go
│   │   ├── message.go
│   │   ├── agent.go
│   │   └── settings.go
│   └── migrations/
│
├── filesystem/
│   ├── workspace.go
│   ├── browser.go
│   └── validator.go
│
├── security/
│   ├── crypto.go            # API 密钥加密
│   ├── whitelist.go         # 命令/路径白名单
│   └── audit.go             # 审计日志
│
└── config/
    ├── config.go
    └── loader.go
```

### 5.2 核心组件说明

#### 5.2.1 进程池 (terminal/pool.go)

管理 CLI 子进程数量，防止资源耗尽：

```go
type ProcessPool struct {
    mu           sync.RWMutex
    sessions     map[string]*ProcessSession
    maxSessions  int           // 最大并发进程数
    idleTimeout  time.Duration // 空闲超时
}

type ProcessSession struct {
    ID           string
    PID          int
    Cmd          *exec.Cmd
    PTY          *os.File
    Workspace    string
    TerminalType string
    OutputChan   chan []byte
    ErrorChan    chan error
    DoneChan     chan struct{}
}
```

#### 5.2.2 安全模块 (security/)

- **命令白名单**: 只允许预定义的 CLI 命令执行
- **路径白名单**: 工作目录限制在允许范围内
- **密钥加密**: API Key 使用 AES-256 加密存储

```go
type Whitelist struct {
    commands map[string]bool
    paths    []string
}

func (w *Whitelist) ValidateCommand(cmd string) error
func (w *Whitelist) ValidatePath(path string) error

type KeyEncryptor struct {
    key []byte // 32 bytes for AES-256
}

func (e *KeyEncryptor) Encrypt(plaintext string) (string, error)
func (e *KeyEncryptor) Decrypt(ciphertext string) (string, error)
```

---

## 六、数据模型

### 6.1 SQLite Schema

```sql
-- 工作区表
CREATE TABLE workspaces (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    last_opened_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL
);

-- 成员表（Agent/CLI 配置）
CREATE TABLE members (
    id TEXT PRIMARY KEY,
    workspace_id TEXT NOT NULL,
    name TEXT NOT NULL,
    role_type TEXT NOT NULL,      -- 'owner' | 'admin' | 'assistant' | 'member'
    role_key TEXT,
    avatar TEXT,
    terminal_type TEXT,           -- 'claude' | 'gemini' | 'codex' | 'custom'
    terminal_command TEXT,
    terminal_path TEXT,
    auto_start_terminal BOOLEAN DEFAULT TRUE,
    status TEXT DEFAULT 'online',
    created_at INTEGER NOT NULL,
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id)
);

-- 会话表
CREATE TABLE conversations (
    id TEXT PRIMARY KEY,
    workspace_id TEXT NOT NULL,
    type TEXT NOT NULL,           -- 'dm' | 'group' | 'channel'
    name TEXT,
    target_id TEXT,
    member_ids TEXT NOT NULL,     -- JSON 数组
    pinned BOOLEAN DEFAULT FALSE,
    muted BOOLEAN DEFAULT FALSE,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id)
);

-- 消息表
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    content TEXT NOT NULL,        -- JSON: {type, text, key}
    is_ai BOOLEAN DEFAULT FALSE,
    attachment TEXT,              -- JSON
    status TEXT DEFAULT 'sent',
    created_at INTEGER NOT NULL,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);

-- 工作流模板表
CREATE TABLE workflows (
    id TEXT PRIMARY KEY,
    workspace_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    steps TEXT NOT NULL,          -- JSON 数组
    created_at INTEGER NOT NULL,
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id)
);

-- 设置表
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL           -- JSON
);

-- API 密钥表（加密存储）
CREATE TABLE api_keys (
    id TEXT PRIMARY KEY,
    provider TEXT NOT NULL,       -- 'anthropic' | 'openai' | 'google' | 'custom'
    encrypted_key TEXT NOT NULL,
    created_at INTEGER NOT NULL
);

-- 索引
CREATE INDEX idx_members_workspace ON members(workspace_id);
CREATE INDEX idx_conversations_workspace ON conversations(workspace_id);
CREATE INDEX idx_messages_conversation ON messages(conversation_id);
CREATE INDEX idx_messages_created ON messages(created_at);
```

### 6.2 多用户预留

所有表预留 `user_id` 字段迁移路径：

```sql
-- 多用户时添加
ALTER TABLE workspaces ADD COLUMN user_id TEXT;
ALTER TABLE members ADD COLUMN user_id TEXT;
ALTER TABLE conversations ADD COLUMN user_id TEXT;
-- ...
CREATE INDEX idx_workspaces_user ON workspaces(user_id);
```

---

## 七、API 设计

### 7.1 REST API

| 方法 | 路径 | 说明 |
|------|------|------|
| **工作区** |||
| GET | /api/workspaces | 获取工作区列表 |
| POST | /api/workspaces | 创建工作区 |
| GET | /api/workspaces/:id | 获取工作区详情 |
| DELETE | /api/workspaces/:id | 删除工作区 |
| GET | /api/workspaces/:id/browse | 浏览工作区文件 |
| **成员** |||
| GET | /api/workspaces/:id/members | 获取成员列表 |
| POST | /api/workspaces/:id/members | 添加成员 |
| PUT | /api/workspaces/:id/members/:memberId | 更新成员 |
| DELETE | /api/workspaces/:id/members/:memberId | 删除成员 |
| **会话** |||
| GET | /api/conversations | 获取会话列表 |
| POST | /api/conversations | 创建会话 |
| GET | /api/conversations/:id/messages | 获取消息列表 |
| POST | /api/conversations/:id/messages | 发送消息 |
| **工作流** |||
| GET | /api/workflows | 获取工作流列表 |
| POST | /api/workflows | 创建工作流 |
| POST | /api/workflows/:id/execute | 执行工作流 |
| **设置** |||
| GET | /api/settings | 获取设置 |
| PUT | /api/settings | 更新设置 |
| POST | /api/settings/api-keys | 添加 API Key |
| DELETE | /api/settings/api-keys/:id | 删除 API Key |

### 7.2 WebSocket

| 路径 | 说明 |
|------|------|
| /ws/terminal/:sessionId | 终端流（双向） |
| /ws/chat | 聊天实时消息（双向） |

#### 终端 WebSocket 协议

```typescript
// 客户端 -> 服务端
{ type: 'input', data: string }      // 终端输入
{ type: 'resize', cols: number, rows: number }  // 终端大小调整
{ type: 'close' }                    // 关闭会话

// 服务端 -> 客户端
{ type: 'output', data: string }     // 终端输出
{ type: 'error', message: string }   // 错误信息
{ type: 'exit', code: number }       // 进程退出
{ type: 'connected', sessionId: string }  // 连接成功
```

---

## 八、配置

### 8.1 配置文件 (config.yaml)

```yaml
server:
  http_addr: ":8080"
  ws_addr: ":8081"

terminal:
  max_sessions: 10          # 最大并发 CLI 进程数
  idle_timeout: 30m         # 空闲超时

security:
  encryption_key: "${ORCHESTRA_ENCRYPTION_KEY}"
  allowed_commands:
    - claude
    - gemini
    - codex
    - qwen
  allowed_paths:
    - /home/user/projects
    - /var/orchestra/workspaces

storage:
  database: "./data/orchestra.db"
  workspaces: "./workspaces"
```

### 8.2 环境变量

| 变量 | 说明 |
|------|------|
| ORCHESTRA_ENCRYPTION_KEY | API 密钥加密密钥（32字节） |
| ORCHESTRA_CONFIG | 配置文件路径（可选） |

---

## 九、安全设计

### 9.1 风险与缓解

| 风险 | 严重程度 | 缓解措施 |
|------|----------|----------|
| 远程代码执行 | 致命 | 命令白名单、路径白名单 |
| API 密钥泄露 | 高 | AES-256 加密存储 |
| 路径遍历 | 高 | 路径白名单验证 |
| 进程资源耗尽 | 高 | 进程池限制并发数 |
| WebSocket 劫持 | 中 | 认证中间件（预留） |

### 9.2 安全最佳实践

1. **密钥管理**: 加密密钥从环境变量读取，不落盘
2. **命令限制**: 只允许预定义的 CLI 命令
3. **路径限制**: 工作目录限制在白名单路径内
4. **审计日志**: 记录所有敏感操作
5. **速率限制**: 防止 API 滥用

---

## 十、扩展路线

### Phase 1 (MVP)

- [ ] 基础 CLI 进程管理
- [ ] WebSocket 终端流
- [ ] SQLite 存储
- [ ] 工作区路径白名单
- [ ] 基础安全控制

### Phase 2 (产品化)

- [ ] 用户认证系统
- [ ] 进程池配额管理
- [ ] 审计日志完善
- [ ] 配置热重载
- [ ] 错误监控

### Phase 3 (多用户)

- [ ] 迁移到 MySQL/PostgreSQL
- [ ] 完整 RBAC 权限
- [ ] 容器化隔离（可选）
- [ ] 高可用部署

---

## 十一、参考

- 对照用桌面参考实现的本地源码路径由团队约定（与 Orchestra 仓库并列或独立 checkout 均可）。
- xterm.js: https://xtermjs.org/
- Vue 3 文档: https://vuejs.org/
- Go 文档: https://golang.org/