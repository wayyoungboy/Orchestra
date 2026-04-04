<template>
  <div class="flex h-full w-full relative">
    <!-- Chat Sidebar -->
    <ChatSidebar
      :conversations="chatStore.sortedConversations"
      :active-conversation-id="chatStore.activeConversationId || ''"
      :workspace-name="workspaceName"
      @select="chatStore.setActiveConversation"
      @new-conversation="handleNewConversation"
    />

    <!-- Main Chat Area -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- Chat Header -->
      <ChatHeader
        v-if="chatStore.activeConversation"
        :title="headerTitle"
        :description="headerDescription"
        :member-count="memberCount"
        @open-members="showMembersDrawer = true"
      />

      <!-- Messages List -->
      <MessagesList
        v-if="chatStore.activeConversation"
        :messages="activeMessages"
        :current-user-id="currentUserId"
        ref="messagesListRef"
      />

      <!-- Empty State -->
      <div
        v-else
        class="flex-1 flex items-center justify-center text-white/50"
      >
        <div class="text-center">
          <div class="w-16 h-16 rounded-full bg-white/5 flex items-center justify-center mx-auto mb-4">
            <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </div>
          <p class="text-sm">{{ t('chat.empty') }}</p>
        </div>
      </div>

      <!-- Chat Input -->
      <ChatInput
        v-if="chatStore.activeConversation"
        v-model="inputValue"
        :placeholder="inputPlaceholder"
        :members="displayMembers"
        @send="handleSendMessage"
      />
    </div>

    <!-- Members Sidebar（邀请入口：单键 + InviteMenu） -->
    <MembersSidebar
      v-if="chatStore.activeConversation && showMembersSidebar"
      :members="displayMembers"
      :current-user-id="currentUserId"
      @mention="handleMention"
      @member-action="handleMemberAction"
    >
      <template #header-action>
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
            @click.stop="toggleInviteMenu"
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
            <InviteMenu position-class="absolute right-0 top-full mt-3 z-50" @select="handleInviteMenuSelect" />
          </template>
        </div>
      </template>
    </MembersSidebar>

    <!-- Add Member Modal -->
    <AddMemberModal
      v-if="showAddModal"
      :mode="showAddModal"
      @close="showAddModal = null"
      @invite="handleInvite"
    />

    <!-- Create Conversation Modal -->
    <CreateConversationModal
      v-if="showCreateModal"
      :members="displayMembers"
      :current-user-id="currentUserId"
      @close="showCreateModal = false"
      @create="handleCreateConversation"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { onTerminalChatStream } from '@/features/terminal/terminalChatBridge'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'
import { useProjectStore } from '@/features/workspace/projectStore'
import { useChatStore } from './chatStore'
import { useTerminalMemberStore, hasTerminalConfig } from '@/features/terminal/terminalMemberStore'
import ChatSidebar from './components/ChatSidebar.vue'
import ChatHeader from './components/ChatHeader.vue'
import MessagesList from './components/MessagesList.vue'
import ChatInput from './components/ChatInput.vue'
import MembersSidebar from './components/MembersSidebar.vue'
import InviteMenu from './components/InviteMenu.vue'
import CreateConversationModal from './components/CreateConversationModal.vue'
import AddMemberModal from '@/features/members/AddMemberModal.vue'
import type { Member, MemberRole, MemberStatus } from '@/shared/types/member'

const { t } = useI18n()
const route = useRoute()
const workspaceStore = useWorkspaceStore()
const projectStore = useProjectStore()

const routeWorkspaceId = computed(() => {
  const p = route.params.id
  if (typeof p === 'string' && p) return p
  if (Array.isArray(p) && p[0]) return p[0]
  return undefined
})
const chatStore = useChatStore()
const terminalMemberStore = useTerminalMemberStore()

// State
const inputValue = ref('')
const showMembersDrawer = ref(false)
const showAddModal = ref<'assistant' | 'admin' | 'member' | 'secretary' | null>(null)
const showInviteMenu = ref(false)

function toggleInviteMenu() {
  showInviteMenu.value = !showInviteMenu.value
}

function handleInviteMenuSelect(type: 'admin' | 'secretary' | 'assistant' | 'member') {
  showInviteMenu.value = false
  showAddModal.value = type
}
const showCreateModal = ref(false)
const messagesListRef = ref<{ jumpToLatest: () => void } | null>(null)

let unsubTerminalChatStream: (() => void) | null = null

