<template>
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
    <div class="bg-panel rounded-3xl w-full max-w-2xl max-h-[80vh] flex flex-col border border-white/10">
      <!-- 头部 -->
      <div class="flex items-center justify-between p-4 border-b border-white/5">
        <h2 class="text-lg font-bold text-white">选择工作目录</h2>
        <button @click="$emit('close')" class="text-white/40 hover:text-white">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- 路径输入 -->
      <div class="p-4 border-b border-white/5">
        <input
          v-model="currentPath"
          type="text"
          class="w-full bg-surface text-white rounded-xl px-4 py-3 border border-white/10 focus:border-primary/50 focus:outline-none"
          placeholder="输入路径..."
          @keyup.enter="loadDirectory(currentPath)"
        />
      </div>

      <!-- 目录列表（仅文件夹，由 API 过滤） -->
      <div class="flex-1 overflow-y-auto p-4">
        <div v-if="loading" class="text-center py-8 text-white/50">
          加载中...
        </div>
        <div v-else-if="error" class="text-center py-8 text-red-400">
          {{ error }}
        </div>
        <div v-else-if="!files.length" class="text-center py-8 text-white/50">
          空目录
        </div>
        <div v-else class="space-y-1">
          <button
            v-for="file in files"
            :key="file.path"
            @click="handleFileClick(file)"
            class="w-full flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-white/5 transition-colors"
          >
            <svg class="w-5 h-5 text-white/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
            <span class="text-white/80">{{ file.name }}</span>
          </button>
        </div>
      </div>

      <!-- 底部 -->
      <div class="flex items-center justify-between p-4 border-t border-white/5">
        <input
          v-model="workspaceName"
          type="text"
          class="flex-1 bg-surface text-white rounded-xl px-4 py-2 mr-4 border border-white/10 focus:border-primary/50 focus:outline-none"
          placeholder="工作区名称"
        />
        <button
          @click="confirm"
          :disabled="!currentPath || !workspaceName"
          class="px-6 py-2 bg-primary text-white rounded-xl font-medium disabled:opacity-50 disabled:cursor-not-allowed hover:bg-primary-hover transition-colors"
        >
          确认
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { workspaceApi } from '@/shared/api/workspace'
import { getApiErrorMessage } from '@/shared/api/errors'
import type { FileInfo, BrowseResult } from '@/shared/types/workspace'

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'select', path: string, name: string): void
}>()

const currentPath = ref('')
const workspaceName = ref('')
const files = ref<FileInfo[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

/** 工作区名称默认 = 当前目录文件夹名（支持 / 与 \\，忽略末尾分隔符） */
function defaultWorkspaceNameFromPath(path: string): string {
  const trimmed = path.replace(/[/\\]+$/, '')
  const parts = trimmed.split(/[/\\]/)
  return parts[parts.length - 1] || trimmed || ''
}

function applyPathAndDefaultName(basePath: string, entries: FileInfo[]) {
  currentPath.value = basePath
  files.value = entries
  workspaceName.value = defaultWorkspaceNameFromPath(basePath)
}

onMounted(async () => {
  await loadRoot()
})

async function loadRoot() {
  loading.value = true
  error.value = null
  try {
    const response = await workspaceApi.browseRoot()
    const data: BrowseResult = response.data
    applyPathAndDefaultName(data.basePath, data.files)
  } catch (e) {
    error.value = getApiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

async function loadDirectory(path: string) {
  loading.value = true
  error.value = null
  try {
    const response = await workspaceApi.browseRoot(path)
    const data: BrowseResult = response.data
    applyPathAndDefaultName(data.basePath, data.files)
  } catch (e) {
    error.value = getApiErrorMessage(e)
  } finally {
    loading.value = false
  }
}

function handleFileClick(file: FileInfo) {
  if (file.isDir) {
    loadDirectory(file.path)
  }
}

function confirm() {
  if (currentPath.value && workspaceName.value) {
    emit('select', currentPath.value, workspaceName.value)
  }
}
</script>