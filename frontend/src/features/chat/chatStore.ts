import { computed, ref } from 'vue'
import { acceptHMRUpdate, defineStore } from 'pinia'
import client from '@/shared/api/client'
import { notifyUserError } from '@/shared/notifyError'
import { useAuthStore } from '@/features/auth/authStore'
import { useProjectStore } from '@/features/workspace/projectStore'
import type { Conversation } from '@/shared/types/chat'

export interface AgentStatus {
  memberId: string
  status: 'thinking' | 'reading_file' | 'writing_code' | 'idle' | 'error'
  message?: string
}

export interface ChatWsEvent {
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

export const useChatStore = defineStore('chat', () => {
  const conversations = ref<Conversation[]>([])
  const activeConversationId = ref<string | null>(null)
  const loading = ref(false)
  const loadingMessages = ref(false)
  const workspaceId = ref<string | null>(null)
  const agentStatuses = ref<Record<string, AgentStatus>>({})
  const connectionStatus = ref<'connected' | 'disconnected' | 'reconnecting'>('disconnected')

  // Pagination state for active conversation
  const oldestMessageId = ref<string | null>(null)
  const hasMoreMessages = ref(true)
  const MESSAGE_PAGE_SIZE = 50
  const MESSAGE_INITIAL_SIZE = 30

  const authStore = useAuthStore()

  let chatWs: WebSocket | null = null
  let chatWsReconnectTimer: ReturnType<typeof setTimeout> | null = null
  let chatWsReconnectAttempts = 0
  const CHAT_WS_MAX_RECONNECT_ATTEMPTS = 10

  const currentUserId = computed(() => authStore.currentUserId || 'default')

  const activeConversation = computed(() =>
    conversations.value.find(c => c.id === activeConversationId.value) || null
  )

  const sortedConversations = computed(() => {
    return [...conversations.value].sort((a, b) => {
      if ((a.unreadCount || 0) !== (b.unreadCount || 0)) {
        return (b.unreadCount || 0) - (a.unreadCount || 0)
      }
      return (b.lastMessageAt || 0) - (a.lastMessageAt || 0)
    })
  })

  /**
   * Load conversations via HTTP (history) then connect WebSocket for real-time updates
   */
  async function loadConversations(wsId: string) {
    workspaceId.value = wsId
    loading.value = true
    try {
      // 1. Fetch conversation list + messages via HTTP (load history)
      await loadConversationsList(wsId)
      if (activeConversationId.value) {
        await loadMessages(wsId, activeConversationId.value)
      }
      // 2. Connect WebSocket for real-time updates
      connectChatWebSocket(wsId)
    } catch (e) {
      notifyUserError('Failed to load chat session', e)
    } finally {
      loading.value = false
    }
  }

  async function loadConversationsList(wsId: string) {
    const convResponse = await client.get(`/workspaces/${wsId}/conversations`)
    const remoteConvs = convResponse.data.timeline || []
    conversations.value = remoteConvs.map((rc: any) => {
      const local = conversations.value.find(lc => lc.id === rc.id)
      return {
        ...rc,
        messages: local?.messages || []
      }
    })
    if (!activeConversationId.value && convResponse.data.defaultChannelId) {
      activeConversationId.value = convResponse.data.defaultChannelId
    }
  }

  /**
   * Connect to chat WebSocket for real-time message broadcasting
   */
  function connectChatWebSocket(wsId: string) {
    // Skip if already connected to the same workspace
    if (chatWs?.readyState === WebSocket.OPEN) {
      console.log('[ChatWS] already connected, skipping')
      return
    }

    // Close existing connection
    if (chatWs) {
      // Clear handlers before closing to prevent stale onclose from firing
      chatWs.onclose = null
      chatWs.onerror = null
      chatWs.onmessage = null
      chatWs.onopen = null
      chatWs.close()
      chatWs = null
    }
    if (chatWsReconnectTimer) {
      clearTimeout(chatWsReconnectTimer)
      chatWsReconnectTimer = null
    }
    chatWsReconnectAttempts = 0

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host === 'localhost:5175' ? 'localhost:8080' : window.location.host
    const token = localStorage.getItem('orchestra.auth.token') || ''
    const wsUrl = `${protocol}//${host}/ws/chat/${wsId}?token=${encodeURIComponent(token)}`

    connectionStatus.value = 'reconnecting'
    chatWs = new WebSocket(wsUrl)

    chatWs.onopen = () => {
      connectionStatus.value = 'connected'
      chatWsReconnectAttempts = 0
      console.log('[ChatWS] connected')
    }

    chatWs.onmessage = (event: MessageEvent) => {
      try {
        const data: ChatWsEvent = JSON.parse(event.data)
        handleWsEvent(data)
      } catch (e) {
        console.warn('[ChatWS] parse error', e)
      }
    }

    chatWs.onerror = () => {
      connectionStatus.value = 'disconnected'
      console.warn('[ChatWS] error')
    }

    chatWs.onclose = () => {
      connectionStatus.value = 'disconnected'
      if (chatWsReconnectAttempts >= CHAT_WS_MAX_RECONNECT_ATTEMPTS) {
        console.log('[ChatWS] max reconnect attempts reached')
        return
      }
      chatWsReconnectAttempts++
      const delay = Math.min(1000 * 2 ** chatWsReconnectAttempts, 30000)
      console.log(`[ChatWS] reconnecting in ${delay}ms (attempt ${chatWsReconnectAttempts}/${CHAT_WS_MAX_RECONNECT_ATTEMPTS})`)
      chatWsReconnectTimer = setTimeout(() => connectChatWebSocket(wsId), delay)
    }
  }

  /**
   * Handle incoming WebSocket chat events
   */
  function handleWsEvent(event: ChatWsEvent) {
    if (!event.workspaceId || event.workspaceId !== workspaceId.value) return

    switch (event.type) {
      case 'new_message': {
        // Find the conversation and append/refresh messages
        const convIndex = conversations.value.findIndex(c => c.id === event.conversationId)
        if (convIndex !== -1) {
          // If it's the active conversation, append the new message directly (don't reload all)
          if (event.conversationId === activeConversationId.value && event.messageId) {
            const conv = conversations.value[convIndex]
            // Check if message already exists (avoid duplicates from rapid sends)
            const exists = conv.messages?.some(m => m.id === event.messageId)
            if (!exists) {
              const newMsg = {
                id: event.messageId,
                senderId: event.senderId,
                senderName: event.senderName ?? '',
                content: { type: 'text' as const, text: event.content || '' },
                isAi: event.isAi || false,
                createdAt: event.createdAt || Date.now()
              }
              conv.messages = [...(conv.messages || []), newMsg]
            }
          } else {
            // For inactive conversations, just refresh the list to update lastMessage/unread
            loadConversationsList(workspaceId.value!)
          }
        } else {
          // New conversation appeared
          loadConversationsList(workspaceId.value!)
        }
        break
      }
      case 'message_status': {
        // Agent status update — handled via agentStatuses record
        if (event.senderId) {
          agentStatuses.value[event.senderId] = {
            memberId: event.senderId,
            status: (event.status as AgentStatus['status']) || 'idle',
            message: event.content
          }
        }
        break
      }
      case 'unread_sync': {
        if (event.conversationId) {
          const conv = conversations.value.find(c => c.id === event.conversationId)
          if (conv) conv.unreadCount = event.unreadCount ?? 0
        }
        break
      }
    }
  }

  function disconnectChatWebSocket() {
    if (chatWs) {
      chatWs.close()
      chatWs = null
    }
    if (chatWsReconnectTimer) {
      clearTimeout(chatWsReconnectTimer)
      chatWsReconnectTimer = null
    }
    connectionStatus.value = 'disconnected'
  }

  function reconnectChatWebSocket() {
    if (workspaceId.value) {
      connectChatWebSocket(workspaceId.value)
    }
  }

  async function loadMessages(wsId: string, convId: string, beforeId?: string) {
    loadingMessages.value = true
    try {
      const params = new URLSearchParams()
      params.set('limit', String(beforeId ? MESSAGE_PAGE_SIZE : MESSAGE_INITIAL_SIZE))
      if (beforeId) params.set('beforeId', beforeId)

      const response = await client.get(`/workspaces/${wsId}/conversations/${convId}/messages?${params}`)
      const index = conversations.value.findIndex(c => c.id === convId)
      if (index !== -1) {
        const newMessages = response.data || []
        if (beforeId) {
          // Prepend older messages
          const existing = conversations.value[index].messages || []
          conversations.value[index] = {
            ...conversations.value[index],
            messages: [...newMessages, ...existing]
          }
        } else {
          conversations.value[index] = {
            ...conversations.value[index],
            messages: newMessages
          }
        }

        // Track oldest message for pagination
        if (newMessages.length > 0) {
          oldestMessageId.value = newMessages[0].id
          hasMoreMessages.value = newMessages.length >= MESSAGE_PAGE_SIZE
        } else {
          hasMoreMessages.value = false
        }
      }
    } catch (e) { /* silent during loading */ }
    finally {
      loadingMessages.value = false
    }
  }

  async function loadOlderMessages(wsId: string, convId: string) {
    if (!hasMoreMessages.value || !oldestMessageId.value) return
    await loadMessages(wsId, convId, oldestMessageId.value)
  }

  async function setActiveConversation(id: string) {
    activeConversationId.value = id
    // Reset pagination state for new conversation
    oldestMessageId.value = null
    hasMoreMessages.value = true
    if (workspaceId.value) {
      await loadMessages(workspaceId.value, id)
      await markAsRead(workspaceId.value, id)
    }
  }

  async function sendMessage(payload: { text: string, conversationId: string }) {
    if (!workspaceId.value) return
    try {
      let senderId = currentUserId.value
      const senderName = authStore.currentUser || 'User'

      if (senderId === 'default') {
        const projectStore = useProjectStore()
        const owner = projectStore.members.find((m: any) => m.roleType === 'owner')
        if (owner) {
          senderId = owner.id
        }
      }

      await client.post(`/workspaces/${workspaceId.value}/conversations/${payload.conversationId}/messages`, {
        text: payload.text,
        senderId,
        senderName
      })
      // Immediate pull after send for optimistic consistency
      await loadMessages(workspaceId.value, payload.conversationId)
    } catch (e) {
      notifyUserError('Failed to send message', e)
    }
  }

  async function createConversation(data: { type: 'channel' | 'dm'; name?: string; memberId?: string }) {
    if (!workspaceId.value) return null
    try {
      const memberIDs = data.memberId
        ? [currentUserId.value, data.memberId]
        : [currentUserId.value]

      const response = await client.post(`/workspaces/${workspaceId.value}/conversations`, {
        type: data.type,
        name: data.name || '',
        memberIDs
      })

      const newConv = response.data
      // Add to local state with empty messages array
      conversations.value = [...conversations.value, { ...newConv, messages: [] }]
      return newConv.id
    } catch (e) {
      notifyUserError('Failed to create conversation', e)
      return null
    }
  }

  async function markAsRead(wsId: string, convId: string) {
    try {
      await client.post(`/workspaces/${wsId}/conversations/${convId}/read`, {
        userId: currentUserId.value
      })
      const index = conversations.value.findIndex(c => c.id === convId)
      if (index !== -1) conversations.value[index].unreadCount = 0
    } catch (e) { /* silent */ }
  }

  async function updatePresence(activity: 'typing' | 'viewing' | 'idle', targetId: string) {
    if (!workspaceId.value) return
    try {
      await client.post(`/workspaces/${workspaceId.value}/members/${currentUserId.value}/presence`, {
        activity,
        targetId,
        targetType: 'conversation'
      })
    } catch (e) { /* silent */ }
  }

  return {
    conversations,
    activeConversationId,
    activeConversation,
    sortedConversations,
    agentStatuses,
    loading,
    loadingMessages,
    hasMoreMessages,
    oldestMessageId,
    connectionStatus,
    currentUserId,
    workspaceId,
    loadConversations,
    setActiveConversation,
    sendMessage,
    createConversation,
    updatePresence,
    disconnectChatWebSocket,
    reconnectChatWebSocket,
    loadOlderMessages,
    getConversationTitle: (c: any) => c.customName || c.id
  }
})

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useChatStore, import.meta.hot))
}
