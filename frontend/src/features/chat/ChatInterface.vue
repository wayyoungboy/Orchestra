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
          :member-count="activeConversation.memberIds?.length"
          @open-members="showMembersSidebar = true"
        />

        <MessagesList
          ref="messagesListRef"
          :messages="activeConversation.messages || []"
          :current-user-id="chatStore.currentUserId"
          :loading="chatStore.loading"
        />

        <ChatInput
          v-model="newMessage"
          :placeholder="t('chat.inputPlaceholder', { name: getConversationTitle(activeConversation) })"
          @send="handleSendMessage"
        />
      </template>

      <!-- No Active Conversation Empty State -->
      <div v-else class="chat-empty-state">
        <div class="empty-glass-card">
          <div class="empty-icon">
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
          </div>
          <h3>选择一个频道开始对话</h3>
          <p>从左侧列表中选择一个现有的频道，或者创建一个新的对话。</p>
        </div>
      </div>
    </div>

    <!-- Right Sidebar: Members (Optional Overlay) -->
    <MembersSidebar
      v-if="showMembersSidebar"
      :workspace-id="currentWorkspace?.id || ''"
      :conversation="activeConversation"
      @close="showMembersSidebar = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { useChatStore } from './chatStore'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'
import ChatSidebar from './components/ChatSidebar.vue'
import ChatHeader from './components/ChatHeader.vue'
import MessagesList from './components/MessagesList.vue'
import ChatInput from './components/ChatInput.vue'
import MembersSidebar from './components/MembersSidebar.vue'

const { t } = useI18n()
const chatStore = useChatStore()
const workspaceStore = useWorkspaceStore()
const { currentWorkspace } = storeToRefs(workspaceStore)
const { activeConversation, getConversationTitle } = storeToRefs(chatStore)

const newMessage = ref('')
const showMembersSidebar = ref(false)
const messagesListRef = ref<any>(null)

async function handleSelectConversation(id: string) {
  chatStore.setActiveConversation(id)
  await nextTick()
  messagesListRef.value?.jumpToLatest?.()
}

function handleNewConversation() {
  // To be implemented: open create conversation modal
}

async function handleSendMessage() {
  if (!newMessage.value.trim() || !activeConversation.value) return
  
  await chatStore.sendMessage({
    text: newMessage.value,
    conversationId: activeConversation.value.id
  })
  newMessage.value = ''
  await nextTick()
  messagesListRef.value?.jumpToLatest?.()
}

onMounted(() => {
  if (currentWorkspace.value) {
    chatStore.loadSession()
  }
})
</script>

<style scoped>
.chat-interface-root {
  display: flex;
  height: 100%;
  width: 100%;
  gap: 12px;
}

.chat-main-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.45);
  backdrop-filter: blur(32px);
  border-radius: 32px;
  border: 1px solid white;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.04);
  overflow: hidden;
}

.chat-empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
}

.empty-glass-card {
  text-align: center;
  max-width: 320px;
}

.empty-icon {
  width: 64px;
  height: 64px;
  background: #f1f5f9;
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 24px;
  color: #94a3b8;
}

.empty-glass-card h3 {
  font-size: 18px;
  font-weight: 800;
  color: #0f172a;
  margin-bottom: 8px;
}

.empty-glass-card p {
  font-size: 14px;
  color: #64748b;
  line-height: 1.6;
}
</style>
