<template>
  <div class="workspace-root" @mousemove="handleMouseMove">
    <!-- Layer 0: Background Layer (1:1 Match with Login Page) -->
    <div class="background-layer">
      <div class="grid-pattern"></div>
      <div class="bg-glow-center"></div>
      
      <!-- Parallax Orbs with Fluid Motion -->
      <div class="orb orb-1 orb-animate-1" :style="parallaxStyle(0.05)"></div>
      <div class="orb orb-2 orb-animate-2" :style="parallaxStyle(-0.03)"></div>
      <div class="orb orb-3 orb-animate-3" :style="parallaxStyle(0.02)"></div>

      <!-- Floating Particles (Optional for deeper consistency) -->
      <div class="particles-container">
        <div v-for="i in 15" :key="i" class="particle" :style="randomParticleStyle()"></div>
      </div>
    </div>

    <!-- Sidebar Navigation -->
    <div class="nav-container">
      <SidebarNav :active-tab="activeTab" :in-workspace="!showWorkspaceSelection && !!workspaceStore.currentWorkspace" @change="setActiveTab" />
    </div>

    <!-- Main Content -->
    <main class="main-body">
      <!-- Top Bar (Only in workspace) -->
      <header
        v-if="workspaceStore.currentWorkspace && !showWorkspaceSelection"
        class="workspace-header"
      >
        <WorkspaceSwitcher
          :current-workspace="workspaceStore.currentWorkspace"
          :workspaces="workspaceStore.workspaces"
          @switch="handleSwitchWorkspace"
          @create="handleCreateWorkspace"
        />
      </header>

      <!-- Content Area -->
      <div class="content-view">
        <WorkspaceSelection v-if="showWorkspaceSelection" />
        
        <ChatInterface v-else-if="activeTab === 'chat' && workspaceStore.currentWorkspace" />
        <MembersPage v-else-if="activeTab === 'members' && workspaceStore.currentWorkspace" />
        <TasksPage v-else-if="activeTab === 'tasks' && workspaceStore.currentWorkspace" />
        <Settings v-else-if="activeTab === 'settings'" />

        <!-- No Selection Placeholder -->
        <div v-else-if="!showWorkspaceSelection" class="empty-state">
          <div class="empty-glass-card">
            <p class="text-lg font-bold text-slate-900 mb-2">{{ t('workspace.noSelectionTitle') }}</p>
            <p class="text-sm text-slate-500">{{ t('workspace.noSelectionHint') }}</p>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, reactive, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'
import { useChatStore } from '@/features/chat/chatStore'
import SidebarNav, { type TabId } from '@/shared/components/SidebarNav.vue'
import WorkspaceSwitcher from '@/shared/components/WorkspaceSwitcher.vue'
import WorkspaceSelection from '@/features/workspace/WorkspaceSelection.vue'
import ChatInterface from '@/features/chat/ChatInterface.vue'
import Settings from '@/features/settings/Settings.vue'
import MembersPage from '@/features/members/MembersPage.vue'
import TasksPage from '@/features/tasks/TasksPage.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const workspaceStore = useWorkspaceStore()
const chatStore = useChatStore()

const activeTab = ref<TabId>('workspaces')

// Mouse Parallax Logic (Identical to Login)
const mouse = reactive({ x: 0, y: 0 })
function handleMouseMove(e: MouseEvent) {
  mouse.x = (e.clientX - window.innerWidth / 2) / 10
  mouse.y = (e.clientY - window.innerHeight / 2) / 10
}
function parallaxStyle(factor: number) {
  return { transform: `translate(${mouse.x * factor}px, ${mouse.y * factor}px)` }
}

function randomParticleStyle() {
  return {
    left: `${Math.random() * 100}%`,
    top: `${Math.random() * 100}%`,
    width: `${Math.random() * 2 + 2}px`,
    height: `${Math.random() * 2 + 2}px`,
    animationDelay: `${Math.random() * 5}s`,
    animationDuration: `${Math.random() * 10 + 10}s`
  }
}

