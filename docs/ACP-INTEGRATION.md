# ACP (Agent Communication Protocol) 集成文档

## 概述

Orchestra 使用 ACP（Agent Communication Protocol）进行 AI 代理通信。ACP 是一种基于 stdin/stdout JSON 流的结构化通信协议，用于与 AI 代理（如 Claude Code、Gemini CLI）进行可靠的消息交换。

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│                     Orchestra ACP 架构                       │
├─────────────────────────────────────────────────────────────┤
│  用户消息 → JSON stdin → AI处理 → JSON stdout → 消息解析 → 聊天│
│                    ↓                                         │
│              工具调用（原生支持）                              │
└─────────────────────────────────────────────────────────────┘
```

## 核心组件

### 后端 (`backend/internal/acp/`)

| 文件 | 说明 |
|------|------|
| `messages.go` | ACP 消息类型定义 |
| `parser.go` | JSON 流解析器 |
| `session.go` | ACP 会话管理 |
| `pool.go` | 会话池管理 |
| `tools.go` | Orchestra 工具定义 |
| `handler.go` | 工具执行处理器 |

### 前端 (`frontend/src/shared/types/`)

| 文件 | 说明 |
|------|------|
| `acp.ts` | ACP TypeScript 类型定义 |

## ACP 消息格式

### 输入消息（发送给 AI）

```json
{"type": "user_message", "content": "你好"}
```

### 输出消息（从 AI 接收）

```json
// AI 回复
{"type": "assistant_message", "content": "你好！有什么可以帮你的？"}

// 工具调用
{"type": "tool_use", "name": "orchestra_chat_send", "tool_use_id": "xxx", "input": {...}}

// 完成通知
{"type": "result", "message": "完成", "cost_usd": 0.01}

// 错误
{"type": "error", "error": "错误信息"}
```

## Orchestra 工具

AI 可以调用以下原生工具：

| 工具名 | 说明 |
|--------|------|
| `orchestra_chat_send` | 发送消息到对话 |
| `orchestra_task_create` | 创建任务（秘书角色） |
| `orchestra_task_start` | 开始任务 |
| `orchestra_task_complete` | 完成任务 |
| `orchestra_task_fail` | 任务失败报告 |
| `orchestra_workload_list` | 查询助手负载（秘书角色） |
| `orchestra_agent_status` | 更新活动状态 |

## 配置

### 成员配置

在创建或更新成员时配置 ACP：

```json
{
  "name": "Claude-Agent",
  "roleType": "assistant",
  "acpEnabled": true,
  "acpCommand": "claude",
  "acpArgs": ["--input-format", "stream-json", "--output-format", "stream-json"]
}
```

### 数据库迁移

```sql
ALTER TABLE members ADD COLUMN acp_enabled INTEGER DEFAULT 0;
ALTER TABLE members ADD COLUMN acp_command TEXT;
ALTER TABLE members ADD COLUMN acp_args TEXT;
```

## WebSocket 协议

### 连接

```
ws://localhost:8080/ws/terminal/{sessionId}
```

ACP 会话 ID 以 "acp-" 开头。

### 消息格式

**客户端 → 服务端：**

```json
{"type": "user_message", "content": "你好"}
{"type": "tool_result", "tool_use_id": "xxx", "tool_result": "结果", "is_error": false}
{"type": "close"}
```

**服务端 → 客户端：**

```json
{"type": "connected", "sessionId": "acp-xxx"}
{"type": "assistant_message", "content": "回复内容"}
{"type": "tool_use", "tool_name": "orchestra_chat_send", "tool_use_id": "xxx", "tool_input": {}}
{"type": "result", "message": "完成"}
{"type": "error", "error": "错误"}
{"type": "exit", "code": 0}
```

## 测试

```bash
# 编译后端
cd backend && go build ./...

# 运行服务
make run

# 创建 ACP 成员进行测试
curl -X POST http://localhost:8080/api/workspaces/{id}/members \
  -H "Content-Type: application/json" \
  -d '{"name":"Claude-ACP","roleType":"assistant","acpEnabled":true,"acpCommand":"claude","acpArgs":["--input-format","stream-json","--output-format","stream-json"]}'
```

## 参考

- Claude Code 文档：`claude --input-format stream-json --output-format stream-json`
- Gemini CLI 文档：`gemini --input-format stream-json --output-format stream-json`