<template>
  <div class="modal-overlay">
    <div class="modal-container animate-in fade-in zoom-in-95 duration-300">
      <!-- Header -->
      <div class="modal-header">
        <div class="header-text">
          <h3 class="modal-title">{{ t('members.editMemberModal.title') }}</h3>
          <p class="modal-subtitle">Update member details and access</p>
        </div>
        <button type="button" @click="$emit('close')" class="close-btn">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="modal-content">
        <!-- Identity Preview -->
        <div class="identity-preview">
          <div class="preview-avatar">
            <span>{{ member.name.charAt(0).toUpperCase() }}</span>
          </div>
          <div class="preview-meta">
            <span class="meta-role">{{ roleLabel }}</span>
            <span class="meta-id">ID: {{ member.id.slice(0, 8) }}</span>
          </div>
        </div>

        <div class="form-sections">
          <div class="input-group">
            <label>{{ t('members.editMemberModal.displayName') }}</label>
            <input
              v-model="name"
              class="modal-input"
              :placeholder="t('members.editMemberModal.displayNamePlaceholder')"
            />
          </div>

          <!-- ACP Configuration (for non-owner/admin) -->
          <div v-if="member.roleType !== 'owner'" class="input-group">
            <label>{{ t('members.editMemberModal.acpHeading') }}</label>
            <div class="acp-desc">{{ t('members.editMemberModal.acpDesc') }}</div>
            <label class="acp-toggle">
              <input type="checkbox" v-model="acpEnabled" class="modal-checkbox" />
              <span>{{ t('members.editMemberModal.acpEnabled') }}</span>
            </label>
            <template v-if="acpEnabled">
              <div class="input-group">
                <label>{{ t('members.editMemberModal.acpCommand') }}</label>
                <input v-model="acpCommand" class="modal-input" :placeholder="t('members.editMemberModal.acpCommandPlaceholder')" />
              </div>
              <div class="input-group">
                <label>{{ t('members.editMemberModal.acpArgs') }}</label>
                <input v-model="acpArgsDisplay" class="modal-input" :placeholder="t('members.editMemberModal.acpArgsPlaceholder')" />
                <div class="input-hint">{{ t('members.editMemberModal.acpArgsHint') }}</div>
              </div>
            </template>
          </div>
        </div>
      </div>

      <!-- Footer Actions -->
      <div class="modal-footer">
        <div class="footer-primary-actions">
          <button type="button" @click="handleSave" class="modal-btn-primary">
            {{ t('members.editMemberModal.save') }}
          </button>
        </div>
        
        <template v-if="showRemove">
          <div class="footer-divider"></div>
          <button type="button" @click="handleRemove" class="modal-btn-danger">
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7a4 4 0 11-8 0 4 4 0 018 0zM9 14a6 6 0 00-6 6v1h12v-1a6 6 0 00-6-6zM21 12h-6" />
            </svg>
            {{ t('members.editMemberModal.remove') }}
          </button>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Member } from '@/shared/types/member'

const { t } = useI18n()

const props = defineProps<{
  member: Member
  showRemove?: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'save', id: string, name: string, acpEnabled: boolean, acpCommand: string, acpArgs: string[]): void
  (e: 'remove', id: string): void
}>()

const name = ref(props.member.name)
const acpEnabled = ref(props.member.acpEnabled ?? false)
const acpCommand = ref(props.member.acpCommand || '')
const acpArgsDisplay = ref((props.member.acpArgs || []).join(' '))

const roleLabel = computed(() => {
  switch (props.member.roleType) {
    case 'owner': return t('members.roleOwner')
    case 'assistant': return t('members.roleAssistant')
    case 'secretary': return t('members.roleSecretary')
    default: return t('members.roleMember')
  }
})

watch(() => props.member, (newMember) => {
  name.value = newMember.name
  acpEnabled.value = newMember.acpEnabled ?? false
  acpCommand.value = newMember.acpCommand || ''
  acpArgsDisplay.value = (newMember.acpArgs || []).join(' ')
})

function handleSave() {
  const args = acpArgsDisplay.value.split(' ').filter(Boolean)
  emit('save', props.member.id, name.value, acpEnabled.value, acpCommand.value, args)
}

function handleRemove() {
  emit('remove', props.member.id)
}
</script>

<style scoped>
.modal-overlay {
  position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center;
  background: rgba(15, 23, 42, 0.15); backdrop-filter: blur(8px); padding: 24px;
}

.modal-container {
  width: 100%; max-width: 440px; background: rgba(255, 255, 255, 0.85); backdrop-filter: blur(40px);
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

.modal-content { padding: 32px 40px; }

.identity-preview {
  display: flex; align-items: center; gap: 20px; margin-bottom: 32px;
  background: white; padding: 16px; border-radius: 20px; border: 1px solid #f1f5f9;
}

.preview-avatar {
  width: 56px; height: 56px; border-radius: 16px; background: rgba(99, 102, 241, 0.1);
  display: flex; align-items: center; justify-content: center;
  font-size: 20px; font-weight: 900; color: #4f46e5;
}

.preview-meta { display: flex; flex-direction: column; gap: 2px; }
.meta-role { font-size: 10px; font-weight: 900; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.1em; }
.meta-id { font-size: 13px; font-weight: 700; color: #0f172a; }

.form-sections { display: flex; flex-direction: column; gap: 24px; }
.input-group { display: flex; flex-direction: column; gap: 10px; }
.input-group label { font-size: 11px; font-weight: 900; color: #64748b; text-transform: uppercase; letter-spacing: 0.1em; margin-left: 4px; }

.acp-desc { font-size: 12px; font-weight: 600; color: #94a3b8; margin-left: 4px; }
.acp-toggle { display: flex; align-items: center; gap: 10px; cursor: pointer; margin-left: 4px; }
.acp-toggle span { font-size: 14px; font-weight: 700; color: #475569; }
.acp-field-label { font-size: 11px; font-weight: 900; color: #64748b; text-transform: uppercase; letter-spacing: 0.1em; margin-left: 4px; display: block; }
.input-hint { font-size: 11px; color: #94a3b8; margin-left: 4px; margin-top: 4px; }

.modal-input {
  width: 100%; padding: 14px 18px; border-radius: 16px; border: 1px solid #e2e8f0;
  background: white; color: #0f172a; font-size: 15px; font-weight: 600; outline: none;
}
.modal-input:focus { border-color: #4f46e5; box-shadow: 0 0 0 4px rgba(79, 70, 229, 0.05); }

.modal-footer {
  padding: 0 40px 40px; display: flex; flex-direction: column; gap: 16px;
}

.footer-primary-actions { display: flex; gap: 12px; }

.modal-btn-primary {
  flex: 1; padding: 14px; background: #4f46e5; color: white; border-radius: 16px;
  font-size: 15px; font-weight: 900; border: none; cursor: pointer;
  box-shadow: 0 10px 25px -5px rgba(79, 70, 229, 0.4); transition: all 0.3s;
}
.modal-btn-primary:hover { background: #4338ca; transform: translateY(-1px); }

.footer-divider { height: 1px; background: rgba(15, 23, 42, 0.05); margin: 8px 0; }

.modal-btn-danger {
  width: 100%; padding: 12px; background: #fef2f2; color: #ef4444; border-radius: 14px;
  font-size: 13px; font-weight: 800; border: 1px solid #fee2e2; cursor: pointer; transition: all 0.2s;
  display: flex; align-items: center; justify-content: center;
}
.modal-btn-danger:hover { background: #fee2e2; border-color: #fecaca; }
</style>
