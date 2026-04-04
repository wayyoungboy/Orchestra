<template>
  <div class="selection-container">
    <div class="selection-content">
      <!-- Left Column: Branding / Welcome -->
      <div class="branding-section">
        <div class="version-badge reveal-item" style="animation-delay: 0.1s">
          <span class="badge-dot"></span>
          <span class="badge-text">VERSION 1.0 ALPHA</span>
        </div>
        
        <div class="welcome-group">
          <h1 class="welcome-title reveal-item" style="animation-delay: 0.2s">
            Ready to <br /><span class="gradient-text">orchestrate?</span>
          </h1>
          <p class="welcome-sub reveal-item" style="animation-delay: 0.3s">
            选择一个现有的工作区开始协作，或者创建一个新的目录来编排您的 AI 劳动力。
          </p>
        </div>

        <!-- Create New Workspace Card -->
        <div class="create-card-wrapper reveal-item" style="animation-delay: 0.4s">
          <button @click="showPathBrowser = true" class="create-btn-card">
            <div class="card-glow"></div>
            <div class="create-icon">
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" /></svg>
            </div>
            <div class="create-text">
              <h3>Create New Workspace</h3>
              <p>Initialize a new project directory on the server</p>
            </div>
            <div class="arrow-icon">
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M9 5l7 7-7 7" /></svg>
            </div>
          </button>
        </div>
      </div>

      <!-- Right Column: Recent Workspaces -->
      <div class="list-section reveal-item" style="animation-delay: 0.5s">
        <div class="list-header">
          <span class="list-label">RECENT WORKSPACES</span>
          <div class="header-line"></div>
        </div>

        <div v-if="!workspaceStore.recentWorkspaces.length" class="empty-list-card">
          <p>No recent workspaces found.</p>
        </div>

        <div v-else class="workspace-grid">
          <button
            v-for="ws in workspaceStore.recentWorkspaces"
            :key="ws.id"
            @click="openWorkspace(ws.id)"
            class="workspace-item-card"
          >
            <div class="ws-icon">
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" /></svg>
            </div>
            <div class="ws-info">
              <h4>{{ ws.name }}</h4>
              <code>{{ ws.path }}</code>
            </div>
            <div class="ws-arrow">
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M13 7l5 5m0 0l-5 5m5-5H6" /></svg>
            </div>
          </button>
        </div>
      </div>
    </div>

    <!-- 路径浏览器弹窗 -->
    <PathBrowser
      v-if="showPathBrowser"
      @close="showPathBrowser = false"
      @select="handlePathSelect"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useWorkspaceStore } from './workspaceStore'
import PathBrowser from './PathBrowser.vue'

const { t } = useI18n()
const router = useRouter()
const workspaceStore = useWorkspaceStore()
const showPathBrowser = ref(false)

onMounted(() => {
  workspaceStore.loadWorkspaces()
})

async function openWorkspace(id: string) {
  await workspaceStore.openWorkspace(id)
  router.push(`/workspace/${id}/chat`)
}

async function handlePathSelect(path: string, name: string) {
  const ws = await workspaceStore.createWorkspace(name, path)
  if (ws) {
    router.push(`/workspace/${ws.id}/chat`)
  }
  showPathBrowser.value = false
}
</script>

