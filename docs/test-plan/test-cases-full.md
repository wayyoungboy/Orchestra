# Orchestra 完整协作操作测试用例集

**版本**: v1.0  
**日期**: 2026-04-06  
**用例总数**: 120

---

## 测试用例索引

| 模块 | 用例范围 | 数量 |
|------|----------|------|
| TC-001 ~ TC-010 | 认证授权 | 10 |
| TC-011 ~ TC-030 | 工作区管理 | 20 |
| TC-031 ~ TC-050 | 成员管理 | 20 |
| TC-051 ~ TC-065 | 终端管理 | 15 |
| TC-066 ~ TC-085 | 对话系统 | 20 |
| TC-086 ~ TC-100 | 秘书协调 | 15 |
| TC-101 ~ TC-110 | 附件管理 | 10 |
| TC-111 ~ TC-120 | 端到端协作流程 | 10 |

---

## 一、认证授权模块 (TC-001 ~ TC-010)

### TC-001 获取认证配置（未启用认证）
**前置条件**: 服务启动，认证未启用  
**操作步骤**: GET /api/auth/config  
**预期结果**: {"enabled": false}  
**验证点**: 返回正确的认证状态

### TC-002 获取认证配置（启用认证）
**前置条件**: 服务启动，认证已启用  
**操作步骤**: GET /api/auth/config  
**预期结果**: {"enabled": true, "allowRegistration": true/false}  
**验证点**: 返回正确的认证配置

### TC-003 用户登录成功
**前置条件**: 认证启用，用户存在  
**操作步骤**: POST /api/auth/login {"username":"orchestra","password":"orchestra"}  
**预期结果**: HTTP 200, 返回 token 和 user 信息  
**验证点**: token 有效，user 信息正确

### TC-004 用户登录失败（错误密码）
**前置条件**: 认证启用  
**操作步骤**: POST /api/auth/login {"username":"orchestra","password":"wrong"}  
**预期结果**: HTTP 401, {"error": "invalid credentials"}  
**验证点**: 正确拒绝错误密码

### TC-005 用户登录失败（不存在的用户）
**前置条件**: 认证启用  
**操作步骤**: POST /api/auth/login {"username":"nonexistent","password":"any"}  
**预期结果**: HTTP 401, {"error": "invalid credentials"}  
**验证点**: 不泄露用户存在性

### TC-006 Token验证成功
**前置条件**: 已获取有效Token  
**操作步骤**: POST /api/auth/validate {"token":"valid_token"}  
**预期结果**: HTTP 200, {"valid": true, "userId": "..."}  
**验证点**: Token验证通过

### TC-007 Token验证失败（过期Token）
**前置条件**: Token已过期  
**操作步骤**: POST /api/auth/validate {"token":"expired_token"}  
**预期结果**: HTTP 200, {"valid": false}  
**验证点**: 正确识别过期Token

### TC-008 Token验证失败（无效Token）
**前置条件**: Token格式错误  
**操作步骤**: POST /api/auth/validate {"token":"invalid_token"}  
**预期结果**: HTTP 200, {"valid": false}  
**验证点**: 正确识别无效Token

### TC-009 获取当前用户信息
**前置条件**: 已认证  
**操作步骤**: GET /api/auth/me (Header: Authorization: Bearer token)  
**预期结果**: HTTP 200, 返回用户信息  
**验证点**: 用户信息正确

### TC-010 未认证访问受保护API
**前置条件**: 未登录  
**操作步骤**: GET /api/workspaces (无Authorization头)  
**预期结果**: HTTP 401  
**验证点**: 正确拒绝未认证请求

---

## 二、工作区管理模块 (TC-011 ~ TC-030)

### TC-011 列出工作区（空列表）
**前置条件**: 数据库无工作区  
**操作步骤**: GET /api/workspaces  
**预期结果**: HTTP 200, []  
**验证点**: 返回空数组

