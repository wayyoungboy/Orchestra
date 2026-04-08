# Orchestra 系统架构与流程图

## 一、系统整体架构

```mermaid
graph TB
    subgraph "客户端层"
        Browser[浏览器]
        WS_Client[WebSocket 客户端]
    end

    subgraph "前端 Vue 3"
        Router[Vue Router]
        Pinia[Pinia Store]
        Components[UI 组件]
        API_Client[API 客户端]
        Socket_Client[Socket 连接]
    end

    subgraph "后端 Go + Gin"
        Middleware[中间件层<br/>CORS/Auth/Logger]
        Handlers[Handler 层<br/>API 处理器]
        Services[服务层<br/>终端池/网关]
        Repository[Repository 层<br/>数据访问]
    end

    subgraph "外部服务"
        CLI[CLI 工具<br/>Claude/Gemini]
    end

    subgraph "数据存储"
        SQLite[(SQLite 数据库)]
        FileSystem[文件系统]
    end

    Browser --> Router
    Router --> Components
    Components --> Pinia
    Pinia --> API_Client
    Pinia --> Socket_Client
    
    API_Client -->|HTTP/REST| Middleware
    Socket_Client -->|WebSocket| Services
    
    Middleware --> Handlers
    Handlers --> Services
    Handlers --> Repository
    Services --> CLI
    Services --> Repository
    
    Repository --> SQLite
    Services --> FileSystem
```

## 二、后端模块架构

```mermaid
graph LR
    subgraph "API 层"
        Router[router.go<br/>路由配置]
        AuthH[auth.go]
        WorkspaceH[workspace.go]
        MemberH[member.go]
        TerminalH[terminal.go]
        ConvH[conversation.go]
        TaskH[task.go]
        AttachH[attachment.go]
    end

    subgraph "中间件"
        CORS[cors.go]
        AuthM[auth.go]
        Logger[logger.go]
    end

    subgraph "服务层"
        Pool[ProcessPool<br/>终端进程池]
        Gateway[Gateway<br/>WebSocket网关]
        ChatBridge[ChatBridge<br/>消息桥接]
        Browser[Browser<br/>文件浏览]
    end

    subgraph "数据层"
        WS_Repo[WorkspaceRepository]
        MemberRepo[MemberRepository]
        ConvRepo[ConversationRepository]
        MsgRepo[MessageRepository]
        TaskRepo[TaskRepository]
        UserRepo[UserRepository]
        AttachRepo[AttachmentRepository]
    end

    subgraph "模型层"
        Workspace[Workspace]
        Member[Member]
        Conversation[Conversation]
        Message[Message]
        Task[Task]
        User[User]
        Attachment[Attachment]
    end

    Router --> CORS
    Router --> AuthM
    Router --> Logger
    Router --> AuthH
    Router --> WorkspaceH
    Router --> MemberH
    Router --> TerminalH
    Router --> ConvH
    Router --> TaskH
    Router --> AttachH

    AuthH --> UserRepo
    WorkspaceH --> WS_Repo
    WorkspaceH --> Browser
    MemberH --> MemberRepo
    TerminalH --> Pool
    ConvH --> ConvRepo
    ConvH --> MsgRepo
    ConvH --> Pool
    TaskH --> TaskRepo
    TaskH --> MemberRepo
    AttachH --> AttachRepo

    Pool --> ChatBridge
    Gateway --> Pool

    WS_Repo --> Workspace
    MemberRepo --> Member
    ConvRepo --> Conversation
    MsgRepo --> Message
    TaskRepo --> Task
```

## 三、前端模块架构

```mermaid
graph TB
    subgraph "应用入口"
        App[App.vue]
        Router[router.ts]
    end

    subgraph "功能模块"
        Auth[认证模块<br/>LoginPage.vue<br/>authStore.ts]
        Workspace[工作区模块<br/>WorkspaceSelection.vue<br/>WorkspaceMain.vue<br/>workspaceStore.ts]
        Chat[对话模块<br/>ChatInterface.vue<br/>ChatSidebar.vue<br/>chatStore.ts]
        Terminal[终端模块<br/>TerminalWorkspace.vue<br/>terminalStore.ts]
        Members[成员模块<br/>MembersPage.vue<br/>MemberRow.vue]
        Settings[设置模块<br/>Settings.vue<br/>settingsStore.ts]
    end

    subgraph "共享组件"
        SidebarNav[SidebarNav.vue<br/>侧边导航]
        ToastStack[ToastStack.vue<br/>消息提示]
        WorkspaceSwitcher[WorkspaceSwitcher.vue<br/>工作区切换]
    end

    subgraph "API & 通信"
        APIClient[api/client.ts<br/>HTTP 客户端]
        ChatSocket[socket/chat.ts<br/>聊天 WebSocket]
        TerminalSocket[socket/terminal.ts<br/>终端 WebSocket]
    end

    subgraph "状态管理"
        Pinia[Pinia Stores]
    end

    App --> Router
    Router --> Auth
    Router --> Workspace
    Router --> Chat
    Router --> Terminal
    Router --> Members
    Router --> Settings

    Auth --> Pinia
    Workspace --> Pinia
    Chat --> Pinia
    Terminal --> Pinia
    Members --> Pinia
    Settings --> Pinia

    Pinia --> APIClient
    Pinia --> ChatSocket
    Pinia --> TerminalSocket
```