<style scoped>
.selection-container {
  height: 100%;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.selection-content {
  width: 100%;
  max-width: 1200px;
  display: flex;
  align-items: flex-start;
  gap: 80px;
  padding: 0 40px;
}

/* Animations */
.reveal-item {
  opacity: 0;
  transform: translateY(20px);
  animation: reveal 0.8s forwards cubic-bezier(0.22, 1, 0.36, 1);
}
@keyframes reveal { to { opacity: 1; transform: translateY(0); } }

/* Left Column: Branding */
.branding-section {
  flex: 6;
  display: flex;
  flex-direction: column;
  gap: 40px;
}

.version-badge {
  display: flex; align-items: center; gap: 8px; padding: 6px 14px;
  background: white; border: 1px solid rgba(15, 23, 42, 0.08); border-radius: 100px;
  width: fit-content; box-shadow: 0 4px 15px rgba(0,0,0,0.02);
}
.badge-dot { width: 6px; height: 6px; background: #6366f1; border-radius: 50%; box-shadow: 0 0 10px #6366f1; }
.badge-text { font-size: 10px; font-weight: 900; color: #64748b; letter-spacing: 0.15em; }

.welcome-title { font-size: 72px; font-weight: 950; line-height: 1.05; color: #0f172a; letter-spacing: -0.03em; }
.gradient-text {
  background: linear-gradient(to right, #4f46e5, #8b5cf6, #10b981);
  background-size: 200% auto; -webkit-background-clip: text; -webkit-text-fill-color: transparent;
  animation: shine-text 6s linear infinite;
}
@keyframes shine-text { to { background-position: 200% center; } }

.welcome-sub { font-size: 18px; color: #475569; line-height: 1.6; max-width: 480px; font-weight: 500; }

/* Create Card */
.create-btn-card {
  width: 100%;
  background: rgba(255, 255, 255, 0.5);
  backdrop-filter: blur(32px);
  border-radius: 32px;
  border: 1px solid white;
  padding: 32px;
  display: flex;
  align-items: center;
  gap: 24px;
  text-align: left;
  position: relative;
  overflow: hidden;
  transition: all 0.4s;
  box-shadow: 0 30px 60px -12px rgba(0,0,0,0.05);
  cursor: pointer;
}

.create-btn-card:hover {
  transform: translateY(-4px);
  background: rgba(255, 255, 255, 0.7);
  box-shadow: 0 40px 80px -12px rgba(99, 102, 241, 0.15);
  border-color: rgba(99, 102, 241, 0.3);
}

.create-icon {
  width: 64px; height: 64px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border-radius: 20px;
  display: flex; align-items: center; justify-content: center;
  color: white; shadow: 0 10px 20px rgba(99, 102, 241, 0.3);
}
.create-icon svg { width: 32px; height: 32px; }

.create-text h3 { font-size: 20px; font-weight: 800; color: #0f172a; margin-bottom: 4px; }
.create-text p { font-size: 14px; color: #64748b; font-weight: 500; }

.arrow-icon { margin-left: auto; color: #cbd5e1; transition: transform 0.3s; }
.create-btn-card:hover .arrow-icon { transform: translateX(4px); color: #6366f1; }

/* Right Column: List */
.list-section {
  flex: 5;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.list-header { display: flex; align-items: center; gap: 16px; margin-bottom: 8px; }
.list-label { font-size: 11px; font-weight: 900; color: #94a3b8; letter-spacing: 0.2em; white-space: nowrap; }
.header-line { flex: 1; height: 1px; background: rgba(15, 23, 42, 0.05); }

.workspace-grid { display: flex; flex-direction: column; gap: 12px; }

.workspace-item-card {
  width: 100%;
  background: rgba(255, 255, 255, 0.4);
  backdrop-filter: blur(20px);
  border-radius: 20px;
  padding: 16px 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  border: 1px solid rgba(255, 255, 255, 0.5);
  transition: all 0.3s;
  cursor: pointer;
}

.workspace-item-card:hover {
  background: white;
  transform: translateX(6px);
  border-color: rgba(99, 102, 241, 0.2);
  box-shadow: 0 10px 30px rgba(0,0,0,0.03);
}

.ws-icon {
  width: 44px; height: 44px;
  background: #f1f5f9; border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  color: #64748b; transition: all 0.3s;
}
.workspace-item-card:hover .ws-icon { background: rgba(99, 102, 241, 0.1); color: #6366f1; }

.ws-info { text-align: left; flex: 1; min-width: 0; }
.ws-info h4 { font-size: 15px; font-weight: 700; color: #0f172a; margin-bottom: 2px; }
.ws-info code { font-size: 11px; color: #94a3b8; font-family: inherit; }

.ws-arrow { opacity: 0; transform: translateX(-10px); transition: all 0.3s; color: #6366f1; }
.workspace-item-card:hover .ws-arrow { opacity: 1; transform: translateX(0); }

.empty-list-card {
  padding: 40px; text-align: center; color: #94a3b8; font-size: 14px; font-weight: 500;
  border: 2px dashed rgba(15, 23, 42, 0.05); border-radius: 24px;
}

@media (max-width: 1024px) {
  .selection-content { flex-direction: column; gap: 60px; padding: 60px 24px; }
  .branding-section { flex: 1; align-items: center; text-align: center; }
  .welcome-title { font-size: 48px; }
  .list-section { flex: 1; width: 100%; }
}
</style>
