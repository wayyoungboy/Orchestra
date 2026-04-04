<template>
  <div class="flex items-center gap-2 relative workspace-switcher z-50">
    <!-- Current Workspace Button -->
    <button
      @click="toggleDropdown"
      class="flex items-center gap-2 px-3 py-1.5 rounded-lg hover:bg-slate-100/50 transition-colors"
    >
      <svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
      </svg>
      <span class="text-slate-800 font-bold text-sm">
        {{ currentWorkspace?.name || t('workspace.switcherPlaceholder') }}
      </span>
      <svg
        class="w-4 h-4 text-slate-400 transition-transform"
        :class="{ 'rotate-180': showDropdown }"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>

    <!-- Dropdown Menu -->
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="showDropdown"
        class="absolute top-full left-0 mt-2 w-72 bg-white/90 backdrop-blur-xl rounded-2xl shadow-[0_10px_40px_-10px_rgba(0,0,0,0.1)] border border-slate-200/60 z-[100] overflow-hidden"
      >
        <!-- Workspace List -->
        <div class="max-h-64 overflow-y-auto custom-scrollbar">
          <button
            v-for="ws in workspaces"
            :key="ws.id"
            @click="handleSwitch(ws.id)"
            :class="[
              'w-full flex items-center gap-3 px-4 py-3 transition-colors',
              ws.id === currentWorkspace?.id
                ? 'bg-primary/5 text-primary'
                : 'hover:bg-slate-50 text-slate-600 hover:text-slate-900'
            ]"
          >
            <svg class="w-5 h-5 text-slate-400" :class="{'text-primary': ws.id === currentWorkspace?.id}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
            <div class="flex-1 min-w-0 text-left">
              <p class="text-sm font-bold truncate" :class="{'text-primary': ws.id === currentWorkspace?.id}">{{ ws.name }}</p>
              <p class="text-xs text-slate-400 truncate">{{ ws.path }}</p>
            </div>
            <svg
              v-if="ws.id === currentWorkspace?.id"
              class="w-5 h-5 text-primary"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </button>
        </div>

        <!-- Divider -->
        <div class="border-t border-slate-100"></div>

        <!-- Create New Workspace -->
        <div class="p-2">
          <button
            @click="handleCreate"
            class="w-full flex items-center gap-3 px-4 py-2.5 rounded-xl hover:bg-slate-50 text-slate-500 hover:text-primary transition-colors font-medium"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span class="text-sm">{{ t('workspace.switcherNew') }}</span>
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Workspace } from '@/shared/types/workspace'

const { t } = useI18n()

defineProps<{
  currentWorkspace: Workspace | null
  workspaces: Workspace[]
}>()

const emit = defineEmits<{
  (e: 'switch', workspaceId: string): void
  (e: 'create'): void
}>()

const showDropdown = ref(false)

function toggleDropdown() {
  showDropdown.value = !showDropdown.value
}

function handleSwitch(workspaceId: string) {
  emit('switch', workspaceId)
  showDropdown.value = false
}

function handleCreate() {
  emit('create')
  showDropdown.value = false
}

function handleClickOutside(event: MouseEvent) {
  const target = event.target as HTMLElement
  if (!target.closest('.workspace-switcher')) {
    showDropdown.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>