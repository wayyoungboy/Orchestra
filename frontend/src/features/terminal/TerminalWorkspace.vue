<template>
  <section class="terminal-workspace">
    <aside class="agent-list">
      <div class="panel-heading">
        <p class="eyebrow">AGENT TERMINALS</p>
        <h1>终端工作区</h1>
        <p>选择已配置的智能体，启动或重新连接其持久会话。</p>
      </div>

      <div v-if="loading" class="agent-empty">正在加载成员…</div>
      <div v-else-if="!agentMembers.length" class="agent-empty">
        暂无已配置 ACP 命令的成员。请在“成员”页面添加 AI 助手并填写 CLI 命令。
      </div>
      <button
        v-for="member in agentMembers"
        :key="member.id"
        type="button"
        class="agent-card"
        :class="{ selected: selectedMember?.id === member.id }"
        @click="selectMember(member)"
      >
        <span class="agent-avatar">{{ member.name.slice(0, 1).toUpperCase() }}</span>
        <span class="agent-details">
          <strong>{{ member.name }}</strong>
          <code>{{ member.acpCommand }}</code>
        </span>
        <span class="agent-status" :class="connectionState">{{ connectionLabel }}</span>
      </button>
    </aside>

    <main class="terminal-panel">
      <template v-if="selectedMember">
        <header class="terminal-header">
          <div>
            <p class="eyebrow">{{ selectedMember.acpCommand }}</p>
            <h2>{{ selectedMember.name }}</h2>
          </div>
          <div class="terminal-actions">
            <button type="button" class="button secondary" :disabled="starting" @click="startOrReconnect">
              {{ starting ? '正在连接…' : sessionId ? '重新连接' : '启动会话' }}
            </button>
            <button v-if="sessionId" type="button" class="button danger" @click="closeSession">结束会话</button>
          </div>
        </header>

        <p v-if="error" class="terminal-error">{{ error }}</p>
        <div ref="terminalHost" class="terminal-host" aria-label="Agent terminal output"></div>
        <p class="terminal-help">输入完整提示后按 Enter 发送给智能体。此会话由 tmux 托管，刷新页面后可重新连接。</p>
      </template>

      <div v-else class="terminal-placeholder">
        <div>⌘</div>
        <h2>选择一个智能体</h2>
        <p>只会展示已启用 ACP 且配置了命令的成员。</p>
      </div>
    </main>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { terminalApi } from '@/shared/api/terminal'
import { getApiErrorMessage } from '@/shared/api/errors'
import { terminalWebSocketURL, type TerminalEvent } from '@/shared/socket/terminal'
import { useProjectStore } from '@/features/workspace/projectStore'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'
import type { Member } from '@/shared/types/member'

const projectStore = useProjectStore()
const workspaceStore = useWorkspaceStore()
const terminalHost = ref<HTMLElement | null>(null)
const selectedMember = ref<Member | null>(null)
const sessionId = ref<string | null>(null)
const loading = ref(false)
const starting = ref(false)
const error = ref<string | null>(null)
const connectionState = ref<'idle' | 'connecting' | 'connected' | 'disconnected'>('idle')

let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
let socket: WebSocket | null = null
let inputBuffer = ''
let resizeObserver: ResizeObserver | null = null

const agentMembers = computed(() =>
  projectStore.sortedMembers.filter((member) => member.acpEnabled && !!member.acpCommand),
)

const connectionLabel = computed(() => {
  switch (connectionState.value) {
    case 'connected': return '已连接'
    case 'connecting': return '连接中'
    case 'disconnected': return '已断开'
    default: return '未启动'
  }
})

function setupTerminal() {
  if (!terminalHost.value || terminal) return
  terminal = new Terminal({
    convertEol: true,
    cursorBlink: true,
    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
    fontSize: 13,
    theme: {
      background: '#080d14',
      foreground: '#d7e1ef',
      cursor: '#a78bfa',
      selectionBackground: '#4338ca66',
    },
  })
  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.open(terminalHost.value)
  fitTerminal()
  terminal.writeln('\x1b[38;5;141mOrchestra agent terminal\x1b[0m')
  terminal.writeln('Select an agent and start a session.\r\n')
  terminal.onData(handleTerminalInput)
  resizeObserver = new ResizeObserver(fitTerminal)
  resizeObserver.observe(terminalHost.value)
}

