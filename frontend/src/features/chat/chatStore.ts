// 聊天状态管理：负责会话列表、消息流与终端联动
import { computed, ref, watch } from 'vue'
import { acceptHMRUpdate, defineStore, storeToRefs } from 'pinia'
import client from '@/shared/api/client'
import { getApiErrorMessage } from '@/shared/api/errors'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'
import { useProjectStore } from '@/features/workspace/projectStore'
import { useSettingsStore } from '@/features/settings/settingsStore'
import { getChatSocket, closeChatSocket, type ChatWebSocketMessage } from '@/shared/socket/chat'
import {
  registerWorkspaceActive,
  unregisterWorkspaceActive,
  onMessage as onQueuedMessage
} from '@/shared/messageQueue'
import {
  initTabSync,
  isLeaderTab,
  broadcastMessage,
  onTabSyncMessage,
  closeTabSync
} from '@/shared/tabSync'
import type {
  Conversation,
  ConversationDto,
  Message,
  MessageDto,
  MessageMentionsPayload,
  MessageContent,
  ChatListResponse,
  ChatMessageCreatedPayload,
  ChatMessageStatusPayload,
  ChatUnreadSyncPayload
} from '@/shared/types/chat'
import type { TerminalChatStreamPayload } from '@/shared/types/terminal'
import { stripAnsiForChat } from '@/shared/utils/stripAnsiForChat'

const MAX_MESSAGE_LENGTH = 1200
const MESSAGES_PAGE_LIMIT = 200
const POLL_INTERVAL_MS = 3000 // 轮询间隔

// 防止并发加载覆盖状态的序号标记
let loadSequence = 0

/**
 * 聊天状态存储
 * 输入：会话加载、消息发送与会话操作
 * 输出：会话列表与操作方法
 */
