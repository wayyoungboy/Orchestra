<template>
  <nav class="h-full w-16 bg-panel/50 border-r border-white/5 flex flex-col items-center py-4">
    <!-- Logo -->
    <button
      @click="handleLogoClick"
      class="w-10 h-10 rounded-xl bg-primary/20 flex items-center justify-center mb-4 hover:bg-primary/30 transition-colors cursor-pointer"
      title="Back to Workspaces"
    >
      <span class="text-xl font-bold text-primary">O</span>
    </button>

    <!-- Divider -->
    <div class="w-8 h-px bg-white/10 rounded-full mb-4"></div>

    <!-- Workspace Navigation - only when in workspace -->
    <div v-if="inWorkspace" class="flex-1 flex flex-col gap-3 w-full px-3">
      <button
        v-for="item in workspaceNavItems"
        :key="item.id"
        @click="$emit('change', item.id)"
        :class="[
          'w-10 h-10 flex items-center justify-center rounded-xl transition-all',
          activeTab === item.id
            ? 'bg-gradient-to-br from-primary/80 to-primary-hover/80 text-white shadow-glow'
            : 'bg-white/5 text-white/40 hover:bg-white/10 hover:text-white'
        ]"
        :title="item.label"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <!-- Chat icon -->
          <path v-if="item.id === 'chat'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          <!-- Terminal icon -->
          <path v-else-if="item.id === 'terminal'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          <!-- Members icon -->
          <path v-else-if="item.id === 'members'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
          <!-- Settings icon -->
          <path v-else-if="item.id === 'settings'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
          <!-- Skills / puzzle（侧栏入口已隐藏）
          <path
            v-else-if="item.id === 'skills'"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-5M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"
          />
          -->
        </svg>
      </button>
    </div>

    <!-- Workspaces button - when not in workspace -->
    <div v-else class="flex-1 flex items-center justify-center w-full px-3">
      <button
        @click="$emit('change', 'workspaces')"
        :class="[
          'w-10 h-10 flex items-center justify-center rounded-xl transition-all',
          'bg-white/5 text-white/40 hover:bg-white/10 hover:text-white'
        ]"
        title="Workspaces"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
      </button>
    </div>
  </nav>
</template>

<script setup lang="ts">
export type TabId = 'chat' | 'terminal' | 'members' | 'settings' | 'skills' | 'workspaces'

defineProps<{
  activeTab: TabId
  inWorkspace?: boolean
}>()

const emit = defineEmits<{
  (e: 'change', tab: TabId): void
}>()

// Navigation items inside workspace
const workspaceNavItems: { id: TabId; label: string }[] = [
  { id: 'chat', label: 'Chat' },
  { id: 'terminal', label: 'Terminal' },
  { id: 'members', label: 'Members' },
  // { id: 'skills', label: 'Skills' },
  { id: 'settings', label: 'Settings' },
]

function handleLogoClick() {
  emit('change', 'workspaces')
}
</script>

<style scoped>
.shadow-glow {
  box-shadow: 0 0 15px rgba(var(--color-primary), 0.4);
}
</style>