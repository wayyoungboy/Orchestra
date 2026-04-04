# Orchestra Phase 4 - 聊天和终端 UI 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** 实现聊天界面和终端 UI，复刻 对照端 的核心交互体验。

**Architecture:** Vue 3 组件 + xterm.js + WebSocket

**Tech Stack:** TypeScript, Tailwind CSS, xterm.js

---

## 文件结构

```
frontend/src/features/
├── chat/
│   ├── ChatInterface.vue
│   ├── chatStore.ts
│   ├── components/
│   │   ├── ChatHeader.vue
│   │   ├── ChatInput.vue
│   │   ├── ChatSidebar.vue
│   │   ├── MessagesList.vue
│   │   └── MembersSidebar.vue
│   └── types.ts
├── terminal/
│   ├── TerminalWorkspace.vue
│   ├── TerminalPane.vue
│   ├── terminalStore.ts
│   └── components/
│       └── TerminalTab.vue
└── settings/
    ├── Settings.vue
    └── settingsStore.ts
```

---

### Task 1: 聊天模块

**Files:**
- Create: `frontend/src/features/chat/types.ts`
- Create: `frontend/src/features/chat/chatStore.ts`
- Create: `frontend/src/features/chat/ChatInterface.vue`

- [ ] **Step 1: 创建聊天类型**

```typescript
// frontend/src/features/chat/types.ts
export type ConversationType = 'dm' | 'group' | 'channel'

export interface MessageContent {
  type: 'text' | 'code' | 'file'
  text?: string
  key?: string
}

export interface Message {
  id: string
  conversationId: string
  senderId: string
  user: string
  avatar: string
  content: MessageContent
  isAi: boolean
  status: 'sending' | 'sent' | 'delivered' | 'read'
  createdAt: string
}

export interface Conversation {
  id: string
  workspaceId: string
  type: ConversationType
  name: string
  targetId?: string
  memberIds: string[]
  pinned: boolean
  muted: boolean
  unreadCount: number
  messages: Message[]
  lastMessageAt?: string
  lastMessagePreview?: string
}
```

- [ ] **Step 2: 创建聊天 Store**

```typescript
// frontend/src/features/chat/chatStore.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Conversation, Message } from './types'

export const useChatStore = defineStore('chat', () => {
  const conversations = ref<Conversation[]>([])
  const activeConversationId = ref<string | null>(null)
  const loading = ref(false)

  const activeConversation = computed(() =>
    conversations.value.find((c) => c.id === activeConversationId.value)
  )

  const sortedConversations = computed(() =>
    [...conversations.value].sort((a, b) => {
      if (a.pinned !== b.pinned) return a.pinned ? -1 : 1
      const timeA = a.lastMessageAt ? new Date(a.lastMessageAt).getTime() : 0
      const timeB = b.lastMessageAt ? new Date(b.lastMessageAt).getTime() : 0
      return timeB - timeA
    })
  )

  function setActiveConversation(id: string) {
    activeConversationId.value = id
    // 标记已读
    const conv = conversations.value.find((c) => c.id === id)
    if (conv) {
      conv.unreadCount = 0
    }
  }

  function addMessage(conversationId: string, message: Message) {
    const conv = conversations.value.find((c) => c.id === conversationId)
    if (conv) {
      conv.messages.push(message)
      conv.lastMessageAt = message.createdAt
      conv.lastMessagePreview = message.content.text?.slice(0, 100) || ''
    }
  }

  function createConversation(data: Partial<Conversation>) {
    const conv: Conversation = {
      id: data.id || crypto.randomUUID(),
      workspaceId: data.workspaceId || '',
      type: data.type || 'group',
      name: data.name || '',
      memberIds: data.memberIds || [],
      pinned: false,
      muted: false,
      unreadCount: 0,
      messages: [],
    }
    conversations.value.push(conv)
    return conv
  }

  return {
    conversations,
    activeConversationId,
    activeConversation,
    sortedConversations,
    loading,
    setActiveConversation,
    addMessage,
    createConversation,
  }
})
```

- [ ] **Step 3: 创建聊天主界面**