### TC-012 列出工作区（有数据）
**前置条件**: 已创建工作区  
**操作步骤**: GET /api/workspaces  
**预期结果**: HTTP 200, 返回工作区数组  
**验证点**: 包含所有工作区，字段完整

### TC-013 创建工作区成功
**前置条件**: 路径存在且允许  
**操作步骤**: POST /api/workspaces {"name":"Test","path":"/valid/path"}  
**预期结果**: HTTP 201, 返回工作区信息  
**验证点**: ID生成，时间戳正确，自动创建Owner

### TC-014 创建工作区失败（路径不存在）
**前置条件**: 路径不存在  
**操作步骤**: POST /api/workspaces {"name":"Test","path":"/nonexistent/path"}  
**预期结果**: HTTP 400, {"error": "path does not exist"}  
**验证点**: 正确验证路径存在性

### TC-015 创建工作区失败（路径不允许）
**前置条件**: 路径在白名单外  
**操作步骤**: POST /api/workspaces {"name":"Test","path":"/etc/passwd"}  
**预期结果**: HTTP 400, {"error": "path not allowed"}  
**验证点**: 白名单机制生效

### TC-016 创建工作区失败（缺少必填字段）
**前置条件**: 无  
**操作步骤**: POST /api/workspaces {"name":"Test"}  
**预期结果**: HTTP 400  
**验证点**: 参数验证生效

### TC-017 创建工作区失败（空名称）
**前置条件**: 无  
**操作步骤**: POST /api/workspaces {"name":"","path":"/valid/path"}  
**预期结果**: HTTP 400  
**验证点**: 名称不能为空

### TC-018 获取工作区成功
**前置条件**: 工作区已创建  
**操作步骤**: GET /api/workspaces/{id}  
**预期结果**: HTTP 200, 返回工作区详情  
**验证点**: 所有字段正确

### TC-019 获取工作区失败（不存在）
**前置条件**: 工作区ID不存在  
**操作步骤**: GET /api/workspaces/nonexistent_id  
**预期结果**: HTTP 404, {"error": "workspace not found"}  
**验证点**: 正确返回404

### TC-020 更新工作区名称
**前置条件**: 工作区已创建  
**操作步骤**: PUT /api/workspaces/{id} {"name":"New Name"}  
**预期结果**: HTTP 200, name已更新  
**验证点**: 只更新名称，其他字段不变

### TC-021 更新工作区路径（有效路径）
**前置条件**: 新路径存在且允许  
**操作步骤**: PUT /api/workspaces/{id} {"path":"/new/valid/path"}  
**预期结果**: HTTP 200, path已更新  
**验证点**: 路径验证生效

### TC-022 更新工作区路径（无效路径）
**前置条件**: 新路径不存在  
**操作步骤**: PUT /api/workspaces/{id} {"path":"/invalid/path"}  
**预期结果**: HTTP 400, {"error": "path does not exist"}  
**验证点**: 路径验证生效

### TC-023 删除工作区成功
**前置条件**: 工作区已创建  
**操作步骤**: DELETE /api/workspaces/{id}  
**预期结果**: HTTP 204  
**验证点**: 工作区已删除

### TC-024 删除工作区级联删除成员
**前置条件**: 工作区有成员  
**操作步骤**: DELETE /api/workspaces/{id}  
**预期结果**: HTTP 204  
**验证点**: 成员被级联删除

### TC-025 删除工作区级联删除对话
**前置条件**: 工作区有对话  
**操作步骤**: DELETE /api/workspaces/{id}  
**预期结果**: HTTP 204  
**验证点**: 对话被级联删除

### TC-026 浏览工作区根目录
**前置条件**: 工作区已创建  
**操作步骤**: GET /api/workspaces/{id}/browse  
**预期结果**: HTTP 200, 返回文件列表  
**验证点**: 文件列表正确

