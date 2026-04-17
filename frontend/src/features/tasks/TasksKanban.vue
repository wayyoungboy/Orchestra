<template>
  <div class="kanban-container">
    <!-- Kanban Columns -->
    <div class="kanban-board">
      <div
        v-for="column in columns"
        :key="column.status"
        class="kanban-column"
        @dragover.prevent="handleDragOver"
        @drop="handleDrop($event, column.status)"
      >
        <!-- Column Header -->
        <div class="column-header">
          <h3 class="column-title">{{ column.label }}</h3>
          <span class="column-count">{{ getTasksInColumn(column.status).length }}</span>
        </div>

        <!-- Task Cards -->
        <div class="column-content custom-scrollbar">
          <div
            v-for="task in getTasksInColumn(column.status)"
            :key="task.id"
            class="kanban-card"
            draggable="true"
            @dragstart="handleDragStart($event, task)"
            @click="$emit('select-task', task)"
          >
            <!-- Card Header -->
            <div class="card-header">
              <span :class="['status-dot', `status-${task.status}`]"></span>
              <span class="card-title-text">{{ task.title }}</span>
            </div>

            <!-- Assignee Badge -->
            <div v-if="task.assignee" class="card-assignee">
              <span class="assignee-dot"></span>
              <span class="assignee-text">{{ task.assignee.name }}</span>
            </div>

            <!-- Card Footer -->
            <div class="card-footer">
              <span class="card-date">{{ formatDate(task.createdAt) }}</span>
            </div>
          </div>

          <!-- Empty State -->
          <div v-if="getTasksInColumn(column.status).length === 0" class="empty-column">
            <p>暂无任务</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { Task } from './taskStore'

const props = defineProps<{
  tasks: Task[]
}>()

const emit = defineEmits<{
  (e: 'select-task', task: Task): void
  (e: 'update-status', taskId: string, newStatus: Task['status']): void
}>()

const draggedTask = ref<Task | null>(null)

interface KanbanColumn {
  status: Task['status']
  label: string
}

const columns: KanbanColumn[] = [
  { status: 'pending', label: '待处理' },
  { status: 'assigned', label: '已分配' },
  { status: 'in_progress', label: '进行中' },
  { status: 'completed', label: '已完成' },
  { status: 'failed', label: '失败' },
  { status: 'cancelled', label: '已取消' }
]

function getTasksInColumn(status: Task['status']): Task[] {
  return props.tasks.filter(t => t.status === status)
}

function handleDragStart(event: DragEvent, task: Task) {
  draggedTask.value = task
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', task.id)
  }
}

function handleDragOver(event: DragEvent) {
  event.preventDefault()
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'move'
  }
}

function handleDrop(event: DragEvent, targetStatus: Task['status']) {
  event.preventDefault()

  if (!draggedTask.value) return

  // Allow only valid status transitions
  const validTransitions: Record<Task['status'], Task['status'][]> = {
    pending: ['assigned', 'in_progress', 'completed', 'failed', 'cancelled'],
    assigned: ['in_progress', 'cancelled'],
    in_progress: ['completed', 'failed', 'cancelled'],
    completed: ['in_progress'],
    failed: ['in_progress'],
    cancelled: []
  }

  const allowedTransitions = validTransitions[draggedTask.value.status] || []
  if (!allowedTransitions.includes(targetStatus)) {
    draggedTask.value = null
    return
  }

  if (draggedTask.value.status !== targetStatus) {
    emit('update-status', draggedTask.value.id, targetStatus)
  }

  draggedTask.value = null
}

function formatDate(dateStr: string): string {
  try {
    const date = new Date(dateStr)
    const now = new Date()
    const diff = now.getTime() - date.getTime()

    if (diff < 60000) return '刚刚'
    if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
    if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
    return date.toLocaleDateString()
  } catch {
    return dateStr
  }
}
</script>

<style scoped>
.kanban-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 24px;
}

.kanban-board {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 20px;
  flex: 1;
  min-height: 0;
}

.kanban-column {
  display: flex;
  flex-direction: column;
  background: #f8fafc;
  border-radius: 16px;
  border: 1px solid #e2e8f0;
  min-height: 0;
  overflow: hidden;
  transition: all 0.2s;
}

.kanban-column:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.column-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid #e2e8f0;
  flex-shrink: 0;
}

.column-title {
  font-size: 15px;
  font-weight: 700;
  color: #0f172a;
}

.column-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: white;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 700;
  color: #64748b;
}

.column-content {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.kanban-card {
  background: white;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  padding: 12px;
  cursor: grab;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.kanban-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  border-color: #cbd5e1;
  transform: translateY(-2px);
}

.kanban-card:active {
  cursor: grabbing;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-pending {
  background: #cbd5e1;
}

.status-in_progress {
  background: #3b82f6;
}

.status-completed {
  background: #10b981;
}

.status-failed {
  background: #ef4444;
}

.status-cancelled {
  background: #9ca3af;
}

.card-title-text {
  font-size: 13px;
  font-weight: 600;
  color: #0f172a;
  line-height: 1.4;
  word-break: break-word;
}

.card-assignee {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
  font-size: 12px;
  color: #64748b;
}

.assignee-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #4f46e5;
  flex-shrink: 0;
}

.assignee-text {
  flex: 1;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 11px;
  color: #94a3b8;
}

.card-date {
  font-weight: 500;
}

.empty-column {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  color: #cbd5e1;
  font-size: 13px;
}
</style>
