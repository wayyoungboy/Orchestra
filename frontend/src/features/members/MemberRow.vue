<template>
  <div
    :class="[
      'member-row-root group',
      menuOpen ? 'is-menu-open' : ''
    ]"
    @contextmenu.prevent="handleContextMenu"
  >
    <div class="avatar-container">
      <div :class="['avatar-circle', avatarClass]">
        <span :class="['avatar-text', avatarTextClass]">{{ initials }}</span>
      </div>
      <div :class="['status-dot', statusColor]"></div>
    </div>

    <div class="member-main-info">
      <div class="name-badge-row">
        <span :class="['member-display-name', member.roleType === 'owner' ? 'is-owner' : '']">
          {{ displayName }}
        </span>
        
        <!-- Role Badges -->
        <span v-if="member.roleType === 'owner'" class="role-badge owner">{{ t('members.roleOwner') }}</span>
        <span v-if="member.roleType === 'assistant'" class="role-badge assistant">{{ t('members.roleAssistant') }}</span>
        <span v-if="member.roleType === 'secretary'" class="role-badge secretary">{{ t('members.roleSecretary') }}</span>
        
        <!-- Terminal Status Badge -->
        <span v-if="false" :class="['terminal-status-badge']"></span>
      </div>
      <span v-if="subtitleSecondary" class="member-status-line truncate">
        {{ subtitleSecondary }}
      </span>
    </div>

    <div class="member-actions-trigger">
      <button
        type="button"
        @click.stop="$emit('toggle-menu', member)"
        :class="['more-btn', menuOpen ? 'is-active' : '']"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
        </svg>
      </button>

      <!-- Dropdown Menu (Light Glass) -->
      <div v-if="menuOpen" class="member-dropdown-menu animate-in fade-in zoom-in-95 duration-200" @click.stop>
        <div class="menu-group">
          <button v-if="canSendMessage" @click="$emit('action', { action: 'send-message', member })" class="menu-item">
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
            {{ t('memberRow.sendMessage') }}
          </button>
          <button @click="$emit('action', { action: 'mention', member })" class="menu-item">
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 12a4 4 0 10-8 0 4 4 0 008 0zm0 0v1.5a2.5 2.5 0 005 0V12a9 9 0 10-9 9m4.5-1.206a8.959 8.959 0 01-4.5 1.207" /></svg>
            {{ t('memberRow.mention') }}
          </button>
        </div>
        <div class="menu-divider"></div>
        <div class="menu-label">{{ t('memberRow.statusHeading') }}</div>
        <div class="menu-group">
          <button v-for="option in statusOptions" :key="option.id" @click="$emit('action', { action: 'set-status', member, status: option.id })" class="menu-item">
            <span :class="['status-dot-small', option.dotClass]"></span>
            {{ option.label }}
            <svg v-if="coercedStatus === option.id" class="ml-auto w-4 h-4 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
          </button>
        </div>
        <template v-if="canRemove">
          <div class="menu-divider"></div>
          <button @click="$emit('action', { action: 'remove', member })" class="menu-item is-danger">
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7a4 4 0 11-8 0 4 4 0 018 0zM9 14a6 6 0 00-6 6v1h12v-1a6 6 0 00-6-6zM21 12h-6" /></svg>
            {{ t('memberRow.remove') }}
          </button>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Member, MemberStatus } from '@/shared/types/member'

const props = defineProps<{
  member: Member
  menuOpen?: boolean
  currentUserId?: string
}>()

const emit = defineEmits<{
  (e: 'toggle-menu', member: Member): void
  (e: 'action', payload: { action: string; member: Member; status?: MemberStatus }): void
}>()

const { t } = useI18n()

const displayName = computed(() => props.member.name?.trim() || t('members.unnamedMember'))
const menuOpen = computed(() => props.menuOpen ?? false)
const canRemove = computed(() => props.currentUserId ? props.member.id !== props.currentUserId : true)
const canSendMessage = computed(() => props.currentUserId ? props.member.id !== props.currentUserId : true)

const initials = computed(() => {
  const n = displayName.value.trim()
  return n ? n.charAt(0).toUpperCase() : '?'
})

const coercedStatus = computed((): MemberStatus => ((props.member.manualStatus ?? props.member.status) || 'online') as MemberStatus)

const statusOptions = computed(() => [
  { id: 'online' as MemberStatus, label: t('memberRow.statusOnline'), dotClass: 'bg-green-500 shadow-[0_0_8px_rgba(16,185,129,0.4)]' },
  { id: 'working' as MemberStatus, label: t('memberRow.statusWorking'), dotClass: 'bg-amber-400' },
  { id: 'dnd' as MemberStatus, label: t('memberRow.statusDnd'), dotClass: 'bg-red-500' },
  { id: 'offline' as MemberStatus, label: t('memberRow.statusOffline'), dotClass: 'bg-slate-400' }
])

