<template>
  <div class="settings-page-root animate-in fade-in slide-in-from-bottom-4 duration-500">
    <!-- Settings Sidebar -->
    <aside class="settings-nav">
      <div class="nav-header">
        <h2 class="nav-title">设置中心</h2>
      </div>
      <div class="nav-items">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          @click="activeTab = tab.id"
          :class="['nav-item', activeTab === tab.id ? 'is-active' : '']"
        >
          <svg class="w-4 h-4 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path v-if="tab.id === 'general'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path v-else-if="tab.id === 'apiKeys'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l3.354-3.354A6 6 0 1121 9z" />
            <path v-else-if="tab.id === 'workspace'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H5a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
            <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
          {{ tab.label }}
        </button>
      </div>
    </aside>

    <!-- Settings Content -->
    <main class="settings-content custom-scrollbar">
      <div v-if="activeTab === 'general'" class="settings-section">
        <h3 class="section-title">通用设置</h3>
        <div class="settings-grid">
          <div class="setting-card">
            <label>语言 (Language)</label>
            <select class="setting-select" :value="currentLocale" @change="handleLocaleChange">
              <option value="zh">简体中文</option>
              <option value="en">English</option>
            </select>
          </div>
          <div class="setting-card">
            <label>主题模式</label>
            <div class="theme-toggle">
              <button
                :class="['theme-btn', theme === 'light' ? 'is-active' : '']"
                @click="handleThemeChange('light')"
              >Light</button>
              <button
                :class="['theme-btn', theme === 'dark' ? 'is-active' : '']"
                @click="handleThemeChange('dark')"
              >Dark</button>
            </div>
          </div>
        </div>
      </div>

      <div v-if="activeTab === 'apiKeys'" class="settings-section">
        <ApiKeysSection />
      </div>

      <div v-if="activeTab === 'workspace'" class="settings-section">
        <h3 class="section-title">工作区配置</h3>
        <form @submit.prevent="handleUpdateWorkspace" class="settings-grid">
          <div class="setting-card">
            <label>工作区名称</label>
            <input
              v-model="editWorkspace.name"
              class="setting-input"
              placeholder="例如: Orchestra Backend"
            />
          </div>
          <div class="setting-card">
            <label>物理路径 (Server Path)</label>
            <input
              v-model="editWorkspace.path"
              class="setting-input"
              placeholder="/volumes/code/project"
            />
          </div>
          <div class="form-actions">
            <button type="submit" :disabled="isSaving" class="save-btn">
              {{ isSaving ? '保存中...' : '保存更改' }}
            </button>
          </div>
        </form>
      </div>

      <div v-if="activeTab === 'account'" class="settings-section">
        <h3 class="section-title">账号与安全</h3>
        <div class="settings-grid">
          <div class="setting-card">
            <label>当前用户</label>
            <p class="setting-value">{{ authStore.currentUser }}</p>
          </div>
          <button @click="handleLogout" class="logout-btn">退出登录</button>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/features/auth/authStore'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'
import ApiKeysSection from './ApiKeysSection.vue'

const { locale } = useI18n()
const authStore = useAuthStore()
const workspaceStore = useWorkspaceStore()
const router = useRouter()
const activeTab = ref('general')
const isSaving = ref(false)
const theme = ref<'light' | 'dark'>((localStorage.getItem('orchestra.theme') as 'light' | 'dark') || 'light')

const currentLocale = computed(() => locale.value)

const editWorkspace = reactive({
  name: '',
  path: ''
})

const tabs = [
  { id: 'general', label: '通用' },
  { id: 'apiKeys', label: 'API 密钥' },
  { id: 'workspace', label: '工作区' },
  { id: 'account', label: '账号' }
]

function handleLocaleChange(event: Event) {
  const newLocale = (event.target as HTMLSelectElement).value
  locale.value = newLocale as 'en' | 'zh'
  localStorage.setItem('orchestra.locale', newLocale)
}

function handleThemeChange(newTheme: 'light' | 'dark') {
  theme.value = newTheme
  localStorage.setItem('orchestra.theme', newTheme)
  applyTheme(newTheme)
}

function applyTheme(t: 'light' | 'dark') {
  const root = document.documentElement
  if (t === 'dark') {
    root.classList.add('dark-theme')
  } else {
    root.classList.remove('dark-theme')
  }
}

onMounted(() => {
  // Restore saved locale
  const savedLocale = localStorage.getItem('orchestra.locale')
  if (savedLocale && (savedLocale === 'en' || savedLocale === 'zh')) {
    locale.value = savedLocale
  }
  // Apply saved theme
  applyTheme(theme.value)

  if (workspaceStore.currentWorkspace) {
    editWorkspace.name = workspaceStore.currentWorkspace.name
    editWorkspace.path = workspaceStore.currentWorkspace.path
  }
})

