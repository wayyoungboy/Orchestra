# Orchestra 系统功能测试方案

**版本**: v1.0  
**日期**: 2026-04-06  
**编写人**: Claude Code  
**项目**: Orchestra 多智能体协作平台

---

## 文档修订历史

| 版本 | 日期 | 修订人 | 修订内容 |
|------|------|--------|----------|
| v1.0 | 2026-04-06 | Claude Code | 初版创建 |

---

## 1. 引言

### 1.1 编写目的

本文档旨在为 Orchestra 多智能体协作平台提供完整的功能测试方案，确保系统各模块功能正常、接口稳定、用户体验良好。测试结果将作为项目验收的重要依据。

### 1.2 测试范围

本测试方案覆盖以下功能模块：

| 模块编号 | 模块名称 | 测试范围 |
|----------|----------|----------|
| M01 | 认证授权 | 用户登录、注册、Token验证、权限控制 |
| M02 | 工作区管理 | 工作区CRUD、路径绑定、文件浏览 |
| M03 | 成员管理 | 成员CRUD、角色类型、在线状态 |
| M04 | 终端管理 | PTY会话、WebSocket通信、进程池 |
| M05 | 对话系统 | 对话CRUD、消息发送、已读状态 |
| M06 | 实时通信 | WebSocket连接、消息推送、状态同步 |
| M07 | 秘书协调 | 任务创建、状态追踪、负载查询 |
| M08 | 文件附件 | 上传、下载、删除、元信息 |
| M09 | 搜索功能 | 消息全文搜索、结果过滤 |
| M10 | 前端界面 | 页面渲染、交互响应、状态管理 |

### 1.3 参考文档

- `docs/superpowers/specs/2026-03-29-orchestra-design.md` - 设计规范
- `backend/internal/api/router.go` - API路由定义
- `frontend/src/app/router.ts` - 前端路由定义
- `CLAUDE.md` - 项目说明

### 1.4 术语定义

| 术语 | 定义 |
|------|------|
| 工作区(Workspace) | 绑定服务器路径的项目空间 |
| 成员(Member) | 工作区内的用户或AI助手 |
| 秘书(Secretary) | 协调者角色，负责任务分配 |
| 助手(Assistant) | 执行者角色，负责完成任务 |
| 终端会话(Terminal Session) | PTY进程实例 |
| 对话(Conversation) | 成员间的消息交流通道 |

---

## 2. 测试环境

### 2.1 硬件环境

| 项目 | 规格 |
|------|------|
| CPU | Apple Silicon (M系列) |
| 内存 | 16GB+ |
| 硬盘 | 100GB+ 可用空间 |
| 网络 | 本地回环 |

### 2.2 软件环境

| 软件 | 版本要求 |
|------|----------|
| 操作系统 | macOS 14.0+ / Linux |
| Go | 1.21+ |
| Node.js | 18.0+ |
| pnpm | 8.0+ |
| SQLite | 3.40+ |

### 2.3 测试工具

| 工具 | 用途 |
|------|------|
| curl | HTTP API测试 |
| jq | JSON解析 |
| wscat | WebSocket测试 |
| Chrome DevTools | 前端调试 |
| SQLite CLI | 数据库验证 |

### 2.4 环境配置

```bash
# 后端配置
cd backend
cp configs/config.yaml.example configs/config.yaml
make reset-data
make run

# 前端配置
cd frontend
pnpm install
pnpm dev
```

---

## 3. 测试用例

### 3.1 M01 认证授权模块

#### TC-M01-001 获取认证配置（未启用认证）

**前置条件**: 服务启动，认证未启用

**测试步骤**:
```bash
curl -s http://127.0.0.1:8080/api/auth/config | jq
```

**预期结果**:
```json
{
  "enabled": false,
  "allowRegistration": false
}
```

**测试类型**: 正常流程

---

#### TC-M01-002 用户登录（认证启用时）