### TC-027 浏览工作区子目录
**前置条件**: 工作区有子目录  
**操作步骤**: GET /api/workspaces/{id}/browse?path=/workspace/subdir  
**预期结果**: HTTP 200, 返回子目录文件列表  
**验证点**: 子目录内容正确

### TC-028 浏览根目录
**前置条件**: 无  
**操作步骤**: GET /api/browse  
**预期结果**: HTTP 200, 返回Home目录内容  
**验证点**: 从Home目录开始

### TC-029 浏览指定路径
**前置条件**: 路径允许访问  
**操作步骤**: GET /api/browse?path=/Users/xxx/projects  
**预期结果**: HTTP 200, 返回指定路径内容  
**验证点**: 路径正确

### TC-030 搜索消息
**前置条件**: 工作区有消息  
**操作步骤**: GET /api/workspaces/{id}/search?q=keyword  
**预期结果**: HTTP 200, 返回匹配消息  
**验证点**: 全文搜索生效

---

## 三、成员管理模块 (TC-031 ~ TC-050)

### TC-031 列出成员（空列表）
**前置条件**: 工作区刚创建，无额外成员  
**操作步骤**: GET /api/workspaces/{id}/members  
**预期结果**: HTTP 200, 返回Owner  
**验证点**: 自动创建Owner

### TC-032 创建秘书成员
**前置条件**: 工作区已创建  
**操作步骤**: POST /api/workspaces/{id}/members {"name":"Alice","roleType":"secretary","program":"claude"}  
**预期结果**: HTTP 201, roleType=secretary  
**验证点**: 秘书角色正确

### TC-033 创建助手成员
**前置条件**: 工作区已创建  
**操作步骤**: POST /api/workspaces/{id}/members {"name":"Bob","roleType":"assistant","program":"claude"}  
**预期结果**: HTTP 201, roleType=assistant  
**验证点**: 助手角色正确

### TC-034 创建管理员成员
**前置条件**: 工作区已创建  
**操作步骤**: POST /api/workspaces/{id}/members {"name":"Admin","roleType":"admin","program":"claude"}  
**预期结果**: HTTP 201, roleType=admin  
**验证点**: 管理员角色正确

### TC-035 创建成员失败（无效角色）
**前置条件**: 工作区已创建  
**操作步骤**: POST /api/workspaces/{id}/members {"name":"Test","roleType":"invalid_role"}  
**预期结果**: HTTP 400  
**验证点**: 角色验证生效

### TC-036 创建成员失败（缺少名称）
**前置条件**: 工作区已创建  
**操作步骤**: POST /api/workspaces/{id}/members {"roleType":"assistant"}  
**预期结果**: HTTP 400  
**验证点**: 名称必填

### TC-037 创建成员失败（工作区不存在）
**前置条件**: 工作区ID无效  
**操作步骤**: POST /api/workspaces/invalid_id/members {...}  
**预期结果**: HTTP 404  
**验证点**: 工作区验证

### TC-038 获取成员详情
**前置条件**: 成员已创建  
**操作步骤**: GET /api/workspaces/{id}/members/{memberId}  
**预期结果**: HTTP 200, 返回成员详情  
**验证点**: 所有字段正确

### TC-039 获取成员失败（不存在）
**前置条件**: 成员ID无效  
**操作步骤**: GET /api/workspaces/{id}/members/invalid_id  
**预期结果**: HTTP 404  
**验证点**: 返回404

### TC-040 更新成员名称
**前置条件**: 成员已创建  
**操作步骤**: PUT /api/workspaces/{id}/members/{memberId} {"name":"New Name"}  
**预期结果**: HTTP 200, 名称已更新  
**验证点**: 更新成功

### TC-041 更新成员系统提示词
**前置条件**: 成员已创建  
**操作步骤**: PUT /api/workspaces/{id}/members/{memberId} {"systemPrompt":"新提示词"}  
**预期结果**: HTTP 200, 提示词已更新  
**验证点**: 更新成功