## 四、数据库模型关系

```mermaid
erDiagram
    WORKSPACES ||--o{ MEMBERS : contains
    WORKSPACES ||--o{ CONVERSATIONS : contains
    WORKSPACES ||--o{ TASKS : contains
    WORKSPACES ||--o{ ATTACHMENTS : contains

    MEMBERS ||--o{ TASKS : assigned_to
    MEMBERS ||--o{ AGENT_STATUS : has

    CONVERSATIONS ||--o{ MESSAGES : contains
    CONVERSATIONS ||--o{ ATTACHMENTS : has
    CONVERSATIONS ||--o{ CONVERSATION_READS : tracks

    USERS ||--o{ WORKSPACES : owns

    WORKSPACES {
        string id PK
        string name
        string path
        datetime lastOpenedAt
        datetime createdAt
    }

    MEMBERS {
        string id PK
        string workspaceId FK
        string name
        string roleType
        string program
        string systemPrompt
        string apiKey
        boolean autoStartTerminal
        string status
    }

    CONVERSATIONS {
        string id PK
        string workspaceId FK
        string type
        string name
        string[] memberIds
        boolean pinned
        boolean muted
    }

    MESSAGES {
        string id PK
        string conversationId FK
        string senderId
        string content
        boolean isAI
        string status
    }

    TASKS {
        string id PK
        string workspaceId FK
        string conversationId FK
        string secretaryId FK
        string assigneeId FK
        string title
        string status
        string resultSummary
        string errorMessage
    }

    ATTACHMENTS {
        string id PK
        string workspaceId FK
        string messageId FK
        string fileName
        string filePath
        integer fileSize
        string mimeType
    }

    USERS {
        string id PK
        string username
        string passwordHash
        datetime createdAt
    }
```

## 五、核心流程图

### 5.1 用户登录流程

```mermaid
sequenceDiagram
    participant U as 用户
    participant F as 前端
    participant B as 后端
    participant D as 数据库

    U->>F: 打开应用
    F->>B: GET /api/auth/config
    B->>F: {enabled: true/false}
    
    alt 认证启用
        F->>U: 显示登录页面
        U->>F: 输入用户名密码
        F->>B: POST /api/auth/login
        B->>D: 查询用户
        D->>B: 用户数据
        B->>B: 验证密码
        B->>B: 生成JWT
        B->>F: {token, user}
        F->>F: 存储token
        F->>F: 跳转到工作区
    else 认证未启用
        F->>F: 直接跳转到工作区
    end
```

### 5.2 工作区管理流程

```mermaid
sequenceDiagram
    participant U as 用户
    participant F as 前端
    participant B as 后端
    participant D as 数据库
    participant FS as 文件系统

    U->>F: 创建工作区
    F->>B: POST /api/workspaces
    B->>FS: 验证路径存在
    FS->>B: 路径验证结果
    
    alt 路径有效
        B->>B: 生成工作区ID
        B->>D: 创建工作区记录
        B->>B: 创建Owner成员
        B->>D: 创建成员记录
        B->>F: 工作区创建成功
        F->>U: 显示工作区
    else 路径无效
        B->>F: 错误：路径不存在/不允许
        F->>U: 显示错误提示
    end
```

### 5.3 消息发送流程

```mermaid
sequenceDiagram
    participant U as 用户
    participant F as 前端
    participant B as 后端
    participant P as 终端进程
    participant D as 数据库
    participant WS as WebSocket客户端

    U->>F: 输入消息
    F->>B: POST /conversations/:id/messages
    B->>D: 存储消息
    B->>P: 发送到终端
    P->>P: AI处理消息
    
    loop 输出流
        P->>B: 终端输出
        B->>B: 解析输出
        B->>D: 存储AI回复
        B->>WS: 广播消息
        WS->>F: 收到消息
        F->>U: 显示消息
    end
```

### 5.4 秘书协调流程

```mermaid
sequenceDiagram
    participant U as 用户
    participant F as 前端
    participant B as 后端
    participant S as 秘书终端
    participant A as 助手终端
    participant D as 数据库

    U->>F: 发送任务请求
    F->>B: POST /conversations/:id/messages
    B->>S: 消息到达秘书终端
    
    S->>S: 分析任务需求
    S->>B: GET /internal/workloads/list
    B->>D: 查询负载统计
    D->>B: 返回负载数据
    B->>S: 负载信息
    
    S->>S: 选择合适助手
    S->>B: POST /internal/tasks/create
    B->>D: 创建任务记录
    B->>A: 通知助手(通过终端)
    
    A->>B: POST /internal/tasks/start
    B->>D: 更新任务状态
    
    A->>A: 执行任务
    
    alt 任务成功
        A->>B: POST /internal/tasks/complete
        B->>D: 更新状态+结果
        B->>S: 通知秘书
    else 任务失败
        A->>B: POST /internal/tasks/fail
        B->>D: 更新状态+错误
        B->>S: 通知秘书
    end
    
    S->>U: 汇报结果
```

