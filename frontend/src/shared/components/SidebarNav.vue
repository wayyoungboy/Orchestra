<template>
  <nav class="h-full w-16 md:w-20 bg-white/40 backdrop-blur-xl border-r border-slate-200/50 flex flex-col items-center py-6 shadow-[4px_0_24px_rgba(0,0,0,0.02)] z-20">
    <!-- Logo -->
    <button
      @click="handleLogoClick"
      class="w-12 h-12 rounded-2xl bg-gradient-primary shadow-glow-primary flex items-center justify-center mb-6 hover:scale-105 transition-transform duration-300 cursor-pointer group"
      title="Back to Workspaces"
    >
      <span class="text-xl font-black text-white group-hover:rotate-12 transition-transform duration-500">O</span>
    </button>

    <!-- Divider -->
    <div class="w-8 h-px bg-slate-200 rounded-full mb-6"></div>

    <!-- Workspace Navigation - only when in workspace -->
    <div v-if="inWorkspace" class="flex-1 flex flex-col gap-4 w-full px-3">
      <button
        v-for="item in workspaceNavItems"
        :key="item.id"
        @click="$emit('change', item.id)"
        :class="[
          'w-12 h-12 mx-auto flex items-center justify-center rounded-2xl transition-all duration-300',
          activeTab === item.id
            ? 'bg-white shadow-[0_4px_15px_rgba(0,0,0,0.05)] text-primary ring-1 ring-slate-100'
            : 'bg-transparent text-slate-400 hover:bg-white/60 hover:text-slate-700'
        ]"
        :title="item.label"
      >
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <!-- Chat icon -->
          <path v-if="item.id === 'chat'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          <!-- Terminal icon -->
          <path v-else-if="item.id === 'terminal'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          <!-- Members icon -->
          <path v-else-if="item.id === 'members'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
          <!-- Tasks icon -->
          <path v-else-if="item.id === 'tasks'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
          <!-- Settings icon -->
          <path v-else-if="item.id === 'settings'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
        </svg>
      </button>
    </div>

    <!-- Workspaces button - when not in workspace -->
    <div v-else class="flex-1 flex flex-col items-center w-full px-3">
      <button
        @click="$emit('change', 'workspaces')"
        :class="[
          'w-12 h-12 flex items-center justify-center rounded-2xl transition-all duration-300',
          'bg-white shadow-[0_4px_15px_rgba(0,0,0,0.05)] text-primary ring-1 ring-slate-100'
        ]"
        title="Workspaces"
      >
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
      </button>
    </div>
  </nav>
</template>

<script setup lang="ts">
export type TabId = 'chat' | 'terminal' | 'members' | 'tasks' | 'settings' | 'workspaces'

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
  { id: 'tasks', label: 'Tasks' },
  { id: 'settings', label: 'Settings' },
]

function handleLogoClick() {
  emit('change', 'workspaces')
}
</script>
