<template>
  <div
    :class="[
      'relative flex items-center gap-3 p-2 rounded-xl hover:bg-white/5 group transition-all duration-200',
      menuOpen ? 'z-40' : 'z-0 group-hover:z-30'
    ]"
    @contextmenu.prevent="handleContextMenu"
  >
    <div class="relative shrink-0">
      <div class="w-10 h-10 rounded-full flex items-center justify-center shadow-md" :class="avatarClass">
        <span class="text-sm font-bold" :class="avatarTextClass">{{ initials }}</span>
      </div>
      <div
        :class="[
          'absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-panel',
          statusColor
        ]"
      />
    </div>

    <div class="flex items-center justify-between gap-2 min-w-0 flex-1">
      <div class="flex flex-col min-w-0">
        <div class="flex items-center gap-2 flex-wrap">
          <span
            :class="[
              'text-[13px] leading-none transition-colors',
              member.roleType === 'owner' ? 'text-primary font-bold' : 'text-white font-medium group-hover:text-primary'
            ]"
          >
            {{ displayName }}
          </span>
          <span
            v-if="terminalBadge"
            class="text-[9px] px-1.5 py-0.5 rounded-md border font-bold uppercase tracking-wide shrink-0"
            :class="terminalBadge.cls"
            :title="terminalBadge.title"
          >
            {{ terminalBadge.text }}
          </span>
          <span
            v-if="member.roleType === 'owner'"
            class="px-1.5 py-0.5 bg-yellow-500/20 text-yellow-500 text-[9px] rounded border border-yellow-500/20 font-bold uppercase tracking-wide"
          >
            {{ t('members.roleOwner') }}
          </span>
          <span
            v-if="member.roleType === 'admin'"
            class="px-1.5 py-0.5 bg-primary/20 text-primary text-[9px] rounded border border-primary/20 font-bold uppercase tracking-wide"
          >
            {{ t('members.roleAdmin') }}
          </span>
          <span
            v-if="member.roleType === 'secretary'"
            class="px-1.5 py-0.5 bg-sky-500/15 text-sky-400 text-[9px] rounded border border-sky-500/25 font-bold uppercase tracking-wide"
          >
            {{ t('members.roleSecretary') }}
          </span>
          <span
            v-if="!isKnownMemberRole"
            class="px-1.5 py-0.5 bg-amber-500/15 text-amber-400 text-[9px] rounded border border-amber-500/25 font-bold uppercase tracking-wide"
          >
            {{ t('members.roleOther') }}
          </span>
        </div>
        <span v-if="subtitleSecondary" class="text-white/30 text-[10px] mt-1.5 font-medium truncate block">
          {{ subtitleSecondary }}
        </span>
      </div>

      <div class="relative shrink-0">
        <button
          type="button"
          @click.stop="$emit('toggle-menu', member)"
          :class="[
            'w-8 h-8 rounded-lg flex items-center justify-center transition-colors',
            menuOpen ? 'bg-white/10 text-white' : 'text-white/30 hover:text-white hover:bg-white/10'
          ]"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
          </svg>
        </button>

        <!-- Dropdown Menu -->
        <div
          v-if="menuOpen"
          class="absolute right-0 top-full mt-2 w-52 rounded-xl bg-panel-strong/95 flex flex-col py-1.5 shadow-2xl overflow-hidden z-50 ring-1 ring-white/10"
          @click.stop
        >
          <!-- Open Terminal for assistants -->
          <template v-if="canOpenTerminal">
            <button
              type="button"
              class="w-full text-left px-4 py-2.5 text-xs font-bold text-white hover:bg-white/15 transition-colors flex items-center gap-3"
              @click="$emit('action', { action: 'open-terminal', member })"
            >
              <svg class="w-4 h-4 opacity-70" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              {{ t('memberRow.openTerminal') }}
            </button>
            <div class="h-px bg-white/10 my-1 mx-2" />
          </template>

          <button
            v-if="canSendMessage"
            type="button"
            class="w-full text-left px-4 py-2.5 text-xs font-bold text-white hover:bg-white/15 transition-colors flex items-center gap-3"
            @click="$emit('action', { action: 'send-message', member })"
          >
            <svg class="w-4 h-4 opacity-70" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
            {{ t('memberRow.sendMessage') }}
          </button>

          <button
            v-if="canRename"
            type="button"
            class="w-full text-left px-4 py-2.5 text-xs font-bold text-white hover:bg-white/15 transition-colors flex items-center gap-3"
            @click="$emit('action', { action: 'rename', member })"
          >
            <svg class="w-4 h-4 opacity-70" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
            </svg>
            {{ t('memberRow.rename') }}
          </button>

          <div class="h-px bg-white/10 my-1 mx-2" />

          <div class="px-4 py-1 text-[10px] font-semibold uppercase tracking-wider text-white/40">
            {{ t('memberRow.statusHeading') }}
          </div>

          <button
            v-for="option in statusOptions"
            :key="option.id"
            type="button"
            class="w-full text-left px-4 py-2.5 text-xs font-bold text-white hover:bg-white/15 transition-colors flex items-center gap-3"
            @click="$emit('action', { action: 'set-status', member, status: option.id })"
          >
            <span :class="['w-2.5 h-2.5 rounded-full', option.dotClass]" />
            {{ option.label }}
            <span v-if="coercedStatus === option.id" class="ml-auto text-white/60">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
            </span>
          </button>

          <template v-if="canRemove">
            <div class="h-px bg-white/10 my-1 mx-2" />
            <button
              type="button"
              class="w-full text-left px-4 py-2.5 text-xs font-bold text-red-400 hover:bg-red-500/20 transition-colors flex items-center gap-3"
              @click="$emit('action', { action: 'remove', member })"
            >
              <svg class="w-4 h-4 opacity-70" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7a4 4 0 11-8 0 4 4 0 018 0zM9 14a6 6 0 00-6 6v1h12v-1a6 6 0 00-6-6zM21 12h-6" />
              </svg>
              {{ t('memberRow.remove') }}
            </button>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Member, MemberStatus } from '@/shared/types/member'
