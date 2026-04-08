<template>
  <div class="modal-overlay">
    <div class="modal-container animate-in fade-in zoom-in-95 duration-300">
      <!-- Header -->
      <div class="modal-header">
        <div class="header-text">
          <h3 class="modal-title">{{ title }}</h3>
          <p class="modal-subtitle">Configure identity and permissions</p>
        </div>
        <button type="button" @click="$emit('close')" class="close-btn">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="modal-content custom-scrollbar">
        <!-- Admin Invitation -->
        <div v-if="mode === 'admin'" class="form-sections">
          <div class="input-group">
            <label>{{ t('members.addMemberModal.adminName') }}</label>
            <input
              v-model="memberName"
              class="modal-input"
              :placeholder="t('members.addMemberModal.adminNamePlaceholder')"
            />
          </div>

          <div class="input-group">
            <label>{{ t('members.addMemberModal.permissions') }}</label>
            <div class="perms-list">
              <label
                v-for="row in adminPermRows"
                :key="row.key"
                class="perm-item"
              >
                <input
                  type="checkbox"
                  v-model="adminPerms[row.key]"
                  class="modal-checkbox"
                />
                <div class="perm-info">
                  <span class="perm-name">{{ t(row.titleKey) }}</span>
                  <span class="perm-desc">{{ t(row.descKey) }}</span>
                </div>
              </label>
            </div>
          </div>
        </div>

        <!-- Member / AI Assistant / Secretary -->
        <div v-else class="form-sections">
          <div v-if="mode === 'member'" class="input-group">
            <label>{{ t('members.addMemberModal.memberName') }}</label>
            <input
              v-model="memberName"
              type="text"
              class="modal-input"
              :placeholder="t('members.addMemberModal.memberNamePlaceholder')"
            />
          </div>
          <div v-else class="input-group">
            <label>{{ t('members.addMemberModal.assistantDisplayName') }}</label>
            <input
              v-model="assistantDisplayName"
              type="text"
              class="modal-input"
              :placeholder="t('members.addMemberModal.assistantDisplayNamePlaceholder')"
            />
          </div>

          <div class="input-group">
            <label>{{ cliPickerLabel }}</label>
            <div class="model-picker-grid">
              <button
                v-for="model in assistantModels"
                :key="model.id"
                type="button"
                @click="selectedModel = model.id"
                :class="['model-card', selectedModel === model.id ? 'is-selected' : '']"
              >
                <div :class="['model-avatar', model.accentClass]">
                  <span class="model-initials">{{ model.initials }}</span>
                </div>
                <div class="model-meta">
                  <span class="model-label">{{ model.label }}</span>
                  <span class="model-tag">Available</span>
                </div>
                <div v-if="selectedModel === model.id" class="check-icon">
                  <svg fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                </div>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer Actions -->
      <div class="modal-footer">
        <button type="button" @click="$emit('close')" class="modal-btn-secondary">
          {{ t('members.addMemberModal.cancel') }}
        </button>
        <button
          type="button"
          @click="mode === 'admin' ? handleAddAdmin() : handleInvite()"
          :disabled="inviteSubmitDisabled"
          class="modal-btn-primary"
        >
          {{ mode === 'admin' ? t('members.addMemberModal.submitAdmin') : (mode === 'member' ? t('members.addMemberModal.submitMember') : t('members.addMemberModal.submitAdd')) }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MemberRole } from '@/shared/types/member'

const { t } = useI18n()

const props = defineProps<{
  mode?: 'assistant' | 'admin' | 'member' | 'secretary'
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'invite', data: { name: string; roleType: MemberRole; command?: string; terminalType?: string }): void
}>()

const mode = computed(() => props.mode || 'assistant')
const title = computed(() => {
  switch (mode.value) {
    case 'admin': return t('members.addMemberModal.titleAdmin')
    case 'member': return t('members.addMemberModal.titleMember')
    case 'secretary': return t('members.addMemberModal.titleSecretary')
    default: return t('members.addMemberModal.titleAssistant')
  }
})

const cliPickerLabel = computed(() => {
  switch (mode.value) {
    case 'secretary': return t('members.addMemberModal.selectSecretary')
    case 'member': return t('members.addMemberModal.selectMemberCli')
    default: return t('members.addMemberModal.selectAssistant')
  }
})

const selectedModel = ref('claude')
const memberName = ref('')
const assistantDisplayName = ref('')

const inviteSubmitDisabled = computed(() => 
  (mode.value === 'admin' && !memberName.value.trim()) ||
  (mode.value !== 'admin' && (!selectedModel.value || (mode.value === 'member' && !memberName.value.trim())))
)

const adminPermRows = [
  { key: 'fullAccess' as const, titleKey: 'members.addMemberModal.permFullAccess', descKey: 'members.addMemberModal.permFullAccessDesc' },
  { key: 'memberManage' as const, titleKey: 'members.addMemberModal.permMemberManage', descKey: 'members.addMemberModal.permMemberManageDesc' },
  { key: 'billing' as const, titleKey: 'members.addMemberModal.permBilling', descKey: 'members.addMemberModal.permBillingDesc' }
]

const adminPerms = reactive({ fullAccess: true, memberManage: true, billing: false })

