<template>
  <aside
    :class="[
      'bg-panel/50 border-l border-white/5 shrink-0 flex-col py-6 px-4 h-full',
      variant === 'drawer' ? 'flex w-72' : 'hidden md:flex w-[280px]'
    ]"
  >
    <!-- Header -->
    <div class="mb-6 flex items-center justify-between px-2">
      <h2 class="text-white font-bold text-[15px]">{{ t('members.title') }}</h2>
      <slot name="header-action">
        <button
          type="button"
          class="w-9 h-9 rounded-xl bg-white/10 border border-white/10 text-white/70 hover:bg-white/20 hover:text-white transition-colors flex items-center justify-center"
          :title="t('friends.invite')"
          @click="$emit('open-invite')"
        >
          <svg class="w-[18px] h-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z"
            />
          </svg>
        </button>
      </slot>
    </div>

    <!-- Members List -->
    <div class="space-y-6 overflow-y-auto flex-1">
      <!-- Owners -->
      <div v-if="owners.length">
        <h3 class="text-white/30 text-[10px] font-bold uppercase tracking-widest mb-3 px-2">
          Owner {{ owners.length > 1 ? `(${owners.length})` : '' }}
        </h3>
        <div class="space-y-1">
          <MemberRow
            v-for="member in owners"
            :key="member.id"
            :member="member"
            :current-user-id="currentUserId"
            :menu-open="openMenuId === member.id"
            @toggle-menu="toggleMenu"
            @action="handleAction"
          />
        </div>
      </div>

      <!-- Admins -->
      <div v-if="admins.length">
        <h3 class="text-white/30 text-[10px] font-bold uppercase tracking-widest mb-3 px-2">
          Admin {{ admins.length > 1 ? `(${admins.length})` : '' }}
        </h3>
        <div class="space-y-1">
          <MemberRow
            v-for="member in admins"
            :key="member.id"
            :member="member"
            :current-user-id="currentUserId"
            :menu-open="openMenuId === member.id"
            @toggle-menu="toggleMenu"
            @action="handleAction"
          />
        </div>
      </div>

      <!-- Secretary -->
      <div v-if="secretaries.length">
        <h3 class="text-white/30 text-[10px] font-bold uppercase tracking-widest mb-3 px-2">
          {{ t('members.roleSecretary') }}{{ secretaries.length > 1 ? ` (${secretaries.length})` : '' }}
        </h3>
        <div class="space-y-1">
          <MemberRow
            v-for="member in secretaries"
            :key="member.id"
            :member="member"
            :current-user-id="currentUserId"
            :menu-open="openMenuId === member.id"
            @toggle-menu="toggleMenu"
            @action="handleAction"
          />
        </div>
      </div>

      <!-- Assistants -->
      <div v-if="assistants.length">
        <h3 class="text-white/30 text-[10px] font-bold uppercase tracking-widest mb-3 px-2">
          Assistant {{ assistants.length > 1 ? `(${assistants.length})` : '' }}
        </h3>
        <div class="space-y-1">
          <MemberRow
            v-for="member in assistants"
            :key="member.id"
            :member="member"
            :current-user-id="currentUserId"
            :menu-open="openMenuId === member.id"
            @toggle-menu="toggleMenu"
            @action="handleAction"
          />
        </div>
      </div>

      <!-- Members -->
      <div v-if="membersGroup.length">
        <h3 class="text-white/30 text-[10px] font-bold uppercase tracking-widest mb-3 px-2">
          Member {{ membersGroup.length > 1 ? `(${membersGroup.length})` : '' }}
        </h3>
        <div class="space-y-1">
          <MemberRow
            v-for="member in membersGroup"
            :key="member.id"
            :member="member"
            :current-user-id="currentUserId"
            :menu-open="openMenuId === member.id"
            @toggle-menu="toggleMenu"
            @action="handleAction"
          />
        </div>
      </div>

      <!-- Unknown / legacy rows (empty roleType in DB, etc.) -->
      <div v-if="otherRoles.length">
        <h3 class="text-white/30 text-[10px] font-bold uppercase tracking-widest mb-3 px-2">
          {{ t('members.roleOther') }}{{ otherRoles.length > 1 ? ` (${otherRoles.length})` : '' }}
        </h3>
        <div class="space-y-1">
          <MemberRow
            v-for="member in otherRoles"
            :key="member.id"
            :member="member"
            :current-user-id="currentUserId"
            :menu-open="openMenuId === member.id"
            @toggle-menu="toggleMenu"
            @action="handleAction"
          />
        </div>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import MemberRow from '@/features/members/MemberRow.vue'
import type { Member, MemberRole, MemberStatus } from '@/shared/types/member'

const { t } = useI18n()

const props = defineProps<{
  members: Member[]
  currentUserId: string
  variant?: 'sidebar' | 'drawer'
}>()

const emit = defineEmits<{
  (e: 'mention', memberId: string): void
  (e: 'open-invite'): void
  (e: 'member-action', payload: ActionPayload): void
}>()

const variant = computed(() => props.variant ?? 'sidebar')
const openMenuId = ref<string | null>(null)

const owners = computed(() => props.members.filter((m) => m.roleType === 'owner'))
const admins = computed(() => props.members.filter((m) => m.roleType === 'admin'))
const secretaries = computed(() => props.members.filter((m) => m.roleType === 'secretary'))
const assistants = computed(() => props.members.filter((m) => m.roleType === 'assistant'))
const membersGroup = computed(() => props.members.filter((m) => m.roleType === 'member'))

const knownMemberRoles = new Set<MemberRole>(['owner', 'admin', 'secretary', 'assistant', 'member'])
const otherRoles = computed(() => props.members.filter((m) => !knownMemberRoles.has(m.roleType as MemberRole)))

function toggleMenu(member: Member) {
  openMenuId.value = openMenuId.value === member.id ? null : member.id
}

interface ActionPayload {
  action: string
  member: Member
  status?: MemberStatus
}

function handleAction(payload: ActionPayload) {
  openMenuId.value = null

  switch (payload.action) {
    case 'mention':
      emit('mention', payload.member.id)
      break
    default:
      emit('member-action', payload)
      break
  }
}
</script>