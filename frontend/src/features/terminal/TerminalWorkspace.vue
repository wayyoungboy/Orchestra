<template>
  <div class="terminal-workspace-root">
    <!-- Tab Bar -->
    <div class="tab-bar">
      <div v-for="session in terminalStore.sessions" :key="session.id" 
           :class="['tab-item', terminalStore.activeSessionId === session.id ? 'is-active' : '']"
           @click="terminalStore.setActiveSession(session.id)">
        <svg class="w-3.5 h-3.5 mr-2 opacity-60" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
        <span class="tab-label truncate">{{ session.name }}</span>
        <button @click.stop="terminalStore.closeSession(session.id)" class="tab-close">×</button>
      </div>
      <button @click="handleCreateSession" class="new-tab-btn">+</button>
    </div>

    <!-- Terminal Canvas Wrapper -->
    <div class="terminal-container">
      <div v-if="!terminalStore.sessions.length" class="no-sessions">
        <div class="empty-glass-card">
          <div class="terminal-icon-placeholder">
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
          </div>
          <h3>没有运行中的终端</h3>
          <p>点击上方 "+" 按钮或从成员列表启动一个新的 PTY 会话。</p>
        </div>
      </div>
      <div v-else class="terminal-canvas-box">
        <TerminalPane
          v-for="session in terminalStore.sessions"
          v-show="terminalStore.activeSessionId === session.id"
          :key="session.id"
          :session-id="session.id"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useTerminalStore } from './terminalStore'
import TerminalPane from './TerminalPane.vue'

const terminalStore = useTerminalStore()

function handleCreateSession() {
  terminalStore.createSession('bash')
}

onMounted(() => {
  terminalStore.loadSessions()
})
</script>

<style scoped>
.terminal-workspace-root {
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.tab-bar {
  display: flex;
  align-items: flex-end;
  gap: 4px;
  padding: 0 16px;
  height: 44px;
}

.tab-item {
  height: 36px;
  min-width: 120px;
  max-width: 200px;
  background: rgba(255, 255, 255, 0.3);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.5);
  border-bottom: none;
  border-top-left-radius: 12px;
  border-top-right-radius: 12px;
  display: flex;
  align-items: center;
  padding: 0 12px;
  cursor: pointer;
  transition: all 0.2s;
  color: #64748b;
  font-size: 13px;
  font-weight: 600;
}

.tab-item:hover { background: rgba(255, 255, 255, 0.5); color: #0f172a; }

.tab-item.is-active {
  height: 40px;
  background: #ffffff;
  color: #4f46e5;
  box-shadow: 0 -4px 15px rgba(0,0,0,0.02);
  z-index: 2;
}

.tab-close {
  margin-left: auto;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  transition: all 0.2s;
}
.tab-close:hover { background: rgba(15, 23, 42, 0.05); color: #ef4444; }

.new-tab-btn {
  width: 32px;
  height: 32px;
  margin-bottom: 4px;
  margin-left: 8px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.3);
  color: #64748b;
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}
.new-tab-btn:hover { background: white; color: #4f46e5; transform: scale(1.05); }

.terminal-container {
  flex: 1;
  background: white;
  border-radius: 24px;
  border-top-left-radius: 0; /* Align with tabs */
  border: 1px solid white;
  box-shadow: 0 40px 100px -20px rgba(0, 0, 0, 0.06);
  overflow: hidden;
  position: relative;
  margin: 0 12px 12px;
}

.no-sessions {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-glass-card { text-align: center; max-width: 360px; }
.terminal-icon-placeholder {
  width: 64px; height: 64px; background: #f1f5f9; border-radius: 20px;
  display: flex; align-items: center; justify-content: center;
  margin: 0 auto 24px; color: #94a3b8;
}
.terminal-icon-placeholder svg { width: 32px; height: 32px; }
.empty-glass-card h3 { font-size: 18px; font-weight: 800; color: #0f172a; margin-bottom: 8px; }
.empty-glass-card p { font-size: 14px; color: #64748b; line-height: 1.6; }

.terminal-canvas-box {
  height: 100%;
  width: 100%;
  background: #0f172a; /* Dark PTY background remains */
}
</style>
