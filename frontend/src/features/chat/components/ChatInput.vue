<template>
  <div class="chat-input-wrapper">
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

const { t } = useI18n()
const chatStore = useChatStore()

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

function handleInput(event: Event) {
  const target = event.target as HTMLTextAreaElement
  emit('update:modelValue', target.value)
  resizeInput()
}

function handleKeydown(event: KeyboardEvent) {
  if (event.isComposing || event.key === 'Process') return

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