### TC-042 更新成员API密钥
**前置条件**: 成员已创建  
**操作步骤**: PUT /api/workspaces/{id}/members/{memberId} {"apiKey":"sk-xxx"}  
**预期结果**: HTTP 200, API密钥已更新（加密存储）  
**验证点**: 加密存储

### TC-043 更新成员自动启动终端
**前置条件**: 成员已创建  
**操作步骤**: PUT /api/workspaces/{id}/members/{memberId} {"autoStartTerminal":true}  
**预期结果**: HTTP 200, 设置已更新  
**验证点**: 更新成功

### TC-044 删除成员成功
**前置条件**: 成员已创建  
**操作步骤**: DELETE /api/workspaces/{id}/members/{memberId}  
**预期结果**: HTTP 204  
**验证点**: 成员已删除

### TC-045 删除成员失败（Owner不可删除）
**前置条件**: 尝试删除Owner  
**操作步骤**: DELETE /api/workspaces/{id}/members/{ownerId}  
**预期结果**: HTTP 403 或 400  
**验证点**: Owner保护

### TC-046 更新在线状态
**前置条件**: 成员已创建  
**操作步骤**: POST /api/workspaces/{id}/members/{memberId}/presence {"status":"online"}  
**预期结果**: HTTP 200  
**验证点**: 状态更新

### TC-047 更新活动信息
**前置条件**: 成员已创建  
**操作步骤**: POST /api/workspaces/{id}/members/{memberId}/presence {"status":"busy","activity":"coding"}  
**预期结果**: HTTP 200  
**验证点**: 活动信息记录

### TC-048 成员数量统计
**前置条件**: 工作区有多个成员  
**操作步骤**: GET /api/workspaces/{id}/members  
**预期结果**: 返回所有成员  
**验证点**: 数量正确

### TC-049 成员角色分布
**前置条件**: 有不同角色成员  
**操作步骤**: GET /api/workspaces/{id}/members  
**预期结果**: 包含owner、secretary、assistant等  
**验证点**: 角色分布正确

### TC-050 创建重复名称成员
**前置条件**: 成员名称已存在  
**操作步骤**: POST /api/workspaces/{id}/members {"name":"SameName",...}  
**预期结果**: HTTP 201（允许同名）或 400（不允许）  
**验证点**: 检查是否允许重名

---

## 四、终端管理模块 (TC-051 ~ TC-065)

### TC-051 创建终端会话（允许的命令）
**前置条件**: 命令在白名单  
**操作步骤**: POST /api/terminals {"workspaceId":"xxx","command":"claude"}  
**预期结果**: HTTP 200, {"sessionId":"..."}  
**验证点**: 会话创建成功

### TC-052 创建终端会话失败（不允许的命令）
**前置条件**: 命令不在白名单  
**操作步骤**: POST /api/terminals {"workspaceId":"xxx","command":"rm"}  
**预期结果**: HTTP 400, {"error": "command not allowed"}  
**验证点**: 白名单生效

### TC-053 创建终端会话失败（缺少工作区ID）
**前置条件**: 无  
**操作步骤**: POST /api/terminals {"command":"claude"}  
**预期结果**: HTTP 400  
**验证点**: 参数验证

### TC-054 删除终端会话
**前置条件**: 会话已创建  
**操作步骤**: DELETE /api/terminals/{sessionId}  
**预期结果**: HTTP 204  
**验证点**: 会话已删除

### TC-055 删除不存在的会话
**前置条件**: 会话ID无效  
**操作步骤**: DELETE /api/terminals/invalid_session  
**预期结果**: HTTP 404 或 204  
**验证点**: 处理正确

### TC-056 获取成员终端会话
**前置条件**: 成员有绑定的会话  
**操作步骤**: GET /api/workspaces/{id}/members/{memberId}/terminal-session  
**预期结果**: HTTP 200, 返回会话信息  
**验证点**: 会话信息正确

