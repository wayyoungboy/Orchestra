import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import client from '@/shared/api/client'

export interface Task {
  id: string
  workspaceId: string
  conversationId: string
  secretaryId: string
  assigneeId: string
  title: string
  description: string
  status: 'pending' | 'in_progress' | 'completed' | 'failed'
  resultSummary?: string
  errorMessage?: string
  createdAt: string
  updatedAt: string
  completedAt?: string

  // Related entities
  assignee?: {
    id: string
    name: string
    roleType: string
  }
  secretary?: {
    id: string
    name: string
    roleType: string
  }
}

export interface Workload {
  memberId: string
  memberName: string
  pendingTasks: number
  inProgressTasks: number
  completedTasks: number
  totalTasks: number
}

export interface TaskCreate {
  workspaceId: string
  conversationId: string
  secretaryId: string
  title: string
  description?: string
  assigneeId?: string
}

export const useTaskStore = defineStore('task', () => {
  const tasks = ref<Task[]>([])
  const workloads = ref<Workload[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const currentWorkspaceId = ref<string | null>(null)

  const pendingTasks = computed(() =>
    tasks.value.filter(t => t.status === 'pending')
  )

  const inProgressTasks = computed(() =>
    tasks.value.filter(t => t.status === 'in_progress')
  )

  const completedTasks = computed(() =>
    tasks.value.filter(t => t.status === 'completed')
  )

  const failedTasks = computed(() =>
    tasks.value.filter(t => t.status === 'failed')
  )

  async function fetchTasks(workspaceId: string) {
    loading.value = true
    error.value = null
    currentWorkspaceId.value = workspaceId

    try {
      const response = await client.get(`/workspaces/${workspaceId}/tasks`)
      tasks.value = response.data?.tasks || []
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch tasks'
    } finally {
      loading.value = false
    }
  }

  async function fetchMyTasks(workspaceId: string) {
    loading.value = true
    error.value = null

    try {
      const response = await client.get(`/workspaces/${workspaceId}/tasks/my-tasks`)
      tasks.value = response.data?.tasks || []
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch my tasks'
    } finally {
      loading.value = false
    }
  }

  async function fetchWorkloads(workspaceId: string) {
    try {
      const response = await client.get(`/internal/workloads/list?workspaceId=${workspaceId}`)
      workloads.value = response.data?.workloads || []
    } catch (e: any) {
      console.error('Failed to fetch workloads:', e)
    }
  }

  async function createTask(task: TaskCreate) {
    loading.value = true
    error.value = null

    try {
      const response = await client.post('/internal/tasks/create', task)
      const newTask = response.data?.task
      if (newTask) {
        tasks.value.push(newTask)
      }
      return newTask
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to create task'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function startTask(taskId: string) {
    try {
      await client.post('/internal/tasks/start', { taskId })
      const task = tasks.value.find(t => t.id === taskId)
      if (task) {
        task.status = 'in_progress'
      }
    } catch (e: any) {
      console.error('Failed to start task:', e)
      throw e
    }
  }

  async function completeTask(taskId: string, resultSummary?: string) {
    try {
      await client.post('/internal/tasks/complete', { taskId, resultSummary })
      const task = tasks.value.find(t => t.id === taskId)
      if (task) {
        task.status = 'completed'
        task.resultSummary = resultSummary
      }
    } catch (e: any) {
      console.error('Failed to complete task:', e)
      throw e
    }
  }

  async function failTask(taskId: string, errorMessage: string) {
    try {
      await client.post('/internal/tasks/fail', { taskId, errorMessage })
      const task = tasks.value.find(t => t.id === taskId)
      if (task) {
        task.status = 'failed'
        task.errorMessage = errorMessage
      }
    } catch (e: any) {
      console.error('Failed to fail task:', e)
      throw e
    }
  }

  function addTask(task: Task) {
    tasks.value.push(task)
  }

  function updateTask(updatedTask: Task) {
    const index = tasks.value.findIndex(t => t.id === updatedTask.id)
    if (index !== -1) {
      tasks.value[index] = updatedTask
    }
  }

  function removeTask(taskId: string) {
    tasks.value = tasks.value.filter(t => t.id !== taskId)
  }

  return {
    tasks,
    workloads,
    loading,
    error,
    currentWorkspaceId,
    pendingTasks,
    inProgressTasks,
    completedTasks,
    failedTasks,
    fetchTasks,
    fetchMyTasks,
    fetchWorkloads,
    createTask,
    startTask,
    completeTask,
    failTask,
    addTask,
    updateTask,
    removeTask
  }
})