const showWorkspaceSelection = computed(() => activeTab.value === 'workspaces')

function setActiveTab(tab: TabId) {
  activeTab.value = tab
  const workspaceId = workspaceStore.currentWorkspace?.id
  if (workspaceId) {
    if (tab === 'workspaces') router.push('/workspaces').catch(() => {})
    else router.push(`/workspace/${workspaceId}/${tab}`).catch(() => {})
  } else if (tab !== 'workspaces' && tab !== 'settings') {
    router.push('/workspaces').catch(() => {})
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

watch(() => route.path, (path) => {
  if (path.includes('/members')) activeTab.value = 'members'
  else if (path.includes('/tasks')) activeTab.value = 'tasks'
  else if (path.includes('/settings')) activeTab.value = 'settings'
  else if (path.includes('/chat')) activeTab.value = 'chat'
  else if (path.includes('/workspaces') || path === '/') activeTab.value = 'workspaces'
}, { immediate: true })

watch(() => route.params.id, async (workspaceId) => {
  if (workspaceId && typeof workspaceId === 'string') {
    if (!workspaceStore.currentWorkspace || workspaceStore.currentWorkspace.id !== workspaceId) {
      await workspaceStore.openWorkspace(workspaceId)
    }
  }
}, { immediate: true })

// Disconnect chat WebSocket when leaving the workspace entirely
onBeforeUnmount(() => {
  chatStore.disconnectChatWebSocket()
})
</script>

<style scoped>
.workspace-root {
  height: 100vh;
  width: 100vw;
  background-color: #f8fafc;
  display: flex;
  position: relative;
  overflow: hidden;
}

/* Layer 0: Background - EXACT MATCH with Login */
.background-layer {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.grid-pattern {
  position: absolute;
  inset: 0;
  background-image: 
    linear-gradient(to right, rgba(128, 128, 128, 0.07) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(128, 128, 128, 0.07) 1px, transparent 1px);
  background-size: 40px 40px;
}

.bg-glow-center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 100%;
  height: 100%;
  background: radial-gradient(circle at 50% 50%, rgba(99, 102, 241, 0.05), transparent 70%);
}

.orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(100px);
  opacity: 0.15;
  transition: transform 0.2s ease-out;
  pointer-events: none;
}

.orb-1 { top: -10%; left: -10%; width: 50%; height: 50%; background-color: #6366f1; }
.orb-2 { bottom: -10%; right: -5%; width: 40%; height: 40%; background-color: #10b981; }
.orb-3 { top: 40%; right: 20%; width: 20%; height: 20%; background-color: #8b5cf6; opacity: 0.1; }

.particles-container { position: absolute; inset: 0; pointer-events: none; }
.particle {
  position: absolute; background: white; border-radius: 50%; opacity: 0.3;
  box-shadow: 0 0 10px white; animation: float-up-ws infinite linear;
}

@keyframes float-up-ws {
  from { transform: translateY(0); opacity: 0; }
  20% { opacity: 0.4; }
  80% { opacity: 0.4; }
  to { transform: translateY(-100vh); opacity: 0; }
}

.nav-container {
  position: relative;
  z-index: 20;
  height: 100%;
}

.main-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 10;
  height: 100%;
}

.workspace-header {
  height: 56px;
  background: rgba(255, 255, 255, 0.4);
  backdrop-filter: blur(24px);
  border-bottom: 1px solid rgba(15, 23, 42, 0.05);
  display: flex;
  align-items: center;
  padding: 0 24px;
}

.content-view {
  flex: 1;
  overflow: hidden;
  position: relative;
}

.empty-state {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-glass-card {
  padding: 40px;
  background: rgba(255, 255, 255, 0.4);
  backdrop-filter: blur(40px);
  border-radius: 32px;
  border: 1px solid rgba(255, 255, 255, 0.8);
  text-align: center;
  box-shadow: 0 20px 40px rgba(0,0,0,0.03);
}
</style>
