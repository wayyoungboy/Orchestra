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
            <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
          {{ tab.label }}
        </button>
      </div>
    </aside>

    <!-- Settings Content -->
    <main class="settings-content">
      <div v-if="activeTab === 'general'" class="settings-section">
        <h3 class="section-title">通用设置</h3>
        <div class="settings-grid">
          <div class="setting-card">
            <label>语言 (Language)</label>
            <select class="setting-select">
              <option>简体中文</option>
              <option>English</option>
            </select>
          </div>
          <div class="setting-card">
            <label>主题模式</label>
            <div class="theme-toggle">
              <div class="toggle-active">清新磨砂 (Light)</div>
            </div>
          </div>
        </div>
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
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/features/auth/authStore'

const authStore = useAuthStore()
const router = useRouter()
const activeTab = ref('general')

const tabs = [
  { id: 'general', label: '通用' },
  { id: 'account', label: '账号' }
]

function handleLogout() {
  authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.settings-page-root {
  height: 100%;
  display: flex;
  gap: 24px;
  padding: 24px;
}

.settings-nav {
  width: 240px;
  background: rgba(255, 255, 255, 0.4);
  backdrop-filter: blur(24px);
  border-radius: 24px;
  padding: 24px;
  border: 1px solid white;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.nav-title { font-size: 18px; font-weight: 900; color: #0f172a; }

.nav-items { display: flex; flex-direction: column; gap: 6px; }
.nav-item {
  width: 100%; height: 44px; display: flex; align-items: center; padding: 0 16px;
  border-radius: 12px; border: none; background: transparent; color: #64748b;
  font-size: 14px; font-weight: 700; cursor: pointer; transition: all 0.2s;
}
.nav-item:hover { background: rgba(255, 255, 255, 0.6); color: #0f172a; }
.nav-item.is-active { background: #4f46e5; color: white; shadow: 0 10px 20px -5px rgba(79, 70, 229, 0.3); }

.settings-content {
  flex: 1;
  background: rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(32px);
  border-radius: 32px;
  padding: 40px;
  border: 1px solid white;
  overflow-y: auto;
}

.section-title { font-size: 22px; font-weight: 900; color: #0f172a; margin-bottom: 32px; }

.settings-grid { display: flex; flex-direction: column; gap: 24px; max-width: 500px; }

.setting-card {
  display: flex; flex-direction: column; gap: 10px;
}
.setting-card label { font-size: 11px; font-weight: 900; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.1em; }

.setting-select {
  width: 100%; padding: 12px 16px; border-radius: 14px; border: 1px solid #e2e8f0;
  background: white; color: #0f172a; font-weight: 600; outline: none;
}

.setting-value { font-size: 16px; font-weight: 700; color: #0f172a; }

.theme-toggle {
  padding: 4px; background: #f1f5f9; border-radius: 12px; width: fit-content;
}
.toggle-active {
  padding: 8px 16px; background: white; border-radius: 10px; shadow: 0 2px 8px rgba(0,0,0,0.05);
  font-size: 13px; font-weight: 700; color: #4f46e5;
}

.logout-btn {
  margin-top: 20px; padding: 14px; border-radius: 14px; border: 1px solid #fee2e2;
  background: #fef2f2; color: #ef4444; font-weight: 800; cursor: pointer; transition: all 0.2s;
}
.logout-btn:hover { background: #fee2e2; transform: translateY(-1px); }
</style>
