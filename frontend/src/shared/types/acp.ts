// ACP (Agent Communication Protocol) TypeScript types
// For structured JSON communication with AI agents

export type ACPMessageType =
  | 'user_message'
  | 'assistant_message'
  | 'tool_use'
  | 'tool_result'
  | 'result'
  | 'error'
  | 'system'

// Input message types (sent to agent)
export interface ACPUserMessage {
  type: 'user_message'
  content: string
}

export interface ACPToolResult {
  type: 'tool_result'
  tool_use_id: string
  content: string
  is_error?: boolean
}

// Output message types (received from agent)
export interface ACPAssistantMessage {
  type: 'assistant_message'
  content: string
}

export interface ACPToolUse {
  type: 'tool_use'
  name: string
  tool_use_id: string
  input: Record<string, unknown>
}

export interface ACPResult {
  type: 'result'
  message?: string
  cost_usd?: number
  duration_ms?: number
}

export interface ACPError {
  type: 'error'
  error: string
}

export interface ACPSystem {
  type: 'system'
  message: string
  level?: 'info' | 'warning' | 'error'
}

export type ACPMessage =
  | ACPUserMessage
  | ACPAssistantMessage
  | ACPToolUse
  | ACPToolResult
  | ACPResult
  | ACPError
  | ACPSystem

// WebSocket message types for ACP sessions
export interface ACPTerminalConnected {
  type: 'connected'
  sessionId: string
}

export interface ACPTerminalAssistantMessage {
  type: 'assistant_message'
  content: string
}

export interface ACPTerminalToolUse {
  type: 'tool_use'
  tool_name: string
  tool_use_id: string
  tool_input: Record<string, unknown>
}

export interface ACPTerminalResult {
  type: 'result'
  message?: string
  cost_usd?: number
  duration_ms?: number
}

export interface ACPTerminalError {
  type: 'error'
  error: string
}

export interface ACPTerminalStatus {
  type: 'status'
  status: string
  message?: string
}

export interface ACPTerminalExit {
  type: 'exit'
  code: number
}

export type ACPTerminalServerMessage =
  | ACPTerminalConnected
  | ACPTerminalAssistantMessage
  | ACPTerminalToolUse
  | ACPTerminalResult
  | ACPTerminalError
  | ACPTerminalStatus
  | ACPTerminalExit

// Client message types
export interface ACPTerminalUserMessage {
  type: 'user_message' | 'input'
  content: string
}

export interface ACPTerminalToolResultMessage {
  type: 'tool_result'
  tool_use_id: string
  tool_result: string
  is_error?: boolean
}

export interface ACPTerminalCloseMessage {
  type: 'close'
}

export type ACPTerminalClientMessage =
  | ACPTerminalUserMessage
  | ACPTerminalToolResultMessage
  | ACPTerminalCloseMessage

// Orchestra tool types
export interface OrchestraToolDefinition {
  name: string
  description: string
  input_schema: Record<string, unknown>
}

export const ORCHESTRA_TOOLS = {
  CHAT_SEND: 'orchestra_chat_send',
  TASK_CREATE: 'orchestra_task_create',
  TASK_START: 'orchestra_task_start',
  TASK_COMPLETE: 'orchestra_task_complete',
  TASK_FAIL: 'orchestra_task_fail',
  WORKLOAD_LIST: 'orchestra_workload_list',
  AGENT_STATUS: 'orchestra_agent_status',
  FILE_READ: 'orchestra_file_read',
  FILE_WRITE: 'orchestra_file_write',
  FILE_LIST: 'orchestra_file_list',
} as const

// Tool input types
export interface ChatSendInput {
  conversationId: string
  text: string
}

export interface TaskCreateInput {
  conversationId: string
  title: string
  description?: string
  assigneeId?: string
  priority?: number
}

export interface TaskStartInput {
  taskId: string
  message?: string
}

export interface TaskCompleteInput {
  taskId: string
  resultSummary?: string
}

export interface TaskFailInput {
  taskId: string
  errorMessage: string
}

export interface WorkloadListInput {
  workspaceId: string
}

export interface AgentStatusInput {
  status: 'idle' | 'thinking' | 'reading_file' | 'writing_code' | 'executing_command' | 'waiting' | 'error'
  message?: string
  progress?: number
}