**前置条件**: 认证已启用，存在用户 orchestra/orchestra

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"orchestra","password":"orchestra"}' | jq
```

**预期结果**:
```json
{
  "token": "eyJ...",
  "user": {
    "id": "...",
    "username": "orchestra"
  }
}
```

**测试类型**: 正常流程

---

#### TC-M01-003 登录失败（错误密码）

**前置条件**: 认证已启用

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"orchestra","password":"wrongpassword"}' | jq
```

**预期结果**:
```json
{
  "error": "invalid credentials"
}
```

**测试类型**: 异常流程

---

#### TC-M01-004 Token验证

**前置条件**: 已获取有效Token

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/auth/validate \
  -H "Content-Type: application/json" \
  -d '{"token":"eyJ..."}' | jq
```

**预期结果**:
```json
{
  "valid": true,
  "userId": "...",
  "username": "orchestra"
}
```

**测试类型**: 正常流程

---

### 3.2 M02 工作区管理模块

#### TC-M02-001 列出工作区

**前置条件**: 服务启动

**测试步骤**:
```bash
curl -s http://127.0.0.1:8080/api/workspaces | jq
```

**预期结果**:
```json
[
  {
    "id": "...",
    "name": "工作区名称",
    "path": "/path/to/workspace",
    "lastOpenedAt": "2026-04-06T...",
    "createdAt": "2026-04-06T..."
  }
]
```

**测试类型**: 正常流程

---

#### TC-M02-002 创建工作区（有效路径）

**前置条件**: 路径存在且在允许列表中

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/workspaces \
  -H "Content-Type: application/json" \
  -d '{"name":"测试工作区","path":"/Users/wangxuyan/projects/test"}' | jq
```

**预期结果**:
```json
{
  "id": "...",
  "name": "测试工作区",
  "path": "/Users/wangxuyan/projects/test",
  "lastOpenedAt": "...",
  "createdAt": "..."
}
```

**测试类型**: 正常流程

---

#### TC-M02-003 创建工作区（无效路径）

**前置条件**: 路径不存在或不在允许列表

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/workspaces \
  -H "Content-Type: application/json" \
  -d '{"name":"测试工作区","path":"/invalid/path"}' | jq
```

**预期结果**:
```json
{
  "error": "path does not exist"
}
```
或
```json
{
  "error": "path not allowed"
}
```

**测试类型**: 异常流程

---

#### TC-M02-004 获取工作区详情

**前置条件**: 工作区已创建

**测试步骤**:
```bash
curl -s http://127.0.0.1:8080/api/workspaces/{workspaceId} | jq
```

**预期结果**: 返回完整工作区信息

**测试类型**: 正常流程

---

#### TC-M02-005 更新工作区

**前置条件**: 工作区已创建

**测试步骤**:
```bash
curl -s -X PUT http://127.0.0.1:8080/api/workspaces/{workspaceId} \
  -H "Content-Type: application/json" \
  -d '{"name":"新名称"}' | jq
```

**预期结果**: 返回更新后的工作区信息，name已更新

**测试类型**: 正常流程

---

#### TC-M02-006 删除工作区

**前置条件**: 工作区已创建

**测试步骤**:
```bash
curl -s -X DELETE http://127.0.0.1:8080/api/workspaces/{workspaceId}
```

**预期结果**: HTTP 204 No Content

**测试类型**: 正常流程

---

#### TC-M02-007 浏览工作区文件

**前置条件**: 工作区已创建

**测试步骤**:
```bash
curl -s "http://127.0.0.1:8080/api/workspaces/{workspaceId}/browse" | jq
```

**预期结果**:
```json
{
  "basePath": "/Users/wangxuyan/projects/test",
  "files": [
    {
      "name": "...",
      "isDir": true/false,
      "size": ...,
      "modTime": "..."
    }
  ]
}
```

**测试类型**: 正常流程

---

#### TC-M02-008 搜索消息

**前置条件**: 工作区内有对话和消息

**测试步骤**:
```bash
curl -s "http://127.0.0.1:8080/api/workspaces/{workspaceId}/search?q=测试" | jq
```

**预期结果**:
```json
{
  "query": "测试",
  "count": 5,
  "results": [...]
}
```

**测试类型**: 正常流程

---

### 3.3 M03 成员管理模块

#### TC-M03-001 列出成员

**前置条件**: 工作区已创建

**测试步骤**:
```bash
curl -s http://127.0.0.1:8080/api/workspaces/{workspaceId}/members | jq
```

**预期结果**: 返回成员列表，包含自动创建的Owner

**测试类型**: 正常流程

---

#### TC-M03-002 创建秘书成员

**前置条件**: 工作区已创建

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/workspaces/{workspaceId}/members \
  -H "Content-Type: application/json" \
  -d '{"name":"Secretary","roleType":"secretary","program":"claude"}' | jq
```

