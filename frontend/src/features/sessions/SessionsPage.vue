<template>
  <div class="sessions-page-root animate-in fade-in zoom-in-95 duration-500">
    <div class="page-header">
      <div class="header-info">
        <h1 class="page-title">Agent 会话</h1>
        <p class="page-subtitle">{{ workspaceStore.currentWorkspace?.name }} · {{ sessions.length }} 个活跃后台会话</p>
      </div>
      <button class="refresh-btn" :disabled="loading" @click="loadSessions">
        <svg class="w-4 h-4" :class="loading ? 'animate-spin' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M4 4v6h6M20 20v-6h-6M20 9a8 8 0 00-14.906-4M4 15a8 8 0 0014.906 4" />
        </svg>
        <span>{{ loading ? '刷新中' : '刷新' }}</span>
      </button>
    </div>

    <div class="sessions-content custom-scrollbar">
      <div v-if="loading" class="loading-state">
        <div class="loading-spinner"></div>
        <p>Loading Sessions...</p>
      </div>

      <div v-else-if="sessions.length === 0" class="empty-state">
        <div class="empty-icon">
          <svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 8h10M7 12h6m-7 8h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
        </div>
        <p class="empty-title">暂无活跃 Agent 会话</p>
        <p class="empty-subtitle">在成员页启动 assistant 或 secretary 后，会话会显示在这里。</p>
      </div>

      <div v-else class="sessions-grid">
        <div v-for="session in enrichedSessions" :key="session.sessionId" class="session-card">
          <div class="card-top">
            <div class="session-avatar">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
              </svg>
            </div>
            <span class="status-pill">运行中</span>
          </div>

          <div class="session-info">
            <h3>{{ session.memberName }}</h3>
            <p>{{ roleLabel(session.roleType) }}</p>
          </div>

          <div class="session-meta">
            <span>Session</span>
            <code>{{ session.sessionId }}</code>
          </div>
          <div class="session-meta">
            <span>Member</span>
            <code>{{ session.memberId }}</code>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import client from '@/shared/api/client'
import { notifyUserError } from '@/shared/notifyError'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'

type WorkspaceSession = {
  memberId: string
  sessionId: string
}

type Member = {
  id: string
  name: string
  roleType: string
}

const workspaceStore = useWorkspaceStore()
const loading = ref(false)
const sessions = ref<WorkspaceSession[]>([])
const members = ref<Member[]>([])

const memberById = computed(() => {
  return new Map(members.value.map((member) => [member.id, member]))
})

const enrichedSessions = computed(() => {
  return sessions.value.map((session) => {
    const member = memberById.value.get(session.memberId)
    return {
      ...session,
      memberName: member?.name || 'Unknown Member',
      roleType: member?.roleType || 'assistant',
    }
  })
})

function roleLabel(role: string) {
  const map: Record<string, string> = {
    owner: '所有者',
    admin: '管理员',
    assistant: 'AI 助手',
    secretary: '秘书',
    member: '成员',
  }
  return map[role] || role
}

async function loadSessions() {
  const wsId = workspaceStore.currentWorkspace?.id
  if (!wsId) return

  loading.value = true
  try {
    const [sessionResponse, memberResponse] = await Promise.all([
      client.get(`/workspaces/${wsId}/terminal-sessions`),
      client.get(`/workspaces/${wsId}/members`),
    ])
    sessions.value = sessionResponse.data?.sessions || []
    members.value = memberResponse.data || []
  } catch (e) {
    notifyUserError('Failed to load agent sessions', e)
  } finally {
    loading.value = false
  }
}

onMounted(loadSessions)
</script>

<style scoped>
.sessions-page-root {
  height: 100%; display: flex; flex-direction: column; gap: 28px; padding: 24px;
}

.page-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
.page-title { font-size: 32px; font-weight: 950; color: #0f172a; }
.page-subtitle { font-size: 15px; font-weight: 600; color: #475569; margin-top: 6px; }

.refresh-btn {
  display: flex; align-items: center; gap: 10px; padding: 12px 18px;
  border-radius: 14px; border: 1px solid #c7d2fe; background: white;
  color: #4338ca; font-size: 14px; font-weight: 900; cursor: pointer;
  box-shadow: 0 10px 24px rgba(79, 70, 229, 0.12);
}
.refresh-btn:disabled { cursor: not-allowed; opacity: 0.7; }

.sessions-content { flex: 1; overflow-y: auto; padding-bottom: 40px; }
.loading-state, .empty-state {
  height: 100%; min-height: 320px; display: flex; flex-direction: column;
  align-items: center; justify-content: center; color: #64748b; text-align: center;
}
.loading-spinner {
  width: 32px; height: 32px; border-radius: 999px; border: 4px solid rgba(79, 70, 229, 0.16);
  border-top-color: #4f46e5; animation: spin 1s linear infinite; margin-bottom: 14px;
}
.loading-state p { font-size: 12px; font-weight: 900; letter-spacing: 0.12em; text-transform: uppercase; }
.empty-icon {
  width: 56px; height: 56px; border-radius: 18px; display: flex; align-items: center; justify-content: center;
  background: #eef2ff; color: #4f46e5; margin-bottom: 16px;
}
.empty-title { font-size: 17px; font-weight: 900; color: #0f172a; }
.empty-subtitle { font-size: 13px; font-weight: 600; color: #64748b; margin-top: 6px; }

.sessions-grid {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 18px;
}
.session-card {
  background: rgba(255, 255, 255, 0.9); border: 1px solid rgba(226, 232, 240, 0.9);
  border-radius: 18px; padding: 20px; display: flex; flex-direction: column; gap: 18px;
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.05);
}
.card-top { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.session-avatar {
  width: 44px; height: 44px; border-radius: 14px; display: flex; align-items: center; justify-content: center;
  background: rgba(16, 185, 129, 0.1); color: #059669;
}
.status-pill {
  padding: 5px 10px; border-radius: 999px; background: #dcfce7; color: #059669;
  font-size: 10px; font-weight: 950; letter-spacing: 0.08em;
}
.session-info h3 { font-size: 17px; font-weight: 900; color: #0f172a; }
.session-info p { font-size: 12px; font-weight: 800; color: #64748b; margin-top: 4px; }
.session-meta { min-width: 0; display: flex; flex-direction: column; gap: 5px; }
.session-meta span { font-size: 10px; font-weight: 950; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.08em; }
.session-meta code {
  display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  padding: 8px 10px; border-radius: 10px; background: #f8fafc; color: #334155;
  font-size: 12px; font-weight: 800;
}

@keyframes spin { to { transform: rotate(360deg); } }
</style>
