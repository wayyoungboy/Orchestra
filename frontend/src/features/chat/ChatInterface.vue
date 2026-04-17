<template>
  <div class="chat-interface-root">
    <!-- Sub Sidebar: Channels & DMs -->
    <ChatSidebar
      :conversations="chatStore.sortedConversations"
      :active-conversation-id="chatStore.activeConversationId || ''"
      :workspace-name="currentWorkspace?.name"
      @select="handleSelectConversation"
      @new-conversation="handleNewConversation"
    />

    <!-- Main Chat Content -->
    <div class="chat-main-container">
      <template v-if="activeConversation">
        <ChatHeader
          :title="getConversationTitle(activeConversation)"
          :description="activeConversation.descriptionKey"
          :member-count="chatMembers.length"
          @open-members="toggleMembersSidebar"
          @open-search="showSearchPanel = true"
        />

        <div class="messages-viewport">
          <MessagesList
            ref="messagesListRef"
            :messages="activeConversation.messages || []"
            :current-user-id="chatStore.currentUserId"
            :members="chatMembers"
            :loading="chatStore.loading"
            :loading-messages="chatStore.loadingMessages"
          />
          
          <!-- AI Status Floating Indicator -->
          <Transition name="slide-up">
            <div v-if="currentAgentStatus" class="ai-status-indicator">
              <div class="status-glass">
                <div class="status-pulse" :class="currentAgentStatus.status"></div>
                <span class="status-msg">{{ currentAgentStatus.message || statusLabel(currentAgentStatus.status) }}</span>
              </div>
            </div>
          </Transition>
        </div>

        <ChatInput
          v-model="newMessage"
          :placeholder="t('chat.inputPlaceholder', { name: getConversationTitle(activeConversation) })"
          :members="chatMembers"
          @send="handleSendMessage"
          @input="handleTyping"
        />
      </template>

      <!-- No Active Conversation Empty State -->
      <div v-else class="chat-empty-state">
        <div class="empty-glass-card">
          <div class="empty-icon">
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
          </div>
          <h3>{{ t('chat.emptyStateTitle') }}</h3>
          <p>{{ t('chat.emptyStateDesc') }}</p>
        </div>
      </div>
    </div>

    <!-- Right Sidebar: Members (Permanent on desktop) -->
    <MembersSidebar
      v-if="showMembersSidebar"
      :members="chatMembers"
      :current-user-id="chatStore.currentUserId"
      @mention="handleMentionMember"
      @close="showMembersSidebar = false"
    />

    <!-- Global Search Panel -->
    <div v-if="showSearchPanel" class="search-overlay" @click="showSearchPanel = false">
      <div class="search-modal" @click.stop>
        <div class="search-header">
          <svg class="search-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
          <input v-model="searchQuery" :placeholder="t('chat.searchPlaceholder')" autofocus @input="handleSearch" />
          <button class="close-search" @click="showSearchPanel = false">Esc</button>
        </div>
        <div class="search-results custom-scrollbar">
          <div v-if="!searchQuery" class="search-hint">{{ t('chat.searchHint') }}</div>
          <div v-for="res in workspaceStore.searchResults" :key="res.id" class="search-item">
            <div class="res-meta">
              <span class="res-sender">{{ res.senderName }}</span>
              <span class="res-time">{{ new Date(res.createdAt).toLocaleDateString() }}</span>
            </div>
            <p class="res-content">{{ res.content }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Conversation Modal -->
    <CreateConversationModal
      v-if="showCreateConvModal"
      :members="chatMembers"
      :current-user-id="chatStore.currentUserId"
      @close="showCreateConvModal = false"
      @create="handleCreateConversation"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { useChatStore } from './chatStore'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'
import { useProjectStore } from '@/features/workspace/projectStore'
import ChatSidebar from './components/ChatSidebar.vue'
import ChatHeader from './components/ChatHeader.vue'
import MessagesList from './components/MessagesList.vue'
import ChatInput from './components/ChatInput.vue'
import MembersSidebar from './components/MembersSidebar.vue'
import CreateConversationModal from './components/CreateConversationModal.vue'

const { t } = useI18n()
const chatStore = useChatStore()
const workspaceStore = useWorkspaceStore()
const projectStore = useProjectStore()
const { currentWorkspace } = storeToRefs(workspaceStore)
const { activeConversation } = storeToRefs(chatStore)
const { getConversationTitle } = chatStore

const newMessage = ref('')
const showMembersSidebar = ref(true)
const showSearchPanel = ref(false)
const showCreateConvModal = ref(false)
const searchQuery = ref('')
const messagesListRef = ref<any>(null)

// Computed members for reactivity - returns the array directly
const chatMembers = computed(() => {
  const m = projectStore.members
  return m && m.length > 0 ? m : []
})

// AI Status Logic
const currentAgentStatus = computed(() => {
  if (!activeConversation.value) return null
  const assistantId = activeConversation.value.memberIds?.find((id: string) => id.includes('assistant') || id.startsWith('ai-'))
  return assistantId ? chatStore.agentStatuses[assistantId] : null
})

function statusLabel(status: string) {
  const key = `chat.status.${status}`
  const translated = t(key)
  return translated === key ? status : translated
}

function toggleMembersSidebar() {
  showMembersSidebar.value = !showMembersSidebar.value
}

async function handleSelectConversation(id: string) {
  await chatStore.setActiveConversation(id)
  await nextTick()
  messagesListRef.value?.jumpToLatest?.()
}

function handleNewConversation() { showCreateConvModal.value = true }

async function handleCreateConversation(data: { type: 'channel' | 'dm'; name?: string; memberId?: string }) {
  showCreateConvModal.value = false
  const newId = await chatStore.createConversation(data)
  if (newId) {
    await chatStore.setActiveConversation(newId)
    await nextTick()
    messagesListRef.value?.jumpToLatest?.()
  }
}

async function handleSendMessage() {
  if (!newMessage.value.trim() || !activeConversation.value) return
  await chatStore.sendMessage({ text: newMessage.value, conversationId: activeConversation.value.id })
  newMessage.value = ''
  await nextTick()
  messagesListRef.value?.jumpToLatest?.()
}

function handleTyping() {
  if (activeConversation.value) chatStore.updatePresence('typing', activeConversation.value.id)
}

// Handle @ mention from MembersSidebar
function handleMentionMember(memberId: string) {
  const member = chatMembers.value.find(m => m.id === memberId)
  if (member) {
    newMessage.value = newMessage.value + `@${member.name} `
  }
}

let searchTimer: any = null
function handleSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => workspaceStore.searchWorkspace(searchQuery.value), 300)
}

