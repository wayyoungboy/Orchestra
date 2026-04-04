<template>
  <Teleport to="body">
    <div v-if="state.open" class="fixed inset-0 z-[200]" @pointerdown="closeMenu" @contextmenu.prevent>
      <div
        ref="menuRef"
        class="absolute rounded-xl bg-panel-strong/95 border border-white/10 shadow-2xl overflow-hidden min-w-[180px] py-1.5 ring-1 ring-white/10"
        :style="menuStyle"
        @pointerdown.stop
        @contextmenu.prevent
      >
        <template v-for="entry in state.entries" :key="entry.id">
          <button
            type="button"
            :class="[
              'w-full text-left px-4 py-2.5 text-xs font-bold flex items-center gap-3 transition-colors',
              entry.danger
                ? 'text-red-400 hover:bg-red-500/20'
                : 'text-white hover:bg-white/15'
            ]"
            @click="handleEntry(entry)"
          >
            <svg v-if="entry.icon" class="w-4 h-4 opacity-70" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path v-if="entry.icon === 'chat'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              <path v-else-if="entry.icon === 'mention'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 12a4 4 0 10-8 0 4 4 0 008 0zm0 0v1.5a2.5 2.5 0 005 0V12a9 9 0 10-9 9" />
              <path v-else-if="entry.icon === 'edit'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              <path v-else-if="entry.icon === 'status'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5.121 17.804A13.937 13.937 0 0112 16c2.5 0 4.847.655 6.879 1.804M15 10a3 3 0 11-6 0 3 3 0 016 0zm6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              <path v-else-if="entry.icon === 'remove'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7a4 4 0 11-8 0 4 4 0 018 0zM9 14a6 6 0 00-6 6v1h12v-1a6 6 0 00-6-6zM21 12h-6" />
            </svg>
            <span class="flex-1">{{ entry.label }}</span>
          </button>
        </template>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useContextMenu, type ContextMenuItem } from './useContextMenu'

const { state, closeMenu } = useContextMenu()
const menuRef = ref<HTMLDivElement | null>(null)
const menuStyle = ref<Record<string, string>>({})

const updatePosition = async () => {
  if (!state.value.open) return
  await nextTick()
  const menu = menuRef.value
  if (!menu) return

  const rect = menu.getBoundingClientRect()
  const padding = 8
  let left = state.value.x
  let top = state.value.y

  const maxLeft = window.innerWidth - rect.width - padding
  const maxTop = window.innerHeight - rect.height - padding

  if (left > maxLeft) left = Math.max(padding, maxLeft)
  if (top > maxTop) top = Math.max(padding, maxTop)

  menuStyle.value = { left: left + 'px', top: top + 'px' }
}

const handleEntry = (entry: ContextMenuItem) => {
  closeMenu()
  entry.action?.()
}

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    event.preventDefault()
    closeMenu()
  }
}

watch(() => [state.value.open, state.value.x, state.value.y], updatePosition, { immediate: true })

onMounted(() => window.addEventListener('keydown', handleKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', handleKeydown))
</script>
