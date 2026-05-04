<template>
  <div class="chat-input-wrapper">
    <!-- @mention Autocomplete -->
    <div v-if="mentionOpen" class="mention-popup">
      <div class="mention-header">成员</div>
      <div
        v-for="(member, idx) in mentionFiltered"
        :key="member.id"
        :class="['mention-item', idx === mentionIndex ? 'is-active' : '']"
        @mousedown.prevent="selectMention(member)"
        @mouseenter="mentionIndex = idx"
      >
        <div :class="['mention-avatar', member.roleType === 'assistant' ? 'is-ai' : '', member.roleType === 'owner' ? 'is-owner' : '']">
          {{ member.name.charAt(0).toUpperCase() }}
        </div>
        <div class="mention-info">
          <span class="mention-name">{{ member.name }}</span>
          <span class="mention-role">{{ roleLabel(member.roleType) }}</span>
        </div>
      </div>
      <div
        :class="['mention-item mention-all', mentionIndex === mentionFiltered.length ? 'is-active' : '']"
        @mousedown.prevent="selectMentionAll"
        @mouseenter="mentionIndex = mentionFiltered.length"
      >
        <div class="mention-avatar is-all">@</div>
        <div class="mention-info">
          <span class="mention-name">all</span>
          <span class="mention-role">通知所有成员</span>
        </div>
      </div>
    </div>

    <!-- Input Container -->
    <div class="chat-input-root">
      <!-- Text Input -->
      <div class="input-area custom-scrollbar">
        <textarea
          ref="inputRef"
          :value="modelValue"
          class="chat-textarea"
          :placeholder="placeholder"
          :maxlength="maxLength"
          spellcheck="false"
          rows="1"
          @input="handleInput"
          @keydown="handleKeydown"
        ></textarea>
      </div>

      <!-- Action Buttons -->
      <div class="action-area">
        <button
          type="button"
          class="send-btn"
          :disabled="!modelValue.trim() || connectionStatus !== 'connected'"
          :title="connectionStatus !== 'connected' ? '已断线，无法发送消息' : '发送消息'"
          @click="handleSend"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Character Count -->
    <div class="input-footer">
      <div class="footer-hint">
        <div class="hint-dot"></div>
        <span>{{ t('chat.inputHint') }}</span>
      </div>
      <span :class="['char-count', modelValue.length > maxLength * 0.9 ? 'is-near-limit' : '']">
        {{ modelValue.length }} / {{ maxLength }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useChatStore } from '@/features/chat/chatStore'
import { useProjectStore } from '@/features/workspace/projectStore'

const { t } = useI18n()
const chatStore = useChatStore()
const projectStore = useProjectStore()

const connectionStatus = computed(() => chatStore.connectionStatus)

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  maxLength?: number
}>(), {
  maxLength: 1200
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'send'): void
}>()

const inputRef = ref<HTMLTextAreaElement | null>(null)

// --- @mention autocomplete ---
const mentionOpen = ref(false)
const mentionQuery = ref('')
const mentionIndex = ref(0)
const mentionStart = ref(-1)

const mentionFiltered = computed(() => {
  const q = mentionQuery.value.toLowerCase()
  return projectStore.members.filter(m =>
    m.name.toLowerCase().includes(q)
  )
})

function roleLabel(role: string) {
  const map: Record<string, string> = { owner: '所有者', admin: '管理员', assistant: 'AI 助手', secretary: '秘书', member: '成员' }
  return map[role] || role
}

function checkMentionTrigger(value: string, cursorPos: number) {
  const before = value.slice(0, cursorPos)
  const atIdx = before.lastIndexOf('@')
  if (atIdx === -1 || (atIdx > 0 && before[atIdx - 1] !== ' ' && before[atIdx - 1] !== '\n')) {
    mentionOpen.value = false
    return
  }
  const query = before.slice(atIdx + 1)
  if (query.includes(' ') || query.includes('\n')) {
    mentionOpen.value = false
    return
  }
  mentionStart.value = atIdx
  mentionQuery.value = query
  mentionIndex.value = 0
  mentionOpen.value = true
}

function selectMention(member: { name: string }) {
  if (mentionStart.value < 0) return
  const before = props.modelValue.slice(0, mentionStart.value)
  const cursorPos = inputRef.value?.selectionStart ?? props.modelValue.length
  const after = props.modelValue.slice(cursorPos)
  const inserted = `@${member.name} `
  emit('update:modelValue', before + inserted + after)
  mentionOpen.value = false
  nextTick(() => {
    if (inputRef.value) {
      const pos = before.length + inserted.length
      inputRef.value.selectionStart = pos
      inputRef.value.selectionEnd = pos
      inputRef.value.focus()
    }
  })
}

function selectMentionAll() {
  if (mentionStart.value < 0) return
  const before = props.modelValue.slice(0, mentionStart.value)
  const cursorPos = inputRef.value?.selectionStart ?? props.modelValue.length
  const after = props.modelValue.slice(cursorPos)
  const inserted = '@all '
  emit('update:modelValue', before + inserted + after)
  mentionOpen.value = false
  nextTick(() => {
    if (inputRef.value) {
      const pos = before.length + inserted.length
      inputRef.value.selectionStart = pos
      inputRef.value.selectionEnd = pos
      inputRef.value.focus()
    }
  })
}

