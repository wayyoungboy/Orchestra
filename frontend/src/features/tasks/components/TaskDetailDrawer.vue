<template>
  <Transition name="slide-right">
    <div v-if="open" class="task-detail-overlay" @click="handleClose" @keydown.esc="handleClose">
      <div class="task-detail-drawer animate-in fade-in duration-300" @click.stop>
        <!-- Header -->
        <div class="drawer-header">
          <h2 class="drawer-title">任务详情</h2>
          <button @click="handleClose" class="close-btn">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Content -->
        <div class="drawer-content custom-scrollbar">
          <div v-if="task" class="task-details">
            <!-- Title Section -->
            <div class="detail-section">
              <h3 class="detail-label">标题</h3>
              <p class="detail-title">{{ task.title }}</p>
            </div>

            <!-- Status Badge -->
            <div class="detail-section">
              <h3 class="detail-label">状态</h3>
              <span :class="['status-badge', `status-${task.status}`]">
                {{ getStatusLabel(task.status) }}
              </span>
            </div>

            <!-- Description -->
            <div v-if="task.description" class="detail-section">
              <h3 class="detail-label">描述</h3>
              <p class="detail-text">{{ task.description }}</p>
            </div>

            <!-- Assignee -->
            <div class="detail-section">
              <h3 class="detail-label">指派给</h3>
              <div class="assignee-badge">
                <div class="assignee-avatar">
                  {{ (task.assignee?.name || '未分配').charAt(0).toUpperCase() }}
                </div>
                <span class="assignee-name">{{ task.assignee?.name || '未分配' }}</span>
              </div>
            </div>

            <!-- Created At -->
            <div class="detail-section">
              <h3 class="detail-label">创建于</h3>
              <p class="detail-time">{{ formatDate(task.createdAt) }}</p>
            </div>

            <!-- Result Summary -->
            <div v-if="task.resultSummary" class="detail-section">
              <h3 class="detail-label">结果摘要</h3>
              <div class="result-box">
                {{ task.resultSummary }}
              </div>
            </div>

            <!-- Error Message -->
            <div v-if="task.errorMessage" class="detail-section">
              <h3 class="detail-label">失败原因</h3>
              <div class="error-box">
                {{ task.errorMessage }}
              </div>
            </div>

            <!-- Completed At -->
            <div v-if="task.completedAt" class="detail-section">
              <h3 class="detail-label">完成于</h3>
              <p class="detail-time">{{ formatDate(task.completedAt) }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import type { Task } from '../taskStore'

defineProps<{
  open: boolean
  task?: Task | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

function handleClose() {
  emit('close')
}

function getStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    pending: '待处理',
    in_progress: '进行中',
    completed: '已完成',
    failed: '失败'
  }
  return labels[status] || status
}

function formatDate(dateStr: string): string {
  try {
    const date = new Date(dateStr)
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch {
    return dateStr
  }
}
</script>

<style scoped>
.task-detail-overlay {
  position: fixed;
  inset: 0;
  z-index: 40;
  background: rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(4px);
}

.task-detail-drawer {
  position: fixed;
  right: 0;
  top: 0;
  bottom: 0;
  width: 380px;
  background: white;
  box-shadow: -20px 0 50px rgba(0, 0, 0, 0.15);
  display: flex;
  flex-direction: column;
  z-index: 50;
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px;
  border-bottom: 1px solid #f1f5f9;
  flex-shrink: 0;
}

.drawer-title {
  font-size: 18px;
  font-weight: 800;
  color: #0f172a;
}

.close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #f1f5f9;
  color: #475569;
}

.drawer-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.task-details {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.detail-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-label {
  font-size: 11px;
  font-weight: 700;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.detail-title {
  font-size: 16px;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.5;
}

.detail-text {
  font-size: 14px;
  color: #475569;
  line-height: 1.6;
  white-space: pre-wrap;
}

.detail-time {
  font-size: 14px;
  color: #64748b;
  font-family: monospace;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 6px 12px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 700;
  width: fit-content;
}

.status-pending {
  background: #f1f5f9;
  color: #64748b;
}

.status-in_progress {
  background: #dbeafe;
  color: #0369a1;
}

.status-completed {
  background: #dcfce7;
  color: #166534;
}

.status-failed {
  background: #fee2e2;
  color: #991b1b;
}

.assignee-badge {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: #f8fafc;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  width: fit-content;
}

.assignee-avatar {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: rgba(99, 102, 241, 0.1);
  color: #4f46e5;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
}

.assignee-name {
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
}

.result-box,
.error-box {
  padding: 12px;
  border-radius: 10px;
  font-size: 14px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

.result-box {
  background: #dcfce7;
  color: #166534;
  border: 1px solid #bbf7d0;
}

.error-box {
  background: #fee2e2;
  color: #991b1b;
  border: 1px solid #fecaca;
}

@keyframes slide-right {
  from {
    transform: translateX(100%);
  }
  to {
    transform: translateX(0);
  }
}

.slide-right-enter-active,
.slide-right-leave-active {
  transition: all 0.3s ease;
}

.slide-right-enter-from {
  transform: translateX(100%);
}

.slide-right-leave-to {
  transform: translateX(100%);
}
</style>
