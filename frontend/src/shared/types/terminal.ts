export type TerminalType = 'native' | 'web' | 'claude' | 'custom'

/** Aligns with common desktop patterns: pending → connecting → connected / working / error / disconnected */
export type TerminalConnectionStatus =
  | 'pending'
  | 'connecting'
  | 'connected'
  | 'working'
  | 'disconnected'
  | 'error'

export interface TerminalTab {
  id: string
  title: string
  sessionId?: string
  hasActivity: boolean
  isBlinking: boolean
  pinned: boolean
  memberId?: string
  terminalType?: TerminalType
  keepAlive?: boolean
}

export type TerminalLayoutMode = 'single' | 'split-vertical' | 'split-horizontal' | 'grid-2x2'

export type TerminalPaneId =
  | 'primary'
  | 'left'
  | 'right'
  | 'top'
  | 'bottom'
  | 'top-left'
  | 'top-right'
  | 'bottom-left'
  | 'bottom-right'

export interface TerminalSession {
  id: string
  pid: number
  cwd: string
  cols: number
  rows: number
  createdAt: number
}

/** Reference-desktop-compatible terminal → chat stream payload (WebSocket). */
export interface TerminalChatStreamPayload {
  terminalId: string
  memberId?: string
  workspaceId?: string
  conversationId?: string
  seq: number
  timestamp: number
  content: string
  type: string
  source: string
  mode: 'delta' | 'final'
  spanId?: string
  messageId?: string
  isAi?: boolean
}

// Re-export from socket types
export type TerminalServerMessage =
  | { type: 'output'; data: string }
  | { type: 'error'; message: string }
  | { type: 'exit'; code: number }
  | { type: 'connected'; sessionId: string }
  | { type: 'terminal_chat_stream'; payload: TerminalChatStreamPayload }