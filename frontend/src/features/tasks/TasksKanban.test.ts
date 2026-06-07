import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import TasksKanban from './TasksKanban.vue'
import type { Task } from './taskStore'

function makeTask(overrides: Partial<Task> = {}): Task {
  return {
    id: 'task-1',
    workspaceId: 'workspace-1',
    conversationId: 'conversation-1',
    secretaryId: 'secretary-1',
    assigneeId: '',
    title: 'Plan work',
    description: '',
    status: 'pending',
    createdAt: '2026-06-07T00:00:00Z',
    updatedAt: '2026-06-07T00:00:00Z',
    ...overrides
  }
}

function makeDragEvent(type: string): DragEvent {
  const event = new Event(type, {
    bubbles: true,
    cancelable: true,
  }) as DragEvent
  Object.defineProperty(event, 'dataTransfer', {
    value: {
      effectAllowed: '',
      dropEffect: '',
      setData: () => undefined,
    },
  })
  return event
}

describe('TasksKanban transitions', () => {
  it('emits only backend-valid lifecycle transitions', async () => {
    const wrapper = mount(TasksKanban, {
      props: {
        tasks: [makeTask()],
      },
    })

    const card = wrapper.get('.kanban-card')
    const completedColumn = wrapper.findAll('.kanban-column')[3]
    const assignedColumn = wrapper.findAll('.kanban-column')[1]

    await card.element.dispatchEvent(makeDragEvent('dragstart'))
    await completedColumn.element.dispatchEvent(makeDragEvent('drop'))

    expect(wrapper.emitted('update-status')).toBeUndefined()

    await card.element.dispatchEvent(makeDragEvent('dragstart'))
    await assignedColumn.element.dispatchEvent(makeDragEvent('drop'))

    expect(wrapper.emitted('update-status')).toEqual([
      ['task-1', 'assigned'],
    ])
  })
})
