import { defineStore } from 'pinia'
import { ref } from 'vue'
import client from '@/shared/api/client'
import { notifyUserError } from '@/shared/notifyError'

export interface TerminalSession {
  id: string
  name: string
  status: 'connecting' | 'connected' | 'disconnected'
  memberId?: string
  terminalType?: 'native' | 'web' | 'claude' | 'custom' | 'gemini' | 'aider'
  keepAlive?: boolean
}

export interface TabOptions {
  title?: string
  memberId?: string
  terminalType?: 'native' | 'web' | 'claude' | 'custom' | 'gemini' | 'aider'
  keepAlive?: boolean
  activate?: boolean
}

// Connection state for each session
const connections = ref<Map<string, WebSocket>>(new Map())

export const useTerminalStore = defineStore('terminal', () => {
  const sessions = ref<TerminalSession[]>([])
  const activeSessionId = ref<string | null>(null)
  const currentWorkspaceId = ref<string | null>(null)

  /**
   * Load existing sessions for the workspace
   */
  async function loadSessions(wsId: string) {
    currentWorkspaceId.value = wsId
    try {
      const response = await client.get(`/workspaces/${wsId}/terminal-sessions`)
      // API returns { sessions: [{ memberId, sessionId, pid }] }
      const sessionsList = response.data?.sessions || []
      sessions.value = sessionsList.map((s: any) => ({
        id: s.sessionId,
        name: 'bash',
        status: 'disconnected'
      }))

      if (sessions.value.length > 0 && !activeSessionId.value) {
        activeSessionId.value = sessions.value[0].id
      }
    } catch (e) {
      console.warn('Failed to fetch existing terminal sessions')
    }
  }

  /**
   * Create a brand new PTY session
   */
  async function createSession(name: string = 'bash', memberId?: string) {
    if (!currentWorkspaceId.value) return
    
    try {
      const response = await client.post('/terminals', {
        workspaceId: currentWorkspaceId.value,
        memberId: memberId || 'default',
        command: '/bin/bash'
      })
      
      const newSession: TerminalSession = {
        id: response.data.sessionId,
        name: name,
        status: 'connecting'
      }
      
      sessions.value.push(newSession)
      activeSessionId.value = newSession.id
      return newSession.id
    } catch (e) {
      notifyUserError('Failed to start terminal session', e)
    }
  }

  function setActiveSession(id: string) {
    activeSessionId.value = id
  }

  function updateSessionStatus(id: string, status: TerminalSession['status']) {
    const session = sessions.value.find(s => s.id === id)
    if (session) session.status = status
  }

  function closeSession(id: string) {
    // Close WebSocket connection first
    closeConnection(id)
    // Delete from backend
    client.delete(`/terminals/${id}`).catch(() => {})
    sessions.value = sessions.value.filter(s => s.id !== id)
    if (activeSessionId.value === id) {
      activeSessionId.value = sessions.value[0]?.id || null
    }
  }

  /**
   * Open a terminal tab (register session in store)
   */
  function openTab(sessionId: string, options?: TabOptions) {
    // Check if session already exists
    const existing = sessions.value.find(s => s.id === sessionId)
    if (existing) {
      // Update existing session with new options
      if (options?.title && existing.name === 'bash') {
        existing.name = options.title
      }
      if (options?.memberId) existing.memberId = options.memberId
      if (options?.terminalType) existing.terminalType = options.terminalType
      if (options?.keepAlive !== undefined) existing.keepAlive = options.keepAlive
      if (options?.activate) {
        activeSessionId.value = sessionId
      }
      return
    }

    // Create new session entry
    const newSession: TerminalSession = {
      id: sessionId,
      name: options?.title || 'bash',
      status: 'connecting',
      memberId: options?.memberId,
      terminalType: options?.terminalType,
      keepAlive: options?.keepAlive ?? false
    }
    sessions.value.push(newSession)
    if (options?.activate !== false) {
      activeSessionId.value = sessionId
    }
  }

  /**
   * Get WebSocket connection for a session
   */
  function getConnection(sessionId: string): WebSocket | undefined {
    return connections.value.get(sessionId)
  }

  /**
   * Create WebSocket connection for a session
   */
  async function createConnection(sessionId: string): Promise<WebSocket> {
    // Close existing connection if any
    closeConnection(sessionId)

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host === 'localhost:5175' ? 'localhost:8080' : window.location.host
    const wsUrl = `${protocol}//${host}/ws/terminal/${sessionId}`

    return new Promise((resolve, reject) => {
      const ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        connections.value.set(sessionId, ws)
        updateSessionStatus(sessionId, 'connected')
        resolve(ws)
      }

      ws.onerror = (error) => {
        updateSessionStatus(sessionId, 'disconnected')
        reject(error)
      }

      ws.onclose = () => {
        connections.value.delete(sessionId)
        updateSessionStatus(sessionId, 'disconnected')
      }
    })
  }

  /**
   * Close WebSocket connection for a session
   */
  function closeConnection(sessionId: string) {
    const ws = connections.value.get(sessionId)
    if (ws) {
      ws.close()
      connections.value.delete(sessionId)
    }
  }

  return {
    sessions,
    activeSessionId,
    loadSessions,
    createSession,
    setActiveSession,
    updateSessionStatus,
    closeSession,
    openTab,
    getConnection,
    createConnection,
    closeConnection
  }
})
