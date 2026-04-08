<template>
  <div ref="terminalRef" class="terminal-pane"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { CanvasAddon } from '@xterm/addon-canvas'
import '@xterm/xterm/css/xterm.css'
import { useTerminalStore } from './terminalStore'

const props = defineProps<{
  sessionId: string
}>()

const terminalStore = useTerminalStore()
const terminalRef = ref<HTMLElement | null>(null)
let term: Terminal | null = null
let fitAddon: FitAddon | null = null
let canvasAddon: CanvasAddon | null = null
let socket: WebSocket | null = null
let resizeObserver: ResizeObserver | null = null

function connect() {
  if (socket) socket.close()

  if (!props.sessionId) {
    console.warn('TerminalPane: no sessionId provided')
    return
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host === 'localhost:5175' ? 'localhost:8080' : window.location.host
  const wsUrl = `${protocol}//${host}/ws/terminal/${props.sessionId}`

  console.log('TerminalPane connecting to:', wsUrl)

  socket = new WebSocket(wsUrl)

  socket.onopen = () => {
    terminalStore.updateSessionStatus(props.sessionId, 'connected')
    // Send initial size
    if (fitAddon && term) {
      fitAddon.fit()
      sendResize()
    }
  }

  socket.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      if (msg.type === 'output' && msg.data && term) {
        term.write(msg.data)
      } else if (msg.type === 'error') {
        console.error('Terminal error:', msg.message)
      }
      // Ignore other message types like 'connected'
    } catch {
      // If not JSON, write directly (backward compatibility)
      if (term) term.write(event.data)
    }
  }

  socket.onclose = () => {
    terminalStore.updateSessionStatus(props.sessionId, 'disconnected')
  }

  socket.onerror = (error) => {
    console.error('Terminal WS Error:', error)
  }
}

function sendResize() {
  if (socket?.readyState === WebSocket.OPEN && term) {
    socket.send(JSON.stringify({
      type: 'resize',
      cols: term.cols,
      rows: term.rows
    }))
  }
}

function fitTerminal() {
  if (fitAddon && term && terminalRef.value) {
    // Only fit if the element is visible
    const rect = terminalRef.value.getBoundingClientRect()
    if (rect.width > 0 && rect.height > 0) {
      fitAddon.fit()
      sendResize()
    }
  }
}

onMounted(async () => {
  if (!terminalRef.value) return

  term = new Terminal({
    cursorBlink: true,
    fontFamily: "'JetBrains Mono', 'Fira Code', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace",
    fontSize: 13,
    scrollback: 5000,
    theme: {
      background: '#0b0f14',
      foreground: '#e5e7eb',
      cursor: '#38bdf8',
      selectionBackground: 'rgba(56, 189, 248, 0.3)'
    }
  })

  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.open(terminalRef.value)

  // Load canvas renderer for better performance
  try {
    canvasAddon = new CanvasAddon()
    term.loadAddon(canvasAddon)
  } catch (e) {
    console.warn('Canvas addon failed to load, using DOM renderer', e)
  }

  // Wait for fonts to be ready
  if ('fonts' in document) {
    try {
      await document.fonts.ready
    } catch {
      // Ignore font ready failure
    }
  }

  await nextTick()
  fitAddon.fit()

  term.onData((data) => {
    if (socket?.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({
        type: 'input',
        data: data
      }))
    }
  })

  // Connect immediately if sessionId is provided
  if (props.sessionId) {
    connect()
  }

  // Use ResizeObserver to detect size changes
  resizeObserver = new ResizeObserver(() => {
    fitTerminal()
  })
  resizeObserver.observe(terminalRef.value)

  window.addEventListener('resize', handleResize)
})

function handleResize() {
  fitTerminal()
}

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  if (resizeObserver) {
    resizeObserver.disconnect()
  }
  if (socket) socket.close()
  if (canvasAddon) {
    canvasAddon.dispose()
    canvasAddon = null
  }
  if (term) term.dispose()
})

// Focus terminal when sessionId becomes active
watch(() => terminalStore.activeSessionId, async (newId) => {
  if (newId === props.sessionId) {
    // Wait for the DOM to update and element to become visible
    await nextTick()
    setTimeout(() => {
      fitTerminal()
      term?.focus()
    }, 50)
  }
})

// Connect when sessionId changes
watch(() => props.sessionId, (newId) => {
  if (newId) {
    connect()
  }
}, { immediate: false })
</script>

<style scoped>
.terminal-pane {
  width: 100%;
  height: 100%;
  background: #0b0f14;
}

:deep(.xterm-viewport) {
  background-color: transparent !important;
}

:deep(.xterm-screen) {
  padding: 8px;
}
</style>
