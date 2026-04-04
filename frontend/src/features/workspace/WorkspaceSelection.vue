<template>
  <div class="min-h-screen flex flex-col items-center justify-center p-6">
    <!-- 打开新工作区 -->
    <div class="w-full max-w-2xl mb-10">
      <button
        @click="showPathBrowser = true"
        class="w-full bg-panel/40 rounded-3xl p-10 flex flex-col items-center justify-center text-center hover:bg-panel/60 transition-all border border-white/5 hover:border-primary/20"
      >
        <div class="w-14 h-14 rounded-full bg-white/5 flex items-center justify-center mb-4 border border-white/10">
          <svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 19a2 2 0 01-2-2V7a2 2 0 012-2h4l2 2h4a2 2 0 012 2v1M5 19h14a2 2 0 002-2v-5a2 2 0 00-2-2H9a2 2 0 00-2 2v5a2 2 0 01-2 2z" />
          </svg>
        </div>
        <h1 class="text-2xl font-bold text-white mb-2">{{ t('workspace.selectionOpenTitle') }}</h1>
        <p class="text-gray-400 font-medium">{{ t('workspace.selectionOpenSubtitle') }}</p>
      </button>
    </div>

    <!-- 最近工作区 -->
    <div class="w-full max-w-6xl">
      <h2 class="text-xs font-bold text-gray-500 tracking-widest uppercase mb-6">{{ t('workspace.selectionRecent') }}</h2>

      <div v-if="!workspaceStore.recentWorkspaces.length" class="text-center text-white/50 py-12">
        <p class="text-sm">{{ t('workspace.selectionRecentEmpty') }}</p>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <button
          v-for="ws in workspaceStore.recentWorkspaces"
          :key="ws.id"
          @click="openWorkspace(ws.id)"
          class="bg-panel/40 rounded-3xl p-5 text-left hover:bg-panel/60 transition-all border border-white/5 hover:border-primary/20"
        >
          <div class="w-10 h-10 rounded-2xl bg-white/10 flex items-center justify-center mb-4 border border-white/10">
            <svg class="w-5 h-5 text-white/60" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
          </div>
          <h3 class="text-lg font-bold text-white mb-1 truncate">{{ ws.name }}</h3>
          <p class="text-xs text-white/40 truncate">{{ ws.path }}</p>
        </button>
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
  router.push(`/workspace/${id}`)
}

async function handlePathSelect(path: string, name: string) {
  const ws = await workspaceStore.createWorkspace(name, path)
  if (ws) {
    router.push(`/workspace/${ws.id}`)
  }
  showPathBrowser.value = false
}
</script>