### TC-057 创建成员终端会话
**前置条件**: 成员无绑定会话  
**操作步骤**: POST /api/workspaces/{id}/members/{memberId}/terminal-session  
**预期结果**: HTTP 200, 返回会话信息  
**验证点**: 会话创建并绑定

### TC-058 列出工作区终端会话
**前置条件**: 工作区有多个会话  
**操作步骤**: GET /api/workspaces/{id}/terminal-sessions  
**预期结果**: HTTP 200, 返回会话列表  
**验证点**: 列表正确

### TC-059 WebSocket终端连接
**前置条件**: 会话已创建  
**操作步骤**: WebSocket连接 /ws/terminal/{sessionId}  
**预期结果**: 连接成功  
**验证点**: WebSocket建立

### TC-060 WebSocket终端数据传输
**前置条件**: WebSocket已连接  
**操作步骤**: 发送终端输入数据  
**预期结果**: 收到终端输出  
**验证点**: 双向通信

### TC-061 WebSocket终端断开
**前置条件**: WebSocket已连接  
**操作步骤**: 关闭连接  
**预期结果**: 资源正确释放  
**验证点**: 无资源泄漏

### TC-062 终端会话超时
**前置条件**: 会话空闲超过超时时间  
**操作步骤**: 等待超时  
**预期结果**: 会话自动关闭  
**验证点**: 超时机制生效

### TC-063 最大会话数限制
**前置条件**: 达到最大会话数  
**操作步骤**: 尝试创建新会话  
**预期结果**: HTTP 429 或 503  
**验证点**: 限制生效

### TC-064 终端进程异常退出
**前置条件**: 会话运行中  
**操作步骤**: 进程异常退出  
**预期结果**: WebSocket收到退出通知  
**验证点**: 异常处理

### TC-065 终端窗口大小调整
**前置条件**: WebSocket已连接  
**操作步骤**: 发送resize消息  
**预期结果**: 终端大小调整  
**验证点**: resize生效

---

## 五、对话系统模块 (TC-066 ~ TC-085)

### TC-066 列出对话
**前置条件**: 工作区已创建  
**操作步骤**: GET /api/workspaces/{id}/conversations  
**预期结果**: HTTP 200, 返回对话列表  
**验证点**: 包含默认频道

### TC-067 创建频道对话
**前置条件**: 工作区已创建  
**操作步骤**: POST /api/workspaces/{id}/conversations {"title":"general","type":"channel"}  
**预期结果**: HTTP 201, type=channel  
**验证点**: 频道创建成功

### TC-068 创建私信对话
**前置条件**: 工作区有多个成员  
**操作步骤**: POST /api/workspaces/{id}/conversations {"type":"dm","memberIds":["id1","id2"]}  
**预期结果**: HTTP 201, type=dm  
**验证点**: 私信创建成功

### TC-069 获取对话详情
**前置条件**: 对话已创建  
**操作步骤**: GET /api/workspaces/{id}/conversations/{convId}  
**预期结果**: HTTP 200, 返回对话详情  
**验证点**: 信息正确

### TC-070 更新对话设置
**前置条件**: 对话已创建  
**操作步骤**: PUT /api/workspaces/{id}/conversations/{convId} {"title":"New Title"}  
**预期结果**: HTTP 200  
**验证点**: 更新成功

### TC-071 删除对话
**前置条件**: 对话已创建  
**操作步骤**: DELETE /api/workspaces/{id}/conversations/{convId}  
**预期结果**: HTTP 204  
**验证点**: 对话已删除

### TC-072 获取消息列表（空）
**前置条件**: 对话刚创建  
**操作步骤**: GET /api/workspaces/{id}/conversations/{convId}/messages  
**预期结果**: HTTP 200, []  
**验证点**: 返回空数组

### TC-073 发送消息
**前置条件**: 对话已创建，成员有终端  
**操作步骤**: POST /api/workspaces/{id}/conversations/{convId}/messages {"senderId":"xxx","text":"Hello"}  
**预期结果**: HTTP 200  
**验证点**: 消息发送

