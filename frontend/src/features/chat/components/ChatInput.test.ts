import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ChatInput from './ChatInput.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => ({
      'chat.inputHint': 'Enter sends, Shift+Enter inserts a new line',
      'chat.inputSend': 'Send',
      'chat.inputUnavailable': 'Disconnected, cannot send messages',
    }[key] ?? key),
  }),
}))

vi.mock('@/features/chat/chatStore', () => ({
  useChatStore: () => ({ connectionStatus: 'connected' }),
}))

vi.mock('@/features/workspace/projectStore', () => ({
  useProjectStore: () => ({ members: [] }),
}))

describe('ChatInput accessibility', () => {
  it('gives the icon-only send button an accessible name', () => {
    const wrapper = mount(ChatInput, {
      props: {
        modelValue: 'hello',
      },
    })

    const sendButton = wrapper.get('button.send-btn')

    expect(sendButton.attributes('aria-label')).toBe('Send')
    expect(sendButton.attributes('title')).toBe('Send')
  })
})
