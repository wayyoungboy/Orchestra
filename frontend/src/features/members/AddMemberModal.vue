<template>
  <Teleport to="body">
    <div class="modal-overlay" @click.self="$emit('close')">
      <div class="modal-card animate-in fade-in zoom-in-95 duration-200">
        <div class="modal-header">
          <h2 class="modal-title">{{ modalTitle }}</h2>
          <button @click="$emit('close')" class="modal-close">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="modal-body">
          <div class="form-sections">
            <h3 class="section-title">{{ t('members.addMemberModal.agentConfig') }}</h3>
            <div class="input-group">
              <label>{{ t('members.addMemberModal.displayName') }}</label>
              <input v-model="assistantDisplayName" :placeholder="t('members.addMemberModal.displayNamePlaceholder')" />
            </div>
            <div class="input-group">
              <label>{{ t('members.addMemberModal.cliCommand') }}</label>
              <input v-model="selectedCli" :placeholder="t('members.addMemberModal.placeholderAssistantCli')" />
            </div>
            <div class="input-group">
              <label>{{ t('members.addMemberModal.cliArgs') }}</label>
              <input v-model="cliArgsDisplay" :placeholder="t('members.addMemberModal.cliArgsPlaceholder')" />
            </div>
            <p class="field-hint">{{ mode === 'secretary' ? t('members.addMemberModal.hintSecretary') : t('members.addMemberModal.hintAssistant') }}</p>
          </div>
        </div>

        <div class="modal-footer">
          <button @click="$emit('close')" class="btn-cancel">{{ t('members.addMemberModal.cancel') }}</button>
          <button @click="handleInvite()" class="btn-invite" :disabled="isFormInvalid">{{ t('members.addMemberModal.submitAdd') }}</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { type MemberRole } from '@/shared/types/member'

const { t } = useI18n()

const props = defineProps<{
  mode?: 'assistant' | 'secretary'
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'invite', data: { name: string; roleType: MemberRole; command: string; terminalType: string; args: string[] }): void
}>()

const assistantDisplayName = ref('')
const selectedCli = ref('claude')
const cliArgsDisplay = ref('')

const cliArgs = computed(() => cliArgsDisplay.value.split(/\s+/).filter(Boolean))

const modalTitle = computed(() => {
  switch (props.mode) {
    case 'secretary': return t('members.addMemberModal.titleSecretary')
    default: return t('members.addMemberModal.titleAssistant')
  }
})

const isFormInvalid = computed(() => {
  return !selectedCli.value || !assistantDisplayName.value.trim()
})

function handleInvite() {
  if (isFormInvalid.value) return
  const roleType: MemberRole = props.mode === 'secretary' ? 'secretary' : 'assistant'
  const name = assistantDisplayName.value.trim()
  emit('invite', {
    name,
    roleType,
    command: selectedCli.value,
    terminalType: selectedCli.value,
    args: cliArgs.value
  })
}
</script>

<style scoped>
.modal-overlay {
  position: fixed; inset: 0; background: rgba(15, 23, 42, 0.4); backdrop-filter: blur(8px);
  display: flex; align-items: center; justify-content: center; z-index: 100;
}
.modal-card {
  background: rgba(255, 255, 255, 0.95); backdrop-filter: blur(32px);
  border-radius: 28px; border: 1px solid white; width: 520px; max-width: 95vw;
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.12);
}
.modal-header {
  display: flex; align-items: center; justify-content: space-between; padding: 28px 32px 16px;
}
.modal-title { font-size: 22px; font-weight: 900; color: #0f172a; }
.modal-close {
  width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center; justify-content: center;
  color: #94a3b8; background: transparent; border: none; cursor: pointer; transition: all 0.2s;
}
.modal-close:hover { background: #f1f5f9; color: #0f172a; }
.modal-body { padding: 0 32px 24px; }
.form-sections { display: flex; flex-direction: column; gap: 20px; }
.section-title { font-size: 13px; font-weight: 900; color: #64748b; text-transform: uppercase; letter-spacing: 0.08em; }
.input-group { display: flex; flex-direction: column; gap: 6px; }
.input-group label { font-size: 12px; font-weight: 800; color: #334155; }
.input-group input, .input-group select {
  padding: 12px 16px; border-radius: 14px; border: 1px solid #e2e8f0;
  font-size: 14px; font-weight: 600; color: #0f172a; outline: none;
  transition: all 0.2s; background: rgba(248, 250, 252, 0.5);
}
.input-group input:focus, .input-group select:focus { border-color: #4f46e5; background: white; box-shadow: 0 0 0 4px rgba(79, 70, 229, 0.08); }
.field-hint { font-size: 11px; color: #94a3b8; font-weight: 600; }
.modal-footer {
  display: flex; gap: 12px; justify-content: flex-end; padding: 20px 32px 28px;
  border-top: 1px solid #f1f5f9;
}
.btn-cancel {
  padding: 12px 24px; border-radius: 14px; border: 1px solid #e2e8f0;
  background: white; color: #475569; font-size: 14px; font-weight: 800;
  cursor: pointer; transition: all 0.2s;
}
.btn-cancel:hover { background: #f8fafc; }
.btn-invite {
  padding: 12px 28px; border-radius: 14px; border: none;
  background: #4f46e5; color: white; font-size: 14px; font-weight: 800;
  cursor: pointer; transition: all 0.2s;
}
.btn-invite:hover:not(:disabled) { background: #4338ca; }
.btn-invite:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