### TC-074 发送消息失败（对话不存在）
**前置条件**: 对话ID无效  
**操作步骤**: POST /api/workspaces/{id}/conversations/invalid/messages {...}  
**预期结果**: HTTP 404  
**验证点**: 验证对话存在

### TC-075 删除单条消息
**前置条件**: 消息已发送  
**操作步骤**: DELETE /api/workspaces/{id}/conversations/{convId}/messages/{messageId}  
**预期结果**: HTTP 204  
**验证点**: 消息已删除

### TC-076 清空对话消息
**前置条件**: 对话有消息  
**操作步骤**: DELETE /api/workspaces/{id}/conversations/{convId}/messages  
**预期结果**: HTTP 204  
**验证点**: 消息已清空

### TC-077 标记对话已读
**前置条件**: 对话有消息  
**操作步骤**: POST /api/workspaces/{id}/conversations/{convId}/read {"userId":"xxx"}  
**预期结果**: HTTP 200  
**验证点**: 已读状态更新

### TC-078 标记全部已读
**前置条件**: 工作区有多个对话  
**操作步骤**: POST /api/workspaces/{id}/conversations/read-all {"userId":"xxx"}  
**预期结果**: HTTP 200  
**验证点**: 所有对话已读

### TC-079 设置对话成员
**前置条件**: 频道已创建  
**操作步骤**: PUT /api/workspaces/{id}/conversations/{convId}/members {"memberIds":["id1","id2"]}  
**预期结果**: HTTP 200  
**验证点**: 成员设置成功

### TC-080 置顶对话
**前置条件**: 对话已创建  
**操作步骤**: PUT /api/workspaces/{id}/conversations/{convId} {"pinned":true}  
**预期结果**: HTTP 200  
**验证点**: 置顶成功

### TC-081 静音对话
**前置条件**: 对话已创建  
**操作步骤**: PUT /api/workspaces/{id}/conversations/{convId} {"muted":true}  
**预期结果**: HTTP 200  
**验证点**: 静音成功

### TC-082 WebSocket聊天连接
**前置条件**: 无  
**操作步骤**: WebSocket连接 /ws/chat/{workspaceId}  
**预期结果**: 连接成功  
**验证点**: WebSocket建立

### TC-083 WebSocket接收消息
**前置条件**: WebSocket已连接  
**操作步骤**: 其他用户发送消息  
**预期结果**: 收到消息推送  
**验证点**: 实时推送

### TC-084 WebSocket接收成员状态
**前置条件**: WebSocket已连接  
**操作步骤**: 成员状态变化  
**预期结果**: 收到状态更新  
**验证点**: 状态同步

### TC-085 内部API发送消息
**前置条件**: 对话已创建  
**操作步骤**: POST /api/internal/chat/send {...}  
**预期结果**: HTTP 200  
**验证点**: 消息发送成功

---

## 六、秘书协调模块 (TC-086 ~ TC-100)

### TC-086 创建任务（分配给助手）
**前置条件**: 秘书和助手已创建  
**操作步骤**: POST /api/internal/tasks/create {"secretaryId":"xxx","assigneeId":"yyy",...}  
**预期结果**: HTTP 201, status=assigned  
**验证点**: 任务创建，状态为assigned

### TC-087 创建任务（未分配）
**前置条件**: 秘书已创建  
**操作步骤**: POST /api/internal/tasks/create {"secretaryId":"xxx",...}（无assigneeId）  
**预期结果**: HTTP 201, status=pending  
**验证点**: 任务创建，状态为pending

### TC-088 创建任务失败（缺少必填字段）
**前置条件**: 无  
**操作步骤**: POST /api/internal/tasks/create {"title":"Test"}  
**预期结果**: HTTP 400  
**验证点**: 参数验证

