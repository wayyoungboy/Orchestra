<template>
  <div class="h-full flex flex-col">
    <!-- Header -->
    <header class="px-6 py-4 border-b border-white/5">
      <div class="flex items-center justify-between">
        <h1 class="text-lg font-bold text-white">{{ t('members.title') }}</h1>
        <div class="relative">
          <button
            type="button"
            data-invite-menu-toggle
            class="w-9 h-9 rounded-xl border flex items-center justify-center transition-colors"
            :class="
              showInviteMenu
                ? 'bg-primary/20 text-primary border-primary/30'
                : 'bg-white/10 border-white/10 text-white/70 hover:bg-white/20 hover:text-white'
            "
            :title="t('friends.invite')"
            @click.stop="showInviteMenu = !showInviteMenu"
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
          <template v-if="showInviteMenu">
            <div
              class="fixed inset-0 bg-background/60 backdrop-blur-[2px] z-40"
              aria-hidden="true"
              @click="showInviteMenu = false"
            />
            <InviteMenu position-class="absolute right-0 top-full mt-2 z-50" @select="handleInviteMenuSelect" />
          </template>
        </div>
      </div>
    </header>

    <!-- Members List -->
    <div class="flex-1 overflow-y-auto px-6 py-4">
      <div v-if="projectStore.loading" class="flex items-center justify-center h-full">
        <div class="text-white/40">{{ t('members.loading') }}</div>
      </div>

      <div v-else-if="projectStore.error" class="flex items-center justify-center h-full">
        <div class="text-red-400">{{ projectStore.error }}</div>
      </div>

      <div v-else class="space-y-8 pb-6">
        <section
          v-for="sec in roleSections"
          :key="sec.key"
          class="rounded-xl border border-white/10 bg-white/[0.03] px-4 py-4"
        >
          <h2 class="text-white/30 text-[10px] font-bold uppercase tracking-widest mb-3">
            {{ t(sec.titleKey) }}
            <span class="tabular-nums text-white/45">({{ sec.members.length }})</span>
          </h2>
          <div v-if="sec.members.length === 0" class="py-6 text-center text-sm text-white/30">
            {{ t(sec.emptyKey) }}
          </div>
          <div v-else class="space-y-2">
            <MemberRow
              v-for="member in sec.members"
              :key="member.id"
              :member="member"
              :current-user-id="currentUserId"
              :menu-open="openMenuId === member.id"
              @toggle-menu="toggleMenu"
              @action="handleAction"
            />
          </div>
        </section>
      </div>
    </div>

    <!-- Add Member Modal -->
    <AddMemberModal
      v-if="showAddModal"
      :mode="showAddModal"
      @close="showAddModal = null"
      @invite="handleInvite"
    />

    <!-- Edit Member Modal -->
    <EditMemberModal
      v-if="editingMember"
      :member="editingMember"
      :show-remove="editingMember.id !== currentUserId"
      @close="editingMember = null"
      @save="handleSave"
      @remove="handleRemove"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useProjectStore } from '@/features/workspace/projectStore'
import { useChatStore } from '@/features/chat/chatStore'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'
import { useTerminalMemberStore, hasTerminalConfig } from '@/features/terminal/terminalMemberStore'
import MemberRow from './MemberRow.vue'
import AddMemberModal from './AddMemberModal.vue'
import EditMemberModal from './EditMemberModal.vue'
import InviteMenu from '@/features/chat/components/InviteMenu.vue'
import type { Member, MemberRole, MemberStatus } from '@/shared/types/member'

const { t } = useI18n()
const route = useRoute()
const projectStore = useProjectStore()
const { members: projectMembers } = storeToRefs(projectStore)
const workspaceStore = useWorkspaceStore()

const routeWorkspaceId = computed(() => {
  const p = route.params.id
  if (typeof p === 'string' && p) return p
  if (Array.isArray(p) && p[0]) return p[0]
  return undefined
})
const terminalMemberStore = useTerminalMemberStore()
const chatStore = useChatStore()

const showAddModal = ref<'assistant' | 'admin' | 'member' | 'secretary' | null>(null)
const showInviteMenu = ref(false)