```vue
<!-- frontend/src/features/chat/ChatInterface.vue -->
<template>
  <div class="flex h-full">
    <!-- 侧边栏 -->
    <ChatSidebar
      :conversations="chatStore.sortedConversations"
      :active-id="chatStore.activeConversationId"
      @select="chatStore.setActiveConversation"
      @create="handleCreateConversation"
    />

    <!-- 主内容区 -->
    <div class="flex-1 flex flex-col">
      <ChatHeader
        v-if="chatStore.activeConversation"
        :conversation="chatStore.activeConversation"
      />

      <MessagesList
        v-if="chatStore.activeConversation"
        :messages="chatStore.activeConversation.messages"
      />

      <div v-else class="flex-1 flex items-center justify-center text-white/50">
        <p>选择一个会话开始聊天</p>
      </div>

      <ChatInput
        v-if="chatStore.activeConversation"
        @send="handleSendMessage"
      />
    </div>

    <!-- 成员侧边栏 -->
    <MembersSidebar
      v-if="chatStore.activeConversation"
      :member-ids="chatStore.activeConversation.memberIds"
    />
  </div>
</template>

<script setup lang="ts">
import { useChatStore } from './chatStore'
import { useProjectStore } from '@/features/workspace/projectStore'
import ChatSidebar from './components/ChatSidebar.vue'
import ChatHeader from './components/ChatHeader.vue'
import MessagesList from './components/MessagesList.vue'
import ChatInput from './components/ChatInput.vue'
import MembersSidebar from './components/MembersSidebar.vue'

const chatStore = useChatStore()
const projectStore = useProjectStore()

function handleCreateConversation() {
  // 创建新会话
}

function handleSendMessage(text: string) {
  if (!chatStore.activeConversation) return

  const message = {
    id: crypto.randomUUID(),
    conversationId: chatStore.activeConversation.id,
    senderId: 'current-user',
    user: 'Owner',
    avatar: '',
    content: { type: 'text' as const, text },
    isAi: false,
    status: 'sending' as const,
    createdAt: new Date().toISOString(),
  }

  chatStore.addMessage(chatStore.activeConversation.id, message)

  // TODO: 通过 WebSocket 发送消息到后端
}
</script>
```

- [ ] **Step 4: 创建聊天子组件**

```vue
<!-- frontend/src/features/chat/components/ChatSidebar.vue -->
<template>
  <div class="w-64 bg-panel/50 border-r border-white/5 flex flex-col">
    <div class="p-4 border-b border-white/5">
      <button
        @click="$emit('create')"
        class="w-full py-2 px-4 bg-primary/20 text-primary rounded-xl text-sm font-medium hover:bg-primary/30 transition-colors"
      >
        + 新建会话
      </button>
    </div>

    <div class="flex-1 overflow-y-auto">
      <button
        v-for="conv in conversations"
        :key="conv.id"
        @click="$emit('select', conv.id)"
        :class="[
          'w-full p-3 text-left hover:bg-white/5 transition-colors',
          conv.id === activeId ? 'bg-white/10' : ''
        ]"
      >
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 rounded-full bg-surface flex items-center justify-center text-sm">
            {{ conv.name.charAt(0).toUpperCase() }}
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-white text-sm font-medium truncate">{{ conv.name }}</p>
            <p v-if="conv.lastMessagePreview" class="text-white/40 text-xs truncate">
              {{ conv.lastMessagePreview }}
            </p>
          </div>
          <span
            v-if="conv.unreadCount > 0"
            class="min-w-[18px] h-[18px] px-1 rounded-full bg-primary text-white text-[10px] flex items-center justify-center"
          >
            {{ conv.unreadCount > 99 ? '99+' : conv.unreadCount }}
          </span>
        </div>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Conversation } from '../types'

defineProps<{
  conversations: Conversation[]
  activeId: string | null
}>()

defineEmits<{
  (e: 'select', id: string): void
  (e: 'create'): void
}>()
</script>
```

