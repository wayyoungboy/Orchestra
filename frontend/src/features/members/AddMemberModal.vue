<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
    <div class="w-full max-w-sm bg-panel/90 border border-white/10 rounded-2xl shadow-2xl p-6">
      <div class="flex justify-between items-center mb-6">
        <h3 class="text-white font-bold text-lg">{{ title }}</h3>
        <button type="button" @click="$emit('close')" class="text-white/40 hover:text-white transition-colors">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Admin Invitation -->
      <div v-if="mode === 'admin'" class="space-y-4">
        <div class="space-y-2">
          <label class="text-xs font-bold text-white/50 uppercase tracking-wider">{{ t('members.addMemberModal.adminName') }}</label>
          <input
            v-model="memberName"
            class="w-full bg-surface/80 border border-white/10 rounded-xl px-4 py-2.5 text-white placeholder-white/30 focus:border-primary/50 outline-none transition-all"
            :placeholder="t('members.addMemberModal.adminNamePlaceholder')"
          />
        </div>

        <div class="space-y-2">
          <label class="text-xs font-bold text-white/50 uppercase tracking-wider">{{ t('members.addMemberModal.permissions') }}</label>
          <div class="bg-surface/50 border border-white/10 rounded-xl overflow-hidden">
            <label
              v-for="row in adminPermRows"
              :key="row.key"
              class="flex items-center gap-3 p-3 hover:bg-white/5 cursor-pointer transition-colors border-b border-white/5 last:border-0"
            >
              <input
                type="checkbox"
                v-model="adminPerms[row.key]"
                class="w-4 h-4 rounded border-white/30 text-primary focus:ring-primary/50"
              />
              <div class="flex flex-col">
                <span class="text-xs font-medium text-white/90">{{ t(row.titleKey) }}</span>
                <span class="text-[10px] text-white/30">{{ t(row.descKey) }}</span>
              </div>
            </label>
          </div>
        </div>

        <div class="flex gap-3 pt-4">
          <button
            type="button"
            @click="$emit('close')"
            class="flex-1 py-2.5 bg-white/10 text-white font-medium rounded-xl hover:bg-white/15 transition-colors"
          >
            {{ t('members.addMemberModal.cancel') }}
          </button>
          <button
            type="button"
            @click="handleAddAdmin"
            :disabled="!memberName.trim()"
            :class="[
              'flex-1 py-2.5 font-bold rounded-xl transition-colors',
              memberName.trim()
                ? 'bg-primary text-on-primary hover:bg-primary-hover'
                : 'bg-white/10 text-white/40 cursor-not-allowed'
            ]"
          >
            {{ t('members.addMemberModal.submitAdmin') }}
          </button>
        </div>
      </div>

      <!-- Member（带 CLI）/ AI Assistant / Secretary -->
      <div v-else-if="mode === 'assistant' || mode === 'secretary' || mode === 'member'" class="space-y-4">
        <div v-if="mode === 'member'" class="space-y-2">
          <label class="text-xs font-bold text-white/50 uppercase tracking-wider">{{ t('members.addMemberModal.memberName') }}</label>
          <input
            v-model="memberName"
            type="text"
            class="w-full bg-surface/80 border border-white/10 rounded-xl px-4 py-2.5 text-white placeholder-white/30 focus:border-primary/50 outline-none transition-all"
            :placeholder="t('members.addMemberModal.memberNamePlaceholder')"
          />
        </div>
        <div v-else class="space-y-2">
          <label class="text-xs font-bold text-white/50 uppercase tracking-wider">{{
            t('members.addMemberModal.assistantDisplayName')
          }}</label>
          <input
            v-model="assistantDisplayName"
            type="text"
            class="w-full bg-surface/80 border border-white/10 rounded-xl px-4 py-2.5 text-white placeholder-white/30 focus:border-primary/50 outline-none transition-all"
            :placeholder="t('members.addMemberModal.assistantDisplayNamePlaceholder')"
          />
        </div>
        <div class="space-y-2">
          <label class="text-xs font-bold text-white/50 uppercase tracking-wider">{{ cliPickerLabel }}</label>
          <div class="space-y-2 max-h-64 overflow-y-auto">
            <button
              v-for="model in assistantModels"
              :key="model.id"
              type="button"
              @click="selectedModel = model.id"
              :class="[
                'w-full flex items-center gap-3 p-3 rounded-xl border transition-all',
                selectedModel === model.id
                  ? 'bg-primary/10 border-primary/20'
                  : 'border-transparent hover:bg-white/5'
              ]"
            >
              <div :class="['w-8 h-8 rounded-lg flex items-center justify-center', model.accentClass]">
                <span class="text-white text-xs font-bold">{{ model.initials }}</span>
              </div>
              <span :class="['text-sm font-medium', selectedModel === model.id ? 'text-white' : 'text-white/70']">
                {{ model.label }}
              </span>
              <svg v-if="selectedModel === model.id" class="w-4 h-4 ml-auto text-primary" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
              </svg>
            </button>
          </div>
        </div>

        <div class="flex gap-3 pt-4">
          <button
            type="button"
            @click="$emit('close')"
            class="flex-1 py-2.5 bg-white/10 text-white font-medium rounded-xl hover:bg-white/15 transition-colors"
          >
            {{ t('members.addMemberModal.cancel') }}
          </button>
          <button
            type="button"
            @click="handleInvite"
            :disabled="inviteSubmitDisabled"
            :class="[
              'flex-1 py-2.5 font-bold rounded-xl transition-colors',
              !inviteSubmitDisabled
                ? 'bg-primary text-on-primary hover:bg-primary-hover'
                : 'bg-white/10 text-white/40 cursor-not-allowed'
            ]"
          >
            {{ mode === 'member' ? t('members.addMemberModal.submitMember') : t('members.addMemberModal.submitAdd') }}
          </button>
        </div>
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
    case 'admin':
      return t('members.addMemberModal.titleAdmin')
    case 'member':
      return t('members.addMemberModal.titleMember')
    case 'secretary':
      return t('members.addMemberModal.titleSecretary')
    default:
      return t('members.addMemberModal.titleAssistant')
  }
})

