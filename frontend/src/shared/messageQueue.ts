// Message queue buffer for WebSocket reconnection scenarios
import type { ChatWebSocketMessage } from '@/shared/socket/chat'

interface QueuedMessage {
  message: ChatWebSocketMessage
  timestamp: number
}

const MAX_QUEUE_SIZE = 100
const QUEUE_EXPIRY_MS = 30000 // 30 seconds

const messageQueue: QueuedMessage[] = []
const activeWorkspaces = new Set<string>()
const listeners = new Set<(message: ChatWebSocketMessage) => void>()

export function registerWorkspaceActive(workspaceId: string): void {
  activeWorkspaces.add(workspaceId)
  processQueueForWorkspace(workspaceId)
}

export function unregisterWorkspaceActive(workspaceId: string): void {
  activeWorkspaces.delete(workspaceId)
}

export function enqueueMessage(message: ChatWebSocketMessage): void {
  pruneExpiredMessages()

  // If we have active listeners and workspace is active, deliver directly
  if (listeners.size > 0 && activeWorkspaces.has(message.workspaceId)) {
    deliverToListeners(message)
    return
  }

  // Otherwise, queue the message
  if (messageQueue.length < MAX_QUEUE_SIZE) {
    messageQueue.push({
      message,
      timestamp: Date.now()
    })
  }
}

export function onMessage(listener: (message: ChatWebSocketMessage) => void): () => void {
  listeners.add(listener)
  // Process any queued messages on new listener
  processQueue()
  return () => listeners.delete(listener)
}

function deliverToListeners(message: ChatWebSocketMessage): void {
  listeners.forEach((fn) => {
    try {
      fn(message)
    } catch (e) {
      console.error('Message listener error:', e)
    }
  })
}

function processQueue(): void {
  pruneExpiredMessages()

  const toProcess = messageQueue.splice(0, messageQueue.length)
  for (const item of toProcess) {
    if (activeWorkspaces.has(item.message.workspaceId)) {
      deliverToListeners(item.message)
    }
  }
}

function processQueueForWorkspace(workspaceId: string): void {
  pruneExpiredMessages()

  const workspaceMessages = messageQueue.filter(
    (item) => item.message.workspaceId === workspaceId
  )

  for (const item of workspaceMessages) {
    const idx = messageQueue.indexOf(item)
    if (idx !== -1) {
      messageQueue.splice(idx, 1)
    }
    deliverToListeners(item.message)
  }
}

function pruneExpiredMessages(): void {
  const now = Date.now()
  while (messageQueue.length > 0 && now - messageQueue[0].timestamp > QUEUE_EXPIRY_MS) {
    messageQueue.shift()
  }
}

export function getQueueStatus(): {
  queueLength: number
  activeWorkspaces: string[]
  listenerCount: number
} {
  return {
    queueLength: messageQueue.length,
    activeWorkspaces: Array.from(activeWorkspaces),
    listenerCount: listeners.size
  }
}

export function clearQueue(): void {
  messageQueue.length = 0
}