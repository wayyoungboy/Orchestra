<template>
  <div class="fixed inset-0 bg-slate-900/20 backdrop-blur-sm flex items-center justify-center z-[100] p-4">
    <div class="bg-white/80 backdrop-blur-2xl rounded-[2.5rem] w-full max-w-3xl max-h-[85vh] flex flex-col border border-white shadow-[0_40px_100px_-20px_rgba(0,0,0,0.15)] overflow-hidden animate-in fade-in zoom-in-95 duration-300">
      <!-- 头部 -->
      <div class="flex items-center justify-between px-8 py-6 border-b border-slate-100">
        <div>
          <h2 class="text-xl font-black text-slate-900 tracking-tight">选择工作目录</h2>
          <p class="text-xs text-slate-400 font-bold uppercase tracking-widest mt-1">Select Server Directory</p>
        </div>
        <button @click="$emit('close')" class="w-10 h-10 rounded-full flex items-center justify-center text-slate-400 hover:bg-slate-100 hover:text-slate-900 transition-all">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- 路径导航输入 -->
      <div class="px-8 py-5 bg-slate-50/50 border-b border-slate-100">
        <div class="relative group">
          <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 group-focus-within:text-primary transition-colors">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
          </div>
          <input
            v-model="currentPath"
            type="text"
            class="w-full bg-white border border-slate-200 rounded-2xl pl-12 pr-4 py-3.5 text-slate-900 text-sm font-medium placeholder-slate-300 focus:border-primary/50 focus:ring-4 focus:ring-primary/5 outline-none transition-all"
            placeholder="输入或粘贴服务器路径..."
            @keyup.enter="loadDirectory(currentPath)"
          />
          <!-- Validation Badges -->
          <div v-if="validation" class="absolute right-4 inset-y-0 flex items-center gap-2">
            <span v-if="validation.exists && validation.writable" class="px-2 py-1 rounded bg-green-50 text-[10px] font-black text-green-600 border border-green-100 uppercase tracking-tight">Writable</span>
            <span v-else-if="validation.exists && !validation.writable" class="px-2 py-1 rounded bg-amber-50 text-[10px] font-black text-amber-600 border border-amber-100 uppercase tracking-tight">Read Only</span>
            <span v-else-if="!validation.exists && currentPath" class="px-2 py-1 rounded bg-red-50 text-[10px] font-black text-red-600 border border-red-100 uppercase tracking-tight">Not Found</span>
          </div>
        </div>
      </div>

      <!-- 目录列表 -->
      <div class="flex-1 overflow-y-auto p-6 custom-scrollbar">
        <div v-if="loading" class="flex flex-col items-center justify-center py-20 text-slate-400">
          <div class="animate-spin h-8 w-8 border-4 border-primary/20 border-t-primary rounded-full mb-4"></div>
          <p class="text-sm font-bold tracking-widest uppercase">Scanning Directory...</p>
        </div>
        
        <div v-else-if="error" class="bg-red-50 border border-red-100 rounded-2xl p-8 text-center">
          <svg class="w-12 h-12 text-red-300 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <p class="text-red-500 font-bold">{{ error }}</p>
        </div>

        <div v-else-if="!files.length" class="flex flex-col items-center justify-center py-20 text-slate-300">
          <svg class="w-16 h-16 mb-4 opacity-20" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0a2 2 0 01-2 2H6a2 2 0 01-2-2m16 0l-8 8-8-8" />
          </svg>
          <p class="text-lg font-bold">空目录</p>
        </div>

        <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-2">
          <button
            v-for="file in files"
            :key="file.path"
            @click="handleFileClick(file)"
            class="group w-full flex items-center gap-4 px-4 py-3 rounded-xl hover:bg-primary/5 border border-transparent hover:border-primary/10 transition-all duration-200 text-left"
          >
            <div class="w-10 h-10 rounded-lg bg-slate-100 flex items-center justify-center group-hover:bg-white group-hover:shadow-sm transition-all">
              <svg class="w-5 h-5 text-slate-400 group-hover:text-primary transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
              </svg>
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-sm font-bold text-slate-700 group-hover:text-slate-900 truncate">{{ file.name }}</p>
              <p class="text-[10px] text-slate-400 font-medium uppercase tracking-tighter">Folder</p>
            </div>
          </button>
        </div>
      </div>

      <!-- 底部操作栏 -->
      <div class="p-8 bg-slate-50/80 border-t border-slate-100 flex items-center gap-4">
        <div class="flex-1 relative">
          <input
            v-model="workspaceName"
            type="text"
            class="w-full bg-white border border-slate-200 rounded-2xl px-5 py-3 text-slate-900 font-bold placeholder-slate-300 focus:border-primary/50 focus:ring-4 focus:ring-primary/5 outline-none transition-all"
            placeholder="工作区显示名称..."
          />
        </div>
        <button
          @click="confirm"
          :disabled="!currentPath || !workspaceName || loading"
          class="px-10 py-3 bg-gradient-primary text-white rounded-2xl font-black text-sm shadow-glow-primary hover:brightness-110 active:scale-[0.98] transition-all disabled:opacity-40 disabled:grayscale disabled:cursor-not-allowed flex items-center gap-2"
        >
          <span>创建工作区</span>
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M13 7l5 5m0 0l-5 5m5-5H6" />
          </svg>
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

function defaultWorkspaceNameFromPath(path: string): string {
  const trimmed = path.replace(/[/\\]+$/, '')
  const parts = trimmed.split(/[/\\]/)
  return parts[parts.length - 1] || trimmed || ''
}

function applyPathAndDefaultName(basePath: string, entries: FileInfo[] | null) {
  currentPath.value = basePath
  files.value = entries || []
  // If user hasn't typed a name yet, or we're just loading, update the name
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

interface PathValidation {
  exists: boolean
  isDir: boolean
  readable: boolean
  writable: boolean
  error?: string
}

const validation = ref<PathValidation | null>(null)

async function loadDirectory(path: string) {
  loading.value = true
  error.value = null
  validation.value = null
  try {
    const response = await workspaceApi.browseRoot(path)
    const data: BrowseResult = response.data
    applyPathAndDefaultName(data.basePath, data.files)
    
    // Auto-validate current path
    if (data.basePath) {
      // Assuming browseRoot also returns validation info in newer API
      // If not, we could call a specific validate endpoint
    }
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
