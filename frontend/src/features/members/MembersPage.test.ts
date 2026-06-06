import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import MembersPage from './MembersPage.vue'

vi.mock('@/features/workspace/workspaceStore', () => ({
  useWorkspaceStore: () => ({
    currentWorkspace: { id: 'ws-1', name: 'Demo Workspace' },
  }),
}))

const clientMock = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}))

vi.mock('@/shared/api/client', () => ({
  default: clientMock,
}))

vi.mock('@/shared/notifyError', () => ({
  notifyUserError: vi.fn(),
}))

function mountPage() {
  return mount(MembersPage, {
    global: {
      stubs: {
        AddMemberModal: true,
        EditMemberModal: true,
        ConfirmModal: true,
      },
    },
  })
}

describe('MembersPage agent sessions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    clientMock.get.mockResolvedValue({
      data: [
        {
          id: 'assistant-1',
          workspaceId: 'ws-1',
          name: 'Claude',
          roleType: 'assistant',
          acpEnabled: true,
          acpCommand: 'claude',
          acpArgs: [],
        },
      ],
    })
  })

  it('starts a configured assistant terminal session from the member card', async () => {
    clientMock.post.mockResolvedValue({
      data: {
        sessionId: 'session-123456',
        status: 'running',
      },
    })

    const wrapper = mountPage()
    await flushPromises()

    await wrapper.get('[data-test="start-agent-session"]').trigger('click')
    await flushPromises()

    expect(clientMock.post).toHaveBeenCalledWith(
      '/workspaces/ws-1/members/assistant-1/terminal-session',
      {},
      { skipErrorToast: true }
    )
    expect(wrapper.text()).toContain('会话: session-1')
  })
})