const assistantModels = computed(() => {
  const defs = [
    { id: 'claude', initials: 'CC', command: 'claude', terminalType: 'claude', accentClass: 'bg-orange-100 text-orange-600' },
    { id: 'gemini', initials: 'GM', command: 'gemini', terminalType: 'gemini', accentClass: 'bg-blue-100 text-blue-600' },
    { id: 'aider', initials: 'AI', command: 'aider', terminalType: 'aider', accentClass: 'bg-emerald-100 text-emerald-600' },
    { id: 'cursor', initials: 'CA', command: 'cursor-agent', terminalType: 'cursor', accentClass: 'bg-violet-100 text-violet-600' }
  ]
  return defs.map(d => ({ ...d, label: t(`members.assistantModels.${d.id}`) }))
})

function handleInvite() {
  const model = assistantModels.value.find(m => m.id === selectedModel.value)
  if (!model) return
  const roleType: MemberRole = mode.value === 'member' ? 'member' : (mode.value === 'secretary' ? 'secretary' : 'assistant')
  const name = mode.value === 'member' ? memberName.value.trim() : (assistantDisplayName.value.trim() || model.label)
  emit('invite', { name, roleType, command: model.command, terminalType: model.terminalType })
}

function handleAddAdmin() {
  if (memberName.value.trim()) emit('invite', { name: memberName.value.trim(), roleType: 'admin' })
}
</script>

<style scoped>
.modal-overlay {
  position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center;
  background: rgba(15, 23, 42, 0.15); backdrop-filter: blur(8px); padding: 24px;
}

.modal-container {
  width: 100%; max-width: 520px; background: rgba(255, 255, 255, 0.85); backdrop-filter: blur(40px);
  border-radius: 40px; border: 1px solid white; shadow: 0 40px 100px -20px rgba(0,0,0,0.15);
  display: flex; flex-direction: column; overflow: hidden;
}

.modal-header {
  padding: 32px 40px; border-bottom: 1px solid rgba(15, 23, 42, 0.05);
  display: flex; align-items: center; justify-content: space-between;
}

.modal-title { font-size: 22px; font-weight: 950; color: #0f172a; letter-spacing: -0.02em; }
.modal-subtitle { font-size: 11px; font-weight: 800; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.15em; margin-top: 4px; }

.close-btn {
  width: 40px; height: 40px; border-radius: 50%; display: flex; align-items: center; justify-content: center;
  color: #cbd5e1; transition: all 0.2s; border: none; background: transparent; cursor: pointer;
}
.close-btn:hover { background: #f1f5f9; color: #0f172a; }

.modal-content { flex: 1; overflow-y: auto; padding: 32px 40px; }
.form-sections { display: flex; flex-direction: column; gap: 32px; }

.input-group { display: flex; flex-direction: column; gap: 10px; }
.input-group label { font-size: 11px; font-weight: 900; color: #64748b; text-transform: uppercase; letter-spacing: 0.1em; margin-left: 4px; }

.modal-input {
  width: 100%; padding: 14px 18px; border-radius: 16px; border: 1px solid #e2e8f0;
  background: white; color: #0f172a; font-size: 15px; font-weight: 600; outline: none;
  transition: all 0.2s;
}
.modal-input:focus { border-color: #4f46e5; box-shadow: 0 0 0 4px rgba(79, 70, 229, 0.05); }

.model-picker-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; }
.model-card {
  padding: 16px; border-radius: 20px; border: 1px solid #f1f5f9; background: white;
  display: flex; align-items: center; gap: 12px; transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
  cursor: pointer; text-align: left; position: relative;
}
.model-card:hover { border-color: #4f46e544; background: #f8fafc; }
.model-card.is-selected { border-color: #4f46e5; box-shadow: 0 10px 25px -5px rgba(79, 70, 229, 0.15); }

.model-avatar {
  width: 40px; height: 40px; border-radius: 12px; display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.model-initials { font-size: 14px; font-weight: 900; }
.model-meta { display: flex; flex-direction: column; min-width: 0; }
.model-label { font-size: 14px; font-weight: 800; color: #0f172a; }
.model-tag { font-size: 9px; font-weight: 700; color: #94a3b8; text-transform: uppercase; }

.check-icon { margin-left: auto; color: #4f46e5; width: 18px; height: 18px; }

.modal-footer {
  padding: 32px 40px; border-top: 1px solid rgba(15, 23, 42, 0.05);
  display: flex; align-items: center; gap: 16px;
}

.modal-btn-primary {
  flex: 1; padding: 14px; background: #4f46e5; color: white; border-radius: 16px;
  font-size: 15px; font-weight: 900; border: none; cursor: pointer;
  box-shadow: 0 10px 25px -5px rgba(79, 70, 229, 0.4); transition: all 0.3s;
}
.modal-btn-primary:hover { background: #4338ca; transform: translateY(-1px); }
.modal-btn-primary:disabled { opacity: 0.5; cursor: not-allowed; grayscale: 1; }

.modal-btn-secondary {
  flex: 1; padding: 14px; background: white; color: #64748b; border-radius: 16px;
  font-size: 15px; font-weight: 800; border: 1px solid #e2e8f0; cursor: pointer; transition: all 0.2s;
}
.modal-btn-secondary:hover { background: #f8fafc; color: #0f172a; border-color: #cbd5e1; }
</style>