```vue
<!-- frontend/src/features/chat/components/MessagesList.vue -->
<template>
  <div class="flex-1 overflow-y-auto p-4 space-y-4">
    <div
      v-for="message in messages"
      :key="message.id"
      class="flex gap-3"
    >
      <div
        class="w-8 h-8 rounded-full bg-surface flex items-center justify-center text-sm flex-shrink-0"
        :style="{ backgroundColor: message.avatar }"
      >
        {{ message.user.charAt(0).toUpperCase() }}
      </div>
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 mb-1">
          <span class="text-white font-medium text-sm">{{ message.user }}</span>
          <span class="text-white/30 text-xs">{{ formatTime(message.createdAt) }}</span>
        </div>
        <p class="text-white/80 text-sm whitespace-pre-wrap">{{ message.content.text }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Message } from '../types'

defineProps<{
  messages: Message[]
}>()

function formatTime(date: string): string {
  return new Date(date).toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>
```

```vue
<!-- frontend/src/features/chat/components/ChatInput.vue -->
<template>
  <div class="p-4 border-t border-white/5">
    <div class="flex gap-2">
      <input
        v-model="text"
        type="text"
        class="flex-1 bg-surface text-white rounded-xl px-4 py-3 border border-white/10 focus:border-primary/50 focus:outline-none"
        placeholder="输入消息..."
        @keyup.enter="send"
      />
      <button
        @click="send"
        :disabled="!text.trim()"
        class="px-6 py-3 bg-primary text-white rounded-xl font-medium disabled:opacity-50 disabled:cursor-not-allowed hover:bg-primary-hover transition-colors"
      >
        发送
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const emit = defineEmits<{
  (e: 'send', text: string): void
}>()

const text = ref('')

function send() {
  if (text.value.trim()) {
    emit('send', text.value.trim())
    text.value = ''
  }
}
</script>
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/features/chat/
git commit -m "feat: add chat module with sidebar, messages list, and input"
```

---

### Task 2: 终端模块

**Files:**
- Create: `frontend/src/features/terminal/terminalStore.ts`
- Create: `frontend/src/features/terminal/TerminalPane.vue`
- Create: `frontend/src/features/terminal/TerminalWorkspace.vue`

- [ ] **Step 1: 创建终端 Store**

```typescript
// frontend/src/features/terminal/terminalStore.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { TerminalSocket } from '@/shared/socket/terminal'

export type TerminalTab = {
  id: string
  title: string
  memberId?: string
  terminalType?: string
  hasActivity: boolean
  isBlinking: boolean
  active: boolean
}

export const useTerminalStore = defineStore('terminal', () => {
  const tabs = ref<TerminalTab[]>([])
  const activeTabId = ref<string | null>(null)
  const sockets = new Map<string, TerminalSocket>()
  const tabCounter = ref(1)

  const activeTab = computed(() =>
    tabs.value.find((t) => t.id === activeTabId.value)
  )

  function createTab(memberId?: string, terminalType?: string): string {
    const id = crypto.randomUUID()
    const title = `终端 ${tabCounter.value}`
    tabCounter.value++

    tabs.value.push({
      id,
      title,
      memberId,
      terminalType,
      hasActivity: false,
      isBlinking: false,
      active: false,
    })

    setActiveTab(id)
    return id
  }

  function setActiveTab(id: string) {
    tabs.value.forEach((t) => (t.active = t.id === id))
    activeTabId.value = id

    const tab = tabs.value.find((t) => t.id === id)
    if (tab) {
      tab.hasActivity = false
      tab.isBlinking = false
    }
  }

  function closeTab(id: string) {
    const socket = sockets.get(id)
    if (socket) {
      socket.close()
      sockets.delete(id)
    }

    tabs.value = tabs.value.filter((t) => t.id !== id)

    if (activeTabId.value === id) {
      activeTabId.value = tabs.value[0]?.id || null
    }
  }

  function markActivity(id: string) {
    const tab = tabs.value.find((t) => t.id === id)
    if (tab && activeTabId.value !== id) {
      tab.hasActivity = true
    }
  }

  function getSocket(id: string): TerminalSocket | undefined {
    return sockets.get(id)
  }

  function setSocket(id: string, socket: TerminalSocket) {
    sockets.set(id, socket)
  }

  return {
    tabs,
    activeTabId,
    activeTab,
    createTab,
    setActiveTab,
    closeTab,
    markActivity,
    getSocket,
    setSocket,
  }
})
```

