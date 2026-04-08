# ACP 协议集成方案设计

## 一、概述

### 1.1 目标
将 Orchestra 的终端通信从纯 PTY 方案升级为 **PTY + ACP 混合方案**，支持两种通信协议：
- **PTY**: 兼容现有 CLI 工具，处理不支持 ACP 的场景
- **ACP**: 结构化 JSON-RPC 通信，提供更好的工具调用和状态管理

### 1.2 架构变化

```
当前架构:
┌─────────┐    WebSocket    ┌─────────┐    PTY    ┌─────────┐
│ 前端    │ ←──────────────→ │ 后端    │ ←───────→ │ CLI工具 │
└─────────┘                 └─────────┘           └─────────┘

新架构:
┌─────────┐    WebSocket    ┌─────────────────────┐
│ 前端    │ ←──────────────→ │       后端          │
└─────────┘                 │  ┌───────────────┐  │
                            │  │ SessionManager│  │
                            │  └───────┬───────┘  │
                            │          │          │
                            │   ┌──────┴──────┐   │
                            │   ↓             ↓   │
                            │ ┌────┐      ┌─────┐│
                            │ │PTY │      │ ACP ││
                            │ └─┬──┘      └──┬──┘│
                            └───┼────────────┼───┘
                                ↓            ↓
                            ┌───────┐   ┌───────┐
                            │旧CLI  │   │新CLI  │
                            └───────┘   └───────┘
```

---

## 二、ACP 协议规范

### 2.1 基础消息格式

基于 JSON-RPC 2.0:

```go
// 请求
type ACPRequest struct {
    JSONRPC string         `json:"jsonrpc"`          // "2.0"
    ID      string         `json:"id"`               // 请求ID
    Method  string         `json:"method"`           // 方法名
    Params  map[string]any `json:"params,omitempty"` // 参数
}

// 响应
type ACPResponse struct {
    JSONRPC string         `json:"jsonrpc"`
    ID      string         `json:"id"`
    Result  map[string]any `json:"result,omitempty"`
    Error   *ACPError      `json:"error,omitempty"`
}

// 错误
type ACPError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
}
```

### 2.2 支持的方法

| 方法 | 说明 |
|------|------|
| `initialize` | 初始化会话，协商能力 |
| `message/create` | 创建消息 |
| `message/stream` | 流式消息（SSE风格） |
| `tools/list` | 列出可用工具 |
| `tools/call` | 调用工具 |
| `tasks/create` | 创建任务 |
| `tasks/update` | 更新任务状态 |
| `status/update` | 更新状态 |

### 2.3 消息类型

```go
type ContentBlock struct {
    Type string `json:"type"` // "text", "image", "tool_use", "tool_result"
    
    // text 类型
    Text string `json:"text,omitempty"`
    
    // image 类型
    Source *ImageSource `json:"source,omitempty"`
    
    // tool_use 类型
    ToolName string         `json:"name,omitempty"`
    Input    map[string]any `json:"input,omitempty"`
    ID       string         `json:"id,omitempty"`
    
    // tool_result 类型
    ToolUseID string `json:"tool_use_id,omitempty"`
    Content   any    `json:"content,omitempty"`
    IsError   bool   `json:"is_error,omitempty"`
}
```

---

## 三、后端实现设计

### 3.1 会话接口抽象