import { useContextMenu } from '@/shared/context-menu/useContextMenu'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'
import { useTerminalStore } from '@/features/terminal/terminalStore'
import { hasTerminalConfig, useTerminalMemberStore } from '@/features/terminal/terminalMemberStore'

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
const { openMenu } = useContextMenu()
const workspaceStore = useWorkspaceStore()
const terminalStore = useTerminalStore()
const terminalMemberStore = useTerminalMemberStore()

const knownMemberRoles = new Set(['owner', 'admin', 'secretary', 'assistant', 'member'])
const isKnownMemberRole = computed(() => knownMemberRoles.has(props.member.roleType))
const displayName = computed(() => {
  const n = props.member.name?.trim()
  return n || t('members.unnamedMember')
})

const menuOpen = computed(() => props.menuOpen ?? false)
const canRemove = computed(() => props.currentUserId ? props.member.id !== props.currentUserId : true)
const canRename = computed(() => props.currentUserId ? props.member.id !== props.currentUserId : true)
const canMention = computed(() => props.currentUserId ? props.member.id !== props.currentUserId : true)
const canSendMessage = computed(() => props.currentUserId ? props.member.id !== props.currentUserId : true)
const canOpenTerminal = computed(() =>
  hasTerminalConfig(props.member.terminalType, props.member.terminalCommand)
)

/** 本地 WebSocket 态 + 服务端 PTY 轮询（REQ-303） */
const terminalBadge = computed(() => {
  if (!canOpenTerminal.value) return null
  const wid = workspaceStore.currentWorkspace?.id
  const sess = terminalMemberStore.getSession(props.member.id, wid)
  const server = terminalMemberStore.getServerTerminalForMember(props.member.id)

  if (sess) {
    const st = terminalStore.connectionStatus[sess.terminalId] ?? 'pending'
    const styles: Record<string, { text: string; title: string; cls: string }> = {
      pending: {
        text: '···',
        title: t('memberRow.badgeWsPending'),
        cls: 'border-slate-500/40 text-slate-300 bg-slate-500/10'
      },
      connected: {
        text: 'tty',
        title: t('memberRow.badgeWsConnected'),
        cls: 'border-emerald-500/40 text-emerald-400 bg-emerald-500/10'
      },
      working: {
        text: '»',
        title: t('memberRow.badgeWsWorking'),
        cls: 'border-amber-500/40 text-amber-300 bg-amber-500/10'
      },
      connecting: {
        text: '···',
        title: t('memberRow.badgeWsConnecting'),
        cls: 'border-amber-500/40 text-amber-300 bg-amber-500/10'
      },
      error: {
        text: 'err',
        title: t('memberRow.badgeWsError'),
        cls: 'border-red-500/40 text-red-400 bg-red-500/10'
      },
      disconnected: {
        text: 'off',
        title: t('memberRow.badgeWsDisconnected'),
        cls: 'border-white/15 text-white/45 bg-white/5'
      }
    }
    return styles[st] ?? styles.pending
  }

  if (server) {
    return {
      text: 'pty',
      title: t('memberRow.badgeServerPty'),
      cls: 'border-sky-500/40 text-sky-300 bg-sky-500/10'
    }
  }

  return {
    text: 'tty',
    title: t('memberRow.badgeNoSession'),
    cls: 'border-white/10 text-white/35 bg-white/5'
  }
})

