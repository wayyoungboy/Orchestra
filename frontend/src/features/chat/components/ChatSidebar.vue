<template>
  <aside class="w-12 lg:w-56 bg-panel/50 border-r border-white/5 flex flex-col shrink-0">
    <!-- Workspace Header -->
    <div class="h-16 flex items-center px-2 lg:px-4 justify-center lg:justify-start">
      <h2 class="text-white font-bold text-sm tracking-wide flex items-center gap-2">
        <svg class="w-4 h-4 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
        <span class="hidden lg:inline">{{ workspaceName || t('chat.workspaceFallback') }}</span>
      </h2>
    </div>

    <!-- Conversations List -->
    <div class="flex-1 overflow-y-auto py-4 px-1 lg:px-2 space-y-6">
      <!-- Channels Section -->
      <div>
        <div class="px-2 mb-2 items-center justify-between group hidden lg:flex">
          <h3 class="text-[11px] font-bold text-white/40 uppercase tracking-wider">{{ t('chat.channels') }}</h3>
          <button
            @click="$emit('new-conversation')"
            class="text-white/20 hover:text-white transition-colors"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
          </button>
        </div>

        <div class="space-y-1">
          <button
            v-for="conv in channelConversations"
            :key="conv.id"
            @click="$emit('select', conv.id)"
            :class="[
              'w-full px-2 lg:px-3 py-2 rounded-xl flex items-center gap-3 transition-all group cursor-pointer justify-center lg:justify-start',
              conv.id === activeConversationId
                ? 'bg-white/10 text-white'
                : 'text-white/60 hover:text-white hover:bg-white/5'
            ]"
          >
            <div class="relative">
              <span
                :class="[
                  'text-lg font-semibold',
                  conv.id === activeConversationId ? 'text-primary' : 'text-white/30'
                ]"
              >#</span>
            </div>
            <div class="hidden lg:flex items-start gap-3 min-w-0 flex-1">
              <div class="min-w-0 flex-1">
                <span class="text-[13px] font-semibold truncate">{{ getConversationTitle(conv) }}</span>
              </div>
              <!-- Unread Count -->
              <span
                v-if="conv.unreadCount && conv.unreadCount > 0"
                class="min-w-[18px] h-[18px] px-1 rounded-full bg-red-500 text-white text-[10px] font-bold flex items-center justify-center"
              >
                {{ conv.unreadCount > 99 ? '99+' : conv.unreadCount }}
              </span>
            </div>
          </button>
        </div>
      </div>

      <!-- Direct Messages Section -->
      <div>
        <div class="px-2 mb-2 items-center justify-between group hidden lg:flex">
          <h3 class="text-[11px] font-bold text-white/40 uppercase tracking-wider">{{ t('chat.directMessages') }}</h3>
        </div>

        <div class="space-y-1">
          <button
            v-for="conv in dmConversations"
            :key="conv.id"
            @click="$emit('select', conv.id)"
            :class="[
              'w-full px-2 lg:px-3 py-2 rounded-xl flex items-center gap-3 transition-all group cursor-pointer justify-center lg:justify-start',
              conv.id === activeConversationId
                ? 'bg-white/10 text-white'
                : 'text-white/60 hover:text-white hover:bg-white/5'
            ]"
          >
            <div class="relative">
              <div class="w-9 h-9 rounded-full bg-white/10 flex items-center justify-center">
                <span class="text-sm font-semibold text-white/60">{{ getInitials(getConversationTitle(conv)) }}</span>
              </div>
              <!-- Online indicator -->
              <div class="absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full bg-green-500 border-2 border-panel"></div>
            </div>
            <div class="hidden lg:flex items-start gap-3 min-w-0 flex-1">
              <div class="min-w-0 flex-1">
                <span class="text-[13px] font-semibold truncate">{{ getConversationTitle(conv) }}</span>
              </div>
              <!-- Unread Count -->
              <span
                v-if="conv.unreadCount && conv.unreadCount > 0"
                class="min-w-[18px] h-[18px] px-1 rounded-full bg-red-500 text-white text-[10px] font-bold flex items-center justify-center"
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
  workspaceName: string
}>()

defineEmits<{
  (e: 'select', id: string): void
  (e: 'new-conversation'): void
}>()

const channelConversations = computed(() =>
  props.conversations.filter(c => c.type === 'channel')
)

const dmConversations = computed(() =>
  props.conversations.filter(c => c.type === 'dm')
)

function getConversationTitle(conv: Conversation): string {
  if (conv.customName) return conv.customName
  if (conv.nameKey) return conv.nameKey
  return conv.id
}

function getInitials(name: string): string {
  return name.charAt(0).toUpperCase()
}
</script>