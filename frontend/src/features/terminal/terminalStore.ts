// 终端标签页状态：管理 session 与 tab 的对应关系及活动提示
import { ref, computed } from 'vue'
import { acceptHMRUpdate, defineStore } from 'pinia'
import { TerminalSocket } from '@/shared/socket/terminal'
import { terminalApi } from '@/shared/api/terminal'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'
import type {
  TerminalTab,
  TerminalLayoutMode,
  TerminalPaneId,
  TerminalServerMessage,
  TerminalChatStreamPayload,
  TerminalConnectionStatus
} from '@/shared/types/terminal'
import { notifyTerminalChatStream } from '@/features/terminal/terminalChatBridge'
import { notifyUserError } from '@/shared/notifyError'

const LAYOUT_PANES: Record<TerminalLayoutMode, TerminalPaneId[]> = {
  single: ['primary'],
  'split-vertical': ['left', 'right'],
  'split-horizontal': ['top', 'bottom'],
  'grid-2x2': ['top-left', 'top-right', 'bottom-left', 'bottom-right']
}

type PaneAssignments = Partial<Record<TerminalPaneId, string | null>>
type SocketMap = Map<string, TerminalSocket>

/**
 * 终端标签页存储
 * 输入：创建/打开/关闭/排序操作
 * 输出：tabs 列表与活动 tab id
 */
