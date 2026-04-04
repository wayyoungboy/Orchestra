<template>
  <div class="flex h-full w-full relative">
    <!-- Sidebar Navigation -->
    <SidebarNav :active-tab="activeTab" :in-workspace="!showWorkspaceSelection && !!workspaceStore.currentWorkspace" @change="setActiveTab" />

    <!-- Main Content -->
    <main class="flex-1 h-full overflow-hidden relative flex flex-col bg-panel/50">
      <!-- Top Bar with Workspace Switcher (only when in workspace) -->
      <header
        v-if="workspaceStore.currentWorkspace && !showWorkspaceSelection"
        class="h-12 bg-panel/50 border-b border-white/5 flex items-center px-4"
      >
        <WorkspaceSwitcher
          :current-workspace="workspaceStore.currentWorkspace"
          :workspaces="workspaceStore.workspaces"
          @switch="handleSwitchWorkspace"
          @create="handleCreateWorkspace"
        />
      </header>

      <!-- Content Area -->
      <div class="flex-1 overflow-hidden">
        <!-- Workspace Selection -->
        <WorkspaceSelection v-if="showWorkspaceSelection" />

        <!-- Chat Interface -->
        <ChatInterface v-else-if="activeTab === 'chat' && workspaceStore.currentWorkspace" />
        <div v-else-if="activeTab === 'chat'" class="h-full flex items-center justify-center text-white/40">
          <div class="text-center">
            <p class="text-lg mb-2">{{ t('workspace.noSelectionTitle') }}</p>
            <p class="text-sm">{{ t('workspace.noSelectionHint') }}</p>
          </div>
        </div>

        <!-- Terminal Workspace -->
        <TerminalWorkspace v-else-if="activeTab === 'terminal' && workspaceStore.currentWorkspace" />
        <div v-else-if="activeTab === 'terminal'" class="h-full flex items-center justify-center text-white/40">
          <div class="text-center">
            <p class="text-lg mb-2">{{ t('workspace.noSelectionTitle') }}</p>
            <p class="text-sm">{{ t('workspace.noSelectionHint') }}</p>
          </div>
        </div>

        <!-- Members Page -->
        <MembersPage v-else-if="activeTab === 'members' && workspaceStore.currentWorkspace" />
        <div v-else-if="activeTab === 'members'" class="h-full flex items-center justify-center text-white/40">
          <div class="text-center">
            <p class="text-lg mb-2">{{ t('workspace.noSelectionTitle') }}</p>
            <p class="text-sm">{{ t('workspace.noSelectionHint') }}</p>
          </div>
        </div>

        <!-- Skills (Web placeholder, REQ-403) -->
        <SkillsPlaceholder v-else-if="activeTab === 'skills' && workspaceStore.currentWorkspace" />
        <div v-else-if="activeTab === 'skills'" class="h-full flex items-center justify-center text-white/40">
          <div class="text-center">
            <p class="text-lg mb-2">{{ t('workspace.noSelectionTitle') }}</p>
            <p class="text-sm">{{ t('workspace.noSelectionHint') }}</p>
          </div>
        </div>

        <!-- Settings -->
        <Settings v-else-if="activeTab === 'settings'" />
      </div>
    </main>

    <!-- Keyboard Shortcuts Help Overlay -->
    <div
      v-if="helpVisible"
      class="absolute inset-0 bg-black/50 flex items-center justify-center z-50"
      @click="helpVisible = false"
    >
      <div
        class="bg-panel rounded-xl border border-white/10 shadow-xl w-80 max-h-96 overflow-hidden"
        @click.stop
      >
        <div class="p-4 border-b border-white/10">
          <h3 class="text-lg font-bold text-white">Keyboard Shortcuts</h3>
        </div>
        <div class="p-4 space-y-3 overflow-y-auto">
          <div class="flex items-center justify-between py-2">
            <span class="text-sm text-white/80">Focus search</span>
            <kbd class="px-2 py-1 rounded bg-surface border border-white/10 text-xs font-mono text-white/60">
              Ctrl+K
            </kbd>
          </div>
          <div class="flex items-center justify-between py-2">
            <span class="text-sm text-white/80">Toggle help</span>
            <kbd class="px-2 py-1 rounded bg-surface border border-white/10 text-xs font-mono text-white/60">
              Ctrl+/
            </kbd>
          </div>
          <div class="flex items-center justify-between py-2">
            <span class="text-sm text-white/80">Close modal</span>
            <kbd class="px-2 py-1 rounded bg-surface border border-white/10 text-xs font-mono text-white/60">
              Esc
            </kbd>
          </div>
          <div class="flex items-center justify-between py-2">
            <span class="text-sm text-white/80">Navigate to Chat</span>
            <kbd class="px-2 py-1 rounded bg-surface border border-white/10 text-xs font-mono text-white/60">
              Ctrl+1
            </kbd>
          </div>
          <div class="flex items-center justify-between py-2">
            <span class="text-sm text-white/80">Navigate to Terminal</span>
            <kbd class="px-2 py-1 rounded bg-surface border border-white/10 text-xs font-mono text-white/60">
              Ctrl+2
            </kbd>
          </div>
          <div class="flex items-center justify-between py-2">
            <span class="text-sm text-white/80">Navigate to Members</span>
            <kbd class="px-2 py-1 rounded bg-surface border border-white/10 text-xs font-mono text-white/60">
              Ctrl+3
            </kbd>
          </div>
          <div class="flex items-center justify-between py-2">
            <span class="text-sm text-white/80">Navigate to Settings</span>
            <kbd class="px-2 py-1 rounded bg-surface border border-white/10 text-xs font-mono text-white/60">
              Ctrl+4
            </kbd>
          </div>
          <div class="flex items-center justify-between py-2">
            <span class="text-sm text-white/80">Navigate to Skills</span>
            <kbd class="px-2 py-1 rounded bg-surface border border-white/10 text-xs font-mono text-white/60">
              Ctrl+5
            </kbd>
          </div>
        </div>
        <div class="p-4 border-t border-white/10">
          <p class="text-xs text-white/40 text-center">
            Press <kbd class="px-1.5 py-0.5 rounded bg-surface border border-white/10 text-xs font-mono">Esc</kbd> to close
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'
import { useAppShortcuts } from '@/shared/composables'
import SidebarNav, { type TabId } from '@/shared/components/SidebarNav.vue'
import WorkspaceSwitcher from '@/shared/components/WorkspaceSwitcher.vue'
import WorkspaceSelection from '@/features/workspace/WorkspaceSelection.vue'
import ChatInterface from '@/features/chat/ChatInterface.vue'
import TerminalWorkspace from '@/features/terminal/TerminalWorkspace.vue'
import Settings from '@/features/settings/Settings.vue'
import MembersPage from '@/features/members/MembersPage.vue'
import SkillsPlaceholder from '@/features/skills/SkillsPlaceholder.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const workspaceStore = useWorkspaceStore()