- [ ] **Step 2: 创建终端面板组件**

```vue
<!-- frontend/src/features/terminal/TerminalPane.vue -->
<template>
  <div ref="containerRef" class="h-full w-full bg-background"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { TerminalSocket } from '@/shared/socket/terminal'
import '@xterm/xterm/css/xterm.css'

const props = defineProps<{
  sessionId: string
}>()

const emit = defineEmits<{
  (e: 'ready'): void
  (e: 'error', message: string): void
}>()

const containerRef = ref<HTMLElement | null>(null)
let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
let socket: TerminalSocket | null = null

onMounted(async () => {
  if (!containerRef.value) return

  // 创建终端实例
  terminal = new Terminal({
    fontFamily: 'JetBrains Mono, monospace',
    fontSize: 14,
    theme: {
      background: '#14141e',
      foreground: '#ffffff',
      cursor: '#ffffff',
      selection: 'rgba(255, 255, 255, 0.2)',
    },
    cursorBlink: true,
    cursorStyle: 'block',
  })

  // 加载插件
  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.loadAddon(new WebLinksAddon())

  // 打开终端
  terminal.open(containerRef.value)
  fitAddon.fit()

  // 连接 WebSocket
  socket = new TerminalSocket()
  await socket.connect(props.sessionId)

  // 处理终端输出
  socket.onMessage((message) => {
    if (message.type === 'output' && terminal) {
      terminal.write(message.data)
    } else if (message.type === 'error') {
      emit('error', message.message)
    } else if (message.type === 'exit') {
      terminal?.writeln(`\r\n进程已退出，退出码: ${message.code}`)
    }
  })

  // 处理终端输入
  terminal.onData((data) => {
    socket?.input(data)
  })

  // 处理终端大小变化
  terminal.onResize(({ cols, rows }) => {
    socket?.resize(cols, rows)
  })

  // 监听窗口大小变化
  window.addEventListener('resize', handleResize)

  emit('ready')
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  socket?.close()
  terminal?.dispose()
})

function handleResize() {
  fitAddon?.fit()
}

// 暴露方法供父组件调用
defineExpose({
  write: (data: string) => terminal?.write(data),
  clear: () => terminal?.clear(),
  resize: () => handleResize(),
})
</script>
```

- [ ] **Step 3: 创建终端工作区**

```vue
<!-- frontend/src/features/terminal/TerminalWorkspace.vue -->
<template>
  <div class="flex flex-col h-full bg-background">
    <!-- 标签栏 -->
    <div class="flex items-center bg-panel/50 border-b border-white/5">
      <div class="flex-1 flex items-center overflow-x-auto">
        <div
          v-for="tab in terminalStore.tabs"
          :key="tab.id"
          :class="[
            'flex items-center gap-2 px-4 py-2 border-r border-white/5 cursor-pointer',
            tab.id === terminalStore.activeTabId ? 'bg-surface' : 'hover:bg-white/5'
          ]"
          @click="terminalStore.setActiveTab(tab.id)"
        >
          <span class="text-white/80 text-sm">{{ tab.title }}</span>
          <span
            v-if="tab.hasActivity"
            class="w-2 h-2 rounded-full bg-primary"
          ></span>
          <button
            @click.stop="terminalStore.closeTab(tab.id)"
            class="text-white/40 hover:text-white ml-1"
          >
            <span class="material-symbols-outlined text-sm">close</span>
          </button>
        </div>
      </div>

      <button
        @click="handleCreateTab"
        class="px-4 py-2 text-white/40 hover:text-white hover:bg-white/5"
      >
        <span class="material-symbols-outlined">add</span>
      </button>
    </div>

    <!-- 终端内容区 -->
    <div class="flex-1 relative">
      <div
        v-for="tab in terminalStore.tabs"
        :key="tab.id"
        v-show="tab.id === terminalStore.activeTabId"
        class="absolute inset-0"
      >
        <TerminalPane
          :session-id="tab.id"
          @ready="handleTerminalReady(tab.id)"
        />
      </div>

      <div
        v-if="!terminalStore.tabs.length"
        class="flex items-center justify-center h-full text-white/50"
      >
        <p>点击 + 创建新终端</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useTerminalStore } from './terminalStore'
import TerminalPane from './TerminalPane.vue'

const terminalStore = useTerminalStore()

function handleCreateTab() {
  terminalStore.createTab()
}

function handleTerminalReady(tabId: string) {
  console.log('Terminal ready:', tabId)
}
</script>
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/terminal/
git commit -m "feat: add terminal module with xterm.js and WebSocket integration"
```