### TC-089 开始任务
**前置条件**: 任务已创建  
**操作步骤**: POST /api/internal/tasks/start {"taskId":"xxx"}  
**预期结果**: HTTP 200, status=in_progress  
**验证点**: 状态更新

### TC-090 开始任务失败（任务不存在）
**前置条件**: 任务ID无效  
**操作步骤**: POST /api/internal/tasks/start {"taskId":"invalid"}  
**预期结果**: HTTP 404  
**验证点**: 任务验证

### TC-091 完成任务
**前置条件**: 任务已开始  
**操作步骤**: POST /api/internal/tasks/complete {"taskId":"xxx","resultSummary":"Done"}  
**预期结果**: HTTP 200, status=completed  
**验证点**: 状态更新，结果记录

### TC-092 完成任务失败（任务未开始）
**前置条件**: 任务状态为pending  
**操作步骤**: POST /api/internal/tasks/complete {"taskId":"xxx"}  
**预期结果**: HTTP 200 或 400  
**验证点**: 状态流转验证

### TC-093 任务失败报告
**前置条件**: 任务已开始  
**操作步骤**: POST /api/internal/tasks/fail {"taskId":"xxx","errorMessage":"Error"}  
**预期结果**: HTTP 200, status=failed  
**验证点**: 状态更新，错误记录

### TC-094 列出工作区任务
**前置条件**: 工作区有任务  
**操作步骤**: GET /api/workspaces/{id}/tasks  
**预期结果**: HTTP 200, 返回任务列表  
**验证点**: 列表正确

### TC-095 列出任务（按状态过滤）
**前置条件**: 工作区有不同状态任务  
**操作步骤**: GET /api/workspaces/{id}/tasks?status=completed  
**预期结果**: HTTP 200, 只返回已完成任务  
**验证点**: 过滤生效

### TC-096 获取任务详情
**前置条件**: 任务已创建  
**操作步骤**: GET /api/workspaces/{id}/tasks/{taskId}  
**预期结果**: HTTP 200, 返回任务详情  
**验证点**: 信息完整

### TC-097 查询成员任务
**前置条件**: 成员有分配的任务  
**操作步骤**: GET /api/workspaces/{id}/tasks/my-tasks?memberId=xxx  
**预期结果**: HTTP 200, 返回该成员任务  
**验证点**: 过滤正确

### TC-098 查询负载统计
**前置条件**: 工作区有助手成员  
**操作步骤**: GET /api/internal/workloads/list?workspaceId=xxx  
**预期结果**: HTTP 200, 返回负载统计  
**验证点**: 统计正确

### TC-099 负载统计准确性
**前置条件**: 助手有多个任务  
**操作步骤**: 查询负载  
**预期结果**: currentTaskCount、completedTaskCount正确  
**验证点**: 数量准确

### TC-100 任务完整生命周期
**前置条件**: 秘书和助手已创建  
**操作步骤**: 创建→开始→完成  
**预期结果**: 状态流转: pending→assigned→in_progress→completed  
**验证点**: 完整流程

---

## 七、附件管理模块 (TC-101 ~ TC-110)

### TC-101 上传附件
**前置条件**: 对话已创建  
**操作步骤**: POST /api/workspaces/{id}/conversations/{convId}/attachments (multipart/form-data)  
**预期结果**: HTTP 201, 返回附件信息  
**验证点**: 文件保存成功

### TC-102 上传附件失败（缺少文件）
**前置条件**: 对话已创建  
**操作步骤**: POST 不带file字段  
**预期结果**: HTTP 400  
**验证点**: 参数验证

### TC-103 上传附件失败（文件过大）
**前置条件**: 文件超过限制  
**操作步骤**: 上传大文件  
**预期结果**: HTTP 400  
**验证点**: 大小限制

### TC-104 列出附件
**前置条件**: 工作区有附件  
**操作步骤**: GET /api/workspaces/{id}/attachments  
**预期结果**: HTTP 200, 返回附件列表  
**验证点**: 列表正确