export const useTerminalStore = defineStore('terminal', () => {
  const tabs = ref<TerminalTab[]>([])
  const activeTabId = ref<string | null>(null)
  const layoutMode = ref<TerminalLayoutMode>('single')
  const paneAssignments = ref<PaneAssignments>({ primary: null })
  const focusedPaneId = ref<TerminalPaneId>('primary')
  const paneIds = ref<TerminalPaneId[]>(LAYOUT_PANES.single)

  // WebSocket 连接管理
  const sockets: SocketMap = new Map()
  const connectionStatus = ref<Record<string, TerminalConnectionStatus>>({})

  // 本地自增序号用于默认标题
  let tabCounter = 1

  const resolvePaneIds = (mode: TerminalLayoutMode) => LAYOUT_PANES[mode] ?? LAYOUT_PANES.single

  const updatePaneIds = (mode: TerminalLayoutMode) => {
    paneIds.value = resolvePaneIds(mode)
  }

  const isTabAlive = (terminalId: string) => tabs.value.some((tab) => tab.id === terminalId)

  const resolveFallbackPaneId = (paneId?: TerminalPaneId | null) => {
    if (paneId && paneIds.value.includes(paneId)) {
      return paneId
    }
    return paneIds.value[0] ?? 'primary'
  }

  const findPaneByTerminalId = (terminalId: string) => {
    for (const paneId of paneIds.value) {
      if (paneAssignments.value[paneId] === terminalId) {
        return paneId
      }
    }
    return null
  }

  const sortTabsByPin = (list: TerminalTab[]) => {
    const pinned: TerminalTab[] = []
    const unpinned: TerminalTab[] = []
    for (const tab of list) {
      if (tab.pinned) {
        pinned.push(tab)
      } else {
        unpinned.push(tab)
      }
    }
    return pinned.concat(unpinned)
  }

  const normalizePinnedTabs = () => {
    const next = sortTabsByPin(tabs.value)
    const current = tabs.value
    if (next.length !== current.length) {
      tabs.value = next
      return
    }
    for (let index = 0; index < next.length; index += 1) {
      if (next[index] !== current[index]) {
        tabs.value = next
        return
      }
    }
  }

  const syncPaneAssignments = () => {
    const nextAssignments: PaneAssignments = {}
    for (const paneId of paneIds.value) {
      const existing = paneAssignments.value[paneId]
      if (existing && isTabAlive(existing)) {
        nextAssignments[paneId] = existing
      } else {
        nextAssignments[paneId] = null
      }
    }
    paneAssignments.value = nextAssignments
    if (!paneIds.value.includes(focusedPaneId.value)) {
      focusedPaneId.value = resolveFallbackPaneId(null)
    }
    const focusedTab = paneAssignments.value[focusedPaneId.value]
    if (focusedTab) {
      activeTabId.value = focusedTab
      return
    }
    const firstAssigned = paneIds.value.map((paneId) => paneAssignments.value[paneId]).find((id) => id)
    activeTabId.value = firstAssigned ?? null
  }

  const buildLayoutCandidates = (preferredId?: string) => {
    const candidates: string[] = []
    const pushCandidate = (id?: string | null) => {
      if (!id || !isTabAlive(id) || candidates.includes(id)) {
        return
      }
      candidates.push(id)
    }
    pushCandidate(preferredId)
    for (const terminalId of Object.values(paneAssignments.value)) {
      pushCandidate(terminalId)
    }
    for (const tab of tabs.value) {
      pushCandidate(tab.id)
    }
    return candidates
  }

  const setLayoutMode = (mode: TerminalLayoutMode, options?: { preferTerminalId?: string }) => {
    if (layoutMode.value === mode) {
      return
    }
    updatePaneIds(mode)
    layoutMode.value = mode
    const candidates = buildLayoutCandidates(options?.preferTerminalId)
    const nextAssignments: PaneAssignments = {}
    for (const paneId of paneIds.value) {
      nextAssignments[paneId] = candidates.shift() ?? null
    }
    paneAssignments.value = nextAssignments
    const preferred = options?.preferTerminalId
    const preferredPane = preferred
      ? paneIds.value.find((paneId) => paneAssignments.value[paneId] === preferred)
      : null
    focusedPaneId.value = preferredPane ?? resolveFallbackPaneId(focusedPaneId.value)
    syncPaneAssignments()
  }

  const unassignTab = (terminalId: string) => {
    let changed = false
    const nextAssignments: PaneAssignments = { ...paneAssignments.value }
    for (const paneId of paneIds.value) {
      if (nextAssignments[paneId] === terminalId) {
        nextAssignments[paneId] = null
        changed = true
      }
    }
    if (!changed) {
      return
    }
    paneAssignments.value = nextAssignments
    syncPaneAssignments()
  }

  const assignTabToPane = (
    terminalId: string,
    paneId: TerminalPaneId,
    options?: { focus?: boolean; activate?: boolean }
  ) => {
    if (!isTabAlive(terminalId)) {
      return
    }
    const resolvedPane = resolveFallbackPaneId(paneId)
    const nextAssignments: PaneAssignments = { ...paneAssignments.value }
    for (const existingPane of paneIds.value) {
      if (nextAssignments[existingPane] === terminalId) {
        nextAssignments[existingPane] = null
      }
    }
    nextAssignments[resolvedPane] = terminalId
    paneAssignments.value = nextAssignments
    if (options?.activate) {
      activeTabId.value = terminalId
    }
    if (options?.focus) {
      focusedPaneId.value = resolvedPane
      activeTabId.value = terminalId
    }
  }

  const assignTabOnOpen = (terminalId: string, shouldActivate: boolean) => {
    const emptyPane = paneIds.value.find((paneId) => !paneAssignments.value[paneId])
    if (shouldActivate) {
      const targetPane = emptyPane ?? resolveFallbackPaneId(focusedPaneId.value)
      assignTabToPane(terminalId, targetPane, { focus: true, activate: true })
      return
    }
    if (emptyPane) {
      assignTabToPane(terminalId, emptyPane, { activate: false })
    }
  }

  /**
   * 创建 WebSocket 连接
   */
  const createConnection = async (sessionId: string): Promise<TerminalSocket> => {
    const existingSocket = sockets.get(sessionId)
    if (existingSocket) {
      return existingSocket
    }

    connectionStatus.value[sessionId] = 'connecting'
    const socket = new TerminalSocket()

    try {
      await socket.connect(sessionId)
      socket.onMessage((msg: TerminalServerMessage) => {
        if (msg.type === 'terminal_chat_stream') {
          const payload = (msg as { payload?: TerminalChatStreamPayload }).payload
          if (payload) {
            notifyTerminalChatStream(sessionId, payload)
          }
        }
      })
      connectionStatus.value[sessionId] = 'connected'
      sockets.set(sessionId, socket)
      return socket
    } catch (error) {
      connectionStatus.value[sessionId] = 'error'
      notifyUserError('Terminal WebSocket', error)
      throw error
    }
  }

  /**
   * 关闭 WebSocket 连接
   */
  const closeConnection = (sessionId: string) => {
    const socket = sockets.get(sessionId)
    if (socket) {
      socket.close()
      sockets.delete(sessionId)
      connectionStatus.value[sessionId] = 'disconnected'
    }
  }

  /**
   * 获取 WebSocket 连接
   */
  const getConnection = (sessionId: string): TerminalSocket | undefined => {
    return sockets.get(sessionId)
  }

  /**
   * 注册消息处理器
   */
  const registerMessageHandler = (
    sessionId: string,
    handler: (message: TerminalServerMessage) => void
  ): (() => void) | null => {
    const socket = sockets.get(sessionId)
    if (!socket) {
      return null
    }
    return socket.onMessage(handler)
  }

  /**
   * 发送终端输入
   */
  const sendInput = (sessionId: string, data: string) => {
    const socket = sockets.get(sessionId)
    if (socket) {
      socket.input(data)
    }
  }

  /**
   * 发送终端 resize
   */
  const sendResize = (sessionId: string, cols: number, rows: number) => {
    const socket = sockets.get(sessionId)
    if (socket) {
      socket.resize(cols, rows)
    }
  }

  /**
   * 创建新的终端会话并加入标签列表
   */
  const createTab = async (options?: { cwd?: string; workspaceId?: string; title?: string }): Promise<string> => {
    tabCounter += 1
    const title = options?.title ?? `Terminal ${tabCounter}`

    // 先调用后端 API 创建 session
    let sessionId: string
    try {
      const workspaceStore = useWorkspaceStore()
      const response = await terminalApi.createSession({
        workspaceId: options?.workspaceId ?? workspaceStore.currentWorkspace?.id,
        terminalType: 'native'
      })
      sessionId = response.sessionId
    } catch {
      /* createSession: axios interceptor surfaces API errors */
      return ''
    }

    tabs.value.push({
      id: sessionId,
      title,
      sessionId,
      hasActivity: false,
      isBlinking: false,
      pinned: false,
      keepAlive: false
    })
    normalizePinnedTabs()
    assignTabOnOpen(sessionId, true)

    // 创建 WebSocket 连接
    try {
      await createConnection(sessionId)
    } catch {
      /* WebSocket failure already reported in createConnection */
    }

    return sessionId
  }

  const openTab = (
    terminalId: string,
    options: {
      title: string
      memberId?: string
      terminalType?: 'native' | 'web' | 'claude' | 'custom'
      keepAlive?: boolean
      activate?: boolean
    }
  ) => {
    const shouldActivate = options.activate !== false
    const existing = tabs.value.find((item) => item.id === terminalId)
    if (existing) {
      existing.title = options.title
      existing.memberId = options.memberId
      existing.terminalType = options.terminalType
      existing.keepAlive = options.keepAlive ?? existing.keepAlive
      existing.isBlinking = false
      if (shouldActivate) {
        activeTabId.value = terminalId
        existing.hasActivity = false
      }
      if (!findPaneByTerminalId(terminalId)) {
        assignTabOnOpen(terminalId, shouldActivate)
      }
      return
    }
    if (connectionStatus.value[terminalId] === undefined) {
      connectionStatus.value[terminalId] = 'pending'
    }
    tabs.value.push({
      id: terminalId,
      title: options.title,
      hasActivity: false,
      isBlinking: false,
      pinned: false,
      memberId: options.memberId,
      terminalType: options.terminalType,
      keepAlive: options.keepAlive ?? false
    })
    normalizePinnedTabs()
    assignTabOnOpen(terminalId, shouldActivate)
  }

  const setActiveTab = (terminalId: string) => {
    activeTabId.value = terminalId
    const tab = tabs.value.find((item) => item.id === terminalId)
    if (tab) {
      tab.hasActivity = false
      tab.isBlinking = false
    }
    const paneId = findPaneByTerminalId(terminalId)
    if (paneId) {
      focusedPaneId.value = paneId
    }
  }

  /**
   * 关闭标签页，必要时终止 WebSocket 连接
   */
  const closeTab = async (terminalId: string) => {
    const index = tabs.value.findIndex((item) => item.id === terminalId)
    if (index === -1) {
      return
    }
    const tab = tabs.value[index]
    if (!tab.keepAlive) {
      closeConnection(terminalId)
      // Also delete the session from backend
      terminalApi.deleteSession(terminalId).catch(() => {
        /* DELETE session: axios interceptor reports failure */
      })
    }
    tabs.value.splice(index, 1)
    unassignTab(terminalId)

    // 如果关闭的是活动标签，选择下一个
    if (activeTabId.value === terminalId && tabs.value.length > 0) {
      const newIndex = Math.min(index, tabs.value.length - 1)
      setActiveTab(tabs.value[newIndex].id)
    } else if (tabs.value.length === 0) {
      activeTabId.value = null
    }
  }

  const moveTab = (fromId: string, toId: string) => {
    const fromIndex = tabs.value.findIndex((item) => item.id === fromId)
    const toIndex = tabs.value.findIndex((item) => item.id === toId)
    if (fromIndex === -1 || toIndex === -1 || fromIndex === toIndex) {
      return
    }
    const [item] = tabs.value.splice(fromIndex, 1)
    tabs.value.splice(toIndex, 0, item)
    normalizePinnedTabs()
  }

  const moveTabToIndex = (fromId: string, insertIndex: number) => {
    const fromIndex = tabs.value.findIndex((item) => item.id === fromId)
    if (fromIndex === -1) {
      return
    }
    const [item] = tabs.value.splice(fromIndex, 1)
    const safeIndex = Math.max(0, Math.min(tabs.value.length, insertIndex))
    tabs.value.splice(safeIndex, 0, item)
    normalizePinnedTabs()
  }

  const swapTabOrder = (fromId: string, toId: string) => {
    if (fromId === toId) {
      return
    }
    const fromIndex = tabs.value.findIndex((item) => item.id === fromId)
    const toIndex = tabs.value.findIndex((item) => item.id === toId)
    if (fromIndex === -1 || toIndex === -1 || fromIndex === toIndex) {
      return
    }
    const fromTab = tabs.value[fromIndex]
    const toTab = tabs.value[toIndex]
    if (fromTab.pinned !== toTab.pinned) {
      return
    }
    const next = tabs.value.slice()
    next[fromIndex] = toTab
    next[toIndex] = fromTab
    tabs.value = next
    normalizePinnedTabs()
  }

  const setTabOrder = (orderedIds: string[]) => {
    if (orderedIds.length === 0) {
      return
    }
    const remaining = new Map(tabs.value.map((tab) => [tab.id, tab]))
    const ordered: TerminalTab[] = []
    for (const id of orderedIds) {
      const tab = remaining.get(id)
      if (tab) {
        ordered.push(tab)
        remaining.delete(id)
      }
    }
    for (const tab of tabs.value) {
      if (remaining.has(tab.id)) {
        ordered.push(tab)
      }
    }
    tabs.value = sortTabsByPin(ordered)
  }

  const markActivity = (terminalId: string) => {
    const tab = tabs.value.find((item) => item.id === terminalId)
    if (tab && activeTabId.value !== terminalId) {
      tab.hasActivity = true
      tab.isBlinking = false
    }
  }

  const clearActivity = (terminalId: string) => {
    const tab = tabs.value.find((item) => item.id === terminalId)
    if (tab) {
      tab.hasActivity = false
      tab.isBlinking = false
    }
  }

  const togglePin = (terminalId: string) => {
    const tab = tabs.value.find((item) => item.id === terminalId)
    if (tab) {
      tab.pinned = !tab.pinned
      normalizePinnedTabs()
    }
  }

  const renameTab = (terminalId: string, title: string) => {
    const tab = tabs.value.find((item) => item.id === terminalId)
    if (tab) {
      tab.title = title
    }
  }

  /**
   * 获取活动标签的 WebSocket 连接
   */
  const activeConnection = computed(() => {
    if (!activeTabId.value) {
      return null
    }
    return sockets.get(activeTabId.value)
  })

  /**
   * 获取活动标签
   */
  const activeTab = computed(() => {
    if (!activeTabId.value) {
      return null
    }
    return tabs.value.find((tab) => tab.id === activeTabId.value)
  })

  /**
   * 重置终端状态
   */
  const reset = () => {
    // 关闭所有连接
    for (const sessionId of sockets.keys()) {
      closeConnection(sessionId)
    }
    tabs.value = []
    activeTabId.value = null
    layoutMode.value = 'single'
    paneAssignments.value = { primary: null }
    focusedPaneId.value = 'primary'
    paneIds.value = LAYOUT_PANES.single
    connectionStatus.value = {}
    tabCounter = 1
  }

  return {
    tabs,
    activeTabId,
    activeTab,
    activeConnection,
    layoutMode,
    paneAssignments,
    focusedPaneId,
    paneIds,
    connectionStatus,
    createConnection,
    closeConnection,
    getConnection,
    registerMessageHandler,
    sendInput,
    sendResize,
    createTab,
    openTab,
    setActiveTab,
    closeTab,
    moveTab,
    moveTabToIndex,
    swapTabOrder,
    setTabOrder,
    markActivity,
    clearActivity,
    togglePin,
    renameTab,
    setLayoutMode,
    assignTabToPane,
    unassignTab,
    syncPaneAssignments,
    reset
  }
})

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useTerminalStore, import.meta.hot))
}