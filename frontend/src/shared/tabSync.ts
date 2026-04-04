// Cross-tab synchronization using BroadcastChannel API
// Implements leader election to ensure only one tab manages WebSocket connections

const TAB_ID = `tab-${Date.now()}-${Math.random().toString(36).slice(2)}`
const LEADER_KEY = 'orchestra-chat-leader'
const LEADER_HEARTBEAT_KEY = 'orchestra-chat-leader-heartbeat'
const HEARTBEAT_INTERVAL = 5000 // 5 seconds
const LEADER_TIMEOUT = 15000 // 15 seconds without heartbeat = leader is dead

type SyncMessageType =
  | 'message_received'
  | 'socket_connected'
  | 'socket_disconnected'
  | 'leader_announce'
  | 'leader_claim'

interface SyncMessage {
  type: SyncMessageType
  sourceTabId: string
  timestamp: number
  payload?: unknown
}

let channel: BroadcastChannel | null = null
let isLeader = false
let heartbeatTimer: ReturnType<typeof setInterval> | null = null
let leaderCheckTimer: ReturnType<typeof setInterval> | null = null
const listeners = new Set<(msg: SyncMessage) => void>()

export function getTabId(): string {
  return TAB_ID
}

export function isLeaderTab(): boolean {
  return isLeader
}

export function initTabSync(): void {
  if (channel) return

  const hasBroadcastChannel = typeof BroadcastChannel !== 'undefined'
  if (!hasBroadcastChannel) {
    console.log('[TabSync] BroadcastChannel not supported, running in standalone mode')
    isLeader = true // Act as leader when no BroadcastChannel
    return
  }

  channel = new BroadcastChannel('orchestra-chat-sync')
  channel.onmessage = (event) => {
    const msg = event.data as SyncMessage
    if (msg.sourceTabId === TAB_ID) return // Ignore own messages
    listeners.forEach((fn) => fn(msg))
  }

  // Start leader election
  startLeaderElection()
}

function startLeaderElection(): void {
  // Check if there's already a leader
  const currentLeader = localStorage.getItem(LEADER_KEY)
  const leaderHeartbeat = localStorage.getItem(LEADER_HEARTBEAT_KEY)

  if (currentLeader && leaderHeartbeat) {
    const lastBeat = parseInt(leaderHeartbeat, 10)
    if (Date.now() - lastBeat < LEADER_TIMEOUT) {
      // There's an active leader, become follower
      isLeader = false
      startLeaderCheck()
      return
    }
  }

  // No active leader, try to become one
  claimLeadership()
}

function claimLeadership(): void {
  localStorage.setItem(LEADER_KEY, TAB_ID)
  isLeader = true
  startHeartbeat()
  broadcast({ type: 'leader_announce', sourceTabId: TAB_ID, timestamp: Date.now() })
}

function startHeartbeat(): void {
  if (heartbeatTimer) clearInterval(heartbeatTimer)
  heartbeatTimer = setInterval(() => {
    if (isLeader) {
      localStorage.setItem(LEADER_HEARTBEAT_KEY, Date.now().toString())
    }
  }, HEARTBEAT_INTERVAL)
}

function startLeaderCheck(): void {
  if (leaderCheckTimer) clearInterval(leaderCheckTimer)
  leaderCheckTimer = setInterval(() => {
    const currentLeader = localStorage.getItem(LEADER_KEY)
    const leaderHeartbeat = localStorage.getItem(LEADER_HEARTBEAT_KEY)

    if (currentLeader && leaderHeartbeat) {
      const lastBeat = parseInt(leaderHeartbeat, 10)
      if (Date.now() - lastBeat > LEADER_TIMEOUT) {
        // Leader is dead, try to take over
        claimLeadership()
      }
    } else {
      // No leader, try to become one
      claimLeadership()
    }
  }, HEARTBEAT_INTERVAL)
}

export function broadcast(payload: SyncMessage): void {
  if (channel) {
    channel.postMessage(payload)
  }
}

export function broadcastMessage(message: unknown): void {
  broadcast({
    type: 'message_received',
    sourceTabId: TAB_ID,
    timestamp: Date.now(),
    payload: message
  })
}

export function onTabSyncMessage(listener: (msg: SyncMessage) => void): () => void {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

export function closeTabSync(): void {
  if (heartbeatTimer) {
    clearInterval(heartbeatTimer)
    heartbeatTimer = null
  }
  if (leaderCheckTimer) {
    clearInterval(leaderCheckTimer)
    leaderCheckTimer = null
  }
  if (isLeader) {
    localStorage.removeItem(LEADER_KEY)
    localStorage.removeItem(LEADER_HEARTBEAT_KEY)
  }
  channel?.close()
  channel = null
  isLeader = false
  listeners.clear()
}

// Handle page unload
if (typeof window !== 'undefined') {
  window.addEventListener('beforeunload', () => {
    closeTabSync()
  })
}