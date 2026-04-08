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

export const useChatStore = defineStore('chat', () => {
  const conversations = ref<Conversation[]>([])
  const activeConversationId = ref<string | null>(null)
  const loading = ref(false)
  const workspaceId = ref<string | null>(null)
  const agentStatuses = ref<Record<string, AgentStatus>>({})
  const authStore = useAuthStore()

  let pollingTimer: any = null

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
   * Load session and start polling
   */
  async function loadConversations(wsId: string) {
    workspaceId.value = wsId
    loading.value = true
    try {
      await refreshAllData()
      startPolling()
    } catch (e) {
      notifyUserError('Failed to load chat session', e)
    } finally {
      loading.value = false
    }
  }

  /**
   * High-frequency polling for messages and status
   */
  function startPolling() {
    stopPolling()
    pollingTimer = setInterval(async () => {
      await refreshAllData()
    }, 10000) // 10 seconds interval to reduce CPU load
  }

  function stopPolling() {
    if (pollingTimer) {
      clearInterval(pollingTimer)
      pollingTimer = null
    }
  }

  async function refreshAllData() {
    if (!workspaceId.value) return
    try {
      // 1. Refresh conversation list (includes unread counts and agent statuses)
      const convResponse = await client.get(`/workspaces/${workspaceId.value}/conversations`)
      const remoteConvs = convResponse.data.timeline || []
      
      // Merge with local state to preserve message history while updating metadata
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

      // 2. Refresh active conversation messages
      if (activeConversationId.value) {
        await loadMessages(workspaceId.value, activeConversationId.value)
      }
    } catch (e) {
      console.warn('Polling tick failed', e)
    }
  }

  async function loadMessages(wsId: string, convId: string) {
    try {
      const response = await client.get(`/workspaces/${wsId}/conversations/${convId}/messages`)
      const index = conversations.value.findIndex(c => c.id === convId)
      if (index !== -1) {
        // Only update if message count changed or new content
        const remoteMsgs = response.data || []
        if (JSON.stringify(conversations.value[index].messages) !== JSON.stringify(remoteMsgs)) {
          conversations.value[index] = {
            ...conversations.value[index],
            messages: remoteMsgs
          }
        }
      }
    } catch (e) { /* silent during polling */ }
  }

  async function setActiveConversation(id: string) {
    activeConversationId.value = id
    if (workspaceId.value) {
      await loadMessages(workspaceId.value, id)
      await markAsRead(workspaceId.value, id)
    }
  }

  async function sendMessage(payload: { text: string, conversationId: string }) {
    if (!workspaceId.value) return
    try {
      // Get real senderId: if currentUserId is 'default', find Owner from member list
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
      // Immediate pull after send
      await loadMessages(workspaceId.value, payload.conversationId)
    } catch (e) {
      notifyUserError('Failed to send message', e)
    }
  }

  async function markAsRead(wsId: string, convId: string) {
    try {
      await client.post(`/workspaces/${wsId}/conversations/${convId}/read`, {})
      const index = conversations.value.findIndex(c => c.id === convId)
      if (index !== -1) conversations.value[index].unreadCount = 0
    } catch (e) { /* silent */ }
  }

  async function updatePresence(activity: 'typing' | 'viewing' | 'idle', targetId: string) {
    if (!workspaceId.value) return
    try {
      await client.post(`/workspaces/${workspaceId.value}/members/me/presence`, {
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
    currentUserId,
    loadConversations,
    setActiveConversation,
    sendMessage,
    updatePresence,
    stopPolling,
    getConversationTitle: (c: any) => c.customName || c.id
  }
})

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useChatStore, import.meta.hot))
}