**预期结果**:
```json
{
  "id": "...",
  "workspaceId": "...",
  "name": "Secretary",
  "roleType": "secretary",
  "autoStartTerminal": true,
  "status": "online",
  "createdAt": "..."
}
```

**测试类型**: 正常流程

---

#### TC-M03-003 创建助手成员

**前置条件**: 工作区已创建

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/workspaces/{workspaceId}/members \
  -H "Content-Type: application/json" \
  -d '{"name":"Assistant-1","roleType":"assistant","program":"claude"}' | jq
```

**预期结果**: 返回新创建的助手成员

**测试类型**: 正常流程

---

#### TC-M03-004 更新成员

**前置条件**: 成员已创建

**测试步骤**:
```bash
curl -s -X PUT http://127.0.0.1:8080/api/workspaces/{workspaceId}/members/{memberId} \
  -H "Content-Type: application/json" \
  -d '{"name":"新名称"}' | jq
```

**预期结果**: 返回更新后的成员信息

**测试类型**: 正常流程

---

#### TC-M03-005 删除成员

**前置条件**: 成员已创建

**测试步骤**:
```bash
curl -s -X DELETE http://127.0.0.1:8080/api/workspaces/{workspaceId}/members/{memberId}
```

**预期结果**: HTTP 204 No Content

**测试类型**: 正常流程

---

#### TC-M03-006 更新在线状态

**前置条件**: 成员已创建

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/workspaces/{workspaceId}/members/{memberId}/presence \
  -H "Content-Type: application/json" \
  -d '{"status":"online"}' | jq
```

**预期结果**: HTTP 200 OK

**测试类型**: 正常流程

---

### 3.4 M04 终端管理模块

#### TC-M04-001 创建终端会话

**前置条件**: 工作区已创建

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/terminals \
  -H "Content-Type: application/json" \
  -d '{"workspaceId":"{workspaceId}","command":"claude"}' | jq
```

**预期结果**:
```json
{
  "sessionId": "...",
  "status": "created"
}
```

**测试类型**: 正常流程

---

#### TC-M04-002 创建不允许的命令会话

**前置条件**: 命令不在白名单

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/terminals \
  -H "Content-Type: application/json" \
  -d '{"workspaceId":"{workspaceId}","command":"rm"}' | jq
```

**预期结果**:
```json
{
  "error": "command not allowed"
}
```

**测试类型**: 异常流程

---

#### TC-M04-003 列出工作区终端会话

**前置条件**: 工作区有终端会话

**测试步骤**:
```bash
curl -s http://127.0.0.1:8080/api/workspaces/{workspaceId}/terminal-sessions | jq
```

**预期结果**: 返回会话列表

**测试类型**: 正常流程

---

#### TC-M04-004 删除终端会话

**前置条件**: 终端会话已创建

**测试步骤**:
```bash
curl -s -X DELETE http://127.0.0.1:8080/api/terminals/{sessionId}
```

**预期结果**: HTTP 204 No Content

**测试类型**: 正常流程

---

### 3.5 M05 对话系统模块

#### TC-M05-001 列出对话

**前置条件**: 工作区已创建

**测试步骤**:
```bash
curl -s http://127.0.0.1:8080/api/workspaces/{workspaceId}/conversations | jq
```

**预期结果**: 返回对话列表

**测试类型**: 正常流程

---

#### TC-M05-002 创建对话

**前置条件**: 工作区已创建

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/workspaces/{workspaceId}/conversations \
  -H "Content-Type: application/json" \
  -d '{"title":"新对话","memberIds":["{memberId1}","{memberId2}"]}' | jq
