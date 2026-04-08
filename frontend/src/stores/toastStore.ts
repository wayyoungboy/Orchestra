import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export type ToastTone = 'info' | 'success' | 'warning' | 'error'

export interface Toast {
  id: number
  message: string
  tone: ToastTone
  duration?: number
}

export const useToastStore = defineStore('toast', () => {
  const toasts = ref<Toast[]>([])
  let nextId = 1

  const activeToasts = computed(() => toasts.value)

  function addToast(message: string, tone: ToastTone = 'info', duration: number = 5000): number {
    const id = nextId++
    toasts.value.push({ id, message, tone, duration })

    if (duration > 0) {
      setTimeout(() => {
        removeToast(id)
      }, duration)
    }

    return id
  }

  function removeToast(id: number) {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  function info(message: string, duration?: number) {
    return addToast(message, 'info', duration)
  }

  function success(message: string, duration?: number) {
    return addToast(message, 'success', duration)
  }

  function warning(message: string, duration?: number) {
    return addToast(message, 'warning', duration)
  }

  function error(message: string, duration?: number) {
    return addToast(message, 'error', duration ?? 8000)
  }

  return {
    toasts,
    activeToasts,
    addToast,
    removeToast,
    info,
    success,
    warning,
    error
  }
})
