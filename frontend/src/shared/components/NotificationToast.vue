<template>
  <Teleport to="body">
    <div class="notification-container">
      <TransitionGroup name="notification">
        <div
          v-for="notification in notifications"
          :key="notification.id"
          class="notification-toast"
          :class="notification.type"
        >
          <div class="toast-icon">
            <!-- Success -->
            <svg v-if="notification.type === 'success'" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
            <!-- Error -->
            <svg v-else-if="notification.type === 'error'" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
            <!-- Warning -->
            <svg v-else-if="notification.type === 'warning'" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <!-- Info -->
            <svg v-else fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div class="toast-content">
            <p class="toast-title">{{ notification.title }}</p>
            <p v-if="notification.message" class="toast-message">
              {{ notification.message }}
            </p>
          </div>
          <button
            @click="removeNotification(notification.id)"
            class="toast-close"
          >
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useNotificationStore } from '@/features/notifications/notificationStore'

const notificationStore = useNotificationStore()
const { notifications } = storeToRefs(notificationStore)
const { removeNotification } = notificationStore
</script>

<style scoped>
.notification-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 12px;
  pointer-events: none;
}

.notification-toast {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-width: 320px;
  max-width: 420px;
  padding: 16px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1), 0 0 0 1px rgba(0, 0, 0, 0.02);
  pointer-events: auto;
  transition: all 0.3s ease;
}

.notification-toast.success {
  border-left: 4px solid #10b981;
}

.notification-toast.error {
  border-left: 4px solid #ef4444;
}

.notification-toast.warning {
  border-left: 4px solid #f59e0b;
}

.notification-toast.info {
  border-left: 4px solid #6366f1;
}

.toast-icon {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.toast-icon svg {
  width: 20px;
  height: 20px;
}

.success .toast-icon { color: #10b981; }
.error .toast-icon { color: #ef4444; }
.warning .toast-icon { color: #f59e0b; }
.info .toast-icon { color: #6366f1; }

.toast-content {
  flex: 1;
  min-width: 0;
}

.toast-title {
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
  margin: 0;
}

.toast-message {
  font-size: 13px;
  color: #64748b;
  margin: 4px 0 0;
  line-height: 1.4;
}

.toast-close {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s;
}

.toast-close:hover {
  background: #f1f5f9;
  color: #475569;
}

.toast-close svg {
  width: 16px;
  height: 16px;
}

/* Animations */
.notification-enter-active {
  animation: slide-in 0.3s ease-out;
}

.notification-leave-active {
  animation: slide-out 0.2s ease-in forwards;
}

@keyframes slide-in {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

@keyframes slide-out {
  from {
    transform: translateX(0);
    opacity: 1;
  }
  to {
    transform: translateX(100%);
    opacity: 0;
  }
}

@media (max-width: 480px) {
  .notification-container {
    left: 16px;
    right: 16px;
    top: 16px;
  }

  .notification-toast {
    min-width: auto;
    max-width: none;
  }
}
</style>