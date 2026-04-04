<template>
  <div
    ref="listRef"
    class="messages-list-container"
  >
    <!-- Empty State -->
    <div v-if="messages.length === 0" class="empty-messages">
      <div class="empty-bubble">还没有消息。开始对话吧！</div>
    </div>

    <!-- Messages Groups -->
    <template v-else>
      <div v-for="(item, index) in groupedMessages" :key="index" class="message-group">
        <!-- Date Divider -->
        <div v-if="item.type === 'divider'" class="date-divider">
          <span>{{ item.label }}</span>
        </div>

        <!-- Message Item -->
        <div
          v-else
          :class="['message-item', isMe(item.message) ? 'is-me' : 'is-other']"
        >
          <!-- Avatar -->
          <div class="avatar-wrap">
            <div :class="['avatar', isMe(item.message) ? 'my-avatar' : 'other-avatar']">
              {{ getInitials(item.message.senderName) }}
            </div>
          </div>

          <!-- Content -->
          <div class="message-content-wrap">
            <div class="message-info">
              <span class="sender-name">{{ item.message.senderName }}</span>
              <span class="message-time">{{ formatTime(item.message.createdAt) }}</span>
            </div>
            
            <div :class="['message-bubble', isMe(item.message) ? 'my-bubble' : 'other-bubble']">
              <p class="message-text">{{ stripAnsiForChat(item.message.content) }}</p>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import type { Message } from '@/shared/types/chat'
import { stripAnsiForChat } from '@/shared/utils/stripAnsiForChat'

const props = defineProps<{
  messages: Message[]
  currentUserId: string
  loading?: boolean
}>()

const listRef = ref<HTMLElement | null>(null)

function isMe(msg: Message) {
  return msg.senderId === props.currentUserId
}

function getInitials(name: string) {
  return name ? name.charAt(0).toUpperCase() : '?'
}

function formatTime(timestamp?: number) {
  if (!timestamp) return ''
  return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

// Grouping logic (simplified)
const groupedMessages = computed(() => {
  return props.messages.map(m => ({ type: 'message', message: m }))
})

function scrollToBottom() {
  if (listRef.value) {
    listRef.value.scrollTop = listRef.value.scrollHeight
  }
}

onMounted(scrollToBottom)

defineExpose({ jumpToLatest: scrollToBottom })
</script>

<style scoped>
.messages-list-container {
  flex: 1;
  overflow-y: auto;
  padding: 32px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.empty-messages {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-bubble {
  background: #f1f5f9;
  padding: 12px 24px;
  border-radius: 100px;
  font-size: 13px;
  font-weight: 600;
  color: #94a3b8;
}

.message-item {
  display: flex;
  gap: 16px;
  max-width: 85%;
}

.is-me {
  flex-direction: row-reverse;
  align-self: flex-end;
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 900;
  color: white;
}

.my-avatar { background: #4f46e5; }
.other-avatar { background: #e2e8f0; color: #64748b; }

.message-content-wrap {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.is-me .message-content-wrap { align-items: flex-end; }

.message-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.sender-name {
  font-size: 13px;
  font-weight: 800;
  color: #0f172a;
}

.message-time {
  font-size: 10px;
  font-weight: 700;
  color: #94a3b8;
  text-transform: uppercase;
}

.message-bubble {
  padding: 14px 18px;
  border-radius: 20px;
  line-height: 1.5;
  font-size: 14px;
  font-weight: 500;
}

.my-bubble {
  background: #4f46e5;
  color: white;
  border-bottom-right-radius: 4px;
  box-shadow: 0 10px 20px -5px rgba(79, 70, 229, 0.3);
}

.other-bubble {
  background: white;
  color: #334155;
  border-bottom-left-radius: 4px;
  border: 1px solid #e2e8f0;
  box-shadow: 0 4px 10px rgba(0,0,0,0.02);
}

.message-text {
  white-space: pre-wrap;
  word-break: break-word;
}

.date-divider {
  text-align: center;
  position: relative;
  margin: 16px 0;
}

.date-divider span {
  background: #f8fafc;
  padding: 4px 12px;
  font-size: 11px;
  font-weight: 800;
  color: #cbd5e1;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  position: relative;
  z-index: 2;
}

.date-divider::after {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  width: 100%;
  height: 1px;
  background: #f1f5f9;
  z-index: 1;
}
</style>