```

**预期结果**: 返回新创建的对话

**测试类型**: 正常流程

---

#### TC-M05-003 获取对话详情

**前置条件**: 对话已创建

**测试步骤**:
```bash
curl -s http://127.0.0.1:8080/api/workspaces/{workspaceId}/conversations/{convId} | jq
```

**预期结果**: 返回对话详情

**测试类型**: 正常流程

---

#### TC-M05-004 删除对话

**前置条件**: 对话已创建

**测试步骤**:
```bash
curl -s -X DELETE http://127.0.0.1:8080/api/workspaces/{workspaceId}/conversations/{convId}
```

**预期结果**: HTTP 204 No Content

**测试类型**: 正常流程

---

#### TC-M05-005 获取消息列表

**前置条件**: 对话已创建

**测试步骤**:
```bash
curl -s "http://127.0.0.1:8080/api/workspaces/{workspaceId}/conversations/{convId}/messages" | jq
```

**预期结果**: 返回消息列表

**测试类型**: 正常流程

---

#### TC-M05-006 发送消息

**前置条件**: 对话已创建，成员已绑定终端

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/workspaces/{workspaceId}/conversations/{convId}/messages \
  -H "Content-Type: application/json" \
  -d '{"senderId":"{memberId}","senderName":"Secretary","text":"你好"}' | jq
```

**预期结果**: 消息发送成功

**测试类型**: 正常流程

---

#### TC-M05-007 标记已读

**前置条件**: 对话有消息

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/workspaces/{workspaceId}/conversations/{convId}/read \
  -H "Content-Type: application/json" \
  -d '{"memberId":"{memberId}"}' | jq
```

**预期结果**: HTTP 200 OK

**测试类型**: 正常流程

---

### 3.6 M07 秘书协调模块

#### TC-M07-001 创建任务

**前置条件**: 工作区、秘书、助手、对话已创建

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/internal/tasks/create \
  -H "Content-Type: application/json" \
  -d '{
    "workspaceId":"{workspaceId}",
    "conversationId":"{convId}",
    "secretaryId":"{secretaryId}",
    "title":"实现登录功能",
    "description":"创建登录页面和API",
    "assigneeId":"{assistantId}"
  }' | jq
```

**预期结果**:
```json
{
  "ok": true,
  "taskId": "task_...",
  "task": {
    "id": "task_...",
    "status": "assigned",
    ...
  }
}
```

**测试类型**: 正常流程

---

#### TC-M07-002 开始任务

**前置条件**: 任务已创建

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/internal/tasks/start \
  -H "Content-Type: application/json" \
  -d '{"taskId":"{taskId}"}' | jq
```

**预期结果**:
```json
{
  "ok": true,
  "taskId": "...",
  "status": "in_progress",
  "conversationId": "...",
  "secretaryId": "..."
}
```

**测试类型**: 正常流程

---

#### TC-M07-003 完成任务

**前置条件**: 任务已开始

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/internal/tasks/complete \
  -H "Content-Type: application/json" \
  -d '{"taskId":"{taskId}","resultSummary":"任务完成"}' | jq
```

**预期结果**:
```json
{
  "ok": true,
  "status": "completed",
  ...
}
```

**测试类型**: 正常流程

---

#### TC-M07-004 任务失败

**前置条件**: 任务已开始

**测试步骤**:
```bash
curl -s -X POST http://127.0.0.1:8080/api/internal/tasks/fail \
  -H "Content-Type: application/json" \
  -d '{"taskId":"{taskId}","errorMessage":"遇到错误"}' | jq
```

**预期结果**:
```json
{
  "ok": true,
  "status": "failed",
  ...
}
```

**测试类型**: 正常流程

---

#### TC-M07-005 查询负载

**前置条件**: 工作区有助手成员

**测试步骤**:
```bash
curl -s "http://127.0.0.1:8080/api/internal/workloads/list?workspaceId={workspaceId}" | jq
```

**预期结果**:
```json
{
  "ok": true,
  "workloads": [
    {
      "memberId": "...",
      "name": "Assistant-1",
      "currentTaskCount": 0,
      "pendingTaskCount": 0,
      "completedTaskCount": 1,
      "status": "idle"
    }
  ]
}
```