```go
// internal/terminal/session.go

package terminal

import (
    "context"
    "io"
)

// ProtocolType 协议类型
type ProtocolType string

const (
    ProtocolPTY ProtocolType = "pty"
    ProtocolACP ProtocolType = "acp"
)

// Session 会话接口
type Session interface {
    // 基本信息
    ID() string
    Protocol() ProtocolType
    PID() int
    
    // 通信
    Write(data []byte) error
    Read() ([]byte, error)
    Close() error
    
    // 状态
    IsRunning() bool
    Wait() error
}

// ACPSession ACP协议会话
type ACPSession struct {
    id       string
    cmd      *exec.Cmd
    stdin    io.WriteCloser
    stdout   io.Reader
    stderr   io.Reader
    encoder  *json.Encoder
    decoder  *json.Decoder
    running  bool
    mu       sync.RWMutex
    
    // 消息处理
    pendingRequests map[string]chan *ACPResponse
    responseHandler func(*ACPResponse)
}

// NewACPSession 创建ACP会话
func NewACPSession(ctx context.Context, command string, args []string, dir string) (*ACPSession, error) {
    cmd := exec.CommandContext(ctx, command, args...)
    cmd.Dir = dir
    
    stdin, err := cmd.StdinPipe()
    if err != nil {
        return nil, fmt.Errorf("create stdin pipe: %w", err)
    }
    
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return nil, fmt.Errorf("create stdout pipe: %w", err)
    }
    
    stderr, err := cmd.StderrPipe()
    if err != nil {
        return nil, fmt.Errorf("create stderr pipe: %w", err)
    }
    
    if err := cmd.Start(); err != nil {
        return nil, fmt.Errorf("start command: %w", err)
    }
    
    session := &ACPSession{
        id:              generateSessionID(),
        cmd:             cmd,
        stdin:           stdin,
        stdout:          stdout,
        stderr:          stderr,
        encoder:         json.NewEncoder(stdin),
        decoder:         json.NewDecoder(stdout),
        pendingRequests: make(map[string]chan *ACPResponse),
        running:         true,
    }
    
    // 启动响应监听
    go session.listenResponses()
    go session.listenStderr()
    
    return session, nil
}

// Initialize 初始化会话
func (s *ACPSession) Initialize(ctx context.Context, params *InitializeParams) (*InitializeResult, error) {
    req := &ACPRequest{
        JSONRPC: "2.0",
        ID:      generateID(),
        Method:  "initialize",
        Params: map[string]any{
            "protocolVersion": "1.0.0",
            "clientInfo":      params.ClientInfo,
            "capabilities":    params.Capabilities,
        },
    }
    
    resp, err := s.sendRequest(req)
    if err != nil {
        return nil, err
    }
    
    if resp.Error != nil {
        return nil, fmt.Errorf("initialize failed: %s", resp.Error.Message)
    }
    
    var result InitializeResult
    if err := mapToStruct(resp.Result, &result); err != nil {
        return nil, err
    }
    
    return &result, nil
}

// SendMessage 发送消息
func (s *ACPSession) SendMessage(ctx context.Context, msg *Message) (*Message, error) {
    req := &ACPRequest{
        JSONRPC: "2.0",
        ID:      generateID(),
        Method:  "message/create",
        Params: map[string]any{
            "role":    msg.Role,
            "content": msg.Content,
        },
    }
    
    resp, err := s.sendRequest(req)
    if err != nil {
        return nil, err
    }
    
    if resp.Error != nil {
        return nil, fmt.Errorf("send message failed: %s", resp.Error.Message)
    }
    
    var result Message
    if err := mapToStruct(resp.Result, &result); err != nil {
        return nil, err
    }
    
    return &result, nil
}

// CallTool 调用工具
func (s *ACPSession) CallTool(ctx context.Context, name string, args map[string]any) (*ToolResult, error) {
    req := &ACPRequest{
        JSONRPC: "2.0",
        ID:      generateID(),
        Method:  "tools/call",
        Params: map[string]any{
            "name":      name,
            "arguments": args,
        },
    }
    
    resp, err := s.sendRequest(req)
    if err != nil {
        return nil, err
    }
    
    if resp.Error != nil {
        return nil, fmt.Errorf("tool call failed: %s", resp.Error.Message)
    }
    
    var result ToolResult
    if err := mapToStruct(resp.Result, &result); err != nil {
        return nil, err
    }
    
    return &result, nil
}

// sendRequest 发送请求并等待响应
func (s *ACPSession) sendRequest(req *ACPRequest) (*ACPResponse, error) {
    s.mu.Lock()
    ch := make(chan *ACPResponse, 1)
    s.pendingRequests[req.ID] = ch
    s.mu.Unlock()
    
    defer func() {
        s.mu.Lock()
        delete(s.pendingRequests, req.ID)
        s.mu.Unlock()
    }()
    
    if err := s.encoder.Encode(req); err != nil {
        return nil, fmt.Errorf("encode request: %w", err)
    }
    
    select {
    case resp := <-ch:
        return resp, nil
    case <-time.After(30 * time.Second):
        return nil, fmt.Errorf("request timeout")
    }
}

// listenResponses 监听响应
func (s *ACPSession) listenResponses() {
    for {
        var resp ACPResponse
        if err := s.decoder.Decode(&resp); err != nil {
            if err == io.EOF {
                s.mu.Lock()
                s.running = false
                s.mu.Unlock()
                return
            }
            log.Printf("decode response error: %v", err)
            continue
        }
        
        s.mu.RLock()
        ch, ok := s.pendingRequests[resp.ID]
        s.mu.RUnlock()
        
        if ok {
            ch <- &resp
        } else if s.responseHandler != nil {
            // 处理服务器主动推送的消息
            s.responseHandler(&resp)
        }
    }
}
```

