# Orchestra Phase 3 - 前端基础架构实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** 搭建 Vue 3 前端基础架构，包括状态管理、路由、API 调用层、WebSocket 层。

**Architecture:** Vue 3 + Pinia + Vue Router + Tailwind + xterm.js

**Tech Stack:** TypeScript, Vite, pnpm

---

## 文件结构

```
frontend/
├── src/
│   ├── app/
│   │   ├── App.vue
│   │   ├── main.ts
│   │   └── router.ts
│   ├── features/
│   │   ├── chat/
│   │   ├── terminal/
│   │   ├── workspace/
│   │   ├── agent/
│   │   └── settings/
│   ├── shared/
│   │   ├── api/
│   │   ├── socket/
│   │   ├── components/
│   │   └── utils/
│   └── assets/
├── index.html
├── vite.config.ts
├── tailwind.config.js
├── tsconfig.json
└── package.json
```

---

### Task 1: 项目初始化

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/tsconfig.json`
- Create: `frontend/tailwind.config.js`
- Create: `frontend/postcss.config.js`
- Create: `frontend/index.html`

- [ ] **Step 1: 创建 package.json**

```json
{
  "name": "orchestra-frontend",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc && vite build",
    "preview": "vite preview",
    "lint": "eslint . --ext .vue,.ts,.tsx --fix"
  },
  "dependencies": {
    "vue": "^3.4.0",
    "vue-router": "^4.3.0",
    "pinia": "^2.1.0",
    "@xterm/xterm": "^5.3.0",
    "@xterm/addon-fit": "^0.10.0",
    "@xterm/addon-web-links": "^0.11.0",
    "axios": "^1.6.0"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0.0",
    "typescript": "^5.3.0",
    "vite": "^5.0.0",
    "vue-tsc": "^2.0.0",
    "tailwindcss": "^3.4.0",
    "postcss": "^8.4.0",
    "autoprefixer": "^10.4.0",
    "eslint": "^8.56.0",
    "eslint-plugin-vue": "^9.20.0",
    "@typescript-eslint/eslint-plugin": "^6.19.0",
    "@typescript-eslint/parser": "^6.19.0"
  }
}
```

- [ ] **Step 2: 创建 vite.config.ts**

```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true
      }
    }
  }
})
```

- [ ] **Step 3: 创建 tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "module": "ESNext",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "preserve",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"]
    }
  },
  "include": ["src/**/*.ts", "src/**/*.tsx", "src/**/*.vue"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

- [ ] **Step 4: 创建 tailwind.config.js**

```javascript
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: 'rgb(var(--color-primary) / <alpha-value>)',
          hover: 'rgb(var(--color-primary-hover) / <alpha-value>)',
        },
        panel: 'rgb(var(--color-panel) / <alpha-value>)',
        surface: 'rgb(var(--color-surface) / <alpha-value>)',
        background: 'rgb(var(--color-background) / <alpha-value>)',
      },
      fontFamily: {
        sans: ['Be Vietnam Pro', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
    },
  },
  plugins: [],
}
```

- [ ] **Step 5: 创建 index.html**

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/vite.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Orchestra</title>
  </head>
  <body class="bg-background text-white">
    <div id="app"></div>
    <script type="module" src="/src/app/main.ts"></script>
  </body>
</html>
```

- [ ] **Step 6: 安装依赖**

Run: `cd frontend && pnpm install`
Expected: 依赖安装成功

- [ ] **Step 7: Commit**

```bash
git add frontend/
git commit -m "feat: initialize Vue 3 frontend with Vite and Tailwind"
```

---

### Task 2: API 调用层

**Files:**
- Create: `frontend/src/shared/api/client.ts`
- Create: `frontend/src/shared/api/workspace.ts`
- Create: `frontend/src/shared/api/member.ts`
- Create: `frontend/src/shared/types/workspace.ts`
- Create: `frontend/src/shared/types/member.ts`

- [ ] **Step 1: 创建 API 客户端**

```typescript
// frontend/src/shared/api/client.ts
import axios from 'axios'

const client = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

client.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('API Error:', error.response?.data || error.message)
    return Promise.reject(error)
  }
)

export default client
```

- [ ] **Step 2: 创建类型定义**

```typescript
// frontend/src/shared/types/workspace.ts
export interface Workspace {
  id: string
  name: string
  path: string
  lastOpenedAt: string
  createdAt: string
}

export interface WorkspaceCreate {
  name: string
  path: string
}

export interface FileInfo {
  name: string
  path: string
  isDir: boolean
  size: number
  modTime: string
  mode: string
}

export interface BrowseResult {
  basePath: string
  home?: string
  files: FileInfo[]
}
```

