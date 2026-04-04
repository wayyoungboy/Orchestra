<template>
  <div class="flex h-full w-full flex-col">
    <!-- Terminal Tabs -->
    <header class="h-12 px-2 border-b border-white/5 flex items-center gap-2 bg-panel/50">
      <button
        v-for="tab in terminalStore.tabs"
        :key="tab.id"
        @click="terminalStore.setActiveTab(tab.id)"
        :class="[
          'h-8 px-4 rounded-lg flex items-center gap-2 text-sm transition-colors',
          terminalStore.activeTabId === tab.id
            ? 'bg-white/10 text-white'
            : 'text-white/40 hover:bg-white/5 hover:text-white/60'
        ]"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
        <span>{{ tab.title }}</span>
        <!-- Activity Indicator -->
        <span
          v-if="tab.hasActivity && terminalStore.activeTabId !== tab.id"
          class="w-2 h-2 rounded-full bg-primary"
        ></span>
        <!-- Pin Indicator -->
        <svg
          v-if="tab.pinned"
          class="w-3 h-3 text-primary"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
        </svg>
        <!-- Close Button -->
        <button
          v-if="terminalStore.tabs.length > 1"
          @click.stop="() => terminalStore.closeTab(tab.id)"
          class="ml-1 p-0.5 rounded hover:bg-white/10"
        >
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </button>

      <!-- Add Tab Button -->
      <button
        @click="() => terminalStore.createTab()"
        class="h-8 w-8 rounded-lg flex items-center justify-center text-white/40 hover:bg-white/5 hover:text-white/60 transition-colors"
        title="New Terminal"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
      </button>

      <!-- Layout Controls -->
      <div class="ml-auto flex items-center gap-1">
        <button
          v-for="layout in layoutOptions"
          :key="layout.mode"
          @click="terminalStore.setLayoutMode(layout.mode)"
          :class="[
            'h-8 w-8 rounded-lg flex items-center justify-center transition-colors',
            terminalStore.layoutMode === layout.mode
              ? 'bg-white/10 text-white'
              : 'text-white/40 hover:bg-white/5 hover:text-white/60'
          ]"
          :title="layout.title"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="layout.icon" />
          </svg>
        </button>
      </div>
    </header>

    <!-- Terminal Content -->
    <main class="flex-1 relative bg-[#0b0f14]">
      <TerminalPane
        v-for="tab in terminalStore.tabs"
        :key="tab.id"
        :terminal-id="tab.id"
        :active="terminalStore.activeTabId === tab.id"
        class="absolute inset-0"
        :class="{ 'pointer-events-none opacity-0': terminalStore.activeTabId !== tab.id }"
      />

      <!-- Empty State -->
      <div
        v-if="terminalStore.tabs.length === 0"
        class="flex items-center justify-center h-full text-white/50"
      >
        <div class="text-center">
          <div class="w-16 h-16 rounded-full bg-white/5 flex items-center justify-center mx-auto mb-4">
            <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
          <p class="text-sm">No terminals open</p>
          <p class="text-xs mt-1">Click + to create a new terminal</p>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useTerminalStore } from './terminalStore'
import TerminalPane from './TerminalPane.vue'
import type { TerminalLayoutMode } from '@/shared/types/terminal'

const terminalStore = useTerminalStore()

const layoutOptions: { mode: TerminalLayoutMode; title: string; icon: string }[] = [
  { mode: 'single', title: 'Single', icon: 'M4 6h16M4 12h16M4 18h16' },
  { mode: 'split-vertical', title: 'Split Vertical', icon: 'M9 6h6M9 12h6M9 18h6' },
  { mode: 'split-horizontal', title: 'Split Horizontal', icon: 'M4 6h16M4 12h16' },
]
</script>