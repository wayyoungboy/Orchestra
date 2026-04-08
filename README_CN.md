# Orchestra

多智能体协作平台 - 一个基于 Web 的系统，用于编排多个 AI 智能体（Claude Code、Gemini CLI、Aider 等）并行运行，支持实时聊天和工作区管理。

[English](README.md)

## 致谢

本项目的设计思路参考了 [golutra](https://github.com/golutra/golutra)。特别感谢该项目及其作者 [seeksky](https://github.com/seekskyworld) 的启发。本项目所有代码均为独立实现。

## 功能特性

- **多智能体终端管理**：并行运行多个 AI 智能体终端，每个终端拥有独立的 PTY 会话
- **实时协作聊天**：内置聊天界面，支持 @提及 功能定向发送消息给特定成员
- **工作区管理**：创建和切换多个工作区，可配置服务器端路径
- **成员角色**：基于角色的权限控制（Owner、Admin、Secretary、Assistant、Member）
- **路径浏览器**：浏览并选择服务器端目录绑定到工作区
- **现代柔光玻璃风格 UI**：基于 Vue 3、TypeScript 和 Tailwind CSS 构建的简洁现代界面
- **WebSocket 终端流**：通过 WebSocket 实时传输终端输出，支持 ANSI 颜色
- **国际化支持**：支持英文和中文

## 技术栈

### 后端
- **Go 1.21+** - 核心运行时
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
- Go 1.21+
- Node.js 18+（推荐使用 pnpm）

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
cd backend
make reset-data  # 删除数据库和 WAL 文件

# 可选：清除浏览器 localStorage
# DevTools → Application → Clear site data
# 或删除键：orchestra-settings, orchestra.auth, orchestra.user
```

### 代码更新后

拉取代码变更后重启两个开发进程：

1. **后端**：停止（`Ctrl+C`）→ `make run`
2. **前端**：停止（`Ctrl+C`）→ `pnpm dev`

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
| `/api/ping` | GET | 测试连接 |
| `/api/workspaces` | GET | 列出所有工作区 |
| `/api/workspaces` | POST | 创建工作区 |
| `/api/workspaces/:id` | GET | 获取工作区详情 |
| `/api/workspaces/:id/members` | GET | 列出工作区成员 |
| `/api/workspaces/:id/members` | POST | 添加成员到工作区 |
| `/api/workspaces/:id/members/:mid` | PUT | 更新成员 |
| `/api/workspaces/:id/members/:mid` | DELETE | 移除成员 |
| `/api/terminals` | POST | 创建终端会话 |
| `/api/terminals/:id` | DELETE | 关闭终端会话 |
| `/api/browse` | GET | 浏览服务器路径 |
| `/api/conversations/:workspaceId` | GET | 获取聊天历史 |
| `/api/conversations/:workspaceId/messages` | POST | 发送聊天消息 |

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
# 后端单元测试
cd backend && make test

# 前端单元测试
cd frontend && pnpm test

# E2E 测试（需要运行后端）
cd frontend && pnpm test:e2e

# E2E 使用自定义后端 URL
ORCHESTRA_API_URL=http://your-server:8080 pnpm test:e2e
```

### Make 命令

```bash
# 后端 Makefile 目标
make build      # 构建二进制
make run        # 运行服务器
make test       # 运行测试
make reset-data # 清理数据库
make clean      # 删除构建产物
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