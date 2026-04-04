<template>
  <div class="p-6 pb-8">
    <!-- Input Container -->
    <div
      class="relative bg-white/5 backdrop-blur-md rounded-2xl shadow-lg flex items-end p-2 gap-3 border border-white/5 focus-within:border-primary/50 focus-within:ring-1 focus-within:ring-primary/25 focus-within:bg-white/[0.07] transition-all duration-300 group"
    >
      <!-- Mention Suggestions Dropdown -->
      <div
        v-if="showMentionSuggestions && mentionSuggestions.length"
        class="absolute bottom-full left-0 mb-2 w-64 bg-panel/95 border border-white/10 rounded-xl shadow-2xl overflow-hidden z-20"
      >
        <button
          v-for="(member, index) in mentionSuggestions"
          :key="member.id"
          type="button"
          :class="[
            'w-full px-3 py-2 flex items-center gap-3 text-left transition-colors',
            index === activeMentionIndex ? 'bg-white/10' : 'hover:bg-white/5'
          ]"
          @click="applyMention(member)"
        >
          <div class="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center">
            <span class="text-xs font-bold text-white">{{ member.name.charAt(0) }}</span>
          </div>
          <div>
            <div class="text-xs font-semibold text-white">@{{ member.name }}</div>
            <div class="text-[10px] text-white/40">{{ roleTypeLabel(member.roleType) }}</div>
          </div>
        </button>
      </div>

      <!-- Text Input -->
      <div class="flex-1 min-h-[44px] max-h-40 overflow-y-auto py-2.5">
        <textarea
          ref="inputRef"
          :value="modelValue"
          class="w-full bg-transparent border-none p-0 text-white placeholder-white/30 focus:ring-0 outline-none text-[15px] font-light resize-none min-h-[24px]"
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
      <div class="flex items-center shrink-0 mb-0.5">
        <!-- Send Button -->
        <button
          type="button"
          :disabled="!modelValue.trim()"
          :class="[
            'h-10 px-5 bg-primary hover:bg-primary-hover text-white text-sm font-bold rounded-xl shadow-glow flex items-center gap-2 transition-all active:scale-95 transform',
            !modelValue.trim() ? 'opacity-50 cursor-not-allowed' : ''
          ]"
          @click="handleSend"
        >
          {{ t('chat.inputSend') }}
        </button>
      </div>
    </div>

    <!-- Character Count -->
    <div class="mt-2 flex items-center justify-between px-2 text-[11px] text-white/30 font-medium tracking-wide">
      <span>{{ t('chat.inputHint') }}</span>
      <span>{{ modelValue.length }}/{{ maxLength }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

function roleTypeLabel(roleType: string) {
  switch (roleType) {
    case 'owner':
      return t('members.roleOwner')
    case 'admin':
      return t('members.roleAdmin')
    case 'assistant':
      return t('members.roleAssistant')
    case 'secretary':
      return t('members.roleSecretary')
    case 'member':
      return t('members.roleMember')
    default:
      return roleType
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

// Computed for mention suggestions
const mentionSuggestions = computed(() => {
  const query = mentionQuery.value.toLowerCase()
  // Show all members when query is empty (just typed @), otherwise filter by query
  if (!query) {
    return props.members.slice(0, 6)
  }
  return props.members
    .filter(m => m.name.toLowerCase().includes(query))
    .slice(0, 6)
})

const showMentionSuggestions = computed(() => {
  // Show dropdown when we're in a mention context (typed @ with optional query)
  const inMentionContext = props.modelValue.slice(0, cursorIndex.value).match(/@([^\s]*)$/)
  return inMentionContext !== null && mentionSuggestions.value.length > 0
})

// Watch for @ symbol
watch(
  () => props.modelValue,
  (value) => {
    // Detect @mention pattern
    const match = value.slice(0, cursorIndex.value).match(/@([^\s]*)$/)
    if (match) {
      mentionQuery.value = match[1] || ''
    } else {
      mentionQuery.value = ''
    }
  }
)

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
    if (typeof start === 'number') {
      cursorIndex.value = start
    }
  }
}

function handleInput(event: Event) {
  const target = event.target as HTMLTextAreaElement
  // Update cursor index BEFORE emitting value, so watcher can detect @ correctly
  cursorIndex.value = target.selectionStart
  emit('update:modelValue', target.value)
  resizeInput()
}

function handleKeydown(event: KeyboardEvent) {
  // IME：回车用于选词/上屏，不能拦截，否则中文 @ 提及后无法继续输入或提交
  if (event.isComposing || event.key === 'Process') {
    return
  }

  syncCursorFromInput()

  // Handle mention suggestion navigation
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
      // 无有效候选项时不吞掉 Enter，交给下方发送逻辑
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

  // Focus back to input and set cursor position after the mention
  nextTick(() => {
    if (inputRef.value) {
      const newCursorPos = beforeMention.length + member.name.length + 2
      inputRef.value.focus()
      inputRef.value.setSelectionRange(newCursorPos, newCursorPos)
    }
  })
}

function handleSend() {
  if (props.modelValue.trim()) {
    emit('send')
  }
}

function resizeInput() {
  if (!inputRef.value) return
  inputRef.value.style.height = 'auto'
  inputRef.value.style.height = `${Math.min(inputRef.value.scrollHeight, 160)}px`
}

watch(
  () => props.modelValue,
  async () => {
    await nextTick()
    resizeInput()
  }
)
</script>

<style scoped>
.shadow-glow {
  box-shadow: 0 0 20px rgba(99, 102, 241, 0.3);
}
</style>