const activeTab = ref<TabId>('workspaces')
const helpVisible = ref(false)

// Keyboard shortcuts
useAppShortcuts({
  onFocusSearch: () => {
    // Focus command palette or search - to be implemented
    console.log('Focus search triggered')
  },
  onToggleHelp: () => {
    helpVisible.value = !helpVisible.value
  },
  onCloseModal: () => {
    helpVisible.value = false
  },
  onNavigateChat: () => {
    if (workspaceStore.currentWorkspace) {
      setActiveTab('chat')
    }
  },
  onNavigateTerminal: () => {
    if (workspaceStore.currentWorkspace) {
      setActiveTab('terminal')
    }
  },
  onNavigateMembers: () => {
    if (workspaceStore.currentWorkspace) {
      setActiveTab('members')
    }
  },
  onNavigateSettings: () => {
    setActiveTab('settings')
  },
  onNavigateSkills: () => {
    if (workspaceStore.currentWorkspace) {
      setActiveTab('skills')
    }
  },
  helpVisible
})

// Only show workspace selection when on workspaces route
const showWorkspaceSelection = computed(() => {
  return activeTab.value === 'workspaces'
})

function setActiveTab(tab: TabId) {
  activeTab.value = tab

  // Update route based on tab
  const workspaceId = workspaceStore.currentWorkspace?.id
  if (workspaceId) {
    if (tab === 'workspaces') {
      router.push('/workspaces')
    } else {
      router.push(`/workspace/${workspaceId}/${tab}`)
    }
  } else if (tab !== 'workspaces' && tab !== 'settings') {
    // If no workspace, redirect to workspaces
    router.push('/workspaces')
  }
}

async function handleSwitchWorkspace(workspaceId: string) {
  await workspaceStore.openWorkspace(workspaceId)
  if (workspaceStore.currentWorkspace) {
    activeTab.value = 'chat'
    router.push(`/workspace/${workspaceId}/chat`)
  }
}

function handleCreateWorkspace() {
  activeTab.value = 'workspaces'
  router.push('/workspaces')
}

// Sync activeTab with route
watch(
  () => route.path,
  (path) => {
    if (path.includes('/terminal')) {
      activeTab.value = 'terminal'
    } else if (path.includes('/members')) {
      activeTab.value = 'members'
    } else if (path.includes('/skills')) {
      activeTab.value = 'skills'
    } else if (path.includes('/settings')) {
      activeTab.value = 'settings'
    } else if (path.includes('/chat')) {
      activeTab.value = 'chat'
    } else if (path.includes('/workspaces') || path === '/') {
      activeTab.value = 'workspaces'
    }
  },
  { immediate: true }
)

// Load workspace on mount
watch(
  () => route.params.id,
  async (workspaceId) => {
    if (workspaceId && typeof workspaceId === 'string') {
      if (!workspaceStore.currentWorkspace || workspaceStore.currentWorkspace.id !== workspaceId) {
        await workspaceStore.openWorkspace(workspaceId)
      }
    }
  },
  { immediate: true }
)
</script>