const handleContextMenu = (event: MouseEvent) => {
  const entries: { id: string; label: string; icon?: string; danger?: boolean; action?: () => void }[] = []

  // Open Terminal for assistants
  if (canOpenTerminal.value) {
    entries.push({
      id: 'open-terminal',
      label: t('memberRow.openTerminal'),
      icon: 'terminal',
      action: () => emit('action', { action: 'open-terminal', member: props.member })
    })
  }

  if (canSendMessage.value) {
    entries.push({
      id: 'send-message',
      label: t('memberRow.sendMessage'),
      icon: 'chat_bubble',
      action: () => emit('action', { action: 'send-message', member: props.member })
    })
  }

  if (canMention.value) {
    entries.push({
      id: 'mention',
      label: `@${displayName.value}`,
      icon: 'alternate_email',
      action: () => emit('action', { action: 'mention', member: props.member })
    })
  }

  if (canRename.value) {
    entries.push({
      id: 'rename',
      label: t('memberRow.rename'),
      icon: 'edit',
      action: () => emit('action', { action: 'rename', member: props.member })
    })
  }

  if (canRemove.value) {
    entries.push({
      id: 'remove',
      label: t('memberRow.remove'),
      icon: 'person_remove',
      danger: true,
      action: () => emit('action', { action: 'remove', member: props.member })
    })
  }

  openMenu(event.clientX, event.clientY, entries)
}

const initials = computed(() => {
  const n = displayName.value.trim()
  if (!n) return '?'
  const ch = n.charAt(0).toUpperCase()
  return /[\w\u4e00-\u9fff]/.test(ch) ? ch : '?'
})

/** 后端可能省略 status 或旧数据异常；UI 上统一映射到四种 presence，避免副标题空白 */
const coercedStatus = computed((): MemberStatus => {
  const raw = props.member.manualStatus ?? props.member.status
  if (raw === 'online' || raw === 'working' || raw === 'dnd' || raw === 'offline') {
    return raw
  }
  return 'online'
})

const statusOptions = computed(() => [
  { id: 'online' as MemberStatus, label: t('memberRow.statusOnline'), dotClass: 'bg-green-500' },
  { id: 'working' as MemberStatus, label: t('memberRow.statusWorking'), dotClass: 'bg-amber-400' },
  { id: 'dnd' as MemberStatus, label: t('memberRow.statusDnd'), dotClass: 'bg-red-500' },
  { id: 'offline' as MemberStatus, label: t('memberRow.statusOffline'), dotClass: 'bg-slate-500' }
])

const statusColor = computed(() => {
  const option = statusOptions.value.find((o) => o.id === coercedStatus.value)
  return option?.dotClass || 'bg-slate-500'
})

const avatarClass = computed(() => {
  switch (props.member.roleType) {
    case 'owner':
      return 'bg-primary/20'
    case 'admin':
      return 'bg-white/10'
    case 'assistant':
      return 'bg-emerald-500/20'
    case 'secretary':
      return 'bg-sky-500/20'
    default:
      return 'bg-white/10'
  }
})

const avatarTextClass = computed(() => {
  switch (props.member.roleType) {
    case 'owner':
      return 'text-white'
    case 'admin':
      return 'text-white/60'
    case 'assistant':
      return 'text-emerald-400'
    case 'secretary':
      return 'text-sky-400'
    default:
      return 'text-white/60'
  }
})

const terminalStateLabel = computed(() => {
  if (!canOpenTerminal.value) return ''
  const wid = workspaceStore.currentWorkspace?.id
  const sess = terminalMemberStore.getSession(props.member.id, wid)
  const server = terminalMemberStore.getServerTerminalForMember(props.member.id)
  if (sess) {
    const st = terminalStore.connectionStatus[sess.terminalId] ?? 'pending'
    switch (st) {
      case 'pending':
        return t('memberRow.terminalLinePending')
      case 'connecting':
        return t('memberRow.terminalLineConnecting')
      case 'error':
        return t('memberRow.terminalLineError')
      case 'disconnected':
        return t('memberRow.terminalLineDisconnected')
      case 'working':
        return t('memberRow.terminalLineWorking')
      case 'connected':
        return t('memberRow.terminalLineConnected')
      default:
        return t('memberRow.terminalLinePending')
    }
  }
  if (server) return t('memberRow.terminalLineServerPty')
  return t('memberRow.terminalLineIdle')
})

const subtitleSecondary = computed(() => {
  const presence = statusOptions.value.find((o) => o.id === coercedStatus.value)?.label ?? ''

  if (canOpenTerminal.value) {
    const term = terminalStateLabel.value
    if (term) return presence ? `${presence} · ${term}` : term
    return presence
  }

  const roleTag = (): string => {
    switch (props.member.roleType) {
      case 'assistant':
        return t('members.roleAssistant')
      case 'secretary':
        return t('members.roleSecretary')
      case 'member':
        return t('members.roleMember')
      case 'admin':
        return t('members.roleAdmin')
      case 'owner':
        return ''
      default:
        return isKnownMemberRole.value ? '' : t('members.roleOther')
    }
  }

  const tag = roleTag()
  if (tag && presence) return `${presence} · ${tag}`
  if (tag) return tag
  return presence
})
</script>