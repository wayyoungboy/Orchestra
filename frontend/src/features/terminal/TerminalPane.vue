<template>
  <div class="h-full w-full relative">
    <!-- Connection / error overlay (mirrors terminalStore; WS is created after tab mount) -->
    <div
      v-if="overlayKind === 'loading'"
      class="absolute inset-0 flex items-center justify-center bg-[#0b0f14] z-10"
    >
      <div class="flex items-center gap-3 text-white/60">
        <svg class="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <span class="text-sm">{{ overlayLoadingText }}</span>
      </div>
    </div>
    <div
      v-else-if="overlayKind === 'error'"
      class="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-[#0b0f14] z-10 px-6 text-center"
    >
      <span class="text-sm text-red-400/90">{{ t('terminal.overlayError') }}</span>
      <span class="text-xs text-white/40">{{ t('terminal.overlayErrorHint') }}</span>
    </div>

    <!-- Terminal Container -->
    <div
      ref="terminalRef"
      class="h-full w-full"
      @contextmenu.prevent="handleContextMenu"
    ></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { SearchAddon } from '@xterm/addon-search'
import '@xterm/xterm/css/xterm.css'
import { useTerminalStore } from './terminalStore'
import type { TerminalServerMessage } from '@/shared/types/terminal'

const props = defineProps<{
  terminalId: string
  active: boolean
}>()

const { t } = useI18n()
const terminalStore = useTerminalStore()
const { connectionStatus } = storeToRefs(terminalStore)
const terminalRef = ref<HTMLDivElement | null>(null)
const terminalReady = ref(false)

let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
let searchAddon: SearchAddon | null = null
let resizeObserver: ResizeObserver | null = null
let unsubscribe: (() => void) | null = null

const storeStatus = computed(() => connectionStatus.value[props.terminalId])

const overlayKind = computed(() => {
  const st = storeStatus.value
  if (st === 'connected' || st === 'working') return 'none'
  if (st === 'error' || st === 'disconnected') return 'error'
  return 'loading'
})

const overlayLoadingText = computed(() => {
  const st = storeStatus.value
  if (st === 'pending') return t('terminal.overlayPending')
  return t('terminal.overlayConnecting')
})

// Terminal theme configuration
const terminalTheme = {
  background: '#0b0f14',
  foreground: '#d4d4d4',
  cursor: '#d4d4d4',
  cursorAccent: '#0b0f14',
  selection: 'rgba(255, 255, 255, 0.1)',
  black: '#1e1e1e',
  red: '#f44747',
  green: '#4ec9b0',
  yellow: '#dcdcaa',
  blue: '#569cd6',
  magenta: '#c586c0',
  cyan: '#4fc1ff',
  white: '#d4d4d4',
  brightBlack: '#808080',
  brightRed: '#f44747',
  brightGreen: '#4ec9b0',
  brightYellow: '#dcdcaa',
  brightBlue: '#569cd6',
  brightMagenta: '#c586c0',
  brightCyan: '#4fc1ff',
  brightWhite: '#ffffff'
}

function initTerminal() {
  if (!terminalRef.value) return

  // 与 Golutra 一致：终端仅展示 PTY 输出，不向进程发送键盘输入（交互走聊天等入口）。
  terminal = new Terminal({
    theme: {
      ...terminalTheme,
      cursor: terminalTheme.background
    },
    fontFamily: 'JetBrains Mono, monospace',
    fontSize: 14,
    lineHeight: 1.2,
    cursorBlink: false,
    cursorStyle: 'block',
    scrollback: 10000,
    allowTransparency: true,
    disableStdin: true
  })

  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  searchAddon = new SearchAddon()
  terminal.loadAddon(searchAddon)

  terminal.attachCustomKeyEventHandler((domEvent) => {
    if (domEvent.ctrlKey && domEvent.key === 'f' && domEvent.type === 'keydown') {
      domEvent.preventDefault()
      const q = window.prompt('Find in terminal', '')
      if (q?.trim() && searchAddon) {
        searchAddon.findNext(q.trim(), { caseSensitive: false })
      }
      return false
    }
    return true
  })

  terminal.open(terminalRef.value)
  terminalReady.value = true

  // Fit terminal to container
  nextTick(() => {
    fitAddon?.fit()
    setupMessageHandler()
  })

  // Handle resize
  resizeObserver = new ResizeObserver(() => {
    fitAddon?.fit()
    const socket = terminalStore.getConnection(props.terminalId)
    if (socket && terminal) {
      socket.resize(terminal.cols, terminal.rows)
    }
  })
  resizeObserver.observe(terminalRef.value)
}

function setupMessageHandler() {
  unsubscribe?.()
  unsubscribe = null
  const socket = terminalStore.getConnection(props.terminalId)
  if (!socket || !terminal) return

  unsubscribe = socket.onMessage((message: TerminalServerMessage) => {
    if (!terminal) return

    switch (message.type) {
      case 'connected':
        terminalStore.connectionStatus[props.terminalId] = 'connected'
        // 服务端会紧跟重放 scrollback；清屏避免与上次连接内容叠在一起
        terminal.clear()
        break
      case 'output':
        terminalStore.connectionStatus[props.terminalId] = 'working'
        terminal.write(message.data)
        break
      case 'error':
        terminalStore.connectionStatus[props.terminalId] = 'error'
        terminal.writeln(`\x1b[31mError: ${message.message}\x1b[0m`)
        break
      case 'exit':
        terminal.writeln(`\x1b[33mProcess exited with code ${message.code}\x1b[0m`)
        break
      case 'terminal_chat_stream':
        break
    }
  })
}

function handleContextMenu() {
  // Can add context menu support later
}

function cleanup() {
  unsubscribe?.()
  unsubscribe = null
  resizeObserver?.disconnect()
  resizeObserver = null
  searchAddon?.dispose()
  searchAddon = null
  terminal?.dispose()
  terminal = null
  fitAddon = null
}

// Watch for active state changes
watch(() => props.active, (active) => {
  if (active && fitAddon) {
    nextTick(() => {
      fitAddon?.fit()
    })
  }
})

// WebSocket is stored in a non-reactive Map — watching getConnection() never re-runs.
// Follow connectionStatus + re-bind when the socket exists (after createConnection resolves).
watch(
  () =>
    [
      terminalReady.value,
      props.terminalId,
      connectionStatus.value[props.terminalId]
    ] as const,
  () => {
    if (!terminalReady.value || !terminal) return
    if (!terminalStore.getConnection(props.terminalId)) return
    nextTick(() => {
      setupMessageHandler()
      if (props.active) fitAddon?.fit()
    })
  },
  { flush: 'post' }
)

onMounted(() => {
  initTerminal()
})

onBeforeUnmount(() => {
  cleanup()
})
</script>

<style scoped>
:deep(.xterm) {
  padding: 8px;
}

:deep(.xterm-viewport) {
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.2) transparent;
}

:deep(.xterm-viewport::-webkit-scrollbar) {
  width: 8px;
}

:deep(.xterm-viewport::-webkit-scrollbar-thumb) {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
}

:deep(.xterm-viewport::-webkit-scrollbar-track) {
  background: transparent;
}
</style>