const statusColor = computed(() => statusOptions.value.find(o => o.id === coercedStatus.value)?.dotClass || 'bg-slate-400')

const avatarClass = computed(() => {
  switch (props.member.roleType) {
    case 'owner': return 'bg-indigo-100'
    case 'assistant': return 'bg-green-100'
    case 'secretary': return 'bg-amber-100'
    default: return 'bg-slate-100'
  }
})

const avatarTextClass = computed(() => {
  switch (props.member.roleType) {
    case 'owner': return 'text-indigo-600'
    case 'assistant': return 'text-green-600'
    case 'secretary': return 'text-amber-600'
    default: return 'text-slate-600'
  }
})

const subtitleSecondary = computed(() => {
  // AI members show agent config instead of presence status
  if (props.member.roleType === 'assistant' || props.member.roleType === 'secretary') {
    const agentLabel = props.member.acpEnabled ? (props.member.acpCommand || 'AI Agent') : '未配置'
    return agentLabel
  }
  const presence = statusOptions.value.find(o => o.id === coercedStatus.value)?.label ?? ''
  const roleLabel = () => {
    switch (props.member.roleType) {
      case 'assistant': return t('members.roleAssistant')
      case 'secretary': return t('members.roleSecretary')
      default: return ''
    }
  }
  const tag = roleLabel()
  return tag ? `${presence} · ${tag}` : presence
})

function handleContextMenu(_event: MouseEvent) {
  // Use emitting for simplicity in this visual refactor
  emit('toggle-menu', props.member)
}
</script>

<style scoped>
.member-row-root {
  display: flex; align-items: center; gap: 12px; padding: 10px 12px;
  border-radius: 14px; transition: all 0.2s cubic-bezier(0.23, 1, 0.32, 1);
  cursor: pointer; position: relative;
}
.member-row-root:hover { background: rgba(15, 23, 42, 0.04); }
.member-row-root.is-menu-open { background: rgba(15, 23, 42, 0.06); z-index: 40; }

.avatar-container { position: relative; flex-shrink: 0; }
.avatar-circle { width: 40px; height: 40px; border-radius: 12px; display: flex; align-items: center; justify-content: center; shadow: 0 4px 10px rgba(0,0,0,0.02); }
.avatar-text { font-size: 15px; font-weight: 900; }

.status-dot {
  position: absolute; bottom: -2px; right: -2px; width: 10px; height: 10px;
  border-radius: 50%; border: 2px solid white;
}

.member-main-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.name-badge-row { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }

.member-display-name { font-size: 13.5px; font-weight: 700; color: #334155; }
.member-display-name.is-owner { color: #4f46e5; }

.role-badge {
  font-size: 9px; font-weight: 900; padding: 1px 6px; border-radius: 6px;
  text-transform: uppercase; letter-spacing: 0.05em;
}
.role-badge.owner { background: #fee2e2; color: #ef4444; }
.role-badge.admin { background: #e0e7ff; color: #4338ca; }
.role-badge.assistant { background: #dcfce7; color: #10b981; }
.role-badge.secretary { background: #fef9c3; color: #ca8a04; }

.member-status-line { font-size: 11px; font-weight: 600; color: #94a3b8; }

.member-actions-trigger { position: relative; }
.more-btn {
  width: 28px; height: 28px; border-radius: 8px; display: flex; align-items: center; justify-content: center;
  color: #cbd5e1; transition: all 0.2s; border: none; background: transparent; cursor: pointer;
}
.member-row-root:hover .more-btn { color: #94a3b8; }
.more-btn:hover, .more-btn.is-active { background: white; color: #0f172a; shadow: 0 2px 8px rgba(0,0,0,0.05); }

.member-dropdown-menu {
  position: absolute; top: 100%; right: 0; margin-top: 8px; width: 200px;
  background: rgba(255, 255, 255, 0.95); backdrop-filter: blur(24px);
  border-radius: 16px; border: 1px solid rgba(226, 232, 240, 0.8);
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1); z-index: 100; padding: 6px;
}

.menu-group { display: flex; flex-direction: column; gap: 2px; }
.menu-item {
  width: 100%; padding: 10px 12px; border-radius: 10px; display: flex; align-items: center; gap: 10px;
  font-size: 12px; font-weight: 700; color: #475569; border: none; background: transparent; cursor: pointer;
  transition: all 0.2s;
}
.menu-item:hover { background: #f8fafc; color: #0f172a; }
.menu-item svg { width: 16px; height: 16px; opacity: 0.6; }
.menu-item.is-danger { color: #ef4444; }
.menu-item.is-danger:hover { background: #fef2f2; }

.menu-divider { height: 1px; background: #f1f5f9; margin: 4px 8px; }
.menu-label { font-size: 9px; font-weight: 900; color: #cbd5e1; text-transform: uppercase; padding: 4px 12px; letter-spacing: 0.1em; }

.status-dot-small { width: 8px; height: 8px; border-radius: 50%; }
</style>
