<template>
  <div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2 max-w-sm">
    <TransitionGroup name="toast">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="toast-item px-4 py-3 rounded-lg shadow-lg cursor-pointer flex items-center gap-3"
        :class="toastClass(toast.tone)"
        @click="dismiss(toast.id)"
      >
        <span class="toast-icon" :class="iconClass(toast.tone)">
          <slot name="icon" :tone="toast.tone">
            <span v-if="toast.tone === 'info'">&#9432;</span>
            <span v-else-if="toast.tone === 'success'">&#10003;</span>
            <span v-else-if="toast.tone === 'warning'">&#9888;</span>
            <span v-else-if="toast.tone === 'error'">&#10007;</span>
          </slot>
        </span>
        <span class="toast-message text-sm font-medium">{{ toast.message }}</span>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useToastStore } from '@/stores/toastStore'
import type { ToastTone } from '@/stores/toastStore'

const toastStore = useToastStore()
const toasts = computed(() => toastStore.activeToasts)

const dismiss = (id: number) => {
  toastStore.removeToast(id)
}

const toastClass = (tone: ToastTone): string => {
  const classes: Record<ToastTone, string> = {
    info: 'bg-slate-700 border border-slate-600 text-slate-100',
    success: 'bg-emerald-700 border border-emerald-600 text-emerald-100',
    warning: 'bg-amber-700 border border-amber-600 text-amber-100',
    error: 'bg-red-700 border border-red-600 text-red-100'
  }
  return classes[tone]
}

const iconClass = (tone: ToastTone): string => {
  const classes: Record<ToastTone, string> = {
    info: 'text-slate-300',
    success: 'text-emerald-300',
    warning: 'text-amber-300',
    error: 'text-red-300'
  }
  return classes[tone]
}
</script>

<style scoped>
.toast-enter-active {
  animation: toast-in 0.3s ease-out;
}

.toast-leave-active {
  animation: toast-out 0.2s ease-in;
}

@keyframes toast-in {
  from {
    opacity: 0;
    transform: translateX(100%);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

@keyframes toast-out {
  from {
    opacity: 1;
    transform: translateX(0);
  }
  to {
    opacity: 0;
    transform: translateX(100%);
  }
}

.toast-item {
  backdrop-filter: blur(8px);
}

.toast-icon {
  font-size: 1.25rem;
  line-height: 1;
}

.toast-message {
  flex: 1;
  min-width: 0;
}
</style>