// Computed
const workspaceName = computed(() => workspaceStore.currentWorkspace?.name ?? t('chat.workspaceFallback'))
const currentUserId = computed(() => {
  const owner = projectStore.members.find((m) => m.roleType === 'owner')
  return owner?.id ?? 'owner'
})
const headerTitle = computed(() => {
  const conv = chatStore.activeConversation
  if (!conv) return '#general'
  return chatStore.getConversationTitle(conv)
})
const headerDescription = computed(() => {
  const conv = chatStore.activeConversation
  if (!conv) return ''
  return conv.type === 'channel'
    ? t('chat.headerChannelInWorkspace', { name: workspaceName.value })
    : t('chat.headerDirectMessage')
})
const memberCount = computed(() => projectStore.members.length)
const displayMembers = computed(() => projectStore.sortedMembers)
const inputPlaceholder = computed(() => t('chat.inputMessagePlaceholder', { channel: headerTitle.value }))
const showMembersSidebar = computed(() => true) // Can be made responsive

// Convert messages for display
const activeMessages = computed(() => {
  const conv = chatStore.activeConversation
  if (!conv) return []
  return conv.messages.map((msg) => ({
    id: msg.id,
    conversationId: conv.id,
    senderId: msg.senderId || '',
    senderName: msg.senderName,
    senderAvatar: msg.senderAvatar,
    content: typeof msg.content === 'string' ? msg.content : (msg.content.text || ''),
    createdAt: typeof msg.createdAt === 'number' ? new Date(msg.createdAt).toISOString() : msg.createdAt
  }))
})

// Initialize — chat session list is loaded by chatStore when workspace changes; only refresh members here.
onMounted(async () => {
  unsubTerminalChatStream = onTerminalChatStream((sessionId, payload) => {
    chatStore.applyTerminalChatStreamEvent(sessionId, payload)
  })
  const wid = routeWorkspaceId.value ?? workspaceStore.currentWorkspace?.id
  if (wid) {
    await workspaceStore.openWorkspace(wid)
    await projectStore.loadMembers(wid)
  }
})

onBeforeUnmount(() => {
  unsubTerminalChatStream?.()
  unsubTerminalChatStream = null
})

watch(
  () => routeWorkspaceId.value,
  async (wid) => {
    if (wid) {
      await workspaceStore.openWorkspace(wid)
      await projectStore.loadMembers(wid)
    }
  }
)

watch(
  () => chatStore.activeConversationId,
  (convId) => {
    showInviteMenu.value = false
    // Ensure messages are loaded for the active conversation
    if (convId) {
      void chatStore.loadConversationMessages(convId)
    }
  },
  { immediate: true }
)

// Methods
function handleNewConversation() {
  showCreateModal.value = true
}

async function handleCreateConversation(data: { type: 'channel' | 'dm'; name?: string; memberId?: string }) {
  showCreateModal.value = false

  if (data.type === 'dm' && data.memberId) {
    await chatStore.createDirectConversation(data.memberId)
  } else if (data.type === 'channel' && data.name) {
    await chatStore.createConversation([], data.name)
  }
}

async function handleSendMessage() {
  const text = inputValue.value.trim()
  if (!text || !chatStore.activeConversationId) return

  await chatStore.sendMessage({
    text,
    conversationId: chatStore.activeConversationId
  })
  inputValue.value = ''

  // Scroll to bottom
  setTimeout(() => {
    messagesListRef.value?.jumpToLatest()
  }, 100)
}

function handleMention(memberId: string) {
  const member = displayMembers.value.find(m => m.id === memberId)
  if (member) {
    inputValue.value += `@${member.name} `
  }
}

async function handleMemberAction(payload: {
  action: string
  member: Member
  status?: MemberStatus
}) {
  const { action, member } = payload
  switch (action) {
    case 'open-terminal':
      await terminalMemberStore.openMemberTerminal(member)
      break
    case 'send-message': {
      const id = await chatStore.ensureDirectMessage(member.id)
      if (id) await chatStore.setActiveConversation(id)
      break
    }
    case 'rename': {
      const name = window.prompt('Rename member', member.name)?.trim()
      if (name) await projectStore.updateMember(member.id, { name }, routeWorkspaceId.value)
      break
    }
    case 'set-status':
      if (payload.status) {
        await projectStore.updateMember(member.id, { manualStatus: payload.status }, routeWorkspaceId.value)
      }
      break
    case 'remove': {
      if (!window.confirm(`Remove ${member.name}?`)) return
      const cleaned = await chatStore.deleteMemberConversations(member.id)
      if (!cleaned) return
      await projectStore.removeMember(member.id, routeWorkspaceId.value)
      break
    }
    default:
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
</script>