**测试类型**: 正常流程

---

#### TC-M07-006 列出任务

**前置条件**: 工作区有任务

**测试步骤**:
```bash
curl -s "http://127.0.0.1:8080/api/workspaces/{workspaceId}/tasks" | jq
```

**预期结果**: 返回任务列表

**测试类型**: 正常流程

---

#### TC-M07-007 获取任务详情

**前置条件**: 任务已创建

**测试步骤**:
```bash
curl -s "http://127.0.0.1:8080/api/workspaces/{workspaceId}/tasks/{taskId}" | jq
```

**预期结果**: 返回任务详情，包含状态变更历史

**测试类型**: 正常流程

---

#### TC-M07-008 查询成员任务

**前置条件**: 成员有分配的任务

**测试步骤**:
```bash
curl -s "http://127.0.0.1:8080/api/workspaces/{workspaceId}/tasks/my-tasks?memberId={memberId}" | jq
```

**预期结果**: 返回该成员的所有任务

**测试类型**: 正常流程

---

### 3.7 M08 文件附件模块

#### TC-M08-001 上传附件

**前置条件**: 对话已创建

**测试步骤**:
```bash
echo "test content" > /tmp/test.txt
curl -s -X POST "http://127.0.0.1:8080/api/workspaces/{workspaceId}/conversations/{convId}/attachments" \
  -F "file=@/tmp/test.txt" \
  -F "uploadedBy={memberId}" | jq
```

**预期结果**: 返回附件信息

**测试类型**: 正常流程

---

#### TC-M08-002 列出附件

**前置条件**: 工作区有附件

**测试步骤**:
```bash
curl -s "http://127.0.0.1:8080/api/workspaces/{workspaceId}/attachments" | jq
```

**预期结果**: 返回附件列表

**测试类型**: 正常流程

---

#### TC-M08-003 下载附件

**前置条件**: 附件已上传

**测试步骤**:
```bash
curl -s "http://127.0.0.1:8080/api/workspaces/{workspaceId}/attachments/{attachmentId}" -o /tmp/downloaded.txt
cat /tmp/downloaded.txt
```

**预期结果**: 文件内容与上传一致

**测试类型**: 正常流程

---

#### TC-M08-004 删除附件

**前置条件**: 附件已上传

**测试步骤**:
```bash
curl -s -X DELETE "http://127.0.0.1:8080/api/workspaces/{workspaceId}/attachments/{attachmentId}"
```

**预期结果**: HTTP 204 No Content

**测试类型**: 正常流程

---

### 3.8 M10 前端界面模块

#### TC-M10-001 页面加载

**前置条件**: 前后端服务运行

**测试步骤**:
1. 打开浏览器访问 http://localhost:5173
2. 检查页面是否正常加载

**预期结果**: 显示工作区选择页面，无控制台错误

**测试类型**: UI测试

---

#### TC-M10-002 工作区选择

**前置条件**: 已创建工作区

**测试步骤**:
1. 查看工作区列表
2. 点击工作区进入

**预期结果**: 显示工作区卡片，点击后进入工作区主页

**测试类型**: UI测试

---

#### TC-M10-003 成员管理页面

**前置条件**: 在工作区内

**测试步骤**:
1. 点击侧边栏"成员"
2. 查看成员列表
3. 点击"添加成员"

**预期结果**: 显示成员列表，可打开添加成员对话框

**测试类型**: UI测试

---

#### TC-M10-004 对话界面

**前置条件**: 在工作区内，有成员

**测试步骤**:
1. 点击侧边栏"对话"
2. 选择或创建对话
3. 在输入框输入消息

**预期结果**: 显示对话列表和消息历史，可发送消息

**测试类型**: UI测试

---

#### TC-M10-005 终端界面

**前置条件**: 在工作区内，成员已启动终端

**测试步骤**:
1. 点击侧边栏"终端"
2. 查看终端列表
3. 选择终端查看

**预期结果**: 显示终端列表，可查看终端输出

**测试类型**: UI测试