export const useChatStore = defineStore('chat', () => {
  const workspaceStore = useWorkspaceStore()
  const projectStore = useProjectStore()
  const settingsStore = useSettingsStore()
  const { currentWorkspace } = storeToRefs(workspaceStore)
  const { members } = storeToRefs(projectStore)
  const { settings } = storeToRefs(settingsStore)

  const streamOutputEnabled = computed(() => settings.value.chat.streamOutput)
  const streamMessageIds = new Map<string, string>()

  const conversations = ref<Conversation[]>([])
  const activeConversationId = ref<string | null>(null)
  const isReady = ref(false)
  const chatError = ref<string | null>(null)
  const defaultChannelId = ref<string | null>(null)
  const totalUnreadCount = ref(0)
  const loading = ref(false)

  const loadedMessages = new Set<string>()
  const loadingMessages = new Set<string>()
  const conversationPaging = new Map<string, { hasMore: boolean; loading: boolean }>()

  // 消息轮询定时器
  let pollTimer: ReturnType<typeof setInterval> | null = null
  let pollSequence = 0 // 防止竞态

  const currentUserId = computed(() => members.value?.find((m) => m.roleType === 'owner')?.id ?? 'owner')

  const activeConversation = computed(() =>
    conversations.value.find((c) => c.id === activeConversationId.value)
  )

  const sortedConversations = computed(() =>
    [...conversations.value].sort((a, b) => {
      if (a.pinned !== b.pinned) {
        return a.pinned ? -1 : 1
      }
      const timeA = a.lastMessageAt ?? 0
      const timeB = b.lastMessageAt ?? 0
      if (timeA !== timeB) {
        return timeB - timeA
      }
      const nameA = a.customName ?? a.id
      const nameB = b.customName ?? b.id
      return nameA.localeCompare(nameB)
    })
  )

  const resolveConversationTitle = (conversation: Conversation) => {
    if (conversation.type === 'dm') {
      const targetId = conversation.targetId ?? ''
      const target = members.value?.find((member) => member.id === targetId)
      return target?.name ?? 'Member'
    }
    const workspaceLabel = currentWorkspace.value?.name?.trim()
    if ((conversation.isDefault || conversation.id === defaultChannelId.value) && workspaceLabel) {
      return workspaceLabel
    }
    if (conversation.customName) {
      return conversation.customName
    }
    if (conversation.nameKey) {
      return conversation.nameKey
    }
    return conversation.id
  }

  const normalizeConversation = (dto: ConversationDto): Conversation => ({
    id: dto.id,
    type: dto.type,
    targetId: dto.targetId,
    memberIds: dto.memberIds ?? [],
    nameKey: undefined,
    customName: dto.customName ?? undefined,
    descriptionKey: undefined,
    pinned: Boolean(dto.pinned),
    muted: Boolean(dto.muted),
    lastMessageAt: dto.lastMessageAt ?? undefined,
    lastMessagePreview: dto.lastMessagePreview ?? undefined,
    lastMessageSenderId: undefined,
    lastMessageSenderName: undefined,
    lastMessageSenderAvatar: undefined,
    lastMessageAttachment: undefined,
    isDefault: dto.isDefault ?? false,
    unreadCount: dto.unreadCount ?? 0,
    messages: []
  })

  const resolvePreviewText = (content: MessageContent) => {
    if (content.type === 'text') return content.text ?? ''
    return ''
  }

  const buildStreamMessageId = (spanId: string) => `terminal-stream:${spanId}`

  const normalizeMessage = (dto: MessageDto): Message => {
    const senderId = dto.senderId
    const member = senderId && members.value ? members.value.find((m) => m.id === senderId) : undefined
    const user = member?.name ?? ''
    const avatar = member?.avatar ?? ''

    return {
      id: dto.id,
      senderId,
      senderName: user,
      senderAvatar: avatar,
      content: dto.content,
      createdAt: dto.createdAt,
      isAi: dto.isAi,
      attachment: dto.attachment,
      status: dto.status
    }
  }

  const sortConversations = (items: Conversation[]) =>
    [...items].sort((a, b) => {
      if (a.pinned !== b.pinned) {
        return a.pinned ? -1 : 1
      }
      const timeA = a.lastMessageAt ?? 0
      const timeB = b.lastMessageAt ?? 0
      if (timeA !== timeB) {
        return timeB - timeA
      }
      const nameA = a.customName ?? a.id
      const nameB = b.customName ?? b.id
      return nameA.localeCompare(nameB)
    })

  const updateConversation = (conversationId: string, updater: (conversation: Conversation) => Conversation) => {
    conversations.value = conversations.value.map((conversation) =>
      conversation.id === conversationId ? updater(conversation) : conversation
    )
  }

  const updateConversationOrder = () => {
    conversations.value = sortConversations(conversations.value)
  }

  const getPagingState = (conversationId: string) => {
    const state = conversationPaging.get(conversationId)
    if (state) {
      return state
    }
    const next = { hasMore: true, loading: false }
    conversationPaging.set(conversationId, next)
    return next
  }

  const updatePagingState = (conversationId: string, updater: (state: { hasMore: boolean; loading: boolean }) => void) => {
    const state = getPagingState(conversationId)
    updater(state)
    conversationPaging.set(conversationId, state)
  }

  /**
   * 加载会话列表与首页统计
   */
  const loadSession = async () => {
    const workspace = currentWorkspace.value
    if (!workspace) {
      conversations.value = []
      defaultChannelId.value = null
      totalUnreadCount.value = 0
      isReady.value = false
      return
    }

    const requestId = ++loadSequence
    loading.value = true
    isReady.value = false
    chatError.value = null

    try {
      const response = await client.get<ChatListResponse>(
        `/workspaces/${workspace.id}/conversations`,
        {
          params: {
            userId: currentUserId.value
          }
        }
      )
      const feed = response.data

      if (requestId !== loadSequence || workspace.id !== currentWorkspace.value?.id) {
        return
      }

      const merged = [...(feed.pinned ?? []), ...(feed.timeline ?? [])]
      const seen = new Set<string>()
      const normalized: Conversation[] = []
      // Preserve existing messages when reloading conversations
      const existingMessages = new Map<string, Message[]>()
      for (const conv of conversations.value) {
        if (conv.messages && conv.messages.length > 0) {
          existingMessages.set(conv.id, conv.messages)
        }
      }
      for (const dto of merged) {
        if (!dto || !dto.id || seen.has(dto.id)) {
          continue
        }
        seen.add(dto.id)
        const normalizedConv = normalizeConversation(dto)
        // Restore preserved messages
        const savedMessages = existingMessages.get(dto.id)
        if (savedMessages) {
          normalizedConv.messages = savedMessages
        }
        normalized.push(normalizedConv)
      }

      conversations.value = sortConversations(normalized)
      defaultChannelId.value = feed.defaultChannelId ?? null
      totalUnreadCount.value = feed.totalUnreadCount ?? 0
      isReady.value = true

      if (!activeConversationId.value && normalized.length > 0) {
        const sorted = sortConversations(normalized)
        void setActiveConversation(sorted[0].id)
      }
    } catch (error) {
      chatError.value = getApiErrorMessage(error)
    } finally {
      loading.value = false
    }
  }

  /**
   * 加载会话消息首屏
   */
  const loadConversationMessages = async (conversationId: string, force = false) => {
    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId || !conversationId) return
    if (loadingMessages.has(conversationId)) return
    if (!force && loadedMessages.has(conversationId)) return
    loadingMessages.add(conversationId)
    updatePagingState(conversationId, (state) => {
      state.loading = true
    })
    try {
      const response = await client.get<MessageDto[]>(
        `/workspaces/${workspaceId}/conversations/${conversationId}/messages`,
        {
          params: { limit: MESSAGES_PAGE_LIMIT }
        }
      )
      console.log('[chatStore] loadConversationMessages response:', { conversationId, dataLength: response.data?.length, firstMsg: response.data?.[0]?.id })
      const messages = response.data.map((dto) => normalizeMessage(dto))
      updateConversation(conversationId, (conversation) => ({
        ...conversation,
        messages
      }))
      loadedMessages.add(conversationId)
      updatePagingState(conversationId, (state) => {
        state.hasMore = response.data.length >= MESSAGES_PAGE_LIMIT
      })
    } catch (error) {
      chatError.value = getApiErrorMessage(error)
    } finally {
      loadingMessages.delete(conversationId)
      updatePagingState(conversationId, (state) => {
        state.loading = false
      })
    }
  }

  /**
   * 加载更早的会话消息
   */
  const loadOlderMessages = async (conversationId: string) => {
    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId || !conversationId) return
    const paging = getPagingState(conversationId)
    if (paging.loading || !paging.hasMore) return
    const conversation = conversations.value.find((item) => item.id === conversationId)
    const beforeId = conversation?.messages[0]?.id
    updatePagingState(conversationId, (state) => {
      state.loading = true
    })
    try {
      const response = await client.get<MessageDto[]>(
        `/workspaces/${workspaceId}/conversations/${conversationId}/messages`,
        {
          params: {
            limit: MESSAGES_PAGE_LIMIT,
            beforeId
          }
        }
      )
      if (response.data.length === 0) {
        updatePagingState(conversationId, (state) => {
          state.hasMore = false
        })
        return
      }
      const messages = response.data.map((dto) => normalizeMessage(dto))
      updateConversation(conversationId, (conversation) => ({
        ...conversation,
        messages: [...messages, ...conversation.messages]
      }))
      updatePagingState(conversationId, (state) => {
        state.hasMore = response.data.length >= MESSAGES_PAGE_LIMIT
      })
    } catch {
      /* API errors surfaced by axios interceptor + toast */
    } finally {
      updatePagingState(conversationId, (state) => {
        state.loading = false
      })
    }
  }

  /**
   * 发送消息并同步到会话列表
   */
  const sendMessage = async (payload: { text: string; conversationId: string; mentions?: MessageMentionsPayload }) => {
    const trimmed = payload.text.trim()
    if (!trimmed) return null
    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId) return null
    const conversation = conversations.value.find((item) => item.id === payload.conversationId)
    if (!conversation) return null

    const text = trimmed.slice(0, MAX_MESSAGE_LENGTH)
    try {
      const senderName = settings.value.account.displayName || 'Owner'
      const clientTraceId =
        typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
          ? crypto.randomUUID()
          : `${Date.now()}-${Math.random().toString(36).slice(2)}`
      const response = await client.post<MessageDto>(
        `/workspaces/${workspaceId}/conversations/${payload.conversationId}/messages`,
        {
          text,
          senderId: currentUserId.value,
          senderName,
          clientTraceId,
          conversationType: conversation.type,
          mentions: payload.mentions ?? { mentionIds: [], mentionAll: false },
          timestamp: Date.now()
        }
      )
      const message = normalizeMessage(response.data)

      updateConversation(payload.conversationId, (conversation) => ({
        ...conversation,
        messages: [...conversation.messages, message],
        lastMessageAt: message.createdAt,
        lastMessagePreview: message.content.type === 'text' ? message.content.text ?? '' : '',
        lastMessageSenderId: message.senderId,
        lastMessageSenderName: message.senderName,
        lastMessageSenderAvatar: message.senderAvatar,
        unreadCount: conversation.unreadCount
      }))
      updateConversationOrder()

      void ensureConversationLatestMessage(payload.conversationId)

      return message
    } catch {
      return null
    }
  }

  /**
   * 创建私聊会话
   */
  const createDirectConversation = async (memberId: string) => {
    if (!memberId || memberId === currentUserId.value) return null
    if (!members.value?.some((member) => member.id === memberId)) return null

    const existing = conversations.value.find(
      (conversation) => conversation.type === 'dm' && conversation.targetId === memberId
    )
    if (existing) {
      setActiveConversation(existing.id)
      return existing
    }

    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId) return null

    try {
      const response = await client.post<ConversationDto>(
        `/workspaces/${workspaceId}/conversations`,
        {
          type: 'dm',
          memberIds: [currentUserId.value, memberId],
          targetId: memberId
        }
      )
      const conversation = normalizeConversation(response.data)
      conversations.value = sortConversations([...conversations.value, conversation])
      setActiveConversation(conversation.id)
      return conversation
    } catch {
      return null
    }
  }

  /**
   * 创建群聊会话
   */
  const createConversation = async (memberIds: string[], customName?: string) => {
    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId) return null
    const nextMembers = Array.from(new Set([currentUserId.value, ...memberIds]))
    if (nextMembers.length < 2) return null

    try {
      const response = await client.post<ConversationDto>(
        `/workspaces/${workspaceId}/conversations`,
        {
          type: 'channel',
          memberIds: nextMembers,
          name: customName
        }
      )
      const conversation = normalizeConversation(response.data)
      conversations.value = sortConversations([...conversations.value, conversation])
      setActiveConversation(conversation.id)
      return conversation
    } catch {
      return null
    }
  }

  /**
   * 设置当前活动会话
   */
  const setActiveConversation = (conversationId: string | null) => {
    activeConversationId.value = conversationId
    if (conversationId) {
      void loadConversationMessages(conversationId)
      void markConversationRead(conversationId)
      startPolling(conversationId)
    } else {
      stopPolling()
    }
  }

  /**
   * 启动消息轮询
   */
  const startPolling = (conversationId: string) => {
    stopPolling()
    pollSequence += 1
    const seq = pollSequence
    pollTimer = setInterval(() => {
      // 检查是否仍是当前活动会话
      if (activeConversationId.value !== conversationId || pollSequence !== seq) {
        return
      }
      void pollNewMessages(conversationId)
    }, POLL_INTERVAL_MS)
  }

  /**
   * 停止消息轮询
   */
  const stopPolling = () => {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  /**
   * 轮询新消息（检查是否有新消息并添加到当前会话）
   */
  const pollNewMessages = async (conversationId: string) => {
    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId || !conversationId) return

    const conversation = conversations.value.find((c) => c.id === conversationId)
    if (!conversation) return

    // 获取当前最新消息的时间戳
    const currentMessages = conversation.messages
    const lastMessageTime = currentMessages.length > 0
      ? currentMessages[currentMessages.length - 1].createdAt
      : 0

    try {
      const response = await client.get<MessageDto[]>(
        `/workspaces/${workspaceId}/conversations/${conversationId}/messages`,
        { params: { limit: 50 } }
      )
      const dtos = response.data
      if (!dtos?.length) return

      // 找出比当前最新消息更新的消息
      const newMessages = dtos
        .filter((dto) => dto.createdAt > lastMessageTime)
        .map((dto) => normalizeMessage(dto))

      if (newMessages.length === 0) return

      // 合并新消息
      updateConversation(conversationId, (conv) => {
        const existingIds = new Set(conv.messages.map((m) => m.id))
        const toAdd = newMessages.filter((m) => !existingIds.has(m.id))
        if (toAdd.length === 0) return conv
        return {
          ...conv,
          messages: [...conv.messages, ...toAdd],
          lastMessageAt: toAdd[toAdd.length - 1].createdAt,
          lastMessagePreview:
            toAdd[toAdd.length - 1].content.type === 'text'
              ? toAdd[toAdd.length - 1].content.text ?? ''
              : '',
          lastMessageSenderId: toAdd[toAdd.length - 1].senderId,
          lastMessageSenderName: toAdd[toAdd.length - 1].senderName,
          lastMessageSenderAvatar: toAdd[toAdd.length - 1].senderAvatar
        }
      })
      updateConversationOrder()
    } catch {
      /* 轮询失败不报错，下次轮询会继续尝试 */
    }
  }

  /**
   * 添加消息到会话（本地操作）
   */
  const addMessage = (conversationId: string, message: Message) => {
    updateConversation(conversationId, (conversation) => ({
      ...conversation,
      messages: [...conversation.messages, message],
      lastMessageAt: message.createdAt,
      lastMessagePreview: message.content.type === 'text' ? message.content.text ?? '' : ''
    }))
    updateConversationOrder()
  }

  /**
   * 切换会话置顶状态
   */
  const toggleConversationPin = async (conversationId: string) => {
    const conversation = conversations.value.find((item) => item.id === conversationId)
    if (!conversation) return
    const nextPinned = !conversation.pinned
    updateConversation(conversationId, (item) => ({ ...item, pinned: nextPinned }))
    updateConversationOrder()

    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId) return
    try {
      await client.put(
        `/workspaces/${workspaceId}/conversations/${conversationId}/settings`,
        { pinned: nextPinned }
      )
    } catch {
      /* optimistic UI; API error toasted by interceptor */
    }
  }

  /**
   * 切换会话静音状态
   */
  const toggleConversationMute = async (conversationId: string) => {
    const conversation = conversations.value.find((item) => item.id === conversationId)
    if (!conversation) return
    const nextMuted = !conversation.muted
    updateConversation(conversationId, (item) => ({ ...item, muted: nextMuted }))

    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId) return
    try {
      await client.put(
        `/workspaces/${workspaceId}/conversations/${conversationId}/settings`,
        { muted: nextMuted }
      )
    } catch {
      /* optimistic UI; API error toasted by interceptor */
    }
  }

  /**
   * 重命名会话
   */
  const renameConversation = async (conversationId: string, name: string) => {
    const trimmed = name.trim()
    updateConversation(conversationId, (conversation) => ({
      ...conversation,
      customName: trimmed ? trimmed : undefined
    }))
    updateConversationOrder()

    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId) return
    try {
      await client.put(
        `/workspaces/${workspaceId}/conversations/${conversationId}`,
        { name: trimmed }
      )
    } catch {
      /* optimistic UI; API error toasted by interceptor */
    }
  }

  /**
   * 清空会话消息
   */
  const clearConversationMessages = async (conversationId: string) => {
    updateConversation(conversationId, (conversation) => ({
      ...conversation,
      messages: [],
      lastMessageAt: undefined,
      lastMessagePreview: undefined
    }))
    updateConversationOrder()
    loadedMessages.add(conversationId)

    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId) return
    try {
      await client.delete(
        `/workspaces/${workspaceId}/conversations/${conversationId}/messages`
      )
    } catch {
      /* optimistic UI; API error toasted by interceptor */
    }
  }

  /**
   * 删除会话并清理本地缓存
   */
  const deleteConversation = async (conversationId: string) => {
    conversations.value = conversations.value.filter((c) => c.id !== conversationId)
    loadedMessages.delete(conversationId)
    loadingMessages.delete(conversationId)
    conversationPaging.delete(conversationId)
    if (activeConversationId.value === conversationId) {
      activeConversationId.value = null
    }

    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId) return
    try {
      await client.delete(
        `/workspaces/${workspaceId}/conversations/${conversationId}`
      )
    } catch {
      /* optimistic UI; API error toasted by interceptor */
    }
  }

  /**
   * 标记会话已读
   */
  const markConversationRead = async (conversationId: string) => {
    if (!conversationId) return
    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId) return

    const conv = conversations.value.find((c) => c.id === conversationId)
    const prevUnread = conv?.unreadCount ?? 0

    updateConversation(conversationId, (conversation) => ({
      ...conversation,
      unreadCount: 0
    }))
    totalUnreadCount.value = Math.max(0, totalUnreadCount.value - prevUnread)

    try {
      await client.post(
        `/workspaces/${workspaceId}/conversations/${conversationId}/read`,
        { userId: currentUserId.value }
      )
    } catch {
      /* optimistic UI; API error toasted by interceptor */
    }
  }

  /**
   * 全部会话标为已读（服务端 + 本地）
   */
  const markAllConversationsRead = async () => {
    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId) return
    try {
      await client.post(`/workspaces/${workspaceId}/conversations/read-all`, {
        userId: currentUserId.value
      })
      conversations.value = conversations.value.map((c) => ({ ...c, unreadCount: 0 }))
      totalUnreadCount.value = 0
    } catch {
      /* optimistic UI; API error toasted by interceptor */
    }
  }

  /**
   * 更新频道成员列表（全量替换）
   */
  const setConversationMembers = async (conversationId: string, memberIds: string[]) => {
    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId || !conversationId) return null
    try {
      const { data } = await client.put<ConversationDto>(
        `/workspaces/${workspaceId}/conversations/${conversationId}/members`,
        { memberIds }
      )
      const next = normalizeConversation(data)
      updateConversation(conversationId, (c) => ({
        ...next,
        messages: c.messages
      }))
      updateConversationOrder()
      return next
    } catch {
      return null
    }
  }

  /** Reference semantics: ensure a DM with a member exists and return conversation id */
  const ensureDirectMessage = async (memberId: string) => {
    const conv = await createDirectConversation(memberId)
    return conv?.id ?? null
  }

  /**
   * 删除与某成员相关的 DM，并将其从频道成员中移除（服务端）
   */
  const deleteMemberConversations = async (memberId: string): Promise<boolean> => {
    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId || !memberId) return false
    try {
      await client.delete(`/workspaces/${workspaceId}/members/${memberId}/conversations`)
      await loadSession()
      return true
    } catch {
      return false
    }
  }

  /**
   * 仅拉取该会话最新一条消息并合并到本地（替代 send 后多次全量刷新）
   */
  const ensureConversationLatestMessage = async (conversationId: string) => {
    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId || !conversationId) return
    try {
      const response = await client.get<MessageDto[]>(
        `/workspaces/${workspaceId}/conversations/${conversationId}/messages`,
        { params: { limit: 1 } }
      )
      const dtos = response.data
      if (!dtos?.length) return
      const latest = normalizeMessage(dtos[dtos.length - 1])
      updateConversation(conversationId, (conversation) => {
        const exists = conversation.messages.some((m) => m.id === latest.id)
        if (exists) return conversation
        return {
          ...conversation,
          messages: [...conversation.messages, latest],
          lastMessageAt: latest.createdAt,
          lastMessagePreview:
            latest.content.type === 'text' ? latest.content.text ?? '' : '',
          lastMessageSenderId: latest.senderId,
          lastMessageSenderName: latest.senderName,
          lastMessageSenderAvatar: latest.senderAvatar
        }
      })
      updateConversationOrder()
    } catch {
      /* non-fatal; API error toasted by interceptor */
    }
  }

  /** 成员信息变更后刷新消息上的显示名/头像 */
  const refreshMessageAuthors = () => {
    const memberMap = new Map((members.value ?? []).map((m) => [m.id, m]))
    conversations.value = conversations.value.map((conv) => ({
      ...conv,
      messages: conv.messages.map((msg) => {
        if (!msg.senderId) return msg
        const m = memberMap.get(msg.senderId)
        if (!m) return msg
        return {
          ...msg,
          senderName: m.name,
          senderAvatar: m.avatar ?? msg.senderAvatar
        }
      })
    }))
  }

  /**
   * 应用未读同步事件
   */
  const applyUnreadSync = (payload: ChatUnreadSyncPayload) => {
    if (!payload) return
    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId || payload.workspaceId !== workspaceId) return

    if (typeof payload.totalUnreadCount === 'number') {
      totalUnreadCount.value = payload.totalUnreadCount
    }
    if (payload.resetAll) {
      conversations.value = conversations.value.map((conversation) => ({
        ...conversation,
        unreadCount: 0
      }))
      return
    }
    const conversationId = payload.conversationId?.trim()
    if (!conversationId || typeof payload.conversationUnreadCount !== 'number') return

    updateConversation(conversationId, (conversation) => ({
      ...conversation,
      unreadCount: payload.conversationUnreadCount ?? 0
    }))
  }

  /**
   * 应用消息状态事件
   */
  const applyMessageStatus = (payload: ChatMessageStatusPayload) => {
    if (!payload) return
    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId || payload.workspaceId !== workspaceId) return
    const conversationId = payload.conversationId?.trim()
    if (!conversationId) return

    updateConversation(conversationId, (conversation) => ({
      ...conversation,
      messages: conversation.messages.map((message) =>
        message.id === payload.messageId ? { ...message, status: payload.status } : message
      )
    }))
  }

  /**
   * Merge streaming terminal output into the current conversation (same idea as reference applyTerminalStreamMessage).
   */
  const applyTerminalStreamMessage = (payload: TerminalChatStreamPayload) => {
    if (!payload || payload.mode === 'final') return
    if (!streamOutputEnabled.value) return
    const conversationId = payload.conversationId?.trim()
    const spanId = payload.spanId?.trim()
    if (!conversationId || !spanId) return
    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId || (payload.workspaceId && payload.workspaceId !== workspaceId)) return
    if (!loadedMessages.has(conversationId)) return
    const rawText = payload.content ?? ''
    const chunk = payload.mode === 'delta' ? rawText : rawText.trimEnd()
    if (!chunk) return
    const messageId = streamMessageIds.get(spanId) ?? buildStreamMessageId(spanId)
    streamMessageIds.set(spanId, messageId)
    updateConversation(conversationId, (conversation) => {
      const messages = [...conversation.messages]
      const index = messages.findIndex((item) => item.id === messageId)
      let nextText: string
      if (index >= 0) {
        const current = messages[index]
        const currentText = current.content.type === 'text' ? current.content.text ?? '' : ''
        const merged = payload.mode === 'delta' ? `${currentText}${chunk}` : chunk
        nextText = stripAnsiForChat(merged)
      } else {
        nextText = stripAnsiForChat(chunk)
      }
      if (!nextText) {
        if (index >= 0) {
          messages.splice(index, 1)
          return { ...conversation, messages }
        }
        return conversation
      }
      const dto: MessageDto = {
        id: messageId,
        senderId: payload.memberId,
        content: { type: 'text', text: nextText },
        createdAt: payload.timestamp || Date.now(),
        isAi: false,
        status: 'sending'
      }
      const message = normalizeMessage(dto)
      if (index >= 0) {
        messages[index] = {
          ...messages[index],
          content: { type: 'text', text: nextText }
        }
      } else {
        messages.push(message)
      }
      return {
        ...conversation,
        messages
      }
    })
  }

  /**
   * 终端行持久化后的最终消息（替换同 span 的流式气泡）。
   */
  const applyTerminalStreamFinal = (payload: TerminalChatStreamPayload) => {
    if (!payload || payload.mode !== 'final') return
    const conversationId = payload.conversationId?.trim()
    if (!conversationId || !payload.messageId) return
    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId || (payload.workspaceId && payload.workspaceId !== workspaceId)) return

    const spanId = payload.spanId?.trim()
    const streamMessageId = spanId ? streamMessageIds.get(spanId) : undefined
    if (spanId) {
      streamMessageIds.delete(spanId)
    }

    const conversation = conversations.value.find((item) => item.id === conversationId)
    if (!conversation) return

    const dto: MessageDto = {
      id: payload.messageId,
      senderId: payload.memberId,
      content: { type: 'text', text: stripAnsiForChat(payload.content ?? '') },
      createdAt: payload.timestamp,
      isAi: payload.isAi ?? false,
      status: 'sent'
    }
    const message = normalizeMessage(dto)

    if (streamMessageId) {
      updateConversation(conversationId, (current) => {
        const messages = [...current.messages]
        const index = messages.findIndex((item) => item.id === streamMessageId)
        if (index >= 0) {
          messages[index] = message
        } else {
          messages.push(message)
        }
        return {
          ...current,
          messages,
          lastMessageAt: message.createdAt,
          lastMessagePreview: resolvePreviewText(message.content),
          lastMessageSenderId: message.senderId,
          lastMessageSenderName: message.senderName,
          lastMessageSenderAvatar: message.senderAvatar,
          unreadCount: current.unreadCount
        }
      })
    } else {
      updateConversation(conversationId, (current) => ({
        ...current,
        messages: [...current.messages, message],
        lastMessageAt: message.createdAt,
        lastMessagePreview: resolvePreviewText(message.content),
        lastMessageSenderId: message.senderId,
        lastMessageSenderName: message.senderName,
        lastMessageSenderAvatar: message.senderAvatar,
        unreadCount: current.unreadCount
      }))
    }
    updateConversationOrder()
  }

  const applyTerminalChatStreamEvent = (_sessionId: string, payload: TerminalChatStreamPayload) => {
    if (!payload) return
    if (payload.mode === 'final') {
      applyTerminalStreamFinal(payload)
    } else {
      applyTerminalStreamMessage(payload)
    }
  }

  const appendTerminalMessage = (payload: ChatMessageCreatedPayload) => {
    if (!payload || !payload.conversationId) return
    const workspaceId = currentWorkspace.value?.id
    if (!workspaceId) return
    if (payload.workspaceId && payload.workspaceId !== workspaceId) return

    const conversation = conversations.value.find((item) => item.id === payload.conversationId)
    if (!conversation) return

    const message = normalizeMessage(payload.message)
    updateConversation(payload.conversationId, (current) => ({
      ...current,
      messages: [...current.messages, message],
      lastMessageAt: message.createdAt,
      lastMessagePreview: message.content.type === 'text' ? message.content.text ?? '' : '',
      lastMessageSenderId: message.senderId,
      lastMessageSenderName: message.senderName,
      lastMessageSenderAvatar: message.senderAvatar
    }))
    updateConversationOrder()
  }

  /**
   * 重置聊天状态
   */
  const reset = () => {
    stopPolling()
    disconnectWebSocket()
    isReady.value = false
    chatError.value = null
    defaultChannelId.value = null
    totalUnreadCount.value = 0
    conversations.value = []
    activeConversationId.value = null
    loadedMessages.clear()
    loadingMessages.clear()
    conversationPaging.clear()
    streamMessageIds.clear()
    loadSequence += 1
    loading.value = false
  }

  // WebSocket connection management
  let chatSocketUnsubscribe: (() => void) | null = null
  let tabSyncUnsubscribe: (() => void) | null = null
  let messageQueueUnsubscribe: (() => void) | null = null

  const connectWebSocket = async (workspaceId: string) => {
    // Initialize tab sync
    initTabSync()
    registerWorkspaceActive(workspaceId)

    // Only leader tab connects to WebSocket
    if (!isLeaderTab()) {
      console.log('[chatStore] Not leader tab, skipping WebSocket connection')
      return
    }

    try {
      const socket = getChatSocket()
      if (socket.isConnected && socket.currentWorkspaceId === workspaceId) {
        return // Already connected
      }

      await socket.connect(workspaceId)

      // Subscribe to WebSocket messages
      chatSocketUnsubscribe = socket.onMessage((message) => {
        handleWebSocketMessage(message)
        // Broadcast to other tabs
        broadcastMessage(message)
      })
    } catch (error) {
      console.error('[chatStore] WebSocket connection failed:', error)
    }
  }

  const disconnectWebSocket = () => {
    if (chatSocketUnsubscribe) {
      chatSocketUnsubscribe()
      chatSocketUnsubscribe = null
    }
    if (messageQueueUnsubscribe) {
      messageQueueUnsubscribe()
      messageQueueUnsubscribe = null
    }
    if (tabSyncUnsubscribe) {
      tabSyncUnsubscribe()
      tabSyncUnsubscribe = null
    }
    const workspaceId = currentWorkspace.value?.id
    if (workspaceId) {
      unregisterWorkspaceActive(workspaceId)
    }
    closeChatSocket()
    closeTabSync()
  }

  const handleWebSocketMessage = (message: ChatWebSocketMessage) => {
    if (message.type === 'new_message' && message.conversationId && message.messageId) {
      const conversationId = message.conversationId
      const msg: MessageDto = {
        id: message.messageId,
        senderId: message.senderId || '',
        content: { type: 'text', text: message.content || '' },
        createdAt: message.createdAt || Date.now(),
        isAi: message.isAi || false,
        status: 'sent'
      }
      const normalizedMsg = normalizeMessage(msg)

      // Check if message already exists
      const conversation = conversations.value.find((c) => c.id === conversationId)
      if (conversation?.messages.some((m) => m.id === msg.id)) {
        return
      }

      // Add message to conversation
      updateConversation(conversationId, (conv) => ({
        ...conv,
        messages: [...conv.messages, normalizedMsg],
        lastMessageAt: normalizedMsg.createdAt,
        lastMessagePreview: normalizedMsg.content.type === 'text' ? normalizedMsg.content.text ?? '' : '',
        lastMessageSenderId: normalizedMsg.senderId,
        lastMessageSenderName: normalizedMsg.senderName,
        lastMessageSenderAvatar: normalizedMsg.senderAvatar
      }))
      updateConversationOrder()
    }
  }

  // Subscribe to cross-tab messages
  const setupTabSync = () => {
    tabSyncUnsubscribe = onTabSyncMessage((msg) => {
      if (msg.type === 'message_received' && msg.payload) {
        handleWebSocketMessage(msg.payload as ChatWebSocketMessage)
      }
    })

    // Subscribe to queued messages
    messageQueueUnsubscribe = onQueuedMessage((message) => {
      handleWebSocketMessage(message)
    })
  }

  // 监听 workspace 变化时自动加载
  watch(
    () => currentWorkspace.value?.id,
    (workspaceId, prevWorkspaceId) => {
      if (!workspaceId || workspaceId !== prevWorkspaceId) {
        reset()
      }
      if (workspaceId) {
        void loadSession()
        setupTabSync()
        void connectWebSocket(workspaceId)
      }
    },
    { immediate: true }
  )

  watch(
    () =>
      (members.value ?? [])
        .map((m) => m.id)
        .sort()
        .join('|'),
    (next, prev) => {
      if (!isReady.value || !next || next === prev) return
      refreshMessageAuthors()
      void loadSession()
    }
  )

  return {
    conversations,
    activeConversation,
    activeConversationId,
    currentUserId,
    isReady,
    chatError,
    defaultChannelId,
    totalUnreadCount,
    loading,
    sortedConversations,
    maxMessageLength: MAX_MESSAGE_LENGTH,
    loadSession,
    loadConversationMessages,
    loadOlderMessages,
    sendMessage,
    createDirectConversation,
    createConversation,
    setActiveConversation,
    addMessage,
    toggleConversationPin,
    toggleConversationMute,
    renameConversation,
    clearConversationMessages,
    deleteConversation,
    markConversationRead,
    markAllConversationsRead,
    setConversationMembers,
    ensureDirectMessage,
    deleteMemberConversations,
    ensureConversationLatestMessage,
    refreshMessageAuthors,
    applyUnreadSync,
    applyMessageStatus,
    applyTerminalStreamMessage,
    applyTerminalStreamFinal,
    applyTerminalChatStreamEvent,
    appendTerminalMessage,
    getConversationTitle: resolveConversationTitle,
    getConversationPaging: (conversationId: string) => getPagingState(conversationId),
    startPolling,
    stopPolling,
    pollNewMessages,
    reset
  }
})

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useChatStore, import.meta.hot))
}