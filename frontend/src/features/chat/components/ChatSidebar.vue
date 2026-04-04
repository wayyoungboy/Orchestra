<template>
  <aside class="chat-sidebar-root">
    <!-- Workspace Header -->
    <div class="sidebar-header">
      <h2 class="workspace-title truncate" :title="workspaceName || t('chat.workspaceFallback')">
        {{ workspaceName || t('chat.workspaceFallback') }}
      </h2>
    </div>

    <!-- Conversations List -->
    <div class="sidebar-content custom-scrollbar">
      <!-- Channels Section -->
      <div class="section-container">
        <div class="section-header">
          <h3 class="section-title">{{ t('chat.channels') }}</h3>
          <button
            @click="$emit('new-conversation')"
            class="add-btn"
            title="New Channel"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
          </button>
        </div>

        <div class="list-container">
          <button
            v-for="conv in channelConversations"
            :key="conv.id"
            @click="$emit('select', conv.id)"
            :class="['channel-item group', conv.id === activeConversationId ? 'is-active' : '']"
          >
            <span class="hash-icon">#</span>
            <div class="channel-info">
              <span class="channel-name">{{ getConversationTitle(conv) }}</span>
              <!-- Unread Count -->
              <span
                v-if="conv.unreadCount && conv.unreadCount > 0"
                class="unread-badge"
              >
                {{ conv.unreadCount > 99 ? '99+' : conv.unreadCount }}
              </span>
            </div>
          </button>
        </div>
      </div>

      <!-- Direct Messages Section -->
      <div class="section-container">
        <div class="section-header">
          <h3 class="section-title">{{ t('chat.directMessages') }}</h3>
        </div>

        <div class="list-container">
          <button
            v-for="conv in dmConversations"
            :key="conv.id"
            @click="$emit('select', conv.id)"
            :class="['dm-item group', conv.id === activeConversationId ? 'is-active' : '']"
          >
            <div class="dm-avatar-wrap">
              <div class="dm-avatar">
                <span>{{ getInitials(getConversationTitle(conv)) }}</span>
              </div>
              <div class="online-indicator"></div>
            </div>
            <div class="channel-info">
              <span class="channel-name">{{ getConversationTitle(conv) }}</span>
              <!-- Unread Count -->
              <span
                v-if="conv.unreadCount && conv.unreadCount > 0"
                class="unread-badge"
              >
                {{ conv.unreadCount > 99 ? '99+' : conv.unreadCount }}
              </span>
            </div>
          </button>
        </div>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Conversation } from '@/shared/types/chat'

const { t } = useI18n()

const props = defineProps<{
  conversations: Conversation[]
  activeConversationId: string
  workspaceName?: string
}>()

defineEmits<{
  (e: 'select', id: string): void
  (e: 'new-conversation'): void
}>()

const channelConversations = computed(() =>
  props.conversations.filter((c) => c.type === 'channel' || c.type === 'thread')
)

const dmConversations = computed(() => props.conversations.filter((c) => c.type === 'dm'))

function getConversationTitle(conv: Conversation): string {
  if (conv.customName) return conv.customName
  if (conv.nameKey) return conv.nameKey
  return conv.id
}

function getInitials(name: string): string {
  return name ? name.charAt(0).toUpperCase() : '?'
}
</script>

<style scoped>
.chat-sidebar-root {
  width: 260px;
  height: 100%;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.45);
  backdrop-filter: blur(40px);
  -webkit-backdrop-filter: blur(40px);
  border-radius: 32px;
  border: 1px solid rgba(255, 255, 255, 0.8);
  box-shadow: 0 40px 80px -15px rgba(0, 0, 0, 0.04);
  padding: 24px;
  gap: 24px;
}

.sidebar-header {
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(15, 23, 42, 0.06);
  display: flex;
  align-items: center;
}

.workspace-title {
  font-size: 15px;
  font-weight: 800;
  color: #0f172a;
}

.sidebar-content {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 32px;
  padding-right: 4px; /* for scrollbar */
}

.section-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 8px;
}

.section-title {
  font-size: 11px;
  font-weight: 900;
  color: #64748b;
  letter-spacing: 0.15em;
  text-transform: uppercase;
}

.add-btn {
  color: #94a3b8;
  padding: 4px;
  border-radius: 8px;
  transition: all 0.2s;
}

.add-btn:hover {
  background: rgba(15, 23, 42, 0.05);
  color: #0f172a;
}

.list-container {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.channel-item, .dm-item {
  width: 100%;
  height: 44px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 12px;
  border-radius: 14px;
  border: 1px solid transparent;
  transition: all 0.2s ease-out;
  cursor: pointer;
  background: transparent;
}

.channel-item:hover, .dm-item:hover {
  background: rgba(255, 255, 255, 0.6);
}

.channel-item.is-active, .dm-item.is-active {
  background: white;
  border-color: rgba(99, 102, 241, 0.15);
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.02), 0 2px 4px rgba(99, 102, 241, 0.05);
}

.hash-icon {
  font-size: 18px;
  font-weight: 700;
  color: #94a3b8;
  transition: color 0.2s;
}

.channel-item:hover .hash-icon { color: #64748b; }
.channel-item.is-active .hash-icon { color: #4f46e5; font-weight: 900; }

.channel-info {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.channel-name {
  font-size: 14px;
  font-weight: 600;
  color: #475569;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: color 0.2s;
}

.channel-item:hover .channel-name, .dm-item:hover .channel-name {
  color: #0f172a;
}

.channel-item.is-active .channel-name, .dm-item.is-active .channel-name {
  color: #4f46e5;
  font-weight: 800;
}

.unread-badge {
  margin-left: auto;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: 10px;
  background: #4f46e5;
  color: white;
  font-size: 10px;
  font-weight: 900;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(79, 70, 229, 0.3);
}

.dm-avatar-wrap {
  position: relative;
  flex-shrink: 0;
}

.dm-avatar {
  width: 24px;
  height: 24px;
  border-radius: 8px;
  background: #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 800;
  color: #64748b;
  transition: all 0.2s;
}

.dm-item.is-active .dm-avatar {
  background: rgba(99, 102, 241, 0.1);
  color: #4f46e5;
}

.online-indicator {
  position: absolute;
  bottom: -2px;
  right: -2px;
  width: 8px;
  height: 8px;
  background: #10b981;
  border-radius: 50%;
  border: 2px solid white;
}

.dm-item.is-active .online-indicator {
  border-color: #f8fafc;
}
</style>
