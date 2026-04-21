<template>
  <div class="tasks-page animate-in fade-in zoom-in-95 duration-500">
    <!-- Header -->
    <div class="page-header">
      <div class="header-info">
        <h1 class="page-title">任务管理</h1>
        <p class="page-subtitle">{{ workspaceName }} · {{ tasks.length }} 个任务</p>
      </div>
      <div class="view-toggle">
        <button
          :class="['view-btn', { active: viewMode === 'list' }]"
          @click="viewMode = 'list'"
          title="列表视图"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
          <span>列表</span>
        </button>
        <button
          :class="['view-btn', { active: viewMode === 'kanban' }]"
          @click="viewMode = 'kanban'"
          title="看板视图"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
          </svg>
          <span>看板</span>
        </button>
      </div>
    </div>

    <!-- Filter Tabs (List View Only) -->
    <div v-if="viewMode === 'list'" class="filter-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.value"
        @click="activeTab = tab.value"
        :class="['tab', { active: activeTab === tab.value }]"
      >
        <span class="tab-label">{{ tab.label }}</span>
        <span class="tab-count">{{ getTabCount(tab.value) }}</span>
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>加载任务中...</p>
    </div>

    <!-- Empty State -->
    <div v-else-if="tasks.length === 0" class="empty-state">
      <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
      </svg>
      <h3>暂无任务</h3>
      <p>任务由秘书自动分配</p>
    </div>

    <!-- List View -->
    <div v-else-if="viewMode === 'list'" class="task-list custom-scrollbar">
      <div v-if="filteredTasks.length === 0" class="empty-state">
        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
        <h3>暂无任务</h3>
        <p>此分类下没有任务</p>
      </div>
      <TaskCard
        v-for="task in filteredTasks"
        v-else
        :key="task.id"
        :task="task"
        :clickable="true"
        @start="handleStartTask"
        @complete="handleCompleteTask"
        @fail="handleFailTask"
        @cancel="handleCancelTask"
        @detail="handleOpenDetail"
      />
    </div>

    <!-- Kanban View -->
    <div v-else-if="viewMode === 'kanban'" class="kanban-view-wrapper">
      <TasksKanban
        :tasks="tasks"
        @select-task="selectedTask = $event; showTaskDetailDrawer = true"
        @update-status="handleUpdateTaskStatus"
      />
    </div>

    <!-- Task Action Modal -->
    <TaskActionModal
      v-if="showTaskActionModal"
      :open="showTaskActionModal"
      :action="taskActionType"
      :task-title="selectedTaskTitle"
      @submit="handleTaskActionSubmit"
      @cancel="showTaskActionModal = false"
    />

    <!-- Task Detail Drawer -->
    <TaskDetailDrawer
      :open="showTaskDetailDrawer"
      :task="selectedTask"
      @close="showTaskDetailDrawer = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import { useTaskStore, type Task } from './taskStore'
import { useWorkspaceStore } from '../workspace/workspaceStore'
import { getChatSocket } from '@/shared/socket/chat'
import TaskCard from './TaskCard.vue'
import TaskActionModal from './components/TaskActionModal.vue'
import TaskDetailDrawer from './components/TaskDetailDrawer.vue'
import TasksKanban from './TasksKanban.vue'

const route = useRoute()
const taskStore = useTaskStore()
const workspaceStore = useWorkspaceStore()

const workspaceId = computed(() => route.params.id as string)
const workspaceName = computed(() => workspaceStore.currentWorkspace?.name || '工作区')

const activeTab = ref('all')
const viewMode = ref<'list' | 'kanban'>('list')
const showTaskActionModal = ref(false)
const taskActionType = ref<'complete' | 'fail'>('complete')
const selectedTaskId = ref('')
const selectedTaskTitle = ref('')
const showTaskDetailDrawer = ref(false)
const selectedTask = ref<Task | null>(null)

const tabs = [
  { label: '全部', value: 'all' },
  { label: '待处理', value: 'pending' },
  { label: '已分配', value: 'assigned' },
  { label: '进行中', value: 'in_progress' },
  { label: '已完成', value: 'completed' },
  { label: '失败', value: 'failed' },
  { label: '已取消', value: 'cancelled' }
]

const tasks = computed(() => taskStore.tasks)
const loading = computed(() => taskStore.loading)

