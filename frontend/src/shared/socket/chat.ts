// ChatSocket - WebSocket connection for real-time chat messages
import { notifyUserError } from '@/shared/notifyError'

type ChatMessageHandler = (message: ChatWebSocketMessage) => void
type ErrorHandler = (error: Event) => void

// Chat WebSocket message types
export interface ChatWebSocketMessage {
  type: 'new_message' | 'message_status' | 'unread_sync'
  workspaceId: string
  conversationId?: string
  messageId?: string
  senderId?: string
  senderName?: string
  content?: string
  createdAt?: number
  isAi?: boolean
  status?: string
  unreadCount?: number
}

export class ChatSocket {
  private ws: WebSocket | null = null
  private messageHandlers: Set<ChatMessageHandler> = new Set()
  private errorHandlers: Set<ErrorHandler> = new Set()
  private reconnectAttempts = 0
  private maxReconnectAttempts = 10
  private workspaceId: string | null = null
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null

  private readonly HEARTBEAT_INTERVAL = 30000 // 30 seconds

  connect(workspaceId: string): Promise<void> {
    return new Promise((resolve, reject) => {
      this.workspaceId = workspaceId
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/ws/chat/${workspaceId}`

      this.ws = new WebSocket(wsUrl)

      this.ws.onopen = () => {
        this.reconnectAttempts = 0
        this.startHeartbeat()
        resolve()
      }

      this.ws.onerror = (error) => {
        this.stopHeartbeat()
        this.errorHandlers.forEach((handler) => handler(error))
        reject(error)
      }

      this.ws.onmessage = (event) => {
        try {
          const message: ChatWebSocketMessage = JSON.parse(event.data)
          this.messageHandlers.forEach((handler) => handler(message))
        } catch (e) {
          notifyUserError('Parse chat WebSocket message', e)
        }
      }

      this.ws.onclose = () => {
        this.stopHeartbeat()
        this.scheduleReconnect()
      }
    })
  }

  private startHeartbeat() {
    this.stopHeartbeat()
    this.heartbeatTimer = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'ping' }))
      }
    }, this.HEARTBEAT_INTERVAL)
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
  }

  private scheduleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts && this.workspaceId) {
      this.reconnectAttempts++
      const delay = Math.min(1000 * this.reconnectAttempts, 10000) // Max 10 seconds
      this.reconnectTimer = setTimeout(() => {
        if (this.workspaceId) {
          this.connect(this.workspaceId).catch((e) => {
            notifyUserError('Chat WebSocket reconnect', e)
          })
        }
      }, delay)
    }
  }

  close(): void {
    this.stopHeartbeat()
    this.ws?.close()
    this.ws = null
    this.workspaceId = null
  }

  onMessage(handler: ChatMessageHandler): () => void {
    this.messageHandlers.add(handler)
    return () => this.messageHandlers.delete(handler)
  }

  onError(handler: ErrorHandler): () => void {
    this.errorHandlers.add(handler)
    return () => this.errorHandlers.delete(handler)
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }

  get currentWorkspaceId(): string | null {
    return this.workspaceId
  }
}

// Global singleton for chat WebSocket
let globalChatSocket: ChatSocket | null = null

export function getChatSocket(): ChatSocket {
  if (!globalChatSocket) {
    globalChatSocket = new ChatSocket()
  }
  return globalChatSocket
}

export function closeChatSocket(): void {
  if (globalChatSocket) {
    globalChatSocket.close()
    globalChatSocket = null
  }
}