### 3.2 会话管理器

```go
// internal/terminal/manager.go

package terminal

import (
    "context"
    "sync"
)

// SessionConfig 会话配置
type SessionConfig struct {
    WorkspaceID string
    MemberID    string
    Command     string
    Args        []string
    Dir         string
    Protocol    ProtocolType // "auto", "pty", "acp"
    Env         map[string]string
}

// SessionManager 会话管理器
type SessionManager struct {
    mu       sync.RWMutex
    sessions map[string]Session
    pool     *ProcessPool // 现有的PTY池
    
    // 协议检测器
    protocolDetector *ProtocolDetector
}

// NewSessionManager 创建会话管理器
func NewSessionManager(maxSessions int, idleTimeout time.Duration) *SessionManager {
    return &SessionManager{
        sessions:        make(map[string]Session),
        pool:           NewProcessPool(maxSessions, idleTimeout),
        protocolDetector: NewProtocolDetector(),
    }
}

// CreateSession 创建会话
func (m *SessionManager) CreateSession(ctx context.Context, config *SessionConfig) (Session, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    // 检查是否已存在
    if sess, ok := m.sessions[config.MemberID]; ok {
        return sess, nil
    }
    
    // 自动检测协议
    protocol := config.Protocol
    if protocol == "auto" || protocol == "" {
        protocol = m.protocolDetector.Detect(config.Command)
    }
    
    var sess Session
    var err error
    
    switch protocol {
    case ProtocolACP:
        sess, err = NewACPSession(ctx, config.Command, config.Args, config.Dir)
        if err != nil {
            // 降级到PTY
            log.Printf("ACP session failed, falling back to PTY: %v", err)
            sess, err = NewPTYSession(ctx, config.Command, config.Args, config.Dir)
        }
    case ProtocolPTY:
        sess, err = NewPTYSession(ctx, config.Command, config.Args, config.Dir)
    default:
        sess, err = NewPTYSession(ctx, config.Command, config.Args, config.Dir)
    }
    
    if err != nil {
        return nil, err
    }
    
    m.sessions[config.MemberID] = sess
    return sess, nil
}

// GetSession 获取会话
func (m *SessionManager) GetSession(memberID string) Session {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.sessions[memberID]
}

// DeleteSession 删除会话
func (m *SessionManager) DeleteSession(memberID string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    sess, ok := m.sessions[memberID]
    if !ok {
        return nil
    }
    
    if err := sess.Close(); err != nil {
        return err
    }
    
    delete(m.sessions, memberID)
    return nil
}
```

### 3.3 协议检测器

