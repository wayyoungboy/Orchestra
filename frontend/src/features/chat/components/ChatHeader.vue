<template>
  <header class="chat-header-root">
    <div class="header-left">
      <h1 class="header-title">{{ title }}</h1>
      <p v-if="description" class="header-sub">{{ description }}</p>
    </div>

    <div class="header-right">
      <!-- Connection Status Indicator -->
      <div
        v-if="connectionStatus === 'connected'"
        class="connection-status connected"
        title="已连接"
      >
        <div class="status-dot"></div>
      </div>
      <div
        v-else-if="connectionStatus === 'reconnecting'"
        class="connection-status reconnecting"
        title="重连中..."
      >
        <div class="status-dot"></div>
        <span class="status-text">重连中...</span>
      </div>
      <div
        v-else
        class="connection-status disconnected"
        title="已断线"
      >
        <div class="status-dot"></div>
        <span class="status-text">已断线</span>
        <button
          type="button"
          class="reconnect-btn"
          @click="handleReconnect"
          title="重新连接"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>
      </div>

      <button
        type="button"
        class="action-btn search-btn"
        @click="$emit('open-search')"
        title="Search Messages"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </button>

      <button
        v-if="memberCount !== undefined"
        type="button"
        class="members-btn"
        @click="$emit('open-members')"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
        <span class="member-count">{{ memberCount }}</span>
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useChatStore } from '@/features/chat/chatStore'

const props = defineProps<{
  title: string
  description?: string
  memberCount?: number
}>()

const emit = defineEmits<{
  (e: 'open-members'): void
  (e: 'open-search'): void
}>()

const chatStore = useChatStore()

const connectionStatus = computed(() => chatStore.connectionStatus)

function handleReconnect() {
  chatStore.reconnectChatWebSocket()
}
</script>

<style scoped>
.chat-header-root {
  width: 100%;
  height: 72px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  background: rgba(255, 255, 255, 0.5);
  backdrop-filter: blur(32px);
  -webkit-backdrop-filter: blur(32px);
  border-radius: 24px;
  border: 1px solid white;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.02);
  margin-bottom: 16px;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.header-title {
  font-size: 18px;
  font-weight: 900;
  color: #0f172a;
}

.header-sub {
  font-size: 12px;
  font-weight: 500;
  color: #64748b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 400px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.action-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  background: white;
  color: #64748b;
  transition: all 0.2s;
  cursor: pointer;
}

.action-btn:hover {
  background: #f8fafc;
  color: #4f46e5;
  border-color: #cbd5e1;
}

.members-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  height: 36px;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  background: white;
  color: #64748b;
  transition: all 0.2s;
  cursor: pointer;
}

.members-btn:hover {
  background: #f8fafc;
  color: #0f172a;
  border-color: #cbd5e1;
}

.member-count {
  font-size: 12px;
  font-weight: 700;
  color: #475569;
}

.connection-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  height: 36px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 600;
  border: 1px solid transparent;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.connection-status.connected {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.connection-status.connected .status-dot {
  background: #10b981;
}

.connection-status.reconnecting {
  background: rgba(234, 179, 8, 0.1);
  color: #ca8a04;
  border-color: rgba(234, 179, 8, 0.2);
}

.connection-status.reconnecting .status-dot {
  background: #eab308;
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

.connection-status.disconnected {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border-color: rgba(239, 68, 68, 0.2);
}

.connection-status.disconnected .status-dot {
  background: #ef4444;
}

.status-text {
  line-height: 1;
}

.reconnect-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 4px;
  background: transparent;
  color: #ef4444;
  border: none;
  cursor: pointer;
  transition: all 0.2s;
}

.reconnect-btn:hover {
  background: rgba(239, 68, 68, 0.1);
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
</style>
