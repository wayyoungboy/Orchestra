import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useTaskStore, type Task } from './taskStore'

function makeTask(overrides: Partial<Task> = {}): Task {
  return {
    id: 'task-1',
    workspaceId: 'workspace-1',
    conversationId: 'conversation-1',
    secretaryId: 'secretary-1',
    assigneeId: 'assistant-old',
    title: 'Original title',
    description: '',
    status: 'pending',
    createdAt: '2026-06-07T00:00:00Z',
    updatedAt: '2026-06-07T00:00:00Z',
    ...overrides
  }
}

describe('taskStore websocket task status handling', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('updates changed task fields from task_status events', () => {
    const store = useTaskStore()
    store.addTask(makeTask())

    store.handleWsTaskStatus({
      taskId: 'task-1',
      status: 'assigned',
      assigneeId: 'assistant-new',
      title: 'Updated title'
    })

    expect(store.tasks[0]).toMatchObject({
      status: 'assigned',
      assigneeId: 'assistant-new',
      title: 'Updated title'
    })
  })

  it('updates task result and error details from task_status events', () => {
    const store = useTaskStore()
    store.addTask(makeTask({ status: 'in_progress' }))

    store.handleWsTaskStatus({
      taskId: 'task-1',
      status: 'completed',
      resultSummary: 'Completed from another browser',
      errorMessage: ''
    })

    expect(store.tasks[0]).toMatchObject({
      status: 'completed',
      resultSummary: 'Completed from another browser',
      errorMessage: ''
    })
  })
})
