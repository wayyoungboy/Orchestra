<template>
  <div ref="containerRef" class="terminal-surface" data-test="terminal-surface"></div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'

const props = defineProps<{
  entries: string[]
  status: 'idle' | 'connecting' | 'connected' | 'closed' | 'error'
}>()

const emit = defineEmits<{
  resize: [dimensions: { cols: number; rows: number }]
  data: [input: string]
}>()

const containerRef = ref<HTMLElement | null>(null)
let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
let resizeObserver: ResizeObserver | null = null
let writtenEntries = 0
let lastDimensions = ''

function normalizeChunk(value: string) {
  return value.replace(/\r?\n/g, '\r\n')
}

function fitTerminal() {
  requestAnimationFrame(() => {
    if (!terminal || !fitAddon) return
    try {
      fitAddon.fit()
      const dimensions = `${terminal.cols}x${terminal.rows}`
      if (dimensions !== lastDimensions) {
        lastDimensions = dimensions
        emit('resize', { cols: terminal.cols, rows: terminal.rows })
      }
    } catch {
      // The fit addon can throw while the element is being mounted or hidden.
    }
  })
}

function writeEntries() {
  if (!terminal) return
  if (props.entries.length < writtenEntries) {
    terminal.clear()
    writtenEntries = 0
  }

  for (const entry of props.entries.slice(writtenEntries)) {
    terminal.write(`${normalizeChunk(entry)}\r\n`)
  }
  writtenEntries = props.entries.length
}

onMounted(async () => {
  await nextTick()
  if (!containerRef.value) return

  const term = new Terminal({
    convertEol: true,
    cursorBlink: props.status === 'connected',
    disableStdin: props.status !== 'connected',
    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
    fontSize: 13,
    lineHeight: 1.35,
    scrollback: 3000,
    theme: {
      background: '#0f172a',
      foreground: '#dbeafe',
      cursor: '#a5b4fc',
      selectionBackground: '#334155',
      black: '#0f172a',
      red: '#f87171',
      green: '#34d399',
      yellow: '#fbbf24',
      blue: '#60a5fa',
      magenta: '#c084fc',
      cyan: '#22d3ee',
      white: '#e2e8f0',
      brightBlack: '#64748b',
      brightRed: '#fca5a5',
      brightGreen: '#86efac',
      brightYellow: '#fde68a',
      brightBlue: '#93c5fd',
      brightMagenta: '#d8b4fe',
      brightCyan: '#67e8f9',
      brightWhite: '#f8fafc',
    },
  })
  const fit = new FitAddon()

  term.loadAddon(fit)
  term.loadAddon(new WebLinksAddon())
  term.onData((input) => {
    if (props.status === 'connected') {
      emit('data', input)
    }
  })
  term.open(containerRef.value)

  terminal = term
  fitAddon = fit
  resizeObserver = new ResizeObserver(fitTerminal)
  resizeObserver.observe(containerRef.value)

  writeEntries()
  fitTerminal()
})

watch(
  () => props.entries,
  () => {
    writeEntries()
    fitTerminal()
  },
  { deep: true }
)

watch(
  () => props.status,
  (status) => {
    if (terminal) {
      terminal.options.cursorBlink = status === 'connected'
      terminal.options.disableStdin = status !== 'connected'
    }
  }
)

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
  terminal?.dispose()
  terminal = null
  fitAddon = null
})
</script>

<style scoped>
.terminal-surface {
  height: 340px;
  min-height: 220px;
  width: 100%;
  overflow: hidden;
  background: #0f172a;
}

.terminal-surface :deep(.xterm) {
  height: 100%;
  padding: 14px 16px;
}

.terminal-surface :deep(.xterm-viewport) {
  scrollbar-width: thin;
}
</style>