---

### Task 3: 主布局和工作区切换器

**Files:**
- Create: `frontend/src/features/workspace/WorkspaceMain.vue`
- Create: `frontend/src/shared/components/SidebarNav.vue`
- Create: `frontend/src/shared/components/WorkspaceSwitcher.vue`

- [ ] **Step 1: 创建工作区主页面**

```vue
<!-- frontend/src/features/workspace/WorkspaceMain.vue -->
<template>
  <div class="flex h-full">
    <!-- 左侧导航 -->
    <SidebarNav
      :active-tab="activeTab"
      :current-workspace="workspaceStore.currentWorkspace"
      @change="setActiveTab"
    />

    <!-- 主内容区 -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- 顶部栏：工作区切换器 -->
      <header class="h-12 bg-panel/50 border-b border-white/5 flex items-center px-4">
        <WorkspaceSwitcher
          :current-workspace="workspaceStore.currentWorkspace"
          :workspaces="workspaceStore.workspaces"
          @switch="handleSwitchWorkspace"
          @create="handleCreateWorkspace"
        />
      </header>

      <!-- 内容区 -->
      <main class="flex-1 overflow-hidden">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useWorkspaceStore } from './workspaceStore'
import SidebarNav from '@/shared/components/SidebarNav.vue'
import WorkspaceSwitcher from '@/shared/components/WorkspaceSwitcher.vue'

const router = useRouter()
const route = useRoute()
const workspaceStore = useWorkspaceStore()

const activeTab = ref('chat')

function setActiveTab(tab: string) {
  activeTab.value = tab
  const workspaceId = workspaceStore.currentWorkspace?.id
  if (workspaceId) {
    router.push(`/workspace/${workspaceId}/${tab}`)
  }
}

async function handleSwitchWorkspace(workspaceId: string) {
  await workspaceStore.openWorkspace(workspaceId)
  router.push(`/workspace/${workspaceId}`)
}

function handleCreateWorkspace() {
  router.push('/workspaces')
}

// 监听路由变化
watch(
  () => route.path,
  (path) => {
    if (path.includes('/chat')) activeTab.value = 'chat'
    else if (path.includes('/terminal')) activeTab.value = 'terminal'
    else if (path.includes('/settings')) activeTab.value = 'settings'
  },
  { immediate: true }
)
</script>
```

- [ ] **Step 2: 创建侧边导航**

```vue
<!-- frontend/src/shared/components/SidebarNav.vue -->
<template>
  <nav class="w-[88px] h-full flex flex-col items-center py-6 bg-panel/50 border-r border-white/5">
    <!-- 用户头像 -->
    <div class="mb-4">
      <div class="w-[52px] h-[52px] rounded-[18px] bg-surface border border-white/10 flex items-center justify-center">
        <span class="material-symbols-outlined text-2xl text-white/60">person</span>
      </div>
    </div>

    <div class="w-8 h-px bg-white/10 rounded-full mb-4"></div>

    <!-- 导航项 -->
    <div class="flex flex-col gap-4 w-full px-4">
      <button
        v-for="item in navItems"
        :key="item.id"
        @click="$emit('change', item.id)"
        :class="[
          'w-12 h-12 flex items-center justify-center rounded-2xl transition-all',
          activeTab === item.id
            ? 'bg-primary text-white'
            : 'bg-white/5 text-white/40 hover:bg-white/10 hover:text-white'
        ]"
        :title="item.tooltip"
      >
        <span class="material-symbols-outlined text-2xl">{{ item.icon }}</span>
      </button>
    </div>

    <!-- 设置按钮 -->
    <button
      @click="$emit('change', 'settings')"
      :class="[
        'w-12 h-12 flex items-center justify-center rounded-2xl transition-all mt-auto',
        activeTab === 'settings'
          ? 'text-white bg-white/10'
          : 'text-white/40 hover:text-white hover:bg-white/5'
      ]"
    >
      <span class="material-symbols-outlined text-2xl">settings</span>
    </button>
  </nav>
</template>

<script setup lang="ts">
const navItems = [
  { id: 'chat', icon: 'chat_bubble', tooltip: '聊天' },
  { id: 'terminal', icon: 'terminal', tooltip: '终端' },
]

defineProps<{
  activeTab: string
  currentWorkspace: { id: string; name: string } | null
}>()

defineEmits<{
  (e: 'change', tab: string): void
}>()
</script>
```

