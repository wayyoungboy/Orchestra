import type { TerminalClientMessage } from './types'
import type { TerminalServerMessage } from '@/shared/types/terminal'
import { notifyUserError } from '@/shared/notifyError'

type MessageHandler = (message: TerminalServerMessage) => void
type ErrorHandler = (error: Event) => void

export class TerminalSocket {
  private ws: WebSocket | null = null
  private messageHandlers: Set<MessageHandler> = new Set()
  private errorHandlers: Set<ErrorHandler> = new Set()
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private sessionId: string | null = null

  connect(sessionId: string): Promise<void> {
    return new Promise((resolve, reject) => {
      this.sessionId = sessionId
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/ws/terminal/${sessionId}`

      this.ws = new WebSocket(wsUrl)

      this.ws.onopen = () => {
        this.reconnectAttempts = 0
        resolve()
      }

      this.ws.onerror = (error) => {
        this.errorHandlers.forEach((handler) => handler(error))
        reject(error)
      }

      this.ws.onmessage = (event) => {
        try {
          const message: TerminalServerMessage = JSON.parse(event.data)
          this.messageHandlers.forEach((handler) => handler(message))
        } catch (e) {
          notifyUserError('Parse terminal WebSocket message', e)
        }
      }

      this.ws.onclose = () => {
        // Auto-reconnect logic
        if (this.reconnectAttempts < this.maxReconnectAttempts && this.sessionId) {
          this.reconnectAttempts++
          setTimeout(() => {
            if (this.sessionId) {
              this.connect(this.sessionId).catch((e) => notifyUserError('Terminal WebSocket reconnect', e))
            }
          }, 1000 * this.reconnectAttempts)
        }
      }
    })
  }

  send(message: TerminalClientMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    }
  }

  input(data: string): void {
    this.send({ type: 'input', data })
  }

  resize(cols: number, rows: number): void {
    this.send({ type: 'resize', cols, rows })
  }

  close(): void {
    this.send({ type: 'close' })
    this.ws?.close()
    this.ws = null
    this.sessionId = null
  }

  onMessage(handler: MessageHandler): () => void {
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
}