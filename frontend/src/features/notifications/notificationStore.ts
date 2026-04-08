import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export type NotificationType = 'success' | 'error' | 'warning' | 'info'

export interface Notification {
  id: string
  type: NotificationType
  title: string
  message?: string
  duration?: number
  createdAt: Date
}

export const useNotificationStore = defineStore('notification', () => {
  const notifications = ref<Notification[]>([])
  const maxNotifications = 5

  const activeNotifications = computed(() =>
    notifications.value.slice(-maxNotifications)
  )

  function addNotification(
    type: NotificationType,
    title: string,
    message?: string,
    duration: number = 5000
  ): Notification {
    const notification: Notification = {
      id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      type,
      title,
      message,
      duration,
      createdAt: new Date()
    }

    notifications.value.push(notification)

    // Auto-remove after duration
    if (duration > 0) {
      setTimeout(() => {
        removeNotification(notification.id)
      }, duration)
    }

    return notification
  }

  function removeNotification(id: string) {
    notifications.value = notifications.value.filter(n => n.id !== id)
  }

  function clearAll() {
    notifications.value = []
  }

  // Convenience methods
  function success(title: string, message?: string, duration?: number) {
    return addNotification('success', title, message, duration)
  }

  function error(title: string, message?: string, duration?: number) {
    return addNotification('error', title, message, duration ?? 8000)
  }

  function warning(title: string, message?: string, duration?: number) {
    return addNotification('warning', title, message, duration)
  }

  function info(title: string, message?: string, duration?: number) {
    return addNotification('info', title, message, duration)
  }

  return {
    notifications,
    activeNotifications,
    addNotification,
    removeNotification,
    clearAll,
    success,
    error,
    warning,
    info
  }
})