function handleInviteMenuSelect(type: 'admin' | 'secretary' | 'assistant' | 'member') {
  showInviteMenu.value = false
  showAddModal.value = type
}
const editingMember = ref<Member | null>(null)
const openMenuId = ref<string | null>(null)

// Current user ID - for now, we'll use the first owner as current user
const currentUserId = computed(() => {
  const owners = projectMembers.value.filter((m) => m.roleType === 'owner')
  return owners[0]?.id || ''
})

// Group members by role
const owners = computed(() => projectMembers.value.filter((m) => m.roleType === 'owner'))
const admins = computed(() => projectMembers.value.filter((m) => m.roleType === 'admin'))
const secretaries = computed(() => projectMembers.value.filter((m) => m.roleType === 'secretary'))
const assistants = computed(() => projectMembers.value.filter((m) => m.roleType === 'assistant'))
const membersGroup = computed(() => projectMembers.value.filter((m) => m.roleType === 'member'))

const knownMemberRoles = new Set<MemberRole>(['owner', 'admin', 'secretary', 'assistant', 'member'])
const otherRoles = computed(() =>
  projectMembers.value.filter((m) => !knownMemberRoles.has(m.roleType as MemberRole))
)

const roleSections = computed(() => [
  { key: 'owner', titleKey: 'members.roleOwner', emptyKey: 'members.emptyRoleOwner', members: owners.value },
  { key: 'admin', titleKey: 'members.roleAdmin', emptyKey: 'members.emptyRoleAdmin', members: admins.value },
  {
    key: 'secretary',
    titleKey: 'members.roleSecretary',
    emptyKey: 'members.emptyRoleSecretary',
    members: secretaries.value
  },
  {
    key: 'assistant',
    titleKey: 'members.roleAssistant',
    emptyKey: 'members.emptyRoleAssistant',
    members: assistants.value
  },
  { key: 'member', titleKey: 'members.roleMember', emptyKey: 'members.emptyRoleMember', members: membersGroup.value },
  {
    key: 'other',
    titleKey: 'members.roleOther',
    emptyKey: 'members.emptyRoleOther',
    members: otherRoles.value
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

  switch (payload.action) {
    case 'open-terminal':
      if (payload.member) {
        terminalMemberStore.openMemberTerminal(payload.member)
      }
      break
    case 'rename':
      editingMember.value = payload.member
      break
    case 'set-status':
      if (payload.status) {
        projectStore.updateMember(payload.member.id, { manualStatus: payload.status }, routeWorkspaceId.value)
      }
      break
    case 'remove':
      editingMember.value = payload.member
      break
  }
}

async function handleInvite(data: { name: string; roleType: MemberRole; command?: string; terminalType?: string }) {
  const newMember = await projectStore.addMember(
    {
      name: data.name,
      roleType: data.roleType,
      terminalType: data.terminalType,
      terminalCommand: data.command
    },
    routeWorkspaceId.value
  )

  if (newMember) {
    showAddModal.value = null
  }

  // Auto-start terminal when role has CLI config (assistant / secretary / member)
  if (
    newMember &&
    (data.roleType === 'assistant' || data.roleType === 'secretary' || data.roleType === 'member') &&
    hasTerminalConfig(data.terminalType, data.command)
  ) {
    await terminalMemberStore.startMemberSession(newMember, { openTab: true, quietAutostart: true })
  }
}

function handleSave(id: string, name: string) {
  projectStore.updateMember(id, { name }, routeWorkspaceId.value)
  editingMember.value = null
}

async function handleRemove(id: string) {
  const cleaned = await chatStore.deleteMemberConversations(id)
  if (!cleaned) return
  await projectStore.removeMember(id, routeWorkspaceId.value)
  editingMember.value = null
}

// Ensure GET workspace (owner bootstrap) completes before listing members; avoids empty list races.
watch(
  () => routeWorkspaceId.value ?? workspaceStore.currentWorkspace?.id,
  async (workspaceId) => {
    showInviteMenu.value = false
    if (!workspaceId) {
      projectStore.reset()
      return
    }
    await workspaceStore.openWorkspace(workspaceId)
    await projectStore.loadMembers(workspaceId)
  },
  { immediate: true }
)
</script>