function fitTerminal() {
  try {
    fitAddon?.fit()
  } catch {
    // The panel can be hidden while the workspace route is changing.
  }
}

function writeLine(text: string) {
  terminal?.writeln(text.replace(/\n/g, '\r\n'))
}

function handleTerminalInput(data: string) {
  if (!terminal) return
  for (const char of data) {
    if (char === '\r') {
      const content = inputBuffer.trim()
      terminal.write('\r\n')
      inputBuffer = ''
      if (content) sendInput(content)
      continue
    }
    if (char === '\u007f') {
      if (inputBuffer.length > 0) {
        inputBuffer = inputBuffer.slice(0, -1)
        terminal.write('\b \b')
      }
      continue
    }
    if (char >= ' ') {
      inputBuffer += char
      terminal.write(char)
    }
  }
}

function sendInput(content: string) {
  if (!socket || socket.readyState !== WebSocket.OPEN) {
    writeLine('\x1b[31mNot connected. Start or reconnect the session first.\x1b[0m')
    return
  }
  socket.send(JSON.stringify({ type: 'user_message', content }))
}

function closeSocket() {
  const current = socket
  socket = null
  if (current && current.readyState <= WebSocket.OPEN) current.close()
}

function renderEvent(event: TerminalEvent) {
  switch (event.type) {
    case 'connected':
      writeLine('\x1b[32m✓ Connected to tmux-backed agent session.\x1b[0m')
      break
    case 'assistant_message':
      writeLine(`\x1b[38;5;81m${selectedMember.value?.name || 'agent'}›\x1b[0m ${event.content || ''}`)
      break
    case 'tool_use':
      writeLine(`\x1b[33m→ tool ${event.tool_name || 'unknown'} ${event.tool_input ? JSON.stringify(event.tool_input) : ''}\x1b[0m`)
      break
    case 'result':
      writeLine(`\x1b[32m✓ ${event.message || 'Agent finished'}\x1b[0m`)
      break
    case 'status':
      writeLine(`\x1b[90m[status] ${event.status || ''}\x1b[0m`)
      break
    case 'error':
      writeLine(`\x1b[31m✗ ${event.error || 'Agent error'}\x1b[0m`)
      break
    case 'exit':
      writeLine('\x1b[90mSession exited.\x1b[0m')
      connectionState.value = 'disconnected'
      break
  }
}

function connect(session: string) {
  closeSocket()
  connectionState.value = 'connecting'
  const nextSocket = new WebSocket(terminalWebSocketURL(session))
  socket = nextSocket
  nextSocket.onopen = () => {
    if (socket === nextSocket) connectionState.value = 'connected'
  }
  nextSocket.onmessage = (message) => {
    try {
      renderEvent(JSON.parse(message.data) as TerminalEvent)
    } catch {
      writeLine(String(message.data))
    }
  }
  nextSocket.onerror = () => {
    error.value = '终端连接失败，请检查后端、tmux 和智能体命令。'
  }
  nextSocket.onclose = () => {
    if (socket === nextSocket) connectionState.value = 'disconnected'
  }
}

async function selectMember(member: Member) {
  if (selectedMember.value?.id === member.id) return
  closeSocket()
  selectedMember.value = member
  sessionId.value = null
  error.value = null
  connectionState.value = 'idle'
  inputBuffer = ''
  await nextTick()
  setupTerminal()
  writeLine(`\r\n\x1b[38;5;141mSelected ${member.name} (${member.acpCommand}).\x1b[0m`)
  fitTerminal()
}

async function startOrReconnect() {
  const workspaceId = workspaceStore.currentWorkspace?.id
  const member = selectedMember.value
  if (!workspaceId || !member) return

  starting.value = true
  error.value = null
  try {
    const { data } = await terminalApi.getOrCreate(workspaceId, member.id)
    sessionId.value = data.sessionId
    connect(data.sessionId)
  } catch (cause) {
    error.value = getApiErrorMessage(cause)
    connectionState.value = 'disconnected'
  } finally {
    starting.value = false
  }
}

