# Orchestra

多智能体协作平台 - 一个基于 Web 的系统，用于编排多个 AI 智能体（Claude Code、Gemini CLI、Aider 等）并行运行，支持实时聊天和工作区管理。

[English](README.md)

## 产品方向

Orchestra 近期路线围绕一个 MVP 闭环：

```text
Workspace -> Members -> Chat mention/DM -> Dispatch -> Agent session -> Output -> Chat/Task state
```

参考桌面端行为与 Golutra 规格用于约束产品语义，但不作为全量 parity backlog。优先级以是否能强化这个闭环为准，而不是追逐所有可迁移功能。

MVP 阶段的 agent 执行路线以 CLI 为主：PTY + tmux 是可持久、可观察的本地会话底座；ACP 在可用时作为结构化协议增强；skills 用于打包和扩展 agent 能力；A2A 等本地 CLI 闭环稳定后再进入主线。成员级 CLI/ACP 配置是创建终端会话时的权威来源，用户添加 assistant 后不需要再次输入启动命令。

成员页已经直接暴露这个入口：配置过命令的 assistant / secretary 会显示已有后端会话状态，也可以从成员卡片启动或复用后端 agent 会话；启动失败会显示在卡片内，不再只依赖聊天派发时才暴露问题。Agent 会话页会列出当前工作区的活跃后台会话，把当前 tmux 面板快照加载到 xterm.js 终端面板，跟随后续实时输出，把终端 resize 事件同步回 tmux，支持直接按键转发和受控整行输入，也可以终止不再需要的后台会话。

## 致谢

