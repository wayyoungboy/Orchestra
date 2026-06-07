import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import TaskDetailDrawer from './TaskDetailDrawer.vue'
import type { Task } from '../taskStore'

function makeTask(status: Task['status']): Task {
  return {
    id: `task-${status}`,
    workspaceId: 'workspace-1',
    conversationId: 'conversation-1',
    secretaryId: 'secretary-1',
    assigneeId: 'assistant-1',
    title: `Task ${status}`,
    description: '',
    status,
    createdAt: '2026-06-07T00:00:00Z',
    updatedAt: '2026-06-07T00:00:00Z',
  }
}

describe('TaskDetailDrawer', () => {
  it('shows localized labels for every task lifecycle status', async () => {
    const wrapper = mount(TaskDetailDrawer, {
      props: {
        open: true,
        task: makeTask('assigned'),
      },
    })

    expect(wrapper.get('.status-badge').text()).toBe('已分配')

    await wrapper.setProps({ task: makeTask('cancelled') })

    expect(wrapper.get('.status-badge').text()).toBe('已取消')
  })
})
