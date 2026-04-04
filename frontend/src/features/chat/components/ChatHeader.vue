<template>
  <header class="h-[72px] px-8 flex items-center justify-between border-b border-white/5 bg-panel/30 backdrop-blur-md sticky top-0 z-10">
    <!-- Center Title -->
    <div class="flex flex-col items-center text-center flex-1">
      <h1 class="text-white font-semibold text-[17px] tracking-tight leading-tight mb-0.5">{{ title }}</h1>
      <p v-if="description" class="text-white/40 text-[11px] font-medium tracking-wide truncate max-w-xs">{{ description }}</p>
    </div>

    <!-- Right Actions -->
    <button
      v-if="memberCount !== undefined"
      type="button"
      class="md:hidden inline-flex items-center gap-2 px-3 py-2 rounded-full bg-white/10 text-white/70 border border-white/10 hover:bg-white/20 hover:text-white transition-colors"
      @click="$emit('open-members')"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
      </svg>
      <span class="text-[11px] font-semibold">{{ t('chat.headerMembers') }}</span>
      <span class="text-[10px] font-bold bg-white/10 px-1.5 py-0.5 rounded-full">{{ memberCount }}</span>
    </button>
  </header>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps<{
  title: string
  description?: string
  memberCount?: number
}>()

defineEmits<{
  (e: 'open-members'): void
}>()
</script>