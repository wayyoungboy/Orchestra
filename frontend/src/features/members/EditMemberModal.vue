<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
    <div class="w-full max-w-sm bg-panel/90 border border-white/10 rounded-2xl shadow-2xl p-6">
      <div class="flex justify-between items-center mb-6">
        <h3 class="text-white font-bold text-lg">{{ t('members.editMemberModal.title') }}</h3>
        <button type="button" @click="$emit('close')" class="text-white/40 hover:text-white transition-colors">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="mb-6 text-center">
        <div class="w-16 h-16 rounded-full bg-primary/20 flex items-center justify-center mx-auto mb-3">
          <span class="text-2xl font-bold text-primary">{{ member.name.charAt(0).toUpperCase() }}</span>
        </div>
        <p class="text-white/40 text-xs uppercase font-bold tracking-widest">{{ roleLabel }}</p>
      </div>

      <div class="space-y-4">
        <div>
          <label class="text-xs font-bold text-white/50 uppercase tracking-wider mb-1.5 block">{{
            t('members.editMemberModal.displayName')
          }}</label>
          <input
            v-model="name"
            class="w-full bg-surface/80 border border-white/10 rounded-xl px-4 py-2.5 text-white focus:border-primary/50 focus:ring-1 focus:ring-primary/50 outline-none transition-all"
          />
        </div>

        <button type="button" @click="handleSave" class="w-full py-2.5 bg-primary text-on-primary font-bold rounded-xl hover:bg-primary-hover transition-colors">
          {{ t('members.editMemberModal.save') }}
        </button>

        <template v-if="showRemove">
          <div class="h-px bg-white/5 my-2"></div>

          <button type="button" @click="handleRemove" class="w-full py-2.5 text-red-400 hover:bg-red-500/10 font-medium rounded-xl transition-colors flex items-center justify-center gap-2">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
  (e: 'save', id: string, name: string): void
  (e: 'remove', id: string): void
}>()

const name = ref(props.member.name)

const roleLabel = computed(() => {
  switch (props.member.roleType) {
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
      return t('members.roleMember')
  }
})

watch(() => props.member, (newMember) => {
  name.value = newMember.name
})

function handleSave() {
  emit('save', props.member.id, name.value)
}

function handleRemove() {
  emit('remove', props.member.id)
}
</script>