async function closeSession() {
  if (!sessionId.value) return
  const currentSession = sessionId.value
  closeSocket()
  try {
    await terminalApi.close(currentSession)
    writeLine('\x1b[90mSession closed.\x1b[0m')
    sessionId.value = null
    connectionState.value = 'idle'
  } catch (cause) {
    error.value = getApiErrorMessage(cause)
  }
}

onMounted(async () => {
  const workspaceId = workspaceStore.currentWorkspace?.id
  if (!workspaceId) return
  loading.value = true
  try {
    await projectStore.loadMembers(workspaceId)
  } finally {
    loading.value = false
  }
})

onBeforeUnmount(() => {
  closeSocket()
  resizeObserver?.disconnect()
  terminal?.dispose()
  terminal = null
  fitAddon = null
})
</script>

<style scoped>
.terminal-workspace { height: 100%; display: grid; grid-template-columns: minmax(230px, 300px) minmax(0, 1fr); gap: 18px; padding: 18px; }
.agent-list, .terminal-panel { min-height: 0; border: 1px solid rgba(148, 163, 184, .25); border-radius: 24px; background: rgba(255,255,255,.54); backdrop-filter: blur(24px); box-shadow: 0 16px 40px rgba(15,23,42,.05); }
.agent-list { padding: 18px; overflow-y: auto; }
.panel-heading { margin-bottom: 18px; }
.eyebrow { margin: 0 0 6px; color: #64748b; font-size: 10px; font-weight: 800; letter-spacing: .12em; }
h1, h2 { margin: 0; color: #0f172a; font-weight: 800; } h1 { font-size: 20px; } h2 { font-size: 18px; }
.panel-heading > p:last-child, .terminal-help { color: #64748b; font-size: 12px; line-height: 1.5; }
.agent-card { display: flex; align-items: center; width: 100%; gap: 10px; margin: 8px 0; padding: 12px; text-align: left; border: 1px solid transparent; border-radius: 14px; background: transparent; cursor: pointer; }
.agent-card:hover, .agent-card.selected { border-color: rgba(99,102,241,.25); background: rgba(238,242,255,.75); }
.agent-avatar { display: grid; width: 32px; height: 32px; place-items: center; border-radius: 10px; background: #4f46e5; color: white; font-weight: 800; }
.agent-details { display: flex; flex: 1; min-width: 0; flex-direction: column; gap: 2px; }.agent-details strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #1e293b; font-size: 13px; }.agent-details code { color: #64748b; font-size: 11px; }
.agent-status { color: #64748b; font-size: 10px; white-space: nowrap; }.agent-status.connected { color: #059669; }.agent-status.connecting { color: #b45309; }
.agent-empty { padding: 18px 8px; color: #64748b; font-size: 13px; line-height: 1.55; }
.terminal-panel { display: flex; min-width: 0; flex-direction: column; padding: 18px; }.terminal-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding-bottom: 14px; }.terminal-actions { display: flex; gap: 8px; }.button { border: 0; border-radius: 10px; padding: 9px 12px; font-size: 12px; font-weight: 700; cursor: pointer; }.button.secondary { background: #4f46e5; color: white; }.button.danger { background: #fee2e2; color: #b91c1c; }.button:disabled { opacity: .6; cursor: wait; }
.terminal-error { margin: 0 0 12px; border-radius: 10px; background: #fef2f2; color: #b91c1c; padding: 10px; font-size: 13px; }.terminal-host { flex: 1; min-height: 260px; overflow: hidden; border: 1px solid #111827; border-radius: 16px; background: #080d14; padding: 10px; }.terminal-help { margin: 10px 0 0; }.terminal-placeholder { display: grid; height: 100%; min-height: 340px; place-content: center; text-align: center; color: #64748b; }.terminal-placeholder > div { margin-bottom: 10px; font-size: 32px; color: #4f46e5; }.terminal-placeholder p { font-size: 13px; }
@media (max-width: 800px) { .terminal-workspace { grid-template-columns: 1fr; overflow-y: auto; }.agent-list { max-height: 230px; }.terminal-panel { min-height: 460px; } }
</style>
