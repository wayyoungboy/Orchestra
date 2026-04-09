<template>
  <div class="members-page-root animate-in fade-in zoom-in-95 duration-500">
    <div class="page-header">
      <div class="header-info">
        <h1 class="page-title">团队成员</h1>
        <p class="page-subtitle">{{ workspaceStore.currentWorkspace?.name }} · {{ members.length }} 位协作成员</p>
      </div>
      <div class="add-member-dropdown">
        <button @click="showAddDropdown = !showAddDropdown" class="add-member-btn">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" />
          </svg>
          <span>添加成员</span>
          <svg class="w-3 h-3 ml-1 transition-transform" :class="showAddDropdown ? 'rotate-180' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>
        <div v-if="showAddDropdown" class="dropdown-menu">
          <button @click="openAddModal('assistant')" class="dropdown-item">
            <span class="item-icon assistant">🤖</span>
            <div class="item-info">
              <span class="item-title">AI 助手</span>
              <span class="item-desc">添加 Claude/Gemini 等智能助手</span>
            </div>
          </button>
          <button @click="openAddModal('secretary')" class="dropdown-item">
            <span class="item-icon secretary">👁️</span>
            <div class="item-info">
              <span class="item-title">秘书</span>
              <span class="item-desc">协调任务分配给多个助手</span>
            </div>
          </button>
        </div>
      </div>
    </div>

    <div class="members-content custom-scrollbar">
      <div v-if="loading" class="flex flex-col items-center justify-center py-20 text-slate-400">
        <div class="animate-spin h-8 w-8 border-4 border-primary/20 border-t-primary rounded-full mb-4"></div>
        <p class="text-sm font-bold tracking-widest uppercase">Loading Members...</p>
      </div>

      <div v-else class="members-grid">
        <div v-for="member in members" :key="member.id" class="member-card">
          <div class="card-top">
            <div :class="['member-avatar', member.roleType === 'owner' ? 'is-owner' : '', member.roleType === 'assistant' ? 'is-ai' : '']">
              <svg v-if="member.roleType === 'assistant'" class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
              </svg>
              <span v-else>{{ member.name.charAt(0).toUpperCase() }}</span>
            </div>
            <div class="member-badge" :class="member.roleType">
              {{ roleLabel(member.roleType) }}
            </div>
          </div>
          
          <div class="member-info">
            <h3 class="member-name">{{ member.name }}</h3>
            <p v-if="member.acpEnabled" class="member-agent-hint">
              Agent: <code>{{ member.acpCommand || '未配置命令' }}</code>
            </p>
            <p v-else-if="member.roleType === 'assistant' || member.roleType === 'secretary'" class="member-agent-hint is-disabled">
              未启用 ACP 智能体
            </p>
            <p class="member-id">ID: {{ member.id.slice(0, 8) }}...</p>
          </div>

          <div class="member-actions">
            <button @click="handleEdit(member)" class="action-btn">配置</button>
            <button v-if="member.roleType !== 'owner'" @click="handleDelete(member.id)" class="action-btn is-danger">移除</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Modals -->
    <AddMemberModal v-if="showAddModal" :mode="addModalMode" @close="showAddModal = false" @invite="handleInviteMember" />
    <EditMemberModal v-if="editingMember" :member="editingMember" :show-remove="editingMember.roleType !== 'owner'" @close="editingMember = null" @save="handleSaveMember" @remove="handleDelete" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import client from '@/shared/api/client'
import { useWorkspaceStore } from '@/features/workspace/workspaceStore'
import { notifyUserError } from '@/shared/notifyError'
import AddMemberModal from './AddMemberModal.vue'
import EditMemberModal from './EditMemberModal.vue'

const workspaceStore = useWorkspaceStore()
const members = ref<any[]>([])
const loading = ref(false)
const showAddModal = ref(false)
const showAddDropdown = ref(false)
const addModalMode = ref<'assistant' | 'secretary'>('assistant')
const editingMember = ref<any>(null)

function openAddModal(mode: 'assistant' | 'secretary') {
  addModalMode.value = mode
  showAddModal.value = true
  showAddDropdown.value = false
}

async function loadMembers() {
  const wsId = workspaceStore.currentWorkspace?.id
  if (!wsId) return
  
  loading.value = true
  try {
    const response = await client.get(`/workspaces/${wsId}/members`)
    members.value = response.data || []
  } catch (e) {
    notifyUserError('Failed to load members', e)
  } finally {
    loading.value = false
  }
}

async function handleInviteMember(data: any) {
  const wsId = workspaceStore.currentWorkspace?.id
  if (!wsId) return

  try {
    await client.post(`/workspaces/${wsId}/members`, {
      name: data.name,
      roleType: data.roleType,
      terminalCommand: data.command,
      terminalType: data.terminalType
    })
    showAddModal.value = false
    await loadMembers()
  } catch (e) {
    notifyUserError('Failed to add member', e)
  }
}

function roleLabel(role: string) {
  const map: any = { owner: '所有者', admin: '管理员', assistant: 'AI 助手', secretary: '秘书', member: '成员' }
  return map[role] || role
}

function handleEdit(member: any) { editingMember.value = member }

async function handleSaveMember(id: string, name: string, acpEnabled: boolean, acpCommand: string, acpArgs: string[]) {
  const wsId = workspaceStore.currentWorkspace?.id
  if (!wsId) return

  try {
    await client.put(`/workspaces/${wsId}/members/${id}`, { name, acpEnabled, acpCommand, acpArgs })
    editingMember.value = null
    await loadMembers()
  } catch (e) {
    notifyUserError('Failed to update member', e)
  }
}