async function handleUpdateWorkspace() {
  if (!workspaceStore.currentWorkspace) return
  
  isSaving.value = true
  try {
    await workspaceStore.updateWorkspace(workspaceStore.currentWorkspace.id, {
      name: editWorkspace.name,
      path: editWorkspace.path
    })
  } catch (e) {
    // Error handled by store
  } finally {
    isSaving.value = false
  }
}

function handleLogout() {
  authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.settings-page-root { height: 100%; display: flex; gap: 24px; padding: 24px; }

.settings-nav {
  width: 240px; background: rgba(255, 255, 255, 0.35); backdrop-filter: blur(32px);
  border-radius: 32px; padding: 24px; border: 1px solid rgba(255, 255, 255, 0.5);
  display: flex; flex-direction: column; gap: 24px;
}

.nav-title { font-size: 18px; font-weight: 900; color: rgb(var(--color-overlay)); }
.nav-items { display: flex; flex-direction: column; gap: 6px; }
.nav-item {
  width: 100%; height: 44px; display: flex; align-items: center; padding: 0 16px;
  border-radius: 14px; border: none; background: transparent; color: rgb(var(--color-overlay) / 0.5);
  font-size: 14px; font-weight: 700; cursor: pointer; transition: all 0.2s;
}
.nav-item:hover { background: rgba(255, 255, 255, 0.4); color: rgb(var(--color-overlay)); }
.nav-item.is-active { background: rgb(var(--color-primary)); color: rgb(var(--color-on-primary)); box-shadow: 0 10px 20px -5px rgba(79, 70, 229, 0.3); }

.settings-content {
  flex: 1; background: rgba(255, 255, 255, 0.25); backdrop-filter: blur(40px);
  border-radius: 40px; padding: 48px; border: 1px solid rgba(255, 255, 255, 0.4); overflow-y: auto;
}

.section-title { font-size: 24px; font-weight: 950; color: rgb(var(--color-overlay)); margin-bottom: 40px; letter-spacing: -0.02em; }
.settings-grid { display: flex; flex-direction: column; gap: 32px; max-width: 520px; }

.setting-card { display: flex; flex-direction: column; gap: 12px; }
.setting-card label { font-size: 11px; font-weight: 900; color: rgb(var(--color-overlay) / 0.4); text-transform: uppercase; letter-spacing: 0.15em; margin-left: 4px; }

.setting-input, .setting-select {
  width: 100%; padding: 14px 18px; border-radius: 16px; border: 1px solid rgba(255, 255, 255, 0.5);
  background: rgba(255, 255, 255, 0.4); color: rgb(var(--color-overlay)); font-size: 15px; font-weight: 600; outline: none;
  backdrop-filter: blur(8px); transition: all 0.2s;
}
.setting-input:focus, .setting-select:focus { border-color: rgb(var(--color-primary)); box-shadow: 0 0 0 4px rgba(79, 70, 229, 0.1); }

.form-actions { margin-top: 12px; }
.save-btn {
  padding: 14px 32px; background: rgb(var(--color-primary)); color: rgb(var(--color-on-primary)); border-radius: 16px;
  font-size: 14px; font-weight: 900; border: none; cursor: pointer;
  box-shadow: 0 10px 25px -5px rgba(79, 70, 229, 0.4); transition: all 0.3s;
}
.save-btn:hover { background: rgb(var(--color-primary-hover)); transform: translateY(-1px); }
.save-btn:disabled { opacity: 0.5; cursor: not-allowed; }

.setting-value { font-size: 16px; font-weight: 700; color: rgb(var(--color-overlay)); }
.theme-toggle { display: flex; gap: 4px; padding: 4px; background: rgba(255, 255, 255, 0.3); border-radius: 12px; width: fit-content; }
.theme-btn {
  padding: 8px 16px; border-radius: 10px; border: none; background: transparent;
  font-size: 13px; font-weight: 700; color: rgb(var(--color-overlay) / 0.5); cursor: pointer; transition: all 0.2s;
}
.theme-btn.is-active { background: rgba(255, 255, 255, 0.6); color: rgb(var(--color-primary)); box-shadow: 0 2px 8px rgba(0,0,0,0.05); }

.logout-btn {
  margin-top: 20px; padding: 14px; border-radius: 16px; border: 1px solid rgba(239, 68, 68, 0.3);
  background: rgba(239, 68, 68, 0.1); color: #ef4444; font-weight: 800; cursor: pointer; transition: all 0.2s;
}
.logout-btn:hover { background: rgba(239, 68, 68, 0.2); transform: translateY(-1px); }
</style>