- [ ] **Step 3: 创建工作区切换器（Orchestra 特有功能）**

```vue
<!-- frontend/src/shared/components/WorkspaceSwitcher.vue -->
<template>
  <div class="flex items-center gap-3">
    <!-- 当前工作区名称 -->
    <button
      @click="showDropdown = !showDropdown"
      class="flex items-center gap-2 px-3 py-1.5 rounded-lg hover:bg-white/5 transition-colors"
    >
      <span class="material-symbols-outlined text-white/60 text-lg">folder</span>
      <span class="text-white font-medium text-sm">
        {{ currentWorkspace?.name || '选择工作区' }}
      </span>
      <span class="material-symbols-outlined text-white/40 text-lg">
        {{ showDropdown ? 'expand_less' : 'expand_more' }}
      </span>
    </button>

    <!-- 下拉菜单 -->
    <div
      v-if="showDropdown"
      class="absolute top-full left-4 mt-2 w-64 bg-panel-strong rounded-xl shadow-2xl border border-white/10 z-50"
    >
      <div class="p-2">
        <button
          v-for="ws in workspaces"
          :key="ws.id"
          @click="handleSwitch(ws.id)"
          :class="[
            'w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-colors',
            ws.id === currentWorkspace?.id ? 'bg-primary/20' : 'hover:bg-white/5'
          ]"
        >
          <span class="material-symbols-outlined text-white/60">folder</span>
          <div class="flex-1 min-w-0 text-left">
            <p class="text-white text-sm font-medium truncate">{{ ws.name }}</p>
            <p class="text-white/40 text-xs truncate">{{ ws.path }}</p>
          </div>
          <span
            v-if="ws.id === currentWorkspace?.id"
            class="material-symbols-outlined text-primary text-lg"
          >
            check
          </span>
        </button>
      </div>

      <div class="border-t border-white/5 p-2">
        <button
          @click="$emit('create')"
          class="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-white/5 text-white/60 hover:text-white transition-colors"
        >
          <span class="material-symbols-outlined">add</span>
          <span class="text-sm">新建工作区</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import type { Workspace } from '@/shared/types/workspace'

const props = defineProps<{
  currentWorkspace: Workspace | null
  workspaces: Workspace[]
}>()

const emit = defineEmits<{
  (e: 'switch', workspaceId: string): void
  (e: 'create'): void
}>()

const showDropdown = ref(false)

function handleSwitch(workspaceId: string) {
  emit('switch', workspaceId)
  showDropdown.value = false
}

function handleClickOutside(event: MouseEvent) {
  const target = event.target as HTMLElement
  if (!target.closest('.workspace-switcher')) {
    showDropdown.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/features/workspace/WorkspaceMain.vue frontend/src/shared/components/
git commit -m "feat: add workspace main layout with switcher for in-page workspace switching"
```

---

## 阶段完成检查

- [ ] 聊天模块完整实现
- [ ] 终端模块完整实现
- [ ] 工作区主布局完成
- [ ] 侧边导航完成
- [ ] 工作区切换器（特有功能）完成

---

**完成后继续:** Phase 5 - 集成测试和优化