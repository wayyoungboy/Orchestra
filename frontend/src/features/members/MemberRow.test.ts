import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import MemberRow from './MemberRow.vue'
import type { Member } from '@/shared/types/member'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, string>) => ({
      'memberRow.actionsLabel': `Actions for ${params?.name ?? 'member'}`,
      'memberRow.sendMessage': 'Send message',
      'memberRow.remove': 'Remove',
      'memberRow.statusHeading': 'Status',
      'memberRow.statusOnline': 'Online',
      'memberRow.statusWorking': 'Working',
      'memberRow.statusDnd': 'Do not disturb',
      'memberRow.statusOffline': 'Offline',
      'members.roleAssistant': 'Assistant',
      'members.unnamedMember': 'Unnamed',
    }[key] ?? key),
  }),
}))

vi.mock('@/features/chat/chatStore', () => ({
  useChatStore: () => ({ agentStatuses: {} }),
}))

const assistantMember: Member = {
  id: 'assistant-1',
  workspaceId: 'ws-1',
  name: 'Claude',
  roleType: 'assistant',
  status: 'online',
  createdAt: '2026-06-07T00:00:00Z',
}

describe('MemberRow accessibility', () => {
  it('gives the icon-only member actions button an accessible name', () => {
    const wrapper = mount(MemberRow, {
      props: {
        member: assistantMember,
      },
    })

    const actionButton = wrapper.get('button.more-btn')

    expect(actionButton.attributes('aria-label')).toBe('Actions for Claude')
    expect(actionButton.attributes('title')).toBe('Actions for Claude')
  })
})
