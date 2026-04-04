import { ref, computed } from 'vue'
import { acceptHMRUpdate, defineStore } from 'pinia'

const MAX_LINES = 200

/**
 * 前端诊断日志（REQ-404）：缓冲最近若干条 API / 运行时信息，供 Settings 查看。
 */
export const useDiagnosticsStore = defineStore('diagnostics', () => {
  const lines = ref<string[]>([])

  const recentLines = computed(() => lines.value)

  const log = (message: string) => {
    const ts = new Date().toISOString()
    const row = `[${ts}] ${message}`
    lines.value = [...lines.value.slice(-(MAX_LINES - 1)), row]
  }

  const clear = () => {
    lines.value = []
  }

  return {
    recentLines,
    log,
    clear
  }
})

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useDiagnosticsStore, import.meta.hot))
}
