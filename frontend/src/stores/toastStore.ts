import { ref, computed } from 'vue'
import { acceptHMRUpdate, defineStore } from 'pinia'

export type ToastTone = 'info' | 'success' | 'warning' | 'error'

export interface ToastOptions {
  tone?: ToastTone
  duration?: number
}

export interface Toast {
  id: number
  message: string
  tone: ToastTone
  duration: number
  createdAt: number
}

const DEFAULT_DURATION = 3000
const DEFAULT_TONE: ToastTone = 'info'

let toastIdCounter = 0

export const useToastStore = defineStore('toast', () => {
  const toasts = ref<Toast[]>([])

  const pushToast = (message: string, options?: ToastOptions): number => {
    const id = ++toastIdCounter
    const tone = options?.tone ?? DEFAULT_TONE
    const duration = options?.duration ?? DEFAULT_DURATION

    const toast: Toast = {
      id,
      message,
      tone,
      duration,
      createdAt: Date.now()
    }

    toasts.value.push(toast)

    // Auto-dismiss after duration
    if (duration > 0) {
      setTimeout(() => {
        removeToast(id)
      }, duration)
    }

    return id
  }

  const removeToast = (id: number): boolean => {
    const index = toasts.value.findIndex((t) => t.id === id)
    if (index !== -1) {
      toasts.value.splice(index, 1)
      return true
    }
    return false
  }

  const clearAll = () => {
    toasts.value = []
  }

  const activeToasts = computed(() => toasts.value)

  return {
    toasts,
    activeToasts,
    pushToast,
    removeToast,
    clearAll
  }
})

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useToastStore, import.meta.hot))
}