const filteredTasks = computed(() => {
  if (activeTab.value === 'all') {
    return tasks.value
  }
  return tasks.value.filter(t => t.status === activeTab.value)
})

function getTabCount(tab: string): number {
  if (tab === 'all') return tasks.value.length
  return tasks.value.filter(t => t.status === tab).length
}

async function handleStartTask(taskId: string) {
  await taskStore.startTask(taskId)
}

function handleCompleteTask(taskId: string) {
  const task = tasks.value.find(t => t.id === taskId)
  if (task) {
    selectedTaskId.value = taskId
    selectedTaskTitle.value = task.title
    taskActionType.value = 'complete'
    showTaskActionModal.value = true
  }
}

function handleFailTask(taskId: string) {
  const task = tasks.value.find(t => t.id === taskId)
  if (task) {
    selectedTaskId.value = taskId
    selectedTaskTitle.value = task.title
    taskActionType.value = 'fail'
    showTaskActionModal.value = true
  }
}

async function handleCancelTask(taskId: string) {
  await taskStore.cancelTask(taskId)
}

function handleOpenDetail(taskId: string) {
  const task = tasks.value.find(t => t.id === taskId)
  if (task) {
    selectedTask.value = task
    showTaskDetailDrawer.value = true
  }
}

async function handleTaskActionSubmit(value: string) {
  if (taskActionType.value === 'complete') {
    await taskStore.completeTask(selectedTaskId.value, value)
  } else {
    await taskStore.failTask(selectedTaskId.value, value)
  }
  showTaskActionModal.value = false
  selectedTaskId.value = ''
  selectedTaskTitle.value = ''
}

async function handleUpdateTaskStatus(taskId: string, newStatus: Task['status']) {
  await taskStore.updateTaskStatus(taskId, newStatus)
}

watch(workspaceId, async (id) => {
  if (id) {
    await taskStore.fetchTasks(id)
  }
}, { immediate: true })

// Listen for task_status WebSocket events
let wsUnsub: (() => void) | null = null
onMounted(() => {
  wsUnsub = getChatSocket().onMessage((msg: any) => {
    if (msg.type === 'task_status') {
      taskStore.handleWsTaskStatus(msg)
    }
  })
})

onBeforeUnmount(() => {
  wsUnsub?.()
})
</script>

<style scoped>
.tasks-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 24px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 16px;
}

.page-title {
  font-size: 32px;
  font-weight: 950;
  color: #0f172a;
  letter-spacing: -0.02em;
}

.page-subtitle {
  font-size: 15px;
  font-weight: 600;
  color: #475569;
  margin-top: 6px;
}

.filter-tabs {
  display: flex;
  gap: 8px;
}

.tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border-radius: 14px;
  font-size: 13px;
  font-weight: 700;
  color: #64748b;
  border: none;
  background: transparent;
  cursor: pointer;
  transition: all 0.2s;
}

.tab:hover {
  background: rgba(15, 23, 42, 0.04);
}

.tab.active {
  background: rgba(99, 102, 241, 0.1);
  color: #4f46e5;
}

.tab-count {
  padding: 2px 8px;
  border-radius: 100px;
  font-size: 11px;
  font-weight: 800;
  background: #f1f5f9;
}

.tab.active .tab-count {
  background: rgba(99, 102, 241, 0.2);
}

.loading-state,
.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  color: #94a3b8;
  padding: 40px;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid #e2e8f0;
  border-top-color: #4f46e5;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-state svg {
  width: 64px;
  height: 64px;
  margin-bottom: 16px;
  opacity: 0.4;
}

.empty-state h3 {
  font-size: 18px;
  font-weight: 800;
  color: #475569;
  margin-bottom: 4px;
}

.empty-state p {
  font-size: 14px;
  color: #94a3b8;
}

.task-list {
  flex: 1;
  overflow-y: auto;
  padding-bottom: 40px;
}

.task-list > * + * {
  margin-top: 16px;
}

.view-toggle {
  display: flex;
  gap: 8px;
  background: #f1f5f9;
  padding: 4px;
  border-radius: 10px;
}

.view-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: #64748b;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
}

.view-btn:hover {
  background: rgba(15, 23, 42, 0.05);
  color: #475569;
}

.view-btn.active {
  background: white;
  color: #4f46e5;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.view-btn svg {
  width: 16px;
  height: 16px;
}

.kanban-view-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
</style>