<template>
  <aside
    :class="[
      'members-sidebar-root',
      variant === 'drawer' ? 'is-drawer' : 'is-sidebar'
    ]"
  >
    <!-- Header -->
    <div class="sidebar-header">
      <h2 class="header-title">{{ t('members.title') }}</h2>
      <slot name="header-action">
        <button
          type="button"
          class="invite-btn"
          :title="t('friends.invite')"
          @click="$emit('open-invite')"
        >
          <svg class="w-4.5 h-4.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2.5"
              d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z"
            />
          </svg>
        </button>
      </slot>
    </div>

    <!-- Members List -->
    <div class="sidebar-content custom-scrollbar">
      <!-- Role Sections -->
      <div v-for="section in roleSections" :key="section.label" class="section-container">
        <div v-if="section.members.length">
          <div class="section-label">
            <span>{{ section.label }}</span>
            <span class="count-badge">{{ section.members.length }}</span>
          </div>
          <div class="member-list">
            <MemberRow
              v-for="member in section.members"
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
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import MemberRow from '@/features/members/MemberRow.vue'
import type { Member, MemberStatus } from '@/shared/types/member'

const { t } = useI18n()

const props = defineProps<{
  members: Member[]
  currentUserId: string
  variant?: 'sidebar' | 'drawer'
}>()

const emit = defineEmits<{
  (e: 'open-invite'): void
  (e: 'member-action', payload: ActionPayload): void
}>()

const variant = computed(() => props.variant ?? 'sidebar')
const openMenuId = ref<string | null>(null)

const roleSections = computed(() => [
  { label: 'Owners', members: props.members.filter(m => m.roleType === 'owner') },
  { label: 'Secretaries', members: props.members.filter(m => m.roleType === 'secretary') },
  { label: 'Assistants', members: props.members.filter(m => m.roleType === 'assistant') },
  {
    label: 'Others',
    members: props.members.filter(m => !['owner', 'secretary', 'assistant'].includes(m.roleType))
  }
])

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
  emit('member-action', payload)
}
</script>

<style scoped>
.members-sidebar-root {
  height: 100%;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.4);
  backdrop-filter: blur(32px);
  -webkit-backdrop-filter: blur(32px);
  border-left: 1px solid rgba(15, 23, 42, 0.05);
  padding: 24px 16px;
  gap: 24px;
}

.is-sidebar { width: 280px; display: none; }
@media (min-width: 768px) { .is-sidebar { display: flex; } }
.is-drawer { width: 280px; display: flex; }

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 8px;
}

.header-title {
  font-size: 15px;
  font-weight: 800;
  color: #0f172a;
}

.invite-btn {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: white;
  border: 1px solid #e2e8f0;
  color: #64748b;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(0,0,0,0.02);
}

.invite-btn:hover {
  background: #f8fafc;
  color: #4f46e5;
  border-color: #cbd5e1;
  transform: translateY(-1px);
}

.sidebar-content {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.section-container {
  display: flex;
  flex-direction: column;
}

.section-label {
  font-size: 10px;
  font-weight: 900;
  color: #94a3b8;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  padding: 0 8px;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.count-badge {
  font-size: 9px;
  background: rgba(15, 23, 42, 0.05);
  padding: 1px 6px;
  border-radius: 6px;
  color: #64748b;
}

.member-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
</style>
