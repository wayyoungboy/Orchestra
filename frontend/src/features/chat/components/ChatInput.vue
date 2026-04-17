<template>
  <div class="chat-input-wrapper">
    <!-- Input Container -->
    <div class="chat-input-root">
      <!-- Mention Suggestions Dropdown -->
      <div
        v-if="showMentionSuggestions && mentionSuggestions.length"
        class="mention-dropdown"
      >
        <div class="dropdown-header">
          <span>{{ t('chat.mentionSuggestions') }}</span>
        </div>
        <button
          v-for="(member, index) in mentionSuggestions"
          :key="member.id"
          type="button"
          :class="['mention-item', index === activeMentionIndex ? 'is-active' : '']"
          @click="applyMention(member)"
        >
          <div class="mention-avatar">
            <span>{{ member.name.charAt(0).toUpperCase() }}</span>
          </div>
          <div class="mention-info">
            <div class="mention-name">@{{ member.name }}</div>
            <div class="mention-role">{{ roleTypeLabel(member.roleType) }}</div>
          </div>
        </button>
      </div>

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
          @keyup="updateCursorIndex"
          @compositionend="updateCursorIndex"
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
import { useProjectStore } from '@/features/workspace/projectStore'
import { useChatStore } from '@/features/chat/chatStore'

const { t } = useI18n()
const projectStore = useProjectStore()
const chatStore = useChatStore()

const connectionStatus = computed(() => chatStore.connectionStatus)

function roleTypeLabel(roleType: string) {
  switch (roleType) {
    case 'owner': return t('members.roleOwner')
    case 'assistant': return t('members.roleAssistant')
    case 'secretary': return t('members.roleSecretary')
    default: return roleType
  }
}

interface Member {
  id: string
  name: string
  roleType: string
}

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  maxLength?: number
  members?: Member[]
}>(), {
  maxLength: 1200,
  members: () => []
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'send'): void
}>()

const inputRef = ref<HTMLTextAreaElement | null>(null)
const cursorIndex = ref(0)
const activeMentionIndex = ref(0)
const mentionQuery = ref('')

// Use projectStore.members directly if props.members is empty
const effectiveMembers = computed(() => {
  if (props.members && props.members.length > 0) {
    return props.members
  }
  return projectStore.members || []
})

const mentionSuggestions = computed(() => {
  const query = mentionQuery.value.toLowerCase()
  if (!query) return effectiveMembers.value.slice(0, 6)
  return effectiveMembers.value.filter(m => m.name.toLowerCase().includes(query)).slice(0, 6)
})

const showMentionSuggestions = computed(() => {
  const inMentionContext = props.modelValue.slice(0, cursorIndex.value).match(/@([^\s]*)$/)
  return inMentionContext !== null && mentionSuggestions.value.length > 0
})

watch(() => props.modelValue, (value) => {
  const match = value.slice(0, cursorIndex.value).match(/@([^\s]*)$/)
  if (match) mentionQuery.value = match[1] || ''
  else mentionQuery.value = ''
})

watch(mentionSuggestions, (list) => {
  if (activeMentionIndex.value >= list.length) {
    activeMentionIndex.value = Math.max(0, list.length - 1)
  }
})

function updateCursorIndex() {
  syncCursorFromInput()
}

function syncCursorFromInput() {
  if (inputRef.value) {
    const start = inputRef.value.selectionStart
    if (typeof start === 'number') cursorIndex.value = start
  }
}

function handleInput(event: Event) {
  const target = event.target as HTMLTextAreaElement
  cursorIndex.value = target.selectionStart
  emit('update:modelValue', target.value)
  resizeInput()
}

function handleKeydown(event: KeyboardEvent) {
  if (event.isComposing || event.key === 'Process') return

  syncCursorFromInput()

  if (showMentionSuggestions.value && mentionSuggestions.value.length > 0) {
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      activeMentionIndex.value = Math.min(activeMentionIndex.value + 1, mentionSuggestions.value.length - 1)
      return
    }
    if (event.key === 'ArrowUp') {
      event.preventDefault()
      activeMentionIndex.value = Math.max(activeMentionIndex.value - 1, 0)
      return
    }
    if (event.key === 'Enter' || event.key === 'Tab') {
      const idx = Math.min(Math.max(0, activeMentionIndex.value), mentionSuggestions.value.length - 1)
      const pick = mentionSuggestions.value[idx]
      if (pick) {
        event.preventDefault()
        applyMention(pick)
        return
      }
    }
    if (event.key === 'Escape') {
      mentionQuery.value = ''
      return
    }
  }

  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    handleSend()
  }
}

function applyMention(member: Member) {
  const beforeMention = props.modelValue.slice(0, cursorIndex.value).replace(/@[^\s]*$/, '')
  const afterCursor = props.modelValue.slice(cursorIndex.value)
  const newValue = `${beforeMention}@${member.name} ${afterCursor}`
  emit('update:modelValue', newValue)
  mentionQuery.value = ''
  activeMentionIndex.value = 0

  nextTick(() => {
    if (inputRef.value) {
      const newCursorPos = beforeMention.length + member.name.length + 2
      inputRef.value.focus()
      inputRef.value.setSelectionRange(newCursorPos, newCursorPos)
    }
  })
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
  padding: 0 24px 24px;
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

.mention-dropdown {
  position: absolute;
  bottom: calc(100% + 12px);
  left: 0;
  width: 320px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(24px);
  border-radius: 20px;
  border: 1px solid rgba(226, 232, 240, 0.8);
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  z-index: 20;
}

.dropdown-header {
  padding: 10px 16px;
  border-bottom: 1px solid #f1f5f9;
}

.dropdown-header span {
  font-size: 10px;
  font-weight: 900;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.15em;
}

.mention-item {
  width: 100%;
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  text-align: left;
  transition: all 0.2s;
  background: transparent;
  border: none;
  cursor: pointer;
}

.mention-item:hover, .mention-item.is-active {
  background: #f8fafc;
}

.mention-item.is-active {
  background: rgba(79, 70, 229, 0.05);
}

.mention-avatar {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: #4f46e5;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 14px;
  font-weight: 900;
  flex-shrink: 0;
}

.mention-info {
  flex: 1;
  min-width: 0;
}

.mention-name {
  font-size: 14px;
  font-weight: 800;
  color: #0f172a;
  truncate: true;
}

.mention-role {
  font-size: 11px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
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