```go
// internal/terminal/protocol_detector.go

package terminal

import (
    "context"
    "os/exec"
    "strings"
    "time"
)

// ProtocolDetector 协议检测器
type ProtocolDetector struct {
    cache map[string]ProtocolType
}

// NewProtocolDetector 创建协议检测器
func NewProtocolDetector() *ProtocolDetector {
    return &ProtocolDetector{
        cache: make(map[string]ProtocolType),
    }
}

// Detect 检测命令支持的协议
func (d *ProtocolDetector) Detect(command string) ProtocolType {
    // 检查缓存
    if protocol, ok := d.cache[command]; ok {
        return protocol
    }
    
    // 已知的ACP支持命令
    acpSupported := map[string]bool{
        "claude":        false, // 当前不支持ACP
        "gemini":        false,
        "copilot":       false,
        "aider":         false,
        "cursor-agent":  false,
    }
    
    baseCmd := getBaseCommand(command)
    if supported, ok := acpSupported[baseCmd]; ok && supported {
        d.cache[command] = ProtocolACP
        return ProtocolACP
    }
    
    // 尝试检测 ACP 支持
    if d.tryDetectACP(command) {
        d.cache[command] = ProtocolACP
        return ProtocolACP
    }
    
    d.cache[command] = ProtocolPTY
    return ProtocolPTY
}

// tryDetectACP 尝试检测ACP支持
func (d *ProtocolDetector) tryDetectACP(command string) bool {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    cmd := exec.CommandContext(ctx, command, "--help")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return false
    }
    
    // 检查帮助信息中是否提到ACP
    helpText := strings.ToLower(string(output))
    if strings.Contains(helpText, "acp") ||
       strings.Contains(helpText, "agent-client-protocol") ||
       strings.Contains(helpText, "--acp") {
        return true
    }
    
    return false
}
```

---

## 四、前端适配

### 4.1 消息类型扩展

```typescript
// src/shared/types/acp.ts

export interface ACPMessage {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: ContentBlock[]
  createdAt: number
}

export interface ContentBlock {
  type: 'text' | 'image' | 'tool_use' | 'tool_result'
  
  // text
  text?: string
  
  // image
  source?: {
    type: 'base64'
    media_type: string
    data: string
  }
  
  // tool_use
  name?: string
  input?: Record<string, any>
  toolUseId?: string
  
  // tool_result
  toolUseId?: string
  content?: any
  isError?: boolean
}

export interface ToolDefinition {
  name: string
  description: string
  inputSchema: {
    type: 'object'
    properties: Record<string, any>
    required?: string[]
  }
}

export interface ACPClientOptions {
  workspaceId: string
  memberId: string
  onMessage?: (message: ACPMessage) => void
  onToolUse?: (toolName: string, input: any) => Promise<any>
  onStatusChange?: (status: 'idle' | 'thinking' | 'working') => void
}
```

### 4.2 ACP 客户端

