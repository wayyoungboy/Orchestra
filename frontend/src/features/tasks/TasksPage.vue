<template>
  <div class="tasks-page animate-in fade-in zoom-in-95 duration-500">
    <!-- Header -->
    <div class="page-header">
      <div class="header-info">
        <h1 class="page-title">任务管理</h1>
        <p class="page-subtitle">{{ workspaceName }} · {{ tasks.length }} 个任务</p>
      </div>
    </div>

    <!-- Filter Tabs -->
    <div class="filter-tabs">
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
    <div v-else-if="filteredTasks.length === 0" class="empty-state">
      <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
      </svg>
      <h3>暂无任务</h3>
      <p>{{ activeTab === 'all' ? '任务由秘书自动分配' : '此分类下没有任务' }}</p>
    </div>

    <!-- Task List -->
    <div v-else class="task-list custom-scrollbar">
      <TaskCard
        v-for="task in filteredTasks"
        :key="task.id"
        :task="task"
        @start="handleStartTask"
        @complete="handleCompleteTask"
        @fail="handleFailTask"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useTaskStore } from './taskStore'
import { useWorkspaceStore } from '../workspace/workspaceStore'
import TaskCard from './TaskCard.vue'

const route = useRoute()
const taskStore = useTaskStore()
const workspaceStore = useWorkspaceStore()

const workspaceId = computed(() => route.params.id as string)
const workspaceName = computed(() => workspaceStore.currentWorkspace?.name || '工作区')

const activeTab = ref('all')

const tabs = [
  { label: '全部', value: 'all' },
  { label: '待处理', value: 'pending' },
  { label: '进行中', value: 'in_progress' },
  { label: '已完成', value: 'completed' },
  { label: '失败', value: 'failed' }
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

async function handleCompleteTask(taskId: string) {
  const result = prompt('请输入任务结果摘要：')
  if (result) {
    await taskStore.completeTask(taskId, result)
  }
}

async function handleFailTask(taskId: string) {
  const reason = prompt('请输入失败原因：')
  if (reason) {
    await taskStore.failTask(taskId, reason)
  }
}

watch(workspaceId, async (id) => {
  if (id) {
    await taskStore.fetchTasks(id)
  }
}, { immediate: true })
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
</style>