onMounted(async () => {
  if (currentWorkspace.value) {
    await chatStore.loadConversations(currentWorkspace.value.id)
    await projectStore.loadMembers(currentWorkspace.value.id)
  }
})

onBeforeUnmount(() => {
  chatStore.disconnectChatWebSocket()
})
</script>

<style scoped>
.chat-interface-root { display: flex; height: 100%; width: 100%; gap: 12px; }
.chat-main-container { flex: 1; display: flex; flex-direction: column; background: rgba(255, 255, 255, 0.45); backdrop-filter: blur(32px); border-radius: 32px; border: 1px solid white; box-shadow: 0 20px 50px rgba(0, 0, 0, 0.04); overflow: hidden; position: relative; }

.messages-viewport { flex: 1; position: relative; overflow: hidden; display: flex; flex-direction: column; }

.ai-status-indicator { position: absolute; bottom: 12px; left: 24px; z-index: 30; }
.status-glass { background: rgba(255, 255, 255, 0.9); backdrop-filter: blur(12px); padding: 8px 16px; border-radius: 100px; display: flex; align-items: center; gap: 10px; border: 1px solid #e2e8f0; box-shadow: 0 10px 25px rgba(0,0,0,0.05); }
.status-pulse { width: 8px; height: 8px; border-radius: 50%; background: #6366f1; }
.status-pulse.thinking { animation: pulse-thinking 1.5s infinite; background: #8b5cf6; }
@keyframes pulse-thinking { 0% { transform: scale(0.8); opacity: 0.5; } 50% { transform: scale(1.2); opacity: 1; } 100% { transform: scale(0.8); opacity: 0.5; } }
.status-msg { font-size: 12px; font-weight: 700; color: #475569; }

.search-overlay { position: absolute; inset: 0; background: rgba(15, 23, 42, 0.1); backdrop-blur: 4px; z-index: 100; display: flex; justify-content: center; padding-top: 80px; }
.search-modal { width: 600px; max-height: 500px; background: rgba(255, 255, 255, 0.95); backdrop-filter: blur(40px); border-radius: 24px; border: 1px solid white; box-shadow: 0 40px 100px rgba(0,0,0,0.15); display: flex; flex-direction: column; overflow: hidden; }
.search-header { display: flex; align-items: center; padding: 20px 24px; border-bottom: 1px solid #f1f5f9; gap: 16px; }
.search-icon { width: 24px; height: 24px; color: #94a3b8; }
.search-header input { flex: 1; background: transparent; border: none; outline: none; font-size: 18px; font-weight: 600; color: #0f172a; }
.close-search { font-size: 11px; font-weight: 900; color: #94a3b8; padding: 4px 10px; background: #f1f5f9; border-radius: 8px; }
.search-results { flex: 1; overflow-y: auto; padding: 12px; }
.search-item { padding: 16px; border-radius: 12px; cursor: pointer; }
.search-item:hover { background: #f8fafc; }
.res-meta { display: flex; justify-content: space-between; margin-bottom: 4px; }
.res-sender { font-size: 12px; font-weight: 800; color: #4f46e5; }
.res-content { font-size: 14px; color: #475569; line-height: 1.5; }

.chat-empty-state { flex: 1; display: flex; align-items: center; justify-content: center; text-align: center; padding: 40px; }
.empty-glass-card { text-align: center; max-width: 320px; }
.empty-icon { width: 64px; height: 64px; background: #f1f5f9; border-radius: 20px; display: flex; align-items: center; justify-content: center; margin: 0 auto 24px; color: #94a3b8; }
.empty-glass-card h3 { font-size: 20px; font-weight: 900; color: #0f172a; margin-bottom: 12px; }
.empty-glass-card p { font-size: 15px; color: #64748b; line-height: 1.6; }
</style>