```typescript
// frontend/src/shared/types/member.ts
export type MemberRole = 'owner' | 'admin' | 'assistant' | 'member'

export interface Member {
  id: string
  workspaceId: string
  name: string
  roleType: MemberRole
  roleKey?: string
  avatar?: string
  terminalType?: string
  terminalCommand?: string
  terminalPath?: string
  autoStartTerminal: boolean
  status: string
  createdAt: string
}

export interface MemberCreate {
  name: string
  roleType: MemberRole
  terminalType?: string
  terminalCommand?: string
}
```

- [ ] **Step 3: 创建工作区 API**

```typescript
// frontend/src/shared/api/workspace.ts
import client from './client'
import type { Workspace, WorkspaceCreate, BrowseResult } from '@/shared/types/workspace'

export const workspaceApi = {
  list: () => client.get<Workspace[]>('/workspaces'),

  get: (id: string) => client.get<Workspace>(`/workspaces/${id}`),

  create: (data: WorkspaceCreate) => client.post<Workspace>('/workspaces', data),

  delete: (id: string) => client.delete(`/workspaces/${id}`),

  browse: (workspaceId: string, path?: string) =>
    client.get<BrowseResult>(`/workspaces/${workspaceId}/browse`, {
      params: { path },
    }),

  browseRoot: (path?: string) =>
    client.get<BrowseResult>('/browse', { params: { path } }),
}
```

- [ ] **Step 4: 创建成员 API**

```typescript
// frontend/src/shared/api/member.ts
import client from './client'
import type { Member, MemberCreate } from '@/shared/types/member'

export const memberApi = {
  list: (workspaceId: string) =>
    client.get<Member[]>(`/workspaces/${workspaceId}/members`),

  create: (workspaceId: string, data: MemberCreate) =>
    client.post<Member>(`/workspaces/${workspaceId}/members`, data),

  update: (workspaceId: string, memberId: string, data: Partial<Member>) =>
    client.put<Member>(`/workspaces/${workspaceId}/members/${memberId}`, data),

  delete: (workspaceId: string, memberId: string) =>
    client.delete(`/workspaces/${workspaceId}/members/${memberId}`),
}
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/shared/api/ frontend/src/shared/types/
git commit -m "feat: add API client and type definitions"
```

---

### Task 3: WebSocket 层

**Files:**
- Create: `frontend/src/shared/socket/terminal.ts`
- Create: `frontend/src/shared/socket/types.ts`

- [ ] **Step 1: 创建 WebSocket 类型**

```typescript
// frontend/src/shared/socket/types.ts
export interface TerminalInput {
  type: 'input'
  data: string
}

export interface TerminalResize {
  type: 'resize'
  cols: number
  rows: number
}

export interface TerminalClose {
  type: 'close'
}

export type TerminalClientMessage = TerminalInput | TerminalResize | TerminalClose

export interface TerminalOutput {
  type: 'output'
  data: string
}

export interface TerminalError {
  type: 'error'
  message: string
}

export interface TerminalExit {
  type: 'exit'
  code: number
}

export interface TerminalConnected {
  type: 'connected'
  sessionId: string
}

export type TerminalServerMessage =
  | TerminalOutput
  | TerminalError
  | TerminalExit
  | TerminalConnected
```

- [ ] **Step 2: 创建终端 WebSocket 客户端**

```typescript
// frontend/src/shared/socket/terminal.ts
import type { TerminalClientMessage, TerminalServerMessage } from './types'

type MessageHandler = (message: TerminalServerMessage) => void
type ErrorHandler = (error: Event) => void

export class TerminalSocket {
  private ws: WebSocket | null = null
  private messageHandlers: Set<MessageHandler> = new Set()
  private errorHandlers: Set<ErrorHandler> = new Set()
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5

  connect(sessionId: string): Promise<void> {
    return new Promise((resolve, reject) => {
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
          console.error('Failed to parse terminal message:', e)
        }
      }

      this.ws.onclose = () => {
        // 自动重连逻辑
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
          this.reconnectAttempts++
          setTimeout(() => this.connect(sessionId), 1000 * this.reconnectAttempts)
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
  }

  onMessage(handler: MessageHandler): () => void {
    this.messageHandlers.add(handler)
    return () => this.messageHandlers.delete(handler)
  }

  onError(handler: ErrorHandler): () => void {
    this.errorHandlers.add(handler)
    return () => this.errorHandlers.delete(handler)
  }
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/shared/socket/
git commit -m "feat: add terminal WebSocket client with reconnect support"
```

---

### Task 4: Pinia Store 基础