```typescript
// src/shared/socket/acp.ts

export class ACPClient {
  private ws: WebSocket | null = null
  private pendingRequests: Map<string, {
    resolve: (value: any) => void
    reject: (error: Error) => void
  }> = new Map()
  
  constructor(private options: ACPClientOptions) {}
  
  async connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(
        `ws://${location.host}/ws/acp/${this.options.workspaceId}/${this.options.memberId}`
      )
      
      this.ws.onopen = () => resolve()
      this.ws.onerror = (e) => reject(e)
      
      this.ws.onmessage = (event) => {
        const response = JSON.parse(event.data)
        this.handleResponse(response)
      }
    })
  }
  
  async sendMessage(content: ContentBlock[]): Promise<ACPMessage> {
    const request = {
      jsonrpc: '2.0',
      id: this.generateId(),
      method: 'message/create',
      params: { content }
    }
    
    return this.sendRequest(request)
  }
  
  async callTool(name: string, args: Record<string, any>): Promise<any> {
    const request = {
      jsonrpc: '2.0',
      id: this.generateId(),
      method: 'tools/call',
      params: { name, arguments: args }
    }
    
    return this.sendRequest(request)
  }
  
  async listTools(): Promise<ToolDefinition[]> {
    const request = {
      jsonrpc: '2.0',
      id: this.generateId(),
      method: 'tools/list',
      params: {}
    }
    
    const result = await this.sendRequest(request)
    return result.tools
  }
  
  private async sendRequest(request: any): Promise<any> {
    return new Promise((resolve, reject) => {
      this.pendingRequests.set(request.id, { resolve, reject })
      this.ws?.send(JSON.stringify(request))
      
      // 超时处理
      setTimeout(() => {
        if (this.pendingRequests.has(request.id)) {
          this.pendingRequests.delete(request.id)
          reject(new Error('Request timeout'))
        }
      }, 30000)
    })
  }
  
  private handleResponse(response: any) {
    // 处理请求响应
    if (response.id && this.pendingRequests.has(response.id)) {
      const { resolve, reject } = this.pendingRequests.get(response.id)!
      this.pendingRequests.delete(response.id)
      
      if (response.error) {
        reject(new Error(response.error.message))
      } else {
        resolve(response.result)
      }
      return
    }
    
    // 处理服务器推送
    if (response.method === 'message/stream') {
      this.options.onMessage?.(response.params)
    }
  }
  
  private generateId(): string {
    return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }
  
  disconnect() {
    this.ws?.close()
    this.ws = null
  }
}
```

---

## 五、工具集成

### 5.1 内置工具定义

```go
// internal/terminal/tools.go

package terminal

// BuiltInTools 内置工具列表
var BuiltInTools = []ToolDefinition{
    {
        Name:        "create_task",
        Description: "创建任务分配给助手",
        InputSchema: ToolInputSchema{
            Type: "object",
            Properties: map[string]any{
                "title": map[string]any{
                    "type":        "string",
                    "description": "任务标题",
                },
                "description": map[string]any{
                    "type":        "string",
                    "description": "任务描述",
                },
                "assigneeId": map[string]any{
                    "type":        "string",
                    "description": "被分配者ID",
                },
                "priority": map[string]any{
                    "type":        "integer",
                    "description": "优先级",
                },
            },
            Required: []string{"title"},
        },
    },
    {
        Name:        "update_task_status",
        Description: "更新任务状态",
        InputSchema: ToolInputSchema{
            Type: "object",
            Properties: map[string]any{
                "taskId": map[string]any{
                    "type":        "string",
                    "description": "任务ID",
                },
                "status": map[string]any{
                    "type":        "string",
                    "enum":        []string{"in_progress", "completed", "failed"},
                    "description": "新状态",
                },
                "result": map[string]any{
                    "type":        "string",
                    "description": "结果摘要",
                },
            },
            Required: []string{"taskId", "status"},
        },
    },
    {
        Name:        "get_workload",
        Description: "获取助手负载信息",
        InputSchema: ToolInputSchema{
            Type: "object",
            Properties: map[string]any{
                "workspaceId": map[string]any{
                    "type":        "string",
                    "description": "工作区ID",
                },
            },
            Required: []string{},
        },
    },
    {
        Name:        "send_message",
        Description: "发送消息到对话",
        InputSchema: ToolInputSchema{
            Type: "object",
            Properties: map[string]any{
                "conversationId": map[string]any{
                    "type":        "string",
                    "description": "对话ID",
                },
                "content": map[string]any{
                    "type":        "string",
                    "description": "消息内容",
                },
            },
            Required: []string{"conversationId", "content"},
        },
    },
}
```

### 5.2 工具执行器

```go
// internal/terminal/tool_executor.go

package terminal

import (
    "context"
    "encoding/json"
)

// ToolExecutor 工具执行器
type ToolExecutor struct {
    taskRepo    repository.TaskRepository
    memberRepo  repository.MemberRepository
    convRepo    repository.ConversationRepository
    msgRepo     repository.MessageRepository
}

