<template>
  <div
    ref="listRef"
    class="flex-1 overflow-y-auto px-8 py-6 space-y-8"
  >
    <!-- Empty State -->
    <div v-if="messages.length === 0" class="flex flex-col items-center justify-center h-full text-white/40">
      <div class="w-16 h-16 rounded-full bg-white/5 flex items-center justify-center mb-4">
        <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
        </svg>
      </div>
      <p class="text-sm font-medium">No messages yet</p>
      <p class="text-xs mt-1">Start the conversation!</p>
    </div>

    <!-- Message Groups -->
    <template v-else>
      <template v-for="item in groupedItems" :key="item.id">
        <!-- Date Separator -->
        <div v-if="item.type === 'separator'" class="relative flex py-2 items-center justify-center">
          <div class="absolute inset-0 flex items-center">
            <div class="w-full border-t border-white/5"></div>
          </div>
          <span class="relative px-4 bg-panel/30 rounded-full text-white/30 text-[11px] font-semibold backdrop-blur-md border border-white/5">
            {{ item.label }}
          </span>
        </div>

        <!-- Message -->
        <div
          v-else
          :class="[
            'flex gap-5 group -mx-6 px-6 py-2 rounded-2xl transition-all hover:bg-white/[0.02]',
            isMe(item.message) ? 'flex-row-reverse' : ''
          ]"
        >
          <!-- Avatar -->
          <div class="mt-1 shrink-0">
            <div
              :class="[
                'w-11 h-11 rounded-[14px] shadow-lg flex items-center justify-center',
                isMe(item.message) ? 'bg-primary/20' : 'bg-white/10'
              ]"
            >
              <span class="text-sm font-bold text-white">
                {{ getInitials(item.message.senderName) }}
              </span>
            </div>
          </div>

          <!-- Message Content -->
          <div :class="['flex flex-col flex-1 min-w-0', isMe(item.message) ? 'items-end' : '']">
            <!-- Sender Info -->
            <div :class="['flex items-baseline gap-2.5', isMe(item.message) ? 'flex-row-reverse' : '']">
              <span class="text-white font-semibold text-[15px] cursor-pointer hover:underline tracking-tight">
                {{ item.message.senderName }}
              </span>
              <span class="text-white/30 text-[11px] font-medium">
                {{ formatTime(item.message.createdAt) }}
              </span>
            </div>

            <!-- Message Bubble -->
            <div
              v-if="isMe(item.message)"
              class="selectable mt-1 bg-white text-slate-900 px-5 py-3 rounded-2xl rounded-tr-sm shadow-lg max-w-[80%] text-[15px] leading-relaxed font-medium"
            >
              <p class="whitespace-pre-wrap">{{ stripAnsiForChat(item.message.content) }}</p>
            </div>
            <div v-else class="selectable text-white/90 text-[15px] leading-relaxed mt-1 font-light tracking-wide max-w-[80%]">
              <p class="whitespace-pre-wrap">{{ stripAnsiForChat(item.message.content) }}</p>
            </div>
          </div>
        </div>
      </template>
    </template>

    <!-- Jump to Latest Button -->
    <button
      v-if="showJumpButton"
      type="button"
      class="sticky bottom-6 self-end mr-2 px-4 py-2 rounded-full bg-panel/80 border border-white/10 text-white/70 hover:text-white hover:bg-panel/80 transition-all shadow-lg backdrop-blur"
      @click="handleJumpToLatest"
    >
      <svg class="w-4 h-4 mr-1 inline-block align-middle" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 14l-7 7m0 0l-7-7m7 7V3" />
      </svg>
      <span class="text-[12px] font-medium">Jump to latest</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { stripAnsiForChat } from '@/shared/utils/stripAnsiForChat'

interface DisplayMessage {
  id: string
  conversationId: string
  senderId: string
  senderName: string
  senderAvatar?: string
  content: string
  createdAt: string
}

const props = defineProps<{
  messages: DisplayMessage[]
  currentUserId: string
}>()

const listRef = ref<HTMLDivElement | null>(null)
const isPinnedToBottom = ref(true)
const showJumpButton = computed(() => !isPinnedToBottom.value)

// Group messages by date
interface SeparatorItem {
  id: string
  type: 'separator'
  label: string
}

interface MessageItem {
  id: string
  type: 'message'
  message: DisplayMessage
}

type GroupedItem = SeparatorItem | MessageItem

const groupedItems = computed<GroupedItem[]>(() => {
  const items: GroupedItem[] = []
  let lastDate = ''

  for (const message of props.messages) {
    const messageDate = formatDate(message.createdAt)
    if (messageDate !== lastDate) {
      items.push({
        id: `separator-${message.id}`,
        type: 'separator',
        label: messageDate
      })
      lastDate = messageDate
    }
    items.push({
      id: message.id,
      type: 'message',
      message
    })
  }

  return items
})

function isMe(message: DisplayMessage): boolean {
  return message.senderId === props.currentUserId
}

function getInitials(name: string): string {
  return name.charAt(0).toUpperCase()
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  if (date.toDateString() === today.toDateString()) {
    return 'Today'
  }
  if (date.toDateString() === yesterday.toDateString()) {
    return 'Yesterday'
  }
  return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
}

function updatePinnedState() {
  if (!listRef.value) return
  const threshold = 120
  const distanceFromBottom = listRef.value.scrollHeight - listRef.value.scrollTop - listRef.value.clientHeight
  isPinnedToBottom.value = distanceFromBottom < threshold
}

function scrollToBottom() {
  if (!listRef.value) return
  listRef.value.scrollTop = listRef.value.scrollHeight
}

function handleJumpToLatest() {
  scrollToBottom()
  isPinnedToBottom.value = true
}

// Auto-scroll on new messages
watch(
  () => props.messages.length,
  async () => {
    await nextTick()
    if (isPinnedToBottom.value) {
      scrollToBottom()
    }
  }
)

onMounted(() => {
  updatePinnedState()
  listRef.value?.addEventListener('scroll', updatePinnedState, { passive: true })
  scrollToBottom()
})

onBeforeUnmount(() => {
  listRef.value?.removeEventListener('scroll', updatePinnedState)
})

defineExpose({
  jumpToLatest: handleJumpToLatest
})
</script>