**Files:**
- Create: `frontend/src/features/workspace/workspaceStore.ts`
- Create: `frontend/src/features/workspace/projectStore.ts`
- Create: `frontend/src/app/main.ts`
- Create: `frontend/src/app/App.vue`

- [ ] **Step 1: 创建工作区 Store**

```typescript
// frontend/src/features/workspace/workspaceStore.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { workspaceApi } from '@/shared/api/workspace'
import type { Workspace, BrowseResult } from '@/shared/types/workspace'

export const useWorkspaceStore = defineStore('workspace', () => {
  const workspaces = ref<Workspace[]>([])
  const currentWorkspace = ref<Workspace | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const recentWorkspaces = computed(() =>
    [...workspaces.value].sort(
      (a, b) => new Date(b.lastOpenedAt).getTime() - new Date(a.lastOpenedAt).getTime()
    )
  )

  async function loadWorkspaces() {
    loading.value = true
    error.value = null
    try {
      const response = await workspaceApi.list()
      workspaces.value = response.data
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load workspaces'
    } finally {
      loading.value = false
    }
  }

  async function openWorkspace(id: string) {
    try {
      const response = await workspaceApi.get(id)
      currentWorkspace.value = response.data
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to open workspace'
    }
  }

  async function createWorkspace(name: string, path: string) {
    loading.value = true
    error.value = null
    try {
      const response = await workspaceApi.create({ name, path })
      workspaces.value.push(response.data)
      return response.data
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to create workspace'
      return null
    } finally {
      loading.value = false
    }
  }

  async function deleteWorkspace(id: string) {
    try {
      await workspaceApi.delete(id)
      workspaces.value = workspaces.value.filter((w) => w.id !== id)
      if (currentWorkspace.value?.id === id) {
        currentWorkspace.value = null
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to delete workspace'
    }
  }

  async function browseWorkspace(workspaceId: string, path?: string) {
    try {
      const response = await workspaceApi.browse(workspaceId, path)
      return response.data
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to browse workspace'
      return null
    }
  }

  function closeWorkspace() {
    currentWorkspace.value = null
  }

  return {
    workspaces,
    currentWorkspace,
    loading,
    error,
    recentWorkspaces,
    loadWorkspaces,
    openWorkspace,
    createWorkspace,
    deleteWorkspace,
    browseWorkspace,
    closeWorkspace,
  }
})
```

- [ ] **Step 2: 创建成员 Store**

```typescript
// frontend/src/features/workspace/projectStore.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { memberApi } from '@/shared/api/member'
import type { Member, MemberCreate } from '@/shared/types/member'
import { useWorkspaceStore } from './workspaceStore'

export const useProjectStore = defineStore('project', () => {
  const members = ref<Member[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const workspaceStore = useWorkspaceStore()

  const sortedMembers = computed(() =>
    [...members.value].sort((a, b) => {
      const roleOrder = { owner: 0, admin: 1, assistant: 2, member: 3 }
      return roleOrder[a.roleType] - roleOrder[b.roleType]
    })
  )

  async function loadMembers() {
    const workspaceId = workspaceStore.currentWorkspace?.id
    if (!workspaceId) return

    loading.value = true
    error.value = null
    try {
      const response = await memberApi.list(workspaceId)
      members.value = response.data
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load members'
    } finally {
      loading.value = false
    }
  }

  async function addMember(data: MemberCreate) {
    const workspaceId = workspaceStore.currentWorkspace?.id
    if (!workspaceId) return null

    try {
      const response = await memberApi.create(workspaceId, data)
      members.value.push(response.data)
      return response.data
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to add member'
      return null
    }
  }

  async function updateMember(memberId: string, data: Partial<Member>) {
    const workspaceId = workspaceStore.currentWorkspace?.id
    if (!workspaceId) return

    try {
      const response = await memberApi.update(workspaceId, memberId, data)
      const index = members.value.findIndex((m) => m.id === memberId)
      if (index !== -1) {
        members.value[index] = response.data
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to update member'
    }
  }

  async function removeMember(memberId: string) {
    const workspaceId = workspaceStore.currentWorkspace?.id
    if (!workspaceId) return

    try {
      await memberApi.delete(workspaceId, memberId)
      members.value = members.value.filter((m) => m.id !== memberId)
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to remove member'
    }
  }

  function reset() {
    members.value = []
    error.value = null
  }

  return {
    members,
    loading,
    error,
    sortedMembers,
    loadMembers,
    addMember,
    updateMember,
    removeMember,
    reset,
  }
})
```

- [ ] **Step 3: 创建 main.ts**

```typescript
// frontend/src/app/main.ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import '../assets/main.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
```

