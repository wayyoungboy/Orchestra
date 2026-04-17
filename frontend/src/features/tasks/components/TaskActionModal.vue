<template>
  <Teleport to="body">
    <div
      class="modal-overlay"
      @click.self="handleCancel"
      @keydown.esc="handleCancel"
    >
      <div class="modal-card animate-in fade-in zoom-in-95 duration-200">
        <div class="modal-header">
          <h2 class="modal-title">{{ title }}</h2>
          <button @click="handleCancel" class="modal-close">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="modal-body">
          <div class="task-title-section">
            <p class="task-title-label">任务</p>
            <p class="task-title-text">{{ taskTitle }}</p>
          </div>

          <div class="input-group">
            <label class="input-label">{{ inputLabel }}</label>
            <textarea
              v-model="inputValue"
              class="input-textarea"
              :placeholder="inputPlaceholder"
              rows="4"
              @keydown.enter.meta="handleSubmit"
            />
          </div>
        </div>

        <div class="modal-footer">
          <button @click="handleCancel" class="btn-cancel">取消</button>
          <button
            @click="handleSubmit"
            :disabled="!inputValue.trim()"
            :class="['btn-submit', actionClass]"
          >
            {{ submitText }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{
  open: boolean
  action: 'complete' | 'fail'
  taskTitle: string
}>()

const emit = defineEmits<{
  (e: 'submit', value: string): void
  (e: 'cancel'): void
}>()

const inputValue = ref('')

const title = computed(() => {
  return props.action === 'complete' ? '完成任务' : '标记失败'
})

const inputLabel = computed(() => {
  return props.action === 'complete' ? '结果摘要' : '失败原因'
})

const inputPlaceholder = computed(() => {
  return props.action === 'complete'
    ? '请输入任务完成的结果摘要...'
    : '请输入失败的原因...'
})

const submitText = computed(() => {
  return props.action === 'complete' ? '标记完成' : '标记失败'
})

const actionClass = computed(() => {
  return `btn-${props.action === 'complete' ? 'success' : 'danger'}`
})

function handleSubmit() {
  if (inputValue.value.trim()) {
    emit('submit', inputValue.value.trim())
    inputValue.value = ''
  }
}

function handleCancel() {
  inputValue.value = ''
  emit('cancel')
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
}

.modal-card {
  width: 100%;
  max-width: 480px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.15);
  overflow: hidden;
  border: 1px solid #e2e8f0;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px;
  border-bottom: 1px solid #f1f5f9;
}

.modal-title {
  font-size: 18px;
  font-weight: 800;
  color: #0f172a;
}

.modal-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.2s;
}

.modal-close:hover {
  background: #f1f5f9;
  color: #475569;
}

.modal-body {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.task-title-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  background: #f8fafc;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
}

.task-title-label {
  font-size: 11px;
  font-weight: 700;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.task-title-text {
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
  line-height: 1.4;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.input-label {
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
}

.input-textarea {
  padding: 12px;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  font-family: inherit;
  font-size: 14px;
  color: #0f172a;
  background: white;
  resize: none;
  transition: all 0.2s;
}

.input-textarea:focus {
  outline: none;
  border-color: #4f46e5;
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
}

.input-textarea::placeholder {
  color: #cbd5e1;
}

.modal-footer {
  display: flex;
  gap: 12px;
  padding: 24px;
  border-top: 1px solid #f1f5f9;
}

.btn-cancel,
.btn-submit {
  flex: 1;
  padding: 12px 16px;
  border-radius: 10px;
  border: none;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel {
  background: #f1f5f9;
  color: #475569;
}

.btn-cancel:hover {
  background: #e2e8f0;
  color: #0f172a;
}

.btn-submit {
  color: white;
  font-weight: 700;
}

.btn-submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-submit:not(:disabled):hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
}

.btn-success {
  background: #10b981;
}

.btn-success:not(:disabled):hover {
  background: #059669;
}

.btn-danger {
  background: #ef4444;
}

.btn-danger:not(:disabled):hover {
  background: #dc2626;
}
</style>
