<template>
  <div class="sessions-page-root animate-in fade-in zoom-in-95 duration-500">
    <div class="page-header">
      <div class="header-info">
        <h1 class="page-title">Agent 会话</h1>
        <p class="page-subtitle">{{ workspaceStore.currentWorkspace?.name }} · {{ sessions.length }} 个活跃后台会话</p>
      </div>
      <button class="refresh-btn" :disabled="loading" @click="loadSessions">
        <svg class="w-4 h-4" :class="loading ? 'animate-spin' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M4 4v6h6M20 20v-6h-6M20 9a8 8 0 00-14.906-4M4 15a8 8 0 0014.906 4" />
        </svg>
        <span>{{ loading ? '刷新中' : '刷新' }}</span>
      </button>
    </div>

    <div class="sessions-content custom-scrollbar">
      <div v-if="loading" class="loading-state">
        <div class="loading-spinner"></div>
        <p>Loading Sessions...</p>
      </div>

      <div v-else-if="sessions.length === 0" class="empty-state">
        <div class="empty-icon">
          <svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 8h10M7 12h6m-7 8h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
        </div>
        <p class="empty-title">暂无活跃 Agent 会话</p>
        <p class="empty-subtitle">在成员页启动 assistant 或 secretary 后，会话会显示在这里。</p>
      </div>

      <div v-else class="sessions-grid">
        <div v-for="session in enrichedSessions" :key="session.sessionId" class="session-card">
          <div class="card-top">
            <div class="session-avatar">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
              </svg>
            </div>
            <span class="status-pill">运行中</span>
          </div>

          <div class="session-info">
            <h3>{{ session.memberName }}</h3>
            <p>{{ roleLabel(session.roleType) }}</p>
          </div>

          <div class="session-meta">
            <span>Session</span>
            <code>{{ session.sessionId }}</code>
          </div>
          <div class="session-meta">
            <span>Member</span>
            <code>{{ session.memberId }}</code>
          </div>
          <div class="session-actions">
            <button class="inspect-btn" @click="connectTerminalStream(session.sessionId)">
              {{ activeSessionId === session.sessionId ? '重新连接' : '查看输出' }}
            </button>
            <button
              class="terminate-btn"
              :disabled="terminatingSessions[session.sessionId]"
              @click="terminateSession(session.sessionId)"
            >
              {{ terminatingSessions[session.sessionId] ? '终止中' : '终止' }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="activeSessionId" class="stream-panel">
        <div class="stream-header">
          <div>
            <p class="stream-title">终端输出</p>
            <p class="stream-subtitle">{{ activeSessionLabel }} · {{ streamStatusText }}</p>
          </div>
          <button class="disconnect-btn" @click="disconnectTerminalStream">断开</button>
        </div>
        <div class="stream-body">
          <TerminalSurface
            v-if="terminalEvents.length"
            :entries="terminalEvents"
            :status="streamStatus"
            @resize="sendTerminalResize"
          />
          <p v-else class="stream-empty">等待 agent 输出...</p>
        </div>
        <form class="stream-input-row" @submit.prevent="sendTerminalInput">
          <textarea
            v-model="terminalInput"
            class="stream-input"
            rows="2"
            placeholder="发送到当前 Agent 会话"
            :disabled="!canSendTerminalInput"
            @keydown.enter.exact.prevent="sendTerminalInput"
          ></textarea>
          <button class="send-btn" type="submit" :disabled="!canSendTerminalInput">发送</button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import client from '@/shared/api/client'
import { notifyUserError } from '@/shared/notifyError'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'
import TerminalSurface from '@/features/sessions/components/TerminalSurface.vue'

type WorkspaceSession = {
  memberId: string
  sessionId: string
}

type Member = {
  id: string
  name: string
  roleType: string
}

const workspaceStore = useWorkspaceStore()
const loading = ref(false)
const sessions = ref<WorkspaceSession[]>([])
const members = ref<Member[]>([])
const activeSessionId = ref('')
const terminalEvents = ref<string[]>([])
const terminalInput = ref('')
const terminatingSessions = ref<Record<string, boolean>>({})
const streamStatus = ref<'idle' | 'connecting' | 'connected' | 'closed' | 'error'>('idle')
let terminalWs: WebSocket | null = null

const memberById = computed(() => {
  return new Map(members.value.map((member) => [member.id, member]))
})

const enrichedSessions = computed(() => {
  return sessions.value.map((session) => {
    const member = memberById.value.get(session.memberId)
    return {
      ...session,
      memberName: member?.name || 'Unknown Member',
      roleType: member?.roleType || 'assistant',
    }
  })
})

const activeSessionLabel = computed(() => {
  const session = enrichedSessions.value.find((item) => item.sessionId === activeSessionId.value)
  return session ? `${session.memberName} / ${session.sessionId.slice(0, 8)}` : activeSessionId.value.slice(0, 8)
})

const streamStatusText = computed(() => {
  const map: Record<typeof streamStatus.value, string> = {
    idle: '未连接',
    connecting: '连接中',
    connected: '已连接',
    closed: '已断开',
    error: '连接错误',
  }
  return map[streamStatus.value]
})

const canSendTerminalInput = computed(() => {
  return !!terminalInput.value.trim() && streamStatus.value === 'connected' && terminalWs?.readyState === WebSocket.OPEN
})

function roleLabel(role: string) {
  const map: Record<string, string> = {
    owner: '所有者',
    admin: '管理员',
    assistant: 'AI 助手',
    secretary: '秘书',
    member: '成员',
  }
  return map[role] || role
}

async function loadSessions() {
  const wsId = workspaceStore.currentWorkspace?.id
  if (!wsId) return

  loading.value = true
  try {
    const [sessionResponse, memberResponse] = await Promise.all([
      client.get(`/workspaces/${wsId}/terminal-sessions`),
      client.get(`/workspaces/${wsId}/members`),
    ])
    sessions.value = sessionResponse.data?.sessions || []
    members.value = memberResponse.data || []
  } catch (e) {
    notifyUserError('Failed to load agent sessions', e)
  } finally {
    loading.value = false
  }
}

function terminalWsUrl(sessionId: string) {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host === 'localhost:5175' ? 'localhost:8080' : window.location.host
  const token = localStorage.getItem('orchestra.auth.token') || ''
  return `${protocol}//${host}/ws/terminal/${sessionId}?token=${encodeURIComponent(token)}`
}

function formatTerminalEvent(raw: string) {
  try {
    const data = JSON.parse(raw)
    const type = data.type || 'message'
    if (data.content) return `[${type}] ${data.content}`
    if (data.message) return `[${type}] ${data.message}`
    if (data.error) return `[${type}] ${data.error}`
    if (data.tool_name) return `[tool_use] ${data.tool_name}`
    if (data.sessionId) return `[${type}] ${data.sessionId}`
    return `[${type}] ${JSON.stringify(data)}`
  } catch {
    return raw
  }
}

function appendTerminalEvent(raw: string, formatted = true) {
  const text = formatted ? formatTerminalEvent(raw) : raw
  terminalEvents.value = [...terminalEvents.value.slice(-199), text]
}

function disconnectTerminalStream() {
  if (terminalWs) {
    terminalWs.onopen = null
    terminalWs.onmessage = null
    terminalWs.onerror = null
    terminalWs.onclose = null
    terminalWs.close()
    terminalWs = null
  }
  terminalInput.value = ''
  streamStatus.value = activeSessionId.value ? 'closed' : 'idle'
}

function closeActiveStreamIfSession(sessionId: string) {
  if (activeSessionId.value !== sessionId) return
  disconnectTerminalStream()
  activeSessionId.value = ''
  terminalEvents.value = []
}

async function loadTerminalSnapshot(sessionId: string) {
  try {
    const response = await client.get(`/terminals/${sessionId}/snapshot?lines=200`, { skipErrorToast: true })
    const content = String(response.data?.content || '').trimEnd()
    if (content) {
      appendTerminalEvent(`[snapshot]\n${content}`, false)
    }
  } catch (e) {
    appendTerminalEvent('[snapshot] 暂时无法读取当前终端屏幕', false)
  }
}

async function connectTerminalStream(sessionId: string) {
  disconnectTerminalStream()
  activeSessionId.value = sessionId
  terminalEvents.value = []
  streamStatus.value = 'connecting'
  await loadTerminalSnapshot(sessionId)

  const ws = new WebSocket(terminalWsUrl(sessionId))
  terminalWs = ws

  ws.onopen = () => {
    streamStatus.value = 'connected'
  }
  ws.onmessage = (event: MessageEvent) => {
    appendTerminalEvent(String(event.data))
  }
  ws.onerror = () => {
    streamStatus.value = 'error'
  }
  ws.onclose = () => {
    if (terminalWs === ws) {
      streamStatus.value = streamStatus.value === 'error' ? 'error' : 'closed'
      terminalWs = null
    }
  }
}

function sendTerminalInput() {
  const content = terminalInput.value.trim()
  if (!content || !terminalWs || terminalWs.readyState !== WebSocket.OPEN) return

  terminalWs.send(JSON.stringify({ type: 'input', content }))
  appendTerminalEvent(`[you]\n${content}`, false)
  terminalInput.value = ''
}

function sendTerminalResize(dimensions: { cols: number; rows: number }) {
  if (!terminalWs || terminalWs.readyState !== WebSocket.OPEN) return
  terminalWs.send(JSON.stringify({ type: 'resize', ...dimensions }))
}

async function terminateSession(sessionId: string) {
  terminatingSessions.value = { ...terminatingSessions.value, [sessionId]: true }
  try {
    await client.delete(`/terminals/${sessionId}`, { skipErrorToast: true })
    closeActiveStreamIfSession(sessionId)
    sessions.value = sessions.value.filter((session) => session.sessionId !== sessionId)
  } catch (e) {
    notifyUserError('Failed to terminate agent session', e)
  } finally {
    terminatingSessions.value = { ...terminatingSessions.value, [sessionId]: false }
  }
}

onMounted(loadSessions)
onBeforeUnmount(disconnectTerminalStream)
</script>

<style scoped>
.sessions-page-root {
  height: 100%; display: flex; flex-direction: column; gap: 28px; padding: 24px;
}

.page-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
.page-title { font-size: 32px; font-weight: 950; color: #0f172a; }
.page-subtitle { font-size: 15px; font-weight: 600; color: #475569; margin-top: 6px; }

.refresh-btn {
  display: flex; align-items: center; gap: 10px; padding: 12px 18px;
  border-radius: 14px; border: 1px solid #c7d2fe; background: white;
  color: #4338ca; font-size: 14px; font-weight: 900; cursor: pointer;
  box-shadow: 0 10px 24px rgba(79, 70, 229, 0.12);
}
.refresh-btn:disabled { cursor: not-allowed; opacity: 0.7; }

.sessions-content { flex: 1; overflow-y: auto; padding-bottom: 40px; }
.loading-state, .empty-state {
  height: 100%; min-height: 320px; display: flex; flex-direction: column;
  align-items: center; justify-content: center; color: #64748b; text-align: center;
}
.loading-spinner {
  width: 32px; height: 32px; border-radius: 999px; border: 4px solid rgba(79, 70, 229, 0.16);
  border-top-color: #4f46e5; animation: spin 1s linear infinite; margin-bottom: 14px;
}
.loading-state p { font-size: 12px; font-weight: 900; letter-spacing: 0.12em; text-transform: uppercase; }
.empty-icon {
  width: 56px; height: 56px; border-radius: 18px; display: flex; align-items: center; justify-content: center;
  background: #eef2ff; color: #4f46e5; margin-bottom: 16px;
}
.empty-title { font-size: 17px; font-weight: 900; color: #0f172a; }
.empty-subtitle { font-size: 13px; font-weight: 600; color: #64748b; margin-top: 6px; }

.sessions-grid {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 18px;
}
.session-card {
  background: rgba(255, 255, 255, 0.9); border: 1px solid rgba(226, 232, 240, 0.9);
  border-radius: 18px; padding: 20px; display: flex; flex-direction: column; gap: 18px;
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.05);
}
.card-top { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.session-avatar {
  width: 44px; height: 44px; border-radius: 14px; display: flex; align-items: center; justify-content: center;
  background: rgba(16, 185, 129, 0.1); color: #059669;
}
.status-pill {
  padding: 5px 10px; border-radius: 999px; background: #dcfce7; color: #059669;
  font-size: 10px; font-weight: 950; letter-spacing: 0.08em;
}
.session-info h3 { font-size: 17px; font-weight: 900; color: #0f172a; }
.session-info p { font-size: 12px; font-weight: 800; color: #64748b; margin-top: 4px; }
.session-meta { min-width: 0; display: flex; flex-direction: column; gap: 5px; }
.session-meta span { font-size: 10px; font-weight: 950; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.08em; }
.session-meta code {
  display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  padding: 8px 10px; border-radius: 10px; background: #f8fafc; color: #334155;
  font-size: 12px; font-weight: 800;
}
.session-actions { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 10px; }
.inspect-btn, .terminate-btn {
  width: 100%; padding: 10px 12px; border-radius: 12px; border: 1px solid #c7d2fe;
  background: white; color: #4338ca; font-size: 13px; font-weight: 900; cursor: pointer;
}
.inspect-btn:hover { background: #eef2ff; border-color: #818cf8; }
.terminate-btn {
  min-width: 70px; border-color: #fecaca; color: #dc2626;
}
.terminate-btn:hover:not(:disabled) { background: #fef2f2; border-color: #fca5a5; }
.terminate-btn:disabled { opacity: 0.65; cursor: not-allowed; }

.stream-panel {
  margin-top: 20px; background: rgba(15, 23, 42, 0.94); border-radius: 18px;
  overflow: hidden; border: 1px solid rgba(15, 23, 42, 0.2);
  box-shadow: 0 18px 42px rgba(15, 23, 42, 0.18);
}
.stream-header {
  display: flex; align-items: center; justify-content: space-between; gap: 16px;
  padding: 16px 18px; border-bottom: 1px solid rgba(148, 163, 184, 0.2);
}
.stream-title { font-size: 14px; font-weight: 950; color: #f8fafc; }
.stream-subtitle { font-size: 12px; font-weight: 700; color: #94a3b8; margin-top: 3px; }
.disconnect-btn {
  flex: 0 0 auto; padding: 8px 12px; border-radius: 10px; border: 1px solid rgba(148, 163, 184, 0.3);
  background: rgba(255, 255, 255, 0.06); color: #e2e8f0; font-size: 12px; font-weight: 900; cursor: pointer;
}
.disconnect-btn:hover { background: rgba(255, 255, 255, 0.12); }
.stream-body {
  min-height: 220px; overflow: hidden;
}
.stream-empty { padding: 18px; color: #94a3b8; font-size: 13px; font-weight: 700; }
.stream-input-row {
  display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 10px;
  padding: 14px 18px 16px; border-top: 1px solid rgba(148, 163, 184, 0.2);
}
.stream-input {
  min-width: 0; resize: vertical; max-height: 120px; border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.28); background: rgba(255, 255, 255, 0.06);
  color: #f8fafc; padding: 10px 12px; font-size: 13px; line-height: 1.4; outline: none;
}
.stream-input::placeholder { color: #64748b; }
.stream-input:focus { border-color: rgba(129, 140, 248, 0.75); }
.stream-input:disabled { opacity: 0.65; cursor: not-allowed; }
.send-btn {
  align-self: stretch; min-width: 72px; border-radius: 12px; border: 1px solid rgba(129, 140, 248, 0.45);
  background: #4f46e5; color: white; font-size: 13px; font-weight: 950; cursor: pointer;
}
.send-btn:disabled { opacity: 0.5; cursor: not-allowed; background: rgba(148, 163, 184, 0.35); }

@keyframes spin { to { transform: rotate(360deg); } }
</style>