- [ ] **Step 4: 创建 App.vue**

```vue
<!-- frontend/src/app/App.vue -->
<template>
  <div class="h-screen w-screen bg-background text-white overflow-hidden">
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'

const workspaceStore = useWorkspaceStore()

onMounted(() => {
  workspaceStore.loadWorkspaces()
})
</script>
```

- [ ] **Step 5: 创建基础 CSS**

```css
/* frontend/src/assets/main.css */
@tailwind base;
@tailwind components;
@tailwind utilities;

:root {
  --color-primary: 99 102 241;
  --color-primary-hover: 79 70 229;
  --color-panel: 30 30 46;
  --color-surface: 40 40 60;
  --color-background: 20 20 30;
}

body {
  font-family: 'Be Vietnam Pro', sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* 终端样式 */
.xterm {
  padding: 8px;
}

/* 自定义滚动条 */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/app/ frontend/src/features/workspace/ frontend/src/assets/
git commit -m "feat: add Pinia stores and main app setup"
```

---

### Task 5: 路由配置

**Files:**
- Create: `frontend/src/app/router.ts`
- Create: `frontend/src/features/workspace/WorkspaceSelection.vue`
- Create: `frontend/src/features/workspace/PathBrowser.vue`

- [ ] **Step 1: 创建路由**

```typescript
// frontend/src/app/router.ts
import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'home',
    redirect: '/workspaces',
  },
  {
    path: '/workspaces',
    name: 'workspaces',
    component: () => import('@/features/workspace/WorkspaceSelection.vue'),
  },
  {
    path: '/workspace/:id',
    name: 'workspace',
    component: () => import('@/features/workspace/WorkspaceMain.vue'),
    children: [
      {
        path: '',
        redirect: { name: 'chat' },
      },
      {
        path: 'chat',
        name: 'chat',
        component: () => import('@/features/chat/ChatInterface.vue'),
      },
      {
        path: 'terminal',
        name: 'terminal',
        component: () => import('@/features/terminal/TerminalWorkspace.vue'),
      },
      {
        path: 'settings',
        name: 'settings',
        component: () => import('@/features/settings/Settings.vue'),
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
```

- [ ] **Step 2: 创建工作区选择页**

```vue
<!-- frontend/src/features/workspace/WorkspaceSelection.vue -->
<template>
  <div class="min-h-screen flex flex-col items-center justify-center p-6">
    <!-- 打开新工作区 -->
    <div class="w-full max-w-2xl mb-10">
      <button
        @click="showPathBrowser = true"
        class="w-full bg-panel/40 rounded-3xl p-10 flex flex-col items-center justify-center text-center hover:bg-panel/60 transition-all border border-white/5 hover:border-primary/20"
      >
        <div class="w-14 h-14 rounded-full bg-white/5 flex items-center justify-center mb-4 border border-white/10">
          <span class="material-symbols-outlined text-3xl text-gray-400">folder_open</span>
        </div>
        <h1 class="text-2xl font-bold text-white mb-2">打开工作空间</h1>
        <p class="text-gray-400 font-medium">选择服务器上的工作目录</p>
      </button>
    </div>

    <!-- 最近工作区 -->
    <div class="w-full max-w-6xl">
      <h2 class="text-xs font-bold text-gray-500 tracking-widest uppercase mb-6">最近使用</h2>

      <div v-if="!workspaceStore.recentWorkspaces.length" class="text-center text-white/50 py-12">
        <p class="text-sm">暂无最近使用的工作区</p>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <button
          v-for="ws in workspaceStore.recentWorkspaces"
          :key="ws.id"
          @click="openWorkspace(ws.id)"
          class="bg-panel/40 rounded-3xl p-5 text-left hover:bg-panel/60 transition-all border border-white/5 hover:border-primary/20"
        >
          <div class="w-10 h-10 rounded-2xl bg-white/10 flex items-center justify-center mb-4 border border-white/10">
            <span class="material-symbols-outlined text-xl text-white/60">folder</span>
          </div>
          <h3 class="text-lg font-bold text-white mb-1 truncate">{{ ws.name }}</h3>
          <p class="text-xs text-white/40 truncate">{{ ws.path }}</p>
        </button>
      </div>
    </div>

    <!-- 路径浏览器弹窗 -->
    <PathBrowser
      v-if="showPathBrowser"
      @close="showPathBrowser = false"
      @select="handlePathSelect"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useWorkspaceStore } from './workspaceStore'
import PathBrowser from './PathBrowser.vue'

const router = useRouter()
const workspaceStore = useWorkspaceStore()
const showPathBrowser = ref(false)

onMounted(() => {
  workspaceStore.loadWorkspaces()
})

async function openWorkspace(id: string) {
  await workspaceStore.openWorkspace(id)
  router.push(`/workspace/${id}`)
}

async function handlePathSelect(path: string, name: string) {
  const ws = await workspaceStore.createWorkspace(name, path)
  if (ws) {
    router.push(`/workspace/${ws.id}`)
  }
  showPathBrowser.value = false
}
</script>
```