function handleInput(event: Event) {
  const target = event.target as HTMLTextAreaElement
  emit('update:modelValue', target.value)
  checkMentionTrigger(target.value, target.selectionStart ?? target.value.length)
  resizeInput()
}

function handleKeydown(event: KeyboardEvent) {
  if (event.isComposing || event.key === 'Process') return

  if (mentionOpen.value) {
    const totalItems = mentionFiltered.value.length + 1
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      mentionIndex.value = (mentionIndex.value + 1) % totalItems
      return
    }
    if (event.key === 'ArrowUp') {
      event.preventDefault()
      mentionIndex.value = (mentionIndex.value - 1 + totalItems) % totalItems
      return
    }
    if (event.key === 'Enter' || event.key === 'Tab') {
      event.preventDefault()
      if (mentionIndex.value < mentionFiltered.value.length) {
        selectMention(mentionFiltered.value[mentionIndex.value])
      } else {
        selectMentionAll()
      }
      return
    }
    if (event.key === 'Escape') {
      event.preventDefault()
      mentionOpen.value = false
      return
    }
  }

  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    handleSend()
  }
}

function handleSend() {
  if (props.modelValue.trim()) emit('send')
}

function resizeInput() {
  if (!inputRef.value) return
  inputRef.value.style.height = 'auto'
  inputRef.value.style.height = `${Math.min(inputRef.value.scrollHeight, 160)}px`
}

watch(() => props.modelValue, async () => {
  await nextTick()
  resizeInput()
})
</script>

<style scoped>
.chat-input-wrapper {
  position: relative;
  padding: 0 24px 24px;
}

/* @mention popup */
.mention-popup {
  position: absolute;
  bottom: 100%;
  left: 24px;
  right: 24px;
  max-height: 240px;
  overflow-y: auto;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border-radius: 16px;
  border: 1px solid rgba(226, 232, 240, 0.8);
  box-shadow: 0 -8px 30px rgba(0, 0, 0, 0.08);
  margin-bottom: 8px;
  padding: 6px;
  z-index: 100;
}

.mention-header {
  padding: 6px 12px;
  font-size: 10px;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #94a3b8;
}

.mention-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 10px;
  cursor: pointer;
  transition: background 0.15s;
}

.mention-item:hover,
.mention-item.is-active {
  background: rgba(79, 70, 229, 0.08);
}

.mention-avatar {
  width: 32px;
  height: 32px;
  border-radius: 10px;
  background: #f1f5f9;
  color: #64748b;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 900;
  flex-shrink: 0;
}

.mention-avatar.is-ai {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.mention-avatar.is-owner {
  background: rgba(99, 102, 241, 0.1);
  color: #4f46e5;
}

.mention-avatar.is-all {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
  font-weight: 900;
}

.mention-info {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.mention-name {
  font-size: 13px;
  font-weight: 700;
  color: #0f172a;
}

.mention-role {
  font-size: 10px;
  font-weight: 600;
  color: #94a3b8;
}

.chat-input-root {
  position: relative;
  display: flex;
  align-items: flex-end;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(32px);
  -webkit-backdrop-filter: blur(32px);
  border-radius: 24px;
  border: 1px solid white;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.04);
  padding: 8px 8px 8px 24px;
  gap: 16px;
  transition: all 0.3s ease;
}

.chat-input-root:focus-within {
  box-shadow: 0 25px 60px rgba(79, 70, 229, 0.08), 0 0 0 4px rgba(79, 70, 229, 0.1);
  border-color: rgba(79, 70, 229, 0.3);
}

.input-area {
  flex: 1;
  min-height: 56px;
  max-height: 160px;
  overflow-y: auto;
  padding: 18px 0;
  display: flex;
  align-items: center;
}

.chat-textarea {
  width: 100%;
  background: transparent;
  border: none;
  padding: 0;
  font-size: 15px;
  font-weight: 500;
  color: #0f172a;
  line-height: 1.5;
  resize: none;
  outline: none;
}

.chat-textarea::placeholder {
  color: #94a3b8;
}

.action-area {
  flex-shrink: 0;
  margin-bottom: 4px;
}

.send-btn {
  width: 48px;
  height: 48px;
  border-radius: 16px;
  background: #4f46e5;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0 10px 20px -5px rgba(79, 70, 229, 0.4);
}

.send-btn:hover:not(:disabled) {
  background: #4338ca;
  transform: translateY(-2px);
  box-shadow: 0 12px 25px -5px rgba(79, 70, 229, 0.5);
}

.send-btn:active:not(:disabled) {
  transform: translateY(0);
}

.send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  filter: grayscale(1);
  box-shadow: none;
}

.input-footer {
  margin-top: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 12px;
}

.footer-hint {
  display: flex;
  align-items: center;
  gap: 6px;
}

.hint-dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: #64748b;
}

.footer-hint span {
  font-size: 10px;
  font-weight: 800;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.char-count {
  font-size: 10px;
  font-weight: 800;
  color: #94a3b8;
  letter-spacing: 0.1em;
}

.is-near-limit {
  color: #ef4444;
}
</style>
