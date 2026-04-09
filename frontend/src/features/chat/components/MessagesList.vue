<template>
  <div
    ref="listRef"
    class="messages-list-container custom-scrollbar"
    @scroll="handleScroll"
  >
    <!-- Loading Older Indicator -->
    <div v-if="loadingOlder" class="loading-older">
      <div class="loading-spinner"></div>
      <span>加载历史消息...</span>
    </div>

    <!-- Load More Trigger (at top of list) -->
    <div ref="loadMoreTrigger" class="load-more-trigger"></div>

    <!-- Empty State -->
    <div v-if="messages.length === 0 && !loading" class="empty-messages">
      <div class="empty-bubble">还没有消息。开始对话吧！</div>
    </div>

    <!-- Messages Groups -->
    <template v-else>
      <div v-for="item in groupedMessages" :key="item.message.id" class="message-group">
        <!-- Message Item -->
        <div
          :class="['message-item', isMe(item.message) ? 'is-me' : 'is-other']"
        >
          <!-- Avatar -->
          <div class="avatar-wrap">
            <div :class="['avatar', isMe(item.message) ? 'my-avatar' : (item.message.isAi ? 'ai-avatar' : 'other-avatar')]">
              <svg v-if="item.message.isAi" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
              </svg>
              <span v-else>{{ getInitials(resolveSenderName(item.message)) }}</span>
            </div>
          </div>

          <!-- Content -->
          <div class="message-content-wrap">
            <div class="message-info">
              <span class="sender-name">{{ resolveSenderName(item.message) }}</span>
              <span class="message-time">{{ formatTime(item.message.createdAt) }}</span>
            </div>

            <div :class="['message-bubble', isMe(item.message) ? 'my-bubble' : 'other-bubble']">
              <div
                class="message-text"
                v-html="renderMessageHtml(item.message)"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useChatStore } from '../chatStore'
import { stripAnsiForChat } from '@/shared/utils/stripAnsiForChat'
import { renderMarkdownSafe } from '@/shared/utils/markdown'

const props = defineProps<{
  messages: any[]
  currentUserId: string
  loading?: boolean
  loadingMessages?: boolean
  members?: any[]
}>()

const chatStore = useChatStore()
const listRef = ref<HTMLElement | null>(null)
const loadMoreTrigger = ref<HTMLElement | null>(null)

// State
const loadingOlder = ref(false)
const wasNearTop = ref(false)
let observer: IntersectionObserver | null = null

function isMe(msg: any) {
  return msg.senderId === props.currentUserId || (!msg.senderId && !msg.isAi)
}

function resolveSenderName(msg: any) {
  if (msg.senderName) return msg.senderName
  if (msg.senderId && props.members) {
    const member = props.members.find((m: any) => m.id === msg.senderId)
    if (member?.name) return member.name
  }
  if (!msg.senderId && !msg.isAi && props.members) {
    const owner = props.members.find((m: any) => m.roleType === 'owner')
    if (owner?.name) return owner.name
  }
  return msg.isAi ? 'AI Assistant' : 'Member'
}

function resolveMessageText(msg: any) {
  const content = msg.content
  const rawText = typeof content === 'string' ? content : (content?.text || '')
  return stripAnsiForChat(rawText)
}

function renderMessageHtml(msg: any) {
  const text = resolveMessageText(msg)
  return renderMarkdownSafe(text)
}

function getInitials(name: string) {
  return name ? name.charAt(0).toUpperCase() : '?'
}

function formatTime(timestamp?: number) {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

const groupedMessages = computed(() => {
  return props.messages.map(m => ({ type: 'message', message: m }))
})

function scrollToBottom(behavior: ScrollBehavior = 'auto') {
  nextTick(() => {
    if (listRef.value) {
      listRef.value.scrollTo({ top: listRef.value.scrollHeight, behavior })
    }
  })
}

function handleScroll() {
  if (!listRef.value) return
  const { scrollTop } = listRef.value
  // If user is near top (within 100px), flag that we should restore scroll position after loading
  wasNearTop.value = scrollTop < 100
}

async function loadOlder() {
  if (loadingOlder.value || !chatStore.hasMoreMessages) return
  loadingOlder.value = true

  // Record current scroll height and position for restoration
  const el = listRef.value
  const prevScrollHeight = el?.scrollHeight || 0
  const prevScrollTop = el?.scrollTop || 0

  await chatStore.loadOlderMessages(chatStore.workspaceId!, chatStore.activeConversationId!)

  // After new messages are prepended, restore scroll position
  await nextTick()
  if (el) {
    const newScrollHeight = el.scrollHeight
    el.scrollTop = prevScrollTop + (newScrollHeight - prevScrollHeight)
  }

  loadingOlder.value = false
}

// Watch for new messages from WebSocket — auto-scroll only if user was already at bottom
let prevMessageCount = props.messages.length
watch(
  () => props.messages.length,
  (newCount) => {
    if (newCount > prevMessageCount) {
      // New message arrived — scroll to bottom
      scrollToBottom('smooth')
    }
    prevMessageCount = newCount
  }
)

onMounted(() => {
  scrollToBottom()

  // Set up IntersectionObserver for infinite scroll
  if (loadMoreTrigger.value) {
    observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && chatStore.hasMoreMessages && !loadingOlder.value) {
          loadOlder()
        }
      },
      { root: listRef.value, rootMargin: '200px 0px 0px 0px', threshold: 0 }
    )
    observer.observe(loadMoreTrigger.value)
  }
})

