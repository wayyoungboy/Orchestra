<template>
  <div class="task-card" :class="statusClass">
    <div class="task-header">
      <div class="status-badge" :class="task.status">
        {{ statusLabel }}
      </div>
      <div class="task-actions">
        <button v-if="task.status === 'pending' || task.status === 'assigned'" @click="$emit('start', task.id)" class="action-btn start">
          开始
        </button>
        <button v-if="task.status === 'in_progress'" @click="$emit('complete', task.id)" class="action-btn complete">
          完成
        </button>
        <button v-if="task.status === 'in_progress'" @click="$emit('fail', task.id)" class="action-btn fail">
          失败
        </button>
        <button v-if="task.status === 'pending' || task.status === 'assigned' || task.status === 'in_progress'" @click="$emit('cancel', task.id)" class="action-btn cancel">
          取消
        </button>
      </div>
    </div>

    <h3 class="task-title">{{ task.title }}</h3>

    <p v-if="task.description" class="task-description">
      {{ task.description }}
    </p>

    <div class="task-meta">
      <div class="meta-item">
        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
        </svg>
        <span>{{ task.assignee?.name || '未分配' }}</span>
      </div>
      <div class="meta-item">
        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>{{ formatDate(task.createdAt) }}</span>
      </div>
    </div>

    <div v-if="task.resultSummary" class="task-result">
      <span class="result-label">结果：</span>
      {{ task.resultSummary }}
    </div>

    <div v-if="task.errorMessage" class="task-error">
      <span class="error-label">错误：</span>
      {{ task.errorMessage }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Task } from './taskStore'

const props = defineProps<{
  task: Task
}>()

defineEmits<{
  (e: 'start', taskId: string): void
  (e: 'complete', taskId: string): void
  (e: 'fail', taskId: string): void
  (e: 'cancel', taskId: string): void
}>()

const statusLabel = computed(() => {
  const labels: Record<string, string> = {
    pending: '待处理',
    assigned: '已分配',
    in_progress: '进行中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消'
  }
  return labels[props.task.status] || props.task.status
})

const statusClass = computed(() => `status-${props.task.status}`)

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  return date.toLocaleDateString()
}
</script>

<style scoped>
.task-card {
  @apply bg-white rounded-xl p-4 border border-slate-100 shadow-sm hover:shadow-md transition-shadow;
}

.task-card.status-pending {
  @apply border-l-4 border-l-slate-400;
}

.task-card.status-assigned {
  @apply border-l-4 border-l-yellow-500;
}

.task-card.status-in_progress {
  @apply border-l-4 border-l-blue-500;
}

.task-card.status-completed {
  @apply border-l-4 border-l-green-500;
}

.task-card.status-failed {
  @apply border-l-4 border-l-red-500;
}

.task-card.status-cancelled {
  @apply border-l-4 border-l-gray-400;
}

.task-header {
  @apply flex items-center justify-between mb-3;
}

.status-badge {
  @apply px-2.5 py-1 rounded-full text-xs font-bold uppercase tracking-wide;
}

.status-badge.pending {
  @apply bg-slate-100 text-slate-600;
}

.status-badge.assigned {
  @apply bg-yellow-100 text-yellow-700;
}

.status-badge.in_progress {
  @apply bg-blue-100 text-blue-700;
}

.status-badge.completed {
  @apply bg-green-100 text-green-700;
}

.status-badge.failed {
  @apply bg-red-100 text-red-700;
}

.task-actions {
  @apply flex gap-2;
}

.action-btn {
  @apply px-3 py-1.5 rounded-lg text-xs font-medium transition-colors;
}

.action-btn.start {
  @apply bg-blue-50 text-blue-600 hover:bg-blue-100;
}

.action-btn.complete {
  @apply bg-green-50 text-green-600 hover:bg-green-100;
}

.action-btn.fail {
  @apply bg-red-50 text-red-600 hover:bg-red-100;
}

.action-btn.cancel {
  @apply bg-gray-100 text-gray-500 hover:bg-gray-200;
}

.task-title {
  @apply text-base font-semibold text-slate-900 mb-2;
}

.task-description {
  @apply text-sm text-slate-500 mb-3 line-clamp-2;
}

.task-meta {
  @apply flex items-center gap-4 text-xs text-slate-400;
}

.meta-item {
  @apply flex items-center gap-1;
}

.meta-item svg {
  @apply w-3.5 h-3.5;
}

.task-result {
  @apply mt-3 p-2 bg-green-50 rounded-lg text-sm text-green-700;
}

.task-error {
  @apply mt-3 p-2 bg-red-50 rounded-lg text-sm text-red-700;
}

.result-label, .error-label {
  @apply font-medium;
}
</style>