本项目的设计思路参考了 [golutra](https://github.com/golutra/golutra)。特别感谢该项目及其作者 [seeksky](https://github.com/seekskyworld) 的启发。本项目所有代码均为独立实现。

## 功能特性

- **多智能体终端管理**：并行运行多个 AI 智能体终端，每个终端拥有独立的 PTY 会话
- **成员级 Agent 启动**：从成员卡片使用已保存的 CLI/ACP 命令检查、启动或复用 assistant / secretary 后端会话
- **Agent 会话检查**：通过独立导航页查看当前工作区活跃后台会话、所属成员、当前 tmux 面板快照、带 tmux resize 同步的 xterm.js 实时输出、直接按键转发、受控整行输入，并终止不再需要的后台会话
- **实时协作聊天**：内置聊天界面，支持 @提及 功能定向发送消息给特定成员
- **工作区管理**：创建和切换多个工作区，可配置服务器端路径
- **成员角色**：基于角色的权限控制（Owner、Admin、Secretary、Assistant、Member）
- **路径浏览器**：浏览并选择服务器端目录绑定到工作区
- **现代柔光玻璃风格 UI**：基于 Vue 3、TypeScript 和 Tailwind CSS 构建的简洁现代界面
- **WebSocket 终端流**：通过 WebSocket 实时传输终端输出，支持 ANSI 颜色
- **国际化支持**：支持英文和中文

## 技术栈

### 后端
- **Go 1.25+** - 核心运行时
- **Gin** - HTTP 框架
- **gorilla/websocket** - WebSocket 处理
- **SQLite** - 持久化存储
- **PTY** - 终端模拟（通过 `creack/pty`）

### 前端
- **Vue 3 + TypeScript** - UI 框架
- **Pinia** - 状态管理
- **Tailwind CSS v4** - 样式
- **xterm.js** - 终端渲染
- **vue-i18n** - 国际化

## 快速开始

### 环境要求
- Go 1.25+
- Node.js 18+（推荐使用 pnpm）
- tmux（agent 会话和 MVP 验证入口需要）

### 后端设置

```bash
cd backend

# 安装依赖
go mod download

# 构建
make build

# 运行（启动于 http://localhost:8080）
make run
```

也可以在仓库根目录启动同一个后端服务：

```bash
make backend-run
# 或
make run
```

### 前端设置

```bash
cd frontend

# 安装依赖
pnpm install

# 开发服务器（启动于 http://localhost:5173）
pnpm dev

# 生产构建
pnpm build
```

### 清理重置（用于测试）

如需全新启动，清除旧数据：

```bash
# 先停止后端（Ctrl+C）
make reset-data  # 在仓库根目录执行；删除数据库和 WAL 文件

# 等价的后端子目录命令：
# cd backend && make reset-data

# 可选：清除浏览器 localStorage
# DevTools → Application → Clear site data
# 或删除键：orchestra-settings, orchestra.auth, orchestra.user
```

### 代码更新后

拉取代码变更后重启两个开发进程：

1. **后端**：停止（`Ctrl+C`）→ 在仓库根目录运行 `make run`，或 `cd backend && make run`
2. **前端**：停止（`Ctrl+C`）→ 在仓库根目录运行 `make frontend-dev`，或 `cd frontend && pnpm dev`

> 注意：Go 无热重载；Vite HMR 可能遗漏某些边缘情况。

## 项目结构

```
Orchestra/
├── backend/              # Go 后端
│   ├── cmd/              # 入口点（main.go）
│   ├── internal/         # 内部模块
│   │   ├── api/          # HTTP 处理器 & 路由
│   │   ├── config/       # 配置加载器
│   │   ├── filesystem/   # 路径浏览器服务
│   │   ├── models/       # 数据模型
│   │   ├── storage/      # SQLite 仓储层
│   │   └── terminal/     # PTY 管理
│   ├── pkg/              # 公共工具
│   ├── configs/          # 配置文件
│   └── Makefile          # 构建命令
├── frontend/             # Vue 前端
│   ├── src/
│   │   ├── app/          # 应用设置、路由、i18n
│   │   ├── assets/       # CSS、静态资源
│   │   ├── features/     # 功能模块
│   │   │   ├── auth/     # 认证
│   │   │   ├── chat/     # 聊天界面
│   │   │   ├── members/  # 成员管理
│   │   │   ├── settings/ # 设置页面
│   │   │   ├── terminal/ # 终端工作区
│   │   │   └── workspace/# 工作区选择
│   │   └── shared/       # 共享组件、API、工具
│   └── public/
├── docs/                 # 文档
│   └── superpowers/      # 规范和计划
├── CLAUDE.md             # 项目说明
├── README.md             # 英文文档
└── README_CN.md          # 中文文档
```

## 成员角色

| 角色 | 说明 |
|------|------|
| **Owner** | 工作区和成员的完全控制权 |
| **Admin** | 可管理成员和工作区设置 |
| **Secretary** | 协调员角色（监控/编排语义） |
| **Assistant** | 可参与聊天和使用终端 |
| **Member** | 基础参与者，权限有限 |

## API 端点

### REST API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/swagger/*any` | GET | Swagger UI 和 OpenAPI 静态资源 |
| `/api/auth/config` | GET | 读取认证模式和注册配置 |
| `/api/auth/login` | POST | 登录并签发认证令牌 |
| `/api/auth/validate` | POST | 校验认证令牌 |
| `/api/auth/me` | GET | 读取当前认证用户 |
| `/api/auth/register` | POST | 注册开放时创建用户 |
| `/api/workspaces` | GET/POST | 列出或创建工作区 |
| `/api/workspaces/validate-path` | POST | 校验服务端工作区路径 |
| `/api/workspaces/:id` | GET/PUT/DELETE | 读取、更新或删除工作区 |
| `/api/browse` | GET | 浏览服务器路径 |
| `/api/workspaces/:id/browse` | GET | 浏览指定工作区路径 |
| `/api/workspaces/:id/search` | GET | 搜索工作区路径 |
| `/api/workspaces/:id/members` | GET/POST | 列出或添加工作区成员 |
| `/api/workspaces/:id/members/:memberId` | GET/PUT/DELETE | 读取、更新或移除成员 |
| `/api/workspaces/:id/members/:memberId/conversations` | DELETE | 删除成员相关会话 |
| `/api/workspaces/:id/members/:memberId/presence` | POST | 更新成员在线状态/活动 |
| `/api/workspaces/:id/members/:memberId/terminal-session` | GET/POST | 检查、启动或复用成员 agent 会话 |
| `/api/workspaces/:id/terminal-sessions` | GET | 列出工作区 agent 会话 |
| `/api/terminals` | POST | 创建终端会话 |
| `/api/terminals/:sessionId/snapshot` | GET | 读取当前终端面板快照 |
| `/api/terminals/:sessionId` | DELETE | 关闭终端会话 |
| `/api/workspaces/:id/conversations` | GET/POST | 列出或创建会话 |
| `/api/workspaces/:id/conversations/:convId` | GET/PUT/DELETE | 读取、更新或删除会话 |
| `/api/workspaces/:id/conversations/direct` | POST | 创建或复用私聊会话 |
| `/api/workspaces/:id/conversations/:convId/members` | PUT | 替换会话成员 |
| `/api/workspaces/:id/conversations/:convId/messages` | GET/POST | 列出或发送会话消息 |
| `/api/workspaces/:id/conversations/:convId/messages` | DELETE | 清空会话消息 |
| `/api/workspaces/:id/conversations/:convId/messages/:messageId` | DELETE | 删除单条消息 |
| `/api/workspaces/:id/conversations/:convId/read` | POST | 标记会话已读 |
| `/api/workspaces/:id/conversations/read-all` | POST | 标记工作区全部会话已读 |
| `/api/internal/chat/send` | POST | 内部 AI 结果消息 API |
| `/api/internal/agent-status` | POST | 内部 agent 状态更新 API |
| `/api/internal/tasks/create` | POST | agent 创建任务 |
| `/api/internal/tasks/assign` | POST | 分派任务给 agent |
| `/api/internal/tasks/start` | POST | agent 开始任务 |
| `/api/internal/tasks/complete` | POST | agent 完成任务 |
| `/api/internal/tasks/fail` | POST | agent 标记任务失败 |
| `/api/internal/tasks/cancel` | POST | agent 取消任务 |
| `/api/internal/tasks/list` | GET | 按 secretary 查询任务 |
| `/api/internal/workloads/list` | GET | 列出 agent 工作负载 |
| `/api/workspaces/:id/tasks` | GET | 列出工作区任务 |
| `/api/workspaces/:id/tasks/:taskId` | GET | 获取任务详情 |
| `/api/workspaces/:id/tasks/my-tasks` | GET | 列出当前 agent 被分派的任务 |
| `/api/workspaces/:id/tasks/:taskId/cancel` | POST | 取消任务 |
| `/api/workspaces/:id/attachments` | GET | 列出工作区附件 |
| `/api/workspaces/:id/conversations/:convId/attachments` | POST | 上传会话附件 |
| `/api/workspaces/:id/attachments/:attachmentId` | GET/DELETE | 下载或删除附件 |
| `/api/workspaces/:id/attachments/:attachmentId/info` | GET | 读取附件元数据 |
| `/api/api-keys` | GET/POST | 列出或创建 API key |
| `/api/api-keys/provider/:provider` | GET | 获取指定 provider 的 API key |
| `/api/api-keys/:id` | DELETE | 删除 API key |
| `/api/api-keys/test` | POST | 测试 API key |
| `/api/workspaces/:id/notifications` | GET | 列出工作区通知 |
| `/api/workspaces/:id/notifications/badge` | GET | 读取通知角标计数 |
| `/api/workspaces/:id/notifications/:notifId/read` | POST | 标记通知已读 |
| `/api/workspaces/:id/notifications/read-all` | POST | 标记工作区全部通知已读 |
| `/api/workspaces/:id/outbox` | GET | 列出工作区 outbox 投递诊断 |

### WebSocket

| 端点 | 说明 |
|------|------|
| `/ws/terminal/:sessionId` | 终端 I/O 流 |
| `/ws/chat/:workspaceId` | 聊天消息流 |

## 配置

配置文件：`backend/configs/config.yaml`

环境变量：
- `ORCHESTRA_ENCRYPTION_KEY` - API 密钥加密密钥（32+ 字节）
- `ORCHESTRA_CONFIG` - 自定义配置文件路径

### 默认配置

```yaml
server:
  http_addr: ":8080"

terminal:
  max_sessions: 10
  idle_timeout: 30m

security:
  allowed_commands:
    - /bin/bash
    - /bin/zsh
    - claude        # Claude Code CLI
    - gemini        # Gemini CLI
    - aider         # Aider
  allowed_paths:
    - ~/projects    # 限制路径浏览范围
  allowed_origins:
    - "http://localhost:5173"

storage:
  database: "./data/orchestra.db"
```

## 开发

### 代码规范
- Go: `gofmt`, `goimports`
- TypeScript: ESLint, Prettier
- 提交: Conventional Commits（`feat:`, `fix:`, `docs:` 等）

### 运行测试

```bash
# 根目录标准快捷命令
make verify
make verify-focused

# 当前 MVP 验证入口（后端测试、前端构建/单测、focused spec typecheck）
./scripts/verify-mvp.sh

# 自动启动临时后端并运行全部 focused browser MVP E2E
./scripts/run-focused-e2e-local.sh

# 在 GitHub Actions 手动运行 CI workflow，并启用
# "Run the full focused browser MVP E2E gate"，即可在远端执行 make verify-focused。

# 一次性纳入全部 focused browser MVP E2E（需要后端和 tmux）
ORCHESTRA_RUN_ALL_FOCUSED_E2E=1 ./scripts/verify-mvp.sh

# 将首屏工作区创建/进入聊天流程纳入验证入口（需要后端）
ORCHESTRA_RUN_WORKSPACE_ONBOARDING_E2E=1 ./scripts/verify-mvp.sh

# 将聚焦的浏览器终端 E2E 纳入验证入口（需要后端和 tmux）
ORCHESTRA_RUN_TERMINAL_E2E=1 ./scripts/verify-mvp.sh

# 将聚焦的浏览器 MVP 聊天流程纳入验证入口（需要后端）
ORCHESTRA_RUN_MVP_CHAT_E2E=1 ./scripts/verify-mvp.sh

# 将聚焦的浏览器 MVP 私聊流程纳入验证入口（需要后端和 tmux）
ORCHESTRA_RUN_MVP_DM_E2E=1 ./scripts/verify-mvp.sh

# 将聚焦的浏览器 MVP 未读同步流程纳入验证入口（需要后端）
ORCHESTRA_RUN_MVP_UNREAD_E2E=1 ./scripts/verify-mvp.sh

# 将聚焦的成员卡片 agent 会话流程纳入验证入口（需要后端和 tmux）
ORCHESTRA_RUN_MVP_MEMBER_SESSION_E2E=1 ./scripts/verify-mvp.sh

# 将聚焦的浏览器 MVP 任务流程纳入验证入口（需要后端）
ORCHESTRA_RUN_MVP_TASK_E2E=1 ./scripts/verify-mvp.sh

# 后端单元测试
cd backend && make test

# 后端 Go 格式检查
make backend-format-check

# 后端静态分析
make backend-vet

# 聚焦的后端终端 API 运行时 smoke（需要 tmux）
cd backend && go test ./internal/api -run TestTerminalRuntimeAPIWorkspaceMemberSessionLifecycle -count=1

# 聚焦的后端结果返回闭环（需要 tmux）
cd backend && go test ./internal/api -run TestAssistantResultCompletesTaskAndForwardsToSecretary -count=1

# 前端单元测试
cd frontend && pnpm test

# 前端 lint 检查
cd frontend && pnpm lint:check

# E2E 测试（需要运行后端）
cd frontend && pnpm test:e2e

# 聚焦的 MVP 聊天浏览器流程（需要运行后端）
cd frontend && pnpm test:e2e:mvp-chat

# 聚焦的 MVP 私聊浏览器流程（需要运行后端和 tmux）
cd frontend && pnpm test:e2e:mvp-dm

# 聚焦的 MVP 未读同步浏览器流程（需要运行后端）
cd frontend && pnpm test:e2e:mvp-unread

# 聚焦的成员卡片 agent 会话浏览器流程（需要后端和 tmux）
cd frontend && pnpm test:e2e:mvp-member-session

# 聚焦的 MVP 任务浏览器流程（需要运行后端）
cd frontend && pnpm test:e2e:mvp-task

# 聚焦的 Agent 终端运行时 E2E（需要后端和 tmux）
cd frontend && pnpm test:e2e:terminal

# E2E 使用自定义后端 URL
ORCHESTRA_API_URL=http://your-server:8080 pnpm test:e2e
```

聚焦的 E2E runner 会为本地浏览器/API 流量清理继承到的 HTTP/SOCKS 代理变量，并把 `127.0.0.1`、`localhost`、`::1` 追加到 `NO_PROXY`，避免全局开发代理影响 localhost 验证。

### Make 命令

```bash
# 根目录 Makefile 目标
make verify           # 后端测试、前端构建/单测、focused spec typecheck
make verify-focused   # 临时后端 + 全部 focused browser MVP E2E
make run              # 启动后端 API 服务（根目录别名）
make test             # 运行后端测试（根目录别名）
make build            # 构建后端服务二进制（根目录别名）
make reset-data       # 清理后端 SQLite 数据（根目录别名）
make backend-run      # 启动后端 API 服务
make backend-test     # 运行后端测试
make backend-reset    # 清理后端 SQLite 数据
make backend-format-check # 检查后端 Go 格式
make backend-vet      # 运行后端 go vet
make frontend-install # 安装前端依赖
make frontend-dev     # 启动前端开发服务器
make frontend-build   # 构建前端
make frontend-test    # 运行前端单测
make frontend-lint    # 运行前端 lint 检查

# ./backend 下的后端 Makefile 目标（也可通过上方根目录别名访问）
make build            # 构建二进制
make run              # 运行服务器
make test             # 运行测试
make reset-data       # 清理数据库
make clean            # 删除构建产物
```

## 开发路线

- [ ] 终端会话持久化和重连
- [ ] 成员在线状态指示
- [ ] 每成员命令历史
- [ ] 工作区模板
- [ ] 每成员 API 密钥管理
- [ ] 导出聊天记录

## 许可证

MIT License

## 贡献

欢迎贡献！请：
1. Fork 仓库
2. 创建功能分支
3. 遵循代码规范（gofmt、ESLint）
4. 为新功能编写测试
5. 提交 PR 并附清晰描述

---

为 AI 辅助开发工作流而构建 ❤️