onBeforeUnmount(() => {
  observer?.disconnect()
})

defineExpose({ jumpToLatest: () => scrollToBottom('smooth') })
</script>

<style scoped>
.messages-list-container {
  flex: 1;
  overflow-y: auto;
  padding: 32px;
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.loading-older {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 16px;
  color: #94a3b8;
  font-size: 13px;
  font-weight: 600;
}

.loading-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid #e2e8f0;
  border-top-color: #4f46e5;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.load-more-trigger {
  height: 1px;
  visibility: hidden;
  pointer-events: none;
}

.empty-messages {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-bubble {
  background: white;
  padding: 12px 24px;
  border-radius: 100px;
  font-size: 13px;
  font-weight: 700;
  color: #94a3b8;
  border: 1px solid #f1f5f9;
}

.message-item {
  display: flex;
  gap: 16px;
  max-width: 85%;
  align-items: flex-start;
}

.is-me {
  flex-direction: row-reverse;
  align-self: flex-end;
}

.avatar-wrap { flex-shrink: 0; }

.avatar {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 900;
  color: white;
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);
}

.my-avatar { background: #4f46e5; }
.ai-avatar { background: #10b981; }
.other-avatar { background: #e2e8f0; color: #64748b; }

.message-content-wrap {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 100px; /* Prevent collapse */
}

.is-me .message-content-wrap { align-items: flex-end; }

.message-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 4px;
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
  padding: 16px 20px;
  border-radius: 22px;
  line-height: 1.6;
  font-size: 14.5px;
  font-weight: 500;
  word-wrap: break-word;
  white-space: pre-wrap;
}

.my-bubble {
  background: #4f46e5;
  color: white;
  border-top-right-radius: 4px;
  box-shadow: 0 15px 30px -10px rgba(79, 70, 229, 0.4);
}

.other-bubble {
  background: white;
  color: #334155;
  border-top-left-radius: 4px;
  border: 1px solid #f1f5f9;
  box-shadow: 0 4px 15px rgba(0,0,0,0.02);
}

.message-text {
  margin: 0;
}

/* Markdown rendered content styles */
.message-text :deep(h1),
.message-text :deep(h2),
.message-text :deep(h3),
.message-text :deep(h4),
.message-text :deep(h5),
.message-text :deep(h6) {
  margin: 0.5em 0 0.3em;
  font-weight: 700;
  line-height: 1.3;
}

.message-text :deep(h1) { font-size: 1.4em; }
.message-text :deep(h2) { font-size: 1.3em; }
.message-text :deep(h3) { font-size: 1.2em; }

.message-text :deep(p) {
  margin: 0.4em 0;
}

.message-text :deep(ul),
.message-text :deep(ol) {
  margin: 0.4em 0;
  padding-left: 1.5em;
}

.message-text :deep(li) {
  margin: 0.2em 0;
}

.message-text :deep(code) {
  background: rgba(0, 0, 0, 0.1);
  padding: 0.1em 0.4em;
  border-radius: 4px;
  font-family: 'SF Mono', 'Monaco', 'Consolas', monospace;
  font-size: 0.9em;
}

.my-bubble .message-text :deep(code) {
  background: rgba(255, 255, 255, 0.2);
}

.message-text :deep(pre) {
  background: rgba(0, 0, 0, 0.05);
  padding: 0.8em 1em;
  border-radius: 8px;
  overflow-x: auto;
  margin: 0.5em 0;
}

.my-bubble .message-text :deep(pre) {
  background: rgba(255, 255, 255, 0.15);
}

.message-text :deep(pre code) {
  background: transparent;
  padding: 0;
  font-size: 0.85em;
}

.message-text :deep(blockquote) {
  border-left: 3px solid rgba(0, 0, 0, 0.2);
  margin: 0.5em 0;
  padding-left: 1em;
  color: rgba(0, 0, 0, 0.7);
}

.my-bubble .message-text :deep(blockquote) {
  border-left-color: rgba(255, 255, 255, 0.4);
  color: rgba(255, 255, 255, 0.9);
}

.message-text :deep(strong) {
  font-weight: 700;
}

.message-text :deep(em) {
  font-style: italic;
}

.message-text :deep(hr) {
  border: none;
  border-top: 1px solid rgba(0, 0, 0, 0.15);
  margin: 0.8em 0;
}

.my-bubble .message-text :deep(hr) {
  border-top-color: rgba(255, 255, 255, 0.3);
}

.message-text :deep(a) {
  color: #4f46e5;
  text-decoration: underline;
}

.my-bubble .message-text :deep(a) {
  color: #a5b4fc;
}
</style>