// Execute 执行工具
func (e *ToolExecutor) Execute(ctx context.Context, name string, args map[string]any) (any, error) {
    switch name {
    case "create_task":
        return e.createTask(ctx, args)
    case "update_task_status":
        return e.updateTaskStatus(ctx, args)
    case "get_workload":
        return e.getWorkload(ctx, args)
    case "send_message":
        return e.sendMessage(ctx, args)
    default:
        return nil, fmt.Errorf("unknown tool: %s", name)
    }
}

func (e *ToolExecutor) createTask(ctx context.Context, args map[string]any) (any, error) {
    task := &models.Task{
        ID:             generateTaskID(),
        WorkspaceID:    args["workspaceId"].(string),
        ConversationID: args["conversationId"].(string),
        SecretaryID:    args["secretaryId"].(string),
        Title:          args["title"].(string),
        Description:    getString(args, "description"),
        Status:         models.TaskStatusPending,
        Priority:       getInt(args, "priority"),
        CreatedAt:      time.Now().UnixMilli(),
        UpdatedAt:      time.Now().UnixMilli(),
    }
    
    if assigneeID, ok := args["assigneeId"].(string); ok && assigneeID != "" {
        task.AssigneeID = assigneeID
        task.Status = models.TaskStatusAssigned
        task.AssignedAt = time.Now().UnixMilli()
    }
    
    if err := e.taskRepo.Create(ctx, task); err != nil {
        return nil, err
    }
    
    return map[string]any{
        "taskId": task.ID,
        "status": task.Status,
    }, nil
}

func (e *ToolExecutor) updateTaskStatus(ctx context.Context, args map[string]any) (any, error) {
    taskID := args["taskId"].(string)
    status := models.TaskStatus(args["status"].(string))
    
    updates := map[string]any{
        "updated_at": time.Now().UnixMilli(),
    }
    
    if status == models.TaskStatusInProgress {
        updates["started_at"] = time.Now().UnixMilli()
    } else if status == models.TaskStatusCompleted || status == models.TaskStatusFailed {
        updates["completed_at"] = time.Now().UnixMilli()
        if result, ok := args["result"]; ok {
            updates["result_summary"] = result
        }
    }
    
    if err := e.taskRepo.UpdateStatus(ctx, taskID, status, updates); err != nil {
        return nil, err
    }
    
    return map[string]any{"ok": true}, nil
}
```

---

## 六、迁移计划

### Phase 1: 基础设施 (1周)
- [ ] 定义 ACP 协议接口
- [ ] 实现 ACPSession
- [ ] 实现 SessionManager
- [ ] 单元测试

### Phase 2: 协议检测 (1周)
- [ ] 实现 ProtocolDetector
- [ ] 添加协议协商逻辑
- [ ] 集成测试

### Phase 3: 工具系统 (1周)
- [ ] 定义工具 Schema
- [ ] 实现 ToolExecutor
- [ ] 前端工具调用 UI

### Phase 4: 前端适配 (1周)
- [ ] 实现 ACPClient
- [ ] 更新消息组件
- [ ] 添加工具调用展示

### Phase 5: 集成测试 (1周)
- [ ] 端到端测试
- [ ] 性能测试
- [ ] 文档更新

---

## 七、兼容性矩阵

| CLI 工具 | PTY 支持 | ACP 支持 | 推荐协议 |
|----------|----------|----------|----------|
| claude | ✅ | ❌ | PTY |
| gemini | ✅ | ❌ | PTY |
| aider | ✅ | ❌ | PTY |
| cursor-agent | ✅ | ❌ | PTY |
| 未来 ACP CLI | ✅ | ✅ | ACP |

---

## 八、预期收益

1. **更好的工具调用体验** - 结构化的工具定义和调用
2. **更简单的消息解析** - 无需处理 ANSI 转义序列
3. **更强的可扩展性** - 支持多模态内容
4. **更好的错误处理** - 明确的错误响应
5. **更清晰的协议边界** - JSON-RPC 标准格式