### TC-105 下载附件
**前置条件**: 附件已上传  
**操作步骤**: GET /api/workspaces/{id}/attachments/{attachmentId}  
**预期结果**: HTTP 200, 文件内容  
**验证点**: 内容正确

### TC-106 下载附件失败（不存在）
**前置条件**: 附件ID无效  
**操作步骤**: GET /api/workspaces/{id}/attachments/invalid  
**预期结果**: HTTP 404  
**验证点**: 返回404

### TC-107 获取附件信息
**前置条件**: 附件已上传  
**操作步骤**: GET /api/workspaces/{id}/attachments/{attachmentId}/info  
**预期结果**: HTTP 200, 返回元信息  
**验证点**: 信息正确

### TC-108 删除附件
**前置条件**: 附件已上传  
**操作步骤**: DELETE /api/workspaces/{id}/attachments/{attachmentId}  
**预期结果**: HTTP 204  
**验证点**: 文件已删除

### TC-109 删除附件失败（不存在）
**前置条件**: 附件ID无效  
**操作步骤**: DELETE /api/workspaces/{id}/attachments/invalid  
**预期结果**: HTTP 404  
**验证点**: 返回404

### TC-110 图片附件预览
**前置条件**: 上传图片  
**操作步骤**: GET附件信息  
**预期结果**: isImage=true  
**验证点**: 图片识别

---

## 八、端到端协作流程 (TC-111 ~ TC-120)

### TC-111 完整工作区创建流程
**流程**: 创建工作区→自动创建Owner→验证成员列表  
**预期**: 工作区和Owner成员都创建成功  
**验证点**: 自动化创建正确

### TC-112 完整成员创建流程
**流程**: 创建秘书→创建助手→创建对话→验证成员关联  
**预期**: 所有成员创建成功，对话可关联  
**验证点**: 关联关系正确

### TC-113 任务分配完整流程
**流程**: 秘书创建任务→分配给助手→助手开始执行  
**预期**: 任务状态正确流转  
**验证点**: 状态机正确

### TC-114 任务完成反馈流程
**流程**: 助手完成任务→返回结果→秘书收到通知  
**预期**: 任务状态更新，结果记录  
**验证点**: 反馈机制生效

### TC-115 任务失败处理流程
**流程**: 助手报告失败→记录错误→秘书收到通知  
**预期**: 错误记录，状态更新  
**验证点**: 错误处理正确

### TC-116 多助手协作流程
**流程**: 秘书创建多个任务→分配给不同助手→查询负载  
**预期**: 负载统计正确  
**验证点**: 多任务协调

### TC-117 消息发送完整流程
**流程**: 用户发送消息→消息存储→WebSocket推送  
**预期**: 消息持久化，实时推送  
**验证点**: 存储和推送都成功

### TC-118 工作区删除级联流程
**流程**: 删除工作区→验证成员、对话、消息都删除  
**预期**: 所有关联数据清理  
**验证点**: 级联删除生效

### TC-119 成员删除影响流程
**流程**: 删除成员→验证任务处理  
**预期**: 任务assignee_id置空或任务取消  
**验证点**: 外键约束正确

### TC-120 完整协作场景模拟
**流程**: 
1. 创建工作区
2. 创建秘书和助手
3. 创建对话
4. 秘书创建任务分配给助手
5. 助手开始任务
6. 助手完成任务
7. 查询负载统计
8. 验证所有数据

**预期**: 整个流程顺利完成  
**验证点**: 端到端完整性

---

## 测试执行要求

1. **执行顺序**: 按模块顺序执行，每个模块内按用例编号执行
2. **数据隔离**: 每个测试用例使用独立的测试数据
3. **结果记录**: 记录每个用例的实际结果和状态
4. **问题追踪**: 发现问题记录到缺陷列表
5. **回归测试**: 修复问题后重新执行相关用例