- [ ] **Step 3: 创建路径浏览器组件**

```vue
<!-- frontend/src/features/workspace/PathBrowser.vue -->
<template>
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
    <div class="bg-panel rounded-3xl w-full max-w-2xl max-h-[80vh] flex flex-col border border-white/10">
      <!-- 头部 -->
      <div class="flex items-center justify-between p-4 border-b border-white/5">
        <h2 class="text-lg font-bold text-white">选择工作目录</h2>
        <button @click="$emit('close')" class="text-white/40 hover:text-white">
          <span class="material-symbols-outlined">close</span>
        </button>
      </div>

      <!-- 路径输入 -->
      <div class="p-4 border-b border-white/5">
        <input
          v-model="currentPath"
          type="text"
          class="w-full bg-surface text-white rounded-xl px-4 py-3 border border-white/10 focus:border-primary/50 focus:outline-none"
          placeholder="输入路径..."
          @keyup.enter="loadDirectory(currentPath)"
        />
      </div>

      <!-- 文件列表 -->
      <div class="flex-1 overflow-y-auto p-4">
        <div v-if="loading" class="text-center py-8 text-white/50">
          加载中...
        </div>
        <div v-else-if="error" class="text-center py-8 text-red-400">
          {{ error }}
        </div>
        <div v-else-if="!files.length" class="text-center py-8 text-white/50">
          空目录
        </div>
        <div v-else class="space-y-1">
          <button
            v-for="file in files"
            :key="file.path"
            @click="handleFileClick(file)"
            class="w-full flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-white/5 transition-colors"
          >
            <span class="material-symbols-outlined text-white/40">
              {{ file.isDir ? 'folder' : 'description' }}
            </span>
            <span class="text-white/80">{{ file.name }}</span>
          </button>
        </div>
      </div>

      <!-- 底部 -->
      <div class="flex items-center justify-between p-4 border-t border-white/5">
        <input
          v-model="workspaceName"
          type="text"
          class="flex-1 bg-surface text-white rounded-xl px-4 py-2 mr-4 border border-white/10 focus:border-primary/50 focus:outline-none"
          placeholder="工作区名称"
        />
        <button
          @click="confirm"
          :disabled="!currentPath || !workspaceName"
          class="px-6 py-2 bg-primary text-white rounded-xl font-medium disabled:opacity-50 disabled:cursor-not-allowed hover:bg-primary-hover transition-colors"
        >
          确认
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { workspaceApi } from '@/shared/api/workspace'
import type { FileInfo, BrowseResult } from '@/shared/types/workspace'

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'select', path: string, name: string): void
}>()

const currentPath = ref('')
const workspaceName = ref('')
const files = ref<FileInfo[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

onMounted(async () => {
  await loadRoot()
})

async function loadRoot() {
  loading.value = true
  error.value = null
  try {
    const response = await workspaceApi.browseRoot()
    const data: BrowseResult = response.data
    currentPath.value = data.basePath
    files.value = data.files
    workspaceName.value = data.basePath.split('/').pop() || ''
  } catch (e) {
    error.value = '无法加载目录'
  } finally {
    loading.value = false
  }
}

async function loadDirectory(path: string) {
  loading.value = true
  error.value = null
  try {
    const response = await workspaceApi.browseRoot(path)
    const data: BrowseResult = response.data
    currentPath.value = data.basePath
    files.value = data.files
  } catch (e) {
    error.value = '无法加载目录'
  } finally {
    loading.value = false
  }
}

function handleFileClick(file: FileInfo) {
  if (file.isDir) {
    loadDirectory(file.path)
  }
}

function confirm() {
  if (currentPath.value && workspaceName.value) {
    emit('select', currentPath.value, workspaceName.value)
  }
}
</script>
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/router.ts frontend/src/features/workspace/
git commit -m "feat: add router and workspace selection UI with path browser"
```

---

## 阶段完成检查

- [ ] 前端项目初始化完成
- [ ] API 调用层正常工作
- [ ] WebSocket 层正常工作
- [ ] Pinia Store 正常工作
- [ ] 路由配置完成
- [ ] 工作区选择页面可用

---

**完成后继续:** Phase 4 - 聊天和终端 UI