async function handleDelete(id: string) {
  const wsId = workspaceStore.currentWorkspace?.id
  if (!wsId) return

  if (confirm('确定要移除该成员吗？所有的聊天记录和终端会话都将被清理。')) {
    try {
      await client.delete(`/workspaces/${wsId}/members/${id}`)
      await loadMembers()
    } catch (e) {
      notifyUserError('Failed to remove member', e)
    }
  }
}

onMounted(loadMembers)
</script>

<style scoped>
.members-page-root {
  height: 100%; display: flex; flex-direction: column; gap: 32px; padding: 24px;
}

.page-header { display: flex; align-items: center; justify-content: space-between; }
.page-title { font-size: 32px; font-weight: 950; color: #0f172a; letter-spacing: -0.02em; }
.page-subtitle { font-size: 15px; font-weight: 600; color: #475569; margin-top: 6px; }

.add-member-btn {
  display: flex; align-items: center; gap: 10px; padding: 14px 28px;
  background: #4f46e5; color: white; border-radius: 18px;
  font-size: 15px; font-weight: 900; border: none; cursor: pointer;
  box-shadow: 0 15px 35px -5px rgba(79, 70, 229, 0.4);
  transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
}
.add-member-btn:hover { background: #4338ca; transform: translateY(-2px); box-shadow: 0 20px 40px -5px rgba(79, 70, 229, 0.5); }

.add-member-dropdown { position: relative; }
.dropdown-menu {
  position: absolute; top: calc(100% + 8px); right: 0; z-index: 50;
  min-width: 280px; background: white; border-radius: 20px;
  border: 1px solid #e2e8f0; box-shadow: 0 20px 50px rgba(0,0,0,0.15);
  padding: 8px; display: flex; flex-direction: column; gap: 4px;
}
.dropdown-item {
  display: flex; align-items: center; gap: 12px; padding: 12px 16px;
  border-radius: 14px; border: none; background: transparent;
  cursor: pointer; transition: all 0.2s; text-align: left;
}
.dropdown-item:hover { background: #f8fafc; }
.item-icon { font-size: 20px; width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center; justify-content: center; }
.item-icon.assistant { background: #dcfce7; }
.item-icon.secretary { background: #fef9c3; }
.item-icon.admin { background: #fee2e2; }
.item-icon.member { background: #f1f5f9; }
.item-info { display: flex; flex-direction: column; gap: 2px; }
.item-title { font-size: 14px; font-weight: 800; color: #0f172a; }
.item-desc { font-size: 11px; color: #64748b; }

.members-content { flex: 1; overflow-y: auto; padding-bottom: 40px; }

.members-grid {
  display: grid; grid-template-cols: repeat(auto-fill, minmax(280px, 1fr)); gap: 24px;
}

.member-card {
  background: rgba(255, 255, 255, 0.85); backdrop-filter: blur(32px);
  border-radius: 32px; padding: 32px; border: 1px solid white;
  display: flex; flex-direction: column; gap: 24px;
  transition: all 0.4s cubic-bezier(0.23, 1, 0.32, 1);
  box-shadow: 0 10px 30px rgba(0,0,0,0.04);
}
.member-card:hover { transform: translateY(-6px); background: white; border-color: rgba(99, 102, 241, 0.2); box-shadow: 0 30px 60px -12px rgba(0,0,0,0.06); }

.card-top { display: flex; align-items: flex-start; justify-content: space-between; }

.member-avatar {
  width: 56px; height: 56px; border-radius: 18px;
  background: #f1f5f9; color: #64748b;
  display: flex; align-items: center; justify-content: center;
  font-size: 20px; font-weight: 900; transition: all 0.3s;
}
.member-avatar.is-owner { background: rgba(99, 102, 241, 0.1); color: #4f46e5; }
.member-avatar.is-ai { background: rgba(16, 185, 129, 0.1); color: #10b981; }

.member-badge {
  padding: 5px 12px; border-radius: 100px; font-size: 10px; font-weight: 950;
  text-transform: uppercase; letter-spacing: 0.1em;
}
.member-badge.owner { background: #fee2e2; color: #ef4444; }
.member-badge.admin { background: #ede9fe; color: #7c3aed; }
.member-badge.assistant { background: #dcfce7; color: #10b981; }
.member-badge.secretary { background: #fef9c3; color: #ca8a04; }
.member-badge.member { background: #f1f5f9; color: #64748b; }

.member-info h3 { font-size: 18px; font-weight: 800; color: #0f172a; }
.member-terminal-hint { font-size: 12px; color: #64748b; margin-top: 4px; }
.member-terminal-hint code { background: rgba(15, 23, 42, 0.05); padding: 2px 6px; border-radius: 6px; font-family: monospace; }
.member-id { font-size: 11px; font-weight: 600; color: #cbd5e1; font-family: monospace; margin-top: 8px; }

.member-actions { display: flex; gap: 10px; }
.action-btn {
  flex: 1; padding: 10px; border-radius: 12px; border: 1px solid #e2e8f0;
  background: white; color: #475569; font-size: 13px; font-weight: 800;
  cursor: pointer; transition: all 0.2s;
}
.action-btn:hover { background: #f8fafc; color: #0f172a; border-color: #cbd5e1; }
.action-btn.is-danger:hover { background: #fef2f2; color: #ef4444; border-color: #fecaca; }
</style>