---

## 4. 测试执行记录

### 4.1 测试执行汇总表

| 模块 | 用例数 | 通过 | 失败 | 阻塞 | 通过率 |
|------|--------|------|------|------|--------|
| M01 认证授权 | 4 | - | - | - | - |
| M02 工作区管理 | 8 | - | - | - | - |
| M03 成员管理 | 6 | - | - | - | - |
| M04 终端管理 | 4 | - | - | - | - |
| M05 对话系统 | 7 | - | - | - | - |
| M07 秘书协调 | 8 | - | - | - | - |
| M08 文件附件 | 4 | - | - | - | - |
| M10 前端界面 | 5 | - | - | - | - |
| **总计** | **46** | - | - | - | - |

### 4.2 缺陷记录表

| 缺陷编号 | 模块 | 描述 | 严重程度 | 状态 |
|----------|------|------|----------|------|
| - | - | - | - | - |

---

## 5. 验收标准

### 5.1 功能验收标准

| 等级 | 标准 |
|------|------|
| 通过 | 所有 P0 功能测试通过率 ≥ 95%，无阻塞性缺陷 |
| 有条件通过 | P0 功能测试通过率 ≥ 80%，存在可接受的小问题 |
| 不通过 | P0 功能测试通过率 < 80%，或存在阻塞性缺陷 |

### 5.2 性能验收标准

| 指标 | 标准 |
|------|------|
| API 响应时间 | P95 < 500ms |
| 页面加载时间 | < 3s |
| WebSocket 连接 | 稳定不中断 |

### 5.3 安全验收标准

| 指标 | 标准 |
|------|------|
| 路径遍历 | 已防护 |
| 命令注入 | 白名单控制 |
| SQL注入 | 参数化查询 |

---

## 6. 附录

### 6.1 测试数据准备脚本

```bash
#!/bin/bash
# prepare_test_data.sh

API="http://127.0.0.1:8080"

# 创建测试目录
mkdir -p /Users/wangxuyan/projects/test

# 创建工作区
WS=$(curl -s -X POST $API/api/workspaces \
  -H "Content-Type: application/json" \
  -d '{"name":"测试工作区","path":"/Users/wangxuyan/projects/test"}')
WS_ID=$(echo $WS | jq -r '.id')
echo "Workspace: $WS_ID"

# 创建秘书
SEC=$(curl -s -X POST $API/api/workspaces/$WS_ID/members \
  -H "Content-Type: application/json" \
  -d '{"name":"Secretary","roleType":"secretary","program":"claude"}')
SEC_ID=$(echo $SEC | jq -r '.id')
echo "Secretary: $SEC_ID"

# 创建助手
ASS=$(curl -s -X POST $API/api/workspaces/$WS_ID/members \
  -H "Content-Type: application/json" \
  -d '{"name":"Assistant-1","roleType":"assistant","program":"claude"}')
ASS_ID=$(echo $ASS | jq -r '.id')
echo "Assistant: $ASS_ID"

# 创建对话
CONV=$(curl -s -X POST $API/api/workspaces/$WS_ID/conversations \
  -H "Content-Type: application/json" \
  -d '{"title":"测试对话"}')
CONV_ID=$(echo $CONV | jq -r '.id')
echo "Conversation: $CONV_ID"

echo "Test data prepared successfully!"
echo "WS_ID=$WS_ID"
echo "SEC_ID=$SEC_ID"
echo "ASS_ID=$ASS_ID"
echo "CONV_ID=$CONV_ID"
```

### 6.2 完整测试执行脚本

```bash
#!/bin/bash
# run_all_tests.sh

set -e

API="http://127.0.0.1:8080"

echo "=== Orchestra 系统功能测试 ==="
echo "开始时间: $(date)"
echo ""

# TC-M01-001
echo ">>> TC-M01-001: 获取认证配置"
curl -s $API/api/auth/config | jq

# TC-M02-001
echo ">>> TC-M02-001: 列出工作区"
curl -s $API/api/workspaces | jq

# ... 更多测试用例

echo ""
echo "=== 测试完成 ==="
echo "结束时间: $(date)"
```