### 5.5 终端会话流程

```mermaid
sequenceDiagram
    participant F as 前端
    participant B as 后端
    participant P as ProcessPool
    participant T as PTY进程
    participant CLI as CLI工具

    F->>B: POST /api/terminals
    B->>P: Acquire(workspaceId, command)
    P->>P: 检查命令白名单
    
    alt 命令允许
        P->>T: 创建PTY会话
        T->>CLI: 启动CLI进程
        P->>B: Session{ID, PID}
        B->>F: {sessionId}
        
        F->>B: WebSocket /ws/terminal/:sessionId
        B->>P: Get(sessionId)
        P->>B: Session
        
        loop 双向通信
            F->>B: 发送输入
            B->>T: 写入PTY
            T->>CLI: 输入
            CLI->>T: 输出
            T->>B: 读取输出
            B->>F: 发送输出
        end
    else 命令不允许
        P->>B: 错误：命令不允许
        B->>F: 错误响应
    end
```

## 六、WebSocket 事件流程

```mermaid
sequenceDiagram
    participant F as 前端
    participant G as Gateway
    participant H as Handler
    participant P as ProcessPool
    participant D as 数据库

    Note over F,G: 终端 WebSocket
    F->>G: Connect /ws/terminal/:sessionId
    G->>P: Get Session
    P->>G: Session
    G->>F: Connection Established
    
    loop 终端数据
        F->>G: Input Data
        G->>H: HandleTerminal
        H->>P: Write to PTY
        P->>G: Output Data
        G->>F: Output Data
    end

    Note over F,G: 聊天 WebSocket
    F->>G: Connect /ws/chat/:workspaceId
    G->>F: Connection Established
    
    loop 聊天事件
        D->>H: 新消息
        H->>G: 广播消息
        G->>F: 消息事件
        
        D->>H: 成员状态变化
        H->>G: 广播状态
        G->>F: 状态事件
        
        D->>H: 任务状态变化
        H->>G: 广播任务
        G->>F: 任务事件
    end
```

## 七、安全架构

```mermaid
graph TB
    subgraph "请求入口"
        Request[HTTP请求]
    end

    subgraph "安全中间件"
        CORS[CORS验证<br/>来源检查]
        Auth[认证中间件<br/>JWT验证]
        Logger[日志记录]
    end

    subgraph "安全检查"
        PathValidator[路径验证器<br/>白名单检查]
        CommandValidator[命令验证器<br/>白名单检查]
        RateLimit[速率限制]
    end

    subgraph "数据处理"
        Sanitizer[输入清理]
        Validator[参数验证]
        Encryptor[加密存储]
    end

    Request --> CORS
    CORS --> Auth
    Auth --> Logger
    Logger --> PathValidator
    Logger --> CommandValidator
    
    PathValidator --> Sanitizer
    CommandValidator --> Sanitizer
    Sanitizer --> Validator
    Validator --> Encryptor
```

## 八、部署架构

```mermaid
graph TB
    subgraph "用户设备"
        Browser[浏览器]
    end

    subgraph "服务器"
        subgraph "Go 后端服务"
            HTTP[HTTP Server<br/>:8080]
            WS[WebSocket Server<br/>:8080]
        end
        
        subgraph "进程管理"
            PTY1[PTY Session 1]
            PTY2[PTY Session 2]
            PTYn[PTY Session N]
        end
    end

    subgraph "存储"
        SQLite[(SQLite<br/>orchestra.db)]
        Uploads[上传目录<br/>./uploads]
        Config[配置文件<br/>config.yaml]
    end

    subgraph "CLI 工具"
        Claude[Claude Code]
        Gemini[Gemini CLI]
        Other[其他 CLI]
    end

    Browser -->|HTTP/WS| HTTP
    Browser -->|WS| WS
    
    HTTP --> PTY1
    HTTP --> PTY2
    HTTP --> PTYn
    
    WS --> PTY1
    WS --> PTY2
    WS --> PTYn
    
    PTY1 --> Claude
    PTY2 --> Gemini
    PTYn --> Other
    
    HTTP --> SQLite
    HTTP --> Uploads
    HTTP --> Config
```

---

## 说明

1. **架构分层**：采用经典的三层架构（表现层、业务层、数据层）
2. **前后端分离**：Vue 3 前端 + Go 后端，通过 REST API 和 WebSocket 通信
3. **实时通信**：WebSocket 用于终端 I/O 和聊天消息推送
4. **进程管理**：PTY 进程池管理多个 CLI 会话
5. **安全机制**：路径/命令白名单、JWT 认证、CORS 验证