const cliPickerLabel = computed(() => {
  switch (mode.value) {
    case 'secretary':
      return t('members.addMemberModal.selectSecretary')
    case 'member':
      return t('members.addMemberModal.selectMemberCli')
    default:
      return t('members.addMemberModal.selectAssistant')
  }
})

const inviteSubmitDisabled = computed(
  () => !selectedModel.value || (mode.value === 'member' && !memberName.value.trim())
)

const selectedModel = ref('claude')
const memberName = ref('')
/** Custom label for assistant/secretary; empty → use selected model’s i18n label */
const assistantDisplayName = ref('')

const adminPermRows = [
  {
    key: 'fullAccess' as const,
    titleKey: 'members.addMemberModal.permFullAccess',
    descKey: 'members.addMemberModal.permFullAccessDesc'
  },
  {
    key: 'memberManage' as const,
    titleKey: 'members.addMemberModal.permMemberManage',
    descKey: 'members.addMemberModal.permMemberManageDesc'
  },
  {
    key: 'billing' as const,
    titleKey: 'members.addMemberModal.permBilling',
    descKey: 'members.addMemberModal.permBillingDesc'
  }
]

const adminPerms = reactive({
  fullAccess: true,
  memberManage: true,
  billing: false
})

const assistantModels = computed(() => {
  const defs = [
    { id: 'claude', i18nKey: 'members.assistantModels.claude', initials: 'CC', command: 'claude --dangerously-skip-permissions', terminalType: 'claude', accentClass: 'bg-gradient-to-br from-orange-500 to-amber-600' },
    { id: 'gemini', i18nKey: 'members.assistantModels.gemini', initials: 'GM', command: 'gemini', terminalType: 'gemini', accentClass: 'bg-gradient-to-br from-blue-500 to-cyan-600' },
    { id: 'aider', i18nKey: 'members.assistantModels.aider', initials: 'AI', command: 'aider', terminalType: 'aider', accentClass: 'bg-gradient-to-br from-green-500 to-emerald-600' },
    { id: 'cursor', i18nKey: 'members.assistantModels.cursor', initials: 'CA', command: 'cursor-agent', terminalType: 'cursor', accentClass: 'bg-gradient-to-br from-purple-500 to-violet-600' },
    { id: 'shell', i18nKey: 'members.assistantModels.shell', initials: 'SH', command: '/bin/bash', terminalType: 'shell', accentClass: 'bg-gradient-to-br from-gray-500 to-slate-600' }
  ]
  return defs.map((d) => ({
    id: d.id,
    label: t(d.i18nKey),
    initials: d.initials,
    command: d.command,
    terminalType: d.terminalType,
    accentClass: d.accentClass
  }))
})

function handleInvite() {
  const model = assistantModels.value.find((m) => m.id === selectedModel.value)
  if (!model) return
  if (mode.value === 'member') {
    const n = memberName.value.trim()
    if (!n) return
    emit('invite', {
      name: n,
      roleType: 'member',
      command: model.command,
      terminalType: model.terminalType
    })
    return
  }
  const roleType: MemberRole = mode.value === 'secretary' ? 'secretary' : 'assistant'
  const custom = assistantDisplayName.value.trim()
  emit('invite', {
    name: custom || model.label,
    roleType,
    command: model.command,
    terminalType: model.terminalType
  })
}

function handleAddAdmin() {
  if (memberName.value.trim()) {
    emit('invite', {
      name: memberName.value.trim(),
      roleType: 'admin'
    })
  }
}
</script>
